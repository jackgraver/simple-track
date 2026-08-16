package controller

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestLoginProtectionLimitsUsernameAndSuppressesRepeatedRejectionLogs(t *testing.T) {
	protection := NewLoginProtection(LoginProtectionConfig{
		Window:         time.Minute,
		MaxPerIP:       100,
		MaxPerUsername: 2,
	}, nil)
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	protection.now = func() time.Time { return now }

	for i := 0; i < 2; i++ {
		if decision := protection.Allow("192.0.2.1", "Alice"); !decision.Allowed {
			t.Fatalf("attempt %d unexpectedly blocked", i+1)
		}
	}
	decision := protection.Allow("192.0.2.2", "alice")
	if decision.Allowed || !decision.LogRejection || decision.RetryAfter != time.Minute {
		t.Fatalf("unexpected first rejection: %#v", decision)
	}
	decision = protection.Allow("192.0.2.3", "ALICE")
	if decision.Allowed || decision.LogRejection {
		t.Fatalf("unexpected repeated rejection: %#v", decision)
	}

	now = now.Add(time.Minute)
	if decision = protection.Allow("192.0.2.3", "alice"); !decision.Allowed {
		t.Fatalf("attempt after reset unexpectedly blocked: %#v", decision)
	}
}

func TestLoginProtectionLimitsIPAcrossUsernames(t *testing.T) {
	protection := NewLoginProtection(LoginProtectionConfig{
		Window:         time.Minute,
		MaxPerIP:       2,
		MaxPerUsername: 100,
	}, nil)
	if !protection.Allow("192.0.2.1", "alice").Allowed {
		t.Fatal("first attempt blocked")
	}
	if !protection.Allow("192.0.2.1", "bob").Allowed {
		t.Fatal("second attempt blocked")
	}
	if protection.Allow("192.0.2.1", "carol").Allowed {
		t.Fatal("IP limit did not block third attempt")
	}
}

func TestLoginProtectionSuccessResetsUsernameLimit(t *testing.T) {
	protection := NewLoginProtection(LoginProtectionConfig{
		Window:         time.Minute,
		MaxPerIP:       100,
		MaxPerUsername: 1,
	}, nil)
	if !protection.Allow("192.0.2.1", "alice").Allowed {
		t.Fatal("first attempt blocked")
	}
	protection.RecordSuccess("192.0.2.1", "Alice")
	if !protection.Allow("192.0.2.2", "alice").Allowed {
		t.Fatal("successful login did not reset username limit")
	}
}

func TestLoginProtectionBlocksCredentialSprayForOneDay(t *testing.T) {
	protection := NewLoginProtection(LoginProtectionConfig{
		Window:                  time.Minute,
		MaxPerIP:                100,
		MaxPerUsername:          100,
		SprayWindow:             10 * time.Minute,
		MaxDistinctUsernames:    3,
		CredentialSprayBlockFor: 24 * time.Hour,
	}, nil)
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	protection.now = func() time.Time { return now }

	for _, username := range []string{"alice", "bob"} {
		if !protection.Allow("192.0.2.1", username).Allowed {
			t.Fatalf("%s was blocked before spray threshold", username)
		}
		if !protection.RecordFailure("192.0.2.1", username).Allowed {
			t.Fatalf("%s triggered spray block too early", username)
		}
	}
	if !protection.Allow("192.0.2.1", "carol").Allowed {
		t.Fatal("triggering attempt was blocked before credential verification")
	}
	decision := protection.RecordFailure("192.0.2.1", "carol")
	if decision.Allowed || decision.Reason != "credential_spray" || decision.RetryAfter != 24*time.Hour || !decision.LogRejection {
		t.Fatalf("unexpected spray decision: %#v", decision)
	}
	decision = protection.Allow("192.0.2.1", "dave")
	if decision.Allowed || decision.Reason != "credential_spray" {
		t.Fatalf("credential spray block was not enforced: %#v", decision)
	}

	now = now.Add(24 * time.Hour)
	if decision = protection.Allow("192.0.2.1", "dave"); !decision.Allowed {
		t.Fatalf("IP remained blocked after one day: %#v", decision)
	}
}

func TestLoginAttemptLogContainsAuditFieldsOnly(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	protection := NewLoginProtection(LoginProtectionConfig{}, logger)
	protection.LogAttempt(context.Background(), "invalid_credentials", "192.0.2.1", "alice")
	logged := output.String()
	for _, value := range []string{"auth.login", "invalid_credentials", "192.0.2.1", "alice"} {
		if !strings.Contains(logged, value) {
			t.Fatalf("missing %q in log %s", value, logged)
		}
	}
	for _, forbidden := range []string{"password", "token", "cookie"} {
		if strings.Contains(strings.ToLower(logged), forbidden) {
			t.Fatalf("credential field %q appeared in log %s", forbidden, logged)
		}
	}
}
