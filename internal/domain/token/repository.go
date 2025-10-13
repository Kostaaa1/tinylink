package token

import (
	"context"
	"time"
)

type Repository interface {
	Save(ctx context.Context, userID uint64, refreshToken string, ttl time.Duration) error
	Revoke(ctx context.Context, userID uint64) error
	Valid(ctx context.Context, userID uint64, refreshToken string) error
}
