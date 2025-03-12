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

	"github.com/Kostaaa1/tinylink/api/handlers"
	"github.com/Kostaaa1/tinylink/internal/middleware"
	"github.com/Kostaaa1/tinylink/internal/middleware/ratelimiter"
	"github.com/Kostaaa1/tinylink/internal/middleware/session"
)

func (app *application) setupRouter() {
	app.router.MethodNotAllowedHandler = http.HandlerFunc(handlers.MethodNotAllowedResponse)
	app.router.NotFoundHandler = http.HandlerFunc(handlers.NotFoundResponse)

	limit := ratelimiter.New(app.cfg.Limiter)
	app.router.Use(middleware.RecoverPanic, limit.Middleware, session.Middleware)

	app.router.HandleFunc("/health", handlers.HealthcheckHandler)

	tlHandler := handlers.NewTinylinkHandler(app.tinylinkService)
	app.router.HandleFunc("/getAll", tlHandler.List).Methods("GET")
	app.router.HandleFunc("/create", tlHandler.Save).Methods("POST")
	app.router.HandleFunc("/{alias}", tlHandler.Redirect).Methods("GET")
	app.router.HandleFunc("/{alias}", tlHandler.Delete).Methods("DELETE")

	userHandler := handlers.NewUserHandler(app.userService)
	userRoutes := app.router.PathPrefix("/users").Subrouter()
	userRoutes.HandleFunc("/register", userHandler.Register).Methods("POST")
	// userRoutes.HandleFunc("/{id}", userHandler.GetByID).Methods("GET")
	// userRoutes.HandleFunc("/create", userHandler.Create).Methods("POST")
}

func (app *application) serve() error {
	if !strings.HasPrefix(app.cfg.Port, ":") {
		app.cfg.Port = ":" + app.cfg.Port
	}

	srv := &http.Server{
		Addr:         "0.0.0.0" + app.cfg.Port,
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
		app.log.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownError <- srv.Shutdown(ctx)
	}()

	app.log.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  "dev",
	})

	err := srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.log.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
