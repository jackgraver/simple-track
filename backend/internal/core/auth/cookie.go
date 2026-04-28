package auth

import (
	"net/http"
	"strings"

	"be-simpletracker/internal/env"
)

const AuthTokenCookieName = "auth_token"

const defaultCookieMaxAgeSec = 365 * 24 * 3600

func CookieMaxAgeSeconds() int {
	return env.IntOr("AUTH_COOKIE_MAX_AGE_SEC", defaultCookieMaxAgeSec)
}

func CookieSameSite() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(env.OptionalString("AUTH_COOKIE_SAMESITE"))) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

func CookieSecure() bool {
	if CookieSameSite() == http.SameSiteNoneMode {
		return true
	}
	s := env.OptionalString("AUTH_COOKIE_SECURE")
	return s == "true" || s == "1"
}
