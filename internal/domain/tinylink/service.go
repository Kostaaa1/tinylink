package tinylink

import (
	"context"
	"errors"

	"github.com/Kostaaa1/tinylink/internal/common/authcontext"
	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
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

type provider interface {
	WithTransaction(txFunc func(dbAdapters DBAdapters) error) error
	GetAdapters() Adapters
}

type Service struct {
	provider      provider
	tinylinkDb    DBRepository
	tinylinkRedis RedisRepository
}

func NewService(provider provider) *Service {
	adapters := provider.GetAdapters()
	return &Service{
		provider:      provider,
		tinylinkDb:    adapters.TinylinkDBRepository,
		tinylinkRedis: adapters.TinylinkRedisRepository,
	}
}

func (s *Service) getStore(ctx context.Context) Repository {
	if authcontext.IsAuthenticated(ctx) {
		return s.tinylinkDb
	}
	return s.tinylinkRedis
}

func (s *Service) List(ctx context.Context, claims *token.Claims) ([]*Tinylink, error) {
	return s.tinylinkDb.List(ctx, claims.UserID)
}

func (s *Service) Insert(ctx context.Context, claims *token.Claims, req InsertTinylinkRequest) (*Tinylink, error) {
	tl := &Tinylink{
		OriginalURL: req.OriginalURL,
		Alias:       req.Alias,
		Domain:      req.Domain,
		Private:     req.Private,
		UserID:      claims.UserID,
	}

	var err error

	if tl.Alias == "" {
		tl.Alias, err = s.tinylinkRedis.GenerateAlias(ctx)
	}

	if err != nil {
		return nil, err
	}

	err = s.provider.WithTransaction(func(dbAdapters DBAdapters) error {
		fetched, err := dbAdapters.TinylinkDBRepository.Get(ctx, tl.Alias)
		if err != nil && err != data.ErrNotFound {
			return err
		}
		if fetched != nil {
			return ErrAliasExists
		}
		return dbAdapters.TinylinkDBRepository.Insert(ctx, tl)
	})

	if err != nil {
		return nil, err
	}

	return tl, nil
}

func (s *Service) Update(ctx context.Context, claims *token.Claims, req UpdateTinylinkRequest) (*Tinylink, error) {
	tl := &Tinylink{
		ID:      req.ID,
		Alias:   req.Alias,
		Domain:  req.Domain,
		Private: req.Private,
		UserID:  claims.UserID,
	}

	s.provider.WithTransaction(func(dbAdapters DBAdapters) error {
		fetched, err := dbAdapters.TinylinkDBRepository.Get(ctx, tl.Alias)
		if err != nil && err != data.ErrNotFound {
			return err
		}
		if fetched != nil {
			return ErrAliasExists
		}
		return dbAdapters.TinylinkDBRepository.Update(ctx, tl)
	})

	return tl, nil
}

func (s *Service) GetPersonal(ctx context.Context, claims *token.Claims, alias string) (*Tinylink, error) {
	var tl *Tinylink

	err := s.provider.WithTransaction(func(dbAdapters DBAdapters) error {
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

	err = s.provider.WithTransaction(func(dbAdapters DBAdapters) error {
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

func (s *Service) Delete(ctx context.Context, claims *token.Claims, alias string) error {
	return s.tinylinkDb.Delete(ctx, claims.UserID, alias)
}
