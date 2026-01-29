-- +goose Up
-- +goose StatementBegin
-- Add activity_level field to users table with default value 'sedentary'
ALTER TABLE users ADD COLUMN activity_level TEXT DEFAULT 'sedentary' CHECK(activity_level IN ('sedentary', 'lightly_active', 'moderate', 'very_active', 'extra_active'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove activity_level field from users table
ALTER TABLE users DROP COLUMN activity_level;
-- +goose StatementEnd
