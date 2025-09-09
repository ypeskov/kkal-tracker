-- +goose Up
-- +goose StatementBegin
-- Add unique constraint to prevent duplicate ingredients per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_ingredients_unique ON user_ingredients(user_id, name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_ingredients_unique;
-- +goose StatementEnd