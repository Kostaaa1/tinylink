package user

import (
	"database/sql"
	"errors"
)

type RepositoryProvider struct {
	db *sql.DB
}

func NewRepositoryProvider(db *sql.DB) *RepositoryProvider {
	return &RepositoryProvider{db: db}
}

func (tp *RepositoryProvider) WithTransaction(txFunc func(adapters Adapters) error) error {
	return runInTx(tp.db, func(tx *sql.Tx) error {
		adapters := Adapters{
			UserDbRepository: &SQLiteUserRepository{db: tx},
		}
		return txFunc(adapters)
	})
}

func (tp *RepositoryProvider) GetAdapters() Adapters {
	return Adapters{
		UserDbRepository: &SQLiteUserRepository{db: tp.db},
	}
}

func runInTx(db *sql.DB, fn func(tx *sql.Tx) error) error {
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
