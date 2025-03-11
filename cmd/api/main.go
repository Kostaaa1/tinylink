package main

import (
	"flag"
	"os"

	redisdb "github.com/Kostaaa1/tinylink/internal/repositories/redis"
	"github.com/Kostaaa1/tinylink/internal/services"
	"github.com/Kostaaa1/tinylink/internal/store"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/Kostaaa1/tinylink/pkg/jsonlog"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type application struct {
	cfg             *config.Config
	log             *jsonlog.Logger
	tinylinkService *services.TinylinkService
	userService     *services.UserService
	router          *mux.Router
}

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func NewStore(cfg *config.Config) *store.Store {
	switch cfg.StorageType {
	case "redis":
		return redisdb.NewRedisStore(cfg)
	// case "sqlite":
	// return redisdb.NewRedisStore(cfg)
	default:
		return redisdb.NewRedisStore(cfg)
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

	store := NewStore(&cfg)

	app := application{
		cfg:             &cfg,
		log:             log,
		tinylinkService: services.NewTinylinkService(store.Tinylink),
		userService:     services.NewUserService(store.User),
		router:          mux.NewRouter(),
	}

	app.setupRouter()

	if err := app.serve(); err != nil {
		app.log.PrintFatal(err, nil)
	}
}
