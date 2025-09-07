-- Sample global ingredients with translations
-- Fruits
INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (52, 0.2, 14, 0.3);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (1, 'en_US', 'apple');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (1, 'uk_UA', 'яблуко');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (1, 'ru_UA', 'яблоко');

INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (89, 0.3, 23, 1.1);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (2, 'en_US', 'banana');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (2, 'uk_UA', 'банан');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (2, 'ru_UA', 'банан');

INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (32, 0.1, 8, 0.6);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (3, 'en_US', 'orange');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (3, 'uk_UA', 'апельсин');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (3, 'ru_UA', 'апельсин');

-- Vegetables
INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (77, 0.1, 18, 2);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (4, 'en_US', 'potato');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (4, 'uk_UA', 'картопля');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (4, 'ru_UA', 'картофель');

INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (25, 0.1, 6, 1.4);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (5, 'en_US', 'tomato');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (5, 'uk_UA', 'помідор');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (5, 'ru_UA', 'помидор');

INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (41, 0.2, 10, 1.9);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (6, 'en_US', 'carrot');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (6, 'uk_UA', 'морква');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (6, 'ru_UA', 'морковь');

-- Grains
INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (365, 6.6, 66, 13);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (7, 'en_US', 'rice');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (7, 'uk_UA', 'рис');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (7, 'ru_UA', 'рис');

INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (264, 2.2, 53, 9);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (8, 'en_US', 'bread');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (8, 'uk_UA', 'хліб');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (8, 'ru_UA', 'хлеб');

-- Proteins
INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (239, 15, 0, 27);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (9, 'en_US', 'chicken breast');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (9, 'uk_UA', 'куряча грудка');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (9, 'ru_UA', 'куриная грудка');

INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (155, 11, 1.1, 13);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (10, 'en_US', 'egg');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (10, 'uk_UA', 'яйце');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (10, 'ru_UA', 'яйцо');

-- Dairy
INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (42, 1, 5, 3.4);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (11, 'en_US', 'milk');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (11, 'uk_UA', 'молоко');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (11, 'ru_UA', 'молоко');

INSERT OR REPLACE INTO global_ingredients (kcal_per_100g, fats, carbs, proteins) VALUES (113, 9, 4.3, 3.2);
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (12, 'en_US', 'cheese');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (12, 'uk_UA', 'сир');
INSERT OR REPLACE INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (12, 'ru_UA', 'сыр');