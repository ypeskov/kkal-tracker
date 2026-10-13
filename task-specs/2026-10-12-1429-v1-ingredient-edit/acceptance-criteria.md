# Acceptance Criteria: Edit and delete ingredients through the external API (`/api/v1`), duplicate-name protection on the website

Ingredient `A` below: `{"name": "Хлеб", "kcal_per_100g": 250, "fats": 1, "carbs": 50, "proteins": 8}` of user U1.

BA acceptance 2026-10-12: verified against the code and the test results reported by QA (go vet, go test 90 tests,
`-race`, shuffled runs, coverage 84% of the changed packages; tsc, lint, build; client against a local mock server;
dev copy equal to the `sync_dev_skill.py` output). Main test files: `internal/server/ingredient_api_test.go`,
`internal/services/ingredient/service_test.go`, `internal/food/food_test.go`. Criteria marked **not verified** need a
browser or the live dev server: the user chose to release without these manual checks (2026-10-12).

## Backend

### `PUT /api/v1/ingredients/:id`
- [x] Valid `X-API-Key` of U1, `{"kcal_per_100g": 260}` on `A` → 200
      `{"ingredient": {"id": A, "name": "Хлеб", "kcal_per_100g": 260, "fats": 1, "carbs": 50, "proteins": 8}}`;
      `GET /api/v1/ingredients` and the website's `GET /api/ingredients` show 260
- [x] All five fields sent at once → all five changed, 200
- [x] Fields not sent keep their values (incl. macros); `"fats": null` is treated as not sent
- [x] `{"name": "  Хлеб​   ржаной "}` → stored and returned as `Хлеб ржаной`
- [x] Boundaries pass: `kcal_per_100g` 0 and 900, `fats`/`carbs`/`proteins` 0 and 100 → 200
- [x] `kcal_per_100g` -1 or 900.1 → 400, message names `kcal_per_100g` and the limit; ingredient unchanged
- [x] `fats` / `carbs` / `proteins` -0.1 or 100.1 → 400, message names the field; ingredient unchanged
- [x] `name` empty or only whitespace/invisible characters → 400; 201-character name → 400; 200 characters → 200
- [x] `{}` (or only `null` values) → 400 "nothing to change" with the accepted field names
- [x] Malformed JSON or `"kcal_per_100g": "abc"` → 400 `Invalid request body`
- [x] `/api/v1/ingredients/abc` → 400 `Invalid ingredient ID`
- [x] Non-existing ID → 404 `Ingredient not found`
- [x] Ingredient of user U2 with U1's key → 404 `Ingredient not found`; U2's ingredient unchanged
- [x] U1 has `Гречка` (id G) and `A`; renaming `A` to `гречка`, ` Гречка `, `ГРЕЧКА` → 409
      `{"message": "an ingredient with this name already exists", "existing_ingredient": {"id": G, "name": "Гречка",
      "kcal_per_100g": ...}}`; `A` unchanged
- [x] U1 has `Ёжик`; renaming another ingredient to `ежик` → 409 with `Ёжик` as `existing_ingredient`
- [x] Renaming `A` to a variant of its own name (`хлеб`, `ХЛЕБ`) → 200, no 409
- [x] Renaming to a name that exists only in U2's list → 200 (no cross-user conflict, nothing about U2 revealed)
- [x] Stored values outside the new limits are kept when not sent: an ingredient created on the website with
      1000 kcal can be renamed via `{"name": ...}` → 200, `kcal_per_100g` stays 1000
- [x] Without `X-API-Key` or with an invalid/revoked key → 401
- [x] JWT-only request (no API key) to `/api/v1/ingredients/:id` → 401 (the route is API-key only)

### `DELETE /api/v1/ingredients/:id`
- [x] Own ingredient → 204, empty body; it is gone from `GET /api/v1/ingredients` and `GET /api/ingredients`
- [x] Same ID again → 404 `Ingredient not found`
- [x] U2's ingredient with U1's key → 404; U2's ingredient still exists
- [x] `/api/v1/ingredients/abc` → 400 `Invalid ingredient ID`
- [x] Without / with an invalid key → 401

### Website API (`/api/ingredients`, JWT)
- [x] U1 has `Гречка` (id G). `POST /api/ingredients {"name": "Гречка", "kcalPer100g": 330}` (exact duplicate, today a
      500) → 409 `{"message": "an ingredient with this name already exists", "existing_ingredient": {"id": G,
      "name": "Гречка", "kcalPer100g": ..., ...}}` (`models.UserIngredient` shape); no new ingredient
- [x] `POST` with `гречка`, ` Гречка `, `ГРЕЧКА` → 409 with G; U1 has `Ёжик` → `POST` `ежик` → 409 with `Ёжик`
- [x] `PUT /api/ingredients/A {"name": "Гречка", "kcalPer100g": 250}` (today a 500) and `{"name": "гречка", ...}` →
      409 with G; `A` unchanged
- [x] `PUT /api/ingredients/A {"name": "ХЛЕБ", "kcalPer100g": 260}` (own-name variant) → 200, saved
- [x] Pre-existing near-duplicates `Рис` (R1) and `рис` (R2) inserted directly (e.g. through the repository in a
      test): `PUT /api/ingredients/R2 {"name": "рис", "kcalPer100g": 120}` → 200 (normalized name unchanged);
      `PUT` on `A` with name `РИС` → 409
      (integration test creates `рис` through the diary path `POST /api/calories`; service test inserts both directly)
- [x] U2 has `Овсянка`, U1 does not: U1 `POST` / `PUT` with `Овсянка` → 201 / 200 (no cross-user conflict)
- [x] Ordinary create and update without a duplicate → 201 / 200 with the same body shape as before; full-replace
      semantics of `PUT` unchanged (macros omitted → cleared); validation unchanged (e.g. `kcalPer100g` 1000 still
      accepted, negative → 400); foreign/missing ingredient on `PUT` → 404; `DELETE` unchanged

### Errors and logging
- [x] No 4xx/5xx response body of the new or changed endpoints contains internal error text (SQL, driver, Go error
      strings, unique-index messages); unexpected errors return 500 with a generic message and are logged with the
      injected logger (`user_id`, `ingredient_id` where applicable)

### Code and tests
- [x] The limits (900 / 100 / 200) and the name cleaning/normalization exist in one place, used by both
      `POST /api/v1/food` and the new update (no copied constants or functions) — `internal/food/food.go`
- [x] The duplicate-name check is one implementation in the ingredient service, used by `POST /api/ingredients`,
      `PUT /api/ingredients/:id` and `PUT /api/v1/ingredients/:id` — `checkDuplicateName`
- [x] Service unit tests cover merge, every limit with boundary values, cleaning, nothing-to-change, duplicate
      (case / spacing / invisible characters / `ё`) for create, full and partial update, own-name variant,
      pre-existing near-duplicate edit, not found
- [x] Integration tests on the fully wired server (`internal/server`) cover the endpoint criteria above for both APIs,
      including the two-user cases
- [x] `go vet ./...` and `go test ./...` pass; `go test -race ./internal/...` passes

## Frontend
- [ ] Food List → Add New → name `Гречка` (U1 already has `Гречка`) → Save: the modal stays open, inputs are kept, an
      inline red message inside the modal says the ingredient already exists and names `Гречка`; no new row in the list
      — **not verified** — user chose to release without manual verification (2026-10-12). Implemented in code
      (`DuplicateIngredientError` → `foodList.duplicateName` with `existing.name`, modal closes only `onSuccess`)
- [ ] Same for `гречка` / `ГРЕЧКА`: the message names the existing `Гречка` (the existing name, not the typed one)
      — **not verified** — user chose to release without manual verification (2026-10-12). Code uses
      `existing_ingredient.name`
- [ ] Food List → click `Хлеб` → rename to `Гречка` → Update: modal stays open, inputs kept, same message; the list is
      unchanged — **not verified** — user chose to release without manual verification (2026-10-12)
- [ ] Rename `Хлеб` → `ХЛЕБ` → Update: saved, modal closes as before — **not verified** — user chose to release
      without manual verification (2026-10-12); the backend side is covered by integration tests
- [ ] After the duplicate message, changing the name to a free one and saving → saved, modal closes; editing the name
      field hides the message; closing and reopening the modal shows no stale message — **not verified** — user chose
      to release without manual verification (2026-10-12). Code: `mutation.reset()` on open, close and name edit
- [x] Any other create/update failure (e.g. 500, network error) → the modal stays open and shows the generic localized
      save error; the raw server response text is never displayed (verified in code: `toSaveError` never reads the
      response text, non-duplicate errors map to `foodList.saveError`, the modal closes only on success)
- [x] Both messages are translated in `en_US`, `uk_UA`, `ru_UA`, `bg_BG` (switch the language and check each); no
      hard-coded user-facing strings; keys exist in all 4 locale files (verified in the locale files and code, both
      messages go through `t()`; not viewed in a browser)
- [ ] Styling uses Tailwind utilities only, consistent with the existing inline error pattern; the modal layout on
      mobile and desktop is not broken — **not verified** (layout) — user chose to release without manual
      verification (2026-10-12). Tailwind-only and the `WeightGoalForm` error pattern are confirmed in code
- [ ] Delete from the Edit modal works as before — **not verified** — user chose to release without manual
      verification (2026-10-12). The delete code path is unchanged
- [x] From `web/`: `npx tsc --noEmit -p tsconfig.app.json`, `npm run lint`, `npm run build` pass

## Diary consistency (findings documented in requirements.md)
- [x] Log `A` (100 g → 250 kcal) via `POST /api/v1/food`, then `PUT /api/v1/ingredients/A {"kcal_per_100g": 300,
      "name": "Хлеб ржаной"}` → the entry in `GET /api/v1/data` is still `Хлеб`, 250 kcal/100 g, 250 kcal; the day total
      is unchanged
- [x] Then `PUT /api/v1/food/<entry> {"ingredient_id": A}` → the entry becomes `Хлеб ржаной`, 300 kcal/100 g, 300 kcal,
      same weight and time
- [x] `DELETE /api/v1/ingredients/A` → the diary entry still exists with its values and counts toward the day total;
      `PUT /api/v1/food/<entry> {"weight": 50}` still works
- [x] After the delete, `POST /api/v1/food` with `{"ingredient_id": A, ...}` → 400 `items[0]: ingredient A not found`,
      nothing saved

## Skill and client
- [x] `kkal_client.py edit-ingredient 94 --kcal 260` sends `PUT /api/v1/ingredients/94` with `{"kcal_per_100g": 260}`
      only and prints the JSON response; `--name`, `--fats`, `--carbs`, `--proteins` map to their fields
- [x] `kkal_client.py edit-ingredient 94` without options → exits 1 with "Nothing to change ..." and sends no request
- [x] `kkal_client.py delete-ingredient 94` → `DELETE /api/v1/ingredients/94`, prints `Ingredient 94 deleted` on 204
- [x] On 404 `Ingredient not found`, 409, 400 the client prints `HTTP <code>` and the JSON body and exits 1 (like `edit`)
- [x] Against a server without the new routes (404/405 without `Ingredient not found`) the client says the server
      probably does not support ingredient editing yet and that nothing was changed
- [x] The client stays standard-library only; module docstring lists the new commands; `API_KEY` is the
      `XXXXXXXXX` placeholder
- [x] `SKILL.md`: the description mentions fixing/removing ingredients of the list; the command list has both commands;
      ingredient IDs are per server (prod/dev routing); 409/404/400/older-server handling for the new commands; the
      sentence "Ingredients created by mistake cannot be removed through this API…" is gone; English, existing style
- [x] `SKILL.md`, Q1 as decided: the ingredient is edited directly only when the user points at the product/list ("в
      справочнике", "на самом деле", "всегда"); otherwise the value is overridden for the entry only and the reply
      offers in one line to update the list, which is changed only after the user agrees
- [x] `SKILL.md`, Q2 as decided: editing does not change logged entries; only entries of the current conversation that
      the correction is clearly about are refreshed (`edit ENTRY_ID --ingredient-id <same id>`) and reported; older
      entries only on the user's request
- [x] `SKILL.md`: rename resets usage statistics; deletion only on an explicit request, entries stay, no undo
- [x] `skills/kkal-tracker-food-log-dev/` regenerated with `python3 skills/sync_dev_skill.py`: matches the production
      copy except the generated differences (name, description, banner, base URL); `api_key` files untouched and not
      in git (`skills/*/api_key` in `.gitignore`; the dev description in `sync_dev_skill.py` also names ingredients as
      the dev-routing exception)

## Docs
- [x] `docs/api.md` lists `PUT /api/v1/ingredients/:id` (body, partial semantics, 409) and
      `DELETE /api/v1/ingredients/:id`, and the 409 of `POST` / `PUT /api/ingredients`
- [x] `docs/features.md` describes ingredient editing/deletion through the API, that diary entries keep their own copy
      of the values, and that ingredient names are unique per user ignoring case, spacing and `ё`/`е`
- [x] `docs/architecture.md` apidata domain description mentions ingredient editing

## Integration
- [ ] End to end with the dev skill against the dev server: log a meal with a new ingredient, correct the ingredient's
      calories with `edit-ingredient`, log it again → the new entry uses the corrected value, the old one is unchanged;
      `delete-ingredient` → gone from `ingredients`, diary entries still listed by `entries` — **not verified** — user
      chose to release without manual verification (2026-10-12); the dev API key was invalid on the dev server. The same
      flow is covered at the API level by `TestIngredientAPIKeepsDiaryEntries`
- [ ] An ingredient renamed by the agent shows the new name on the Food List page after reload; trying to create the
      old agent-created name again on the website while it still exists → the duplicate message — **not verified** —
      user chose to release without manual verification (2026-10-12). The API side (v1 rename visible in
      `GET /api/ingredients`, website 409) is covered by integration tests
