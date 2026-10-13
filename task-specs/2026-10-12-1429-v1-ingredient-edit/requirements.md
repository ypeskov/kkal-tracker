# Edit and delete ingredients through the external API (`/api/v1`), duplicate-name protection on the website

## Overview
The food logging skill for AI agents (`skills/kkal-tracker-food-log/`) can create ingredients in the user's ingredient
list (`new_ingredient` in `POST /api/v1/food`) but cannot fix or remove them. When an estimate was wrong ("у этого
продукта на самом деле 260 ккал", "в справочнике неправильно", a typo in the name, a duplicate created by mistake), the
user has to open the Food List page of the website. The website already can do it (Food List → EditIngredientModal →
`PUT /api/ingredients/:id`, `DELETE /api/ingredients/:id`), but those routes accept only the JWT, while the agent
authenticates with `X-API-Key`, which works only on `/api/v1/*` (`internal/server/server.go:197`).

While analysing it, a website bug was found and the user decided to fix it in this task: creating or renaming an
ingredient on the Food List page to the name of an existing one fails with a generic 500 and the modal gives no
feedback at all.

Business value: the user corrects the ingredient list by voice in the same conversation in which the agent logs food,
the list (which the agent matches foods against) stays accurate and free of duplicates, and the website explains a
duplicate name instead of silently failing.

## Current State

### Routes and auth
- `/api/v1` group: rate limiter + `apiKeyMiddleware.RequireAPIKey` (`internal/server/server.go:194-198`), handler
  `internal/handlers/apidata/handler.go`, routes `GET /data`, `GET /ingredients`, `POST /food`, `PUT /food/:id`,
  `DELETE /food/:id`. DTOs in `internal/handlers/apidata/dto.go` (snake_case JSON, `IngredientInfo`).
- `/api/ingredients` (JWT, `internal/handlers/ingredients/handler.go`), responses use `models.UserIngredient`
  (`id`, `user_id`, `name`, `kcalPer100g`, `fats`, `carbs`, `proteins`, `global_ingredient_id`, `created_at`,
  `updated_at`), the frontend type `Ingredient` in `web/src/api/ingredients.ts` matches it:
  - `POST ""` — create; `name` and `kcalPer100g` required, values `>= 0`, no upper limits, the name is stored as sent,
    **no duplicate check**: an exact duplicate name violates the unique index `idx_user_ingredients_unique
    (user_id, name)` (`migrations/20250907200003_...`) → 500 `Internal server error`. Near-duplicates (`Гречка` /
    `гречка`, `ё`/`е`) are accepted (the index is exact and case-sensitive in SQLite).
  - `PUT /:id` — **full replace** (macros omitted → NULL), same validation, **same 500** on renaming to the exact name
    of another ingredient, near-duplicates accepted.
  - `DELETE /:id` → 204, 404 when the row is not the user's.
- No handler or service tests exist for ingredients; the smoke test only checks `GET /api/ingredients/999999` → 404.

### Website (Food List)
- `web/src/pages/FoodList.tsx`: `createMutation` / `updateMutation` have only `onSuccess` (close modal, invalidate
  `['ingredients']`); **no `onError`**: on any failure the modal stays open and nothing is shown.
- `web/src/api/ingredients.ts`: `createIngredient` / `updateIngredient` throw `new Error(await response.text())`, i.e.
  the raw server body; the status is lost.
- `AddIngredientModal.tsx` / `EditIngredientModal.tsx`: forms without any error display; they get `onCreate` /
  `onUpdate`, `isCreating` / `isUpdating` from the page.
- Existing inline error pattern to follow: `WeightGoalForm.tsx:359`
  (`p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700`). Locales: `web/src/i18n/locales/{en_US,
  uk_UA,ru_UA,bg_BG}.json`, section `foodList`.
- Other website path that creates ingredients: adding a diary entry (`POST /api/calories` → `CalorieService.CreateEntry`
  → `CreateOrUpdateUserIngredient`) upserts by **exact** name and never returns 500 on a duplicate. It is out of scope
  and stays unchanged (it can still create a near-duplicate such as `гречка` next to `Гречка`).

### Ingredient service (`internal/services/ingredient/`)
- `Servicer`: `GetAllIngredients`, `GetIngredientsWithUsage`, `GetIngredientByID(userID, id)`, `CreateIngredient`,
  `UpdateIngredient(*UpdateIngredientRequest)` (full replace), `DeleteIngredient(userID, id)`.
- Ownership is enforced in SQL: `UPDATE ... WHERE user_id = ? AND id = ?`, `DELETE ... WHERE user_id = ? AND id = ?`,
  `SELECT ... WHERE user_id = ? AND id = ?`. Not found / foreign → `sql.ErrNoRows` (update: from the re-read after the
  update; delete: `RowsAffected == 0`).

### Validation and name normalization of `POST /api/v1/food` (`internal/services/calorie/service.go`)
- Limits (unexported constants): `maxKcalPer100g = 900`, `maxNutrientPer100g = 100`, `maxFoodNameLength = 200` (runes).
- `cleanFoodName`: drops control/format characters (e.g. zero-width space), collapses whitespace, trims. The cleaned
  name is what gets stored.
- `normalizeFoodName`: `cleanFoodName` + lower case + `ё` → `е`; used for the duplicate check that returns
  `DuplicateIngredientError` → handler 409 `{message, item_index, existing_ingredient}`.
- These helpers live in the `calorie` package and are unexported; the ingredient service cannot use them today.

### Diary entries vs. ingredients (findings for the skill text)
Verified in the schema (`migrations/20250907200000_initial_schema.sql`, no later migration touches `calorie_entries`),
`internal/repositories/queries.go`, `internal/services/calorie/service.go` and `web/src/pages/FoodList.tsx`:

1. **Diary entries copy the ingredient, they do not reference it.** `calorie_entries` has no ingredient column and no
   foreign key to `user_ingredients`; an entry stores its own `food` (name), `kcal_per_100g`, `fats`, `carbs`,
   `proteins` and the computed `calories`. `CreateMeal` and `UpdateMealEntry` copy these values from the ingredient
   at the moment of the request.
2. **Editing an ingredient does not change any existing diary entry** — not the name, not the calories, not the macros,
   not the day totals. Only entries logged afterwards with that `ingredient_id` get the new values. The website behaves
   the same way: after an update it only refreshes the ingredient list.
3. **Applying a corrected ingredient to an already logged entry** is possible with the existing
   `PUT /api/v1/food/:id {"ingredient_id": <same id>}`: it re-copies name, calories and macros from the (now updated)
   ingredient and recalculates the calories, keeping weight and time. `kcal_per_100g` on `PUT /food/:id` changes only
   the calorie value of that entry, not the macros.
4. **Renaming resets the usage statistics.** `times_used` / `last_used` in `GET /api/v1/ingredients` are computed by an
   exact name match (`calorie_entries.food = user_ingredients.name`, `QueryGetUserIngredientsWithUsage`). Old entries
   keep the old name, so after a rename (even a case-only one) the ingredient shows `times_used: 0` until it is logged
   again; if another ingredient later gets the old name, it inherits those statistics. The agent loses the "habit"
   signal it uses for matching, so it should rename only when the user asks for it or the name is clearly wrong.
5. **Deleting an ingredient leaves all diary entries untouched**: no cascade, they keep their name and calories, still
   count toward the day totals and can still be edited with `PUT /food/:id` (weight, `kcal_per_100g`, time, switching to
   another existing ingredient) and deleted with `DELETE /food/:id`. The deleted `ingredient_id` becomes invalid:
   `POST /food` and `PUT /food/:id` with it return 400 `ingredient N not found`. Deletion cannot be undone through the
   API (the ingredient would have to be created again as a `new_ingredient`, with a new ID).
6. The website keeps a copy of the ingredient list in `sessionStorage`; an edit made by the agent shows up there after
   the list is reloaded. This is the existing behavior for ingredients created by the agent as well; no change.

## Requirements

### Backend

#### Common duplicate-name rule (all ingredient writes of this task)
Applies to `PUT /api/v1/ingredients/:id`, `PUT /api/ingredients/:id` and `POST /api/ingredients`; `POST /api/v1/food`
already follows it.
- Names are compared normalized exactly like `POST /api/v1/food` (`normalizeFoodName`: case, spacing, invisible
  characters, `ё`/`е`), only within the same user's list.
- Create: the normalized name equals the normalized name of any of the user's ingredients → 409, nothing saved.
- Update: the check runs **only when the normalized name changes** (normalized new name ≠ normalized current name of
  this ingredient). Consequences: renaming to a variant of the own name (`гречка` → `Гречка`, `ёжик` → `ежик`) is
  allowed; and an ingredient that already has a near-duplicate in the list from before this change (e.g. `Гречка` and
  `гречка` both exist) can still be edited (calories, macros, case of its own name) without a 409. Renaming it to a
  normalized name that belongs to another ingredient → 409, nothing changed.
- One implementation of the check in the ingredient service, used by all three endpoints (no copies per handler).
- A race between the check and the write (unique index violation on the exact name) must not leak the database error:
  either 409 or the generic 500.
- 409 body: `{"message": "an ingredient with this name already exists", "existing_ingredient": {...}}`. The key and the
  message are the same on both APIs; the inner object uses the ingredient shape **of the API it is returned from**:
  `IngredientInfo` (snake_case) on `/api/v1`, `models.UserIngredient` (the shape of `GET /api/ingredients/:id`, which
  the frontend `Ingredient` type already parses) on `/api/ingredients`. Reason for the difference: each client then
  reads the existing ingredient with the type it already has for that API. No `item_index` outside `POST /food`.

#### `PUT /api/v1/ingredients/:id` — partial update of a user ingredient
- Auth and rate limit: the existing `/api/v1` group (`X-API-Key`), no new route groups.
- Request body (all fields optional, JSON `null` is treated the same as "not sent"):
  ```json
  {"name": "Хлеб ржаной", "kcal_per_100g": 260, "fats": 1.2, "carbs": 48, "proteins": 8.5}
  ```
  Fields that are not sent keep their stored values (same semantics as `PUT /api/v1/food/:id`). Clearing a macro
  (setting it back to "unknown") is not supported through this endpoint; `0` can be sent instead.
- Validation (only the fields that are sent; stored values that are not sent are kept even if they are outside the
  limits, so that e.g. a rename is not blocked by an old website value of 1000 kcal):
  - `name`: cleaned like `new_ingredient.name` in `POST /food` (`cleanFoodName`), the cleaned name is stored; empty
    after cleaning → 400; longer than 200 characters → 400
  - `kcal_per_100g`: 0–900 inclusive, otherwise 400
  - `fats`, `carbs`, `proteins`: 0–100 inclusive (grams per 100 g), otherwise 400
  - nothing to change (empty object, `{}` or only `null`s) → 400 with a message listing the accepted fields
  - malformed JSON or wrong types → 400 `Invalid request body`
  - non-numeric `:id` → 400 `Invalid ingredient ID`
  - 400 messages name the field and the limit (like the `POST /food` / `PUT /food` messages) and never contain
    internal error text
- Ownership: an ingredient that does not exist or belongs to another user → 404 `Ingredient not found` (the two cases
  are indistinguishable for the client). The other user's ingredient must stay unchanged.
- Duplicate name → 409 per the common rule:
  ```json
  {"message": "an ingredient with this name already exists", "existing_ingredient": {"id": 12, "name": "...", "kcal_per_100g": 250, "fats": 1, "carbs": 2, "proteins": 3}}
  ```
- Success → 200 with the updated ingredient in the `IngredientInfo` shape, wrapped like `PUT /food/:id` wraps its entry:
  ```json
  {"ingredient": {"id": 94, "name": "Хлеб ржаной", "kcal_per_100g": 260, "fats": 1.2, "carbs": 48, "proteins": 8.5}}
  ```
- Diary entries are not touched (see Current State, findings 1–4).
- Unexpected errors: logged with the injected logger (`error`, `user_id`, `ingredient_id`), client gets 500 with a
  generic message (`Failed to update ingredient`), never `err.Error()`.

#### `DELETE /api/v1/ingredients/:id`
- Success → 204, no body. The ingredient disappears from `GET /api/v1/ingredients` and from the website's
  `GET /api/ingredients`; diary entries are untouched (finding 5).
- Not existing / another user's / already deleted → 404 `Ingredient not found`; the other user's ingredient stays.
- Non-numeric `:id` → 400 `Invalid ingredient ID`.
- Unexpected errors → logged, 500 `Failed to delete ingredient`.
- Reuses `ingredientService.DeleteIngredient` as is.

#### Website API: `POST /api/ingredients` and `PUT /api/ingredients/:id` (JWT)
- Add the duplicate-name rule: 409 with `existing_ingredient` in the `models.UserIngredient` shape instead of the
  current 500. Applies to exact duplicates and to near-duplicates (case, spacing, invisible characters, `ё`/`е`).
- Everything else stays as it is: full-replace semantics of `PUT`, the request field names (`kcalPer100g`), the current
  validation (`>= 0`, no upper limits — the website limits are not aligned with the API limits in this task), the name
  stored as sent, the success responses (201 / 200 with `models.UserIngredient`), 404 for a foreign or missing
  ingredient on `PUT`.
- Unexpected errors keep returning the generic 500 and are logged (no `err.Error()` in responses).

#### Implementation constraints
- Reuse the existing ingredient service (`internal/services/ingredient`) and repository (`UpdateUserIngredient`,
  `DeleteUserIngredient`, `GetUserIngredientByID`, `GetAllUserIngredients`, `CreateUserIngredient`); no new SQL is
  expected, if any is needed it goes to `internal/repositories/queries.go`. No migration; existing near-duplicates in
  user lists are not merged or changed.
- The validation limits and the name cleaning/normalization must be **the same code** as in `POST /api/v1/food`, not a
  copy that can drift (see Open Questions for developers, Q-D1).
- Route registration of the v1 endpoints in `apidata.Handler.RegisterRoutes`; request/response DTOs in
  `internal/handlers/apidata/dto.go`. The website 409 body is built in `internal/handlers/ingredients/handler.go`.

#### Tests
- Service unit tests:
  - partial update: merge of sent/not sent fields, every limit (boundary values 0, 900, 100 pass; -0.1, 900.1, 100.1
    fail), name cleaning, empty name, 201-character name, nothing to change, not found / foreign ingredient
  - duplicate rule for create, full update and partial update: case / spacing / invisible character / `ё` variants →
    duplicate error carrying the existing ingredient; own-name variant allowed; editing an ingredient that already has
    a pre-existing near-duplicate without changing its normalized name allowed
- Integration tests on the fully wired server (`internal/server`, same environment as the smoke test; helpers
  `signUp`, `addIngredient`, `listIngredients` in `food_api_test.go`):
  - v1: edit, partial edit, 409, 400s, 404 for a foreign ingredient with a second user, delete, delete twice, diary
    entries unchanged after edit and delete, `ingredient_id` of a deleted ingredient rejected by `POST /food`, re-sync
    of an entry with `PUT /food/:id {"ingredient_id"}` after an edit, 401 without a key
  - website API: `POST` and `PUT /api/ingredients` → 409 with `existing_ingredient` for exact and near-duplicates,
    own-name variant → 200, other users' names do not conflict, ordinary create/update still 201/200
  - extending `TestSmokeUserJourney` is optional (developer's call)
- `go vet ./...` and `go test ./...` green (after `make build-frontend`).

### Skill and client (part of the backend subtask)

#### `skills/kkal-tracker-food-log/scripts/kkal_client.py`
- `edit-ingredient INGREDIENT_ID [--name NAME] [--kcal N] [--fats N] [--carbs N] [--proteins N]` →
  `PUT /ingredients/{id}` with only the passed options; no options → `fail("Nothing to change: pass --name, --kcal,
  --fats, --carbs or --proteins")` without a request; expects 200 and prints the JSON. Help texts make clear that
  `--kcal` here changes the ingredient for all future entries (unlike `edit --kcal`, which is entry-only).
- `delete-ingredient INGREDIENT_ID` → `DELETE /ingredients/{id}`, expects 204, prints `Ingredient N deleted`.
- Same style as `edit` / `delete` (`request`, `expect`, `fail`, argparse subparser, `type=int` / `type=float`), module
  docstring lists the two new commands. Standard library only.
- A 404 / 405 that is not the server's `Ingredient not found` answer (an older server without these routes) must not be
  presented as "no such ingredient": print a hint that the server probably does not support ingredient editing yet
  and that nothing was changed (in the spirit of the existing "older server" message in `request`).

#### `skills/kkal-tracker-food-log/SKILL.md` (English, existing style and tone)
- Frontmatter `description`: mention fixing or removing a food in the user's ingredient list (e.g. "исправь в
  справочнике", "у этого продукта на самом деле 260 ккал") so that the skill triggers for it.
- "Production and dev" section: ingredient IDs are separate on each server too; edit or delete an ingredient with the
  copy whose ingredient list it came from (same rule as for entries).
- Command list in "API basics": the two new commands.
- Step 3 (the sentence about `kcal_per_100g` overriding a value for one entry) and "Fixing mistakes": when to edit the
  ingredient vs. override one entry (decided by the user):
  - **Edit the ingredient directly** (`edit-ingredient`) only when the user points at the product or the list: "в
    справочнике", "на самом деле", "всегда" (e.g. "у этого продукта на самом деле 260 ккал", "в справочнике
    неправильно", "он всегда 260"), or asks to fix a typo in the name / a bad estimate of an ingredient the agent
    created.
  - **Otherwise override for the entry only** (`kcal_per_100g` on the item / `edit --kcal`), as now — e.g. "этот хлеб
    260 ккал" while logging — and **offer in one line** of the reply to update the value in the list as well. Update the
    list only after the user agrees.
  - Editing an ingredient does not change entries already in the diary (findings 1–2). **Which entries to refresh
    (decided by the user):** only entries from the current conversation that the correction is clearly about (e.g. the
    meal just logged with the wrong value) — refresh them with `edit ENTRY_ID --ingredient-id <same id>` (finding 3)
    and report them. Older entries are not touched unless the user asks.
  - Renaming resets `times_used` / `last_used` (finding 4): rename only when asked or when the name is clearly wrong.
  - Report the change in the reply: old → new values, and which entries were refreshed.
- **Delete an ingredient only on the user's explicit request** ("удали из справочника", "убери этот продукт из
  списка"); never as a side effect of undoing an entry, never to "clean up" on the agent's own initiative. Say that the
  diary entries stay as they are (finding 5) and that deletion cannot be undone through the API.
- Replace the sentence "Ingredients created by mistake cannot be removed through this API; tell the user to delete them
  on the Food List page of the website." with the new rules (an ingredient created by mistake → `delete-ingredient`
  when the user asks for it; a wrong estimate → `edit-ingredient`).
- "Error handling": `edit-ingredient` 409 → another ingredient already has that name, `existing_ingredient` shows it,
  nothing changed: tell the user and ask (keep both under different names, or use the existing one); do not delete or
  merge anything on your own. `404` on `edit-ingredient` / `delete-ingredient` → no such ingredient (already deleted,
  wrong ID, or the other server's ID): re-fetch `ingredients`. `400` → the message names the field and the limit.
  The "older server" hint from the client → tell the user the server does not support it yet, do not retry.
- Keep the limits line consistent: `kcal_per_100g` 0–900, macros 0–100 g also apply to `edit-ingredient`.
- Regenerate the dev copy: `python3 skills/sync_dev_skill.py` (no argument, default dev URL). Do not touch the
  `api_key` files; the repository keeps the `XXXXXXXXX` placeholder.

### Docs
Backend subtask:
- `docs/api.md`: `PUT /api/v1/ingredients/:id` (`{name?, kcal_per_100g?, fats?, carbs?, proteins?}`, partial, 409 on
  a duplicate name) and `DELETE /api/v1/ingredients/:id`; for `/api/ingredients` note that `POST` and `PUT /:id` return
  409 with `existing_ingredient` on a duplicate name.
- `docs/features.md`, "Food Logging API for AI agents": the agent can correct and delete ingredients of the list;
  diary entries keep their own copy of name and values, so editing or deleting an ingredient does not change them.
  An ingredient line (or the existing "Ingredient Database" bullet): names are unique per user ignoring case, spacing
  and `ё`/`е` for creating and renaming on the website and through the API.
- `docs/architecture.md` line 51 (apidata domain description): add ingredient editing.

### Frontend
- `web/src/api/ingredients.ts`: `createIngredient` and `updateIngredient` recognise HTTP 409 and throw a typed error
  (e.g. a `DuplicateIngredientError` class with the parsed `existing_ingredient` as `Ingredient`) so that the page can
  tell a duplicate from other failures. Other non-OK responses throw an error that keeps the status; the raw response
  text is never shown to the user.
- `FoodList.tsx` + `AddIngredientModal.tsx` + `EditIngredientModal.tsx`:
  - On 409 the modal stays open with the user's input kept and shows an inline error inside the modal (above the
    buttons) with a localized message that names the existing ingredient, e.g. en_US `An ingredient named
    "{{name}}" already exists` (`{{name}}` = `existing_ingredient.name`, which may differ from the typed name in case or
    `ё`).
  - On any other failure of create/update the modal stays open and shows a generic localized save error (today the
    failure is silent).
  - The message disappears when the user edits the name field, when the modal is closed or reopened, and is replaced
    on the next submit; a successful retry closes the modal as now.
  - Delete behavior is unchanged.
- Styling: Tailwind utilities only, the existing inline error pattern
  (`p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700`, see `WeightGoalForm.tsx`), see the
  `frontend-style` skill.
- i18n: new keys in the `foodList` section (e.g. `foodList.duplicateName`, `foodList.saveError`) in all 4 locales
  (`en_US`, `uk_UA`, `ru_UA`, `bg_BG`), every string through `t()`. Suggested texts (the developer may polish them):
  - en_US: `An ingredient named "{{name}}" already exists` / `Could not save the ingredient. Please try again.`
  - ru_UA: `Продукт «{{name}}» уже есть в списке` / `Не удалось сохранить продукт. Попробуйте ещё раз.`
  - uk_UA: `Продукт «{{name}}» вже є у списку` / `Не вдалося зберегти продукт. Спробуйте ще раз.`
  - bg_BG: `Продукт „{{name}}“ вече съществува в списъка` / `Продуктът не можа да бъде запазен. Опитайте отново.`
- `npx tsc --noEmit -p tsconfig.app.json`, `npm run lint`, `npm run build` pass (from `web/`).

### DevOps
Regular release after acceptance (`deploy` skill). No migration, no new environment variables. The updated skill must
reach the user's agent after the server release (the client prints a hint against an older server, see above).

## Subtasks

Chain (full-stack): BA → Backend Dev → Backend Review → Frontend Dev → Frontend Review → Security Review → QA Review →
BA Acceptance → DevOps.

### Backend Developer
1. Make the limits and the name cleaning/normalization of `POST /api/v1/food` reusable by the ingredient service
   without duplicating them (approach: Q-D1).
2. Ingredient service: one duplicate-name check (common rule above) with a typed duplicate error carrying the existing
   ingredient; use it in `CreateIngredient`, `UpdateIngredient` and the new partial update.
3. Partial update in the ingredient service (approach: Q-D2): ownership check, validation of sent fields, duplicate
   check, merge, update through the existing repository method; typed errors for not found / validation.
4. `apidata` handler: `PUT /ingredients/:id`, `DELETE /ingredients/:id`, DTOs, error mapping (400/404/409/500 generic).
5. `ingredients` handler (JWT): map the duplicate error to 409 with `existing_ingredient` on `POST` and `PUT /:id`;
   nothing else changes.
6. Unit and integration tests (see Tests).
7. `kkal_client.py`: `edit-ingredient`, `delete-ingredient`, older-server hint.
8. `SKILL.md` changes, then `python3 skills/sync_dev_skill.py`.
9. `docs/api.md`, `docs/features.md`, `docs/architecture.md`.
10. `go vet ./...`, `go test ./...` green; `python3 -m py_compile` on the client; `status.md` → `backend-review`.

### Frontend Developer
1. `web/src/api/ingredients.ts`: typed 409 error with the existing ingredient; status-preserving errors for other
   failures; no raw server text for the user.
2. `FoodList.tsx`, `AddIngredientModal.tsx`, `EditIngredientModal.tsx`: inline duplicate and generic save errors as
   specified, input kept, message reset rules.
3. New `foodList.*` keys in all 4 locales.
4. `npx tsc --noEmit -p tsconfig.app.json`, `npm run lint`, `npm run build`; `status.md` → `frontend-review`.

### DevOps Engineer
Release after BA acceptance and the user's commit (workflow steps 7.5 and 8).

### Reviews
- `backend-code-reviewer`: Go code, tests, client script.
- `frontend-code-reviewer`: error handling in the API service and modals, i18n completeness, Tailwind patterns.
- `security-reviewer`: ownership (IDOR) on the v1 endpoints, no internal errors in responses (including the 409 path
  and the unique-index race), `existing_ingredient` never exposes another user's data, input limits, the skill's rule
  about explicit consent for deletion, no raw server text rendered in the UI, no real API key in the repository.
- `qa-engineer`: acceptance criteria coverage, boundary values, race detector, manual check of the Food List modals.

## Decisions (answers of the user, 2026-10-12)
- **Q1 → proposed default accepted.** Ambiguous calorie statement while logging → override for the entry only and offer
  in one line to update the list; edit the ingredient directly only when the user points at the product/list ("в
  справочнике", "на самом деле", "всегда").
- **Q2 → proposed default accepted.** After an ingredient correction refresh only the entries of the current
  conversation that the correction is clearly about; older entries untouched unless the user asks.
- **Q3 → fix in this task.** The website's `PUT /api/ingredients/:id` returns 409 on a duplicate name (same
  normalization as v1, response shape consistent with v1) and the Food List shows a localized message. Checked:
  `POST /api/ingredients` has the same 500 on a duplicate name → included. The task is now full-stack.

## Open Questions

### For the developers (technical, decided by the backend developer, checked in backend review)
- **Q-D1.** Where the shared limits and `cleanFoodName` / `normalizeFoodName` live: e.g. a small package used by both
  the `calorie` and `ingredient` services, or exported from one of them. Requirement: one implementation, no copy.
- **Q-D2.** Partial update as a new method of the ingredient service vs. extending `UpdateIngredientRequest`. The
  website's full-replace behavior must not change either way.
