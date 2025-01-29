package repositories

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/domain/entities"
)

type TinylinkRepository interface {
	List(ctx context.Context, qp entities.QueryParams) ([]*entities.Tinylink, error)
	Delete(ctx context.Context, qp entities.QueryParams) error
	Save(ctx context.Context, tl *entities.Tinylink, qp entities.QueryParams) error
	Get(ctx context.Context, qp entities.QueryParams) (*entities.Tinylink, error)
	Exists(ctx context.Context, id string) (bool, error)
	SetAlias(ctx context.Context, alias string) error
	// SetOriginalURL(ctx context.Context, clientID, URL string) error
}
