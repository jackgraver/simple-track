package database

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Per-request database handle (docs/multi-user/README.md, "Per-request database handle").
//
// Middleware gives every request a lazy transaction. Request code reaches the database only
// through From(ctx): the first call begins the transaction and sets the request's identity
// with set_config(..., true), which is transaction-local, so it never leaks to the next user
// of the pooled connection. A request that never calls From never holds a connection.
//
// The transaction commits just before the response is written when the status is below 400,
// and rolls back on 4xx/5xx, on a panic, and when the request context ends first (client gone).
// A nested db.Transaction(...) inside a request is a savepoint.

// Identity is what row-level security sees for a request: app.user_id and app.is_admin.
type Identity struct {
	UserID  uint
	IsAdmin bool
}

// ErrRequestDone is the error From's handle carries once the request transaction has been
// committed or rolled back (the response is already on its way).
var ErrRequestDone = errors.New("database: request transaction already finished")

// Request is one request's lazy transaction. Create it with WithTx.
type Request struct {
	mu       sync.Mutex
	root     *gorm.DB
	ctx      context.Context
	identity *Identity
	tx       *gorm.DB
	begunErr error
	done     bool
}

// handle is the context value: a request transaction or, for tools and tests, a plain *gorm.DB.
type handle struct {
	req *Request
	db  *gorm.DB
}

type handleKey struct{}

// WithTx returns ctx carrying a lazy transaction on root. The caller owns the Request and must
// end it with Commit or Rollback (Middleware does both).
func WithTx(ctx context.Context, root *gorm.DB) (context.Context, *Request) {
	req := &Request{root: root, ctx: ctx}
	return context.WithValue(ctx, handleKey{}, handle{req: req}), req
}

// WithDB returns ctx whose From is db itself, with no request transaction and no identity.
// For tools, tests and the auth lookup that runs before the request has a user; never for
// request handlers.
func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, handleKey{}, handle{db: db})
}

// From returns the database handle for ctx: the request transaction (begun on first use) or the
// db given to WithDB. It panics when ctx carries neither, which is a wiring bug (a route outside
// Middleware, or a test that forgot WithDB). A *gin.Context is read through its request context.
//
// If the transaction can't begin, or has already finished, the returned handle carries that
// error, so the query fails like any other database error.
func From(ctx context.Context) *gorm.DB {
	if gc, ok := ctx.(*gin.Context); ok && gc.Request != nil {
		ctx = gc.Request.Context()
	}
	h, ok := ctx.Value(handleKey{}).(handle)
	switch {
	case !ok:
		panic("database.From: no database handle on the context (route outside database.Middleware, or a test without database.WithDB)")
	case h.db != nil:
		return h.db.WithContext(ctx)
	}
	tx, err := h.req.begin()
	if err != nil {
		errDB := h.req.root.Session(&gorm.Session{NewDB: true, Context: ctx})
		_ = errDB.AddError(err)
		return errDB
	}
	return tx.Session(&gorm.Session{NewDB: true, Context: ctx})
}

// SetIdentity sets the request's identity: AuthMiddleware calls it once it knows the user. It
// applies to the transaction when it begins, or right away if it already has. ctx without a
// request transaction (WithDB, or none) is a no-op.
func SetIdentity(ctx context.Context, id Identity) error {
	if gc, ok := ctx.(*gin.Context); ok && gc.Request != nil {
		ctx = gc.Request.Context()
	}
	h, ok := ctx.Value(handleKey{}).(handle)
	if !ok || h.req == nil {
		return nil
	}
	return h.req.setIdentity(id)
}

func (r *Request) setIdentity(id Identity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.identity = &id
	if r.tx == nil || r.done {
		return nil
	}
	return applyIdentity(r.tx, id)
}

func (r *Request) begin() (*gorm.DB, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case r.done:
		return nil, ErrRequestDone
	case r.begunErr != nil:
		return nil, r.begunErr
	case r.tx != nil:
		return r.tx, nil
	}
	// BeginTx with the request context: database/sql rolls the transaction back by itself if
	// the request ends (client disconnect, server shutdown) before Commit.
	tx := r.root.WithContext(r.ctx).Begin()
	if tx.Error != nil {
		r.begunErr = tx.Error
		return nil, tx.Error
	}
	if r.identity != nil {
		if err := applyIdentity(tx, *r.identity); err != nil {
			tx.Rollback()
			r.begunErr = err
			return nil, err
		}
	}
	r.tx = tx
	return tx, nil
}

// applyIdentity sets app.user_id and app.is_admin for the rest of the transaction. SQLite (module
// tests) has neither set_config nor row-level security, so it is skipped there.
func applyIdentity(tx *gorm.DB, id Identity) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	return tx.Exec("SELECT set_config('app.user_id', ?, true), set_config('app.is_admin', ?, true)",
		strconv.FormatUint(uint64(id.UserID), 10), strconv.FormatBool(id.IsAdmin)).Error
}

// Begun reports whether the transaction has started (tests use it to check that a request
// didn't hold a connection).
func (r *Request) Begun() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tx != nil
}

// Commit commits the transaction if it began. Later From calls get ErrRequestDone.
func (r *Request) Commit() error { return r.finish(true) }

// Rollback rolls the transaction back if it began. Later From calls get ErrRequestDone.
func (r *Request) Rollback() error { return r.finish(false) }

func (r *Request) finish(commit bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done {
		return nil
	}
	r.done = true
	if r.tx == nil {
		return nil
	}
	if commit {
		return r.tx.Commit().Error
	}
	return r.tx.Rollback().Error
}

// CreateInSavepoint inserts value in a savepoint, for callers that recover from the insert
// failing (a unique violation from a concurrent replay, then reading the winner's row). On
// Postgres a failed statement aborts the whole transaction; the savepoint keeps the request's
// transaction usable for that follow-up read.
func CreateInSavepoint(db *gorm.DB, value any) error {
	return db.Transaction(func(sp *gorm.DB) error { return sp.Create(value).Error })
}
