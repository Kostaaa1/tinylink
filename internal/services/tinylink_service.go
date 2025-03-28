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
	return s.getStore(ctx).List(ctx, userID)
}

func (s *TinylinkService) Insert(ctx context.Context, req data.InsertTinylinkRequest) (*data.Tinylink, error) {
	token := auth.AuthTokenFromCtx(ctx)
	tl := &data.Tinylink{
		OriginalURL: req.OriginalURL,
		Alias:       req.Alias,
		UserID:      token.UserID,
		Domain:      req.Domain,
		Private:     req.Private,
		UsageCount:  0,
	}

	store := s.getStore(ctx)
	if err := store.Insert(ctx, tl); err != nil {
		return nil, err
	}

	return tl, nil
}

func (s *TinylinkService) Update(ctx context.Context, req data.UpdateTinylinkRequest) (*data.Tinylink, error) {
	tl := &data.Tinylink{
		ID:      req.ID,
		Alias:   req.Alias,
		Domain:  req.Domain,
		Private: req.Private,
	}
	if err := s.getStore(ctx).Update(ctx, tl); err != nil {
		return nil, err
	}
	return tl, nil
}

func (s *TinylinkService) Get(ctx context.Context, alias string) (*data.Tinylink, error) {
	user := auth.UserFromCtx(ctx)
	userID := user.GetID()

	if userID == "" {
		// getPublic should work only for sqlite
		tl, err := s.PrimaryStore.GetPublic(ctx, alias)
		if err != nil {
			return nil, err
		}
		return tl, nil
	}

	tl, err := s.getStore(ctx).Get(ctx, userID, alias)
	if err != nil {
		return nil, err
	}

	if err := s.getStore(ctx).IncrementUsageCount(ctx, alias); err != nil {
		return nil, err
	}

	return tl, nil
}

func (s *TinylinkService) Delete(ctx context.Context, userID, alias string) error {
	return s.getStore(ctx).Delete(ctx, userID, alias)
}
