package auth

import (
	"be-simpletracker/internal/core/auth/models"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestGenerateToken_verifyToken_roundTrip(t *testing.T) {
	tok, err := GenerateToken(models.User{Model: gorm.Model{ID: 42}, Username: "alice", TokenVersion: 3})
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	got, err := VerifyToken(tok)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if got.Username != "alice" || got.UID != 42 || got.Ver != 3 {
		t.Fatalf("claims: got %+v want alice uid 42 ver 3", got)
	}
}

func TestVerifyToken_badSignature(t *testing.T) {
	tok, err := GenerateToken(models.User{Model: gorm.Model{ID: 1}, Username: "bob", TokenVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(tok, ".")
	parts[2] += "x"
	badTok := strings.Join(parts, ".")
	if _, err := VerifyToken(badTok); err == nil {
		t.Fatal("expected error for corrupted signature")
	}
}

func TestVerifyToken_badFormat(t *testing.T) {
	if _, err := VerifyToken("two.parts"); err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyToken_expiredClaims(t *testing.T) {
	pastExp := time.Now().Unix() - 3600
	tok := mustSignedJWT(t, Claims{
		Username:  "eve",
		Timestamp: time.Now().Unix(),
		Iat:       pastExp - 7200,
		Exp:       pastExp,
	})
	if _, err := VerifyToken(tok); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func mustSignedJWT(tb testing.TB, claims Claims) string {
	tb.Helper()
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	hj, err := json.Marshal(header)
	if err != nil {
		tb.Fatal(err)
	}
	cj, err := json.Marshal(claims)
	if err != nil {
		tb.Fatal(err)
	}
	he := base64.RawURLEncoding.EncodeToString(hj)
	ce := base64.RawURLEncoding.EncodeToString(cj)
	sig, err := createSignature(he + "." + ce)
	if err != nil {
		tb.Fatal(err)
	}
	return he + "." + ce + "." + sig
}

func TestGenerateToken_missingSecretReturnsError(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if _, err := GenerateToken(models.User{Model: gorm.Model{ID: 1}, Username: "u"}); err == nil {
		t.Fatal("expected error when JWT_SECRET is empty")
	}
}

func TestGenerateToken_unsavedUserRefused(t *testing.T) {
	if _, err := GenerateToken(models.User{Username: "u"}); err == nil {
		t.Fatal("expected error for a user with no ID")
	}
}

func TestVerifyToken_rejectionsWrapErrInvalidToken(t *testing.T) {
	for _, tok := range []string{"two.parts", "a.b.c"} {
		if _, err := VerifyToken(tok); !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("%q: got %v, want ErrInvalidToken", tok, err)
		}
	}
}

func signedWithHeader(tb testing.TB, header map[string]string, claims Claims) string {
	tb.Helper()
	hj, _ := json.Marshal(header)
	cj, _ := json.Marshal(claims)
	he := base64.RawURLEncoding.EncodeToString(hj)
	ce := base64.RawURLEncoding.EncodeToString(cj)
	sig, err := createSignature(he + "." + ce)
	if err != nil {
		tb.Fatal(err)
	}
	return he + "." + ce + "." + sig
}

func TestVerifyToken_requiresExpAndIat(t *testing.T) {
	now := time.Now().Unix()
	for name, c := range map[string]Claims{
		"no exp":  {UID: 1, Ver: 1, Iat: now},
		"no iat":  {UID: 1, Ver: 1, Exp: now + 60},
		"neither": {UID: 1, Ver: 1},
	} {
		if _, err := VerifyToken(mustSignedJWT(t, c)); !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("%s: got %v, want ErrInvalidToken", name, err)
		}
	}
}

func TestVerifyToken_rejectsLifetimeLongerThanConfigured(t *testing.T) {
	t.Setenv("AUTH_COOKIE_MAX_AGE_SEC", "3600")
	now := time.Now().Unix()
	if _, err := VerifyToken(mustSignedJWT(t, Claims{UID: 1, Ver: 1, Iat: now, Exp: now + 3600})); err != nil {
		t.Fatalf("lifetime at the limit: %v", err)
	}
	if _, err := VerifyToken(mustSignedJWT(t, Claims{UID: 1, Ver: 1, Iat: now, Exp: now + 3601})); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("lifetime over the limit: got %v", err)
	}
}

func TestVerifyToken_requiresHS256Header(t *testing.T) {
	now := time.Now().Unix()
	c := Claims{UID: 1, Ver: 1, Iat: now, Exp: now + 60}
	for _, alg := range []string{"none", "HS512", ""} {
		tok := signedWithHeader(t, map[string]string{"alg": alg, "typ": "JWT"}, c)
		if _, err := VerifyToken(tok); !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("alg %q: got %v, want ErrInvalidToken", alg, err)
		}
	}
}
