package user

import (
	"context"
)

type UserRepository interface {
	Insert(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, userID string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, userID string) error

	// remove
	// HandleGoogleLogin(ctx context.Context, googleUser *GoogleUser) (*User, error)

	// UpdateGoogleUser(ctx context.Context, googleUser *GoogleUser)
	// InsertGoogleUser(ctx context.Context, googleUser *GoogleUser)
}
