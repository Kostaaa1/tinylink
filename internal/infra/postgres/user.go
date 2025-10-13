package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Kostaaa1/tinylink/core/transactor"
	"github.com/Kostaaa1/tinylink/core/transactor/adapters"
	"github.com/Kostaaa1/tinylink/internal/constants"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db adapters.PgxQuerier
}

func NewUserRepository(db adapters.PgxQuerier) user.Repository {
	return &UserRepository{db: db}
}

func (p *UserRepository) WithRepositoryTx(tx transactor.Tx) user.Repository {
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		panic("tx does not match the pgx.Tx")
	}
	return &UserRepository{db: pgxTx}
}

func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*user.User, error) {
	query := `SELECT u.id, u.name, u.email, u.password_hash, u.version, u.created_at,
		gu.google_id, gu.name, gu.given_name, gu.family_name, gu.picture, gu.is_verified, gu.created_at
		FROM users u
		LEFT JOIN google_users_data gu ON gu.user_id = u.id
		WHERE u.id = $1`

	userData := &user.User{}
	var pwHash []byte

	var gID, gName, gGivenName, gFamilyName, gPicture sql.NullString
	var gVerified sql.NullBool
	var googlCreatedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&userData.ID,
		&userData.Name,
		&userData.Email,
		// &userData.Password.Hash,
		&pwHash,
		&userData.Version,
		&userData.CreatedAt,

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
			return nil, constants.ErrNotFound
		}
		return nil, err
	}

	userData.SetPassword(pwHash)

	if gID.Valid {
		userData.Google = &user.GoogleUser{
			UserID:        userData.ID,
			ID:            gID.String,
			Email:         userData.Email,
			VerifiedEmail: gVerified.Bool,
			FamilyName:    gFamilyName.String,
			Name:          gName.String,
			GivenName:     gGivenName.String,
			Picture:       gPicture.String,
			CreatedAt:     googlCreatedAt.Time,
		}
	}

	return userData, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `SELECT u.id, u.name, u.email, u.password_hash, u.version, u.created_at,
		gu.google_id, gu.name, gu.given_name, gu.family_name, gu.picture, gu.is_verified, gu.created_at
		FROM users u
		LEFT JOIN google_users_data gu ON gu.user_id = u.id
		WHERE u.email = $1`

	userData := &user.User{}
	var pwHash []byte

	var gID, gName, gGivenName, gFamilyName, gPicture sql.NullString
	var gVerified sql.NullBool
	var googlCreatedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, email).Scan(
		&userData.ID,
		&userData.Name,
		&userData.Email,
		// &userData.Password.Hash,
		&pwHash,
		&userData.Version,
		&userData.CreatedAt,

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
			return nil, constants.ErrNotFound
		}
		return nil, err
	}

	userData.SetPassword(pwHash)

	if gID.Valid {
		userData.Google = &user.GoogleUser{
			UserID:        userData.ID,
			ID:            gID.String,
			Email:         userData.Email,
			VerifiedEmail: gVerified.Bool,
			FamilyName:    gFamilyName.String,
			Name:          gName.String,
			GivenName:     gGivenName.String,
			Picture:       gPicture.String,
			CreatedAt:     googlCreatedAt.Time,
		}
	}

	return userData, nil
}

// func isUniqueConstraintErr(err error) bool {
// 	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
// 	// if sqliteError, ok := err.(sqlite3.Error); ok {
// 	// 	if sqliteError.Code == sqlite3.ErrConstraint && sqliteError.ExtendedCode == sqlite3.ErrConstraintUnique {
// 	// 		return true
// 	// 	}
// 	// }
// 	// return false
// }

func (r *UserRepository) Insert(ctx context.Context, user *user.User) error {
	query := `INSERT INTO users (name, email, password_hash)
            VALUES ($1, $2, $3)
            RETURNING id, created_at, version`

	args := []interface{}{user.Name, user.Email, user.Password.Hash}
	if err := r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version); err != nil {
		// if isUniqueConstraintErr(err) {
		// 	return data.ErrRecordExists
		// }
		return err
	}

	if user.Google != nil {
		query := `INSERT INTO google_users_data
			(user_id, google_id, email, name, given_name, family_name, picture, is_verified)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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

		if err := r.db.QueryRow(ctx, query, args...).Scan(&user.Google.CreatedAt); err != nil {
			// if isUniqueConstraintErr(err) {
			// 	return data.ErrRecordExists
			// }
			return err
		}
	}

	return nil
}

func (r *UserRepository) InsertGoogleUser(ctx context.Context, googleUser *user.GoogleUser) error {
	query := `INSERT INTO google_users_data
			(user_id, google_id, email, name, given_name, family_name, picture, is_verified)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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

	if err := r.db.QueryRow(ctx, query, args...).Scan(&googleUser.CreatedAt); err != nil {
		// if isUniqueConstraintErr(err) {
		// 	return data.ErrRecordExists
		// }
		return err
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, userID string) error {
	query := "DELETE FROM users WHERE id = ?"
	res, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("no user found with id %s", userID)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *user.User) error {
	query := `
        UPDATE users
        SET name = $1, email = $2, password_hash = $3, version = version + 1
        WHERE id = $4
        RETURNING version
    `

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.Hash,
		user.ID,
	}

	err := r.db.QueryRow(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return constants.ErrNotFound
		default:
			return err
		}
	}

	return nil
}
