package sqlitedb

import (
	"context"
	"database/sql"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/store"
)

type SQLiteTinylinkRepository struct {
	db *sql.DB
}

func NewSQLiteTinylinkRepository(db *sql.DB) store.TinylinkRepository {
	return &SQLiteTinylinkRepository{
		db: db,
	}
}

func (r *SQLiteTinylinkRepository) Save(ctx context.Context, tl *data.Tinylink, qp data.QueryParams) error {
	return nil
}
func (r *SQLiteTinylinkRepository) Get(ctx context.Context, qp data.QueryParams) (*data.Tinylink, error) {
	return nil, nil
}
func (r *SQLiteTinylinkRepository) List(ctx context.Context, qp data.QueryParams) ([]*data.Tinylink, error) {
	return nil, nil
}
func (r *SQLiteTinylinkRepository) Delete(ctx context.Context, qp data.QueryParams) error {
	return nil
}
func (r *SQLiteTinylinkRepository) Exists(ctx context.Context, id string) (bool, error) {
	return false, nil
}
func (r *SQLiteTinylinkRepository) SetAlias(ctx context.Context, alias string) error {
	return nil
}
