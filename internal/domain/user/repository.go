package user

import (
	"context"
	"database/sql"
)

type db interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type UserRepository interface {
	Insert(ctx context.Context, user *User) error
	InsertGoogleUser(ctx context.Context, googleUser *GoogleUser) error
	Exists(ctx context.Context, email string, checkGoogle bool) (bool, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, userID string) error
	////////////////////////////////////////////////////////////////////////////////////
	// GetByID(ctx context.Context, userID string) (*User, error)
}
