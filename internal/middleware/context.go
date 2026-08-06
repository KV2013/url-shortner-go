package middleware

import (
	"context"
)

type contextKey string

const UserIDContextKey contextKey = "userID"

func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(UserIDContextKey).(string)
	return id
}
