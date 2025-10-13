package db

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenPostgresPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	pgConf, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgConf)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	if err := runMigrations(connString); err != nil {
		return nil, err
	}

	return pool, nil
}

func runMigrations(dsn string) error {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Join(filepath.Dir(b), "migrations")

	m, err := migrate.New(
		fmt.Sprintf("file://%s", basepath),
		dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to init migrate: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			fmt.Printf("source close error: %v", srcErr)
		}
		if dbErr != nil {
			fmt.Printf("database close error: %v", dbErr)
		}
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
