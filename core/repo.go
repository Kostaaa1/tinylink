package core

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryProvider[T any] interface {
	Repos() T
	WithTransaction(ctx context.Context, txFunc func(repos T) error) error
}

type PgxPoolExecer interface{}

type PgxTransactionProvider[T any] interface {
	WithTx(tx *sql.Tx) T
}

// main provider struct that holds db/tx instance
type PostgresRepositoryProvider[T any] struct {
	db    *pgxpool.Pool
	repos T
}

func PostgresSQLRepositoryProvider[T any](db *pgxpool.Pool, repos T) RepositoryProvider[T] {
	return &PostgresRepositoryProvider[T]{
		db:    db,
		repos: repos,
	}
}

func (provider *PostgresRepositoryProvider[T]) Repos() T {
	return provider.repos
}

func (provider *PostgresRepositoryProvider[T]) WithTransaction(ctx context.Context, txFunc func(adapters T) error) error {
	tx, err := provider.db.Begin(ctx)
	if err != nil {
		return err
	}

	repos := provider.Repos()

	err = txFunc(repos)
	if err == nil {
		return tx.Commit(ctx)
	}

	rollbackErr := tx.Rollback(ctx)
	if rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}

type SqlExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type SQLTransactionProvider[T any] interface {
	WithTx(tx *sql.Tx) T
}

// main provider struct that holds db/tx instance
type SQLRepositoryProvider[T SQLTransactionProvider[T]] struct {
	db    *sql.DB
	repos T
}

func NewSQLRepositoryProvider[T SQLTransactionProvider[T]](db *sql.DB, repos T) RepositoryProvider[T] {
	return &SQLRepositoryProvider[T]{
		db:    db,
		repos: repos,
	}
}

func (provider *SQLRepositoryProvider[T]) Repos() T {
	return provider.repos
}

func (provider *SQLRepositoryProvider[T]) WithTransaction(ctx context.Context, txFunc func(adapters T) error) error {
	tx, err := provider.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	adapters := provider.Repos().WithTx(tx)

	err = txFunc(adapters)
	if err == nil {
		return tx.Commit()
	}

	rollbackErr := tx.Rollback()
	if rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}
