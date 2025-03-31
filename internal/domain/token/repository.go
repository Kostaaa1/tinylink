package token

import (
	"context"
)

type Repository interface {
	Store(ctx context.Context, token *Token) error
	Get(ctx context.Context, tokenText string) (*Token, error)
	RevokeAll(ctx context.Context, userID string, scope *Scope) error
}
