package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	myerr "github.com/Kostaaa1/tinylink/internal/errors"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/middleware"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/middleware/ratelimiter"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/middleware/session"
	"github.com/Kostaaa1/tinylink/internal/interface/utils/jsonutil"
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
	// Add a 4 second delay.
	time.Sleep(4 * time.Second)
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

	router.HandleFunc("/healthcheck", a.healthcheck)

	return router
}

func (app *app) serve(r *mux.Router) error {
	// Initialize the HTTP server with configuration settings.
	srv := &http.Server{
		Addr:         ":" + app.cfg.Port,
		Handler:      r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Create a channel to receive errors from the graceful shutdown process.
	shutdownError := make(chan error)

	// Start a goroutine to handle server shutdown gracefully.
	go func() {
		// Create a channel to listen for interrupt or termination signals.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Block until a signal is received.
		s := <-quit

		// Log the shutdown signal.
		app.log.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		// Create a context with a 5-second timeout for the shutdown process.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Attempt to gracefully shut down the server and send any errors to the shutdownError channel.
		shutdownError <- srv.Shutdown(ctx)
	}()

	// Log that the server is starting.
	app.log.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  "dev",
	})

	// Start the server and listen for incoming requests.
	err := srv.ListenAndServe()

	// If the error is NOT http.ErrServerClosed, it means the server failed to start.
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Wait for the shutdown process to complete and check for errors.
	err = <-shutdownError
	if err != nil {
		return err
	}

	// Log that the server has stopped successfully.
	app.log.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}

// func (a *app) serve(r *mux.Router) error {
// 	srv := &http.Server{
// 		Addr:           ":" + a.cfg.Port,
// 		Handler:        r,
// 		IdleTimeout:    1 * time.Minute,
// 		ReadTimeout:    10 * time.Second,
// 		WriteTimeout:   30 * time.Second,
// 		MaxHeaderBytes: 2 * 1024 * 1024,
// 	}

// 	shutdownError := make(chan error, 1)

// 	go func() {
// 		quit := make(chan os.Signal, 1)

// 		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

// 		s := <-quit

// 		a.log.PrintInfo("shuting down server", map[string]string{
// 			"signal": s.String(),
// 		})

// 		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 		defer cancel()

// 		shutdownError <- srv.Shutdown(ctx)
// 	}()

// 	a.log.PrintInfo("starting server", map[string]string{
// 		"port": srv.Addr,
// 		"env":  "development",
// 	})

// 	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
// 		return err
// 	}

// 	err := <-shutdownError
// 	if err != nil {
// 		return err
// 	}

// 	a.log.PrintInfo("stopped server", map[string]string{
// 		"addr": srv.Addr,
// 	})

// 	return nil
// }
