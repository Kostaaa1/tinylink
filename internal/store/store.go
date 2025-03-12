package store

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/data"
)

type Store struct {
	User     UserRepository
	Tinylink TinylinkRepository
}

type TinylinkRepository interface {
	List(ctx context.Context, qp data.QueryParams) ([]*data.Tinylink, error)
	Delete(ctx context.Context, qp data.QueryParams) error
	Save(ctx context.Context, tl *data.Tinylink, qp data.QueryParams) error
	Get(ctx context.Context, qp data.QueryParams) (*data.Tinylink, error)
	Exists(ctx context.Context, id string) (bool, error)
	SetAlias(ctx context.Context, alias string) error
}

type UserRepository interface {
	Insert(user *data.User) error
	GetByEmail(email string) (*data.User, error)
	Update(user *data.User) error
}
