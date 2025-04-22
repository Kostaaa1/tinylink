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
	ID        uint64         `db:"id"`
	Alias     sql.NullString `db:"alias"`
	URL       sql.NullString `db:"original_url"`
	UserID    sql.NullInt64  `db:"user_id"`
	Private   bool           `db:"is_private"`
	Domain    sql.NullString `db:"domain"`
	Version   uint64         `db:"version"`
	CreatedAt int64          `db:"created_at"`
	ExpiresAt int64          `db:"expires_at"`
}

func fromDomain(tl *Tinylink) *TinylinkDb {
	var userID sql.NullInt64
	if tl.UserID != "" {
		if parsedID, err := strconv.ParseInt(tl.UserID, 10, 64); err == nil {
			userID = sql.NullInt64{Int64: parsedID, Valid: true}
		} else {
			userID = sql.NullInt64{Valid: false}
		}
	}
	var domain sql.NullString
	if tl.Domain != nil {
		domain = sql.NullString{String: *tl.Domain, Valid: true}
	}
	var URL sql.NullString
	if tl.URL != "" {
		URL = sql.NullString{String: tl.URL, Valid: true}
	}
	var alias sql.NullString
	if tl.Alias != "" {
		alias = sql.NullString{String: tl.Alias, Valid: true}
	}
	return &TinylinkDb{
		ID:        tl.ID,
		Alias:     alias,
		URL:       URL,
		UserID:    userID,
		Private:   tl.Private,
		Domain:    domain,
		Version:   tl.Version,
		CreatedAt: tl.CreatedAt,
		ExpiresAt: tl.ExpiresAt,
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

	query := `
		UPDATE tinylinks 
		SET 
			alias = CASE WHEN ? IS NOT NULL THEN ? ELSE alias END,
			domain = CASE WHEN ? IS NOT NULL THEN ? ELSE domain END,
			is_private = ?, 
			original_url = CASE WHEN ? IS NOT NULL THEN ? ELSE original_url END,
			version = version + 1,
			expires_at = CASE WHEN ? != 0 THEN ? ELSE expires_at END
		WHERE id = ? AND user_id = ?
		RETURNING version, alias, domain, is_private, original_url, expires_at, created_at`

	args := []interface{}{
		record.Alias, record.Alias,
		record.Domain, record.Domain,
		record.Private,
		record.URL, record.URL,
		record.ExpiresAt, record.ExpiresAt,
		record.ID, record.UserID,
	}

	var (
		alias     sql.NullString
		url       sql.NullString
		domain    sql.NullString
		expiresAt sql.NullInt64
	)

	err := s.db.QueryRowContext(ctx, query, args...).Scan(&tl.Version, &alias, &domain, &tl.Private, &url, &expiresAt, &tl.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return data.ErrNotFound
		}
		return err
	}

	if alias.Valid {
		tl.Alias = alias.String
	}
	if domain.Valid {
		tl.Domain = &domain.String
	}
	if url.Valid {
		tl.URL = url.String
	}
	if expiresAt.Valid {
		tl.ExpiresAt = expiresAt.Int64
	}

	return nil
}

func (s *TinylinkSQLRepository) Create(ctx context.Context, tl *Tinylink) error {
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

func (s *TinylinkSQLRepository) ListUserLinks(ctx context.Context, userID string) ([]*Tinylink, error) {
	query := `
		SELECT id, alias, original_url, user_id, is_private, domain, version, created_at, expires_at 
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
			&tl.Domain,
			&tl.Version,
			&tl.CreatedAt,
			&tl.ExpiresAt,
		); err != nil {
			return nil, err
		}
		tinylinks = append(tinylinks, tl)
	}

	return tinylinks, nil
}

func (s *TinylinkSQLRepository) Exists(ctx context.Context, userID string, alias string) (bool, error) {
	var query string
	var args []interface{}

	if userID == "" {
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
		args = []interface{}{alias, alias, userID}
	}

	var dummy int
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *TinylinkSQLRepository) GetByUserID(ctx context.Context, userID, alias string) (*Tinylink, error) {
	query := `
		SELECT id, alias, original_url, user_id, is_private, domain, version, created_at, expires_at 
		FROM tinylinks
		WHERE alias = ? AND user_id = ?
	`

	tl := &Tinylink{}

	if err := s.db.QueryRowContext(ctx, query, alias, userID).Scan(
		&tl.ID,
		&tl.Alias,
		&tl.URL,
		&tl.UserID,
		&tl.Private,
		&tl.Domain,
		&tl.Version,
		&tl.CreatedAt,
		&tl.ExpiresAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, data.ErrNotFound
		}
		return nil, err
	}

	return tl, nil
}

func (s *TinylinkSQLRepository) Get(ctx context.Context, alias string) (*Tinylink, error) {
	query := `
		SELECT id, alias, original_url, user_id, is_private, domain, version, created_at, expires_at
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
		&tl.Domain,
		&tl.Version,
		&tl.CreatedAt,
		&tl.ExpiresAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, data.ErrNotFound
		}
		return nil, err
	}

	return tl, nil
}

func (s *TinylinkSQLRepository) GetURL(ctx context.Context, alias string) (uint64, string, error) {
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

func (s *TinylinkSQLRepository) GetPersonalURL(ctx context.Context, userID, alias string) (uint64, string, error) {
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

// func (s *TinylinkSQLRepository) UpdateUsage(ctx context.Context, rowID uint64) error {
// 	fmt.Println("UPDATING USAGE FOR ROW: ", rowID)
// 	query := `UPDATE tinylinks
// 	SET usage_count = usage_count + 1, last_visited = strftime('%s', 'now')
// 	WHERE id = ?
// 	RETURNING last_visited`
// 	var dummy int
// 	err := s.db.QueryRowContext(ctx, query, rowID).Scan(&dummy)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return data.ErrNotFound
// 		}
// 		return err
// 	}
// 	return nil
// }

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
