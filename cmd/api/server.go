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

func (app *application) serve() error {
	if !strings.HasPrefix(app.conf.Port, ":") {
		app.conf.Port = ":" + app.conf.Port
	}

	srv := &http.Server{
		Addr:         app.conf.Port,
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
		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.Info("starting server",
		"addr", srv.Addr,
		"env", app.conf.Env,
	)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}

// func (app *application) serve() error {
// 	if !strings.HasPrefix(app.conf.Port, ":") {
// 		app.conf.Port = ":" + app.conf.Port
// 	}

// 	srv := &http.Server{
// 		Addr:         app.conf.Port,
// 		Handler:      app.router,
// 		ReadTimeout:  3 * time.Second,
// 		WriteTimeout: 4 * time.Second,
// 		IdleTimeout:  time.Minute,
// 	}

// 	shutdownError := make(chan error)

// 	go func() {
// 		quit := make(chan os.Signal, 1)
// 		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
// 		s := <-quit
// 		fmt.Println("received SIGNAL", s)

// 		app.logger.Info("shutting down server", "signal", s.String())

// 		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 		defer cancel()

// 		shutdownError <- srv.Shutdown(ctx)
// 	}()

// 	app.logger.Info("starting server", "addr", srv.Addr, "env", "dev")

// 	err := srv.ListenAndServe()

// 	if !errors.Is(err, http.ErrServerClosed) {
// 		return err
// 	}

// 	err = <-shutdownError
// 	if err != nil {
// 		return err
// 	}

// 	app.logger.Info("server stopped", "addr", srv.Addr)

// 	return nil
// }
