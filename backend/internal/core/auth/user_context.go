package auth

import (
	"context"

	"github.com/gin-gonic/gin"
)

// User is the signed-in user of a request, as AuthMiddleware resolved it from the users row
// (not from the token), so IsAdmin reflects the current row.
type User struct {
	ID       uint
	Username string
	IsAdmin  bool
}

type userContextKey struct{}

// WithUser returns ctx carrying user. AuthMiddleware calls it; tests can too.
func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

// CurrentUser returns the request's user, if AuthMiddleware set one. A *gin.Context is
// accepted too (its request context is read, whatever the engine's ContextWithFallback).
func CurrentUser(ctx context.Context) (User, bool) {
	if ctx == nil {
		return User{}, false
	}
	if gc, ok := ctx.(*gin.Context); ok {
		if gc.Request == nil {
			return User{}, false
		}
		ctx = gc.Request.Context()
	}
	user, ok := ctx.Value(userContextKey{}).(User)
	if !ok || user.ID == 0 {
		return User{}, false
	}
	return user, true
}

// UserID returns the request's user ID, or false when there is no signed-in user.
// Callers must treat false as unauthenticated (401), never fall back to a default user.
func UserID(ctx context.Context) (uint, bool) {
	user, ok := CurrentUser(ctx)
	return user.ID, ok
}

// MustUser returns the request's user and panics if there is none. Use it only in code that
// runs behind AuthMiddleware, where a missing user is a programming error.
func MustUser(ctx context.Context) User {
	user, ok := CurrentUser(ctx)
	if !ok {
		panic("auth.MustUser: no user on the context (route not behind AuthMiddleware?)")
	}
	return user
}
