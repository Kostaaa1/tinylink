package config

type Config struct {
	Port    string
	Env     string
	Redis   RedisConfig
	SQL     SQLConfig
	Limiter RatelimitConfig
}

type SQLConfig struct {
	DSN          string
	SQLitePath   string
	MaxOpenConns int
	MaxIdleConns int
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
