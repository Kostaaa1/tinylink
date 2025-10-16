package user

import (
	"context"
)

type Repository interface {
	Insert(ctx context.Context, user *User) error
	InsertGoogleUser(ctx context.Context, googleUser *GoogleUser) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uint64) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, userID string) error
}
