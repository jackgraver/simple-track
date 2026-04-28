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
