-- +goose Up
-- +goose StatementBegin
-- Per-user UI preferences: one row per user and key, the value is JSON encoded by the server
CREATE TABLE user_preferences (
    user_id INTEGER NOT NULL,
    pref_key TEXT NOT NULL,
    value TEXT NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, pref_key),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_preferences;
-- +goose StatementEnd
