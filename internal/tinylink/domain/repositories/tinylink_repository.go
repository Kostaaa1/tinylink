package repositories

import "github.com/Kostaaa1/tinylink/internal/tinylink/domain/entities"

type Tinylink interface {
	List(qp entities.QueryParams) ([]*entities.Tinylink, error)
	Delete(qp entities.QueryParams) error
	Save(tl *entities.Tinylink, qp entities.QueryParams) error
	Get(qp entities.QueryParams) (*entities.Tinylink, error)
	////////////////
	CheckAlias(alias string) error
	CheckOriginalURL(clientID, URL string) error
	// Ping(ctx context.Context) error
}
