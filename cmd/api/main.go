package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

type config struct {
	port        string
	env         string
	storageType string
	redis       struct {
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type app struct {
	cfg         config
	logger      *slog.Logger
	cookiestore *sessions.CookieStore
	// service     *service.StorageService
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
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter requests-per-second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// storage := NewStorageRepository(cfg)

	a := app{
		logger:      logger,
		cfg:         cfg,
		cookiestore: newCookieStore(),
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
		fmt.Println("No key found in env. Generating...")
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

// needs to be moved. Config needs to be seperated. Add support for config files.
// func NewStorageRepository(cfg config) domain.StorageRepository {
// 	var repo domain.StorageRepository
// 	switch cfg.storageType {
// 	case "redis":
// 		client := redis.NewClient(&redis.Options{
// 			Addr:     "localhost:6379",
// 			Password: "lagaosiprovidnokopas",
// 			DB:       0,
// 		})
// 		repo = repository.NewRedisRepo(client)
// 	case "sqlite":
// 	default:
// 	}
// 	return repo
// }

// func newStorage(storageType string) domain.Storage {
// 	ctx := context.Background()
// 	var storage domain.Storage
// 	switch storageType {
// 	case "redis":
// 		storage = redis.NewClient(&redis.Options{
// 			Addr:     "localhost:6379",
// 			Password: "lagaosiprovidnokopas",
// 			DB:       0,
// 		})
// 	case "sqlite":
// 	default:
// 	}
// 	if err := storage.Ping(ctx); err != nil {
// 		log.Fatalf("Storage ping failed for %s: %v", storageType, err)
// 	}
// 	return storage
// }
