package tinylink

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/mattn/go-sqlite3"
)

type TinylinkSQLRepository struct {
	db db
}

type TinylinkDb struct {
	ID          int            `db:"id"`
	Alias       string         `db:"alias"`
	URL         string         `db:"original_url"`
	UserID      sql.NullInt64  `db:"user_id"`
	Private     bool           `db:"is_private"`
	UsageCount  int            `db:"usage_count"`
	Domain      sql.NullString `db:"domain"`
	Version     uint64         `db:"version"`
	CreatedAt   int64          `db:"created_at"`
	ExpiresAt   int64          `db:"expires_at"`
	LastVisited int64          `db:"last_visited"`
}

func fromDomain(tl *Tinylink) *TinylinkDb {
	var userID sql.NullInt64
	if tl.UserID != nil {
		if parsedID, err := strconv.ParseInt(*tl.UserID, 10, 64); err == nil {
			userID = sql.NullInt64{Int64: parsedID, Valid: true}
		} else {
			userID = sql.NullInt64{Valid: false}
		}
	}
	var domain sql.NullString
	if tl.Domain != nil {
		domain = sql.NullString{String: *tl.Domain, Valid: true}
	}
	return &TinylinkDb{
		Alias:       tl.Alias,
		URL:         tl.URL,
		UserID:      userID,
		Private:     tl.Private,
		UsageCount:  tl.UsageCount,
		Domain:      domain,
		Version:     tl.Version,
		CreatedAt:   tl.CreatedAt,
		ExpiresAt:   tl.ExpiresAt,
		LastVisited: tl.LastVisited,
	}
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
	record := fromDomain(tl)
	query := `UPDATE tinylinks SET alias = ?, domain = ?, is_private = ?, version = version + 1, expires_at = ? 
	WHERE id = ? AND user_id = ?
	RETURNING version`

	args := []interface{}{record.Alias, record.Domain, record.Private, record.ExpiresAt, record.ID, record.UserID}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&tl.Version)

	if err != nil {
		if err == sql.ErrNoRows {
			return data.ErrNotFound
		}
		return err
	}

	return nil
}

func (s *TinylinkSQLRepository) Insert(ctx context.Context, tl *Tinylink) error {
	record := fromDomain(tl)
	query := `INSERT INTO tinylinks (user_id, alias, original_url, domain, is_private) 
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, created_at, version`

	args := []interface{}{record.UserID, record.Alias, record.URL, record.Domain, record.Private}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&tl.ID, &tl.CreatedAt, &tl.Version)
	if err != nil {
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
			&tl.URL,
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

func (s *TinylinkSQLRepository) Exists(ctx context.Context, userID *string, alias string) (bool, error) {
	var query string
	var args []interface{}

	if userID == nil {
		query = `
			SELECT 1
			FROM tinylinks
			WHERE (alias = ? AND is_private = 0)
		`
		args = []interface{}{alias}
	} else {
		query = `
			SELECT 1
			FROM tinylinks
			WHERE (alias = ? AND is_private = 0) OR (alias = ? AND is_private = 1 AND user_id = ?)
		`
		args = []interface{}{alias, userID}
	}

	err := s.db.QueryRowContext(ctx, query, args...).Err()
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *TinylinkSQLRepository) Get(ctx context.Context, alias string) (*Tinylink, error) {
	query := `
		SELECT id, alias, original_url, user_id, is_private, usage_count, domain, version, created_at, expires_at, last_visited
		FROM tinylinks
		WHERE alias = ?
	`

	tl := &Tinylink{}

	if err := s.db.QueryRowContext(ctx, query, alias).Scan(
		&tl.ID,
		&tl.Alias,
		&tl.URL,
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

func (s *TinylinkSQLRepository) Redirect(ctx context.Context, alias string) (uint64, string, error) {
	query := `
		SELECT id, original_url
		FROM tinylinks
		WHERE alias = ? AND is_private = 0
	`
	var rowID uint64
	var URL string
	if err := s.db.QueryRowContext(ctx, query, alias).Scan(&rowID, &URL); err != nil {
		if err == sql.ErrNoRows {
			return 0, "", data.ErrNotFound
		}
		return 0, "", err
	}
	return rowID, URL, nil
}

func (s *TinylinkSQLRepository) RedirectPersonal(ctx context.Context, userID, alias string) (uint64, string, error) {
	query := `
		SELECT id, original_url
		FROM tinylinks
		WHERE alias = ? AND user_id = ?
	`
	var url string
	var rowID uint64
	if err := s.db.QueryRowContext(ctx, query, alias, userID).Scan(&rowID, &url); err != nil {
		if err == sql.ErrNoRows {
			return 0, "", data.ErrNotFound
		}
		return 0, "", err
	}
	return rowID, url, nil
}

func (s *TinylinkSQLRepository) UpdateUsage(ctx context.Context, rowID uint64) error {
	query := `UPDATE tinylinks 
	SET usage_count = usage_count + 1, last_visited = strftime('%s', 'now')
	WHERE id = ?`
	err := s.db.QueryRowContext(ctx, query, rowID).Err()
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
