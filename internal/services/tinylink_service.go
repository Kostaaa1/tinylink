package services

import (
	"context"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/Kostaaa1/tinylink/db"
	"github.com/Kostaaa1/tinylink/internal/data"
)

type TinylinkService struct {
	primaryStore db.TinylinkStore
	cacheStore   db.TinylinkStore
}

func NewTinylinkService(primaryStore, cacheStore db.TinylinkStore) *TinylinkService {
	return &TinylinkService{
		primaryStore: primaryStore,
		cacheStore:   cacheStore,
	}
}

func (s *TinylinkService) getStore() db.TinylinkStore {
	// if false {
	// 	return s.primaryStore
	// }
	return s.cacheStore
}

func (s *TinylinkService) List(ctx context.Context, userID string) ([]*data.Tinylink, error) {
	links, err := s.getStore().List(ctx, userID)
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second * 15)
	return links, nil
}

func (s *TinylinkService) Save(ctx context.Context, userID, URL, alias string) (*data.Tinylink, error) {
	if alias == "" {
		s := userID + URL
		alias = fmt.Sprintf("%x", sha1.Sum([]byte(s)))[:8]
	}

	tl, err := data.NewTinylink("http://localhost:3000", URL, alias)
	if err != nil {
		return nil, err
	}

	// zakucano
	ttl := time.Duration(5 * time.Minute)
	if err := s.getStore().Save(ctx, tl, userID, ttl); err != nil {
		return nil, err
	}
	return tl, nil
}

func (s *TinylinkService) Get(ctx context.Context, userID, alias string) (*data.Tinylink, error) {
	tl, err := s.getStore().Get(ctx, userID, alias)
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (s *TinylinkService) Delete(ctx context.Context, userID, alias string) error {
	return s.getStore().Delete(ctx, userID, alias)
}
