package sqlitedb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/store"
)

type SQLiteUserRepository struct {
	db *sql.DB
}

func NewSqliteUserRepository(db *sql.DB) store.UserRepository {
	return &SQLiteUserRepository{
		db: db,
	}
}

func (s *SQLiteUserRepository) Insert(user *data.User) error {
	query := `INSERT INTO users (name, email, password_hash, activated) 
        VALUES (?, ?, ?, ?)
        RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	args := []interface{}{user.Name, user.Email, user.Password.Hash, user.Activated}

	row := s.db.QueryRowContext(ctx, query, args...)

	var createdAtStr string
	err := row.Scan(&user.ID, &createdAtStr, &user.Version)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.emai") {
			return data.ErrDuplicateEmail
		}
		fmt.Println("error while inserting user: ", err.Error())
		return err
	}
	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)

	return nil
}

func (s *SQLiteUserRepository) GetByID(id int64) (*data.User, error) {
	query := `SELECT id, created_at, name, email, password_hash, activated, version 
        FROM users 
        WHERE id = ?`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var user data.User

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &user, err
}

func (s *SQLiteUserRepository) Update(user *data.User) error {
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
		fmt.Println("error while running update query: ", err)
		switch {
		case err == sql.ErrNoRows:
			return data.ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *SQLiteUserRepository) GetByEmail(email string) (*data.User, error) {
	query := `SELECT id, created_at, name, email, password_hash, activated, version 
        FROM users 
        WHERE email = ?`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var user data.User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	return &user, err
}
