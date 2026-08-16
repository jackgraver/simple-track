package controller

import (
	"be-simpletracker/internal/env"
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"
)

const maxLoginRateEntries = 10000

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
	}
}

func (p *LoginProtection) Allow(clientIP, username string) LoginDecision {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := p.now()
	p.cleanupExpired(now)
	ipKey := "ip:" + normalizeLoginKey(clientIP)
	usernameKey := "username:" + normalizeLoginKey(username)

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

	ipBlocked, ipRetry, ipLog := p.isBlocked(ipKey, p.config.MaxPerIP, now)
	usernameBlocked, usernameRetry, usernameLog := p.isBlocked(usernameKey, p.config.MaxPerUsername, now)
	if ipBlocked || usernameBlocked {
		return LoginDecision{
			Allowed:      false,
			RetryAfter:   maxDuration(ipRetry, usernameRetry),
			LogRejection: ipLog || usernameLog,
			Reason:       "rate_limit",
		}
	}

	newEntries := 0
	if _, ok := p.entries[ipKey]; !ok {
		newEntries++
	}
	if _, ok := p.entries[usernameKey]; !ok {
		newEntries++
	}
	if len(p.entries)+len(p.sprays)+newEntries > maxLoginRateEntries {
		logRejection := now.Sub(p.lastCapacityWarning) >= time.Minute
		if logRejection {
			p.lastCapacityWarning = now
		}
		return LoginDecision{
			Allowed:      false,
			RetryAfter:   p.config.Window,
			LogRejection: logRejection,
			Reason:       "capacity",
		}
	}

	p.increment(ipKey, now)
	p.increment(usernameKey, now)
	return LoginDecision{Allowed: true}
}

func (p *LoginProtection) RecordFailure(clientIP, username string) LoginDecision {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := p.now()
	ipKey := "ip:" + normalizeLoginKey(clientIP)
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
	delete(p.entries, "username:"+usernameKey)
	if spray, ok := p.sprays["ip:"+normalizeLoginKey(clientIP)]; ok {
		delete(spray.Usernames, usernameKey)
	}
}

func (p *LoginProtection) LogAttempt(ctx context.Context, outcome, clientIP, username string) {
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
		slog.String("username", strings.TrimSpace(username)),
		slog.String("client_ip", strings.TrimSpace(clientIP)),
	)
}

func (p *LoginProtection) cleanupExpired(now time.Time) {
	if len(p.entries)+len(p.sprays) < maxLoginRateEntries && !p.lastCleanup.IsZero() && now.Sub(p.lastCleanup) < time.Minute {
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

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
