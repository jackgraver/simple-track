package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestDayOffsetMiddleware_dateTakesPrecedence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var got int
	r := gin.New()
	r.GET("/", DayOffsetMiddleware(), func(c *gin.Context) {
		got = GetDayOffset(c)
	})
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/?offset=5&date="+yesterday, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if got != 1 {
		t.Fatalf("GetDayOffset: got %d want 1", got)
	}
}

func TestDayOffsetMiddleware_invalidDateReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", DayOffsetMiddleware(), func(c *gin.Context) {})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/?date=22-11-2026", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d want %d", w.Code, http.StatusBadRequest)
	}
}

func TestOffsetForDate_acrossDST(t *testing.T) {
	loc, err := time.LoadLocation("America/Edmonton")
	if err != nil {
		t.Skip("no tzdata")
	}
	now := time.Date(2026, 3, 9, 1, 0, 0, 0, loc) // the day after spring-forward
	day := time.Date(2026, 3, 7, 0, 0, 0, 0, loc)
	if got := OffsetForDate(day, now); got != 2 {
		t.Fatalf("got %d want 2", got)
	}
	if got := OffsetForDate(time.Date(2026, 3, 10, 0, 0, 0, 0, loc), now); got != -1 {
		t.Fatalf("future: got %d want -1", got)
	}
}
