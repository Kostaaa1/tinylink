-- PostgreSQL migration script for creating tables
-- OPTIONS
-- Goal is to have NON-UNIQUE aliases for private tinylinks (but public aliases must be UNIQUE) which will be accessed via - /p/:alias route (or just /:alias for all tinylinks but if accessed alias is private and user is logged in then access its private)
-- Option 1: have one table for all tinylinks, the problem is that i always need to check and validate alias before upserting.
-- Option 2: have two separate tables for public and private tinylinks, where i wouldn't really need to check alias since i could use UNIQUE constraint on both tables (all private and public aliases must be unique)

-- CREATE EXTENSION IF NOT EXISTS citext;
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- -- USERS
-- CREATE TABLE IF NOT EXISTS users (
--     id SERIAL PRIMARY KEY,
--     email CITEXT UNIQUE NOT NULL,
--     name TEXT NOT NULL,
--     password_hash BYTEA NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     version INTEGER NOT NULL DEFAULT 0
--     tinylink_count INTEGER DEFAULT 0,
-- );
-- CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
-- CREATE INDEX IF NOT EXISTS idx_users_name ON users(name);
-- CREATE INDEX IF NOT EXISTS idx_users_email ON users(name);
-- CREATE INDEX IF NOT EXISTS idx_users_version ON users(version);

-- -- Google users
-- CREATE TABLE IF NOT EXISTS google_users_data (
--     user_id INTEGER PRIMARY KEY NOT NULL REFERENCES users(id) ON DELETE CASCADE,
--     google_id TEXT NOT NULL,
--     email TEXT NOT NULL,
--     name TEXT,
--     given_name TEXT,
--     family_name TEXT,
--     picture TEXT,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     is_verified BOOLEAN NOT NULL DEFAULT FALSE
-- );
-- CREATE INDEX IF NOT EXISTS idx_google_users_data_google_id ON google_users_data(google_id);

-- TINYLINKS table
-- CREATE TABLE IF NOT EXISTS tinylinks (
-- 	id SERIAL PRIMARY KEY,
-- 	alias TEXT NOT NULL,
-- 	url TEXT NOT NULL,
-- 	domain TEXT,
-- 	private BOOLEAN NOT NULL DEFAULT FALSE,
-- 	user_id INTEGER DEFAULT NULL REFERENCES users(id),
-- 	guest_id UUID DEFAULT NULL,
-- 	version INTEGER NOT NULL DEFAULT 1,
-- 	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
-- 	updated_at TIMESTAMP DEFAULT NULL,
-- 	expiration TIMESTAMP DEFAULT NULL
--     -- CONSTRAINT chk_user_or_guest_id_not_null 
--     -- CHECK (user_id IS NOT NULL OR guest_id IS NOT NULL)
-- );
-- CREATE UNIQUE INDEX IF NOT EXISTS uniq_public_alias ON tinylinks(alias) WHERE private = FALSE;
-- CREATE UNIQUE INDEX IF NOT EXISTS uniq_alias_per_user ON tinylinks(alias, user_id) WHERE user_id IS NOT NULL;
-- CREATE UNIQUE INDEX IF NOT EXISTS uniq_alias_per_guest ON tinylinks(alias, guest_id) WHERE guest_id IS NOT NULL;

-- CREATE INDEX IF NOT EXISTS ifx_tinylinks_private ON tinylinks(private);
-- CREATE INDEX IF NOT EXISTS idx_tinylinks_user_id ON tinylinks(user_id);
-- CREATE INDEX IF NOT EXISTS idx_tinylinks_guest_id ON tinylinks(guest_id);
-- CREATE INDEX IF NOT EXISTS idx_tinylinks_created_at ON tinylinks(created_at);
-- CREATE INDEX IF NOT EXISTS idx_tinylinks_expiration ON tinylinks(expiration);

-- roles and user permissions
-- CREATE TYPE roles_enum AS ENUM ('admin', 'subscriber', 'user', 'guest');

-- CREATE TABLE IF NOT EXISTS roles (
--     id SERIAL PRIMARY KEY,
--     name roles_enum UNIQUE NOT NULL,
--     description TEXT
-- );

-- CREATE TABLE IF NOT EXISTS user_roles (
--     user_id INT REFERENCES users(id) ON DELETE CASCADE,
--     role_id INT REFERENCES roles(id) ON DELETE CASCADE,
--     PRIMARY KEY (user_id, role_id)
-- );

-- CREATE TYPE IF NOT EXISTS resources AS ENUM ('all');

-- CREATE TYPE IF NOT EXISTS permissions AS ENUM ('write', 'read', 'update', 'bulk-insert', 'metrics');

-- CREATE TABLE IF NOT EXISTS permissions (
--     id SERIAL PRIMARY KEY,
--     permission VARCHAR(50) UNIQUE NOT NULL,
--     description TEXT
-- );

-- CREATE TABLE IF NOT EXISTS role_permissions (
--     permission_id INT REFERENCES permissions(id),
--     role_id INT REFERENCES roles(id),
--     resource_type 
-- )

-- visit_log