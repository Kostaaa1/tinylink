package sqlitedb

import (
	"database/sql"
)

type SQLiteTinylinkRepository struct {
	db *sql.DB
}

// func SQLitedisTinylinkRepository(db *sql.DB) store.TinylinkRepository {
// return &SQLiteTinylinkRepository{
// 	db: db,
// }
// }
