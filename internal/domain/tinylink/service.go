package tinylink

import (
	"context"
	"errors"
	"time"

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
	provider provider
	db       DBRepository
	redis    RedisRepository
}

func NewService(provider provider) *Service {
	adapters := provider.GetAdapters()
	return &Service{
		provider: provider,
		db:       adapters.TinylinkDBRepository,
		redis:    adapters.TinylinkRedisRepository,
	}
}

func (s *Service) List(ctx context.Context, claims token.Claims) ([]*Tinylink, error) {
	return s.db.List(ctx, claims.UserID)
}

func (s *Service) checkAlias(ctx context.Context, userID string, alias string, isPrivate bool) error {
	if !isPrivate {
		exists, err := s.redis.Exists(ctx, alias)
		if err != nil && err != data.ErrNotFound {
			return err
		}
		if exists {
			return ErrAliasExists
		}
	}

	exists, err := s.db.Exists(ctx, userID, alias)
	if err != nil {
		return err
	}
	if exists {
		return ErrAliasExists
	}
	return nil
}

func (s *Service) Insert(ctx context.Context, claims token.Claims, req InsertTinylinkRequest) (*Tinylink, error) {
	tl := &Tinylink{
		URL:     req.URL,
		Alias:   req.Alias,
		Domain:  &req.Domain,
		Private: req.Private,
	}

	hasUserID := claims.UserID != ""
	if hasUserID {
		tl.UserID = claims.UserID
	} else {
		tl.Private = false
		tl.ExpiresAt = time.Now().Add(time.Duration(anonTTL)).Unix()
	}

	if tl.Alias == "" {
		alias, err := s.redis.GenerateAlias(ctx)
		if err != nil {
			return nil, err
		}
		tl.Alias = alias
	} else {
		if err := s.checkAlias(ctx, tl.UserID, tl.Alias, tl.Private); err != nil {
			return nil, err
		}
	}

	if hasUserID {
		if err := s.db.Insert(ctx, tl); err != nil {
			return nil, err
		}
	} else {
		if err := s.redis.Insert(ctx, tl); err != nil {
			return nil, err
		}
	}

	return tl, nil
}

func (s *Service) Update(ctx context.Context, claims token.Claims, req UpdateTinylinkRequest) (*Tinylink, error) {
	tl := &Tinylink{
		UserID:  claims.UserID,
		Private: req.Private,
	}
	if req.Domain != nil {
		tl.Domain = req.Domain
	}
	if req.Alias != nil {
		tl.Alias = *req.Alias
	}
	if req.URL != nil {
		tl.URL = *req.URL
	}

	err := s.provider.WithTransaction(func(dbAdapters DBAdapters) error {
		fetched, err := dbAdapters.TinylinkDBRepository.GetByUserID(ctx, claims.UserID, tl.Alias)
		if err != nil && err != data.ErrNotFound {
			return err
		}
		if fetched != nil {
			return err
		}
		return dbAdapters.TinylinkDBRepository.Update(ctx, tl)
	})

	if err != nil {
		return nil, err
	}

	return tl, nil
}

// Cache?????
func (s *Service) RedirectPersonal(ctx context.Context, claims token.Claims, alias string) (uint64, string, error) {
	rowID, url, err := s.db.RedirectPersonal(ctx, claims.UserID, alias)
	if err != nil {
		return 0, "", err
	}
	return rowID, url, nil
}

func (s *Service) Redirect(ctx context.Context, alias string) (uint64, string, error) {
	var rowID uint64
	var url string
	var err error

	rowID, url, err = s.redis.Redirect(ctx, alias)
	if err != nil && !errors.Is(err, data.ErrNotFound) {
		return 0, "", err
	}

	err = s.provider.WithTransaction(func(dbAdapters DBAdapters) error {
		if url == "" {
			if rowID, url, err = dbAdapters.TinylinkDBRepository.Redirect(ctx, alias); err != nil {
				return err
			}
			ttl := time.Now().Add(cacheTTL).Unix()
			if err := s.redis.Insert(ctx, &Tinylink{ID: rowID, Alias: alias, URL: url, ExpiresAt: ttl}); err != nil {
				return err
			}
		}
		return data.ErrNotFound
	})

	if err != nil {
		return 0, "", err
	}

	return rowID, url, nil
}

func (s *Service) Delete(ctx context.Context, claims token.Claims, alias string) error {
	return s.db.Delete(ctx, claims.UserID, alias)
}
