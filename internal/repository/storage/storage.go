package storage

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/models"
)

type QueryParams struct {
	ClientID    string
	Alias       string
	CheckUnique bool
}

type Storage interface {
	GetAll(ctx context.Context, qp QueryParams) ([]*models.Tinylink, error)
	Delete(ctx context.Context, qp QueryParams) error
	Create(ctx context.Context, tl *models.Tinylink, qp QueryParams) (*models.Tinylink, error)
	Get(ctx context.Context, qp QueryParams) (*models.Tinylink, error)
	Ping(ctx context.Context) error
	ValidAlias(ctx context.Context, alias string) (bool, error)
	ValidOriginalURL(ctx context.Context, URL string, qp QueryParams) (bool, error)
}
