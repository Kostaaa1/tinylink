package sqlitedb

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/stretchr/testify/require"

	_ "embed"
)

func StartTest(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "tinylink_test.db")

	conf := config.SQLConfig{
		SQLitePath:   dbPath,
		MaxOpenConns: 25,
		MaxIdleConns: 25,
	}

	db, err := StartDB(conf)
	require.NoError(t, err)

	file, err := os.ReadFile("../../../sql/tables.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(file))
	require.NoError(t, err)

	return db
}

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

	_, filaPath, _, _ := runtime.Caller(0)
	basePath := filepath.Join(filepath.Dir(filaPath), "../../../../sql/tables.sql")
	tablesSql, err := os.ReadFile(basePath)
	if err != nil {
		return nil, err
	}
	db.Exec(string(tablesSql))

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	return db, nil
}

// Enable Write-Ahead Logging ?
// if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
// 	log.Fatal("Failed to enable foreign keys:", err)
// }
