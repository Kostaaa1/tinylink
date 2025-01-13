package main

import (
	"context"
	"encoding/hex"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Kostaaa1/tinylink/internal/repository/storage"
	"github.com/Kostaaa1/tinylink/pkg/repos/redisrepo"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
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
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	var cfg config

	flag.StringVar(&cfg.port, "port", "3000", "Server address port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.storageType, "storage-type", "redis", "Storage (redis|sqlite|pocketbase)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	a := app{
		logger:      logger,
		config:      cfg,
		cookiestore: newCookieStore(),
		storage:     newStorage(cfg.storageType),
	}

	srv := &http.Server{
		Addr:           ":" + cfg.port,
		Handler:        a.Routes(),
		IdleTimeout:    1 * time.Minute,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 2 * 1024 * 1024,
	}

	logger.Info("Server running on", "port", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func newCookieStore() *sessions.CookieStore {
	authKeyHex := os.Getenv("TINYLINK_AUTH_KEY")
	if authKeyHex == "" {
		authKeyHex = generateRandHex(32)
	}
	authKey, _ := hex.DecodeString(authKeyHex)

	encryptionKeyHex := os.Getenv("TINYLINK_ENCRYPTION_KEY")
	if encryptionKeyHex == "" {
		encryptionKeyHex = generateRandHex(16)
	}
	encryptionKey, _ := hex.DecodeString(encryptionKeyHex)

	return sessions.NewCookieStore(authKey, encryptionKey)
}

func newStorage(storageType string) storage.Storage {
	ctx := context.Background()

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
		log.Fatalf("Storage ping failed for %s: %v", storageType, err)
	}

	return storage
}
