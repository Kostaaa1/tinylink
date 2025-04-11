package token

import (
	"context"
)

type TokenRepository interface {
	Revoke(ctx context.Context, tokenID string) error
	Store(ctx context.Context, tokenID, userID string) error
	GetUserID(ctx context.Context, tokenID string) (string, error)
	TxDelOldAndInsertNew(ctx context.Context, userID, oldToken, newToken string) (string, error)
}
