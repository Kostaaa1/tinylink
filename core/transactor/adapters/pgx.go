package adapters

import (
	"context"

	"github.com/Kostaaa1/tinylink/core/transactor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxQuerier interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type PgxProvider[T transactor.Transactor[T]] struct {
	repos      T
	txBeginner transactor.TxBeginner
}

func NewPgxProvider[T transactor.Transactor[T]](plainRepos T, txBeginner transactor.TxBeginner) *PgxProvider[T] {
	return &PgxProvider[T]{plainRepos, txBeginner}
}

type pgxPoolAdapter struct {
	db *pgxpool.Pool
}

func WithPgxPool(pool *pgxpool.Pool) transactor.TxBeginner {
	return &pgxPoolAdapter{db: pool}
}

func (a *pgxPoolAdapter) Begin(ctx context.Context) (transactor.Tx, error) {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (p *PgxProvider[T]) WithTx(ctx context.Context, txFunc func(repos T) error) error {
	tx, err := p.txBeginner.Begin(ctx)
	if err != nil {
		return err
	}

	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		panic("interface is not implemented")
	}

	repos := p.Repos().WithRepositoryTx(pgxTx)

	if err := txFunc(repos); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (p *PgxProvider[T]) Repos() T {
	return p.repos
}
