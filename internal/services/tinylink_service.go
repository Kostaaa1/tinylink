package services

import (
	"context"

	"github.com/Kostaaa1/tinylink/db"
	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/middleware/auth"
)

type TinylinkService struct {
	PrimaryStore db.TinylinkStore
	CacheStore   db.TinylinkStore
	Token        db.TokenStore
}

func NewTinylinkService(primaryStore, cacheStore db.TinylinkStore, tokenStore db.TokenStore) *TinylinkService {
	return &TinylinkService{
		PrimaryStore: primaryStore,
		CacheStore:   cacheStore,
		Token:        tokenStore,
	}
}

func (s *TinylinkService) getStore(ctx context.Context) db.TinylinkStore {
	if auth.IsAuthenticated(ctx) {
		return s.PrimaryStore
	}
	return s.CacheStore
}

func (s *TinylinkService) List(ctx context.Context, userID string) ([]*data.Tinylink, error) {
	links, err := s.getStore(ctx).List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return links, nil
}

func (s *TinylinkService) Create(ctx context.Context, URL, alias string) (*data.Tinylink, error) {
	token := auth.AuthTokenFromCtx(ctx)

	tinylink, err := data.NewTinylink(token.UserID, "http://localhost:3000", URL, alias)
	if err != nil {
		return nil, err
	}

	store := s.getStore(ctx)
	if err := store.Save(ctx, tinylink); err != nil {
		return nil, err
	}

	return tinylink, nil
}

func (s *TinylinkService) Get(ctx context.Context, userID, alias string) (*data.Tinylink, error) {
	tl, err := s.getStore(ctx).Get(ctx, userID, alias)
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (s *TinylinkService) Delete(ctx context.Context, userID, alias string) error {
	return s.getStore(ctx).Delete(ctx, userID, alias)
}
