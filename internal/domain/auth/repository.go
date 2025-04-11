package auth

import "context"

type TokenRepository interface {
	Store(ctx context.Context, token *RefreshToken) error
	Valid(ctx context.Context, uuid, userID string) error
	Revoke(ctx context.Context, uuid string) error
	RevokeAll(ctx context.Context, userID string) error
}
