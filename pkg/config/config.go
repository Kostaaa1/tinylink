package config

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
	Password string
	DB       int
	PoolSize int
}
