package auth

import (
	"be-simpletracker/internal/env"
	"errors"
	"fmt"
	"net/url"
)

func ValidateProductionConfig() error {
	if !env.IsProduction() {
		return nil
	}
	if env.StringOr("ALLOW_BYPASS", "false") == "true" {
		return errors.New("ALLOW_BYPASS must be false in production")
	}
	if env.OptionalString("DEV_AUTH_TOKEN") != "" {
		return errors.New("DEV_AUTH_TOKEN must not be set in production")
	}
	if env.StringOr("REGISTER_ENABLED", "false") == "true" {
		return errors.New("REGISTER_ENABLED must be false in production")
	}
	jwtSecret, err := env.String("JWT_SECRET")
	if err != nil || len([]byte(jwtSecret)) < 32 {
		return errors.New("JWT_SECRET must contain at least 32 bytes in production")
	}
	origins, err := env.Slice("CORS_ORIGINS", ",")
	if err != nil {
		return errors.New("CORS_ORIGINS must be set in production")
	}
	for _, origin := range origins {
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("production CORS origin must use HTTPS: %q", origin)
		}
	}
	return nil
}
