package sqlitedb

import (
	"log"
	"time"

	"github.com/Kostaaa1/tinylink/internal/store"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Provider struct {
	db       *sqlx.DB
	tinylink store.TinylinkStore
	user     store.UserStore
}

func NewProvider(dbPath string) *Provider {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	provider := &Provider{
		db: db,
	}

	// define services that will use sqlite
	provider.tinylink = &SQLiteTinylinkStore{db: db}
	provider.user = &SQLiteUserStore{db: db}

	return provider
}

func (p *Provider) Tinylink() store.TinylinkStore {
	return p.tinylink
}

func (p *Provider) User() store.UserStore {
	return p.user
}

func (p *Provider) Close() error {
	return p.db.Close()
}
