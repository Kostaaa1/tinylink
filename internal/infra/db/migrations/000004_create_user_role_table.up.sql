DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'roles_enum'
    ) THEN
        CREATE TYPE roles_enum AS ENUM ('admin', 'subscriber', 'user', 'guest');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name roles_enum UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    role_id INT REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);