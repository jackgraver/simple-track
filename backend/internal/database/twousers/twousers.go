// Package twousers is the two-user Postgres test harness for ownership tests
// (docs/multi-user/README.md "Rules for every lane": Alice's data, Bob's token).
//
//	h := twousers.New(t)                       // skipped without TEST_DATABASE_URL (fails in CI)
//	api := h.Router(func(g *gin.RouterGroup) { water.RegisterWaterRoutes(g.Group("/water")) })
//	res := api.Do(h.Alice, "POST", "/water/presets", map[string]any{"name": "Cup", "amount_oz": 8})
//	res.Status(t, 200)
//	api.Do(h.Bob, "GET", "/water/presets", nil)   // Bob sees an empty list
//
// Each harness is a fresh, fully migrated database. Requests go through the same chain as the
// server (database.Middleware, then auth.AuthMiddleware) on a pool that runs as the app role, so
// row-level security applies. Admin is the first user (the account migrations backfill to);
// Alice and Bob are regular users.
//
// New sets JWT_SECRET with t.Setenv, so tests using it can't call t.Parallel.
package twousers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"be-simpletracker/internal/core/auth"
	authm "be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/pgtest"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// User is one of the harness's accounts.
type User struct {
	ID       uint
	Username string
	IsAdmin  bool
	token    string
}

// Anonymous is no user: requests go without a session cookie.
var Anonymous = User{}

// Harness is a migrated database with three users.
type Harness struct {
	t testing.TB
	// DSN of the test database (owner role).
	DSN string
	// Owner is the schema owner. It bypasses row-level security: use it to arrange rows (setting
	// user_id yourself, as seeds do) and to check what really is in a table.
	Owner *gorm.DB
	// App runs as the server's role (simpletrack_app), so row-level security applies.
	App *gorm.DB

	Admin, Alice, Bob User
}

// New creates the database and the users admin, alice and bob.
func New(t testing.TB) *Harness {
	t.Helper()
	dsn := pgtest.Migrated(t, "twousers")
	h := &Harness{t: t, DSN: dsn, Owner: pgtest.Open(t, dsn), App: pgtest.OpenAsApp(t, dsn)}
	t.Setenv("JWT_SECRET", "twousers-test-secret-0123456789abcdef")
	h.Admin = h.NewUser("admin", true)
	h.Alice = h.NewUser("alice", false)
	h.Bob = h.NewUser("bob", false)
	return h
}

// NewUser adds another account (as the owner) and signs a session token for it.
func (h *Harness) NewUser(username string, admin bool) User {
	h.t.Helper()
	row := authm.User{
		Username:     username,
		Password:     "unused", // sessions come from tokens; nobody logs in with a password here
		Email:        username + "@example.test",
		IsAdmin:      admin,
		TokenVersion: authm.InitialTokenVersion,
	}
	if err := h.Owner.Create(&row).Error; err != nil {
		h.t.Fatalf("create user %s: %v", username, err)
	}
	tok, err := auth.GenerateToken(row)
	if err != nil {
		h.t.Fatal(err)
	}
	return User{ID: row.ID, Username: username, IsAdmin: admin, token: tok}
}

// Ctx returns a request-shaped context for u: a lazy transaction on the app pool with u as its
// identity (app.user_id, app.is_admin) and as auth.CurrentUser, plus the function that ends it
// (commit=true commits, false rolls back). Prefer Run.
func (h *Harness) Ctx(u User) (context.Context, func(commit bool) error) {
	h.t.Helper()
	ctx, req := database.WithTx(context.Background(), h.App)
	if u.ID != 0 {
		ctx = auth.WithUser(ctx, auth.User{ID: u.ID, Username: u.Username, IsAdmin: u.IsAdmin})
		if err := database.SetIdentity(ctx, database.Identity{UserID: u.ID, IsAdmin: u.IsAdmin}); err != nil {
			h.t.Fatal(err)
		}
	}
	return ctx, func(commit bool) error {
		if commit {
			return req.Commit()
		}
		return req.Rollback()
	}
}

// Run calls fn the way a request by u would run it (service and repository tests): in its own
// request transaction as u, committed when fn returns nil and rolled back otherwise.
func (h *Harness) Run(u User, fn func(ctx context.Context) error) error {
	h.t.Helper()
	ctx, end := h.Ctx(u)
	if err := fn(ctx); err != nil {
		_ = end(false)
		return err
	}
	return end(true)
}

// Router builds an API like cmd/server's: database.Middleware on the app pool, then
// auth.AuthMiddleware, then the routes register adds to the group (mounted at "/").
func (h *Harness) Router(register func(g *gin.RouterGroup)) *API {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery(), database.Middleware(h.App))
	register(r.Group("/", auth.AuthMiddleware(h.App)))
	return &API{t: h.t, engine: r}
}

// Engine is Router for modules that register on the engine itself and take the auth middleware
// as an argument (workout.RegisterRoutes): same stack, the routes keep their own paths.
func (h *Harness) Engine(register func(r *gin.Engine, authMiddleware gin.HandlerFunc)) *API {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery(), database.Middleware(h.App))
	register(r, auth.AuthMiddleware(h.App))
	return &API{t: h.t, engine: r}
}

// API sends requests to a Router.
type API struct {
	t      testing.TB
	engine *gin.Engine
}

// Response is a recorded response.
type Response struct {
	Code int
	Body []byte
	req  string
}

// Do sends method path as u (Anonymous: no cookie). body, if not nil, is sent as JSON (a string
// or []byte is sent as is).
func (a *API) Do(u User, method, path string, body any) *Response {
	a.t.Helper()
	var r io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		r = bytes.NewBufferString(b)
	case []byte:
		r = bytes.NewReader(b)
	default:
		buf, err := json.Marshal(b)
		if err != nil {
			a.t.Fatal(err)
		}
		r = bytes.NewReader(buf)
	}
	req := httptest.NewRequest(method, path, r)
	if r != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if u.token != "" {
		req.AddCookie(&http.Cookie{Name: auth.AuthTokenCookieName, Value: u.token})
	}
	rec := httptest.NewRecorder()
	a.engine.ServeHTTP(rec, req)
	who := u.Username
	if who == "" {
		who = "anonymous"
	}
	return &Response{Code: rec.Code, Body: rec.Body.Bytes(), req: fmt.Sprintf("%s %s as %s", method, path, who)}
}

// Status fails the test unless the response has the given status.
func (r *Response) Status(t testing.TB, want int) *Response {
	t.Helper()
	if r.Code != want {
		t.Fatalf("%s: status %d, want %d; body %s", r.req, r.Code, want, r.Body)
	}
	return r
}

// JSON decodes the body into v.
func (r *Response) JSON(t testing.TB, v any) {
	t.Helper()
	if err := json.Unmarshal(r.Body, v); err != nil {
		t.Fatalf("%s: decode %s: %v", r.req, r.Body, err)
	}
}

// NotFound fails the test unless the response is the "not found or not yours" 404, the same
// body whether the row is missing or someone else's.
func (r *Response) NotFound(t testing.TB) {
	t.Helper()
	r.Status(t, http.StatusNotFound)
	var body struct{ Error string }
	r.JSON(t, &body)
	if body.Error != "not found" {
		t.Fatalf("%s: 404 body %s, want {\"error\":\"not found\"}", r.req, r.Body)
	}
}
