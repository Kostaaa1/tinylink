// package main

// import (
// 	"context"
// )

// type TinylinkRepository interface {
// 	List(ctx context.Context, qp QueryParams) ([]*Tinylink, error)
// 	Delete(ctx context.Context, qp QueryParams) error
// 	Save(ctx context.Context, tl *Tinylink, qp QueryParams) error
// 	Get(ctx context.Context, qp QueryParams) (*Tinylink, error)
// 	////////////////
// 	CheckAlias(ctx context.Context, alias string) error
// 	CheckOriginalURL(ctx context.Context, clientID, URL string) error
// 	// Ping(ctx context.Context) error
// }
