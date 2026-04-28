package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the JWT stored in the httpOnly auth_token cookie from the SPA.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ""
		if ct, err := c.Cookie(AuthTokenCookieName); err == nil {
			token = ct
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

// GetUsername extracts the username from the context (set by AuthMiddleware)
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	return username.(string)
}
