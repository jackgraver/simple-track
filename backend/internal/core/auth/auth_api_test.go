package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRouteIsNeverEnabledInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("REGISTER_ENABLED", "true")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/auth/register", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", recorder.Code)
	}
}
