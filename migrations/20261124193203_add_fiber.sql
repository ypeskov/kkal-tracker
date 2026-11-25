-- +goose Up
-- +goose StatementBegin
-- Dietary fibre in grams per 100 g, next to fats / carbs / proteins. NULL means unknown: existing rows are not backfilled
ALTER TABLE global_ingredients ADD COLUMN fiber REAL;
ALTER TABLE user_ingredients ADD COLUMN fiber REAL;
ALTER TABLE calorie_entries ADD COLUMN fiber REAL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE calorie_entries DROP COLUMN fiber;
ALTER TABLE user_ingredients DROP COLUMN fiber;
ALTER TABLE global_ingredients DROP COLUMN fiber;
-- +goose StatementEnd
