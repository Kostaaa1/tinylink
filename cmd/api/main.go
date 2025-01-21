package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type config struct {
	port        string
	env         string
	storageType string
	redis       struct {
		addr string
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
	storage     TinylinkRepository
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
	flag.StringVar(&cfg.redis.addr, "redis-addr", "redis://:lagaosiprovidnokopas@localhost:6379/0", "Redis addres [redis://:password@localhost:6379/0]")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	a := app{
		logger:      logger,
		cfg:         cfg,
		cookiestore: newCookieStore(),
		storage:     newStorage(cfg),
	}

	r := mux.NewRouter()

	r.MethodNotAllowedHandler = http.HandlerFunc(a.methodNotAllowedResponse)
	r.NotFoundHandler = http.HandlerFunc(a.notFoundResponse)
	r.Use(a.recoverPanic, a.rateLimit, a.persistSessionMW)

	srv := &http.Server{
		Addr:           ":" + cfg.port,
		Handler:        r,
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

func newStorage(cfg config) TinylinkRepository {
	ctx := context.Background()

	var tinylinkRepo TinylinkRepository

	switch cfg.storageType {
	case "redis":
		client := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "lagaosiprovidnokopas",
			DB:       0,
		})
		if err := client.Ping(ctx).Err(); err != nil {
			log.Fatalf("Storage ping failed for %s: %v", cfg.storageType, err)
		}
		tinylinkRepo = NewRedisTinylinkRepository(client)
	case "sqlite":
		// Add sqlite initialization here
	default:
		log.Fatalf("Unsupported storage type: %s", cfg.storageType)
	}

	return tinylinkRepo
}
