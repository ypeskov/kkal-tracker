-- +goose Up
-- +goose StatementBegin
-- Global ingredients (master data, admin managed)
CREATE TABLE global_ingredients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    kcal_per_100g REAL NOT NULL,
    fats REAL,
    carbs REAL,
    proteins REAL,
    created_at DATETIME DEFAULT (datetime('now')),
    updated_at DATETIME DEFAULT (datetime('now'))
);

CREATE TABLE global_ingredient_names (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ingredient_id INTEGER NOT NULL,
    language_code TEXT NOT NULL,
    name TEXT NOT NULL,
    FOREIGN KEY (ingredient_id) REFERENCES global_ingredients (id) ON DELETE CASCADE,
    UNIQUE (ingredient_id, language_code)
);

-- User-specific ingredients (copied from global on registration)
CREATE TABLE user_ingredients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    kcal_per_100g REAL NOT NULL,
    fats REAL,
    carbs REAL,
    proteins REAL,
    global_ingredient_id INTEGER, -- Reference to original global ingredient (nullable for custom user ingredients)
    created_at DATETIME DEFAULT (datetime('now')),
    updated_at DATETIME DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (global_ingredient_id) REFERENCES global_ingredients (id) ON DELETE SET NULL
);

-- Indexes for performance
CREATE INDEX idx_global_ingredient_names_search ON global_ingredient_names (language_code, name);
CREATE INDEX idx_global_ingredient_names_ingredient ON global_ingredient_names (ingredient_id);
CREATE INDEX idx_user_ingredients_user ON user_ingredients (user_id);
CREATE INDEX idx_user_ingredients_name ON user_ingredients (user_id, name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_user_ingredients_name;
DROP INDEX idx_user_ingredients_user;
DROP INDEX idx_global_ingredient_names_ingredient;
DROP INDEX idx_global_ingredient_names_search;
DROP TABLE user_ingredients;
DROP TABLE global_ingredient_names;
DROP TABLE global_ingredients;
-- +goose StatementEnd
