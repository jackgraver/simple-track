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
