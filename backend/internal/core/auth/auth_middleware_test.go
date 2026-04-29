package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const testDevAuthToken = "dev-bypass-token-for-tests-only"

func TestAuthMiddleware_validCookie_setsUsername(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	gin.SetMode(gin.TestMode)
	tok, err := GenerateToken("wanda")
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/x", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"u": c.GetString("username")})
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
	t.Setenv("APP_ENV", "prod")
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
	t.Setenv("APP_ENV", "prod")
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

func TestAuthMiddleware_invalidCookie_returnsInvalidTokenError(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", AuthMiddleware(), func(c *gin.Context) {})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: AuthTokenCookieName, Value: "not-a-valid-jwt"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d want %d", rec.Code, http.StatusUnauthorized)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !strings.HasPrefix(body.Error, "Invalid token: ") {
		t.Fatalf("error body: got %q want prefix %q", body.Error, "Invalid token: ")
	}
}

func TestAuthMiddleware_devTokenDisabledByDefault_evenInDevEnv(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DEV_AUTH_TOKEN", "")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", AuthMiddleware(), func(c *gin.Context) {})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("dev env without DEV_AUTH_TOKEN must NOT bypass: got %d want 401", rec.Code)
	}
}

func TestAuthMiddleware_devToken_unsetAppEnv_isTreatedAsProd(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("DEV_AUTH_TOKEN", testDevAuthToken)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", AuthMiddleware(), func(c *gin.Context) {})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: AuthTokenCookieName, Value: testDevAuthToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unset APP_ENV must be treated as prod: got %d want 401", rec.Code)
	}
}

func TestAuthMiddleware_devToken_cookieAuthorizes(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DEV_AUTH_TOKEN", testDevAuthToken)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"u": c.GetString("username")})
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.AddCookie(&http.Cookie{Name: AuthTokenCookieName, Value: testDevAuthToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var body struct {
		U string `json:"u"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.U != "dev" {
		t.Fatalf("dev user: got %q want %q", body.U, "dev")
	}
}

func TestAuthMiddleware_devToken_overrideUser(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DEV_AUTH_TOKEN", testDevAuthToken)
	t.Setenv("DEV_AUTH_USER", "alice")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"u": c.GetString("username")})
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.AddCookie(&http.Cookie{Name: AuthTokenCookieName, Value: testDevAuthToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var body struct {
		U string `json:"u"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.U != "alice" {
		t.Fatalf("dev user: got %q want %q", body.U, "alice")
	}
}

func TestAuthMiddleware_devToken_refusesInProductionEnv(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("DEV_AUTH_TOKEN", testDevAuthToken)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", AuthMiddleware(), func(c *gin.Context) {})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: AuthTokenCookieName, Value: testDevAuthToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("APP_ENV=prod must disable bypass: got %d want 401", rec.Code)
	}
}
