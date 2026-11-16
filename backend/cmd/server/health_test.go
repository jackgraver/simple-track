package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestHealthFailsWhenDatabaseIsUnreachable(t *testing.T) {
	// Nothing listens on port 1; DisableAutomaticPing lets Open succeed so
	// the handler is what finds out.
	db, err := gorm.Open(postgres.Open("host=127.0.0.1 port=1 user=x dbname=x sslmode=disable connect_timeout=1"),
		&gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", healthHandler(db))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503; body %s", rec.Code, rec.Body)
	}
}
