package tinylink

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type RepositoryProvider struct {
	db     *sqlx.DB
	client *redis.Client
}

func NewRepositoryProvider(db *sqlx.DB, client *redis.Client) *RepositoryProvider {
	fmt.Println(db)
	return &RepositoryProvider{
		db:     db,
		client: client,
	}
}

func (p *RepositoryProvider) GetAdapters() Adapters {
	return Adapters{
		DBAdapters: DBAdapters{
			TinylinkDBRepository: &TinylinkSQLRepository{db: p.db},
		},
		RedisAdapters: RedisAdapters{
			TinylinkRedisRepository: &TinylinkRedisRepository{client: p.client},
		},
	}
}

func (p *RepositoryProvider) WithTransaction(txFunc func(dbAdapters DBAdapters) error) error {
	return runInTx(p.db, func(tx *sql.Tx) error {
		adapters := DBAdapters{
			TinylinkDBRepository: &TinylinkSQLRepository{db: tx},
		}
		return txFunc(adapters)
	})
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
