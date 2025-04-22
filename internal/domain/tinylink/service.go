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
	Adapters() Adapters
}

type Service struct {
	provider provider
	db       DBRepository
	redis    RedisRepository
}

func NewService(provider provider) *Service {
	adapters := provider.Adapters()
	return &Service{
		provider: provider,
		db:       adapters.TinylinkDBRepository,
		redis:    adapters.TinylinkRedisRepository,
	}
}

func (s *Service) List(ctx context.Context, sessionID string, claims token.Claims) ([]*Tinylink, error) {
	if claims.UserID != "" {
		return s.db.ListUserLinks(ctx, claims.UserID)
	}
	if sessionID != "" {
		return s.redis.ListUserLinks(ctx, sessionID)
	}
	return nil, data.ErrUnauthenticated
}

func (s *Service) isAliasValid(ctx context.Context, sessionID, userID string, alias string, isPrivate bool) error {
	if !isPrivate {
		exists, err := s.redis.Exists(ctx, sessionID, alias)
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

func (s *Service) Create(ctx context.Context, userID, sessionID string, req CreateTinylinkRequest) (*Tinylink, error) {
	if userID == "" && sessionID == "" {
		return nil, data.ErrUnauthenticated
	}

	tl := &Tinylink{
		URL:       req.URL,
		Alias:     req.Alias,
		Private:   req.Private,
		Domain:    req.Domain,
		CreatedAt: time.Now().Unix(),
	}

	if tl.Alias == "" {
		alias, err := s.redis.GenerateAlias(ctx)
		if err != nil {
			return nil, err
		}
		tl.Alias = alias
	} else {
		if err := s.isAliasValid(ctx, sessionID, userID, tl.Alias, tl.Private); err != nil {
			return nil, err
		}
	}

	switch {
	case userID != "":
		tl.UserID = userID
		if err := s.db.Create(ctx, tl); err != nil {
			return nil, err
		}
	case sessionID != "":
		tl.Private = false
		if err := s.redis.StoreBySessionID(ctx, sessionID, ToMap(tl)); err != nil {
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

	// If not found, maybe check redis?
	err := s.provider.Adapters().TinylinkDBRepository.Update(ctx, tl)
	if err != nil {
		return nil, err
	}

	return tl, nil
}

func (s *Service) Redirect(ctx context.Context, userID *string, alias string, isPrivate bool) (uint64, string, error) {
	var rowID uint64
	var url string
	var err error

	// check cache first, if found return
	rowID, url, err = s.redis.GetURL(ctx, alias)
	if err != nil && !errors.Is(err, data.ErrNotFound) {
		return 0, "", err
	}

	if url == "" {
		err = s.provider.WithTransaction(func(dbAdapters DBAdapters) error {
			// based on route (/p/ or public), GET the URL by alias/alias-user-id.
			if isPrivate && *userID != "" {
				if rowID, url, err = dbAdapters.TinylinkDBRepository.GetPersonalURL(ctx, *userID, alias); err != nil {
					return err
				}
			} else {
				if rowID, url, err = dbAdapters.TinylinkDBRepository.GetURL(ctx, alias); err != nil {
					return err
				}
			}

			// If found, insert it into redis cache for faster lookups
			if url != "" {
				if err := s.redis.CacheURL(ctx, rowID, alias, url); err != nil {
					return err
				}
				return nil
			}

			return data.ErrNotFound
		})
	}

	if err != nil {
		return 0, "", err
	}

	return rowID, url, nil
}

func (s *Service) Delete(ctx context.Context, claims token.Claims, alias string) error {
	return s.db.Delete(ctx, claims.UserID, alias)
}
