package userctx

import "context"

type ctxKey struct{}

// WithUserID returns a copy of ctx carrying the given user ID.
func WithUserID(ctx context.Context, id uint64) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// UserID returns the user ID carried by ctx, if any.
func UserID(ctx context.Context) (uint64, bool) {
	id, ok := ctx.Value(ctxKey{}).(uint64)
	return id, ok
}
