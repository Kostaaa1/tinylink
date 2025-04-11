package token

import (
	"context"
	"time"
)

type TokenRepository interface {
	Revoke(ctx context.Context, tokenID string) error
	Store(ctx context.Context, tokenID, userID string, ttl time.Duration) error
	GetUserID(ctx context.Context, tokenID string) (string, error)
	TxDelOldAndInsertNew(ctx context.Context, userID, oldToken, newToken string, ttl time.Duration) (string, error)
}
