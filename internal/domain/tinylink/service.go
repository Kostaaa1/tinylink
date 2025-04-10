package tinylink

import (
	"context"
	"errors"

	"github.com/Kostaaa1/tinylink/internal/common/auth"
	"github.com/Kostaaa1/tinylink/internal/common/data"
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

func (s *Service) List(ctx context.Context, claims *auth.Claims) ([]*Tinylink, error) {
	return s.tinylinkDb.List(ctx, claims.UserID)
}

func (s *Service) Insert(ctx context.Context, claims *auth.Claims, req InsertTinylinkRequest) (*Tinylink, error) {
	tl := &Tinylink{
		OriginalURL: req.OriginalURL,
		Alias:       req.Alias,
		Domain:      req.Domain,
		Private:     req.Private,
		UserID:      claims.UserID,
	}

	if tl.Alias == "" {
		alias, err := s.tinylinkRedis.GenerateAlias(ctx)
		if err != nil {
			return nil, err
		}
		tl.Alias = alias
	}

	var err error
	if tl.UserID != "" {
		err = s.tinylinkDb.Insert(ctx, tl)
	} else {
		err = s.tinylinkRedis.Insert(ctx, tl)
	}

	if err != nil {
		return nil, err
	}

	return tl, nil
}

func (s *Service) Update(ctx context.Context, claims *auth.Claims, req UpdateTinylinkRequest) (*Tinylink, error) {
	tl := &Tinylink{
		ID:      req.ID,
		Alias:   req.Alias,
		Domain:  req.Domain,
		Private: req.Private,
		UserID:  claims.UserID,
	}
	if err := s.getStore(ctx).Update(ctx, tl); err != nil {
		return nil, err
	}
	return tl, nil
}

func (s *Service) GetPersonal(ctx context.Context, claims *auth.Claims, alias string) (*Tinylink, error) {
	var tl *Tinylink

	err := s.txProvider.WithTransaction(func(dbAdapters DBAdapters) error {
		tl, err := dbAdapters.TinylinkDBRepository.GetByUserID(ctx, claims.UserID, alias)
		if err != nil {
			return err
		}
		return dbAdapters.TinylinkDBRepository.UpdateUsage(ctx, tl)
	})

	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (s *Service) Get(ctx context.Context, alias string) (*Tinylink, error) {
	var tl *Tinylink
	var err error

	err = s.txProvider.WithTransaction(func(dbAdapters DBAdapters) error {
		tl, err = s.tinylinkRedis.Get(ctx, alias)
		if err != nil && !errors.Is(err, data.ErrNotFound) {
			return err
		}

		if tl == nil {
			if tl, err = dbAdapters.TinylinkDBRepository.Get(ctx, alias); err != nil {
				return err
			}
			if err := s.tinylinkRedis.Insert(ctx, tl); err != nil {
				return err
			}
		}

		if err := dbAdapters.TinylinkDBRepository.UpdateUsage(ctx, tl); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tl, nil
}

func (s *Service) Delete(ctx context.Context, claims *auth.Claims, alias string) error {
	return s.getStore(ctx).Delete(ctx, claims.UserID, alias)
}
