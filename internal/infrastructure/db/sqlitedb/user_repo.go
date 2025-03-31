package sqlitedb

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/jmoiron/sqlx"
)

type SQLiteUserRepository struct {
	db *sqlx.DB
}

func (s *SQLiteUserRepository) GetByID(ctx context.Context, userID string) (*user.User, error) {
	query := `SELECT id, created_at, name, email, password_hash, activated, version FROM users WHERE id = ?`

	var userData user.User
	var createdAt int64

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&userData.ID,
		&createdAt,
		&userData.Name,
		&userData.Email,
		&userData.Password.Hash,
		&userData.Activated,
		&userData.Version,
	)
	userData.CreatedAt = time.Unix(createdAt, 0)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	return &userData, err
}

func (s *SQLiteUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `SELECT id, created_at, name, email, password_hash, activated, version FROM users WHERE email = ?`

	var userData user.User
	var createdAt int64

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&userData.ID,
		&createdAt,
		&userData.Name,
		&userData.Email,
		&userData.Password.Hash,
		&userData.Activated,
		&userData.Version,
	)
	userData.CreatedAt = time.Unix(createdAt, 0)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	return &userData, err
}

func (s *SQLiteUserRepository) Insert(ctx context.Context, userData *user.User) error {
	query := `INSERT INTO users (name, email, password_hash, activated) 
        VALUES (?, ?, ?, ?)
        RETURNING id, created_at, version`

	args := []interface{}{userData.Name, userData.Email, userData.Password.Hash, userData.Activated}

	row := s.db.QueryRowContext(ctx, query, args...)

	var createdAt int64

	err := row.Scan(&userData.ID, &createdAt, &userData.Version)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.emai") {
			return user.ErrDuplicateEmail
		}
		return err
	}
	userData.CreatedAt = time.Unix(createdAt, 0)

	return nil
}

func (s *SQLiteUserRepository) Update(ctx context.Context, user *user.User) error {
	query := `
        UPDATE users 
        SET name = ?, email = ?, password_hash = ?, activated = ?, version = version + 1 
        WHERE id = ? AND version = ? 
        RETURNING version
    `

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password,
		user.Activated,
		user.ID,
		user.Version,
	}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return data.ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}
