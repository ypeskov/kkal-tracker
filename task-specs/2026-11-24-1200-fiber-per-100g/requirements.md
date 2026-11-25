# Dietary fibre per 100 g on ingredients and diary entries

## Overview
The tracker stores dietary fibre (grams per 100 g, field `fiber`) next to kcal / fats / carbs / proteins on
ingredients and on diary entries, so that the day view and the AI skill can show fibre totals for a meal and a day.
Existing data stays unknown (NULL); the AI skill will backfill the ingredients through the API after the release
(the backfill itself is not part of this task).

Naming: the field, column, JSON key and CLI option are `fiber`; prose says "fibre".

## Current State (from the code)

### Data
- `user_ingredients`, `global_ingredients` (`migrations/20250907200002_add_ingredients_table.sql`) and
  `calorie_entries` (`migrations/20250907200000_initial_schema.sql`) have nullable `fats`, `carbs`, `proteins REAL`.
- Models: `internal/models/ingredient.go` (`GlobalIngredient`, `UserIngredient`), `internal/models/calorie_entry.go`
  (`CalorieEntry`, `NewMealEntry`), all `*float64` with `json:"...,omitempty"`.
- All SQL is in `internal/repositories/queries.go`. Queries with the macro columns: `QueryCopyGlobalIngredients`,
  `QueryCopyGlobalIngredientsToUser`, `QueryInsertCalorieEntry`, `QueryGetCalorieEntryByID`,
  `QueryGetCalorieEntriesByUserID`, `QueryGetCalorieEntriesByDateRange`, `QueryUpdateCalorieEntry`,
  `QueryGetAllUserIngredients`, `QueryGetUserIngredientsWithUsage`, `QueryGetUserIngredientByName`,
  `QueryGetUserIngredientByID`, `QueryUpdateUserIngredient`, `QueryUpdateUserIngredientByName`,
  `QueryInsertUserIngredient`, `QueryInsertGlobalIngredient`, `QueryGetGlobalIngredientByID` (each in a SQLite and an
  unwired PostgreSQL variant).
- Repositories: `internal/repositories/ingredient.go`, `internal/repositories/calorie_entry.go` (incl. `CreateMeal`,
  which builds the result from the inserted values), signatures in `internal/repositories/interfaces.go`
  (`fats, carbs, proteins *float64` parameter lists).
- Limits: `internal/food/food.go` — `ValidNutrientPer100g` (0–100), `MaxNutrientPer100g`.

### External API `/api/v1` (X-API-Key) — `internal/handlers/apidata/{handler.go,dto.go}`
- DTOs with macros: `IngredientInfo` (and `IngredientItem`), `FoodItemRequest`, `NewIngredientRequest`,
  `UpdateFoodRequest`, `CreateIngredientRequest`, `UpdateIngredientRequest`, `FoodEntry` (in `handler.go`);
  mappers `mapIngredient`, `mapFoodEntry`.
- `internal/services/calorie/service.go`: `CreateMeal` copies the macros from the ingredient into the entry and applies
  per-item overrides (`overrideNutrients`), rejects top-level macros on a `new_ingredient` item (`firstSet`), validates
  `new_ingredient` macros 0–100; `UpdateMealEntry` copies name/kcal/macros on `ingredient_id`, applies overrides on
  top, `nil` keeps the value (so a macro cannot be reset to NULL). DTOs in `internal/services/calorie/dto.go`.
- `internal/services/ingredient/service.go`: `AddIngredient` / `PatchIngredient` (`validateValues`, 0–100; `nil`
  keeps the value on patch); DTOs in `internal/services/ingredient/dto.go`.
- The server totals only calories (`total_calories`, `day_total_calories`); macro grams are summed by the client.

### Internal web API (JWT)
- `/api/calories` — `internal/handlers/calories/{handler.go,dto.go}` → `calorie.Service.CreateEntry` / `UpdateEntry`.
  The entry takes its macros **from the request body** (the form), not from the ingredient on the server. `PUT`
  replaces all values. No range validation of macros. `CreateEntry` creates the ingredient (with the sent macros) when
  no ingredient has exactly this name.
- `/api/ingredients` — `internal/handlers/ingredients/handler.go` (`CreateRequest`, `UpdateRequest`:
  `validate:"omitempty,min=0"`, no upper limit) → `ingredient.Service.CreateIngredient` / `UpdateIngredient`
  (`validateCreateRequest` / `validateUpdateRequest`: `>= 0`). `PUT` replaces all values: a nutrient that is not sent
  is cleared. Responses are the models themselves.

### Other backend places that read macros
- Excel export: `internal/services/export/excel.go` `writeFoodSheet` — columns G/H/I are fats/carbs/proteins per
  100 g (empty cell for NULL); headers from `internal/i18n/locales/{en_US,uk_UA,ru_UA,bg_BG}.json`
  (`export.columns.*`).
- AI insights: `internal/handlers/ai/handler.go` `aggregateNutritionData` sums macro grams per day (NULL as 0) into
  `NutritionDataPoint` (`internal/services/ai/types.go`), rendered by `internal/services/ai/prompts/user.txt`.
- Reports (`/api/reports/data`): no macros at all — nothing to change.
- Seed: `cmd/seed/sql/01_ingredients.sql` (global ingredients) and `cmd/seed/sql/02_user_data.sql` insert with
  explicit column lists, so they keep working without a change.

### Frontend (`web/src`)
- Types: `api/ingredients.ts` (`Ingredient`, `CreateIngredientData`, `UpdateIngredientData`), `api/calories.ts`
  (`CalorieEntry`).
- Ingredient forms: `components/AddIngredientModal.tsx`, `components/EditIngredientModal.tsx` (fats, carbs, proteins
  in one grid; an empty field is sent as `undefined`, `0` is sent as `0`); list: `components/IngredientsTable.tsx`
  (`?? '-'` for unknown); autocomplete line: `components/FoodAutocomplete.tsx`.
- Diary forms: `components/AddFoodEntryForm.tsx` (macros prefilled from the selected ingredient and sent with the
  entry), `components/EditEntryModal.tsx` (sends the whole entry; `PUT` replaces all values). Both drop a macro that is
  `0` (`parseFloat(x) > 0`, `ingredient.fats ? … : ''`), so a web entry cannot hold a macro of 0.
- Day view: `components/CalorieEntriesList.tsx` groups entries **by day only**; there is no per-meal grouping and no
  per-meal totals. Per day it shows calories and fats/carbs/proteins grams (collapsed header and the expanded footer),
  each hidden when its total is 0. `components/EntryListItem.tsx` shows the grams of each macro of an entry.
  `components/FilterSection.tsx` shows the totals of the whole filtered range (computed in
  `pages/DashboardPage.tsx`). The UI order is already К-Ж-У-Б.
- **There is no "without macros" / "understated" hint in the UI today**: an unknown macro silently counts as 0. The
  hint exists only in the skill client (`without_macros`).
- Locales: `web/src/i18n/locales/{en_US,uk_UA,ru_UA,bg_BG}.json` — `dashboard.{fats,carbs,proteins}` (with
  "(г/100г)"), `dashboard.{fats,carbs,proteins}Consumed`, `foodList.{fats,carbs,proteins}` (no unit), `common.g`.

### Skill template (`skills/kkal-tracker-food-log/`)
- `scripts/kkal_client.py`: `NUTRIENTS = ("fats", "carbs", "proteins")`; `ingredients` prints
  `id name kcal_per_100g fats carbs proteins times_used last_used`; `add` forwards the JSON as is; `edit`,
  `add-ingredient`, `edit-ingredient` have `--fats --carbs --proteins`; `totals` sums whole grams (halves up, NULL as
  0) for `meal` and `day_total` and lists `without_macros` = entries where **all three** macros are NULL
  (`{id, food, in_meal}`). `SKILL_VERSION = "2026-10-16"`.
- `SKILL.md`: reply table columns Ккал, Ж, У, Б; override rules; limits.
- Tests of the client run in Go: `internal/server/skill_client_test.go` (asserts the exact key lists of the `totals`
  output), `internal/server/skill_api_test.go`, `internal/skill/skill_test.go` (frontmatter, title and key anchors of
  the template must stay intact).

## Decisions (existing macro behaviour followed)
1. **Storage**: `fiber REAL` nullable, no default, no backfill, on `user_ingredients`, `calorie_entries` and also
   `global_ingredients` (it carries the macros and is copied to a new user's list; both copy queries copy `fiber`).
   Seed SQL stays unchanged, so global values are NULL.
2. **JSON**: key `fiber`, omitted when NULL (`omitempty`), exactly like `fats` / `carbs` / `proteins`, in both APIs.
3. **`/api/v1` semantics** mirror the macros one to one: 0–100 via `food.ValidNutrientPer100g`; not sent or `null`
   = unknown on create and "keep" on edit; cannot be reset to NULL through `PUT /food/:id` or `PUT /ingredients/:id`;
   override on an `ingredient_id` item for this entry only; top-level `fiber` on a `new_ingredient` item is rejected;
   `ingredient_id` in `PUT /food/:id` copies the ingredient's `fiber` (including NULL), an explicit `fiber` goes on
   top. Error texts follow the existing pattern with the field name `fiber`. Validation order: fats, carbs, proteins,
   fiber.
4. **Calories** are not affected by fibre. The server still totals only calories; fibre grams are summed by the
   clients (skill `totals`, web UI), like the macros.
5. **Web API replace-all**: `PUT /api/ingredients/:id` and `PUT /api/calories/:id` replace all values, so a request
   without `fiber` clears it — the same as for the macros today (documented in `docs/api.md`). The "cannot be reset to
   unknown" rule is the `/api/v1` rule. Consequence: the web entry-edit form **must** carry `fiber`, otherwise every
   web edit of an entry would wipe it.
6. **Web diary forms get a fibre field** (`AddFoodEntryForm`, `EditEntryModal`), after proteins, although the task
   text names only the ingredient form: a web entry takes its nutrients from the form, so without the field an entry
   logged on the website could not copy the ingredient's fibre (and see decision 5).
7. **No per-meal totals in the web day view** (entries are grouped by day only), so fibre is added to the day total,
   to the per-entry line and to the filtered-range total; no per-meal block is introduced.
8. **Order**: fibre always comes after proteins: forms, table column, day totals, export column, `ingredients` output
   of the client (`… proteins fiber times_used last_used`), `totals` rows, reply table (Ккал, Ж, У, Б, Клетчатка).
9. **Excel export**: the food sheet gets a tenth column J with the entry's fibre per 100 g (empty when NULL), header
   `export.columns.fiber` in the 4 backend locales, in the style of the existing headers ("Клетчатка (г)").
10. **Skill `totals`**: `fiber` is added to the `meal` and `day_total` rows as whole grams, halves up, NULL as 0, like
    the macros. A new top-level key `without_fiber` (same shape as `without_macros`: `{id, food, in_meal}`) lists the
    entries whose `fiber` is NULL. `without_macros` keeps its meaning (all of fats, carbs, proteins NULL): fibre does
    not take part in it.
11. **Reports**: no change (no macros there). **Autocomplete suggestion line** (`FoodAutocomplete.tsx`): no fibre
    added (not requested, the line is already long).
12. **PostgreSQL query variants** in `queries.go` are updated together with the SQLite ones, as the file is kept today.

## Decisions confirmed by the user (2026-11-24)
These had no precedent in the code or were ambiguous in the task text; the user answered them.

A. **UI hint**. The UI has no "understated" hint for macros to copy, so the rule is defined here, per group of entries
   (a day, or the filtered range):
   - no entry has a known fibre value → fibre is not shown in the total (nothing is known; not a zero);
   - at least one entry has a known value → the total is shown, also when it is 0.0 (known zero, e.g. a meat-only day);
   - additionally, when at least one entry of the group has unknown fibre → the total carries a visible hint that it
     is understated, with the number of entries without data (ru: "занижено, записей без данных: N").
   The macro totals keep their current behaviour.
B. **0 is a value for fibre in the web forms**: `0` typed or prefilled must be stored as 0 and shown as 0 (the
   existing macro fields of the diary forms turn 0 into unknown; copying that would mark every meat/fish/dairy entry
   as "understated"). Only fibre keeps 0: fats / carbs / proteins in these forms behave as today (out of scope).
C. **Skill reply**: when `without_fiber` is not empty the reply says that fibre is understated and names those
   foods; a row in which no entry has fibre shows `—` in the Клетчатка cell instead of 0.
D. **Skill estimates fibre for new foods** together with calories and macros (step 4 of `SKILL.md`), and label values
   include fibre when the label has it.
E. **Web API validation**: `fiber` is checked 0–100 on the website API too (`/api/ingredients`, `/api/calories`,
   `POST` and `PUT`); out of range → 400, nothing stored. An absent `fiber` means unknown and is not range-checked;
   responses omit the key when NULL. The existing validation of fats / carbs / proteins on these endpoints is not
   touched (still `>= 0` on ingredients, none on entries).
F. **AI insights are unchanged**: fibre is not added to `NutritionDataPoint`, `aggregateNutritionData` or the
   prompts.

## Requirements

### Backend
1. **Migration** `migrations/<timestamp>_add_fiber.sql` (`make migrate-create NAME=add_fiber`): add nullable `fiber
   REAL` to `user_ingredients`, `calorie_entries`, `global_ingredients`; no backfill; `Down` removes the columns.
   Existing rows read as NULL. `PRAGMA foreign_key_check` after `up` and `down` stays clean.
2. **Models**: `Fiber *float64 \`json:"fiber,omitempty"\`` on `GlobalIngredient`, `UserIngredient`, `CalorieEntry`;
   `Fiber *float64` on `NewMealEntry`.
3. **Queries and repositories**: every query listed in Current State reads/writes `fiber` after `proteins`;
   repository methods and `interfaces.go` signatures carry it (`Create`, `CreateMeal`, `Update` of the entry
   repository; `CreateOrUpdateUserIngredient`, `CreateUserIngredient`, `UpdateUserIngredient`,
   `CreateGlobalIngredient` and all scans of the ingredient repository). `CreateMeal` sets `Fiber` on the returned
   entry and on the created ingredient.
4. **`/api/v1`** (`internal/handlers/apidata`, `internal/services/calorie`, `internal/services/ingredient`):
   - `GET /ingredients`: each row has `fiber` when known.
   - `POST /ingredients`: accepts `fiber` (optional, 0–100, `null`/absent = unknown); 201 body includes it.
   - `PUT /ingredients/:id`: accepts `fiber`; absent/`null` keeps the value; a body with only `fiber` is a valid
     change; "nothing to change" message lists `fiber`.
   - `POST /food`: `new_ingredient.fiber` (stored on the entry and, unless `one_off`, on the new ingredient);
     `fiber` override on an `ingredient_id` item; without an override the entry copies the ingredient's value
     (NULL stays NULL); top-level `fiber` on a `new_ingredient` item → 400
     `items[<i>]: fiber of a new food belongs inside new_ingredient`; out of range → 400
     `items[<i>]: fiber must be between 0 and 100 (grams per 100 g)` /
     `items[<i>]: new_ingredient.fiber must be between 0 and 100 (grams per 100 g)`; the meal stays atomic.
   - `PUT /food/:id`: `fiber` field; alone it is a valid change; with `ingredient_id` the ingredient's `fiber` is
     copied first and the sent `fiber` is applied on top; cannot be reset to NULL; "nothing to change" message lists
     `fiber`.
   - Entries in the responses of `POST /food`, `PUT /food/:id` and `GET /data` include `fiber` per 100 g when known.
   - 409 bodies (`existing_ingredient`) include `fiber` through `IngredientInfo`.
   - Requests without `fiber` behave exactly as before; responses only gain an optional key.
5. **Web API**:
   - `/api/ingredients` `POST` / `PUT /:id`: optional `fiber`, validated **0–100** (decision E); `PUT` replaces it
     like the other nutrients (absent → NULL). `GET` responses include `fiber` when it is known.
   - `/api/calories` `POST` / `PUT /:id`: optional `fiber`, validated 0–100 (decision E), stored on the entry;
     `POST` passes it to the ingredient it creates for a new name; `PUT` replaces it. `GET` includes `fiber` when it
     is known.
   - The validation of fats / carbs / proteins on these endpoints stays exactly as it is.
   - Validation failures are 400 and never expose internal errors (AGENTS.md rule).
6. **Excel export**: column J as in decision 9; header style range extended to J1; `export.columns.fiber` added to
   `internal/i18n/locales/{en_US,uk_UA,ru_UA,bg_BG}.json`.
7. **AI insights**: unchanged (decision F): no edits in `internal/handlers/ai`, `internal/services/ai`.
8. **Tests** (`go vet ./... && go test ./...` green; `make build-frontend` first): extend
   `internal/services/calorie/service_test.go`, `internal/services/ingredient/service_test.go`,
   `internal/handlers/apidata/handler_test.go`, `internal/server/food_api_test.go`,
   `internal/server/ingredient_api_test.go`, `internal/server/smoke_test.go` (or a new end-to-end test) for the
   criteria in `acceptance-criteria.md`, including the backward-compatibility cases (requests without `fiber`) and an
   export test for column J.

### Skill template (edited only in `skills/kkal-tracker-food-log/`)
1. `scripts/kkal_client.py`:
   - `ingredients`: column `fiber` after `proteins` (`-` when unknown); `--json` unchanged (raw response).
   - `add`: unchanged pass-through (the JSON may carry `fiber` in `new_ingredient` or as an item override).
   - `edit`, `add-ingredient`, `edit-ingredient`: option `--fiber` (float, grams per 100 g); `0` is sent, an option
     that is not passed is not sent; the "Nothing to change" messages list `--fiber`.
   - `totals`: `fiber` in `meal` and `day_total` (after `proteins`), `without_fiber` after `without_macros`
     (decision 10). Output key order: `day, meal, day_total, missing_meal_ids, without_macros, without_fiber, check`;
     row keys `…, kcal, fats, carbs, proteins, fiber`.
   - Module docstring and `--help` texts mention fibre; `SKILL_VERSION` bumped to the date of the change.
   - Standard library only; the `API_KEY = "API_KEY_HERE"` line and `DEFAULT_BASE_URL` line stay as they are.
2. `SKILL.md` (frontmatter lines and the title line stay intact; description length limit of
   `internal/skill/skill_test.go` holds):
   - the field `fiber` (grams per 100 g; unknown vs 0; 0–100) in: the sample `ingredients` table, the override rules
     of step 3, new foods of step 4 (decision D), the JSON example and the limits of step 5, the response description;
   - step 6: `totals` sample output with `fiber` and `without_fiber`, and what `without_fiber` means;
   - step 7: table columns Ккал, Ж, У, Б, Клетчатка (fibre last, in every language), example table updated, reply
     rule of decision C (name the foods of `without_fiber`; `—` in the Клетчатка cell of a row in which no entry
     has fibre);
   - "Fixing mistakes" and "Correcting the ingredient list": `--fiber` in the examples and in the "cannot be reset to
     unknown, send 0" sentences; `--ingredient-id` copies fibre too;
   - the command list at the top (`edit … --fiber`, `add-ingredient … --fiber`).
3. Tests: `internal/server/skill_client_test.go` updated for the new key lists and extended with fibre sums
   (including a half that rounds up, an entry with fibre 0 that is **not** in `without_fiber`, an entry with NULL that
   is), `--fiber` of `edit` / `add-ingredient` / `edit-ingredient`, and the `ingredients` column;
   `TestSkillDownloadMatchesTheRepository` and `internal/skill` tests stay green.

### Frontend
1. Types: `fiber?: number` in `Ingredient`, `CreateIngredientData`, `UpdateIngredientData` (`api/ingredients.ts`) and
   `CalorieEntry` (`api/calories.ts`).
2. **Ingredient form** (`AddIngredientModal.tsx`, `EditIngredientModal.tsx`): field "Клетчатка, г/100 г" after
   proteins, `CalculatorInput`, optional (placeholder `foodList.optional` in the add modal like the macros); empty →
   not sent (unknown), `0` → sent as 0; the edit modal shows a stored 0 as `0`. Layout per the `frontend-style` skill
   (the grid currently holds three fields; mobile and desktop must stay tidy with four).
3. **Ingredient list** (`IngredientsTable.tsx`): column "Клетчатка" after proteins, `-` when unknown, `0` when zero;
   empty-state `colSpan` adjusted; usable on mobile.
4. **Diary forms** (`AddFoodEntryForm.tsx`, `EditEntryModal.tsx`): fibre field after proteins, in both the mobile and
   the desktop layout of the add form; prefilled from the selected ingredient (0 stays `0`, unknown stays empty);
   sent with the entry when not empty, including 0 (decision B); the edit modal preloads the entry's value, so saving
   an unrelated change keeps the fibre.
5. **Day view**:
   - `EntryListItem.tsx`: fibre grams of the entry (`fiber * weight / 100`, one decimal) after proteins when the
     value is known, `0.0` for a known zero, nothing when unknown.
   - `CalorieEntriesList.tsx` (collapsed day header and expanded footer) and `FilterSection.tsx` with
     `DashboardPage.tsx` (filtered-range total): fibre total after proteins by the rule of decision A, with the
     understated hint and the number of entries without data.
6. **I18n**, all 4 locales (`en_US`, `uk_UA`, `ru_UA`, `bg_BG`), every new string through `t()`, units through
   `common.g`: ingredient form label (ru "Клетчатка, г/100 г"), table header / short name (ru "Клетчатка"), diary form
   label (in the style of `dashboard.proteins`), consumed label for totals and entry lines, the understated hint with
   `{{count}}` (worded so that it needs no plural forms).
7. Checks from `web/`: `npx tsc --noEmit -p tsconfig.app.json`, `npm run lint`, `npm run build`. Tailwind utilities
   only; no CSP-relevant changes.

### Docs
- `docs/api.md`: `fiber` in `/api/ingredients`, `/api/calories` and every `/api/v1` endpoint that lists the macros
  (request shapes, limits, error texts, copy/override rules).
- `docs/features.md`: calorie tracking / ingredient database (fibre stored, unknown vs 0), day view totals and the
  understated hint, skill reply table (Ккал, Ж, У, Б, Клетчатка) and `totals`, per-entry overrides incl. `fiber`,
  export column.
- `README.md`: the `/api/v1` meal example and the override / `PUT /food/:id` bullets, the skill line about `totals`.
- `.agents/skills/frontend-style/SKILL.md`: the list of numeric inputs with the calculator gains fibre.
- `AGENTS.md`: no change expected.

### DevOps
- The release contains a migration. Migrations are never applied automatically: apply `make migrate-up` on
  production as part of the release per the `deploy` skill, before the new binary serves requests (the new code
  selects `fiber`). The previous binary works with the added columns (it uses explicit column lists), so the order
  "migrate, then roll out" is safe.
- The release ships the new skill template; installed copies stay as they are until downloaded again, and keep
  working (they never send `fiber` and ignore the new response key).

## Subtasks

### Backend Developer
1. Migration (Backend 1) and apply it to the dev database.
2. Models, queries (both dialect variants), repositories, interfaces (Backend 2–3).
3. Calorie and ingredient services + `/api/v1` handlers and DTOs (Backend 4).
4. Web API handlers `/api/ingredients`, `/api/calories` (Backend 5).
5. Excel export column and backend locale keys (Backend 6).
6. Skill template: `kkal_client.py`, `SKILL.md`, `SKILL_VERSION` (Skill template 1–2).
7. Tests (Backend 8, Skill template 3).
8. Docs: `docs/api.md`, `docs/features.md` (backend and skill parts), `README.md`.

### Frontend Developer
1. Types and the ingredient form, ingredient table (Frontend 1–3).
2. Diary add and edit forms (Frontend 4).
3. Day view: entry line, day totals, filtered total, understated hint (Frontend 5).
4. Locale keys in all 4 locales (Frontend 6).
5. Docs: `docs/features.md` (UI part), `.agents/skills/frontend-style/SKILL.md`.

### Reviews
`backend-code-reviewer` (incl. the Python client and `SKILL.md`), `frontend-code-reviewer`, `security-reviewer`
(input validation of the new field on all four write paths, no internal errors exposed, no key material touched in
the skill template), `qa-engineer` (criteria coverage, backward compatibility, NULL vs 0).

### DevOps Engineer
1. Release with the migration applied (DevOps section), after the user's confirmation.

## Out of scope
- Backfilling fibre for existing ingredients or entries (done later by the AI skill through the API), fibre values in
  the seed data.
- Fibre goals / daily norm, fibre in reports and charts.
- Changing how the existing macro fields treat 0 in the web diary forms, and the range validation of the macros on
  the website API (decisions B and E).
- Fibre in AI insights (decision F).

## Technical choices left to the developers
There are no open questions for the user. These are implementation choices, to be settled with the reviewers:
- SQLite `ALTER TABLE … DROP COLUMN` in the `Down` migration vs a table rebuild: developer's choice; the migration
  runs with foreign keys off and must pass `foreign_key_check`.
- Whether to add `fiber` to the long positional parameter lists of the repositories or to introduce a small nutrients
  struct: developer's choice, to be agreed with the backend reviewer.
- Hint presentation in the day totals (text, icon with `title`, `≥` prefix): frontend developer's choice within the
  `frontend-style` skill, as long as the understated state and the count are visible without hovering on mobile.
