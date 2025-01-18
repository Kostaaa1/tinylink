package storage

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/models"
)

type QueryParams struct {
	ClientID string
	Alias    string
}

type Storage interface {
	GetAll(ctx context.Context, qp QueryParams) ([]*models.Tinylink, error)
	Delete(ctx context.Context, qp QueryParams) error
	Create(ctx context.Context, tl *models.Tinylink, qp QueryParams) error
	Get(ctx context.Context, qp QueryParams) (*models.Tinylink, error)
	Ping(ctx context.Context) error
	ValidateAlias(ctx context.Context, alias string) error
	ValidateOriginalURL(ctx context.Context, clientID, URL string) error
}
