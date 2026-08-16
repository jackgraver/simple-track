package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBenchmarkEndpointIsNotRegisteredInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(BenchmarkMiddleware(router))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/benchmark", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", recorder.Code)
	}
}
