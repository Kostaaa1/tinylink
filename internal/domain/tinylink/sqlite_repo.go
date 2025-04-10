package tinylink

import (
	"context"
	"database/sql"

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
	query := `UPDATE tinylinks SET alias = ?, domain = ?, is_private = ?, version = version + 1, expires_at = ? 
	WHERE id = ? AND user_id = ?
	RETURNING version`

	args := []interface{}{tl.Alias, tl.Domain, tl.Private, tl.ExpiresAt, tl.ID, tl.UserID}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&tl.Version)

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
		RETURNING id, created_at, version`

	args := []interface{}{tl.UserID, tl.Alias, tl.OriginalURL, tl.Domain, tl.Private}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(&tl.ID, &tl.CreatedAt, &tl.Version)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return ErrAliasExists
		}
		return err
	}

	return nil
}

func (s *TinylinkSQLRepository) List(ctx context.Context, userID string) ([]*Tinylink, error) {
	query := `
		SELECT id, alias, original_url, user_id, is_private, usage_count, domain, version, created_at, expires_at, last_visited 
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
		if err := rows.Scan(
			&tl.ID,
			&tl.Alias,
			&tl.OriginalURL,
			&tl.UserID,
			&tl.Private,
			&tl.UsageCount,
			&tl.Domain,
			&tl.Version,
			&tl.CreatedAt,
			&tl.ExpiresAt,
			&tl.LastVisited,
		); err != nil {
			return nil, err
		}
		tinylinks = append(tinylinks, tl)
	}

	return tinylinks, nil
}

func (s *TinylinkSQLRepository) Get(ctx context.Context, alias string) (*Tinylink, error) {
	query := `
		SELECT id, alias, original_url, user_id, is_private, usage_count, domain, version, created_at, expires_at, last_visited
		FROM tinylinks
		WHERE alias = ? AND is_private = 0
	`

	tl := &Tinylink{}

	if err := s.db.QueryRowContext(ctx, query, alias).Scan(
		&tl.ID,
		&tl.Alias,
		&tl.OriginalURL,
		&tl.UserID,
		&tl.Private,
		&tl.UsageCount,
		&tl.Domain,
		&tl.Version,
		&tl.CreatedAt,
		&tl.ExpiresAt,
		&tl.LastVisited,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, data.ErrNotFound
		}
		return nil, err
	}

	return tl, nil
}

func (s *TinylinkSQLRepository) GetByUserID(ctx context.Context, userID, alias string) (*Tinylink, error) {
	query := `
		SELECT id, alias, original_url, user_id, is_private, usage_count, domain, version, created_at, expires_at, last_visited
		FROM tinylinks
		WHERE alias = ? AND user_id = ?
	`

	tl := &Tinylink{}
	if err := s.db.QueryRowContext(ctx, query, alias, userID).Scan(
		&tl.ID,
		&tl.Alias,
		&tl.OriginalURL,
		&tl.UserID,
		&tl.Private,
		&tl.UsageCount,
		&tl.Domain,
		&tl.Version,
		&tl.CreatedAt,
		&tl.ExpiresAt,
		&tl.LastVisited,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, data.ErrNotFound
		}
		return nil, err
	}

	return tl, nil
}

func (s *TinylinkSQLRepository) UpdateUsage(ctx context.Context, tl *Tinylink) error {
	query := `UPDATE tinylinks 
	SET usage_count = usage_count + 1, last_visited = strftime('%s', 'now')
	WHERE id = ?
	RETURNING usage_count, last_visited`

	err := s.db.QueryRowContext(ctx, query, tl.ID).Scan(
		&tl.UsageCount,
		&tl.LastVisited,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return data.ErrNotFound
		}
		return err
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
