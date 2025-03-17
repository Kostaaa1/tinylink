package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// func (app *application) setupRouter() {
// 	r := mux.NewRouter()

// 	r.MethodNotAllowedHandler = http.HandlerFunc(app.handler.MethodNotAllowedResponse)
// 	r.NotFoundHandler = http.HandlerFunc(app.handler.NotFoundResponse)
// redisStore := redisdb.NewRedisStore(&cfg.Redis)
// auth.AuthMiddleware(&redisStore.Token)
// 	authMiddleware := auth.Middleware()
// 	limit := ratelimiter.New(app.cfg.Limiter)
// 	r.Use(middleware.RecoverPanic, limit.Middleware, authMiddleware)
// 	r.HandleFunc("/health", app.handler.HealthcheckHandler)

// 	app.handler.Tinylink.RegisterRoutes(r)
// 	app.handler.User.RegisterRoutes(r)

// 	app.router = r
// }

func (app *application) serve() error {
	if !strings.HasPrefix(app.cfg.Port, ":") {
		app.cfg.Port = ":" + app.cfg.Port
	}

	srv := &http.Server{
		Addr:         app.cfg.Port,
		Handler:      app.router,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 4 * time.Second,
		IdleTimeout:  time.Minute,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", "dev")

	err := srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("server stopped", "addr", srv.Addr)

	return nil
}
