package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
)

// NOT USED //,,
// func NewSQLiteRepository(db db) *SQLiteUserRepository {
// 	return &SQLiteUserRepository{db: db}
// }

type db interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

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

// func (r *SQLiteUserRepository) HandleGoogleLogin(ctx context.Context, gUser *GoogleUser) (UserDTO, error) {
// 	return UserDTO{}, nil

// newUserDTO := UserDTO{}

// userData := &User{
// 	Email: gUser.Email,
// 	Name:  gUser.Name,
// }

// var userCreatedAt int64

// query := `SELECT id, created_at FROM users WHERE email = ?`
// err := r.db.QueryRowContext(ctx, query, gUser.Email).Scan(&userData.ID, &userCreatedAt)
// if err != nil {
// 	if err == sql.ErrNoRows {
// 		insertQuery := "INSERT INTO users (name, email) VALUES (?, ?) RETURNING id, created_at"
// 		r.db.QueryRowContext(ctx,
// 			insertQuery,
// 			userData.Name,
// 			userData.Email,
// 		).Scan(&userData.ID, &userCreatedAt)
// 	} else {
// 		return err
// 	}
// }

// var googleCreatedAt int64

// queryGoogleUser := "SELECT * FROM google_users_data WHERE id = ?"
// err = r.db.QueryRowContext(ctx, queryGoogleUser, userData.ID).Err()
// if err != nil {
// 	if err == sql.ErrNoRows {
// 		insertQuery := `INSERT INTO google_users_data
// 							(user_id, email, google_id, name, given_name, family_name, picture, is_verified)
// 							VALUES
// 							(?, ?, ?, ?, ?, ?, ?, ?)
// 							RETURNING created_at`

// 		if err := r.db.QueryRowContext(ctx,
// 			insertQuery,
// 			userData.ID,
// 			gUser.Email,
// 			gUser.ID,
// 			gUser.Name,
// 			gUser.GivenName,
// 			gUser.FamilyName,
// 			gUser.Picture,
// 			gUser.VerifiedEmail,
// 		).Scan(&googleCreatedAt); err != nil {
// 			return fmt.Errorf("failed to insert Google user data: %w", err)
// 		}
// 	} else {
// 		return fmt.Errorf("failed to query Google user: %w", err)
// 	}
// } else {
// 	updateQuery := `UPDATE google_users_data
// 		SET
// 		google_id = ?,
// 		email = ?,
// 		name = ?,
// 		given_name = ?,
// 		family_name = ?,
// 		picture = ?,
// 		is_verified = ?
// 		WHERE user_id = ?
// 		RETURNING created_at`

// 	if err := r.db.QueryRowContext(
// 		ctx,
// 		updateQuery,
// 		gUser.ID,
// 		gUser.Email,
// 		gUser.Name,
// 		gUser.GivenName,
// 		gUser.FamilyName,
// 		gUser.Picture,
// 		gUser.VerifiedEmail,
// 		userData.ID,
// 	).Scan(&googleCreatedAt); err != nil {
// 		return fmt.Errorf("failed to update Google user data: %w", err)
// 	}
// }

// userData.CreatedAt = time.Unix(userCreatedAt, 0)
// gUser.CreatedAt = time.Unix(googleCreatedAt, 0)
// userData.Google = gUser

// newUserDTO = user.NewUserDTO(userData)

// return nil
// // })

// if err != nil {
// 	return newUserDTO, err
// }

// return newUserDTO, nil
// }

// func (r *SQLiteUserRepository) HandleGoogleLogin(ctx context.Context, gUser *GoogleUser) (UserDTO, error) {
// 	tx, err := r.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return UserDTO{}, err
// 	}
// 	defer tx.Rollback()

// 	query := `SELECT id FROM users WHERE email = ?`

// 	userData := &User{
// 		Email: gUser.Email,
// 		Name:  gUser.Name,
// 	}

// 	err = tx.QueryRowContext(ctx, query, gUser.Email).Scan(&userData.ID)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			query := `INSERT INTO users (name, email) VALUES (?, ?) RETURNING id, created_at`
// 			var createdAt int64
// 			err := tx.QueryRowContext(
// 				ctx,
// 				query,
// 				gUser.Name,
// 				gUser.Email,
// 			).Scan(&userData.ID, &createdAt)
// 			if err != nil {
// 				return UserDTO{}, err
// 			}
// 			userData.CreatedAt = time.Unix(createdAt, 0)
// 		}
// 	} else {
// 		return UserDTO{}, nil
// 	}

// 	query = `SELECT * FROM google_users_data WHERE email = ? OR user_id = ?`
// 	err = tx.QueryRowContext(ctx, query, gUser.Email, userData.ID).Err()

// 	var googleCreatedAt int64

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			insertQuery := `INSERT INTO google_users_data
// 				(user_id, email, google_id, name, given_name, family_name, picture, is_verified)
// 				VALUES
// 				(?, ?, ?, ?, ?, ?, ?, ?)
// 				RETURNING created_at`

// 			if err := tx.QueryRowContext(
// 				ctx,
// 				insertQuery,
// 				userData.ID,
// 				gUser.Email,
// 				gUser.ID,
// 				gUser.Name,
// 				gUser.GivenName,
// 				gUser.FamilyName,
// 				gUser.Picture,
// 				gUser.VerifiedEmail,
// 			).Scan(&googleCreatedAt); err != nil {
// 				return UserDTO{}, fmt.Errorf("failed to insert Google user data: %w", err)
// 			}
// 		} else {
// 			return UserDTO{}, fmt.Errorf("failed to query Google user: %w", err)
// 		}
// 	} else {
// 		updateQuery := `UPDATE google_users_data
// 			SET
// 			google_id = ?,
// 			email = ?,
// 			name = ?,
// 			given_name = ?,
// 			family_name = ?,
// 			picture = ?,
// 			is_verified = ?
// 			WHERE user_id = ?
// 			RETURNING created_at`

// 		if err := tx.QueryRowContext(
// 			ctx,
// 			updateQuery,
// 			gUser.ID,
// 			gUser.Email,
// 			gUser.Name,
// 			gUser.GivenName,
// 			gUser.FamilyName,
// 			gUser.Picture,
// 			gUser.VerifiedEmail,
// 			userData.ID,
// 		).Scan(&googleCreatedAt); err != nil {
// 			return UserDTO{}, fmt.Errorf("failed to update Google user data: %w", err)
// 		}
// 	}
// 	if err = tx.Commit(); err != nil {
// 		return UserDTO{}, fmt.Errorf("failed to commit transaction: %w", err)
// 	}
// 	gUser.CreatedAt = time.Unix(googleCreatedAt, 0)
// 	userData.Google = gUser
// 	return user.NewUserDTO(userData), nil
// }

func (r *SQLiteUserRepository) Insert(ctx context.Context, user *User) error {
	query := `INSERT INTO users (name, email, password_hash) 
                VALUES (?, ?, ?) 
                RETURNING id, created_at, version`

	args := []interface{}{user.Name, user.Email, user.Password.Hash}

	row := r.db.QueryRowContext(ctx, query, args...)

	var createdAt int64
	if err := row.Scan(&user.ID, &createdAt, &user.Version); err != nil {
		return err
	}
	user.CreatedAt = time.Unix(createdAt, 0)

	// if user.Google != nil {
	// 	query2 := `INSERT INTO google_users_data
	//               (user_id, google_id, email, name, given_name, family_name, picture, is_verified)
	//               VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	//               RETURNING google_users_data.created_at`
	// 	args2 := []interface{}{
	// 		user.ID,
	// 		user.Google.ID,
	// 		user.Google.Email,
	// 		user.Google.Name,
	// 		user.Google.GivenName,
	// 		user.Google.FamilyName,
	// 		user.Google.Picture,
	// 		user.Google.VerifiedEmail,
	// 	}
	// 	var googleCreatedAt int64
	// 	if err := row2.Scan(&googleCreatedAt); err != nil {
	// 		return err
	// 	}
	// 	user.Google.CreatedAt = time.Unix(googleCreatedAt, 0)
	// }

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
