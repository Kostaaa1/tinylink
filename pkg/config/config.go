package config

type Config struct {
	Port       string
	Env        string
	SQLitePath string
	Redis      RedisConfig
	Limiter    RatelimitConfig
}

type RatelimitConfig struct {
	RPS     float64
	Burst   int
	Enabled bool
}

type RedisConfig struct {
	Enabled  bool
	Addr     string
	Password string
	DB       int
	PoolSize int
}
