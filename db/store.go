package db

import (
	"context"
	"time"

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
	New(userID int64, ttl time.Duration, scope data.Scope) *data.Token
	Store(token *data.Token) error
	Get(token string) (*data.Token, error)
	Revoke(token string) error
	Validate(token string) error
	ListActive() ([]*data.Token, error)
}

type TinylinkStore interface {
	List(ctx context.Context, userID string) ([]*data.Tinylink, error)
	Delete(ctx context.Context, userID, id string) error
	Save(ctx context.Context, tl *data.Tinylink, userID string, ttl time.Duration) error
	Get(ctx context.Context, userID, id string) (*data.Tinylink, error)
	Exists(ctx context.Context, id string) (bool, error)
}

type UserStore interface {
	Insert(user *data.User) error
	GetByEmail(email string) (*data.User, error)
	Update(user *data.User) error
}
