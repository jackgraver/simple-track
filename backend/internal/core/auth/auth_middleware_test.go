package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const testDevAuthToken = "dev-bypass-token-for-tests-only"

// setupUsers gives the test its own SQLite users table as the global DB.
func setupUsers(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	database.SetDB(db)
	t.Cleanup(func() { database.SetDB(nil) })
	return db
}

func createUser(t *testing.T, db *gorm.DB, username string, admin bool) models.User {
	t.Helper()
	user := models.User{Username: username, Password: "hash", Email: username + "@example.com", IsAdmin: admin, TokenVersion: models.InitialTokenVersion}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	return user
}

func tokenFor(t *testing.T, user models.User) string {
	t.Helper()
	tok, err := GenerateToken(user)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

type meBody struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	GinUser  string `json:"gin_user"`
	Error    string `json:"error"`
}

// serve runs one request through AuthMiddleware to a handler that echoes the context user.
func serve(t *testing.T, cookie string) (int, meBody) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", AuthMiddleware(), func(c *gin.Context) {
		u := MustUser(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{"id": u.ID, "username": u.Username, "admin": u.IsAdmin, "gin_user": c.GetString("username")})
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: AuthTokenCookieName, Value: cookie})
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var body meBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}
	return rec.Code, body
}

func expectRejected(t *testing.T, code int, body meBody) {
	t.Helper()
	if code != http.StatusUnauthorized {
		t.Fatalf("status %d want 401 (body %+v)", code, body)
	}
	if body.ID != 0 {
		t.Fatalf("rejected request reached the handler: %+v", body)
	}
}

func TestAuthMiddleware_validToken_putsUserOnContext(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	db := setupUsers(t)
	createUser(t, db, "first", false)
	wanda := createUser(t, db, "wanda", true)

	code, body := serve(t, tokenFor(t, wanda))
	if code != http.StatusOK {
		t.Fatalf("status %d body %+v", code, body)
	}
	if body.ID != wanda.ID || body.Username != "wanda" || !body.Admin || body.GinUser != "wanda" {
		t.Fatalf("context user: got %+v want id=%d wanda admin", body, wanda.ID)
	}
}

func TestAuthMiddleware_adminFlagComesFromTheRowNotTheToken(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	db := setupUsers(t)
	user := createUser(t, db, "wanda", true)
	tok := tokenFor(t, user)
	if err := db.Model(&user).Update("is_admin", false).Error; err != nil {
		t.Fatal(err)
	}
	code, body := serve(t, tok)
	if code != http.StatusOK || body.Admin {
		t.Fatalf("demoted admin: status %d body %+v, want 200 non-admin", code, body)
	}
}

func TestAuthMiddleware_missingCookie_unauthorized(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	setupUsers(t)
	code, body := serve(t, "")
	expectRejected(t, code, body)
}

func TestAuthMiddleware_authorizationBearerIgnored_withoutCookie(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	db := setupUsers(t)
	tok := tokenFor(t, createUser(t, db, "zeus", false))
	gin.SetMode(gin.TestMode)
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

func TestAuthMiddleware_invalidCookie_genericError(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	setupUsers(t)
	code, body := serve(t, "not-a-valid-jwt")
	expectRejected(t, code, body)
	if body.Error != "Invalid or expired session" {
		t.Fatalf("error body: got %q; the reason must not be echoed to the client", body.Error)
	}
}

func TestAuthMiddleware_tokenWithoutUID_rejected(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	db := setupUsers(t)
	createUser(t, db, "wanda", true)
	// The shape every token had before multi-user: a valid signature, username, no uid.
	now := time.Now().Unix()
	old := mustSignedJWT(t, Claims{Username: "wanda", Timestamp: now, Iat: now, Exp: now + 3600})
	code, body := serve(t, old)
	expectRejected(t, code, body)
}

func TestAuthMiddleware_tokenWithoutVersion_rejected(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	db := setupUsers(t)
	user := createUser(t, db, "wanda", true)
	now := time.Now().Unix()
	tok := mustSignedJWT(t, Claims{UID: user.ID, Username: "wanda", Timestamp: now, Iat: now, Exp: now + 3600})
	code, body := serve(t, tok)
	expectRejected(t, code, body)
}

func TestAuthMiddleware_revokedToken_rejected(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	db := setupUsers(t)
	user := createUser(t, db, "wanda", true)
	tok := tokenFor(t, user)
	if err := db.Model(&user).Update("token_version", gorm.Expr("token_version + 1")).Error; err != nil {
		t.Fatal(err)
	}
	code, body := serve(t, tok)
	expectRejected(t, code, body)

	// A token issued after the bump works.
	var fresh models.User
	if err := db.First(&fresh, user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if code, body := serve(t, tokenFor(t, fresh)); code != http.StatusOK || body.ID != user.ID {
		t.Fatalf("new token after revoke: status %d body %+v", code, body)
	}
}

func TestAuthMiddleware_disabledUser_rejected(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	db := setupUsers(t)
	user := createUser(t, db, "wanda", false)
	tok := tokenFor(t, user)
	if err := db.Model(&user).Update("disabled_at", time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	code, body := serve(t, tok)
	expectRejected(t, code, body)
	if strings.Contains(strings.ToLower(body.Error), "disabled") {
		t.Fatalf("error body leaks the reason: %q", body.Error)
	}
}

func TestAuthMiddleware_deletedUser_rejected(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	db := setupUsers(t)
	user := createUser(t, db, "wanda", false)
	tok := tokenFor(t, user)
	if err := db.Delete(&user).Error; err != nil {
		t.Fatal(err)
	}
	code, body := serve(t, tok)
	expectRejected(t, code, body)
}

func TestAuthMiddleware_uidPicksTheUser_notTheUsernameClaim(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	db := setupUsers(t)
	alice := createUser(t, db, "alice", false)
	tok := tokenFor(t, alice)
	if err := db.Model(&alice).Update("username", "alicia").Error; err != nil {
		t.Fatal(err)
	}
	code, body := serve(t, tok)
	if code != http.StatusOK || body.ID != alice.ID || body.Username != "alicia" {
		t.Fatalf("status %d body %+v", code, body)
	}
}

func TestAuthMiddleware_databaseError_isNot401(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	db := setupUsers(t)
	user := createUser(t, db, "wanda", false)
	tok := tokenFor(t, user)
	if err := db.Migrator().DropTable(&models.User{}); err != nil {
		t.Fatal(err)
	}
	code, body := serve(t, tok)
	if code != http.StatusServiceUnavailable || body.ID != 0 {
		t.Fatalf("DB failure: status %d body %+v, want 503 so the client keeps its session", code, body)
	}
}

func TestAuthMiddleware_devTokenDisabledByDefault_evenInDevEnv(t *testing.T) {
	t.Setenv("ALLOW_BYPASS", "")
	t.Setenv("DEV_AUTH_TOKEN", "")
	setupUsers(t)
	code, body := serve(t, "")
	expectRejected(t, code, body)
}

func TestAuthMiddleware_devToken_unsetAllowBypass_disabled(t *testing.T) {
	t.Setenv("ALLOW_BYPASS", "")
	t.Setenv("DEV_AUTH_TOKEN", testDevAuthToken)
	db := setupUsers(t)
	createUser(t, db, "dev", true)
	code, body := serve(t, testDevAuthToken)
	expectRejected(t, code, body)
}

func TestAuthMiddleware_devToken_resolvesDevUsersRealID(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("ALLOW_BYPASS", "true")
	t.Setenv("DEV_AUTH_TOKEN", testDevAuthToken)
	t.Setenv("DEV_AUTH_USER", "")
	db := setupUsers(t)
	createUser(t, db, "someone", false)
	dev := createUser(t, db, "dev", true)
	code, body := serve(t, testDevAuthToken)
	if code != http.StatusOK {
		t.Fatalf("status %d body %+v", code, body)
	}
	if body.ID != dev.ID || body.Username != "dev" || !body.Admin {
		t.Fatalf("dev user: got %+v want id=%d dev admin", body, dev.ID)
	}
}

func TestAuthMiddleware_devToken_missingDevUser_unauthorized(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("ALLOW_BYPASS", "true")
	t.Setenv("DEV_AUTH_TOKEN", testDevAuthToken)
	t.Setenv("DEV_AUTH_USER", "")
	setupUsers(t)
	code, body := serve(t, testDevAuthToken)
	expectRejected(t, code, body)
}

func TestAuthMiddleware_devTokenNeverAuthorizesInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ALLOW_BYPASS", "true")
	t.Setenv("DEV_AUTH_TOKEN", testDevAuthToken)
	db := setupUsers(t)
	createUser(t, db, "dev", true)
	code, body := serve(t, testDevAuthToken)
	expectRejected(t, code, body)
}

func TestAuthMiddleware_devToken_overrideUser(t *testing.T) {
	t.Setenv("ALLOW_BYPASS", "true")
	t.Setenv("DEV_AUTH_TOKEN", testDevAuthToken)
	t.Setenv("DEV_AUTH_USER", "alice")
	db := setupUsers(t)
	createUser(t, db, "dev", true)
	alice := createUser(t, db, "alice", false)
	code, body := serve(t, testDevAuthToken)
	if code != http.StatusOK || body.ID != alice.ID || body.Username != "alice" || body.Admin {
		t.Fatalf("status %d body %+v want alice id=%d", code, body, alice.ID)
	}
}

func TestAuthMiddleware_devToken_disabledDevUser_rejected(t *testing.T) {
	t.Setenv("ALLOW_BYPASS", "true")
	t.Setenv("DEV_AUTH_TOKEN", testDevAuthToken)
	t.Setenv("DEV_AUTH_USER", "")
	db := setupUsers(t)
	dev := createUser(t, db, "dev", true)
	if err := db.Model(&dev).Update("disabled_at", time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	code, body := serve(t, testDevAuthToken)
	expectRejected(t, code, body)
}

func TestAuthMiddleware_devToken_allowBypassFalse_disabled(t *testing.T) {
	t.Setenv("ALLOW_BYPASS", "false")
	t.Setenv("DEV_AUTH_TOKEN", testDevAuthToken)
	db := setupUsers(t)
	createUser(t, db, "dev", true)
	code, body := serve(t, testDevAuthToken)
	expectRejected(t, code, body)
}
