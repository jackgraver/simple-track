package database

import (
	"strings"
	"testing"
	"time"
)

func TestPoolFromEnvDefaults(t *testing.T) {
	for _, k := range []string{"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME", "DB_CONN_MAX_IDLE_TIME"} {
		t.Setenv(k, "")
	}
	p, err := PoolFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if p != DefaultPool {
		t.Fatalf("got %+v want %+v", p, DefaultPool)
	}
}

func TestPoolFromEnvOverrides(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "8")
	t.Setenv("DB_MAX_IDLE_CONNS", "0")
	t.Setenv("DB_CONN_MAX_LIFETIME", "1h")
	t.Setenv("DB_CONN_MAX_IDLE_TIME", "90s")
	p, err := PoolFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	want := Pool{MaxOpen: 8, MaxIdle: 0, MaxLifetime: time.Hour, MaxIdleTime: 90 * time.Second}
	if p != want {
		t.Fatalf("got %+v want %+v", p, want)
	}
}

func TestPoolFromEnvCapsIdleAtOpen(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "4")
	t.Setenv("DB_MAX_IDLE_CONNS", "10")
	p, err := PoolFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if p.MaxIdle != 4 {
		t.Fatalf("max idle %d, want 4", p.MaxIdle)
	}
}

func TestPoolFromEnvRejectsInvalid(t *testing.T) {
	cases := map[string]string{
		"DB_MAX_OPEN_CONNS":     "0",
		"DB_MAX_IDLE_CONNS":     "-1",
		"DB_CONN_MAX_LIFETIME":  "30", // no unit
		"DB_CONN_MAX_IDLE_TIME": "-5m",
	}
	for key, val := range cases {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, val)
			if _, err := PoolFromEnv(); err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("%s=%q: want an error naming %s, got %v", key, val, key, err)
			}
		})
	}
}

func TestOwnerDSNFallsBackToAppDSN(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://simpletrack_app:x@localhost/app?sslmode=disable")
	t.Setenv("DATABASE_OWNER_URL", "")
	got, err := OwnerDSN()
	if err != nil || got != "postgres://simpletrack_app:x@localhost/app?sslmode=disable" {
		t.Fatalf("got %q, %v", got, err)
	}
	t.Setenv("DATABASE_OWNER_URL", "postgres://owner:y@localhost/app?sslmode=disable")
	if got, _ := OwnerDSN(); got != "postgres://owner:y@localhost/app?sslmode=disable" {
		t.Fatalf("got %q", got)
	}
}

func TestOwnerDSNChecksProductionTLS(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_ALLOW_INSECURE", "false")
	t.Setenv("DATABASE_OWNER_URL", "postgres://owner:y@db.example/app?sslmode=disable")
	if _, err := OwnerDSN(); err == nil || !strings.Contains(err.Error(), "DATABASE_OWNER_URL") {
		t.Fatalf("want a TLS error naming DATABASE_OWNER_URL, got %v", err)
	}
}

func TestScramVerifierFormat(t *testing.T) {
	v, err := scramVerifierWithSalt("pencil", []byte("0123456789abcdef"), 4096)
	if err != nil {
		t.Fatal(err)
	}
	// SCRAM-SHA-256$<iterations>:<salt>$<StoredKey>:<ServerKey>, all base64 (32-byte keys).
	if !strings.HasPrefix(v, "SCRAM-SHA-256$4096:MDEyMzQ1Njc4OWFiY2RlZg==$") || len(v) != len("SCRAM-SHA-256$4096:MDEyMzQ1Njc4OWFiY2RlZg==$")+44+1+44 {
		t.Fatalf("unexpected verifier %q", v)
	}
	if _, err := scramVerifier("has space"); err == nil {
		t.Fatal("password with a space was accepted")
	}
}
