package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDayOffsetMiddleware_missingQueryDefaultsToZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var got int
	r := gin.New()
	r.GET("/", DayOffsetMiddleware(), func(c *gin.Context) {
		got = GetDayOffset(c)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if got != 0 {
		t.Fatalf("GetDayOffset: got %d want 0", got)
	}
}

func TestDayOffsetMiddleware_validOffset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var got int
	r := gin.New()
	r.GET("/", DayOffsetMiddleware(), func(c *gin.Context) {
		got = GetDayOffset(c)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/?offset=7", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if got != 7 {
		t.Fatalf("GetDayOffset: got %d want 7", got)
	}
}

func TestDayOffsetMiddleware_invalidOffsetReturns400AndSkipsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handlerRan := false
	r := gin.New()
	r.GET("/", DayOffsetMiddleware(), func(c *gin.Context) {
		handlerRan = true
		_ = GetDayOffset(c)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/?offset=xyz", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d want %d", w.Code, http.StatusBadRequest)
	}
	if handlerRan {
		t.Fatal("handler ran after invalid offset")
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error != "offset must be an integer" {
		t.Fatalf("error message: got %q", body.Error)
	}
}
