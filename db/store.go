package db

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/data"
)

type RedisStore struct {
	Tinylink TinylinkStore
	Token    TokenStore
}

type SQLiteStore struct {
	Tinylink TinylinkStore
	User     UserStore
}

type TokenStore interface {
	Store(ctx context.Context, token *data.Token) error
	Get(ctx context.Context, tokenText string) (*data.Token, error)
	RevokeAll(ctx context.Context, userID string, scope *data.Scope) error
}

type TinylinkStore interface {
	List(ctx context.Context, userID string) ([]*data.Tinylink, error)
	Delete(ctx context.Context, userID, id string) error
	Save(ctx context.Context, tl *data.Tinylink) error
	Get(ctx context.Context, userID, id string) (*data.Tinylink, error)
}

type UserStore interface {
	Insert(ctx context.Context, user *data.User) error
	GetByEmail(ctx context.Context, email string) (*data.User, error)
	GetByID(ctx context.Context, userID string) (*data.User, error)
	Update(ctx context.Context, user *data.User) error
}
