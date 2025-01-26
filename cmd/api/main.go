package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Kostaaa1/tinylink/internal/application/services"
	"github.com/Kostaaa1/tinylink/internal/errors"
	"github.com/Kostaaa1/tinylink/internal/infrastructure"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/middleware"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/middleware/ratelimiter"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/middleware/session"
	"github.com/Kostaaa1/tinylink/internal/interface/handlers"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.New()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	r := mux.NewRouter()
	r.MethodNotAllowedHandler = http.HandlerFunc(errors.MethodNotAllowedResponse)
	r.NotFoundHandler = http.HandlerFunc(errors.NotFoundResponse)

	limit := ratelimiter.New(cfg.Limiter)
	r.Use(middleware.RecoverPanic, limit.Middleware, session.Middleware)

	repos, err := infrastructure.NewRepositories(cfg)
	if err != nil {
		panic(err)
	}
	handlers.NewTinylinkHandler(r, services.NewTinylinkService(repos.Tinylink))

	srv := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        r,
		IdleTimeout:    1 * time.Minute,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 2 * 1024 * 1024,
	}

	logger.Info("Server running on", "port", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
