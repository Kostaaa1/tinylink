package sqlitedb

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/store"
	"github.com/jmoiron/sqlx"
)

type SQLiteUserStore struct {
	db *sqlx.DB
}

func NewSQLiteUserStore(db *sqlx.DB) store.UserStore {
	return &SQLiteUserStore{
		db: db,
	}
}

func (s *SQLiteUserStore) GetByEmail(email string) (*data.User, error) {
	query := `SELECT id, created_at, name, email, password_hash, activated, version FROM users WHERE email = ?`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

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

func (s *SQLiteUserStore) Insert(user *data.User) error {
	query := `INSERT INTO users (name, email, password_hash, activated) 
        VALUES (?, ?, ?, ?)
        RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

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

func (s *SQLiteUserStore) GetByID(id int64) (*data.User, error) {
	query := `SELECT id, created_at, name, email, password_hash, activated, version 
        FROM users 
        WHERE id = ?`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var user data.User

	var createdAt int64
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&createdAt,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Activated,
		&user.Version,
	)
	user.CreatedAt = time.Unix(createdAt, 0)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &user, err
}

func (s *SQLiteUserStore) Update(user *data.User) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

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
