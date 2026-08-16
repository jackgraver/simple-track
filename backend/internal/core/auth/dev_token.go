package auth

import (
	"be-simpletracker/internal/env"
	"crypto/subtle"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func devTokenMatches(value string) bool {
	secret := env.OptionalString("DEV_AUTH_TOKEN")
	if secret == "" || !bypassAllowed() {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(value), []byte(secret)) == 1
}

func applyDevAuthUser(c *gin.Context) {
	user := env.StringOr("DEV_AUTH_USER", "dev")
	log.Printf("[auth] DEV_AUTH_TOKEN cookie bypass active (user=%q) — DO NOT USE IN PRODUCTION", user)
	c.Set("username", user)
	c.Set("timestamp", time.Now().Unix())
	c.Next()
}

func bypassAllowed() bool {
	return !env.IsProduction() && env.StringOr("ALLOW_BYPASS", "false") == "true"
}
