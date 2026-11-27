package auth

import (
	"errors"
	"log"
	"net/http"

	"be-simpletracker/internal/core/auth/models"
	authrepo "be-simpletracker/internal/core/auth/repository"

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
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(AuthTokenCookieName)

		if devTokenMatches(token) {
			user, err := resolveDevUser()
			if err != nil {
				rejectSession(c, err)
				return
			}
			log.Printf("[auth] DEV_AUTH_TOKEN cookie bypass active (user=%q id=%d) — DO NOT USE IN PRODUCTION", user.Username, user.ID)
			setRequestUser(c, user)
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
		user, err := resolveTokenUser(claims)
		if err != nil {
			rejectSession(c, err)
			return
		}

		setRequestUser(c, user)
		c.Next()
	}
}

// resolveTokenUser checks a signature-verified token against the user's current row.
func resolveTokenUser(claims *Claims) (models.User, error) {
	if claims.UID == 0 {
		return models.User{}, errNoUID
	}
	user, err := authrepo.FindUserByID(claims.UID)
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

func setRequestUser(c *gin.Context, user models.User) {
	c.Request = c.Request.WithContext(WithUser(c.Request.Context(), User{
		ID:       user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
	}))
	// Kept for handlers that still read the gin key; new code uses UserID(ctx).
	c.Set("username", user.Username)
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
