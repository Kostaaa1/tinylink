package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func OpenRedisConn(ctx context.Context, connString string) (*redis.Client, error) {
	opts, err := redis.ParseURL(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis conn string: %v", err)
	}

	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
