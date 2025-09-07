-- +goose Up
-- +goose StatementBegin
ALTER TABLE calorie_entries ADD COLUMN weight REAL;
ALTER TABLE calorie_entries ADD COLUMN kcal_per_100g REAL;

-- Update existing records with default values based on current calories
-- Assuming 100g portions for existing entries
UPDATE calorie_entries SET weight = 100.0, kcal_per_100g = calories WHERE weight IS NULL;

-- Make the new columns NOT NULL after setting default values
-- Note: SQLite doesn't support ALTER COLUMN, so we recreate the table
CREATE TABLE calorie_entries_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    food TEXT NOT NULL,
    calories INTEGER NOT NULL,
    weight REAL NOT NULL,
    kcal_per_100g REAL NOT NULL,
    date DATE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

INSERT INTO calorie_entries_new (id, user_id, food, calories, weight, kcal_per_100g, date, created_at)
SELECT id, user_id, food, calories, weight, kcal_per_100g, date, created_at FROM calorie_entries;

DROP TABLE calorie_entries;
ALTER TABLE calorie_entries_new RENAME TO calorie_entries;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE calorie_entries_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    food TEXT NOT NULL,
    calories INTEGER NOT NULL,
    date DATE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

INSERT INTO calorie_entries_old (id, user_id, food, calories, date, created_at)
SELECT id, user_id, food, calories, date, created_at FROM calorie_entries;

DROP TABLE calorie_entries;
ALTER TABLE calorie_entries_old RENAME TO calorie_entries;
-- +goose StatementEnd
