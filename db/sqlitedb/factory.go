package sqlitedb

import (
	"context"
	"log"
	"time"

	"github.com/Kostaaa1/tinylink/db"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func NewSQLiteStore(dbPath string) *db.SQLiteStore {
	sqldb, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	sqldb.SetMaxIdleConns(25)
	sqldb.SetMaxOpenConns(25)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := sqldb.PingContext(ctx); err != nil {
		panic(err)
	}

	return &db.SQLiteStore{
		Tinylink: &SQLiteTinylinkStore{db: sqldb},
		User:     &SQLiteUserStore{db: sqldb},
	}
}
