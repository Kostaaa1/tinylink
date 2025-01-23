// package main

// import (
// 	"context"

// 	"github.com/Kostaaa1/tinylink/internal/application/interfaces"
// )

// type TinylinkService struct {
// 	tinylinkRepo TinylinkRepository
// }

// func NewTinylinkService(tinylinkRepo TinylinkRepository) interfaces.TinylinkService {
// 	return TinylinkService{tinylinkRepo: tinylinkRepo}
// }

// func (s *TinylinkService) List(ctx context.Context, sessionID string) ([]*Tinylink, error) {
// 	links, err := s.tinylinkRepo.List(ctx, QueryParams{SessionID: sessionID})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return links, nil
// }

// func (s *TinylinkService) Create(ctx context.Context, sessionID, URL, alias string) (*Tinylink, error) {
// 	if err := s.tinylinkRepo.CheckOriginalURL(ctx, sessionID, URL); err != nil {
// 		return nil, err
// 	}

// 	if alias == "" {
// 		alias = createHashAlias(sessionID, URL, 8)
// 	} else {
// 		if err := s.tinylinkRepo.CheckAlias(ctx, alias); err != nil {
// 			return nil, err
// 		}
// 	}

// 	tl, err := NewTinylink("http://localhost:3000", URL, alias)
// 	if err != nil {
// 		return nil, err
// 	}

// 	qp := QueryParams{SessionID: sessionID, Alias: alias}
// 	if err := s.tinylinkRepo.Save(ctx, tl, qp); err != nil {
// 		return nil, err
// 	}

// 	return tl, nil
// }

// func (s *TinylinkService) Get(ctx context.Context, sessionID, alias string) (*Tinylink, error) {
// 	tl, err := s.tinylinkRepo.Get(ctx, QueryParams{SessionID: sessionID, Alias: alias})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return tl, nil
// }

// func (s *TinylinkService) Delete(ctx context.Context, sessionID, alias string) error {
// 	return s.tinylinkRepo.Delete(ctx, QueryParams{SessionID: sessionID, Alias: alias})
// }
