package auth

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/pgtest"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// pgUsers is a fresh Postgres database with a users table (TEST_DATABASE_URL; skipped without).
func pgUsers(t *testing.T) *gorm.DB {
	t.Helper()
	// Skipped without Postgres, or fails when TEST_DATABASE_REQUIRED is set (CI).
	adminDB, adminURL := pgtest.Admin(t)
	admin := adminURL.String()
	name := fmt.Sprintf("auth_test_%d", rand.Uint32())
	if _, err := adminDB.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(admin)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if d, err := db.DB(); err == nil {
			d.Close()
		}
		adminDB.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)")
	})
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	return db
}

// The signed-in user becomes the request transaction's identity, which row-level security reads.
func TestAuthMiddleware_setsRequestIdentityForRLS(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	db := pgUsers(t)
	createUser(t, db, "first", false)
	wanda := createUser(t, db, "wanda", true)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(database.Middleware(db))
	r.GET("/x", AuthMiddleware(db), func(c *gin.Context) {
		var row struct{ UID, Admin string }
		if err := database.From(c.Request.Context()).
			Raw("SELECT current_setting('app.user_id', true) AS uid, current_setting('app.is_admin', true) AS admin").
			Scan(&row).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, row)
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.AddCookie(&http.Cookie{Name: AuthTokenCookieName, Value: tokenFor(t, wanda)})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	want := fmt.Sprintf(`{"UID":"%d","Admin":"true"}`, wanda.ID)
	if rec.Code != http.StatusOK || rec.Body.String() != want {
		t.Fatalf("status %d body %s, want %s", rec.Code, rec.Body.String(), want)
	}
}
