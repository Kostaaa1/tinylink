package user

import (
	"context"
)

type Repository interface {
	Insert(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, userID string) (*User, error)
	Update(ctx context.Context, user *User) error
}
