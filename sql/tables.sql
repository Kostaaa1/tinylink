CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL COLLATE NOCASE,
    name TEXT NOT NULL,
    password_hash BLOB,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    version INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS google_users_data (
    user_id INTEGER PRIMARY KEY NOT NULL,
    google_id TEXT NOT NULL,
    email TEXT NOT NULL,
    name TEXT,
    given_name TEXT,
    family_name TEXT,
    picture TEXT,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    is_verified INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tinylinks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    alias TEXT NOT NULL,
    original_url TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    is_private INTEGER NOT NULL DEFAULT 0,
    domain TEXT,
    version INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')), 
    expires_at INTEGER NOT NULL DEFAULT 0, 
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS visit_log (
    tinylink_id INTEGER NOT NULL,
    `timestamp` INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    usage_count INTEGER NOT NULL DEFAULT 0,
    ip_address TEXT,
    user_agent TEXT,
    referrer TEXT,
    geo_city TEXT,
    geo_country TEXT,
    browser TEXT,
    os TEXT,
    FOREIGN KEY (tinylink_id) REFERENCES tinylinks (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email COLLATE NOCASE);

CREATE INDEX IF NOT EXISTS idx_google_users_data_google_id ON google_users_data(google_id);
CREATE INDEX IF NOT EXISTS idx_google_users_data_email ON google_users_data(email);

CREATE INDEX IF NOT EXISTS idx_tinylinks_alias ON tinylinks(alias);
CREATE INDEX IF NOT EXISTS idx_tinylinks_user_id ON tinylinks(user_id);
CREATE INDEX IF NOT EXISTS idx_tinylinks_created_at ON tinylinks(created_at);
CREATE INDEX IF NOT EXISTS idx_tinylinks_is_private ON tinylinks(is_private);

CREATE INDEX IF NOT EXISTS idx_visit_log_timestamp ON visit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_visit_log_geo_country ON visit_log(geo_country);
CREATE INDEX IF NOT EXISTS idx_visit_log_referrer ON visit_log(referrer);
CREATE INDEX IF NOT EXISTS idx_visit_log_browser ON visit_log(browser);
CREATE INDEX IF NOT EXISTS idx_visit_log_os ON visit_log(os);