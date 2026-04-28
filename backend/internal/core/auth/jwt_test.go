package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGenerateToken_verifyToken_roundTrip(t *testing.T) {
	tok, err := GenerateToken("alice")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	got, err := VerifyToken(tok)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if got.Username != "alice" {
		t.Fatalf("username: got %q want alice", got.Username)
	}
}

func TestVerifyToken_badSignature(t *testing.T) {
	tok, err := GenerateToken("bob")
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
	if _, err := GenerateToken("u"); err == nil {
		t.Fatal("expected error when JWT_SECRET is empty")
	}
}
