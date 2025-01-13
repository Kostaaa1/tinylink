package storage

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/models"
)

type QueryParams struct {
	UserID string
	ID     string
}

type Storage interface {
	GetAll(ctx context.Context, qp QueryParams) ([]*models.Tinylink, error)
	Delete(ctx context.Context, qp QueryParams) error
	Create(ctx context.Context, tl *models.Tinylink, qp QueryParams) (*models.Tinylink, error)
	Get(ctx context.Context, qp QueryParams) (*models.Tinylink, error)
	Ping(ctx context.Context) error
}
