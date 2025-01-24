package config

import (
	"flag"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Env         string
	StorageType string
	Redis       RedisConfig
	Limiter     RatelimitConfig
}

type RatelimitConfig struct {
	RPS     float64
	Burst   int
	Enabled bool
}

type RedisConfig struct {
	Addr     string
	Host     string
	Password string
	Port     string
}

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func New() *Config {
	var cfg Config

	flag.StringVar(&cfg.Port, "port", "3000", "Server address port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.StorageType, "storage-type", "redis", "Storage (redis|sqlite|pocketbase)")
	flag.Float64Var(&cfg.Limiter.RPS, "limiter-rps", 2, "Rate limiter requests-per-second")
	flag.IntVar(&cfg.Limiter.Burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.Limiter.Enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Parse()

	return &cfg
}
