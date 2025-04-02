package tinylink

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/common/auth"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
)

type Service struct {
	primary DBRepository
	cache   RedisRepository
	token   token.Repository
}

func NewService(primary DBRepository, cache RedisRepository, token token.Repository) *Service {
	return &Service{
		primary: primary,
		cache:   cache,
		token:   token,
	}
}

func (s *Service) getStore(ctx context.Context) Repository {
	if auth.IsAuthenticated(ctx) {
		return s.primary
	}
	return s.cache
}

func (s *Service) List(ctx context.Context, userID string) ([]*Tinylink, error) {
	return s.getStore(ctx).List(ctx, userID)
}

func (s *Service) Insert(ctx context.Context, alias, originalURL, domain string, private bool) (*Tinylink, error) {
	// token := auth.TokenFromCtx(ctx)

	// tl := &Tinylink{
	// 	OriginalURL: originalURL,
	// 	Alias:       alias,
	// 	UserID:      token.UserID,
	// 	Domain:      domain,
	// 	Private:     private,
	// 	UsageCount:  0,
	// }

	// if err := s.getStore(ctx).Insert(ctx, tl); err != nil {
	// 	return nil, err
	// }

	// return tl, nil
	return nil, nil
}

func (s *Service) Update(ctx context.Context, id uint64, alias, domain string, private bool) (*Tinylink, error) {
	tl := &Tinylink{
		ID:      id,
		Alias:   alias,
		Domain:  domain,
		Private: private,
	}
	if err := s.getStore(ctx).Update(ctx, tl); err != nil {
		return nil, err
	}
	return tl, nil
}

func (s *Service) Get(ctx context.Context, alias string) (*Tinylink, error) {
	// user := auth.UserFromCtx(ctx)
	// userID := user.GetID()
	// if userID == "" {
	// 	// getPublic should work only for sqlite
	// 	tl, err := s.primary.GetPublic(ctx, alias)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return tl, nil
	// }
	// tl, err := s.getStore(ctx).Get(ctx, userID, alias)
	// if err != nil {
	// 	return nil, err
	// }
	// if err := s.primary.IncrementUsageCount(ctx, alias); err != nil {
	// 	return nil, err
	// }
	// return tl, nil

	return nil, nil
}

func (s *Service) Delete(ctx context.Context, userID, alias string) error {
	return s.getStore(ctx).Delete(ctx, userID, alias)
}
