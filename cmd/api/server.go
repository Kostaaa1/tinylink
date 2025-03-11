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
	"github.com/Kostaaa1/tinylink/api/utils/jsonutil"
	"github.com/Kostaaa1/tinylink/internal/middleware"
	"github.com/Kostaaa1/tinylink/internal/middleware/ratelimiter"
	"github.com/Kostaaa1/tinylink/internal/middleware/session"
	myerr "github.com/Kostaaa1/tinylink/pkg/errors"
)

type envelope map[string]interface{}

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": "devlopment",
			"version":     "1.0",
		},
	}
	err := jsonutil.WriteJSON(w, http.StatusOK, env, nil)
	if err != nil {
		myerr.ServerErrorResponse(w, r, err)
	}
}

func (app *application) setupRouter() {
	app.router.MethodNotAllowedHandler = http.HandlerFunc(myerr.MethodNotAllowedResponse)
	app.router.NotFoundHandler = http.HandlerFunc(myerr.NotFoundResponse)

	limit := ratelimiter.New(app.cfg.Limiter)
	app.router.Use(middleware.RecoverPanic, limit.Middleware, session.Middleware)

	app.router.HandleFunc("/health", app.healthcheck)

	tlHandler := handlers.NewTinylinkHandler(app.tinylinkService)
	app.router.HandleFunc("/getAll", tlHandler.List).Methods("GET")
	app.router.HandleFunc("/create", tlHandler.Save).Methods("POST")
	app.router.HandleFunc("/{alias}", tlHandler.Redirect).Methods("GET")
	app.router.HandleFunc("/{alias}", tlHandler.Delete).Methods("DELETE")
	// userHandler := handlers.NewTinylinkHandler(a.tinylinkService)
}

func (app *application) serve() error {
	if !strings.HasPrefix(app.cfg.Port, ":") {
		app.cfg.Port = ":" + app.cfg.Port
	}

	srv := &http.Server{
		Addr:         app.cfg.Port,
		Handler:      app.router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
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
