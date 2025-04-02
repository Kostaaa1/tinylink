package sqlitedb

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	Tinylink *TinylinkRepository
	User     *UserRepository
}

func NewRepositoriesFromDB(db *sqlx.DB) *Repositories {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}
	return &Repositories{
		Tinylink: &TinylinkRepository{db: db},
		User:     &UserRepository{db: db},
	}
}

func NewRepositories(dbPath string) *Repositories {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxIdleConns(25)
	db.SetMaxOpenConns(25)

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatal("Failed to enable foreign keys:", err)
	}

	// Enable Write-Ahead Logging ?
	// if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
	// 	log.Fatal("Failed to enable foreign keys:", err)
	// }

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	return &Repositories{
		Tinylink: &TinylinkRepository{db: db},
		User:     &UserRepository{db: db},
	}
}
