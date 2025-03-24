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

	if _, err := sqldb.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatal("Failed to enable foreign keys:", err)
	}

	// Enable Write-Ahead Logging ?
	// if _, err := sqldb.Exec("PRAGMA journal_mode = WAL;"); err != nil {
	// 	log.Fatal("Failed to enable foreign keys:", err)
	// }

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
