package database

import (
	"strings"
	"testing"
)

func TestResolvePostgresDSNRequiresExplicitProductionURL(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_PRODUCTION", "")
	_, err := AppDSN()
	if err == nil {
		t.Fatal("expected missing production database URL error")
	}
}

func TestResolvePostgresDSNRejectsDisabledProductionTLS(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://user:pass@db.example/app?sslmode=disable")
	t.Setenv("DATABASE_ALLOW_INSECURE", "false")
	_, err := AppDSN()
	if err == nil || !strings.Contains(err.Error(), "sslmode") {
		t.Fatalf("expected TLS error, got %v", err)
	}
}

func TestResolvePostgresDSNAllowsExplicitInsecureProductionDatabase(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_PRODUCTION", "postgres://user:pass@postgres/app?sslmode=disable")
	t.Setenv("DATABASE_ALLOW_INSECURE", "true")
	if _, err := AppDSN(); err != nil {
		t.Fatal(err)
	}
}

func TestResolvePostgresDSNAcceptsProductionTLS(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	want := "postgres://user:pass@db.example/app?sslmode=verify-full"
	t.Setenv("DATABASE_URL", want)
	got, err := AppDSN()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolvePostgresDSNAllowsLocalDevelopmentWithoutTLS(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://localhost/app?sslmode=disable")
	if _, err := AppDSN(); err != nil {
		t.Fatal(err)
	}
}
