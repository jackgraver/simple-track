package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// hardeningRouter echoes the body length, or 400 if reading it failed.
func hardeningRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(hardening(16, map[string]int64{"POST /big/:id": 64}))
	echo := func(c *gin.Context) {
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"n": len(b)})
	}
	r.POST("/small", echo)
	r.POST("/big/:id", echo)
	r.GET("/small", echo)
	return r
}

// chunked hides the length so only MaxBytesReader can catch it.
type chunked struct{ io.Reader }

func TestHardeningBodyLimit(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		size    int
		chunked bool
		want    int
	}{
		{"default under", "/small", 16, false, http.StatusOK},
		{"default over", "/small", 17, false, http.StatusRequestEntityTooLarge},
		{"default over chunked", "/small", 17, true, http.StatusBadRequest},
		{"override under", "/big/7", 64, false, http.StatusOK},
		{"override over", "/big/7", 65, false, http.StatusRequestEntityTooLarge},
		{"override over chunked", "/big/7", 65, true, http.StatusBadRequest},
		{"unknown route", "/nope", 17, false, http.StatusRequestEntityTooLarge},
	}
	r := hardeningRouter()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader = strings.NewReader(strings.Repeat("x", tc.size))
			if tc.chunked {
				body = chunked{body}
			}
			req := httptest.NewRequest(http.MethodPost, tc.path, body)
			if tc.chunked {
				req.ContentLength = -1
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d; body %s", rec.Code, tc.want, rec.Body)
			}
		})
	}
}

func TestHardeningNoStore(t *testing.T) {
	r := hardeningRouter()
	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/small", nil),
		httptest.NewRequest(http.MethodPost, "/small", strings.NewReader(strings.Repeat("x", 100))), // 413
		httptest.NewRequest(http.MethodGet, "/nope", nil),                                           // 404
	} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("%s %s (%d): Cache-Control = %q, want no-store", req.Method, req.URL.Path, rec.Code, got)
		}
	}
}

// Every bodyLimits key must name a real route, or the larger limit silently stops applying.
func TestBodyLimitRoutesExist(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=127.0.0.1 port=1 user=x dbname=x sslmode=disable connect_timeout=1"),
		&gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	CreateFeatures(db, router)
	routes := map[string]bool{}
	for _, r := range router.Routes() {
		routes[r.Method+" "+r.Path] = true
	}
	for key, limit := range bodyLimits {
		if !routes[key] {
			t.Errorf("bodyLimits[%q] matches no registered route", key)
		}
		if limit > 6_000_000 {
			t.Errorf("bodyLimits[%q] = %d is above Caddy's request_body max_size (6MB)", key, limit)
		}
	}
}
