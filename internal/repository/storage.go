package repository

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/data"
)

type StorageParams struct {
	UserID string
	ID     string
}

type Storage interface {
	GetAll(ctx context.Context, qp StorageParams) ([]data.Tinylink, error)
	Delete(ctx context.Context, qp StorageParams) error
	Create(ctx context.Context, tl data.Tinylink, qp StorageParams) error
	Get(ctx context.Context, qp StorageParams) (data.Tinylink, error)
	Ping(ctx context.Context) error
}
