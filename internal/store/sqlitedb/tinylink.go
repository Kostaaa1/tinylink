package sqlitedb

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/store"
	"github.com/jmoiron/sqlx"
)

type SQLiteTinylinkStore struct {
	db *sqlx.DB
}

func NewSQLiteTinylinkStore(db *sqlx.DB) store.TinylinkStore {
	return &SQLiteTinylinkStore{
		db: db,
	}
}

func (r *SQLiteTinylinkStore) Save(ctx context.Context, tl *data.Tinylink, qp data.QueryParams) error {
	return nil
}
func (r *SQLiteTinylinkStore) Get(ctx context.Context, qp data.QueryParams) (*data.Tinylink, error) {
	return nil, nil
}
func (r *SQLiteTinylinkStore) List(ctx context.Context, qp data.QueryParams) ([]*data.Tinylink, error) {
	return nil, nil
}
func (r *SQLiteTinylinkStore) Delete(ctx context.Context, qp data.QueryParams) error {
	return nil
}
func (r *SQLiteTinylinkStore) Exists(ctx context.Context, id string) (bool, error) {
	return false, nil
}
func (r *SQLiteTinylinkStore) SetAlias(ctx context.Context, alias string) error {
	return nil
}
