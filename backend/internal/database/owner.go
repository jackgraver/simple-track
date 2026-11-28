package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Owner kit (docs/multi-user/README.md "Ownership", DECISIONS.md "B3 owner kit").
//
// Owned tables: the model embeds Owner, reads go through Owned(ctx), creates are stamped by
// Owner's BeforeCreate hook. Shared catalogs: the model embeds Creator, stamped the same way.
// Row-level security is the backstop under both (it applies to the server's app role).

// ErrNoIdentity is returned when an owned query or create runs without a signed-in user on the
// context: a route outside AuthMiddleware, or a test that forgot WithIdentity.
var ErrNoIdentity = errors.New("database: no user on the context (owned data needs AuthMiddleware or database.WithIdentity)")

type identityKey struct{}

// WithIdentity returns ctx carrying id, for contexts without a request transaction: module tests
// on SQLite (WithDB) and tools. Request code never calls it; AuthMiddleware sets the identity
// through SetIdentity. A request transaction's own identity wins over this one.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, id)
}

// CurrentIdentity is the user owned queries and creates act as: the request's identity (set by
// AuthMiddleware), else the one from WithIdentity. ok is false when there is none, or its user ID
// is 0.
func CurrentIdentity(ctx context.Context) (Identity, bool) {
	if ctx == nil {
		return Identity{}, false
	}
	if gc, ok := ctx.(*gin.Context); ok && gc.Request != nil {
		ctx = gc.Request.Context()
	}
	if h, ok := ctx.Value(handleKey{}).(handle); ok && h.req != nil {
		h.req.mu.Lock()
		id := h.req.identity
		h.req.mu.Unlock()
		if id != nil && id.UserID != 0 {
			return *id, true
		}
	}
	if id, ok := ctx.Value(identityKey{}).(Identity); ok && id.UserID != 0 {
		return id, true
	}
	return Identity{}, false
}

// Owned is the gorm scope for owned tables: only the current user's rows.
//
//	database.From(ctx).Scopes(database.Owned(ctx)).First(&row, id)
//	db.Preload("Items", database.Owned(ctx))   // preloads don't inherit scopes
//
// The column is qualified with the statement's table, so joins don't make it ambiguous. It
// composes with Unscoped() (soft-deleted rows stay filtered by user). Without a user on ctx the
// query fails with ErrNoIdentity rather than running unfiltered.
func Owned(ctx context.Context) func(*gorm.DB) *gorm.DB {
	id, ok := CurrentIdentity(ctx)
	return func(db *gorm.DB) *gorm.DB {
		if !ok {
			_ = db.AddError(ErrNoIdentity)
			return db
		}
		return db.Where(clause.Eq{
			Column: clause.Column{Table: clause.CurrentTable, Name: "user_id"},
			Value:  id.UserID,
		})
	}
}

// Owner is embedded in every owned-table model. Its BeforeCreate hook is the only code that sets
// user_id (BACKLOG #1): request code never assigns UserID, and `json:"-"` keeps it out of request
// bodies and responses.
//
// A model that declares its own BeforeCreate must call Owner.BeforeCreate from it (the rules test
// fails otherwise).
type Owner struct {
	UserID uint `json:"-" gorm:"column:user_id;not null;index"`
}

// BeforeCreate stamps the row with the current user. With no user on the context (seeds, tools on
// the owner role) the caller must have set UserID itself; otherwise the create fails with
// ErrNoIdentity. A UserID that differs from the current user's is refused, never overwritten:
// it means code is trying to create a row for someone else.
func (o *Owner) BeforeCreate(tx *gorm.DB) error {
	return stamp(tx, &o.UserID, "user_id")
}

// Creator is embedded in shared-catalog models (foods, exercises, ...): who added the row.
// Stamped on create like Owner. A model embeds Owner or Creator, never both (the two hooks would
// cancel out; the rules test refuses it). It's visible in responses (the UI can offer "edit" to the creator)
// but request bodies must not bind into a model that has it (rules test).
type Creator struct {
	CreatedBy uint `json:"created_by" gorm:"column:created_by;not null;index"`
}

// BeforeCreate stamps created_by with the current user; same rules as Owner.BeforeCreate.
func (c *Creator) BeforeCreate(tx *gorm.DB) error {
	return stamp(tx, &c.CreatedBy, "created_by")
}

func stamp(tx *gorm.DB, field *uint, column string) error {
	id, ok := CurrentIdentity(tx.Statement.Context)
	switch {
	case ok && *field == 0:
		*field = id.UserID
	case ok && *field != id.UserID:
		return fmt.Errorf("database: %s %d set by hand on create, but the request user is %d", column, *field, id.UserID)
	case !ok && *field == 0:
		return fmt.Errorf("%w (creating a row with %s)", ErrNoIdentity, column)
	}
	return nil
}
