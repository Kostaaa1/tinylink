package authcontext

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/domain/token"
)

type contextKey string

var (
	claimsKey contextKey = "claims"
)

func WithClaims(ctx context.Context, claims *token.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func GetClaims(ctx context.Context) *token.Claims {
	return ctx.Value(claimsKey).(*token.Claims)
}

func ClaimsFromCtx(ctx context.Context) *token.Claims {
	claims, _ := ctx.Value(claimsKey).(*token.Claims)
	return claims
}

func IsAuthenticated(ctx context.Context) bool {
	return ClaimsFromCtx(ctx) != nil
}
