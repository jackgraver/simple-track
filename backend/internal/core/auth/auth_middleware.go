package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the JWT stored in the httpOnly auth_token cookie from the SPA.
//
// The default is strict: missing or invalid tokens always produce 401.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(AuthTokenCookieName)

		if devTokenMatches(token) {
			applyDevAuthUser(c)
			return
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		claims, err := VerifyToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Set("timestamp", claims.Timestamp)
		c.Next()
	}
}