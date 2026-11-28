package auth

import (
	"context"
	"errors"
	"log"
	"net/http"

	"be-simpletracker/internal/core/auth/models"
	authrepo "be-simpletracker/internal/core/auth/repository"
	"be-simpletracker/internal/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Reasons a session is rejected. They wrap ErrInvalidToken and are logged, never returned to
// the client: every rejection answers 401 with the same generic body.
var (
	errNoUID           = invalidToken("token has no uid (issued before multi-user)")
	errUserGone        = invalidToken("user no longer exists")
	errUserDisabled    = invalidToken("user is disabled")
	errTokenRevoked    = invalidToken("token version does not match the user's (revoked)")
	errDevUserNotFound = invalidToken("DEV_AUTH_USER not found (run cmd/devuser)")
)

// AuthMiddleware validates the JWT stored in the httpOnly auth_token cookie from the SPA and
// puts the signed-in user on the request context (read it with UserID / MustUser).
//
// Every request loads the user row by primary key, so a disabled user, a bumped
// token_version or a removed admin flag takes effect on the next request. Missing or
// invalid tokens always produce 401; there is no default user.
//
// The user lookup runs on db (the pool), outside the request transaction, so that transaction
// still begins lazily: a handler that does external work before touching the database (media
// lookups) doesn't hold a connection meanwhile. The user then becomes the transaction's
// identity (app.user_id, app.is_admin) through database.SetIdentity.
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		lookup := database.WithDB(c.Request.Context(), db)
		if crossSiteWrite(c.Request) {
			log.Printf("[auth] refused %s %s: Sec-Fetch-Site=%q", c.Request.Method, c.Request.URL.Path, c.GetHeader("Sec-Fetch-Site"))
			c.JSON(http.StatusForbidden, gin.H{"error": "Cross-site request refused"})
			c.Abort()
			return
		}

		token, _ := c.Cookie(AuthTokenCookieName)

		if devTokenMatches(token) {
			user, err := resolveDevUser(lookup)
			if err != nil {
				rejectSession(c, err)
				return
			}
			log.Printf("[auth] DEV_AUTH_TOKEN cookie bypass active (user=%q id=%d) — DO NOT USE IN PRODUCTION", user.Username, user.ID)
			if !setRequestUser(c, user) {
				return
			}
			c.Next()
			return
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		claims, err := VerifyToken(token)
		if err != nil {
			rejectSession(c, err)
			return
		}
		user, err := resolveTokenUser(lookup, claims)
		if err != nil {
			rejectSession(c, err)
			return
		}

		if !setRequestUser(c, user) {
			return
		}
		c.Next()
	}
}

// OutsideRequestTx runs the rest of the route on the pool, without the request transaction. For
// /auth/login and /auth/register: there is no user yet, `users` has no row-level security, and
// they must not hold a connection while bcrypt runs (tens to hundreds of ms per attempt, so a
// burst of login attempts would otherwise pin the pool).
func OutsideRequestTx(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(database.WithDB(c.Request.Context(), db))
		c.Next()
	}
}

// crossSiteWrite reports a state-changing request a browser says came from another origin.
// The SPA only calls the API same-origin (Caddy and vite proxy /api), so anything else with
// the cookie attached is CSRF, including from sibling subdomains that SameSite=Lax lets through.
// Requests without the header (non-browser clients, browsers too old to send it) fall back to
// SameSite=Lax.
func crossSiteWrite(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	}
	switch r.Header.Get("Sec-Fetch-Site") {
	case "", "same-origin", "none":
		return false
	}
	return true
}

// resolveTokenUser checks a signature-verified token against the user's current row.
func resolveTokenUser(ctx context.Context, claims *Claims) (models.User, error) {
	if claims.UID == 0 {
		return models.User{}, errNoUID
	}
	user, err := authrepo.FindUserByID(ctx, claims.UID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, errUserGone
		}
		return models.User{}, err
	}
	if user.DisabledAt != nil {
		return models.User{}, errUserDisabled
	}
	if claims.Ver != user.TokenVersion {
		return models.User{}, errTokenRevoked
	}
	return user, nil
}

// setRequestUser puts user on the request context and makes them the request transaction's
// identity. It answers 503 and returns false if that fails.
func setRequestUser(c *gin.Context, user models.User) bool {
	if err := database.SetIdentity(c.Request.Context(), database.Identity{UserID: user.ID, IsAdmin: user.IsAdmin}); err != nil {
		rejectSession(c, err)
		return false
	}
	c.Request = c.Request.WithContext(WithUser(c.Request.Context(), User{
		ID:       user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
	}))
	// Kept for handlers that still read the gin key; new code uses UserID(ctx).
	c.Set("username", user.Username)
	return true
}

func rejectSession(c *gin.Context, reason error) {
	if errors.Is(reason, ErrInvalidToken) {
		log.Printf("[auth] rejected session: %v", reason)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
	} else {
		// A database or config error is not the client's fault and must not sign them out.
		log.Printf("[auth] could not verify session: %v", reason)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Could not verify session"})
	}
	c.Abort()
}
