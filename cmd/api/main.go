package main

import (
	"os"

	"github.com/Kostaaa1/tinylink/internal/application/services"
	"github.com/Kostaaa1/tinylink/internal/infrastructure"
	"github.com/Kostaaa1/tinylink/internal/interface/handlers"
	"github.com/Kostaaa1/tinylink/internal/jsonlog"
	"github.com/Kostaaa1/tinylink/pkg/config"
)

type app struct {
	cfg *config.Config
	log *jsonlog.Logger
}

func main() {
	cfg := config.New()
	log := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	a := app{
		cfg: cfg,
		log: log,
	}
	router := a.setupRouter()

	repos, err := infrastructure.NewRepositories(cfg)
	if err != nil {
		log.PrintFatal(err, nil)
	}
	tlService := services.NewTinylinkService(repos.Tinylink)
	handlers.NewTinylinkHandler(router, tlService)

	if err := a.serve(router); err != nil {
		a.log.PrintFatal(err, nil)
	}
}
