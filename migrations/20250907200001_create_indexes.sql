-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_calorie_entries_user_id ON calorie_entries(user_id);
CREATE INDEX idx_calorie_entries_meal_datetime ON calorie_entries(meal_datetime);
CREATE INDEX idx_calorie_entries_user_date ON calorie_entries(user_id, date(meal_datetime));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_calorie_entries_user_id;
DROP INDEX idx_calorie_entries_meal_datetime;
DROP INDEX idx_calorie_entries_user_date;
-- +goose StatementEnd