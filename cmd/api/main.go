package main

import (
	"flag"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/db/redisdb"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/db/sqlitedb"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/handler"
	"github.com/Kostaaa1/tinylink/internal/middleware"
	loggermw "github.com/Kostaaa1/tinylink/internal/middleware/logger"
	"github.com/Kostaaa1/tinylink/internal/middleware/ratelimiter"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type application struct {
	cfg     *config.Config
	handler *handler.Handler
	router  *mux.Router
	logger  *slog.Logger
}

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func setupLogger(w io.Writer, cfg *config.Config) *slog.Logger {
	var logHandler slog.Handler
	if cfg.Env == "development" {
		logHandler = slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		logHandler = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelError})
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)
	return logger
}

func main() {
	var cfg config.Config

	flag.StringVar(&cfg.Port, "port", "3000", "server address port")
	flag.StringVar(&cfg.Env, "env", "development", "environment (development|staging|production)")
	flag.Float64Var(&cfg.Limiter.RPS, "limiter-rps", 2, "rate limiter requests-per-second")
	flag.IntVar(&cfg.Limiter.Burst, "limiter-burst", 4, "rate limiter maximum burst")
	flag.BoolVar(&cfg.Limiter.Enabled, "limiter-enabled", true, "enable rate limiter")
	flag.StringVar(&cfg.SQLitePath, "sqlite-db-path", "tinylink.db", "path to the sqlite database")
	flag.BoolVar(&cfg.Redis.Enabled, "redis-enabled", false, "enable redis")
	flag.StringVar(&cfg.Redis.Addr, "redis-addr", "localhost:6379", "redis server address")
	flag.StringVar(&cfg.Redis.Password, "redis-password", "", "redis password")
	flag.IntVar(&cfg.Redis.DB, "redis-db", 0, "redis database number")
	flag.IntVar(&cfg.Redis.PoolSize, "redis-pool-size", 10, "redis connection pool size")

	flag.Parse()

	logger := setupLogger(os.Stdout, &cfg)

	redisRepo := redisdb.NewRepositories(&cfg.Redis)
	sqliteRepo := sqlitedb.NewRepositories(cfg.SQLitePath)

	tlService := tinylink.NewService(sqliteRepo.Tinylink, redisRepo.Tinylink, redisRepo.Token)
	userService := user.NewService(sqliteRepo.User, redisRepo.Token)

	errHandler := handler.NewErrorHandler(logger)
	tinylinkHandler := handler.NewTinylinkHandler(tlService, errHandler)
	userHandler := handler.NewUserHandler(userService, errHandler)

	app := application{
		cfg:    &cfg,
		logger: logger,
		handler: &handler.Handler{
			ErrorHandler: errHandler,
			Tinylink:     tinylinkHandler,
			User:         userHandler,
		},
	}

	r := mux.NewRouter()
	r.MethodNotAllowedHandler = http.HandlerFunc(app.handler.MethodNotAllowedResponse)
	r.NotFoundHandler = http.HandlerFunc(app.handler.NotFoundResponse)

	limit := ratelimiter.New(app.cfg.Limiter)
	// authMW := auth.Middleware(redisStore.Token, sqliteStore.User)
	r.Use(middleware.RecoverPanic, limit.Middleware, loggermw.Middleware)

	// app.handler.Tinylink.RegisterRoutes(r, nil)
	app.handler.User.RegisterRoutes(r)
	app.router = r

	if err := app.serve(); err != nil {
		logger.Error("app.serve()", "error", err)
	}
}
