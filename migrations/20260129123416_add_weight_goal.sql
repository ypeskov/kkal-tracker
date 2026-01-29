-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN target_weight REAL;
ALTER TABLE users ADD COLUMN target_date DATE;
ALTER TABLE users ADD COLUMN goal_set_at DATETIME;
ALTER TABLE users ADD COLUMN initial_weight_at_goal REAL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN target_weight;
ALTER TABLE users DROP COLUMN target_date;
ALTER TABLE users DROP COLUMN goal_set_at;
ALTER TABLE users DROP COLUMN initial_weight_at_goal;
-- +goose StatementEnd
