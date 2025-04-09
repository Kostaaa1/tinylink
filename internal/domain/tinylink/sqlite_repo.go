package tinylink

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/mattn/go-sqlite3"
)

type TinylinkSQLRepository struct {
	db db
}

type flatTL struct {
	ID          uint64         `db:"id"`
	Alias       string         `db:"alias"`
	OriginalURL string         `db:"original_url"`
	UserID      string         `db:"user_id"`
	CreatedAt   int64          `db:"created_at"`
	Private     bool           `db:"private"`
	Domain      string         `db:"domain"`
	UsageCount  int            `db:"usage_count"`
	QRData      []byte         `db:"data"`
	QRWidth     sql.NullString `db:"width"`
	QRHeight    sql.NullString `db:"height"`
	QRSize      sql.NullString `db:"size"`
	QRMimeType  sql.NullString `db:"mimetype"`
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
			return data.ErrRecordNotFound
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
		SELECT 
			t.id, t.alias, t.original_url, t.user_id, t.created_at, 
			q.data, q.width, q.height, q.size, q.mime_type as mimetype
		FROM tinylinks t
		LEFT JOIN qrcodes q ON t.id = q.tinylink_id
		WHERE t.user_id = ?
	`

	tinylinks := []*Tinylink{}

	var links []flatTL
	if err := s.db.SelectContext(ctx, &links, query, userID); err != nil {
		return tinylinks, err
	}

	for _, r := range links {
		tl := &Tinylink{
			ID:          r.ID,
			Alias:       r.Alias,
			UserID:      r.UserID,
			OriginalURL: r.OriginalURL,
			CreatedAt:   time.Unix(r.CreatedAt, 0),
		}
		tinylinks = append(tinylinks, tl)
	}

	return tinylinks, nil
}

func (s *TinylinkSQLRepository) Get(ctx context.Context, alias string) (*Tinylink, error) {
	query := `
		SELECT
			t.id, t.alias, t.original_url, t.user_id, t.created_at, 
			q.data, q.width, q.height, q.size, q.mime_type as mimetype
		FROM tinylinks t
		LEFT JOIN qrcodes q ON t.id = q.tinylink_id
		WHERE t.alias = ? AND t.is_private = 1
	`

	var flat flatTL
	if err := s.db.GetContext(ctx, &flat, query, alias); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	tl := &Tinylink{
		ID:          flat.ID,
		Alias:       flat.Alias,
		OriginalURL: flat.OriginalURL,
		UserID:      flat.UserID,
		CreatedAt:   time.Unix(flat.CreatedAt, 0),
	}

	return tl, nil
}

func (s *TinylinkSQLRepository) GetTest(ctx context.Context, alias string) (*Tinylink, error) {
	query := `
		SELECT
			t.id, t.alias, t.original_url, t.user_id, t.created_at, 
			q.data, q.width, q.height, q.size, q.mime_type as mimetype
		FROM tinylinks t
		LEFT JOIN qrcodes q ON t.id = q.tinylink_id
		WHERE t.alias = ? AND t.is_private = 1
	`

	var flat flatTL
	if err := s.db.GetContext(ctx, &flat, query, alias); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	tl := &Tinylink{
		ID:          flat.ID,
		Alias:       flat.Alias,
		OriginalURL: flat.OriginalURL,
		UserID:      flat.UserID,
		CreatedAt:   time.Unix(flat.CreatedAt, 0),
	}

	return tl, nil
}

func (s *TinylinkSQLRepository) GetByUserID(ctx context.Context, userID, alias string) (*Tinylink, error) {
	query := `
		SELECT 
			t.id, t.alias, t.original_url, t.user_id, t.created_at,
			q.data, q.width, q.height, q.size, q.mime_type as mimetype
		FROM tinylinks t
		LEFT JOIN qrcodes q ON t.id = q.tinylink_id
		WHERE t.alias = ?
	`

	var flat flatTL
	if err := s.db.GetContext(ctx, &flat, query, alias); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	tl := &Tinylink{
		ID:          flat.ID,
		Alias:       flat.Alias,
		OriginalURL: flat.OriginalURL,
		UserID:      flat.UserID,
		CreatedAt:   time.Unix(flat.CreatedAt, 0),
	}

	return tl, nil
}

func (s *TinylinkSQLRepository) IncrementUsageCount(ctx context.Context, alias string) error {
	query := "UPDATE tinylinks SET usage_count = usage_count + 1 WHERE id = ?"

	res, err := s.db.ExecContext(ctx, query, alias)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return data.ErrRecordNotFound
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
		return data.ErrRecordNotFound
	}

	return nil
}
