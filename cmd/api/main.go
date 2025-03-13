package main

import (
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/Kostaaa1/tinylink/cmd/api/handler"
	"github.com/Kostaaa1/tinylink/internal/services"
	"github.com/Kostaaa1/tinylink/internal/store"
	"github.com/Kostaaa1/tinylink/internal/store/redisdb"
	"github.com/Kostaaa1/tinylink/internal/store/sqlitedb"
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
	flag.StringVar(&cfg.SQLitePath, "sqlite-db-path", "./db/tinylink.db", "Path to the SQLite database")
	flag.BoolVar(&cfg.Redis.Enabled, "redis-enabled", false, "Enable redis")
	flag.StringVar(&cfg.Redis.Addr, "redis-addr", "localhost:6379", "Redis server address")
	flag.StringVar(&cfg.Redis.Password, "redis-password", "", "Redis password")
	flag.IntVar(&cfg.Redis.DB, "redis-db", 0, "Redis database number")
	flag.IntVar(&cfg.Redis.PoolSize, "redis-pool-size", 10, "Redis connection pool size")
	flag.Parse()

	logger := setupLogger(os.Stdout, &cfg)

	storeRegistry := store.NewRegistry()
	defer storeRegistry.Close()

	sqliteProvider := sqlitedb.NewProvider(cfg.SQLitePath)
	storeRegistry.RegisterProvider(store.SQLite, sqliteProvider)

	redisProvider := redisdb.NewProvider(&cfg.Redis)
	storeRegistry.RegisterProvider(store.Redis, redisProvider)

	tinylinkService := services.NewTinylinkService(
		storeRegistry.GetTinylinkStore(store.Redis),
		storeRegistry.GetTinylinkStore(store.SQLite),
	)
	userService := services.NewUserService(storeRegistry.GetUserStore(store.SQLite))

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

	app.setupRouter()

	if err := app.serve(); err != nil {
		logger.Error("app.serve()", "error", err)
	}
}
