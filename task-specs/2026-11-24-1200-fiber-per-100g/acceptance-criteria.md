# Acceptance Criteria: Dietary fibre per 100 g on ingredients and diary entries

Letters in brackets refer to the user-confirmed decisions A–F of `requirements.md`.

BA acceptance 2026-11-24: all criteria met. Verified by reading the code and docs, with the QA test results
(`go vet`, `go test`, `go test -race`, `tsc`, lint, frontend build) as evidence. Criteria that rest on manual runs
only (no committed test): the previous skill client against the new server, every Frontend criterion, the two
Integration criteria that involve the web UI, `make seed`, and the `SKILL.md` wording.

## Backend

### Migration and storage
- [x] `make migrate-up` adds nullable `fiber REAL` to `user_ingredients`, `calorie_entries`, `global_ingredients`;
      every existing row has `fiber` NULL; `foreign_key_check` passes
- [x] `make migrate-down` removes the three columns and leaves all other data intact; `foreign_key_check` passes
- [x] A user registered after the migration gets the ingredient list copied as before (fibre NULL, since the global
      values are NULL); a global ingredient with a fibre value is copied with that value
- [x] `make seed` works unchanged

### `/api/v1` ingredients
- [x] `POST /api/v1/ingredients {"name":"Овсянка","kcal_per_100g":370,"fiber":10.6}` → 201, body has `"fiber":10.6`;
      `GET /api/v1/ingredients` shows the row with `"fiber":10.6`
- [x] `POST` with `"fiber":0` → 201, `"fiber":0` is returned and listed (zero is not unknown)
- [x] `POST` without `fiber` (or `"fiber":null`) → 201, no `fiber` key in the body and in the list row
- [x] `POST` with `"fiber":-1` or `"fiber":100.1` → 400 `fiber must be between 0 and 100 (grams per 100 g)`,
      nothing created; `"fiber":100` is accepted
- [x] `PUT /api/v1/ingredients/:id {"fiber":2.5}` → 200, `ingredient.fiber` 2.5, name / kcal / macros unchanged
- [x] `PUT /api/v1/ingredients/:id {"kcal_per_100g":260}` on an ingredient with fibre 2.5 → fibre stays 2.5
- [x] `PUT /api/v1/ingredients/:id {"fiber":null}` alone → 400 "nothing to change …" (message lists `fiber`), value
      kept; `{"fiber":0}` → 200 and the value becomes 0
- [x] `PUT /api/v1/ingredients/:id {"fiber":150}` → 400 naming `fiber`, nothing changed
- [x] 409 `existing_ingredient` bodies include `fiber` when the existing ingredient has it

### `/api/v1` food
- [x] `POST /api/v1/food` with `{"ingredient_id":<id with fiber 10.6>,"weight":50}` → the entry has `"fiber":10.6`
- [x] The same with `"fiber":8` on the item → the entry has 8, the ingredient still has 10.6
- [x] An `ingredient_id` item whose ingredient has unknown fibre and no override → the entry has no `fiber` key
- [x] `{"new_ingredient":{"name":"X","kcal_per_100g":100,"fiber":3},"weight":100}` → entry and the new ingredient
      both have fibre 3; with `"one_off":true` the entry has 3 and no ingredient is created
- [x] `{"new_ingredient":{…},"weight":100,"fiber":3}` → 400
      `items[0]: fiber of a new food belongs inside new_ingredient`, nothing saved
- [x] Item override `"fiber":101` → 400 `items[<i>]: fiber must be between 0 and 100 (grams per 100 g)`;
      `new_ingredient.fiber` 101 → 400 `items[<i>]: new_ingredient.fiber must be between 0 and 100 (grams per 100 g)`;
      in a meal of several items nothing is saved
- [x] `PUT /api/v1/food/:id {"fiber":4}` → 200, `entry.fiber` 4, calories / weight / macros / time unchanged
- [x] `PUT /api/v1/food/:id {"ingredient_id":<id with fiber 10.6>}` → entry fibre 10.6; with
      `{"ingredient_id":<same>,"fiber":7}` → 7; with an ingredient of unknown fibre and no override → the entry's
      fibre becomes unknown (copied like the macros)
- [x] `PUT /api/v1/food/:id {"weight":120}` on an entry with fibre 4 → fibre stays 4
- [x] `PUT /api/v1/food/:id {"fiber":null}` alone → 400 "nothing to change …" (lists `fiber`); `{"fiber":-0.1}` → 400
      naming `fiber`, nothing changed
- [x] `GET /api/v1/data?type=food&from=…&to=…` returns `fiber` per 100 g on entries that have it and omits it on the
      others
- [x] Editing the ingredient's fibre afterwards does not change entries logged before

### Backward compatibility
- [x] Every existing request shape without `fiber` (`POST`/`PUT /api/v1/ingredients`, `POST`/`PUT /api/v1/food`,
      `/api/ingredients`, `/api/calories`) returns the same status and the same values as before; the existing test
      suites pass with changes limited to places that assert the full key list of a response
- [x] `day_total_calories` and `total_calories` are not affected by fibre
- [x] The previous skill client (version 2026-10-16) runs `ingredients`, `add`, `entries`, `totals`, `edit`,
      `add-ingredient`, `edit-ingredient` against the new server without errors

### Web API
- [x] `POST /api/ingredients {"name":"N","kcalPer100g":100,"fiber":5}` → 201 with `"fiber":5`; `GET /api/ingredients`
      and `GET /api/ingredients/:id` include it
- [x] `PUT /api/ingredients/:id` with `"fiber":0` stores 0; without `fiber` clears it (replace-all, like the macros)
- [x] (E) `POST`/`PUT /api/ingredients` with `"fiber":-1` or `101` → 400, nothing stored / nothing changed;
      `"fiber":0` and `"fiber":100` are accepted
- [x] (E) `POST`/`PUT /api/ingredients` and `/api/calories` without `fiber` succeed as before (absent is unknown and
      is not range-checked) and the response has no `fiber` key
- [x] (E) Macro validation on the website API is unchanged: `POST /api/ingredients` with `"fats":150` is still
      accepted and `"fats":-1` is still rejected; `POST /api/calories` with `"fats":150` is still accepted
- [x] `POST /api/calories` with `"fiber":3` → entry has `"fiber":3`; for a food name that is not in the list the
      created ingredient has fibre 3; with `"fiber":0` the entry has 0; without `fiber` the entry has none
- [x] `PUT /api/calories/:id` with `"fiber":6` → 6; `GET /api/calories?dateFrom=…&dateTo=…` includes `fiber`
- [x] (E) `POST`/`PUT /api/calories` with `"fiber":-1` or `101` → 400, nothing stored / nothing changed
- [x] No error response of the changed handlers contains internal error text

### AI insights
- [x] (F) `internal/handlers/ai` and `internal/services/ai` (types and prompts) are not changed: the prompt built for
      a day with fibre data is the same as before and does not mention fibre

### Export
- [x] The food sheet has a tenth column (J) with the translated header in en_US, uk_UA, ru_UA, bg_BG, bold like the
      others; a row with fibre 10.6 shows 10.6, a row with 0 shows 0, a row with unknown fibre has an empty cell;
      columns A–I are unchanged

### Quality
- [x] `go vet ./...` and `go test ./...` are green (after `make build-frontend`), including the end-to-end tests
- [x] No direct `log/slog` use added in handlers/services beyond the injected logger; comments in English

## Skill template
- [x] `kkal_client.py ingredients` prints the header
      `id name kcal_per_100g fats carbs proteins fiber times_used last_used` (tab separated) and `-` for unknown
      fibre, `0` for zero
- [x] `add --json` with `fiber` inside `new_ingredient` or as an item override reaches the server and the response
      shows it
- [x] `edit ID --fiber 4` changes only the fibre; `edit ID --fiber 0` stores 0; `edit ID` without options fails with
      a message that lists `--fiber`
- [x] `add-ingredient --name N --kcal 100 --fiber 3` creates the ingredient with fibre 3; without `--fiber` the
      fibre is unknown; `--fiber 0` stores 0
- [x] `edit-ingredient ID --fiber 2.5` changes only the fibre; without options the failure message lists `--fiber`
- [x] `totals --day D --meal …`: `meal` and `day_total` end with `fiber` (whole grams, halves up, e.g. 2.5 g → 3);
      the output keys are `day, meal, day_total, missing_meal_ids, without_macros, without_fiber, check`
- [x] `without_fiber` lists `{id, food, in_meal}` of exactly the entries of the day whose fibre is unknown; an entry
      with fibre 0 is not listed; `without_macros` is the same as before for the same data
- [x] Against entries that all lack fibre: `fiber` is 0 in both rows and every entry is in `without_fiber`
- [x] `SKILL_VERSION` is a later date than `2026-10-16` (the date of the change) and `version` prints it
- [x] `SKILL.md` documents `fiber` (unit, unknown vs 0, limits), the override rules for `add` / `edit`, `--fiber` of
      the three commands, `without_fiber`, and the reply table has the columns Ккал, Ж, У, Б, Клетчатка with the
      example updated
- [x] (C) `SKILL.md` tells the agent to say that fibre is understated and to name the foods of `without_fiber`, and
      to show `—` (not 0) in the Клетчатка cell of a row in which no entry has fibre; a row with at least one known
      value shows the number, also when it is 0
- [x] (D) `SKILL.md` step 4 tells the agent to estimate fibre per 100 g for a new food together with calories, fats,
      carbs and proteins, to use the label value when the label has one, and the `new_ingredient` examples carry
      `fiber`
- [x] The template still holds the `API_KEY_HERE` placeholder exactly once, no real key; `internal/skill` tests and
      `TestSkillDownloadMatchesTheRepository` pass; the downloaded zip contains the new client and `SKILL.md`
- [x] No installed copy (`~/.claude/skills/…`) was edited

## Frontend
- [x] Add ingredient modal: a "Клетчатка, г/100 г" field (ru_UA) after proteins; saving with 10.6 creates the
      ingredient with fibre 10.6; left empty → unknown; `0` → 0
- [x] Edit ingredient modal: shows the stored value (`0` for zero, empty for unknown); changing it saves; clearing it
      makes the value unknown; saving another field keeps the fibre
- [x] Ingredient table: column "Клетчатка" after proteins; `-` for unknown, `0` for zero; the empty-state row spans
      all columns; no horizontal breakage on a phone width
- [x] Add entry form (mobile and desktop layouts): fibre field after proteins; selecting an ingredient with fibre
      10.6 prefills 10.6, with 0 prefills `0`, with unknown leaves it empty; the saved entry has the same value
- [x] Edit entry modal: shows the entry's fibre; changing only the weight keeps the fibre (also when it is 0);
      changing the fibre saves it
- [x] (B) Only fibre keeps 0 in the diary forms: typing `0` in the fibre field of the add or edit entry form stores
      fibre 0, while fats / carbs / proteins of 0 are still not sent, exactly as before the change
- [x] Entry line: fibre grams after proteins (`fiber × weight / 100`, one decimal); `0.0` for a known zero; nothing
      for unknown
- [x] (A) Day total (collapsed header and expanded footer), all entries with known fibre: "Клетчатка: X.X г" after
      proteins, no hint; equals the sum of the entry lines
- [x] (A) Day with entries of fibre 0 only: the total shows 0.0 without a hint
- [x] (A) Day with at least one known and at least one unknown entry: the total is shown with a visible
      "understated" hint that includes the number of entries without data (e.g. 2)
- [x] (A) Day where no entry has fibre: no fibre total and no hint; the macro totals look as before
- [x] (A) Filtered-range total (`FilterSection`) follows the same rules for the whole range
- [x] Order everywhere: calories, fats, carbs, proteins, fibre
- [x] All new strings exist in `en_US`, `uk_UA`, `ru_UA`, `bg_BG` and are rendered through `t()`; grams use
      `common.g`; switching the language changes every new label and the hint
- [x] Fibre inputs use the integrated calculator like the other numeric inputs
- [x] `npx tsc --noEmit -p tsconfig.app.json`, `npm run lint`, `npm run build` pass; Tailwind utilities only; no
      inline scripts or new external hosts

## Docs
- [x] `docs/api.md` lists `fiber` on every endpoint that lists the macros, with limits and the copy / override /
      "cannot be reset" rules
- [x] `docs/features.md` describes fibre storage (unknown vs 0), the day view total with the hint, the skill table
      and `totals`
- [x] `README.md` API and skill paragraphs mention `fiber`; `.agents/skills/frontend-style/SKILL.md` lists the fibre
      input

## Integration
- [x] End to end through `/api/v1`: create an ingredient with fibre 10 → `add` a meal of 50 g of it and 100 g of a
      one-off food with fibre 0 and 100 g of a one-off food without fibre → the entries have 10, 0 and no value →
      `totals` for the meal and the day returns `fiber: 5`, and `without_fiber` holds only the third entry
- [x] The same day in the web day view shows "Клетчатка: 5.0 г" with the hint "1 entry without data"
- [x] `edit-ingredient --fiber 12` afterwards: the ingredient has 12, the logged entry still has 10;
      `edit ENTRY --ingredient-id <same id>` copies 12 into the entry
- [x] An ingredient created with fibre in the web UI is returned with the same value by `GET /api/v1/ingredients`,
      and one created through `/api/v1` shows its value in the web ingredient table and edit modal
- [x] Excel export of that day contains the fibre values of the three entries (10 / 0 / empty)
- [x] After the release migration all pre-existing ingredients and entries show unknown fibre and every pre-existing
      screen and API call works as before
