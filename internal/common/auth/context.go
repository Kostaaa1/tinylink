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

// func TokenFromCtx(ctx context.Context) *token.Token {
// 	token, _ := ctx.Value(authTokenContextKey).(*token.Token)
// 	return token
// }

// func TempTokenFromCtx(ctx context.Context) *user.User {
// 	token, _ := ctx.Value(tempTokenContextKey).(*user.User)
// 	return token
// }

// func UserFromCtx(ctx context.Context) *user.User {
// 	user, _ := ctx.Value(userContextKey).(*user.User)
// 	return user
// }

func ClaimsFromCtx(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsContextKey).(*Claims)
	return claims
}

func IsAuthenticated(ctx context.Context) bool {
	return ClaimsFromCtx(ctx) != nil
}
