package sqlitedb

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/jmoiron/sqlx"
)

type SQLiteUserStore struct {
	db *sqlx.DB
}

func (s *SQLiteUserStore) GetByID(ctx context.Context, userID string) (*data.User, error) {
	query := `SELECT id, created_at, name, email, password_hash, activated, version FROM users WHERE id = ?`

	var user data.User
	var createdAt int64

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&createdAt,
		&user.Name,
		&user.Email,
		&user.Password.Hash,
		&user.Activated,
		&user.Version,
	)
	user.CreatedAt = time.Unix(createdAt, 0)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	return &user, err
}

func (s *SQLiteUserStore) GetByEmail(ctx context.Context, email string) (*data.User, error) {
	query := `SELECT id, created_at, name, email, password_hash, activated, version FROM users WHERE email = ?`

	var user data.User
	var createdAt int64

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&createdAt,
		&user.Name,
		&user.Email,
		&user.Password.Hash,
		&user.Activated,
		&user.Version,
	)
	user.CreatedAt = time.Unix(createdAt, 0)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	return &user, err
}

func (s *SQLiteUserStore) Insert(ctx context.Context, user *data.User) error {
	query := `INSERT INTO users (name, email, password_hash, activated) 
        VALUES (?, ?, ?, ?)
        RETURNING id, created_at, version`

	args := []interface{}{user.Name, user.Email, user.Password.Hash, user.Activated}

	row := s.db.QueryRowContext(ctx, query, args...)

	var createdAt int64
	err := row.Scan(&user.ID, &createdAt, &user.Version)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.emai") {
			return data.ErrDuplicateEmail
		}
		return err
	}
	user.CreatedAt = time.Unix(createdAt, 0)

	return nil
}

func (s *SQLiteUserStore) Update(ctx context.Context, user *data.User) error {
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
