package authcontext

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/domain/auth"
)

type contextKey string

var (
	claimsKey contextKey = "claims"
)

func WithClaims(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func GetClaims(ctx context.Context) *auth.Claims {
	return ctx.Value(claimsKey).(*auth.Claims)
}

func ClaimsFromCtx(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(claimsKey).(*auth.Claims)
	return claims
}

func IsAuthenticated(ctx context.Context) bool {
	return ClaimsFromCtx(ctx) != nil
}
