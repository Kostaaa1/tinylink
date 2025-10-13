package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/Kostaaa1/tinylink/core/transactor"
	"github.com/Kostaaa1/tinylink/core/transactor/adapters"
	"github.com/Kostaaa1/tinylink/internal/constants"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TinylinkRepository struct {
	db adapters.PgxQuerier
}

func NewTinylinkRepository(db adapters.PgxQuerier) tinylink.DbRepository {
	return &TinylinkRepository{db: db}
}

func (p *TinylinkRepository) WithRepositoryTx(tx transactor.Tx) tinylink.DbRepository {
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		panic("tx does not match type of pgx.Tx")
	}
	return &TinylinkRepository{db: pgxTx}
}

func (r *TinylinkRepository) Insert(ctx context.Context, tl *tinylink.Tinylink) error {
	query := `INSERT INTO tinylinks 
			(alias, url, private, user_id, guest_id, domain, expiration)
			VALUES
			($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, version, updated_at, domain, expiration, guest_id
		`
	args := []interface{}{tl.Alias, tl.URL, tl.Private, tl.UserID, tl.GuestUUID, tl.Domain, tl.Expiration}

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&tl.ID,
		&tl.CreatedAt,
		&tl.Version,
		&tl.UpdatedAt,
		&tl.Domain,
		&tl.Expiration,
		&tl.GuestUUID,
	)
	if err != nil {
		// handle chk_user_or_guest_id_not_null constraint
		if isAliasUniqueErr(err) {
			return tinylink.ErrAliasExists
		}
		return err
	}

	return nil
}

func isAliasUniqueErr(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "uniq_public_alias") || strings.Contains(pgErr.ConstraintName, "uniq_alias_per_user") {
			return true
		}
	}
	return false
}

func (r *TinylinkRepository) Update(ctx context.Context, tl *tinylink.Tinylink) error {
	query := `UPDATE tinylinks 
		SET alias = $1, url = $2, private = $3, expiration = $4, domain = $5, 
		version = version + 1, updated_at = NOW()
		WHERE user_id = $6 AND id = $7
		RETURNING alias, url, domain, private, user_id, guest_id, version, created_at, updated_at, expiration`

	args := []interface{}{
		tl.Alias,
		tl.URL,
		tl.Private,
		tl.Expiration,
		tl.Domain,
		tl.UserID,
		tl.ID,
	}

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&tl.Alias,
		&tl.URL,
		&tl.Domain,
		&tl.Private,
		&tl.UserID,
		&tl.GuestUUID,
		&tl.Version,
		&tl.CreatedAt,
		&tl.UpdatedAt,
		&tl.Expiration,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return constants.ErrNotFound
		}
		// TODO: handle chk_user_or_guest_id_not_null constraint
		if isAliasUniqueErr(err) {
			return tinylink.ErrAliasExists
		}
		return err
	}

	return nil
}

func (r *TinylinkRepository) ListByGuestUUID(ctx context.Context, uuid string) ([]*tinylink.Tinylink, error) {
	query := `SELECT id, alias, url, user_id, guest_id, version, domain, private, created_at, updated_at, expiration FROM tinylinks WHERE guest_id = $1`

	rows, err := r.db.Query(ctx, query, uuid)
	if err != nil {
		return nil, err
	}

	tinylinks := make([]*tinylink.Tinylink, 0)

	for rows.Next() {
		tl := &tinylink.Tinylink{}
		err := rows.Scan(
			&tl.ID,
			&tl.Alias,
			&tl.URL,
			&tl.UserID,
			&tl.GuestUUID,
			&tl.Version,
			&tl.Domain,
			&tl.Private,
			&tl.CreatedAt,
			&tl.UpdatedAt,
			&tl.Expiration,
		)
		if err != nil {
			return nil, err
		}
		tinylinks = append(tinylinks, tl)
	}

	return tinylinks, nil
}

func (r *TinylinkRepository) ListByUserID(ctx context.Context, userID uint64) ([]*tinylink.Tinylink, error) {
	query := `SELECT id, alias, url, user_id, guest_id, version, domain, private, created_at, updated_at, expiration FROM tinylinks WHERE user_id = $1`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	tinylinks := make([]*tinylink.Tinylink, 0)

	for rows.Next() {
		tl := &tinylink.Tinylink{}
		err := rows.Scan(
			&tl.ID,
			&tl.Alias,
			&tl.URL,
			&tl.UserID,
			&tl.GuestUUID,
			&tl.Version,
			&tl.Domain,
			&tl.Private,
			&tl.CreatedAt,
			&tl.UpdatedAt,
			&tl.Expiration,
		)
		if err != nil {
			return nil, err
		}
		tinylinks = append(tinylinks, tl)
	}

	return tinylinks, nil
}

func (r *TinylinkRepository) Redirect(ctx context.Context, alias string, userID *uint64) (*tinylink.RedirectValue, error) {
	query := `SELECT id, url FROM tinylinks WHERE alias = $1 AND (user_id = $2 OR $2 IS NULL)`
	var redirect tinylink.RedirectValue
	err := r.db.QueryRow(ctx, query, alias, userID).Scan(&redirect.RowID, &redirect.URL)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, constants.ErrNotFound
		}
		return nil, err
	}

	return &redirect, nil
}

func (r *TinylinkRepository) Get(ctx context.Context, rowID uint64) (*tinylink.Tinylink, error) {
	query := `SELECT id, alias, url, user_id, guest_id, version, domain, private, created_at, updated_at, expiration FROM tinylinks WHERE id = $1`

	var tl tinylink.Tinylink
	err := r.db.QueryRow(ctx, query, rowID).
		Scan(
			&tl.ID,
			&tl.Alias,
			&tl.URL,
			&tl.UserID,
			&tl.GuestUUID,
			&tl.Version,
			&tl.Domain,
			&tl.Private,
			&tl.CreatedAt,
			&tl.UpdatedAt,
			&tl.Expiration,
		)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, constants.ErrNotFound
		}
		return nil, err
	}

	return &tl, nil
}

func (r *TinylinkRepository) Delete(ctx context.Context, userID uint64, alias string) error {
	res, err := r.db.Exec(ctx, `DELETE FROM TINYLINKS WHERE alias = $1 and user_id = $2`, alias, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return constants.ErrNotFound
		}
		return err
	}

	if res.RowsAffected() > 0 {
		return nil
	}

	return errors.New("no rows affected upon delete")
}
