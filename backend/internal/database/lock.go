package database

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Advisory locks per user (BACKLOG #2). A lock key without the user makes two users wait on each
// other (two first gym loads on the same date) and, worse, lets a check-then-insert under that
// lock see the other user's row. Every key is built by LockKey, which starts with the user ID;
// the rules test forbids pg_advisory calls outside this package and internal/migrate.

// LockKey is the advisory-lock key for parts, scoped to the current user:
// "u<user id>:" + parts joined by ":" (e.g. "u7:workout_logs.date:2026-01-15"). It fails with
// ErrNoIdentity when ctx has no user.
func LockKey(ctx context.Context, parts ...string) (string, error) {
	id, ok := CurrentIdentity(ctx)
	if !ok {
		return "", fmt.Errorf("%w (advisory lock %q)", ErrNoIdentity, strings.Join(parts, ":"))
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("database: LockKey needs at least one part")
	}
	return fmt.Sprintf("u%d:%s", id.UserID, strings.Join(parts, ":")), nil
}

// XactLock takes pg_advisory_xact_lock on LockKey(ctx, parts...) in db's transaction (inside a
// request, until the request ends; see README "Gotchas"). On SQLite (module tests) it only checks
// the key can be built, since SQLite has no advisory locks and serializes writers anyway.
//
// db must be the transaction the guarded check and write run in (the tx of a Transaction
// callback, or From(ctx)); ctx supplies the user.
func XactLock(ctx context.Context, db *gorm.DB, parts ...string) error {
	key, err := LockKey(ctx, parts...)
	if err != nil {
		return err
	}
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	return db.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", key).Error
}
