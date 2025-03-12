package sqlitedb

import (
	"database/sql"
	"log"
	"time"

	"github.com/Kostaaa1/tinylink/internal/store"
	_ "github.com/mattn/go-sqlite3"
)

func NewSqliteStore() *store.Store {
	db, err := sql.Open("sqlite3", "./tinylink.db?_datetime=rfc3339")
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	return &store.Store{
		User: NewSqliteUserRepository(db),
		// Tinylink: NewSQLiteTinylinkRepository(db),
	}
}
