package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/Kostaaa1/tinylink/internal/domain/token"
	"github.com/Kostaaa1/tinylink/internal/infra/db"
	"github.com/Kostaaa1/tinylink/internal/infra/middleware"
	"github.com/Kostaaa1/tinylink/internal/infra/redis"
	"github.com/Kostaaa1/tinylink/pkg/errhandler"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Env         string
	PostgresDSN string
	RedisDSN    string
}

const (
	version = "1.0.0"
)

// @title Tinylink API
// @version 1.0
// @description API documentation for Tinylink
// @host localhost:8000
// @BasePath /
func main() {
	godotenv.Load()

	var conf Config

	flag.StringVar(&conf.Port, "port", "8000", "server port")
	flag.StringVar(&conf.Env, "env", "development", "environment (development|production)")
	flag.StringVar(&conf.PostgresDSN, "postgres-dsn", os.Getenv("POSTGRES_DSN"), "")
	flag.StringVar(&conf.RedisDSN, "redis-dsn", os.Getenv("REDIS_DSN"), "")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbPool, err := db.OpenPostgresPool(ctx, conf.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := db.OpenRedisConn(ctx, conf.RedisDSN)
	if err != nil {
		log.Fatal(err)
	}

	a := &application{
		conf:   conf,
		router: mux.NewRouter(),
		log:    slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	tokenRepo := redis.NewTokenRepository(redisClient)
	tokenService := token.NewService(tokenRepo)

	errHandler := errhandler.New(a.log)
	mw := middleware.New(errHandler, tokenService, a.log)

	a.router.Use(mw.Global)

	a.registerSwagger()
	a.registerUsers(dbPool, tokenRepo, errHandler, mw.RouteProtector)
	a.registerTinylink(dbPool, redisClient, errHandler, mw.RouteProtector)

	if err := a.serve(); err != nil {
		log.Fatal(err)
	}
}
