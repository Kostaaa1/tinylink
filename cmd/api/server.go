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

	"github.com/Kostaaa1/tinylink/api/utils/jsonutil"
	"github.com/Kostaaa1/tinylink/internal/middleware"
	"github.com/Kostaaa1/tinylink/internal/middleware/ratelimiter"
	"github.com/Kostaaa1/tinylink/internal/middleware/session"
	myerr "github.com/Kostaaa1/tinylink/pkg/errors"
	"github.com/gorilla/mux"
)

type envelope map[string]interface{}

func (a *app) healthcheck(w http.ResponseWriter, r *http.Request) {
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

func (a *app) setupRouter() *mux.Router {
	router := mux.NewRouter()

	router.MethodNotAllowedHandler = http.HandlerFunc(myerr.MethodNotAllowedResponse)
	router.NotFoundHandler = http.HandlerFunc(myerr.NotFoundResponse)

	limit := ratelimiter.New(a.cfg.Limiter)
	router.Use(middleware.RecoverPanic, limit.Middleware, session.Middleware)

	router.HandleFunc("/health", a.healthcheck)

	return router
}

func (app *app) serve(r *mux.Router) error {
	if !strings.HasPrefix(app.cfg.Port, ":") {
		app.cfg.Port = ":" + app.cfg.Port
	}

	srv := &http.Server{
		Addr:         app.cfg.Port,
		Handler:      r,
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
