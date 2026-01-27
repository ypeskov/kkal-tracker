-- +goose Up
-- +goose StatementBegin
-- Add gender field to users table
ALTER TABLE users ADD COLUMN gender TEXT CHECK(gender IN ('male', 'female'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove gender field from users table
-- Note: SQLite doesn't support DROP COLUMN directly in older versions
-- For production, consider creating a new table without the column
ALTER TABLE users DROP COLUMN gender;
-- +goose StatementEnd
