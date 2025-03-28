CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL COLLATE NOCASE,
    password_hash BLOB NOT NULL,
    activated INTEGER NOT NULL DEFAULT 1,
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS tinylinks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    alias TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    is_private INTEGER NOT NULL DEFAULT 0,
    usage_count INTEGER NOT NULL DEFAULT 0,
    domain TEXT,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')), 
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_users_email ON users(email);
CREATE UNIQUE INDEX idx_users_email_unique ON users(email);
CREATE INDEX idx_users_id ON users(id);

CREATE INDEX idx_tinylinks_user_id ON tinylinks(user_id);
CREATE INDEX idx_tinylinks_user_alias ON tinylinks(user_id, alias);
CREATE UNIQUE INDEX idx_tinylinks_alias ON tinylinks(alias);