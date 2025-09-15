-- +goose Up
-- +goose StatementBegin
-- Add profile fields to users table
ALTER TABLE users ADD COLUMN first_name TEXT;
ALTER TABLE users ADD COLUMN last_name TEXT;
ALTER TABLE users ADD COLUMN age INTEGER;
ALTER TABLE users ADD COLUMN height REAL; -- Height in cm
ALTER TABLE users ADD COLUMN language TEXT DEFAULT 'en_US';

-- Create weight history table for tracking weight over time
CREATE TABLE weight_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    weight REAL NOT NULL, -- Weight in kg
    recorded_at DATETIME DEFAULT (datetime('now')),
    created_at DATETIME DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Create index for efficient weight history queries
CREATE INDEX idx_weight_history_user_recorded ON weight_history(user_id, recorded_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop weight history table
DROP INDEX IF EXISTS idx_weight_history_user_recorded;
DROP TABLE IF EXISTS weight_history;

-- Remove profile fields from users table
-- Note: SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
CREATE TABLE users_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT (datetime('now')),
    updated_at DATETIME DEFAULT (datetime('now'))
);

INSERT INTO users_backup (id, email, password_hash, created_at, updated_at)
SELECT id, email, password_hash, created_at, updated_at FROM users;

DROP TABLE users;

ALTER TABLE users_backup RENAME TO users;
-- +goose StatementEnd