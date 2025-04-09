package tinylink

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/common/auth"
)

type DBAdapters struct {
	TinylinkDBRepository DBRepository
}

type RedisAdapters struct {
	TinylinkRedisRepository RedisRepository
}

type Adapters struct {
	DBAdapters
	RedisAdapters
}

type txProvider interface {
	WithTransaction(txFunc func(dbAdapters DBAdapters) error) error
	GetAdapters() Adapters
}

type Service struct {
	txProvider    txProvider
	tinylinkDb    DBRepository
	tinylinkRedis RedisRepository
}

func NewService(txProvider txProvider) *Service {
	adapters := txProvider.GetAdapters()
	return &Service{
		txProvider:    txProvider,
		tinylinkDb:    adapters.TinylinkDBRepository,
		tinylinkRedis: adapters.TinylinkRedisRepository,
	}
}

func (s *Service) getStore(ctx context.Context) Repository {
	if auth.IsAuthenticated(ctx) {
		return s.tinylinkDb
	}
	return s.tinylinkRedis
}

func (s *Service) List(ctx context.Context, userID string) ([]*Tinylink, error) {
	return nil, nil
}

func (s *Service) Insert(ctx context.Context, alias, originalURL, domain string, private bool) (*Tinylink, error) {
	claims := auth.ClaimsFromCtx(ctx)
	newTl := &Tinylink{
		UserID:      claims.ID,
		OriginalURL: originalURL,
		Alias:       alias,
		Domain:      domain,
		Private:     private,
		UsageCount:  0,
	}
	if err := s.getStore(ctx).Insert(ctx, newTl); err != nil {
		return nil, err
	}
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

// Auth user - can access public/private
// Non auth user - can access only public
// IF authenticated,
// If not, check public redis cache, if not found get from db. It found increment usage count
func (s *Service) Get(ctx context.Context, alias string) (*Tinylink, error) {
	claims := auth.ClaimsFromCtx(ctx)
	userID := claims.ID

	var err error
	var tl *Tinylink

	if userID == "" {
		tl, err = s.tinylinkRedis.Get(ctx, alias)
		if err != nil {
			return nil, err
		}
	} else {
		tl, err = s.tinylinkDb.GetByUserID(ctx, userID, alias)
		if err != nil {
			return nil, err
		}
	}

	if tl == nil {
		tl, err = s.tinylinkDb.Get(ctx, alias)
		if err != nil {
			return nil, err
		}
		if err := s.tinylinkDb.IncrementUsageCount(ctx, alias); err != nil {
			return nil, err
		}
	}

	return tl, nil
}

func (s *Service) Delete(ctx context.Context, userID, alias string) error {
	return s.getStore(ctx).Delete(ctx, userID, alias)
}
