package main

import (
	"flag"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/db/redisdb"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/db/sqlitedb"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/handler"
	"github.com/Kostaaa1/tinylink/internal/middleware"
	"github.com/Kostaaa1/tinylink/internal/middleware/ratelimiter"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/Kostaaa1/tinylink/pkg/errorhandler"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type application struct {
	conf    *config.Config
	handler handler.Handler
	router  *mux.Router
	logger  *slog.Logger
}

var (
	conf config.Config
)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	flag.StringVar(&conf.Port, "port", "3000", "server address port")
	flag.StringVar(&conf.Env, "env", "development", "environment (development|staging|production)")
	flag.Float64Var(&conf.Limiter.RPS, "limiter-rps", 2, "rate limiter requests-per-second")
	flag.IntVar(&conf.Limiter.Burst, "limiter-burst", 4, "rate limiter maximum burst")
	flag.BoolVar(&conf.Limiter.Enabled, "limiter-enabled", true, "enable rate limiter")

	flag.StringVar(&conf.SQL.DSN, "dsn", "tinylink.db", "database connection string")
	flag.StringVar(&conf.SQL.SQLitePath, "sqlite-db-path", "tinylink.db", "path to the sqlite database")
	flag.IntVar(&conf.SQL.MaxIdleConns, "sql-max-idle-conns", 25, "database connection string")
	flag.IntVar(&conf.SQL.MaxOpenConns, "sql-max-open-conns", 25, "path to the sqlite database")

	flag.BoolVar(&conf.Redis.Enabled, "redis-enabled", false, "enable redis")
	flag.StringVar(&conf.Redis.Addr, "redis-addr", "localhost:6379", "redis server address")
	flag.StringVar(&conf.Redis.Password, "redis-password", "", "redis password")
	flag.IntVar(&conf.Redis.DB, "redis-db", 0, "redis database number")
	flag.IntVar(&conf.Redis.PoolSize, "redis-pool-size", 10, "redis connection pool size")

	flag.Parse()
}

func setupLogger(w io.Writer, conf *config.Config) *slog.Logger {
	var logHandler slog.Handler
	if conf.Env == "development" {
		logHandler = slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		logHandler = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelError})
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)
	return logger
}

func main() {
	logger := setupLogger(os.Stdout, &conf)

	db, err := sqlitedb.StartDB(conf.SQL)
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := redisdb.StartRedis(conf.Redis)
	if err != nil {
		log.Fatal(err)
	}

	userRepoProvider := user.NewRepositoryProvider(db)
	userService := user.NewService(userRepoProvider)

	tinylinkProvider := tinylink.NewRepositoryProvider(db, redisClient)
	tinylinkService := tinylink.NewService(tinylinkProvider)

	// errHandler := handler.NewErrorHandler(logger)
	errHandler := errorhandler.New(logger)
	userHandler := handler.NewUserHandler(userService, errHandler)
	tinylinkHandler := handler.NewTinylinkHandler(tinylinkService, errHandler)

	app := application{
		conf:   &conf,
		logger: logger,
		handler: handler.Handler{
			ErrorHandler: errHandler,
			User:         userHandler,
			Tinylink:     tinylinkHandler,
		},
	}

	r := mux.NewRouter()
	r.MethodNotAllowedHandler = http.HandlerFunc(app.handler.MethodNotAllowedResponse)
	r.NotFoundHandler = http.HandlerFunc(app.handler.NotFoundResponse)

	limit := ratelimiter.New(app.conf.Limiter)
	mw := middleware.MW{ErrorHandler: errHandler}
	r.Use(mw.RecoverPanic, limit.Middleware, mw.Logger)

	app.handler.User.RegisterRoutes(r, mw)
	app.handler.Tinylink.RegisterRoutes(r, mw)
	app.router = r

	if err := app.serve(); err != nil {
		logger.Error("app.serve()", "error", err)
	}
}
