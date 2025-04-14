package sqlitedb

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/config"

	_ "embed"
)

func StartDB(conf config.SQLConfig) (*sql.DB, error) {
	var err error
	var db *sql.DB

	switch {
	case conf.DSN != "":
		db, err = sql.Open("sqlite3", conf.DSN)
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite connection with DSN: %w", err)
		}
	case conf.SQLitePath != "":
		db, err = sql.Open("sqlite3", conf.SQLitePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite connection from path: %w", err)
		}
	default:
		log.Fatal("No DSN or SQLite path provided")
	}

	db.SetMaxIdleConns(conf.MaxIdleConns)
	db.SetMaxOpenConns(conf.MaxOpenConns)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, err
	}

	// Enable Write-Ahead Logging ?
	// if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
	// 	log.Fatal("Failed to enable foreign keys:", err)
	// }
	// _, b, _, _ := runtime.Caller(0)
	// basePath := filepath.Join(filepath.Dir(b), "../../../../sql/tables.sql")
	// tablesSql, err := os.ReadFile(basePath)
	// if err != nil {
	// 	return nil, err
	// }
	// db.Exec(string(tablesSql))

	return db, nil
}
