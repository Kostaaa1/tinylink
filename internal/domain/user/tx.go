package user

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type RepositoryProvider struct {
	db *sqlx.DB
}

func NewRepositoryProvider(db *sqlx.DB) *RepositoryProvider {
	return &RepositoryProvider{db: db}
}

func (tp *RepositoryProvider) GetAdapters() Adapters {
	return Adapters{
		UserRepository: &SQLiteUserRepository{db: tp.db},
	}
}

func (tp *RepositoryProvider) WithTransaction(txFunc func(adapters Adapters) error) error {
	return runInTx(tp.db, func(tx *sql.Tx) error {
		adapters := Adapters{
			UserRepository: &SQLiteUserRepository{db: tx},
		}
		return txFunc(adapters)
	})
}

func runInTx(db *sqlx.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			fmt.Println("commiting")
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// func runInTx(db *sqlx.DB, fn func(tx *sql.Tx) error) error {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return err
// 	}
// 	defer func() {
// 		tx.Rollback()
// 	}()
// 	err = fn(tx)
// 	if err != nil {
// 		return err
// 	}
// 	return tx.Commit()
// }

// func runInTx(db *sqlx.DB, fn func(tx *sql.Tx) error) error {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return err
// 	}

// 	err = fn(tx)
// 	if err == nil {
// 		return tx.Commit()
// 	}

// 	fmt.Println("running rollback")
// 	rollbackErr := tx.Rollback()
// 	if rollbackErr != nil {
// 		fmt.Println("rollback failed?")
// 		return errors.Join(err, rollbackErr)
// 	}

// 	return err
// }
