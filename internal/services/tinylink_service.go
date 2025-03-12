package services

import (
	"context"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/store"
)

type TinylinkService struct {
	authStore    store.TinylinkStore
	nonauthStore store.TinylinkStore
}

func NewTinylinkService(authStore, nonauthStore store.TinylinkStore) *TinylinkService {
	return &TinylinkService{
		authStore:    authStore,
		nonauthStore: nonauthStore,
	}
}

func (s *TinylinkService) getStore() store.TinylinkStore {
	// if false {
	// 	return s.authStore
	// }
	return s.nonauthStore
}

func (s *TinylinkService) List(ctx context.Context, sessionID string) ([]*data.Tinylink, error) {
	links, err := s.getStore().List(ctx, data.QueryParams{SessionID: sessionID})
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second * 15)
	return links, nil
}

func (s *TinylinkService) Save(ctx context.Context, sessionID, URL, alias string) (*data.Tinylink, error) {
	if alias == "" {
		s := sessionID + URL
		alias = fmt.Sprintf("%x", sha1.Sum([]byte(s)))[:8]
	} else {
		if err := s.getStore().SetAlias(ctx, alias); err != nil {
			return nil, err
		}
	}

	tl, err := data.NewTinylink("http://localhost:3000", URL, alias)
	if err != nil {
		return nil, err
	}

	qp := data.QueryParams{SessionID: sessionID, Alias: alias}
	if err := s.getStore().Save(ctx, tl, qp); err != nil {
		return nil, err
	}

	return tl, nil
}

func (s *TinylinkService) Get(ctx context.Context, sessionID, alias string) (*data.Tinylink, error) {
	tl, err := s.getStore().Get(ctx, data.QueryParams{SessionID: sessionID, Alias: alias})
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (s *TinylinkService) Delete(ctx context.Context, sessionID, alias string) error {
	return s.getStore().Delete(ctx, data.QueryParams{SessionID: sessionID, Alias: alias})
}
