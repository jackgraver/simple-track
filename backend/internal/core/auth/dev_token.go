package auth

import (
	"be-simpletracker/internal/env"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

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
