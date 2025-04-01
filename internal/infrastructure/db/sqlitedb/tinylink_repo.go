package sqlitedb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
)

type TinylinkRepository struct {
	db *sqlx.DB
}

type flatTL struct {
	ID          uint64         `db:"id"`
	Alias       string         `db:"alias"`
	OriginalURL string         `db:"original_url"`
	UserID      string         `db:"user_id"`
	CreatedAt   string         `db:"created_at"`
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

func (s *TinylinkRepository) Update(ctx context.Context, tl *tinylink.Tinylink) error {
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
			return tinylink.ErrAliasExists
		}
		return err
	}

	return nil
}

func (s *TinylinkRepository) Insert(ctx context.Context, tl *tinylink.Tinylink) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := `INSERT INTO tinylinks (user_id, alias, original_url, domain, is_private) 
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, created_at`

	args := []interface{}{tl.UserID, tl.Alias, tl.OriginalURL, tl.Domain, tl.Private}

	var createdAt int64
	err = tx.QueryRowContext(ctx, query, args...).Scan(&tl.ID, &createdAt)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return tinylink.ErrAliasExists
		}

		tx.Rollback()
		return err
	}
	tl.CreatedAt = time.Unix(createdAt, 0)

	return tx.Commit()
}

func (s *TinylinkRepository) List(ctx context.Context, userID string) ([]*tinylink.Tinylink, error) {
	query := `
		SELECT 
			t.id, t.alias, t.original_url, t.user_id, 
			datetime(t.created_at, 'unixepoch') as created_at,
			q.data, q.width, q.height, q.size, q.mime_type as mimetype
		FROM tinylinks t
		LEFT JOIN qrcodes q ON t.id = q.tinylink_id
		WHERE t.user_id = ?
	`

	tinylinks := []*tinylink.Tinylink{}

	var links []flatTL
	if err := s.db.SelectContext(ctx, &links, query, userID); err != nil {
		return tinylinks, err
	}

	for _, r := range links {
		createdAt, _ := time.Parse("2006-01-02 15:04:05", r.CreatedAt)
		tl := &tinylink.Tinylink{
			ID:          r.ID,
			Alias:       r.Alias,
			UserID:      r.UserID,
			OriginalURL: r.OriginalURL,
			CreatedAt:   createdAt,
		}
		tinylinks = append(tinylinks, tl)
	}

	return tinylinks, nil
}

func (s *TinylinkRepository) GetPublic(ctx context.Context, alias string) (*tinylink.Tinylink, error) {
	query := `
		SELECT 
			t.id, t.alias, t.original_url, t.user_id, 
			datetime(t.created_at, 'unixepoch') as created_at,
			q.data, q.width, q.height, q.size, q.mime_type as mimetype
		FROM tinylinks t
		LEFT JOIN qrcodes q ON t.id = q.tinylink_id
		WHERE t.alias = ? AND t.is_private = 1
	`

	var r flatTL
	if err := s.db.GetContext(ctx, &r, query, alias); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	tl := &tinylink.Tinylink{
		ID:          r.ID,
		Alias:       r.Alias,
		OriginalURL: r.OriginalURL,
		UserID:      r.UserID,
		CreatedAt:   createdAt,
	}

	return tl, nil
}

func (s *TinylinkRepository) Get(ctx context.Context, userID, alias string) (*tinylink.Tinylink, error) {
	query := `
		SELECT 
			t.id, t.alias, t.original_url, t.user_id, 
			datetime(t.created_at, 'unixepoch') as created_at,
			q.data, q.width, q.height, q.size, q.mime_type as mimetype
		FROM tinylinks t
		LEFT JOIN qrcodes q ON t.id = q.tinylink_id
		WHERE t.user_id = ? AND t.alias = ?
	`

	var r flatTL
	if err := s.db.GetContext(ctx, &r, query, userID, alias); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	tl := &tinylink.Tinylink{
		ID:          r.ID,
		Alias:       r.Alias,
		OriginalURL: r.OriginalURL,
		UserID:      r.UserID,
		CreatedAt:   createdAt,
	}

	return tl, nil
}

func (s *TinylinkRepository) IncrementUsageCount(ctx context.Context, alias string) error {
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

func (s *TinylinkRepository) Delete(ctx context.Context, userID, alias string) error {
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
