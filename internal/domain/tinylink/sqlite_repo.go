package tinylink

import (
	"context"
	"database/sql"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/mattn/go-sqlite3"
)

type TinylinkSQLRepository struct {
	db db
}

func isUniqueConstraintErr(err error) bool {
	if sqliteError, ok := err.(sqlite3.Error); ok {
		if sqliteError.Code == sqlite3.ErrConstraint && sqliteError.ExtendedCode == sqlite3.ErrConstraintUnique {
			return true
		}
	}
	return false
}

func (s *TinylinkSQLRepository) Update(ctx context.Context, tl *Tinylink) error {
	query := `UPDATE tinylinks SET alias = ?, domain = ?, is_private = ? WHERE id = ? 
	RETURNING user_id, original_url, usage_count, domain, created_at`

	args := []interface{}{tl.Alias, tl.Domain, tl.Private, tl.ID}
	var createdAt int64
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&tl.UserID,
		&tl.OriginalURL,
		&tl.UsageCount,
		&tl.Domain,
		&createdAt,
	)
	tl.CreatedAt = time.Unix(createdAt, 0)

	if err != nil {
		if err == sql.ErrNoRows {
			return data.ErrNotFound
		}
		if isUniqueConstraintErr(err) {
			return ErrAliasExists
		}
		return err
	}

	return nil
}

func (s *TinylinkSQLRepository) Insert(ctx context.Context, tl *Tinylink) error {
	query := `INSERT INTO tinylinks (user_id, alias, original_url, domain, is_private) 
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, created_at`

	args := []interface{}{tl.UserID, tl.Alias, tl.OriginalURL, tl.Domain, tl.Private}

	var createdAt int64
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&tl.ID, &createdAt)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return ErrAliasExists
		}
		return err
	}
	tl.CreatedAt = time.Unix(createdAt, 0)

	return nil
}

func (s *TinylinkSQLRepository) List(ctx context.Context, userID string) ([]*Tinylink, error) {
	query := `
		SELECT id, user_id, alias, original_url, is_private, usage_count, domain, created_at
		FROM tinylinks
		WHERE user_id = ?
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tinylinks := []*Tinylink{}
	for rows.Next() {
		tl := &Tinylink{}

		var createdAt int64
		if err := rows.Scan(
			&tl.ID,
			&tl.UserID,
			&tl.Alias,
			&tl.OriginalURL,
			&tl.Private,
			&tl.UsageCount,
			&tl.Domain,
			&createdAt,
		); err != nil {
			return nil, err
		}
		tl.CreatedAt = time.Unix(createdAt, 0)

		tinylinks = append(tinylinks, tl)
	}

	return tinylinks, nil
}

func (s *TinylinkSQLRepository) Get(ctx context.Context, alias string) (*Tinylink, error) {
	query := `
		SELECT id, user_id, alias, original_url, is_private, usage_count, domain, created_at
		FROM tinylinks
		WHERE alias = ? AND is_private = 0
	`

	tl := &Tinylink{}

	var createdAt int64
	if err := s.db.QueryRowContext(ctx, query, alias).Scan(
		&tl.ID,
		&tl.UserID,
		&tl.Alias,
		&tl.OriginalURL,
		&tl.Private,
		&tl.UsageCount,
		&tl.Domain,
		&createdAt,
	); err != nil {
		return nil, err
	}
	tl.CreatedAt = time.Unix(createdAt, 0)

	return tl, nil
}

func (s *TinylinkSQLRepository) GetByUserID(ctx context.Context, userID, alias string) (*Tinylink, error) {
	query := `
		SELECT id, user_id, alias, original_url, is_private, usage_count, domain, created_at
		FROM tinylinks
		WHERE alias = ? AND user_id = ?
	`

	tl := &Tinylink{}

	var createdAt int64
	if err := s.db.QueryRowContext(ctx, query, alias, userID).Scan(
		&tl.ID,
		&tl.UserID,
		&tl.Alias,
		&tl.OriginalURL,
		&tl.Private,
		&tl.UsageCount,
		&tl.Domain,
		&createdAt,
	); err != nil {
		return nil, err
	}
	tl.CreatedAt = time.Unix(createdAt, 0)

	return tl, nil
}

func (s *TinylinkSQLRepository) IncrementUsageCount(ctx context.Context, rowId uint64) error {
	query := "UPDATE tinylinks SET usage_count = usage_count + 1 WHERE id = ?"

	res, err := s.db.ExecContext(ctx, query, rowId)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return data.ErrNotFound
	}

	return nil
}

func (s *TinylinkSQLRepository) Delete(ctx context.Context, userID, alias string) error {
	query := `DELETE FROM tinylinks WHERE user_id = ? AND alias = ?`
	res, err := s.db.ExecContext(ctx, query, userID, alias)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return data.ErrNotFound
	}

	return nil
}
