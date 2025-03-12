package services

import (
	"context"
	"crypto/sha1"
	"fmt"

	"github.com/Kostaaa1/tinylink/internal/domain/entities"
	"github.com/Kostaaa1/tinylink/internal/store"
)

type TinylinkService struct {
	tinylinkRepo store.TinylinkRepository
}

func NewTinylinkService(tinylinkRepo store.TinylinkRepository) *TinylinkService {
	return &TinylinkService{tinylinkRepo: tinylinkRepo}
}

func (s *TinylinkService) List(ctx context.Context, sessionID string) ([]*entities.Tinylink, error) {
	links, err := s.tinylinkRepo.List(ctx, entities.QueryParams{SessionID: sessionID})
	if err != nil {
		return nil, err
	}
	return links, nil
}

func (s *TinylinkService) Save(ctx context.Context, sessionID, URL, alias string) (*entities.Tinylink, error) {
	if alias == "" {
		s := sessionID + URL
		alias = fmt.Sprintf("%x", sha1.Sum([]byte(s)))[:8]
	} else {
		if err := s.tinylinkRepo.SetAlias(ctx, alias); err != nil {
			return nil, err
		}
	}

	tl, err := entities.NewTinylink("http://localhost:3000", URL, alias)
	if err != nil {
		return nil, err
	}

	qp := entities.QueryParams{SessionID: sessionID, Alias: alias}
	if err := s.tinylinkRepo.Save(ctx, tl, qp); err != nil {
		return nil, err
	}

	return tl, nil
}

func (s *TinylinkService) Get(ctx context.Context, sessionID, alias string) (*entities.Tinylink, error) {
	tl, err := s.tinylinkRepo.Get(ctx, entities.QueryParams{SessionID: sessionID, Alias: alias})
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (s *TinylinkService) Delete(ctx context.Context, sessionID, alias string) error {
	return s.tinylinkRepo.Delete(ctx, entities.QueryParams{SessionID: sessionID, Alias: alias})
}
