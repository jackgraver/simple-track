package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseQueryInt_optionalDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/?foo=1", nil)

	v, err := ParseQueryInt(c, QueryIntVar{Key: "missing", Default: 42})
	if err != nil || v != 42 {
		t.Fatalf("got %d, %v", v, err)
	}
}

func TestParseQueryInt_invalidUsesErrInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/?n=abc", nil)

	_, err := ParseQueryInt(c, QueryIntVar{Key: "n", ErrInvalid: "bad number"})
	if err == nil || err.Error() != "bad number" {
		t.Fatalf("got %v", err)
	}
}

func TestParseQueryInt_requiredMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := ParseQueryInt(c, QueryIntVar{Key: "id", Required: true, ErrMissing: "id required"})
	if err == nil || err.Error() != "id required" {
		t.Fatalf("got %v", err)
	}
}
