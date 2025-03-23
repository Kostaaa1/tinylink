package main

import (
	"flag"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/Kostaaa1/tinylink/cmd/api/handler"
	"github.com/Kostaaa1/tinylink/db/redisdb"
	"github.com/Kostaaa1/tinylink/db/sqlitedb"
	"github.com/Kostaaa1/tinylink/internal/middleware"
	"github.com/Kostaaa1/tinylink/internal/middleware/auth"
	"github.com/Kostaaa1/tinylink/internal/middleware/ratelimiter"
	"github.com/Kostaaa1/tinylink/internal/services"
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
	flag.StringVar(&cfg.Port, "port", "3000", "Server address port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|staging|production)")
	flag.Float64Var(&cfg.Limiter.RPS, "limiter-rps", 2, "Rate limiter requests-per-second")
	flag.IntVar(&cfg.Limiter.Burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.Limiter.Enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.StringVar(&cfg.SQLitePath, "sqlite-db-path", "tinylink.db", "Path to the SQLite database")
	flag.BoolVar(&cfg.Redis.Enabled, "redis-enabled", false, "Enable redis")
	flag.StringVar(&cfg.Redis.Addr, "redis-addr", "localhost:6379", "Redis server address")
	flag.StringVar(&cfg.Redis.Password, "redis-password", "", "Redis password")
	flag.IntVar(&cfg.Redis.DB, "redis-db", 0, "Redis database number")
	flag.IntVar(&cfg.Redis.PoolSize, "redis-pool-size", 10, "Redis connection pool size")
	flag.Parse()

	logger := setupLogger(os.Stdout, &cfg)

	// make regitry for stores
	sqliteStore := sqlitedb.NewSQLiteStore(cfg.SQLitePath)
	redisStore := redisdb.NewRedisStore(&cfg.Redis)
	userService := services.NewUserService(
		sqliteStore.User,
		redisStore.Token,
	)
	tinylinkService := services.NewTinylinkService(
		sqliteStore.Tinylink,
		redisStore.Tinylink,
		redisStore.Token,
	)
	errHandler := handler.NewErrorHandler(logger)
	tinylinkHandler := handler.NewTinylinkHandler(tinylinkService, errHandler)
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

	authMiddleware := auth.Middleware(redisStore.Token, sqliteStore.User)
	r.Use(middleware.RecoverPanic, limit.Middleware, authMiddleware)

	app.handler.Tinylink.RegisterRoutes(r)
	app.handler.User.RegisterRoutes(r)

	app.router = r

	if err := app.serve(); err != nil {
		logger.Error("app.serve()", "error", err)
	}
}
