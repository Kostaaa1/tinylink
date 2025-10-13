package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kostaaa1/tinylink/core/transactor"
	"github.com/Kostaaa1/tinylink/core/transactor/adapters"
	tinylinkHandler "github.com/Kostaaa1/tinylink/internal/api/tinylink"
	userHandler "github.com/Kostaaa1/tinylink/internal/api/user"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/Kostaaa1/tinylink/internal/infra/postgres"
	"github.com/Kostaaa1/tinylink/internal/infra/redis"
	"github.com/Kostaaa1/tinylink/pkg/errhandler"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type application struct {
	conf   Config
	router *mux.Router
	log    *slog.Logger
}

func (a *application) serve() error {
	srv := &http.Server{
		Addr:         ":" + a.conf.Port,
		Handler:      a.router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownErr := make(chan error, 1)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		a.log.Info("Shutting down server gracefully")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownErr <- srv.Shutdown(ctx)
	}()

	a.log.Info("Server started on port", "port", a.conf.Port, "env", a.conf.Env, "version", version)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return <-shutdownErr
}

func (a *application) registerUsers(
	db *pgxpool.Pool,
	tokenRepo token.Repository,
	errHandler errhandler.ErrorHandler,
	authMW mux.MiddlewareFunc,
) {
	userRepo := postgres.NewUserRepository(db)
	userProvider := transactor.NewProvider(userRepo, adapters.WithPgxPool(db))
	userService := user.NewService(userRepo, tokenRepo, userProvider)
	userHandler := userHandler.NewUserHandler(userService, errHandler, a.log)
	userHandler.RegisterRoutes(a.router, authMW)
}

func (a *application) registerTinylink(
	db *pgxpool.Pool,
	redisClient *goredis.Client,
	errHandler errhandler.ErrorHandler,
	authMW mux.MiddlewareFunc,
) {
	tlRepo := postgres.NewTinylinkRepository(db)
	tlCacheRepo := redis.NewTinylinkRepository(redisClient)
	tlProvider := transactor.NewProvider(tlRepo, adapters.WithPgxPool(db))
	tlService := tinylink.NewService(tlProvider, tlCacheRepo)
	tlHandler := tinylinkHandler.NewTinylinkHandler(tlService, errHandler, a.log)
	tlHandler.RegisterRoutes(a.router, authMW)
}

func (a *application) registerSwagger() {
	a.router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
}
