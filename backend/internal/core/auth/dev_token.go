package auth

import (
	"be-simpletracker/internal/core/auth/models"
	authrepo "be-simpletracker/internal/core/auth/repository"
	"be-simpletracker/internal/env"
	"crypto/subtle"
	"errors"

	"gorm.io/gorm"
)

func devTokenMatches(value string) bool {
	secret := env.OptionalString("DEV_AUTH_TOKEN")
	if secret == "" || !bypassAllowed() {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(value), []byte(secret)) == 1
}

// resolveDevUser loads DEV_AUTH_USER (default "dev") so the bypass acts as that real user,
// with its real ID. Looked up on every request: e2e reseeds the DB between tests.
func resolveDevUser() (models.User, error) {
	user, err := authrepo.FindUserByUsername(env.StringOr("DEV_AUTH_USER", "dev"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, errDevUserNotFound
		}
		return models.User{}, err
	}
	if user.DisabledAt != nil {
		return models.User{}, errUserDisabled
	}
	return user, nil
}

func bypassAllowed() bool {
	return !env.IsProduction() && env.StringOr("ALLOW_BYPASS", "false") == "true"
}
