package interfaces

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/domain/entities"
)

type TinylinkService interface {
	List(ctx context.Context, sessionID string) ([]*entities.Tinylink, error)
	Delete(ctx context.Context, sessionID, alias string) error
	Save(ctx context.Context, sessionID, url, alias string) (*entities.Tinylink, error)
	Get(ctx context.Context, sessionID, alias string) (*entities.Tinylink, error)
}
