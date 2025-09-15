-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN weight REAL DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN weight;
-- +goose StatementEnd