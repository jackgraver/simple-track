package models

import "time"

// Setup link purposes: "setup" for an account that has never had a password, "reset" otherwise.
const (
	SetupPurposeSetup = "setup"
	SetupPurposeReset = "reset"
)

// SetupToken is a one-time link to set or reset a user's password. Only TokenHash (SHA-256 of
// the token, hex) is stored; the token itself exists only in the link the admin texts.
type SetupToken struct {
	ID        uint       `gorm:"primaryKey"`
	UserID    uint       `gorm:"not null;index"`
	TokenHash string     `gorm:"not null;uniqueIndex"`
	Purpose   string     `gorm:"not null"`
	CreatedBy *uint      ``
	CreatedAt time.Time  `gorm:"not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time ``
	RevokedAt *time.Time ``
}

func (SetupToken) TableName() string { return "user_setup_tokens" }

// AuthEvent is one row of the append-only audit log (written with raw INSERTs, see
// authrepo.InsertEvent; the gorm model is for reads and SQLite module tests).
type AuthEvent struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"not null"`
	Event     string    `gorm:"not null"`
	UserID    *uint     ``
	ActorID   *uint     ``
	ClientIP  string    `gorm:"not null;default:''"`
	Detail    string    `gorm:"not null;default:''"`
}

func (AuthEvent) TableName() string { return "auth_events" }

// auth_events.event values. Detail never holds a password or an attempted username that
// doesn't exist (BACKLOG #19); user_id identifies the account when there is one.
const (
	EventLoginOK              = "login_ok"
	EventLoginFailed          = "login_failed"       // detail: wrong_password, disabled, no_password, unknown_user
	EventLoginRateLimited     = "login_rate_limited" // detail: the limiter's reason
	EventUserCreated          = "user_created"
	EventUserDisabled         = "user_disabled"
	EventUserEnabled          = "user_enabled"
	EventSetupLinkCreated     = "setup_link_created" // detail: setup or reset
	EventSetupLinkUsed        = "setup_link_used"    // password set and signed in
	EventSetupLinkRejected    = "setup_link_rejected"
	EventPasswordChanged      = "password_changed"
	EventPasswordChangeFailed = "password_change_failed"
)
