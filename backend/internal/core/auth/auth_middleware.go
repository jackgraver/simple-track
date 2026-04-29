package auth

import (
	"log"
	"net/http"
	"time"

	"be-simpletracker/internal/env"

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

func devTokenMatches(value string) bool {
	secret := env.OptionalString("DEV_AUTH_TOKEN")
	if secret == "" || isProdEnv() {
		return false
	}
	return value == secret
}

func applyDevAuthUser(c *gin.Context) {
	user := env.StringOr("DEV_AUTH_USER", "dev")
	log.Printf("[auth] DEV_AUTH_TOKEN cookie bypass active (user=%q) — DO NOT USE IN PRODUCTION", user)
	c.Set("username", user)
	c.Set("timestamp", time.Now().Unix())
	c.Next()
}

func isProdEnv() bool {
	return env.StringOr("APP_ENV", "prod") == "prod"
}
