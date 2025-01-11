package repository

import "github.com/Kostaaa1/tinylink/internal/data"

type Store interface {
	GetAll(key string) (map[string]data.Tinylink, error)
	Delete()
	Set(id, key string, data interface{}) error
	Check() bool
	Get(key, shortURL string) (data.Tinylink, error)
	Ping() error
}
