package controller

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/core/auth/services"
	"be-simpletracker/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupLoginRouteTest(t *testing.T, config LoginProtectionConfig) (*gin.Engine, *bytes.Buffer) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	database.SetDB(db)
	t.Cleanup(func() {
		database.SetDB(nil)
	})

	service := services.NewAuthService(func(string) (string, error) {
		return "server-only-token", nil
	})
	if _, err := service.Register(services.RegisterInput{
		Username: "alice",
		Password: "correct-password",
	}); err != nil {
		t.Fatal(err)
	}

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	protection := NewLoginProtection(config, logger)
	cookie := CookieConfig{
		Name:     "auth_token",
		MaxAge:   3600,
		SameSite: http.SameSiteLaxMode,
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", func(c *gin.Context) {
		Login(c, service, cookie, protection)
	})
	return router, &logs
}

func performLogin(router *gin.Engine, username, password string) *httptest.ResponseRecorder {
	body := `{"username":"` + username + `","password":"` + password + `"}`
	request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestLoginKeepsTokenOutOfResponseAndLogsSuccess(t *testing.T) {
	router, logs := setupLoginRouteTest(t, LoginProtectionConfig{
		Window:         time.Minute,
		MaxPerIP:       10,
		MaxPerUsername: 10,
	})
	recorder := performLogin(router, "alice", "correct-password")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "server-only-token") || strings.Contains(recorder.Body.String(), `"token"`) {
		t.Fatalf("login response exposed token: %s", recorder.Body.String())
	}
	if recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected cache control %q", recorder.Header().Get("Cache-Control"))
	}
	if cookie := recorder.Header().Get("Set-Cookie"); !strings.Contains(cookie, "HttpOnly") {
		t.Fatalf("auth cookie is not HttpOnly: %s", cookie)
	}
	if logged := logs.String(); !strings.Contains(logged, `"outcome":"success"`) || strings.Contains(logged, "correct-password") {
		t.Fatalf("unexpected audit log: %s", logged)
	}
}

func TestLoginRateLimitLogsFailureAndReturnsRetryAfter(t *testing.T) {
	router, logs := setupLoginRouteTest(t, LoginProtectionConfig{
		Window:         time.Minute,
		MaxPerIP:       10,
		MaxPerUsername: 1,
	})
	if recorder := performLogin(router, "alice", "wrong-password"); recorder.Code != http.StatusUnauthorized {
		t.Fatalf("first status %d body %s", recorder.Code, recorder.Body.String())
	}
	recorder := performLogin(router, "alice", "wrong-password")
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("second status %d body %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("rate-limited response is missing Retry-After")
	}
	logged := logs.String()
	if !strings.Contains(logged, `"outcome":"invalid_credentials"`) || !strings.Contains(logged, `"outcome":"rate_limited"`) {
		t.Fatalf("missing login audit events: %s", logged)
	}
	if strings.Contains(logged, "wrong-password") {
		t.Fatalf("password appeared in login audit log: %s", logged)
	}
}
