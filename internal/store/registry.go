package store

import (
	"io"
	"sync"
)

type StorageType string

var (
	Redis  StorageType = "redis"
	SQLite StorageType = "sqlite"
)

type Registry struct {
	mutex          sync.Mutex
	providers      []io.Closer
	tinylinkStores map[StorageType]TinylinkStore
	userStores     map[StorageType]UserStore
}

func NewRegistry() *Registry {
	return &Registry{
		providers:      make([]io.Closer, 0),
		tinylinkStores: make(map[StorageType]TinylinkStore),
		userStores:     make(map[StorageType]UserStore),
	}
}

func (r *Registry) RegisterProvider(name StorageType, provider interface{}) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if closer, ok := provider.(io.Closer); ok {
		r.providers = append(r.providers, closer)
	}

	// Register individual stores with their provider name
	// Type-assert and register stores
	if p, ok := provider.(interface{ Tinylink() TinylinkStore }); ok {
		r.tinylinkStores[name] = p.Tinylink()
	}
	if p, ok := provider.(interface{ User() UserStore }); ok {
		r.userStores[name] = p.User()
	}
}

func (r *Registry) GetTinylinkStore(provider StorageType) TinylinkStore {
	return r.tinylinkStores[provider]
}

func (r *Registry) GetUserStore(provider StorageType) UserStore {
	return r.userStores[provider]
}

func (r *Registry) Close() error {
	var lastErr error
	for _, provider := range r.providers {
		if err := provider.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
