CREATE TABLE IF NOT EXISTS tinylinks (
	id SERIAL PRIMARY KEY,
	alias TEXT NOT NULL,
	url TEXT NOT NULL,
	domain TEXT,
	private BOOLEAN NOT NULL DEFAULT FALSE,
	user_id INTEGER DEFAULT NULL REFERENCES users(id),
	guest_id UUID DEFAULT NULL,
	version INTEGER NOT NULL DEFAULT 1,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT NULL,
	expiration TIMESTAMP DEFAULT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_public_alias ON tinylinks(alias) WHERE private = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_alias_per_user ON tinylinks(alias, user_id) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_alias_per_guest ON tinylinks(alias, guest_id) WHERE guest_id IS NOT NULL;