CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY,
    email       VARCHAR(255) NOT NULL UNIQUE,
    password    TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);