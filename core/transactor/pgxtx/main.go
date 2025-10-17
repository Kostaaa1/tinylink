package pgxtx

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxtxQuerier struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *pgxtxQuerier {
	return &pgxtxQuerier{db: db}
}

func (p *pgxtxQuerier) WithTx(ctx context.Context, txFunc func(ctx context.Context) error) error {
	return nil
}
