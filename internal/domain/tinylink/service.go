package tinylink

import (
	"context"
	"errors"
	"fmt"
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

func (s *Service) List(ctx context.Context, claims *token.Claims) ([]*Tinylink, error) {
	return s.db.List(ctx, claims.UserID)
}

func (s *Service) checkIfAliasExists(ctx context.Context, userID *string, alias string) error {
	exists, err := s.redis.Exists(ctx, alias)
	if err != nil && err != data.ErrNotFound {
		return err
	}
	if exists {
		return ErrAliasExists
	}

	exists, err = s.db.Exists(ctx, userID, alias)
	fmt.Println("EXISTS FROM SQLITE: ", exists, userID, alias)
	if err != nil {
		return err
	}
	if exists {
		return ErrAliasExists
	}

	return nil
}

func (s *Service) Insert(ctx context.Context, claims *token.Claims, req InsertTinylinkRequest) (*Tinylink, error) {
	tl := &Tinylink{
		URL:     req.URL,
		Alias:   req.Alias,
		Domain:  &req.Domain,
		Private: req.Private,
	}

	var userID *string
	if claims != nil {
		userID = &claims.UserID
	}

	if tl.Alias == "" {
		alias, err := s.redis.GenerateAlias(ctx)
		if err != nil {
			return nil, err
		}
		tl.Alias = alias
	} else {
		if err := s.checkIfAliasExists(ctx, userID, tl.Alias); err != nil {
			return nil, err
		}
	}

	hasUserID := claims != nil && claims.UserID != ""
	if hasUserID {
		tl.UserID = &claims.UserID
		if err := s.db.Insert(ctx, tl); err != nil {
			return nil, err
		}
	} else {
		tl.Private = false
		tl.ExpiresAt = time.Now().Add(time.Duration(anonTTL)).Unix()
		if err := s.redis.Insert(ctx, tl); err != nil {
			return nil, err
		}
	}

	return tl, nil
}

func (s *Service) Update(ctx context.Context, claims *token.Claims, req UpdateTinylinkRequest) (*Tinylink, error) {
	tl := &Tinylink{
		ID:      req.ID,
		Alias:   req.Alias,
		Domain:  &req.Domain,
		Private: req.Private,
		UserID:  &claims.UserID,
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

// Cache?????
func (s *Service) RedirectPersonal(ctx context.Context, claims *token.Claims, alias string) (string, error) {
	var URL string
	err := s.provider.WithTransaction(func(dbAdapters DBAdapters) error {
		rowID, url, err := dbAdapters.TinylinkDBRepository.RedirectPersonal(ctx, claims.UserID, alias)
		if err != nil {
			return err
		}
		URL = url
		return dbAdapters.TinylinkDBRepository.UpdateUsage(ctx, rowID)
	})
	if err != nil {
		return "", err
	}
	return URL, nil
}

func (s *Service) Redirect(ctx context.Context, alias string) (string, error) {
	var rowID uint64
	var url string
	var err error

	err = s.provider.WithTransaction(func(dbAdapters DBAdapters) error {
		rowID, url, err = s.redis.Redirect(ctx, alias)
		if err != nil && !errors.Is(err, data.ErrNotFound) {
			return err
		}
		if url == "" {
			if rowID, url, err = dbAdapters.TinylinkDBRepository.Redirect(ctx, alias); err != nil {
				return err
			}
			return s.redis.Insert(ctx, &Tinylink{ID: rowID, Alias: alias, URL: url, ExpiresAt: int64(cacheTTL)})
		}

		if url == "" {
			return data.ErrNotFound
		}

		if err := dbAdapters.TinylinkDBRepository.UpdateUsage(ctx, rowID); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return url, nil
}

// func (s *Service) Get(ctx context.Context, alias string) (*Tinylink, error) {
// 	var tl *Tinylink
// 	var err error
// 	err = s.provider.WithTransaction(func(dbAdapters DBAdapters) error {
// 		tl, err = s.tinylinkRedis.Get(ctx, alias)
// 		if err != nil && !errors.Is(err, data.ErrNotFound) {
// 			return err
// 		}
// 		if tl == nil {
// 			if tl, err = dbAdapters.TinylinkDBRepository.Get(ctx, alias); err != nil {
// 				return err
// 			}
// 			if err := s.tinylinkRedis.Insert(ctx, tl); err != nil {
// 				return err
// 			}
// 		}
// 		if err := dbAdapters.TinylinkDBRepository.UpdateUsage(ctx, tl); err != nil {
// 			return err
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return tl, nil
// }

func (s *Service) Delete(ctx context.Context, claims *token.Claims, alias string) error {
	return s.db.Delete(ctx, claims.UserID, alias)
}
