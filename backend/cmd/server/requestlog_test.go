package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestLogger_logsRoutePatternNotPathOrQuery(t *testing.T) {
	var buf bytes.Buffer
	old := gin.DefaultWriter
	gin.DefaultWriter = &buf
	defer func() { gin.DefaultWriter = old }()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(requestLogger()...)
	r.GET("/media/:id", func(c *gin.Context) { c.Status(http.StatusOK) })
	for _, path := range []string{"/media/42?q=private+search", "/setup/secret-token"} {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}
	out := buf.String()
	for _, leak := range []string{"private", "search", "/media/42", "secret-token"} {
		if strings.Contains(out, leak) {
			t.Errorf("log contains %q:\n%s", leak, out)
		}
	}
	if !strings.Contains(out, "GET     /media/:id") || !strings.Contains(out, "(no route)") {
		t.Errorf("log lacks the route pattern or the no-route marker:\n%s", out)
	}
}
