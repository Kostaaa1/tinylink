package user

import (
	"context"

	"github.com/Kostaaa1/tinylink/core/transactor"
)

type Repository interface {
	Insert(ctx context.Context, user *User) error
	InsertGoogleUser(ctx context.Context, googleUser *GoogleUser) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uint64) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, userID string) error
	WithTx(tx transactor.Tx) Repository
	////////////////////////////////////////////////////////////////////////////////////
	// GetByID(ctx context.Context, userID string) (*User, error)
}
