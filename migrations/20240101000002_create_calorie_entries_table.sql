-- +goose Up
-- +goose StatementBegin
CREATE TABLE calorie_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    food TEXT NOT NULL,
    calories INTEGER NOT NULL,
    date DATE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE calorie_entries;
-- +goose StatementEnd