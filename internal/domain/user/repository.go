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
	////////////////////////////////////////////////////////////////////////////////////
	// InsertGoogleUser(ctx context.Context, googleUser *GoogleUser) error
}
