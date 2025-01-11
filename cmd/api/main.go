package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/redis/go-redis/v9"
)

type config struct {
	port  string
	env   string
	store string
}

type app struct {
	config
	logger *slog.Logger
	rdb    *redis.Client
	store  *sessions.CookieStore
}

func main() {
	var cfg config
	flag.StringVar(&cfg.port, "port", "3000", "Server address port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	rdb, err := initRedis()
	if err != nil {
		logger.Info("Failed to start redis")
		log.Fatal(err)
	}

	logger.Info("Redis successfully started")

	a := app{
		logger: logger,
		config: cfg,
		rdb:    rdb,
		store:  sessions.NewCookieStore([]byte(securecookie.GenerateRandomKey(32))),
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

func initRedis() (*redis.Client, error) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "lagaosiprovidnokopas",
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}
