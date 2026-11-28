package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"be-simpletracker/internal/core/media/lookup"
	"be-simpletracker/internal/core/media/models"
	"be-simpletracker/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// External lookups must not run while the request holds a database connection or transaction
// (the app role's idle_in_transaction_session_timeout would kill a slow one). The request
// transaction is lazy and the lookup routes never call database.From.
func TestLookupHoldsNoDatabaseConnection(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.MediaItem{}, &models.UserMedia{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}

	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if inUse := sqlDB.Stats().InUse; inUse != 0 {
			t.Errorf("%d database connections in use during the external call", inUse)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"results":[{"collectionId":1,"collectionName":"Album","artistName":"A","releaseDate":"2020-01-01T00:00:00Z"}]}`)
	}))
	defer upstream.Close()
	svc := lookup.NewService("")
	svc.ITunesBase = upstream.URL

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(database.Middleware(db))
	// What AuthMiddleware does to the request transaction.
	r.Use(func(c *gin.Context) {
		if err := database.SetIdentity(c.Request.Context(), database.Identity{UserID: 1}); err != nil {
			t.Fatal(err)
		}
	})
	RegisterMediaRoutes(r.Group("/media"), svc)

	for _, path := range []string{"/media/lookup?type=album&q=x", "/media/lookup/details?source=itunes&id=album/1"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
			t.Fatalf("%s: status %d body %s", path, rec.Code, rec.Body.String())
		}
	}
	if calls == 0 {
		t.Fatal("the fake upstream was never called")
	}
}

func TestRespondLookupErrorFixedMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		err    error
		status int
		msg    string
	}{
		{fmt.Errorf("%w: Get \"https://x?api_key=SECRET\"", lookup.ErrUpstream), 502, "Lookup service unavailable"},
		{fmt.Errorf("%w: set SECRET", lookup.ErrNotConfigured), 503, "Lookup is not configured"},
		{fmt.Errorf("%w: SECRET", lookup.ErrUnsupported), 400, "Lookup is not supported for this media type"},
		{fmt.Errorf("%w: SECRET", lookup.ErrNotFound), 404, "Lookup result not found"},
		{errors.New("SECRET"), 502, "Lookup service unavailable"},
	}
	for _, tc := range cases {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/lookup", nil)
		respondLookupError(c, tc.err)
		if rec.Code != tc.status || strings.Contains(rec.Body.String(), "SECRET") || !strings.Contains(rec.Body.String(), tc.msg) {
			t.Errorf("%v: got %d %s", tc.err, rec.Code, rec.Body.String())
		}
	}
}
