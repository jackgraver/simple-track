package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/core/auth/models"
	authrepo "be-simpletracker/internal/core/auth/repository"
	"be-simpletracker/internal/database"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SetupLinkTTL is how long a setup or reset link works.
const SetupLinkTTL = 48 * time.Hour

var (
	ErrSelfDisable      = errors.New("you can't disable your own account")
	ErrUserDisabled     = errors.New("user is disabled")
	ErrSetupLinkInvalid = errors.New("setup link is invalid, expired or already used")
	ErrWrongPassword    = errors.New("current password is incorrect")
	ErrSamePassword     = errors.New("new password must differ from the current one")
	ErrUserNotFound     = gorm.ErrRecordNotFound
)

// Actor is who performs an account action and from where, for the audit log.
type Actor struct {
	UserID   uint // 0: nobody signed in (login, setup link)
	ClientIP string
}

func (a Actor) id() *uint {
	if a.UserID == 0 {
		return nil
	}
	id := a.UserID
	return &id
}

// SetupLink is a freshly generated link. Token appears only here, never in the database.
type SetupLink struct {
	Path      string    `json:"path"` // /setup/<token>; the client prefixes its own origin
	Purpose   string    `json:"purpose"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RecordEvent writes an audit event inside the request transaction, so it commits or rolls back
// with the change it describes (a "password changed" row for a change that was rolled back would
// be false). Use RecordRejection for attempts that fail.
func RecordEvent(ctx context.Context, ev models.AuthEvent) error {
	ev.CreatedAt = clock.Now()
	return authrepo.InsertEvent(database.From(ctx), ev)
}

// RecordRejection writes an audit event for a rejected attempt through database.Always, so it
// survives the request's rollback. Its failure is logged, never returned.
func RecordRejection(ctx context.Context, ev models.AuthEvent) {
	ev.CreatedAt = clock.Now()
	database.Always(ctx, func(db *gorm.DB) error { return authrepo.InsertEvent(db, ev) })
}

func uintPtr(v uint) *uint { return &v }

// CreateUser creates an account with a username only (no password, so it can't sign in) and
// returns it with its setup link.
func (s *AuthService) CreateUser(ctx context.Context, actor Actor, username string) (models.User, SetupLink, error) {
	username, err := NormalizeUsername(username)
	if err != nil {
		return models.User{}, SetupLink{}, err
	}
	if _, err := authrepo.FindUserByUsername(ctx, username); err == nil {
		return models.User{}, SetupLink{}, ErrUsernameExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, SetupLink{}, err
	}
	user := models.User{Username: username}
	if err := authrepo.CreateUser(ctx, &user); err != nil {
		return models.User{}, SetupLink{}, fmt.Errorf("%w: %v", ErrUserCreation, err)
	}
	if err := RecordEvent(ctx, models.AuthEvent{Event: models.EventUserCreated, UserID: uintPtr(user.ID), ActorID: actor.id(), ClientIP: actor.ClientIP}); err != nil {
		return models.User{}, SetupLink{}, err
	}
	link, err := s.CreateSetupLink(ctx, actor, user.ID)
	if err != nil {
		return models.User{}, SetupLink{}, err
	}
	user.Password = ""
	return user, link, nil
}

// CreateSetupLink makes a new one-time link for userID ("setup" if it has no password yet,
// "reset" otherwise) and revokes the user's older unused links.
func (s *AuthService) CreateSetupLink(ctx context.Context, actor Actor, userID uint) (SetupLink, error) {
	user, err := authrepo.FindUserByID(ctx, userID)
	if err != nil {
		return SetupLink{}, err
	}
	if user.DisabledAt != nil {
		return SetupLink{}, ErrUserDisabled
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return SetupLink{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	purpose := models.SetupPurposeReset
	if user.Password == "" {
		purpose = models.SetupPurposeSetup
	}
	now := clock.Now()
	row := models.SetupToken{
		UserID:    user.ID,
		TokenHash: HashSetupToken(token),
		Purpose:   purpose,
		CreatedBy: actor.id(),
		CreatedAt: now,
		ExpiresAt: now.Add(SetupLinkTTL),
	}
	if err := authrepo.ReplaceSetupToken(ctx, &row, now); err != nil {
		return SetupLink{}, err
	}
	if err := RecordEvent(ctx, models.AuthEvent{Event: models.EventSetupLinkCreated, UserID: uintPtr(user.ID), ActorID: actor.id(), ClientIP: actor.ClientIP, Detail: purpose}); err != nil {
		return SetupLink{}, err
	}
	return SetupLink{Path: "/setup/" + token, Purpose: purpose, ExpiresAt: row.ExpiresAt}, nil
}

// HashSetupToken is how a link's token is stored and looked up.
func HashSetupToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// ListUsers is the admin's user list.
func (s *AuthService) ListUsers(ctx context.Context) ([]authrepo.UserSummary, error) {
	return authrepo.ListUsers(ctx, clock.Now())
}

// SetDisabled disables or re-enables userID. Both sign the user out everywhere; disabling also
// revokes any unused link.
func (s *AuthService) SetDisabled(ctx context.Context, actor Actor, userID uint, disabled bool) (models.User, error) {
	if disabled && userID == actor.UserID {
		return models.User{}, ErrSelfDisable
	}
	var at *time.Time
	event := models.EventUserEnabled
	if disabled {
		now := clock.Now()
		at = &now
		event = models.EventUserDisabled
		if err := authrepo.RevokeSetupTokens(ctx, userID, now); err != nil {
			return models.User{}, err
		}
	}
	user, err := authrepo.SetUserDisabled(ctx, userID, at)
	if err != nil {
		return models.User{}, err
	}
	if err := RecordEvent(ctx, models.AuthEvent{Event: event, UserID: uintPtr(user.ID), ActorID: actor.id(), ClientIP: actor.ClientIP}); err != nil {
		return models.User{}, err
	}
	user.Password = ""
	return user, nil
}

// ListEvents is the admin's view of the audit log.
func (s *AuthService) ListEvents(ctx context.Context, beforeID uint, limit int) ([]authrepo.EventRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return authrepo.ListEvents(ctx, beforeID, limit)
}

// SetupInfo is what the setup page shows before the password is set.
type SetupInfo struct {
	Username  string    `json:"username"`
	Purpose   string    `json:"purpose"`
	ExpiresAt time.Time `json:"expires_at"`
}

// liveSetupToken loads the link for token and its user, or ErrSetupLinkInvalid with the reason
// (for the audit log) and the user it belonged to, if any.
func liveSetupToken(ctx context.Context, token string, forUpdate bool) (models.SetupToken, models.User, string, error) {
	if token == "" || len(token) > 128 {
		return models.SetupToken{}, models.User{}, "malformed", ErrSetupLinkInvalid
	}
	row, err := authrepo.FindSetupToken(ctx, HashSetupToken(token), forUpdate)
	if errors.Is(err, authrepo.ErrNoSetupToken) {
		return models.SetupToken{}, models.User{}, "unknown", ErrSetupLinkInvalid
	}
	if err != nil {
		return models.SetupToken{}, models.User{}, "", err
	}
	user, err := authrepo.FindUserByID(ctx, row.UserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, models.User{}, "user_gone", ErrSetupLinkInvalid
	}
	if err != nil {
		return models.SetupToken{}, models.User{}, "", err
	}
	switch {
	case row.UsedAt != nil:
		return row, user, "used", ErrSetupLinkInvalid
	case row.RevokedAt != nil:
		return row, user, "revoked", ErrSetupLinkInvalid
	case !clock.Now().Before(row.ExpiresAt):
		return row, user, "expired", ErrSetupLinkInvalid
	case user.DisabledAt != nil:
		return row, user, "disabled", ErrSetupLinkInvalid
	}
	return row, user, "", nil
}

// InspectSetupLink tells the setup page whose link it is. A dead link is ErrSetupLinkInvalid
// (not audited: opening the page is not an attempt).
func (s *AuthService) InspectSetupLink(ctx context.Context, token string) (SetupInfo, error) {
	row, user, _, err := liveSetupToken(ctx, token, false)
	if err != nil {
		return SetupInfo{}, err
	}
	return SetupInfo{Username: user.Username, Purpose: row.Purpose, ExpiresAt: row.ExpiresAt}, nil
}

// CompleteSetup sets the password through a link, uses the link up and signs the user in. The
// password is checked and hashed before the database is touched, so the request transaction
// isn't open while bcrypt runs. Every other session of the user ends (token_version bump).
func (s *AuthService) CompleteSetup(ctx context.Context, actor Actor, token, password string) (AuthResult, error) {
	if err := ValidatePassword(password); err != nil {
		return AuthResult{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrPasswordHash, err)
	}
	row, user, reason, err := liveSetupToken(ctx, token, true)
	if errors.Is(err, ErrSetupLinkInvalid) {
		ev := models.AuthEvent{Event: models.EventSetupLinkRejected, ClientIP: actor.ClientIP, Detail: reason}
		if user.ID != 0 {
			ev.UserID = uintPtr(user.ID)
		}
		RecordRejection(ctx, ev)
		return AuthResult{}, err
	}
	if err != nil {
		return AuthResult{}, err
	}
	if ok, err := authrepo.MarkSetupTokenUsed(ctx, row.ID, clock.Now()); err != nil {
		return AuthResult{}, err
	} else if !ok {
		// Used by a concurrent request between our read and the update (FOR UPDATE makes this
		// unreachable on Postgres; kept for SQLite).
		RecordRejection(ctx, models.AuthEvent{Event: models.EventSetupLinkRejected, UserID: uintPtr(user.ID), ClientIP: actor.ClientIP, Detail: "used"})
		return AuthResult{}, ErrSetupLinkInvalid
	}
	user, err = authrepo.SetUserPassword(ctx, user.ID, string(hash))
	if err != nil {
		return AuthResult{}, err
	}
	if err := RecordEvent(ctx, models.AuthEvent{Event: models.EventSetupLinkUsed, UserID: uintPtr(user.ID), ClientIP: actor.ClientIP, Detail: row.Purpose}); err != nil {
		return AuthResult{}, err
	}
	tok, err := s.generateToken(user)
	if err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}
	user.Password = ""
	return AuthResult{Token: tok, User: user}, nil
}

// ChangePassword changes the signed-in user's own password. Every other session ends; the
// returned token (new token_version) keeps the current one signed in.
func (s *AuthService) ChangePassword(ctx context.Context, actor Actor, current, next string) (AuthResult, error) {
	if err := ValidatePassword(next); err != nil {
		return AuthResult{}, err
	}
	if current == next {
		return AuthResult{}, ErrSamePassword
	}
	user, err := authrepo.FindUserByID(ctx, actor.UserID)
	if err != nil {
		return AuthResult{}, err
	}
	if len(current) > PasswordMaxBytes || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(current)) != nil {
		RecordRejection(ctx, models.AuthEvent{Event: models.EventPasswordChangeFailed, UserID: uintPtr(user.ID), ActorID: actor.id(), ClientIP: actor.ClientIP, Detail: "wrong_password"})
		return AuthResult{}, ErrWrongPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(next), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrPasswordHash, err)
	}
	user, err = authrepo.SetUserPassword(ctx, user.ID, string(hash))
	if err != nil {
		return AuthResult{}, err
	}
	if err := RecordEvent(ctx, models.AuthEvent{Event: models.EventPasswordChanged, UserID: uintPtr(user.ID), ActorID: actor.id(), ClientIP: actor.ClientIP}); err != nil {
		return AuthResult{}, err
	}
	tok, err := s.generateToken(user)
	if err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}
	user.Password = ""
	return AuthResult{Token: tok, User: user}, nil
}
