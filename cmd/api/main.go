package main

import (
	"flag"
	"os"

	"github.com/Kostaaa1/tinylink/api/handlers"
	redisdb "github.com/Kostaaa1/tinylink/internal/repositories/redis"
	"github.com/Kostaaa1/tinylink/internal/services"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/Kostaaa1/tinylink/pkg/jsonlog"
	"github.com/joho/godotenv"
)

type app struct {
	cfg *config.Config
	log *jsonlog.Logger
}

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func main() {
	var cfg config.Config

	flag.StringVar(&cfg.Port, "port", "3000", "Server address port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.StorageType, "storage-type", "redis", "Storage (redis|sqlite|pocketbase)")
	flag.Float64Var(&cfg.Limiter.RPS, "limiter-rps", 2, "Rate limiter requests-per-second")
	flag.IntVar(&cfg.Limiter.Burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.Limiter.Enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Parse()

	log := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	a := app{
		cfg: &cfg,
		log: log,
	}

	store := redisdb.NewRedisStore(&cfg)
	router := a.setupRouter()

	tinylinkService := services.NewTinylinkService(store.Tinylink)
	handlers.NewTinylinkHandler(router, tinylinkService)

	userService := services.NewUserService(store.User)
	handlers.NewUserHandler(router, userService)

	if err := a.serve(router); err != nil {
		a.log.PrintFatal(err, nil)
	}
}
