package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"database/sql"

	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/core/auth/services"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/migrate"
	"be-simpletracker/migrations"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// accountsEnv is a fully migrated Postgres database with the auth routes served the way
// cmd/server serves them, on a pool that acts as simpletrack_app, so row-level security applies
// exactly as in production. owner bypasses RLS, for setup and for checking what was written.
type accountsEnv struct {
	t      *testing.T
	owner  *gorm.DB
	app    *gorm.DB
	router *gin.Engine
}

func newAccountsEnv(t *testing.T) *accountsEnv {
	t.Helper()
	admin := os.Getenv("TEST_DATABASE_URL")
	if admin == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	adminDB, err := sql.Open("pgx", admin)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { adminDB.Close() })
	if err := adminDB.Ping(); err != nil {
		t.Skipf("TEST_DATABASE_URL unreachable: %v", err)
	}
	name := fmt.Sprintf("accounts_test_%d", rand.Uint32())
	if _, err := adminDB.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(admin)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	sqlDB, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatal(err)
	}
	m, err := migrate.New(sqlDB, migrations.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	owner, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), cfg)
	if err != nil {
		t.Fatal(err)
	}
	// Every connection of this pool starts as simpletrack_app (the role is NOLOGIN in test
	// clusters, but the test's superuser may assume it at connect time).
	q := u.Query()
	q.Set("options", "-c role="+database.AppRole)
	u.RawQuery = q.Encode()
	app, err := gorm.Open(postgres.Open(u.String()), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if d, err := app.DB(); err == nil {
			d.Close()
		}
		sqlDB.Close()
		adminDB.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)")
	})
	var role string
	if err := app.Raw("SELECT current_user").Scan(&role).Error; err != nil || role != database.AppRole {
		t.Fatalf("app pool runs as %q (%v), want %s", role, err, database.AppRole)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery(), database.Middleware(app))
	RegisterRoutes(r, app)
	return &accountsEnv{t: t, owner: owner, app: app, router: r}
}

// seedUser creates a user with a password directly (as the owner).
func (e *accountsEnv) seedUser(username, password string, admin bool) models.User {
	e.t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		e.t.Fatal(err)
	}
	user := models.User{Username: username, Password: string(hash), IsAdmin: admin, TokenVersion: models.InitialTokenVersion}
	if err := e.owner.Omit("Email").Create(&user).Error; err != nil {
		e.t.Fatal(err)
	}
	return user
}

type resp struct {
	Code   int
	Body   map[string]any
	Raw    string
	Cookie string // auth cookie set by the response, if any
}

func (e *accountsEnv) do(method, path, cookie string, body any) resp {
	e.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			e.t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: AuthTokenCookieName, Value: cookie})
	}
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	out := resp{Code: rec.Code, Raw: rec.Body.String()}
	_ = json.Unmarshal(rec.Body.Bytes(), &out.Body)
	for _, c := range rec.Result().Cookies() {
		if c.Name == AuthTokenCookieName && c.Value != "" {
			out.Cookie = c.Value
		}
	}
	return out
}

func (e *accountsEnv) expect(r resp, code int, what string) resp {
	e.t.Helper()
	if r.Code != code {
		e.t.Fatalf("%s: status %d (%s), want %d", what, r.Code, r.Raw, code)
	}
	return r
}

func (e *accountsEnv) login(username, password string) resp {
	e.t.Helper()
	return e.do(http.MethodPost, "/auth/login", "", map[string]string{"username": username, "password": password})
}

// events returns "event:detail" for every audit row about userID (0: rows with no user), oldest first.
func (e *accountsEnv) events(userID uint) []string {
	e.t.Helper()
	q := e.owner.Model(&models.AuthEvent{}).Order("id")
	if userID == 0 {
		q = q.Where("user_id IS NULL")
	} else {
		q = q.Where("user_id = ?", userID)
	}
	var rows []models.AuthEvent
	if err := q.Find(&rows).Error; err != nil {
		e.t.Fatal(err)
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Event + ":" + r.Detail
	}
	return out
}

func tokenOf(t *testing.T, r resp) string {
	t.Helper()
	link, _ := r.Body["link"].(map[string]any)
	path, _ := link["path"].(string)
	if !strings.HasPrefix(path, "/setup/") {
		t.Fatalf("no setup link in %s", r.Raw)
	}
	return strings.TrimPrefix(path, "/setup/")
}

func userIDOf(t *testing.T, r resp) uint {
	t.Helper()
	user, _ := r.Body["user"].(map[string]any)
	id, _ := user["ID"].(float64)
	if id == 0 {
		t.Fatalf("no user id in %s", r.Raw)
	}
	return uint(id)
}

// The C6 flow end to end: create -> link -> set password -> signed in -> link dead.
func TestAccounts_createLinkSetPasswordSignIn(t *testing.T) {
	e := newAccountsEnv(t)
	e.seedUser("owner", "owner-password", true)
	adminCookie := e.expect(e.login("owner", "owner-password"), http.StatusOK, "admin login").Cookie

	created := e.expect(e.do(http.MethodPost, "/admin/users", adminCookie, map[string]string{"username": "  Bob "}), http.StatusCreated, "create user")
	bobID := userIDOf(t, created)
	token := tokenOf(t, created)
	if user := created.Body["user"].(map[string]any); user["username"] != "bob" {
		t.Fatalf("username stored as %v, want bob", user["username"])
	}
	if link := created.Body["link"].(map[string]any); link["purpose"] != "setup" {
		t.Fatalf("purpose %v, want setup", link["purpose"])
	}
	var stored models.SetupToken
	e.owner.Where("user_id = ?", bobID).First(&stored)
	if stored.TokenHash == token || strings.Contains(stored.TokenHash, token) || time.Until(stored.ExpiresAt) < 47*time.Hour {
		t.Fatalf("stored token row %+v: want a hash, 48 h expiry", stored)
	}

	// No password yet: can't sign in.
	e.expect(e.login("bob", ""), http.StatusBadRequest, "login with empty password")
	e.expect(e.login("bob", "anything-at-all"), http.StatusUnauthorized, "login before setup")

	info := e.expect(e.do(http.MethodPost, "/auth/setup/inspect", "", map[string]string{"token": token}), http.StatusOK, "inspect")
	if info.Body["username"] != "bob" || info.Body["purpose"] != "setup" {
		t.Fatalf("inspect: %s", info.Raw)
	}
	e.expect(e.do(http.MethodPost, "/auth/setup", "", map[string]string{"token": token, "password": "short"}), http.StatusBadRequest, "short password")
	e.expect(e.do(http.MethodPost, "/auth/setup", "", map[string]string{"token": token, "password": strings.Repeat("é", 37)}), http.StatusBadRequest, "74-byte password")

	done := e.expect(e.do(http.MethodPost, "/auth/setup", "", map[string]string{"token": token, "password": "bob-password-1"}), http.StatusOK, "setup")
	if done.Cookie == "" {
		t.Fatal("setup didn't sign bob in")
	}
	me := e.expect(e.do(http.MethodGet, "/auth/me", done.Cookie, nil), http.StatusOK, "me after setup")
	if me.Body["username"] != "bob" {
		t.Fatalf("me: %s", me.Raw)
	}

	// The link is dead now, and the rejected attempt is audited although the request rolled back.
	e.expect(e.do(http.MethodPost, "/auth/setup/inspect", "", map[string]string{"token": token}), http.StatusGone, "inspect used link")
	e.expect(e.do(http.MethodPost, "/auth/setup", "", map[string]string{"token": token, "password": "bob-password-2"}), http.StatusGone, "reuse link")
	e.expect(e.do(http.MethodPost, "/auth/setup", "", map[string]string{"token": "no-such-token", "password": "bob-password-2"}), http.StatusGone, "unknown link")

	// Case-insensitive username at login.
	e.expect(e.login("BOB", "bob-password-1"), http.StatusOK, "login as BOB")
	e.expect(e.login("bob", "bob-password-2"), http.StatusUnauthorized, "login with rejected password")

	want := []string{"user_created:", "setup_link_created:setup", "login_failed:no_password", "setup_link_used:setup", "setup_link_rejected:used", "login_ok:", "login_failed:wrong_password"}
	if got := e.events(bobID); strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("bob's audit events:\n got %v\nwant %v", got, want)
	}
	if got := e.events(0); strings.Join(got, " ") != "setup_link_rejected:unknown" {
		t.Fatalf("audit events without a user: %v", got)
	}
}

func TestAccounts_newLinkRevokesOlder_resetSignsOutEverywhere(t *testing.T) {
	e := newAccountsEnv(t)
	e.seedUser("owner", "owner-password", true)
	bob := e.seedUser("bob", "bob-password-1", false)
	adminCookie := e.expect(e.login("owner", "owner-password"), http.StatusOK, "admin login").Cookie
	bobCookie := e.expect(e.login("bob", "bob-password-1"), http.StatusOK, "bob login").Cookie

	path := fmt.Sprintf("/admin/users/%d/setup-link", bob.ID)
	first := tokenOf(t, e.expect(e.do(http.MethodPost, path, adminCookie, nil), http.StatusCreated, "first link"))
	secondResp := e.expect(e.do(http.MethodPost, path, adminCookie, nil), http.StatusCreated, "second link")
	if secondResp.Body["link"].(map[string]any)["purpose"] != "reset" {
		t.Fatalf("link for a user with a password: %s, want purpose reset", secondResp.Raw)
	}
	second := tokenOf(t, secondResp)

	e.expect(e.do(http.MethodPost, "/auth/setup", "", map[string]string{"token": first, "password": "bob-password-2"}), http.StatusGone, "revoked link")
	e.expect(e.do(http.MethodPost, "/auth/setup", "", map[string]string{"token": second, "password": "bob-password-2"}), http.StatusOK, "latest link")
	e.expect(e.do(http.MethodGet, "/auth/me", bobCookie, nil), http.StatusUnauthorized, "bob's old session after reset")
	e.expect(e.login("bob", "bob-password-2"), http.StatusOK, "login with new password")

	// Expired links are dead too.
	third := tokenOf(t, e.expect(e.do(http.MethodPost, path, adminCookie, nil), http.StatusCreated, "third link"))
	e.owner.Model(&models.SetupToken{}).Where("token_hash = ?", hashOf(third)).Update("expires_at", time.Now().Add(-time.Minute))
	e.expect(e.do(http.MethodPost, "/auth/setup", "", map[string]string{"token": third, "password": "bob-password-3"}), http.StatusGone, "expired link")
	events := strings.Join(e.events(bob.ID), " ")
	for _, want := range []string{"setup_link_rejected:revoked", "setup_link_rejected:expired", "setup_link_created:reset"} {
		if !strings.Contains(events, want) {
			t.Errorf("audit events %q lack %s", events, want)
		}
	}
}

func hashOf(token string) string { return services.HashSetupToken(token) }

func TestAccounts_adminOnly(t *testing.T) {
	e := newAccountsEnv(t)
	e.seedUser("owner", "owner-password", true)
	bob := e.seedUser("bob", "bob-password-1", false)
	bobCookie := e.expect(e.login("bob", "bob-password-1"), http.StatusOK, "bob login").Cookie
	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/admin/users"},
		{http.MethodPost, "/admin/users"},
		{http.MethodPost, fmt.Sprintf("/admin/users/%d/setup-link", bob.ID)},
		{http.MethodPost, fmt.Sprintf("/admin/users/%d/disable", bob.ID)},
		{http.MethodPost, fmt.Sprintf("/admin/users/%d/enable", bob.ID)},
		{http.MethodGet, "/admin/auth-events"},
	} {
		e.expect(e.do(c.method, c.path, bobCookie, map[string]string{"username": "eve"}), http.StatusForbidden, "bob "+c.method+" "+c.path)
		e.expect(e.do(c.method, c.path, "", nil), http.StatusUnauthorized, "anonymous "+c.method+" "+c.path)
	}
	var n int64
	e.owner.Model(&models.User{}).Where("username = 'eve'").Count(&n)
	if n != 0 {
		t.Fatal("a non-admin created a user")
	}
}

func TestAccounts_disableEnable(t *testing.T) {
	e := newAccountsEnv(t)
	owner := e.seedUser("owner", "owner-password", true)
	bob := e.seedUser("bob", "bob-password-1", false)
	adminCookie := e.expect(e.login("owner", "owner-password"), http.StatusOK, "admin login").Cookie
	bobCookie := e.expect(e.login("bob", "bob-password-1"), http.StatusOK, "bob login").Cookie
	link := tokenOf(t, e.expect(e.do(http.MethodPost, fmt.Sprintf("/admin/users/%d/setup-link", bob.ID), adminCookie, nil), http.StatusCreated, "link"))

	e.expect(e.do(http.MethodPost, fmt.Sprintf("/admin/users/%d/disable", owner.ID), adminCookie, nil), http.StatusBadRequest, "disable self")
	e.expect(e.do(http.MethodPost, fmt.Sprintf("/admin/users/%d/disable", bob.ID), adminCookie, nil), http.StatusOK, "disable bob")
	e.expect(e.do(http.MethodGet, "/auth/me", bobCookie, nil), http.StatusUnauthorized, "disabled bob's session")
	e.expect(e.login("bob", "bob-password-1"), http.StatusUnauthorized, "disabled bob's login")
	e.expect(e.do(http.MethodPost, "/auth/setup", "", map[string]string{"token": link, "password": "bob-password-2"}), http.StatusGone, "link revoked by disable")
	e.expect(e.do(http.MethodPost, fmt.Sprintf("/admin/users/%d/setup-link", bob.ID), adminCookie, nil), http.StatusConflict, "link for disabled user")

	e.expect(e.do(http.MethodPost, fmt.Sprintf("/admin/users/%d/enable", bob.ID), adminCookie, nil), http.StatusOK, "enable bob")
	e.expect(e.do(http.MethodGet, "/auth/me", bobCookie, nil), http.StatusUnauthorized, "session from before the disable")
	e.expect(e.login("bob", "bob-password-1"), http.StatusOK, "bob signs in again")
	e.expect(e.do(http.MethodPost, "/admin/users/9999/disable", adminCookie, nil), http.StatusNotFound, "disable unknown user")

	got := strings.Join(e.events(bob.ID), " ")
	want := "login_ok: setup_link_created:reset user_disabled: login_failed:disabled setup_link_rejected:revoked user_enabled: login_ok:"
	if got != want {
		t.Fatalf("bob's events:\n got %s\nwant %s", got, want)
	}

	users := e.expect(e.do(http.MethodGet, "/admin/users", adminCookie, nil), http.StatusOK, "list users")
	list := users.Body["users"].([]any)
	if len(list) != 2 || list[1].(map[string]any)["username"] != "bob" || list[1].(map[string]any)["last_login_at"] == nil {
		t.Fatalf("users: %s", users.Raw)
	}
}

func TestAccounts_usernameRules(t *testing.T) {
	e := newAccountsEnv(t)
	e.seedUser("owner", "owner-password", true)
	e.seedUser("Jack", "jack-password", false) // a pre-existing mixed-case name
	adminCookie := e.expect(e.login("owner", "owner-password"), http.StatusOK, "admin login").Cookie
	for _, name := range []string{"jack", "JACK"} {
		e.expect(e.do(http.MethodPost, "/admin/users", adminCookie, map[string]string{"username": name}), http.StatusConflict, "duplicate "+name)
	}
	for _, name := range []string{"a", "has space", "émile", "-dash", strings.Repeat("x", 33), "semi;colon"} {
		e.expect(e.do(http.MethodPost, "/admin/users", adminCookie, map[string]string{"username": name}), http.StatusBadRequest, "invalid "+name)
	}
	e.expect(e.login("jack", "jack-password"), http.StatusOK, "pre-existing Jack logs in as jack")
	// Two username-only accounts don't collide on the (empty) email.
	e.expect(e.do(http.MethodPost, "/admin/users", adminCookie, map[string]string{"username": "mia"}), http.StatusCreated, "mia")
	e.expect(e.do(http.MethodPost, "/admin/users", adminCookie, map[string]string{"username": "leo.k"}), http.StatusCreated, "leo.k")
}

func TestAccounts_changeOwnPassword(t *testing.T) {
	e := newAccountsEnv(t)
	bob := e.seedUser("bob", "bob-password-1", false)
	other := e.expect(e.login("bob", "bob-password-1"), http.StatusOK, "other session").Cookie
	cookie := e.expect(e.login("bob", "bob-password-1"), http.StatusOK, "this session").Cookie

	change := func(current, next string) resp {
		return e.do(http.MethodPost, "/auth/password", cookie, map[string]string{"current_password": current, "new_password": next})
	}
	e.expect(change("wrong-password", "bob-password-2"), http.StatusBadRequest, "wrong current password")
	e.expect(change("bob-password-1", "short"), http.StatusBadRequest, "short new password")
	e.expect(change("bob-password-1", "bob-password-1"), http.StatusBadRequest, "same password")
	ok := e.expect(change("bob-password-1", "bob-password-2"), http.StatusOK, "change")
	if ok.Cookie == "" {
		t.Fatal("no fresh cookie for the current session")
	}
	e.expect(e.do(http.MethodGet, "/auth/me", ok.Cookie, nil), http.StatusOK, "current session keeps working")
	e.expect(e.do(http.MethodGet, "/auth/me", other, nil), http.StatusUnauthorized, "other session ends")
	e.expect(e.login("bob", "bob-password-2"), http.StatusOK, "login with new password")

	got := strings.Join(e.events(bob.ID), " ")
	want := "login_ok: login_ok: password_change_failed:wrong_password password_changed: login_ok:"
	if got != want {
		t.Fatalf("events:\n got %s\nwant %s", got, want)
	}
}

func TestAccounts_adminEventView_noAttemptedUsernames(t *testing.T) {
	e := newAccountsEnv(t)
	e.seedUser("owner", "owner-password", true)
	e.expect(e.login("my-secret-password", "oops"), http.StatusUnauthorized, "password typed as username")
	adminCookie := e.expect(e.login("owner", "owner-password"), http.StatusOK, "admin login").Cookie
	r := e.expect(e.do(http.MethodGet, "/admin/auth-events?limit=10", adminCookie, nil), http.StatusOK, "events")
	if strings.Contains(r.Raw, "my-secret-password") {
		t.Fatalf("attempted username stored: %s", r.Raw)
	}
	events := r.Body["events"].([]any)
	if len(events) != 2 || events[0].(map[string]any)["username"] != "owner" || events[1].(map[string]any)["detail"] != "unknown_user" {
		t.Fatalf("events: %s", r.Raw)
	}
	older := e.expect(e.do(http.MethodGet, fmt.Sprintf("/admin/auth-events?before=%v", events[0].(map[string]any)["id"]), adminCookie, nil), http.StatusOK, "older")
	if n := len(older.Body["events"].([]any)); n != 1 {
		t.Fatalf("page before the newest: %d events", n)
	}
}

// Row-level security on C6's tables, as the app role.
func TestAccounts_rowLevelSecurity(t *testing.T) {
	e := newAccountsEnv(t)
	admin := e.seedUser("owner", "owner-password", true)
	bob := e.seedUser("bob", "bob-password-1", false)
	adminCookie := e.expect(e.login("owner", "owner-password"), http.StatusOK, "admin login").Cookie
	token := tokenOf(t, e.expect(e.do(http.MethodPost, fmt.Sprintf("/admin/users/%d/setup-link", bob.ID), adminCookie, nil), http.StatusCreated, "link"))

	as := func(id database.Identity, setupHash string, fn func(db *gorm.DB) error) error {
		ctx, req := database.WithTx(context.Background(), e.app)
		defer req.Rollback()
		if err := database.SetIdentity(ctx, id); err != nil {
			return err
		}
		db := database.From(ctx)
		if setupHash != "" {
			if err := db.Exec("SELECT set_config('app.setup_token_hash', ?, true)", setupHash).Error; err != nil {
				return err
			}
		}
		return fn(db)
	}
	count := func(id database.Identity, hash, table string) int64 {
		t.Helper()
		var n int64
		if err := as(id, hash, func(db *gorm.DB) error { return db.Table(table).Count(&n).Error }); err != nil {
			t.Fatal(err)
		}
		return n
	}
	anon := database.Identity{}
	asBob := database.Identity{UserID: bob.ID}
	asAdmin := database.Identity{UserID: admin.ID, IsAdmin: true}

	if n := count(asAdmin, "", "auth_events"); n == 0 {
		t.Fatal("admin sees no auth events")
	}
	if n := count(asBob, "", "auth_events") + count(anon, "", "auth_events"); n != 0 {
		t.Fatalf("non-admins see %d auth events", n)
	}
	if n := count(asAdmin, "", "user_setup_tokens"); n != 1 {
		t.Fatalf("admin sees %d setup tokens, want 1", n)
	}
	if n := count(asBob, "", "user_setup_tokens") + count(anon, "", "user_setup_tokens") + count(anon, hashOf("wrong"), "user_setup_tokens"); n != 0 {
		t.Fatalf("setup tokens visible without their hash: %d", n)
	}
	if n := count(anon, hashOf(token), "user_setup_tokens"); n != 1 {
		t.Fatalf("token holder sees %d rows, want 1", n)
	}
	// The audit log is append-only for the app, whoever it is.
	for _, stmt := range []string{"UPDATE auth_events SET detail = 'x'", "DELETE FROM auth_events", "TRUNCATE auth_events"} {
		if err := as(asAdmin, "", func(db *gorm.DB) error { return db.Exec(stmt).Error }); err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("%s as admin app: %v, want permission denied", stmt, err)
		}
	}
	// Anyone may insert (failed logins have no identity), but not read back.
	if err := as(anon, "", func(db *gorm.DB) error {
		return db.Exec("INSERT INTO auth_events (event) VALUES ('test')").Error
	}); err != nil {
		t.Fatalf("anonymous insert: %v", err)
	}
	// Bob can't mint a token for himself.
	err := as(asBob, "", func(db *gorm.DB) error {
		return db.Exec("INSERT INTO user_setup_tokens (user_id, token_hash, purpose, expires_at) VALUES (?, 'h', 'reset', now() + interval '1 day')", bob.ID).Error
	})
	if err == nil || !strings.Contains(err.Error(), "row-level security") {
		t.Fatalf("bob inserting a setup token: %v, want an RLS violation", err)
	}
}
