package sqlitedb

import (
	"context"
	"time"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/jmoiron/sqlx"
)

type SQLiteTinylinkStore struct {
	db *sqlx.DB
}

type TinylinkStore interface {
	List(ctx context.Context, userID string) ([]*data.Tinylink, error)
	Delete(ctx context.Context, userID, id string) error
	Save(ctx context.Context, tl *data.Tinylink, userID string) error
	Get(ctx context.Context, userID, id string) (*data.Tinylink, error)
	Exists(ctx context.Context, id string) (bool, error)
}

func (r *SQLiteTinylinkStore) Save(ctx context.Context, tl *data.Tinylink, userID string, ttl time.Duration) error {
	return nil
}

func (r *SQLiteTinylinkStore) Get(ctx context.Context, userID, id string) (*data.Tinylink, error) {
	return nil, nil
}

func (r *SQLiteTinylinkStore) List(ctx context.Context, userID string) ([]*data.Tinylink, error) {
	return nil, nil
}

func (r *SQLiteTinylinkStore) Delete(ctx context.Context, userID, id string) error {
	return nil
}

func (r *SQLiteTinylinkStore) Exists(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (r *SQLiteTinylinkStore) SetAlias(ctx context.Context, alias string) error {
	return nil
}
