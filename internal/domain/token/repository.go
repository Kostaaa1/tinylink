package token

import (
	"context"
)

type TokenRepository interface {
	Save(ctx context.Context, userID, tokenID string) error
	Delete(ctx context.Context, userID string) error
	Valid(ctx context.Context, userID string, token string) error
}
