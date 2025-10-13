package transactor

import "context"

// Tx represents a generic database transaction interface
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// TxBeginner is an interface for beginning transactions
type TxBeginner interface {
	Begin(ctx context.Context) (Tx, error)
}

// Transactor is an interface for types that can work with transactions
type Transactor[T any] interface {
	WithRepositoryTx(tx Tx) T
}

// Provider is a generic provider that works with any database implementation
type Provider[T Transactor[T]] struct {
	repos      T
	txBeginner TxBeginner
}

// NewProvider creates a newProvider
func NewProvider[T Transactor[T]](plainRepos T, txBeginner TxBeginner) *Provider[T] {
	return &Provider[T]{plainRepos, txBeginner}
}

// WithTx executes a function within a transaction
func (p *Provider[T]) WithTx(ctx context.Context, txFunc func(repos T) error) error {
	tx, err := p.txBeginner.Begin(ctx)
	if err != nil {
		return err
	}

	repos := p.repos.WithRepositoryTx(tx)

	if err := txFunc(repos); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// Repos returns the repositories without transaction
func (p *Provider[T]) Repos() T {
	return p.repos
}
