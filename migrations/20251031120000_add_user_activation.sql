-- +goose Up
-- +goose StatementBegin

-- Add is_active column to users table
ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT 0 NOT NULL;

-- Create activation_tokens table
CREATE TABLE activation_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT (datetime('now')) NOT NULL,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_activation_tokens_token ON activation_tokens(token);
CREATE INDEX idx_activation_tokens_user_id ON activation_tokens(user_id);
CREATE INDEX idx_activation_tokens_expires_at ON activation_tokens(expires_at);

-- Set existing users as active (backward compatibility)
UPDATE users SET is_active = 1;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS idx_activation_tokens_expires_at;
DROP INDEX IF EXISTS idx_activation_tokens_user_id;
DROP INDEX IF EXISTS idx_activation_tokens_token;

-- Drop activation_tokens table
DROP TABLE IF EXISTS activation_tokens;

-- Remove is_active column from users
ALTER TABLE users DROP COLUMN is_active;

-- +goose StatementEnd
