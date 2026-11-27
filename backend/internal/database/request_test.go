package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type note struct {
	ID   uint
	Body string
}

func sqliteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if d, err := db.DB(); err == nil {
		d.SetMaxOpenConns(1) // one in-memory database
	}
	if err := db.AutoMigrate(&note{}); err != nil {
		t.Fatal(err)
	}
	return db
}

// pgDB is an empty database on TEST_DATABASE_URL with a notes table, on a pool of one
// connection, so consecutive requests reuse the same session. Skipped without Postgres.
func pgDB(t *testing.T) *gorm.DB {
	t.Helper()
	admin, u := pgAdmin(t)
	name := fmt.Sprintf("request_test_%d", rand.Uint32())
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	sqlDB, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		sqlDB.Close()
		admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)")
	})
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&note{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func countNotes(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := db.Model(&note{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

// serveOne runs one request through gin.Recovery and Middleware to handler.
func serveOne(t *testing.T, db *gorm.DB, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery(), Middleware(db))
	r.POST("/x", handler)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", nil))
	return rec
}

func insertNote(c *gin.Context) error {
	return From(c.Request.Context()).Create(&note{Body: "x"}).Error
}

func TestMiddleware_commitsOnSuccess(t *testing.T) {
	db := sqliteDB(t)
	rec := serveOne(t, db, func(c *gin.Context) {
		if err := insertNote(c); err != nil {
			t.Fatal(err)
		}
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})
	if rec.Code != http.StatusCreated || countNotes(t, db) != 1 {
		t.Fatalf("status %d, notes %d: want 201 and the note kept", rec.Code, countNotes(t, db))
	}
}

func TestMiddleware_rollsBackOnErrorStatus(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError} {
		db := sqliteDB(t)
		rec := serveOne(t, db, func(c *gin.Context) {
			if err := insertNote(c); err != nil {
				t.Fatal(err)
			}
			c.JSON(status, gin.H{"error": "nope"})
		})
		if rec.Code != status || countNotes(t, db) != 0 {
			t.Fatalf("status %d: notes %d, want the write rolled back", rec.Code, countNotes(t, db))
		}
	}
}

func TestMiddleware_rollsBackWhenNothingIsWritten(t *testing.T) {
	db := sqliteDB(t)
	serveOne(t, db, func(c *gin.Context) {
		if err := insertNote(c); err != nil {
			t.Fatal(err)
		}
		c.AbortWithStatus(http.StatusConflict)
	})
	if countNotes(t, db) != 0 {
		t.Fatal("aborted request kept its write")
	}
	serveOne(t, db, func(c *gin.Context) {
		if err := insertNote(c); err != nil {
			t.Fatal(err)
		}
		// No body: the transaction ends when the handler returns, with the default 200.
	})
	if countNotes(t, db) != 1 {
		t.Fatal("request with an implicit 200 lost its write")
	}
}

func TestMiddleware_rollsBackOnPanic(t *testing.T) {
	db := sqliteDB(t)
	rec := serveOne(t, db, func(c *gin.Context) {
		if err := insertNote(c); err != nil {
			t.Fatal(err)
		}
		panic("boom")
	})
	if rec.Code != http.StatusInternalServerError || countNotes(t, db) != 0 {
		t.Fatalf("status %d, notes %d: want 500 and nothing kept", rec.Code, countNotes(t, db))
	}
}

func TestMiddleware_lazy_noQueryNoTransaction(t *testing.T) {
	db := sqliteDB(t)
	rec := serveOne(t, db, func(c *gin.Context) {
		if c.Request.Context().Value(handleKey{}).(handle).req.Begun() {
			t.Error("transaction began before any From call")
		}
		c.Status(http.StatusOK)
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if sqlDB, err := db.DB(); err != nil || sqlDB.Stats().InUse != 0 || sqlDB.Stats().OpenConnections > 1 {
		t.Fatalf("pool after a request without queries: %+v, %v", sqlDB.Stats(), err)
	}
}

func TestFrom_afterTheResponseCarriesErrRequestDone(t *testing.T) {
	db := sqliteDB(t)
	var late error
	serveOne(t, db, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
		late = insertNote(c)
	})
	if !errors.Is(late, ErrRequestDone) {
		t.Fatalf("query after the response: %v, want ErrRequestDone", late)
	}
	if countNotes(t, db) != 0 {
		t.Fatal("a query after the response was written")
	}
}

// The error handle can't open a stray transaction on the pool either.
func TestFrom_errorHandleCannotBeginATransaction(t *testing.T) {
	db := sqliteDB(t)
	var err error
	serveOne(t, db, func(c *gin.Context) {
		c.Status(http.StatusOK)
		c.Writer.WriteHeaderNow()
		err = From(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			return tx.Create(&note{Body: "late"}).Error
		})
		if sqlDB, _ := db.DB(); sqlDB.Stats().InUse != 0 {
			t.Errorf("%d connections in use after a Transaction on the error handle", sqlDB.Stats().InUse)
		}
	})
	// gorm joins Begin's own error onto it with %v, so match the text.
	if err == nil || !strings.Contains(err.Error(), ErrRequestDone.Error()) || countNotes(t, db) != 0 {
		t.Fatalf("err %v, notes %d: want ErrRequestDone and nothing written", err, countNotes(t, db))
	}
}

func TestFrom_withoutHandlePanics(t *testing.T) {
	defer func() {
		if p := recover(); p == nil || !strings.Contains(fmt.Sprint(p), "no database handle") {
			t.Fatalf("panic = %v", p)
		}
	}()
	From(context.Background())
}

func TestFrom_withDBUsesItDirectly(t *testing.T) {
	db := sqliteDB(t)
	ctx := WithDB(context.Background(), db)
	if err := From(ctx).Create(&note{Body: "x"}).Error; err != nil {
		t.Fatal(err)
	}
	if countNotes(t, db) != 1 {
		t.Fatal("WithDB write missing")
	}
}

// A failed nested transaction is a savepoint: it undoes only its own work, and the request can
// keep going and commit the rest.
func TestNestedTransactionIsASavepoint(t *testing.T) {
	for name, open := range map[string]func(*testing.T) *gorm.DB{"sqlite": sqliteDB, "postgres": pgDB} {
		t.Run(name, func(t *testing.T) {
			db := open(t)
			rec := serveOne(t, db, func(c *gin.Context) {
				ctx := c.Request.Context()
				if err := From(ctx).Create(&note{Body: "outer"}).Error; err != nil {
					t.Fatal(err)
				}
				err := From(ctx).Transaction(func(tx *gorm.DB) error {
					if err := tx.Create(&note{Body: "inner"}).Error; err != nil {
						return err
					}
					// A real SQL error, which on Postgres aborts the transaction up to the savepoint.
					return tx.Exec("SELECT * FROM no_such_table").Error
				})
				if err == nil {
					t.Fatal("inner transaction should fail")
				}
				// A best-effort read that fails doesn't poison the request either.
				if err := Optional(ctx, func() error {
					return From(ctx).Exec("SELECT * FROM no_such_table").Error
				}); err == nil {
					t.Fatal("optional read should fail")
				}
				if err := From(ctx).Create(&note{Body: "after"}).Error; err != nil {
					t.Fatalf("request unusable after a failed savepoint: %v", err)
				}
				c.Status(http.StatusOK)
			})
			var bodies []string
			db.Model(&note{}).Order("id").Pluck("body", &bodies)
			if rec.Code != http.StatusOK || strings.Join(bodies, ",") != "outer,after" {
				t.Fatalf("status %d, notes %v: want outer,after", rec.Code, bodies)
			}
		})
	}
}

// The identity is transaction-local: visible to the request's queries, gone for the next user
// of the same pooled connection (pgDB has exactly one).
func TestIdentityIsTransactionLocal(t *testing.T) {
	db := pgDB(t)
	setting := func(q *gorm.DB) (uid, admin string) {
		t.Helper()
		row := struct{ UID, Admin sql.NullString }{}
		if err := q.Raw("SELECT current_setting('app.user_id', true) AS uid, current_setting('app.is_admin', true) AS admin").Scan(&row).Error; err != nil {
			t.Fatal(err)
		}
		return row.UID.String, row.Admin.String
	}
	serveOne(t, db, func(c *gin.Context) {
		ctx := c.Request.Context()
		// Set before the transaction begins (as AuthMiddleware does)...
		if err := SetIdentity(ctx, Identity{UserID: 42, IsAdmin: true}); err != nil {
			t.Fatal(err)
		}
		if uid, admin := setting(From(ctx)); uid != "42" || admin != "true" {
			t.Errorf("in request: app.user_id=%q app.is_admin=%q", uid, admin)
		}
		// ...and changed after: applied to the open transaction right away.
		if err := SetIdentity(ctx, Identity{UserID: 7}); err != nil {
			t.Fatal(err)
		}
		if uid, admin := setting(From(ctx)); uid != "7" || admin != "false" {
			t.Errorf("after SetIdentity: app.user_id=%q app.is_admin=%q", uid, admin)
		}
		c.Status(http.StatusOK)
	})
	if uid, admin := setting(db); uid != "" || admin != "" {
		t.Fatalf("identity leaked to the pooled connection: app.user_id=%q app.is_admin=%q", uid, admin)
	}
}

// A commit that fails must not reach the client as a success: the handler's 200 becomes a 500.
func TestMiddleware_failedCommitAnswers500(t *testing.T) {
	db := pgDB(t)
	// A deferred foreign key is only checked at COMMIT.
	if err := db.Exec(`CREATE TABLE gate (id int PRIMARY KEY, other int REFERENCES gate (id) DEFERRABLE INITIALLY DEFERRED)`).Error; err != nil {
		t.Fatal(err)
	}
	rec := serveOne(t, db, func(c *gin.Context) {
		if err := insertNote(c); err != nil {
			t.Fatal(err)
		}
		if err := From(c.Request.Context()).Exec("INSERT INTO gate VALUES (1, 99)").Error; err != nil {
			t.Fatal(err)
		}
		c.SetCookie("auth_token", "for-a-rolled-back-user", 60, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"saved": true})
	})
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "saved") {
		t.Fatalf("status %d body %s: want a 500 instead of the success body", rec.Code, rec.Body.String())
	}
	if c := rec.Header().Get("Set-Cookie"); c != "" {
		t.Fatalf("the 500 carried the handler's cookie: %s", c)
	}
	if countNotes(t, db) != 0 {
		t.Fatal("note kept although the commit failed")
	}
}
