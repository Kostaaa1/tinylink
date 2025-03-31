package auth

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/domain/token"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
)

type contextKey string

var (
	SessionKey                     = "tinylink_session"
	userContextKey      contextKey = "user"
	authTokenContextKey contextKey = "auth"
	tempTokenContextKey contextKey = "temp_auth"
)

func TokenFromCtx(ctx context.Context) *token.Token {
	token, _ := ctx.Value(authTokenContextKey).(*token.Token)
	return token
}

func TempTokenFromCtx(ctx context.Context) *user.User {
	token, _ := ctx.Value(tempTokenContextKey).(*user.User)
	return token
}

func UserFromCtx(ctx context.Context) *user.User {
	user, _ := ctx.Value(userContextKey).(*user.User)
	return user
}

func IsAuthenticated(ctx context.Context) bool {
	return UserFromCtx(ctx) != nil
}
