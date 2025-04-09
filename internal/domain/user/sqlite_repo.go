package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/mattn/go-sqlite3"
)

type SQLiteUserRepository struct {
	db db
}

func (r *SQLiteUserRepository) GetByID(ctx context.Context, userID string) (*User, error) {
	query := `SELECT u.id, u.name, u.email, u.password_hash, u.version, u.created_at,
		gu.google_id, gu.name, gu.given_name, gu.family_name, gu.picture, gu.is_verified, gu.created_at 
		FROM users u
		LEFT JOIN google_users_data gu ON gu.user_id = u.id
		WHERE id = ?`

	var userData User
	var createdAt int64

	var gID, gName, gGivenName, gFamilyName, gPicture sql.NullString
	var gVerified sql.NullBool
	var googlCreatedAt sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&userData.ID,
		&userData.Name,
		&userData.Email,
		&userData.Password.Hash,
		&userData.Version,
		&createdAt,
		&gID,
		&gName,
		&gGivenName,
		&gFamilyName,
		&gPicture,
		&gVerified,
		&googlCreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}
	userData.CreatedAt = time.Unix(createdAt, 0)

	if gID.Valid {
		userData.Google = &GoogleUser{
			UserID:        userData.ID,
			ID:            gID.String,
			Email:         userData.Email,
			VerifiedEmail: gVerified.Bool,
			FamilyName:    gFamilyName.String,
			Name:          gName.String,
			GivenName:     gGivenName.String,
			Picture:       gPicture.String,
		}
		if googlCreatedAt.Valid {
			userData.Google.CreatedAt = time.Unix(googlCreatedAt.Int64, 0)
		}
	}

	return &userData, err
}

func (r *SQLiteUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT u.id, u.name, u.email, u.password_hash, u.version, u.created_at,
		gu.google_id, gu.name, gu.given_name, gu.family_name, gu.picture, gu.is_verified, gu.created_at 
		FROM users u
		LEFT JOIN google_users_data gu ON gu.user_id = u.id
		WHERE u.email = ?`

	userData := &User{
		Password: password{},
	}
	var createdAt int64

	var gID, gName, gGivenName, gFamilyName, gPicture sql.NullString
	var gVerified sql.NullBool
	var googlCreatedAt sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&userData.ID,
		&userData.Name,
		&userData.Email,
		&userData.Password.Hash,
		&userData.Version,
		&createdAt,
		&gID,
		&gName,
		&gGivenName,
		&gFamilyName,
		&gPicture,
		&gVerified,
		&googlCreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}
	userData.CreatedAt = time.Unix(createdAt, 0)

	if gID.Valid {
		userData.Google = &GoogleUser{
			UserID:        userData.ID,
			ID:            gID.String,
			Email:         userData.Email,
			VerifiedEmail: gVerified.Bool,
			FamilyName:    gFamilyName.String,
			Name:          gName.String,
			GivenName:     gGivenName.String,
			Picture:       gPicture.String,
		}
		if googlCreatedAt.Valid {
			userData.Google.CreatedAt = time.Unix(googlCreatedAt.Int64, 0)
		}
	}

	return userData, err
}

func (r *SQLiteUserRepository) GetGoogleUser(ctx context.Context, email string) (*GoogleUser, error) {
	query := `SELECT user_id, google_id, email, name, given_name, family_name, picture, is_verified, created_at 
		FROM google_users_data WHERE email = ?`

	var gUser GoogleUser
	var createdAt int64
	var isVerified sql.NullBool

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&gUser.UserID,
		&gUser.ID,
		&gUser.Email,
		&gUser.Name,
		&gUser.GivenName,
		&gUser.FamilyName,
		&gUser.Picture,
		&isVerified,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}
	gUser.VerifiedEmail = isVerified.Bool
	gUser.CreatedAt = time.Unix(createdAt, 0)
	return &gUser, nil
}

func isUniqueConstraintErr(err error) bool {
	if sqliteError, ok := err.(sqlite3.Error); ok {
		if sqliteError.Code == sqlite3.ErrConstraint && sqliteError.ExtendedCode == sqlite3.ErrConstraintUnique {
			return true
		}
	}
	return false
}

func (r *SQLiteUserRepository) Insert(ctx context.Context, user *User) error {
	query := `INSERT INTO users (name, email, password_hash) 
                VALUES (?, ?, ?) 
                RETURNING id, created_at, version`

	args := []interface{}{user.Name, user.Email, user.Password.Hash}

	var createdAt int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &createdAt, &user.Version); err != nil {
		if isUniqueConstraintErr(err) {
			return data.ErrRecordExists
		}
		return err
	}
	user.CreatedAt = time.Unix(createdAt, 0)

	if user.Google != nil {
		query := `INSERT INTO google_users_data
	              (user_id, google_id, email, name, given_name, family_name, picture, is_verified)
	              VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	              RETURNING google_users_data.created_at`

		args := []interface{}{
			user.ID,
			user.Google.ID,
			user.Google.Email,
			user.Google.Name,
			user.Google.GivenName,
			user.Google.FamilyName,
			user.Google.Picture,
			user.Google.VerifiedEmail,
		}

		var googleCreatedAt int64
		if err := r.db.QueryRowContext(ctx, query, args...).Scan(&googleCreatedAt); err != nil {
			if isUniqueConstraintErr(err) {
				return data.ErrRecordExists
			}
			return err
		}
		user.Google.CreatedAt = time.Unix(googleCreatedAt, 0)
	}

	return nil
}

func (r *SQLiteUserRepository) InsertGoogleUser(ctx context.Context, googleUser *GoogleUser) error {
	query := `INSERT INTO google_users_data
	              (user_id, google_id, email, name, given_name, family_name, picture, is_verified)
	              VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	              RETURNING google_users_data.created_at`

	args := []interface{}{
		googleUser.UserID,
		googleUser.ID,
		googleUser.Email,
		googleUser.Name,
		googleUser.GivenName,
		googleUser.FamilyName,
		googleUser.Picture,
		googleUser.VerifiedEmail,
	}

	var googleCreatedAt int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&googleCreatedAt); err != nil {
		if isUniqueConstraintErr(err) {
			return data.ErrRecordExists
		}
		return err
	}
	googleUser.CreatedAt = time.Unix(googleCreatedAt, 0)

	return nil
}

func (r *SQLiteUserRepository) Delete(ctx context.Context, userID string) error {
	query := "DELETE FROM users WHERE id = ?"
	res, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with id %s", userID)
	}
	return nil
}

func (r *SQLiteUserRepository) Update(ctx context.Context, user *User) error {
	query := `
        UPDATE users 
        SET name = ?, email = ?, password_hash = ?, version = version + 1 
        WHERE id = ?
        RETURNING version
    `

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.Hash,
		user.ID,
	}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.Version)
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
