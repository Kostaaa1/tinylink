package sqlitedb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/jmoiron/sqlx"
)

type SQLiteTinylinkStore struct {
	db *sqlx.DB
}

type flatTL struct {
	ID         uint64         `db:"id"`
	Alias      string         `db:"alias"`
	URL        string         `db:"original_url"`
	UserID     string         `db:"user_id"`
	CreatedAt  string         `db:"created_at"`
	Public     bool           `db:"public"`
	Domain     string         `db:"domain"`
	UsageCount int            `db:"usage_count"`
	QRData     []byte         `db:"data"`
	QRWidth    sql.NullString `db:"width"`
	QRHeight   sql.NullString `db:"height"`
	QRSize     sql.NullString `db:"size"`
	QRMimeType sql.NullString `db:"mimetype"`
}

func (s *SQLiteTinylinkStore) Update(ctx context.Context, tl *data.Tinylink) error {
	query := `UPDATE tinylinks SET alias = ?, domain = ?, is_public = ? WHERE id = ? 
	RETURNING user_id, original_url, usage_count, domain, created_at`

	args := []interface{}{tl.Alias, tl.Domain, tl.Public, tl.ID}
	var createdAt int64
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&tl.UserID,
		&tl.URL,
		&tl.UsageCount,
		&tl.Domain,
		&createdAt,
	)
	tl.CreatedAt = time.Unix(createdAt, 0)

	if err != nil {
		if err == sql.ErrNoRows {
			return data.ErrRecordNotFound
		}
		return err
	}

	return nil
}

func (s *SQLiteTinylinkStore) Insert(ctx context.Context, tl *data.Tinylink) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := `INSERT INTO tinylinks (user_id, alias, original_url, domain, is_public) 
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, created_at`

	args := []interface{}{tl.UserID, tl.Alias, tl.URL, tl.Domain, tl.Public}

	var createdAt int64
	err = tx.QueryRowContext(ctx, query, args...).Scan(&tl.ID, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return data.ErrAliasExists
		}
		tx.Rollback()
		return err
	}
	tl.CreatedAt = time.Unix(createdAt, 0)

	if tl.QR != nil {
		queryQR := `INSERT INTO qrcodes 
			(
				tinylink_id,
				data,
				width,
				height,
				size,
				mime_type
			) 
			VALUES (?, ?, ?, ?, ?, ?)
		`
		argsQR := []interface{}{tl.ID, tl.QR.Data, tl.QR.Width, tl.QR.Height, tl.QR.Size, tl.QR.MimeType}
		_, err = tx.ExecContext(ctx, queryQR, argsQR...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLiteTinylinkStore) List(ctx context.Context, userID string) ([]*data.Tinylink, error) {
	query := `
		SELECT 
			t.id, t.alias, t.original_url, t.user_id, 
			datetime(t.created_at, 'unixepoch') as created_at,
			q.data, q.width, q.height, q.size, q.mime_type as mimetype
		FROM tinylinks t
		LEFT JOIN qrcodes q ON t.id = q.tinylink_id
		WHERE t.user_id = ?
	`

	tinylinks := []*data.Tinylink{}

	var links []flatTL
	if err := s.db.SelectContext(ctx, &links, query, userID); err != nil {
		return tinylinks, err
	}

	for _, r := range links {
		createdAt, _ := time.Parse("2006-01-02 15:04:05", r.CreatedAt)
		tl := &data.Tinylink{
			ID:        r.ID,
			Alias:     r.Alias,
			UserID:    r.UserID,
			URL:       r.URL,
			CreatedAt: createdAt,
		}

		if r.QRData != nil && r.QRWidth.Valid && r.QRHeight.Valid && r.QRSize.Valid && r.QRMimeType.Valid {
			tl.QR = &data.QR{
				Data:     r.QRData,
				Width:    r.QRWidth.String,
				Height:   r.QRHeight.String,
				Size:     r.QRSize.String,
				MimeType: r.QRMimeType.String,
			}
		}

		tinylinks = append(tinylinks, tl)
	}

	return tinylinks, nil
}

func (s *SQLiteTinylinkStore) GetPublic(ctx context.Context, alias string) (*data.Tinylink, error) {
	query := `
		SELECT 
			t.id, t.alias, t.original_url, t.user_id, 
			datetime(t.created_at, 'unixepoch') as created_at,
			q.data, q.width, q.height, q.size, q.mime_type as mimetype
		FROM tinylinks t
		LEFT JOIN qrcodes q ON t.id = q.tinylink_id
		WHERE t.alias = ? AND t.is_public = 1
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

	tl := &data.Tinylink{
		ID:        r.ID,
		Alias:     r.Alias,
		URL:       r.URL,
		UserID:    r.UserID,
		CreatedAt: createdAt,
	}

	if r.QRData != nil && r.QRWidth.Valid && r.QRHeight.Valid && r.QRSize.Valid && r.QRMimeType.Valid {
		tl.QR = &data.QR{
			Data:     r.QRData,
			Width:    r.QRWidth.String,
			Height:   r.QRHeight.String,
			Size:     r.QRSize.String,
			MimeType: r.QRMimeType.String,
		}
	}

	return tl, nil
}

func (s *SQLiteTinylinkStore) Get(ctx context.Context, userID, alias string) (*data.Tinylink, error) {
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

	tl := &data.Tinylink{
		ID:        r.ID,
		Alias:     r.Alias,
		URL:       r.URL,
		UserID:    r.UserID,
		CreatedAt: createdAt,
	}

	if r.QRData != nil && r.QRWidth.Valid && r.QRHeight.Valid && r.QRSize.Valid && r.QRMimeType.Valid {
		tl.QR = &data.QR{
			Data:     r.QRData,
			Width:    r.QRWidth.String,
			Height:   r.QRHeight.String,
			Size:     r.QRSize.String,
			MimeType: r.QRMimeType.String,
		}
	}

	return tl, nil
}

func (s *SQLiteTinylinkStore) IncrementUsageCount(ctx context.Context, alias string) error {
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

func (s *SQLiteTinylinkStore) Delete(ctx context.Context, userID, alias string) error {
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
