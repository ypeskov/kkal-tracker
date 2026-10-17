-- +goose Up
-- +goose StatementBegin
-- Remove the rows whose parent row is missing, left behind while foreign keys were not enforced. Every condition is
-- "the parent is missing", so this is safe on any copy of the database and changes nothing on a clean one.
-- No schema change.
DELETE FROM calorie_entries WHERE NOT EXISTS (SELECT 1 FROM users WHERE users.id = calorie_entries.user_id);
DELETE FROM weight_history WHERE NOT EXISTS (SELECT 1 FROM users WHERE users.id = weight_history.user_id);
DELETE FROM activation_tokens WHERE NOT EXISTS (SELECT 1 FROM users WHERE users.id = activation_tokens.user_id);
DELETE FROM api_keys WHERE NOT EXISTS (SELECT 1 FROM users WHERE users.id = api_keys.user_id);
DELETE FROM user_preferences WHERE NOT EXISTS (SELECT 1 FROM users WHERE users.id = user_preferences.user_id);
DELETE FROM user_ingredients WHERE NOT EXISTS (SELECT 1 FROM users WHERE users.id = user_ingredients.user_id);

DELETE FROM global_ingredient_names
WHERE NOT EXISTS (SELECT 1 FROM global_ingredients WHERE global_ingredients.id = global_ingredient_names.ingredient_id);

-- Mirrors ON DELETE SET NULL: the user's ingredient stays, only the link to the missing global ingredient goes
UPDATE user_ingredients SET global_ingredient_id = NULL
WHERE global_ingredient_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM global_ingredients WHERE global_ingredients.id = user_ingredients.global_ingredient_id);
-- +goose StatementEnd

-- +goose Down
-- The deleted rows cannot be restored, so there is nothing to roll back
