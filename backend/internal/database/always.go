package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Writes that survive a rollback (docs/multi-user/DECISIONS.md, "database.Always").
//
// Everything a request writes through From is rolled back when it answers 4xx/5xx. A few writes
// must outlive that outcome, above all audit rows for rejected attempts ("setup link expired").
// Always is the one way to say so, and every use is visible with grep -rn 'database.Always'.

// alwaysTimeout bounds the transaction Always functions run in. It runs after the response, on a
// context the client can't cancel, so it needs its own limit (the app role's statement_timeout
// bounds each statement, not the whole transaction).
const alwaysTimeout = 5 * time.Second

// Always registers fn to run once the request transaction has committed or rolled back (also
// after a panic), in a short transaction of its own with the request's identity, so row-level
// security applies as usual. Several calls run in order, in that one transaction. fn's error, or
// panic, is logged and changes nothing else: the response is already decided, and fn gets no
// result back to the handler.
//
// fn can't rely on rows the request wrote (they may have been rolled back), and must use the db
// it is given, not From(ctx). Without a request transaction (WithDB: OutsideRequestTx routes,
// tools, tests) fn runs right away, in a transaction on that db.
//
// Use it only for writes that must outlive the request's outcome (audit and security events).
func Always(ctx context.Context, fn func(db *gorm.DB) error) {
	if gc, ok := ctx.(*gin.Context); ok && gc.Request != nil {
		ctx = gc.Request.Context()
	}
	h, ok := ctx.Value(handleKey{}).(handle)
	switch {
	case !ok:
		panic("database.Always: no database handle on the context (route outside database.Middleware, or a test without database.WithDB)")
	case h.db != nil:
		if err := runAlways(context.WithoutCancel(ctx), h.db, nil, []func(*gorm.DB) error{fn}); err != nil {
			log.Printf("database: always: %v", err)
		}
		return
	}
	h.req.mu.Lock()
	defer h.req.mu.Unlock()
	h.req.always = append(h.req.always, fn)
}

// RunAlways runs the functions registered with Always, once, after the transaction has ended
// (it rolls back an unfinished one first). Middleware calls it; other WithTx owners call it
// after Commit or Rollback. Errors are returned for logging only.
func (r *Request) RunAlways() error {
	_ = r.Rollback() // no-op once committed or rolled back
	r.mu.Lock()
	fns := r.always
	r.always = nil
	identity := r.identity
	r.mu.Unlock()
	if len(fns) == 0 {
		return nil
	}
	return runAlways(context.WithoutCancel(r.ctx), r.root, identity, fns)
}

func runAlways(ctx context.Context, db *gorm.DB, identity *Identity, fns []func(*gorm.DB) error) error {
	ctx, cancel := context.WithTimeout(ctx, alwaysTimeout)
	defer cancel()
	var errs []error
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if identity != nil {
			if err := applyIdentity(tx, *identity); err != nil {
				return fmt.Errorf("set identity: %w", err)
			}
		}
		// Each function in its own savepoint, so one that fails (on Postgres a failed statement
		// aborts the transaction) doesn't take the others' writes with it.
		for i, fn := range fns {
			if err := tx.Transaction(func(sp *gorm.DB) error { return callAlways(sp, fn) }); err != nil {
				errs = append(errs, fmt.Errorf("function %d of %d: %w", i+1, len(fns), err))
			}
		}
		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// callAlways runs fn, turning a panic into an error so it rolls back only fn's savepoint.
func callAlways(db *gorm.DB, fn func(*gorm.DB) error) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic: %v", p)
		}
	}()
	return fn(db.Session(&gorm.Session{NewDB: true}))
}
