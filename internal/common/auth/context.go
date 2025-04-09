package auth

import (
	"context"
)

type contextKey string

var (
	SessionKey                     = "tinylink_session"
	userContextKey      contextKey = "user"
	claimsContextKey    contextKey = "claims"
	authTokenContextKey contextKey = "token"
	tempTokenContextKey contextKey = "temp_token"
)

func ClaimsFromCtx(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsContextKey).(*Claims)
	return claims
}

func IsAuthenticated(ctx context.Context) bool {
	return ClaimsFromCtx(ctx) != nil
}
