package controller

import (
	"be-simpletracker/internal/env"
	"context"
	"log/slog"
	"net/netip"
	"sort"
	"strings"
	"sync"
	"time"
)

const maxLoginRateEntries = 10000

// Username-wide pressure (all sources combined) never blocks; once a username has seen
// 2*MaxPerUsername attempts in a window, each further attempt is delayed by
// usernameDelayBase * 2^(excess-1), capped at usernameDelayMax.
const (
	usernameDelayBase = 250 * time.Millisecond
	usernameDelayMax  = 5 * time.Second
)

type LoginProtectionConfig struct {
	Window                  time.Duration
	MaxPerIP                int
	MaxPerUsername          int
	SprayWindow             time.Duration
	MaxDistinctUsernames    int
	CredentialSprayBlockFor time.Duration
}

func LoginProtectionConfigFromEnv() LoginProtectionConfig {
	return LoginProtectionConfig{
		Window:                  time.Duration(env.IntOr("LOGIN_RATE_LIMIT_WINDOW_SEC", 15*60)) * time.Second,
		MaxPerIP:                env.IntOr("LOGIN_RATE_LIMIT_MAX_IP", 30),
		MaxPerUsername:          env.IntOr("LOGIN_RATE_LIMIT_MAX_USERNAME", 5),
		SprayWindow:             time.Duration(env.IntOr("LOGIN_SPRAY_WINDOW_SEC", 10*60)) * time.Second,
		MaxDistinctUsernames:    env.IntOr("LOGIN_SPRAY_MAX_USERNAMES", 5),
		CredentialSprayBlockFor: time.Duration(env.IntOr("LOGIN_SPRAY_BLOCK_SEC", 24*60*60)) * time.Second,
	}
}

type LoginDecision struct {
	Allowed      bool
	RetryAfter   time.Duration
	LogRejection bool
	Reason       string
	// Delay is how long the caller should wait before checking credentials (username-wide
	// backoff). Only set on allowed decisions.
	Delay time.Duration
}

type loginRateWindow struct {
	Count           int
	ResetAt         time.Time
	RejectionLogged bool
}

type credentialSprayWindow struct {
	Usernames       map[string]struct{}
	ResetAt         time.Time
	BlockedUntil    time.Time
	RejectionLogged bool
}

type LoginProtection struct {
	mu                  sync.Mutex
	config              LoginProtectionConfig
	entries             map[string]*loginRateWindow
	sprays              map[string]*credentialSprayWindow
	logger              *slog.Logger
	now                 func() time.Time
	sleep               func(ctx context.Context, d time.Duration)
	maxEntries          int
	lastCleanup         time.Time
	lastCapacityWarning time.Time
}

func NewLoginProtection(config LoginProtectionConfig, logger *slog.Logger) *LoginProtection {
	if config.Window <= 0 {
		config.Window = 15 * time.Minute
	}
	if config.MaxPerIP <= 0 {
		config.MaxPerIP = 30
	}
	if config.MaxPerUsername <= 0 {
		config.MaxPerUsername = 5
	}
	if config.SprayWindow <= 0 {
		config.SprayWindow = 10 * time.Minute
	}
	if config.MaxDistinctUsernames <= 0 {
		config.MaxDistinctUsernames = 5
	}
	if config.CredentialSprayBlockFor <= 0 {
		config.CredentialSprayBlockFor = 24 * time.Hour
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &LoginProtection{
		config:  config,
		entries: make(map[string]*loginRateWindow),
		sprays:  make(map[string]*credentialSprayWindow),
		logger:  logger,
		now:     time.Now,
		sleep: func(ctx context.Context, d time.Duration) {
			t := time.NewTimer(d)
			defer t.Stop()
			select {
			case <-t.C:
			case <-ctx.Done():
			}
		},
		maxEntries: maxLoginRateEntries,
	}
}

// Wait pauses for a decision's Delay (no-op when zero). Injectable via p.sleep for tests.
func (p *LoginProtection) Wait(ctx context.Context, d time.Duration) {
	if d > 0 {
		p.sleep(ctx, d)
	}
}

func (p *LoginProtection) Allow(clientIP, username string) LoginDecision {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := p.now()
	p.cleanupExpired(now)
	ip := loginClientKey(clientIP)
	user := normalizeLoginKey(username)
	ipKey := "ip:" + ip
	pairKey := "pair:" + ip + "|" + user
	usernameKey := "username:" + user

	if spray, ok := p.sprays[ipKey]; ok && now.Before(spray.BlockedUntil) {
		logRejection := !spray.RejectionLogged
		spray.RejectionLogged = true
		return LoginDecision{
			Allowed:      false,
			RetryAfter:   spray.BlockedUntil.Sub(now),
			LogRejection: logRejection,
			Reason:       "credential_spray",
		}
	}

	// Strict limits: per source IP, and per (source IP, username). The username-wide counter
	// is deliberately not a block, so one source can't lock a username for everyone.
	ipBlocked, ipRetry, ipLog := p.isBlocked(ipKey, p.config.MaxPerIP, now)
	pairBlocked, pairRetry, pairLog := p.isBlocked(pairKey, p.config.MaxPerUsername, now)
	if ipBlocked || pairBlocked {
		return LoginDecision{
			Allowed:      false,
			RetryAfter:   maxDuration(ipRetry, pairRetry),
			LogRejection: ipLog || pairLog,
			Reason:       "rate_limit",
		}
	}

	p.makeRoom(now, []string{ipKey, pairKey, usernameKey})
	p.increment(ipKey, now)
	p.increment(pairKey, now)
	p.increment(usernameKey, now)
	return LoginDecision{Allowed: true, Delay: p.usernameDelay(usernameKey)}
}

// usernameDelay is the backoff for the username-wide counter (already incremented).
func (p *LoginProtection) usernameDelay(key string) time.Duration {
	entry, ok := p.entries[key]
	if !ok {
		return 0
	}
	excess := entry.Count - 2*p.config.MaxPerUsername
	if excess <= 0 {
		return 0
	}
	d := usernameDelayMax
	if excess-1 < 16 {
		if v := usernameDelayBase << uint(excess-1); v < d {
			d = v
		}
	}
	return d
}

// makeRoom evicts the oldest entries (earliest reset time) when adding the given new keys
// would exceed the cap, instead of rejecting logins. It frees ~10% extra to amortize the sort.
func (p *LoginProtection) makeRoom(now time.Time, keys []string) {
	newCount := 0
	for _, k := range keys {
		if _, ok := p.entries[k]; !ok {
			newCount++
		}
	}
	if len(p.entries)+len(p.sprays)+newCount <= p.maxEntries {
		return
	}
	target := p.maxEntries - p.maxEntries/10 - newCount
	if target < 0 {
		target = 0
	}
	type aged struct {
		key   string
		at    time.Time
		spray bool
	}
	all := make([]aged, 0, len(p.entries)+len(p.sprays))
	for k, e := range p.entries {
		all = append(all, aged{k, e.ResetAt, false})
	}
	for k, sp := range p.sprays {
		// Active spray blocks go last so they outlive plain counters.
		at := sp.ResetAt
		if now.Before(sp.BlockedUntil) {
			at = sp.BlockedUntil.Add(p.config.Window * 1000)
		}
		all = append(all, aged{k, at, true})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].at.Before(all[j].at) })
	for i := 0; i < len(all) && len(p.entries)+len(p.sprays) > target; i++ {
		if all[i].spray {
			delete(p.sprays, all[i].key)
		} else {
			delete(p.entries, all[i].key)
		}
	}
	if now.Sub(p.lastCapacityWarning) >= time.Minute {
		p.lastCapacityWarning = now
		p.logger.Warn("login limiter full, evicted oldest entries", slog.String("event", "auth.login_limiter"))
	}
}

func (p *LoginProtection) RecordFailure(clientIP, username string) LoginDecision {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := p.now()
	ipKey := "ip:" + loginClientKey(clientIP)
	usernameKey := normalizeLoginKey(username)
	spray, ok := p.sprays[ipKey]
	if !ok || (!now.Before(spray.ResetAt) && !now.Before(spray.BlockedUntil)) {
		spray = &credentialSprayWindow{
			Usernames: make(map[string]struct{}),
			ResetAt:   now.Add(p.config.SprayWindow),
		}
		p.sprays[ipKey] = spray
	}
	if now.Before(spray.BlockedUntil) {
		return LoginDecision{
			Allowed:    false,
			RetryAfter: spray.BlockedUntil.Sub(now),
			Reason:     "credential_spray",
		}
	}
	spray.Usernames[usernameKey] = struct{}{}
	if len(spray.Usernames) >= p.config.MaxDistinctUsernames {
		spray.BlockedUntil = now.Add(p.config.CredentialSprayBlockFor)
		spray.RejectionLogged = true
		return LoginDecision{
			Allowed:      false,
			RetryAfter:   p.config.CredentialSprayBlockFor,
			LogRejection: true,
			Reason:       "credential_spray",
		}
	}
	return LoginDecision{Allowed: true}
}

func (p *LoginProtection) RecordSuccess(clientIP, username string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	usernameKey := normalizeLoginKey(username)
	ip := loginClientKey(clientIP)
	delete(p.entries, "username:"+usernameKey)
	delete(p.entries, "pair:"+ip+"|"+usernameKey)
	if spray, ok := p.sprays["ip:"+ip]; ok {
		delete(spray.Usernames, usernameKey)
	}
}

// LogAttempt records a login attempt. knownUser says whether username exists in users; an
// unknown name (possibly a mistyped password) is logged as "unknown", never verbatim.
func (p *LoginProtection) LogAttempt(ctx context.Context, outcome, clientIP, username string, knownUser bool) {
	loggedName := "unknown"
	if knownUser {
		loggedName = strings.TrimSpace(username)
	}
	level := slog.LevelInfo
	if outcome == "invalid_credentials" || outcome == "rate_limited" || outcome == "credential_spray_blocked" {
		level = slog.LevelWarn
	}
	if outcome == "server_error" {
		level = slog.LevelError
	}
	p.logger.LogAttrs(
		ctx,
		level,
		"authentication login attempt",
		slog.String("event", "auth.login"),
		slog.String("outcome", outcome),
		slog.String("username", loggedName),
		slog.String("client_ip", strings.TrimSpace(clientIP)),
	)
}

func (p *LoginProtection) cleanupExpired(now time.Time) {
	if len(p.entries)+len(p.sprays) < p.maxEntries && !p.lastCleanup.IsZero() && now.Sub(p.lastCleanup) < time.Minute {
		return
	}
	for key, entry := range p.entries {
		if !now.Before(entry.ResetAt) {
			delete(p.entries, key)
		}
	}
	for key, spray := range p.sprays {
		blockExpired := spray.BlockedUntil.IsZero() || !now.Before(spray.BlockedUntil)
		if blockExpired && !now.Before(spray.ResetAt) {
			delete(p.sprays, key)
		}
	}
	p.lastCleanup = now
}

func (p *LoginProtection) isBlocked(key string, limit int, now time.Time) (bool, time.Duration, bool) {
	entry, ok := p.entries[key]
	if !ok || !now.Before(entry.ResetAt) || entry.Count < limit {
		return false, 0, false
	}
	logRejection := !entry.RejectionLogged
	entry.RejectionLogged = true
	return true, entry.ResetAt.Sub(now), logRejection
}

func (p *LoginProtection) increment(key string, now time.Time) {
	entry, ok := p.entries[key]
	if !ok || !now.Before(entry.ResetAt) {
		p.entries[key] = &loginRateWindow{
			Count:   1,
			ResetAt: now.Add(p.config.Window),
		}
		return
	}
	entry.Count++
}

func normalizeLoginKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "unknown"
	}
	return value
}

// loginClientKey normalizes a client address: IPv4 (and IPv4-mapped IPv6) per address, IPv6
// by its /64 prefix, so rotating addresses inside one allocation shares a budget.
func loginClientKey(clientIP string) string {
	raw := strings.TrimSpace(clientIP)
	addr, err := netip.ParseAddr(raw)
	if err != nil {
		return normalizeLoginKey(raw)
	}
	addr = addr.WithZone("").Unmap()
	if addr.Is6() {
		prefix, err := addr.Prefix(64)
		if err != nil {
			return normalizeLoginKey(raw)
		}
		return prefix.String()
	}
	return addr.String()
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
