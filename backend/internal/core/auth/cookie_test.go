package auth

import "testing"

func TestCookieMaxAgeSeconds_defaultWithoutEnv(t *testing.T) {
	t.Setenv("AUTH_COOKIE_MAX_AGE_SEC", "")
	if got := CookieMaxAgeSeconds(); got != defaultCookieMaxAgeSec {
		t.Fatalf("default max age: got %d want %d", got, defaultCookieMaxAgeSec)
	}
}

func TestCookieMaxAgeSeconds_fromEnv(t *testing.T) {
	t.Setenv("AUTH_COOKIE_MAX_AGE_SEC", "3600")
	if got := CookieMaxAgeSeconds(); got != 3600 {
		t.Fatalf("got %d want 3600", got)
	}
}

func TestCookieMaxAgeSeconds_invalidEnvUsesDefault(t *testing.T) {
	t.Setenv("AUTH_COOKIE_MAX_AGE_SEC", "not-a-number")
	if got := CookieMaxAgeSeconds(); got != defaultCookieMaxAgeSec {
		t.Fatalf("invalid env: got %d want default %d", got, defaultCookieMaxAgeSec)
	}
	t.Setenv("AUTH_COOKIE_MAX_AGE_SEC", "0")
	if got := CookieMaxAgeSeconds(); got != defaultCookieMaxAgeSec {
		t.Fatalf("zero env: got %d want default %d", got, defaultCookieMaxAgeSec)
	}
}

func TestCookieSecureAlwaysTrueInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_COOKIE_SECURE", "false")
	t.Setenv("AUTH_COOKIE_SAMESITE", "lax")
	if !CookieSecure() {
		t.Fatal("production cookie must be secure")
	}
}

func TestCookieSecureDefaultsFalseInDevelopment(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("AUTH_COOKIE_SECURE", "")
	t.Setenv("AUTH_COOKIE_SAMESITE", "lax")
	if CookieSecure() {
		t.Fatal("development cookie should allow HTTP")
	}
}
