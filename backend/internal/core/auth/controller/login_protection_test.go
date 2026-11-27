package controller

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestLoginProtectionLimitsPairAndSuppressesRepeatedRejectionLogs(t *testing.T) {
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
	decision := protection.Allow("192.0.2.1", "alice")
	if decision.Allowed || !decision.LogRejection || decision.RetryAfter != time.Minute {
		t.Fatalf("unexpected first rejection: %#v", decision)
	}
	decision = protection.Allow("192.0.2.1", "ALICE")
	if decision.Allowed || decision.LogRejection {
		t.Fatalf("unexpected repeated rejection: %#v", decision)
	}
	// Other usernames from the same IP are unaffected by the pair block.
	if !protection.Allow("192.0.2.1", "bob").Allowed {
		t.Fatal("pair block leaked to another username")
	}

	now = now.Add(time.Minute)
	if decision = protection.Allow("192.0.2.1", "alice"); !decision.Allowed {
		t.Fatalf("attempt after reset unexpectedly blocked: %#v", decision)
	}
}

func TestLoginProtectionOneIPCannotLockUsernameForOthers(t *testing.T) {
	protection := NewLoginProtection(LoginProtectionConfig{
		Window:         15 * time.Minute,
		MaxPerIP:       1000,
		MaxPerUsername: 5,
	}, nil)
	for i := 0; i < 200; i++ {
		protection.Allow("192.0.2.1", "jack")
	}
	if protection.Allow("192.0.2.1", "jack").Allowed {
		t.Fatal("hammering IP was not blocked")
	}
	d := protection.Allow("198.51.100.7", "jack")
	if !d.Allowed {
		t.Fatalf("owner from another IP blocked: %#v", d)
	}
	// Rejected attempts don't count username-wide, so one IP adds at most MaxPerUsername.
	if d.Delay < 0 || d.Delay > usernameDelayMax {
		t.Fatalf("unbounded delay %v", d.Delay)
	}
	protection.RecordSuccess("198.51.100.7", "jack")
	if d := protection.Allow("198.51.100.7", "jack"); !d.Allowed || d.Delay != 0 {
		t.Fatalf("success did not clear username delay: %#v", d)
	}
}

func TestLoginProtectionUsernameDelayGrowsAndIsCapped(t *testing.T) {
	protection := NewLoginProtection(LoginProtectionConfig{
		Window:         time.Hour,
		MaxPerIP:       100000,
		MaxPerUsername: 2,
	}, nil)
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	protection.now = func() time.Time { return now }
	var delays []time.Duration
	for i := 0; i < 60; i++ {
		// A distinct IP each time so only the username-wide counter grows; each client waits
		// out its delay before the next one arrives.
		d := protection.Allow(fmt.Sprintf("10.0.%d.%d", i/250, i%250), "jack")
		if !d.Allowed {
			t.Fatalf("attempt %d blocked: %#v", i, d)
		}
		delays = append(delays, d.Delay)
		now = now.Add(d.Delay)
	}
	// 2*MaxPerUsername = 4 free attempts, then 250ms doubling.
	for i := 0; i < 4; i++ {
		if delays[i] != 0 {
			t.Fatalf("attempt %d delayed %v", i, delays[i])
		}
	}
	if delays[4] != usernameDelayBase || delays[5] != 2*usernameDelayBase || delays[6] != 4*usernameDelayBase {
		t.Fatalf("delay did not double: %v", delays[:8])
	}
	for i := 1; i < len(delays); i++ {
		if delays[i] < delays[i-1] || delays[i] > usernameDelayMax {
			t.Fatalf("delay not monotonic/capped at %d: %v", i, delays[i])
		}
	}
	if delays[len(delays)-1] != usernameDelayMax {
		t.Fatalf("delay never reached cap: %v", delays[len(delays)-1])
	}
}

// Many sources in parallel can't multiply the username-wide rate: their delays queue up one
// slot after another, and once the queue is longer than usernameQueueMax they're turned away.
func TestLoginProtectionUsernameBackoffSerializesParallelSources(t *testing.T) {
	protection := NewLoginProtection(LoginProtectionConfig{
		Window:         time.Hour,
		MaxPerIP:       100000,
		MaxPerUsername: 2,
	}, nil)
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	protection.now = func() time.Time { return now }
	allowed, rejected := 0, 0
	var last time.Duration
	for i := 0; i < 200; i++ {
		d := protection.Allow(fmt.Sprintf("10.2.%d.%d", i/250, i%250), "jack")
		if !d.Allowed {
			if d.Reason != "username_backoff" || d.RetryAfter <= 0 {
				t.Fatalf("attempt %d: unexpected rejection %#v", i, d)
			}
			rejected++
			continue
		}
		allowed++
		if d.Delay < last {
			t.Fatalf("attempt %d: delay %v shorter than previous %v (not serialized)", i, d.Delay, last)
		}
		last = d.Delay
		if d.Delay > usernameQueueMax+usernameDelayMax {
			t.Fatalf("attempt %d: delay %v exceeds the queue bound", i, d.Delay)
		}
	}
	if rejected == 0 || allowed > 20 {
		t.Fatalf("parallel attempts not bounded: allowed=%d rejected=%d", allowed, rejected)
	}
	// Rejected attempts don't extend the queue: once it drains, a login gets through again.
	now = now.Add(time.Minute)
	if d := protection.Allow("192.0.2.1", "jack"); !d.Allowed || d.Delay > usernameDelayMax {
		t.Fatalf("after the queue drained: %#v", d)
	}
}

func TestLoginProtectionWaitUsesInjectedSleep(t *testing.T) {
	protection := NewLoginProtection(LoginProtectionConfig{}, nil)
	var slept time.Duration
	protection.sleep = func(_ context.Context, d time.Duration) { slept += d }
	protection.Wait(context.Background(), 0)
	protection.Wait(context.Background(), time.Second)
	if slept != time.Second {
		t.Fatalf("slept %v", slept)
	}
}

func TestLoginClientKeyNormalization(t *testing.T) {
	same := [][2]string{
		{"2001:db8:1:2::1", "2001:db8:1:2:ffff:ffff:ffff:ffff"},
		{"::ffff:192.0.2.1", "192.0.2.1"},
		{" 2001:DB8:1:2::5 ", "2001:db8:1:2::6"},
	}
	for _, p := range same {
		if loginClientKey(p[0]) != loginClientKey(p[1]) {
			t.Fatalf("%q and %q should share a key (%q vs %q)", p[0], p[1], loginClientKey(p[0]), loginClientKey(p[1]))
		}
	}
	different := [][2]string{
		{"2001:db8:1:2::1", "2001:db8:1:3::1"},
		{"192.0.2.1", "192.0.2.2"},
		{"::ffff:192.0.2.1", "::ffff:192.0.2.2"},
	}
	for _, p := range different {
		if loginClientKey(p[0]) == loginClientKey(p[1]) {
			t.Fatalf("%q and %q should not share a key", p[0], p[1])
		}
	}
	if loginClientKey("") != "unknown" || loginClientKey("not-an-ip") != "not-an-ip" {
		t.Fatal("unparseable addresses should fall back to the normalized string")
	}
}

func TestLoginProtectionIPv6SameSlash64SharesIPLimit(t *testing.T) {
	protection := NewLoginProtection(LoginProtectionConfig{Window: time.Minute, MaxPerIP: 2, MaxPerUsername: 100}, nil)
	if !protection.Allow("2001:db8:1:2::1", "a").Allowed || !protection.Allow("2001:db8:1:2::2", "b").Allowed {
		t.Fatal("first attempts blocked")
	}
	if protection.Allow("2001:db8:1:2::3", "c").Allowed {
		t.Fatal("rotating inside a /64 escaped the IP limit")
	}
	if !protection.Allow("2001:db8:1:3::1", "c").Allowed {
		t.Fatal("different /64 was blocked")
	}
}

func TestLoginProtectionFullLimiterEvictsOldestAndStillAccepts(t *testing.T) {
	protection := NewLoginProtection(LoginProtectionConfig{
		Window:         time.Hour,
		MaxPerIP:       1000000,
		MaxPerUsername: 1000000,
	}, nil)
	protection.maxEntries = 100
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	protection.now = func() time.Time { return now }

	for i := 0; i < 500; i++ {
		now = now.Add(time.Second)
		d := protection.Allow(fmt.Sprintf("10.1.%d.%d", i/250, i%250), fmt.Sprintf("user%d", i))
		if !d.Allowed {
			t.Fatalf("login %d rejected with %q when limiter full", i, d.Reason)
		}
		if n := len(protection.entries) + len(protection.sprays); n > protection.maxEntries {
			t.Fatalf("limiter grew to %d", n)
		}
	}
	if _, ok := protection.entries["ip:10.1.0.0"]; ok {
		t.Fatal("oldest entry was not evicted")
	}
	if _, ok := protection.entries["ip:10.1.1.249"]; !ok {
		t.Fatal("newest entry missing")
	}
}

func TestLoginProtectionFullLimiterKeepsBlockedEntries(t *testing.T) {
	protection := NewLoginProtection(LoginProtectionConfig{
		Window:         time.Hour,
		MaxPerIP:       1000000,
		MaxPerUsername: 2,
	}, nil)
	protection.maxEntries = 100
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	protection.now = func() time.Time { return now }

	// The attacker's own (ip, username) pair is blocked first, so it is the oldest entry.
	for i := 0; i < 2; i++ {
		protection.Allow("203.0.113.9", "jack")
	}
	if d := protection.Allow("203.0.113.9", "jack"); d.Allowed {
		t.Fatal("pair not blocked")
	}
	// Then floods the limiter with fresh keys to try to evict that block.
	for i := 0; i < 500; i++ {
		now = now.Add(time.Second)
		protection.Allow(fmt.Sprintf("10.3.%d.%d", i/250, i%250), fmt.Sprintf("user%d", i))
	}
	if d := protection.Allow("203.0.113.9", "jack"); d.Allowed {
		t.Fatal("flooding the limiter evicted an active block")
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
	protection.LogAttempt(context.Background(), "invalid_credentials", "192.0.2.1", "alice", true)
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

func TestLoginAttemptLogOmitsUnknownUsername(t *testing.T) {
	var output bytes.Buffer
	protection := NewLoginProtection(LoginProtectionConfig{}, slog.New(slog.NewJSONHandler(&output, nil)))
	protection.LogAttempt(context.Background(), "invalid_credentials", "192.0.2.1", "hunter2-typed-as-username", false)
	logged := output.String()
	if strings.Contains(logged, "hunter2") || !strings.Contains(logged, `"username":"unknown"`) {
		t.Fatalf("unknown username leaked or not masked: %s", logged)
	}
}
