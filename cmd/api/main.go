package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Kostaaa1/tinylink/cmd/api/repos/redisrepo"
	"github.com/Kostaaa1/tinylink/internal/repository/storage"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/redis/go-redis/v9"
)

type config struct {
	port        string
	env         string
	storageType string
}

type app struct {
	config
	logger      *slog.Logger
	cookiestore *sessions.CookieStore
	storage     storage.Storage
}

func main() {
	var cfg config

	flag.StringVar(&cfg.port, "port", "3000", "Server address port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.storageType, "storage-type", "redis", "Storage (redis|sqlite|pocketbase)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx := context.Background()

	storage, err := initStorage(ctx, cfg.storageType)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("Successfully started", "storage", cfg.storageType)

	a := app{
		logger:      logger,
		config:      cfg,
		cookiestore: sessions.NewCookieStore([]byte(securecookie.GenerateRandomKey(32))),
		storage:     storage,
	}

	srv := &http.Server{
		Addr:         ":" + cfg.port,
		Handler:      a.Routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Info("Server running on", "port", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func initStorage(ctx context.Context, storageType string) (storage.Storage, error) {
	var storage storage.Storage

	switch storageType {
	case "redis":
		storage = redisrepo.NewRedisRepo(ctx, &redis.Options{
			Addr:     "localhost:6379",
			Password: "lagaosiprovidnokopas",
			DB:       0,
		})
	case "sqlite":
	default:
	}

	if err := storage.Ping(ctx); err != nil {
		return nil, fmt.Errorf("storage ping failed %w", err)
	}

	return storage, nil
}
