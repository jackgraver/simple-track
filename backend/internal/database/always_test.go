package database

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func noteBodies(t *testing.T, db *gorm.DB) string {
	t.Helper()
	var bodies []string
	if err := db.Model(&note{}).Order("id").Pluck("body", &bodies).Error; err != nil {
		t.Fatal(err)
	}
	return strings.Join(bodies, ",")
}

func auditNote(body string) func(*gorm.DB) error {
	return func(db *gorm.DB) error { return db.Create(&note{Body: body}).Error }
}

// Both databases: SQLite for the module-test path, Postgres for the real one. pgDB's pool has one
// connection, so an Always that ran while the request still held its connection would block
// until its timeout and the audit row would be missing.
var alwaysDBs = map[string]func(*testing.T) *gorm.DB{"sqlite": sqliteDB, "postgres": pgDB}

func TestAlways_runsAfterCommit(t *testing.T) {
	for name, open := range alwaysDBs {
		t.Run(name, func(t *testing.T) {
			db := open(t)
			var sawRequestRow bool
			var lateErr error
			rec := serveOne(t, db, func(c *gin.Context) {
				ctx := c.Request.Context()
				if err := insertNote(c); err != nil {
					t.Fatal(err)
				}
				Always(ctx, func(tx *gorm.DB) error {
					// After the commit: the request's row is visible from a new transaction, and
					// the request handle is finished.
					sawRequestRow = countNotes(t, tx) == 1
					lateErr = From(ctx).Create(&note{Body: "late"}).Error
					return tx.Create(&note{Body: "audit"}).Error
				})
				c.JSON(http.StatusCreated, gin.H{})
			})
			if rec.Code != http.StatusCreated || noteBodies(t, db) != "x,audit" {
				t.Fatalf("status %d, notes %q: want 201 and x,audit", rec.Code, noteBodies(t, db))
			}
			if !sawRequestRow {
				t.Error("Always ran before the request committed")
			}
			if !errors.Is(lateErr, ErrRequestDone) {
				t.Errorf("From inside Always: %v, want ErrRequestDone", lateErr)
			}
		})
	}
}

func TestAlways_survivesRollback(t *testing.T) {
	for name, open := range alwaysDBs {
		t.Run(name, func(t *testing.T) {
			db := open(t)
			rec := serveOne(t, db, func(c *gin.Context) {
				if err := insertNote(c); err != nil {
					t.Fatal(err)
				}
				Always(c, auditNote("rejected"))
				c.JSON(http.StatusGone, gin.H{"error": "link expired"})
			})
			if rec.Code != http.StatusGone || noteBodies(t, db) != "rejected" {
				t.Fatalf("status %d, notes %q: want 410 and only the Always row", rec.Code, noteBodies(t, db))
			}
		})
	}
}

func TestAlways_survivesPanic(t *testing.T) {
	for name, open := range alwaysDBs {
		t.Run(name, func(t *testing.T) {
			db := open(t)
			rec := serveOne(t, db, func(c *gin.Context) {
				if err := insertNote(c); err != nil {
					t.Fatal(err)
				}
				Always(c, auditNote("before panic"))
				panic("boom")
			})
			if rec.Code != http.StatusInternalServerError || noteBodies(t, db) != "before panic" {
				t.Fatalf("status %d, notes %q: want 500 and only the Always row", rec.Code, noteBodies(t, db))
			}
		})
	}
}

// Calls run in registration order, and each in its own savepoint: one that fails with a SQL error
// (which on Postgres aborts the transaction) or panics loses only its own write.
func TestAlways_inOrder_failureIsolatedAndResponseUnchanged(t *testing.T) {
	for name, open := range alwaysDBs {
		t.Run(name, func(t *testing.T) {
			db := open(t)
			rec := serveOne(t, db, func(c *gin.Context) {
				if err := insertNote(c); err != nil {
					t.Fatal(err)
				}
				Always(c, auditNote("first"))
				Always(c, func(tx *gorm.DB) error {
					if err := tx.Create(&note{Body: "lost"}).Error; err != nil {
						return err
					}
					return tx.Exec("SELECT * FROM no_such_table").Error
				})
				Always(c, func(*gorm.DB) error { panic("audit bug") })
				Always(c, auditNote("last"))
				c.JSON(http.StatusOK, gin.H{"saved": true})
			})
			if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "saved") {
				t.Fatalf("status %d body %s: a failing Always changed the response", rec.Code, rec.Body.String())
			}
			if got := noteBodies(t, db); got != "x,first,last" {
				t.Fatalf("notes %q, want x,first,last", got)
			}
		})
	}
}

// The Always transaction has the request's identity, also when it was set after registering.
func TestAlways_seesIdentity(t *testing.T) {
	db := pgDB(t)
	var uid, admin sql.NullString
	serveOne(t, db, func(c *gin.Context) {
		ctx := c.Request.Context()
		Always(ctx, func(tx *gorm.DB) error {
			return tx.Raw("SELECT current_setting('app.user_id', true), current_setting('app.is_admin', true)").Row().Scan(&uid, &admin)
		})
		if err := SetIdentity(ctx, Identity{UserID: 42, IsAdmin: true}); err != nil {
			t.Fatal(err)
		}
		c.Status(http.StatusForbidden)
	})
	if uid.String != "42" || admin.String != "true" {
		t.Fatalf("Always saw app.user_id=%q app.is_admin=%q, want 42/true", uid.String, admin.String)
	}
	var leaked sql.NullString
	if err := db.Raw("SELECT current_setting('app.user_id', true)").Row().Scan(&leaked); err != nil || leaked.String != "" {
		t.Fatalf("identity leaked to the pooled connection: %q, %v", leaked.String, err)
	}
}

// A client that goes away cancels the request (and rolls it back) but not its audit write.
// Postgres only: database/sql discards a connection whose transaction was cancelled, and with it
// SQLite's in-memory database.
func TestAlways_survivesClientDisconnect(t *testing.T) {
	db := pgDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery(), Middleware(db))
	ctx, cancel := context.WithCancel(context.Background())
	r.POST("/x", func(c *gin.Context) {
		if err := insertNote(c); err != nil {
			t.Fatal(err)
		}
		Always(c, auditNote("audit"))
		cancel()
		c.Status(http.StatusOK)
	})
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/x", nil).WithContext(ctx))
	if got := noteBodies(t, db); got != "audit" {
		t.Fatalf("notes %q: want only the audit row (request rolled back, audit kept)", got)
	}
}

// Without a request transaction (OutsideRequestTx, tools, tests) Always runs at once.
func TestAlways_withDBRunsImmediately(t *testing.T) {
	db := sqliteDB(t)
	ctx := WithDB(context.Background(), db)
	Always(ctx, auditNote("now"))
	if noteBodies(t, db) != "now" {
		t.Fatal("Always on a WithDB context didn't run immediately")
	}
	// A failure is logged, not returned or panicked.
	Always(ctx, func(*gorm.DB) error { return errors.New("nope") })
}

func TestAlways_withoutHandlePanics(t *testing.T) {
	defer func() {
		if p := recover(); p == nil || !strings.Contains(p.(string), "no database handle") {
			t.Fatalf("panic = %v", p)
		}
	}()
	Always(context.Background(), auditNote("x"))
}
