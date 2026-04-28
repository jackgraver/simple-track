package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware_validCookie_setsUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tok, err := GenerateToken("wanda")
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/x", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"u": GetUsername(c)})
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.AddCookie(&http.Cookie{Name: AuthTokenCookieName, Value: tok})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestAuthMiddleware_missingCookie_unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", AuthMiddleware(), func(c *gin.Context) {})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_authorizationBearerIgnored_withoutCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tok, err := GenerateToken("zeus")
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/", AuthMiddleware(), func(c *gin.Context) {})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Bearer without cookie should be rejected; got %d", rec.Code)
	}
}
