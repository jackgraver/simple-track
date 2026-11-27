package database

import (
	"errors"

	"gorm.io/gorm"
)

// Not found or not yours (docs/multi-user/README.md "Ownership"): a row that doesn't exist and a
// row that belongs to someone else get the same answer, a 404 with the same body. Owned(ctx) and
// row-level security both make the other user's row invisible, so gorm reports it like a missing
// one. These helpers keep update/delete paths and handlers saying the same thing.

// ErrNotFound is gorm.ErrRecordNotFound: the sentinel for "not found or not yours". Existing
// errors.Is(err, gorm.ErrRecordNotFound) checks keep working.
var ErrNotFound = gorm.ErrRecordNotFound

// IsNotFound reports whether err means the row doesn't exist or isn't the caller's.
func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// Affected turns an update or delete by ID into the same error a lookup gives: res.Error, or
// ErrNotFound when no row was touched (missing, someone else's, or already deleted).
//
//	return database.Affected(db.Scopes(database.Owned(ctx)).Delete(&Row{}, id))
func Affected(res *gorm.DB) error {
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
