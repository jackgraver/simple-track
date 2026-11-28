package authrepo

import (
	"context"
	"errors"
	"time"

	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserSummary is a row of the admin's user list.
type UserSummary struct {
	ID            uint       `json:"id"`
	Username      string     `json:"username"`
	IsAdmin       bool       `json:"is_admin"`
	HasPassword   bool       `json:"has_password"`
	CreatedAt     time.Time  `json:"created_at"`
	DisabledAt    *time.Time `json:"disabled_at"`
	LinkExpiresAt *time.Time `json:"link_expires_at"`
	LinkPurpose   *string    `json:"link_purpose"`
	LastLoginAt   *time.Time `json:"last_login_at"`
}

// ListUsers returns every live user, oldest first. Needs the admin identity (the setup-token
// and auth-event columns are admin-only under RLS).
func ListUsers(ctx context.Context, now time.Time) ([]UserSummary, error) {
	var rows []UserSummary
	err := database.From(ctx).Raw(`
SELECT u.id, u.username, u.is_admin, u.password <> '' AS has_password, u.created_at, u.disabled_at,
       t.expires_at AS link_expires_at, t.purpose AS link_purpose,
       (SELECT max(e.created_at) FROM auth_events e WHERE e.user_id = u.id AND e.event IN ?) AS last_login_at
FROM users u
LEFT JOIN LATERAL (
  SELECT expires_at, purpose FROM user_setup_tokens s
  WHERE s.user_id = u.id AND s.used_at IS NULL AND s.revoked_at IS NULL AND s.expires_at > ?
  ORDER BY s.created_at DESC LIMIT 1
) t ON true
WHERE u.deleted_at IS NULL
ORDER BY u.id`, []string{models.EventLoginOK, models.EventSetupLinkUsed}, now).Scan(&rows).Error
	return rows, err
}

// SetUserPassword stores a new bcrypt hash and bumps token_version (every session signs out),
// returning the updated user.
func SetUserPassword(ctx context.Context, id uint, hash string) (models.User, error) {
	return bumpUser(ctx, id, map[string]any{"password": hash})
}

// SetUserDisabled disables (at) or enables (nil) a user and bumps token_version either way.
func SetUserDisabled(ctx context.Context, id uint, at *time.Time) (models.User, error) {
	return bumpUser(ctx, id, map[string]any{"disabled_at": at})
}

func bumpUser(ctx context.Context, id uint, fields map[string]any) (models.User, error) {
	fields["token_version"] = gorm.Expr("token_version + 1")
	fields["updated_at"] = clock.Now()
	res := database.From(ctx).Model(&models.User{}).Where("id = ?", id).Updates(fields)
	if res.Error != nil {
		return models.User{}, res.Error
	}
	if res.RowsAffected == 0 {
		return models.User{}, gorm.ErrRecordNotFound
	}
	return FindUserByID(ctx, id)
}

// ReplaceSetupToken revokes the user's unused links and stores a new one. Admin identity.
func ReplaceSetupToken(ctx context.Context, token *models.SetupToken, now time.Time) error {
	db := database.From(ctx)
	if err := db.Model(&models.SetupToken{}).
		Where("user_id = ? AND used_at IS NULL AND revoked_at IS NULL", token.UserID).
		Update("revoked_at", now).Error; err != nil {
		return err
	}
	return db.Omit(clause.Associations).Create(token).Error
}

// RevokeSetupTokens revokes every unused link of a user (on disable). Admin identity.
func RevokeSetupTokens(ctx context.Context, userID uint, now time.Time) error {
	return database.From(ctx).Model(&models.SetupToken{}).
		Where("user_id = ? AND used_at IS NULL AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

// ErrNoSetupToken: no row has that hash (or RLS hides it).
var ErrNoSetupToken = errors.New("setup token not found")

// FindSetupToken loads the link with tokenHash, locked for update when forUpdate. Without the
// admin identity, row-level security shows a token only to the request that names its hash in
// app.setup_token_hash, which this sets for the rest of the request transaction.
func FindSetupToken(ctx context.Context, tokenHash string, forUpdate bool) (models.SetupToken, error) {
	db := database.From(ctx)
	if db.Dialector.Name() == "postgres" {
		if err := db.Exec("SELECT set_config('app.setup_token_hash', ?, true)", tokenHash).Error; err != nil {
			return models.SetupToken{}, err
		}
	}
	q := database.From(ctx).Where("token_hash = ?", tokenHash)
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var token models.SetupToken
	if err := q.First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.SetupToken{}, ErrNoSetupToken
		}
		return models.SetupToken{}, err
	}
	return token, nil
}

// MarkSetupTokenUsed sets used_at on an unused link; false if it was used meanwhile.
func MarkSetupTokenUsed(ctx context.Context, id uint, now time.Time) (bool, error) {
	res := database.From(ctx).Model(&models.SetupToken{}).
		Where("id = ? AND used_at IS NULL AND revoked_at IS NULL", id).
		Update("used_at", now)
	return res.RowsAffected == 1, res.Error
}

// InsertEvent appends to auth_events. Plain INSERT without RETURNING: under RLS anyone may
// insert but only the admin may read, and RETURNING would need the read policy to pass.
func InsertEvent(db *gorm.DB, ev models.AuthEvent) error {
	return db.Exec(`INSERT INTO auth_events (created_at, event, user_id, actor_id, client_ip, detail) VALUES (?, ?, ?, ?, ?, ?)`,
		ev.CreatedAt, ev.Event, ev.UserID, ev.ActorID, ev.ClientIP, ev.Detail).Error
}

// EventRow is an auth_events row with usernames, for the admin view.
type EventRow struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	Event         string    `json:"event"`
	UserID        *uint     `json:"user_id"`
	Username      *string   `json:"username"`
	ActorID       *uint     `json:"actor_id"`
	ActorUsername *string   `json:"actor_username"`
	ClientIP      string    `json:"client_ip"`
	Detail        string    `json:"detail"`
}

// ListEvents returns up to limit events older than beforeID (0: newest), newest first. Admin
// identity (RLS shows nothing to anyone else).
func ListEvents(ctx context.Context, beforeID uint, limit int) ([]EventRow, error) {
	q := database.From(ctx).Table("auth_events e").
		Select("e.id, e.created_at, e.event, e.user_id, u.username, e.actor_id, a.username AS actor_username, e.client_ip, e.detail").
		Joins("LEFT JOIN users u ON u.id = e.user_id").
		Joins("LEFT JOIN users a ON a.id = e.actor_id").
		Order("e.id DESC").Limit(limit)
	if beforeID > 0 {
		q = q.Where("e.id < ?", beforeID)
	}
	var rows []EventRow
	return rows, q.Scan(&rows).Error
}
