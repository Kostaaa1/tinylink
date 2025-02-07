package infrastructure

import (
	"context"
	"fmt"

	"github.com/Kostaaa1/tinylink/internal/domain/repositories"
	redisdb "github.com/Kostaaa1/tinylink/internal/infrastructure/redis"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/redis/go-redis/v9"
)

type Repositories struct {
	Tinylink repositories.TinylinkRepository
	// User repositories.UserRepository
}

func NewRepositories(cfg *config.Config) (*Repositories, error) {
	pingctx := context.Background()
	repos := &Repositories{}

	switch cfg.StorageType {
	case "redis":
		client := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "lagaosiprovidnokopas",
			DB:       0,
		})
		if err := client.Ping(pingctx).Err(); err != nil {
			return nil, fmt.Errorf("storage ping failed for %s: %v", cfg.StorageType, err)
		}
		repos.Tinylink = redisdb.NewTinylinkRepository(client)
	case "sqlite":
	default:
		return nil, fmt.Errorf("not supported storage type: %s", cfg.StorageType)
	}

	return repos, nil
}
