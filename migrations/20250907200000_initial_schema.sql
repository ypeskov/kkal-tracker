-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT (datetime('now')),
    updated_at DATETIME DEFAULT (datetime('now'))
);

CREATE TABLE calorie_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    food TEXT NOT NULL,
    calories INTEGER NOT NULL,
    weight REAL NOT NULL,
    kcal_per_100g REAL NOT NULL,
    fats REAL,
    carbs REAL,
    proteins REAL,
    meal_datetime DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE calorie_entries;
DROP TABLE users;
-- +goose StatementEnd