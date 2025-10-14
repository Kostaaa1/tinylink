package tinylink

import (
	"context"
	"errors"

	"github.com/Kostaaa1/tinylink/internal/constants"
	"github.com/Kostaaa1/tinylink/internal/domain/auth"
)

type CreateTinylinkParams struct {
	URL       string  `json:"url"`
	Alias     *string `json:"alias"`
	Domain    *string `json:"domain,omitempty"`
	Private   bool    `json:"private"`
	UserID    *uint64
	GuestUUID string
}

type UpdateTinylinkParams struct {
	ID      uint64  `json:"id"`
	URL     *string `json:"url"`
	UserID  uint64  `json:"user_id"`
	Alias   *string `json:"alias"`
	Domain  *string `json:"domain,omitempty"`
	Private bool    `json:"private"`
}

type Service struct {
	// repo     DbRepository
	// provider *transactor.Provider[DbRepository]
	cache CacheRepository
	repo  DbRepository
}

func NewService(dbRepo DbRepository, cacheRepo CacheRepository) *Service {
	return &Service{repo: dbRepo, cache: cacheRepo}
}

func (s *Service) List(ctx context.Context, userCtx auth.UserContext) ([]*Tinylink, error) {
	if userCtx.UserID != nil {
		return s.repo.ListByUserID(ctx, *userCtx.UserID)
	}
	return s.repo.ListByGuestUUID(ctx, userCtx.GuestUUID)
}

func (s *Service) Create(ctx context.Context, params CreateTinylinkParams) (*Tinylink, error) {
	// guestUUID always needs to be passed,
	if len(params.GuestUUID) == 0 {
		return nil, errors.Join(constants.ErrUnauthenticated, errors.New("missing guestUUID"))
	}

	if params.UserID == nil && params.Private {
		return nil, constants.ErrUnauthenticated
	}

	tl := &Tinylink{URL: params.URL, GuestUUID: params.GuestUUID}

	if params.Alias != nil {
		tl.Alias = *params.Alias
	} else {
		alias, err := s.cache.GenerateAlias(ctx)
		if err != nil {
			return nil, err
		}
		tl.Alias = alias
	}

	if params.Domain != nil {
		tl.Domain = *params.Domain
	}
	if params.UserID != nil {
		tl.UserID = params.UserID
	}

	if err := s.repo.Insert(ctx, tl); err != nil {
		return nil, err
	}

	return tl, nil
}

func (s *Service) Update(ctx context.Context, req UpdateTinylinkParams) (*Tinylink, error) {
	tl := &Tinylink{
		ID:      req.ID,
		Private: req.Private,
		UserID:  &req.UserID,
	}
	if req.Domain != nil {
		tl.Domain = *req.Domain
	}
	if req.Alias != nil {
		tl.Alias = *req.Alias
	}
	if req.URL != nil {
		tl.URL = *req.URL
	}

	if err := s.repo.Update(ctx, tl); err != nil {
		return nil, err
	}

	return tl, nil
}

// only authenticated users can delete their records. The cleanup will be based on in-active tinylinks
func (s *Service) Delete(ctx context.Context, userID uint64, alias string) error {
	return s.repo.Delete(ctx, userID, alias)
}

func (s *Service) Redirect(ctx context.Context, userID *uint64, alias string) (uint64, string, error) {
	val, err := s.cache.Redirect(ctx, alias)

	if err != nil && !errors.Is(err, constants.ErrNotFound) {
		return 0, "", err
	}

	if err == nil && val.URL != "" && val.RowID > 0 {
		return val.RowID, val.URL, nil
	}

	val, err = s.repo.Redirect(ctx, userID, alias)
	if err != nil {
		return 0, "", err
	}

	if val.URL == "" {
		return 0, "", constants.ErrNotFound
	}

	return val.RowID, val.URL, nil

	// implement some logic to collect metrics and analysis when redirect happens. Use something like rabitMQ, kafka, redis pub/sub, or even my own event pool. This should not impact performance in any way.
	// cache it - add hit count - implement worker pool

	// err = s.cache.Cache(ctx, RedirectValue{
	// 	RowID: val.RowID,
	// 	Alias: val.Alias,
	// 	URL:   val.URL,
	// }, defaultTTL)

	// if err != nil {
	// 	return 0, "", err
	// }
}

// func (s *Service) Redirect(ctx context.Context, userID *uint64, alias string) (uint64, string, error) {
// 	val, err := s.repo.Redirect(ctx, alias, userID)
// 	if err != nil {
// 		return 0, "", err
// 	}

// 	if val.URL == "" {
// 		return 0, "", constants.ErrNotFound
// 	}

// 	return val.RowID, val.URL, nil
// }
