-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_calorie_entries_user_id ON calorie_entries(user_id);
CREATE INDEX idx_calorie_entries_date ON calorie_entries(date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_calorie_entries_user_id;
DROP INDEX idx_calorie_entries_date;
-- +goose StatementEnd