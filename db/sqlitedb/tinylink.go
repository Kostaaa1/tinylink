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

func (r *SQLiteTinylinkStore) Save(ctx context.Context, tl *data.Tinylink) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := `INSERT OR IGNORE INTO tinylinks (user_id, alias, original_url) 
		VALUES (?, ?, ?)
		RETURNING id, created_at`

	args := []interface{}{tl.UserID, tl.Alias, tl.URL}

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

	fmt.Println("new ID: ", tl.ID)

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

	return tx.Commit()
}

func (r *SQLiteTinylinkStore) Get(ctx context.Context, userID, id string) (*data.Tinylink, error) {
	query := "SELECT id, user_id, alias, original_url, created_at FROM tinylinks WHERE id = ?"

	var tl data.Tinylink
	var createdAt int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tl.ID,
		&tl.UserID,
		&tl.Alias,
		&tl.URL,
		&createdAt,
	)
	tl.CreatedAt = time.Unix(createdAt, 0)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	return &tl, nil
}

func (r *SQLiteTinylinkStore) List(ctx context.Context, userID string) ([]*data.Tinylink, error) {
	query := `SELECT id, user_id, alias, original_url, created_at FROM tinylinks 
		WHERE user_id = ?`

	var links []*data.Tinylink
	if err := r.db.SelectContext(ctx, &links, query, userID); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *SQLiteTinylinkStore) Delete(ctx context.Context, userID, id string) error {
	query := "DELETE FROM tinylinks WHERE id = ?"

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return fmt.Errorf("no record found to delete")
	}

	return nil
}
