package auth

import (
	"net/http"
	"os"
	"strconv"
	"strings"
)

const defaultCookieMaxAgeSec = 365 * 24 * 3600

// CookieMaxAgeSeconds returns AUTH_COOKIE_MAX_AGE_SEC or a long default (1 year).
func CookieMaxAgeSeconds() int {
	s := os.Getenv("AUTH_COOKIE_MAX_AGE_SEC")
	if s == "" {
		return defaultCookieMaxAgeSec
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return defaultCookieMaxAgeSec
	}
	return v
}

// CookieSameSite returns AUTH_COOKIE_SAMESITE: lax (default), strict, or none.
// Use "none" when the SPA and API are on different sites (cross-origin XHR); browsers require Secure with none.
func CookieSameSite() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_COOKIE_SAMESITE"))) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

// CookieSecure is true when AUTH_COOKIE_SECURE is set, or when SameSite is none (required by browsers).
func CookieSecure() bool {
	if CookieSameSite() == http.SameSiteNoneMode {
		return true
	}
	s := os.Getenv("AUTH_COOKIE_SECURE")
	return s == "true" || s == "1"
}
