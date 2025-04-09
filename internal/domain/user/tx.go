package user

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type RepositoryProvider struct {
	db *sqlx.DB
}

func NewRepositoryProvider(db *sqlx.DB) *RepositoryProvider {
	return &RepositoryProvider{db: db}
}

func (tp *RepositoryProvider) WithTransaction(txFunc func(adapters Adapters) error) error {
	return runInTx(tp.db, func(tx *sql.Tx) error {
		adapters := Adapters{
			UserRepository: &SQLiteUserRepository{db: tx},
		}
		return txFunc(adapters)
	})
}

func (tp *RepositoryProvider) GetDbAdapters() Adapters {
	return Adapters{
		UserRepository: &SQLiteUserRepository{db: tp.db},
	}
}

func runInTx(db *sqlx.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = fn(tx)
	if err == nil {
		return tx.Commit()
	}

	rollbackErr := tx.Rollback()
	if rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}
