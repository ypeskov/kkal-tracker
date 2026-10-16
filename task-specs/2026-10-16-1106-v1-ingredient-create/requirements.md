# Create an ingredient through the external API (`POST /api/v1/ingredients`) and the `add-ingredient` skill command

## Overview
The food logging skill (Settings → AI → Download skill: `SKILL.md` + `scripts/kkal_client.py`) can list, edit and
delete ingredients of the user's list. It can create an ingredient only as a side effect of logging a meal
(`new_ingredient` in `POST /api/v1/food`). To add a product without eating it (typically from a label photo:
"заведи продукт", "добавь в справочник"), the agent today logs a 1 g entry and deletes it. That is a hack: it briefly
changes the day total, leaves a create/delete trace in the diary and costs extra calls.

This task adds a dedicated endpoint for creating one ingredient, exposes it in the client as `add-ingredient` and
teaches the skill when to use it.

Business value: the user builds their ingredient list from labels in the same voice/chat conversation, without diary
side effects, and later meals match the product with its real values.

## Current State

### Routes, auth, rate limit
- `internal/server/server.go:205-207`: `v1Group := apiGroup.Group("/v1", v1RateLimiter, apiKeyMiddleware.RequireAPIKey)`
  (`api-v1` limiter: burst 20, then 1 req/s per IP), `apiDataHandler.RegisterRoutes(v1Group)`.
- `internal/handlers/apidata/handler.go:59-67` registers `GET /data`, `GET /ingredients`, `PUT /ingredients/:id`,
  `DELETE /ingredients/:id`, `POST /food`, `PUT /food/:id`, `DELETE /food/:id`. **`POST /ingredients` is free**: no
  clash with the v1 routes. The JWT route `POST /api/ingredients` (website) lives under another prefix and is not
  affected.
- Error format of the group: `echo.NewHTTPError(status, "message")` → `{"message": "..."}`; 409 bodies are built with
  `c.JSON`. Unexpected errors: logged with the injected logger, generic 500 message.

### Ingredient list row in the v1 API (`GET /api/v1/ingredients`)
`IngredientItem` (`internal/handlers/apidata/dto.go:14-19`) = embedded `IngredientInfo` + usage statistics. JSON keys in
this order:
```json
{"id": 94, "name": "Хлеб ржаной", "kcal_per_100g": 260, "fats": 1.2, "carbs": 48, "proteins": 8.5, "times_used": 3, "last_used": "2026-10-12"}
```
- `fats`, `carbs`, `proteins` are `omitempty` pointers: absent when null.
- `times_used` is always present; `last_used` (YYYY-MM-DD) is absent when null.
- The statistics come from `QueryGetUserIngredientsWithUsage`: an **exact name match** of `calorie_entries.food` and
  `user_ingredients.name`. A newly created ingredient therefore has `times_used: 0` and no `last_used`, **unless** the
  diary already has entries with exactly that name (e.g. one-off entries, or entries of a deleted ingredient). Then it
  "inherits" those statistics, see finding 4 of `task-specs/2026-10-12-1429-v1-ingredient-edit/requirements.md`.
- For comparison: `PUT /api/v1/ingredients/:id` returns the ingredient wrapped, `{"ingredient": IngredientInfo}`,
  without statistics.

### How `new_ingredient` in `POST /api/v1/food` is validated and deduplicated
- Handler (`apidata/handler.go:238-249`): `new_ingredient.kcal_per_100g` missing/null → 400
  `items[i]: new_ingredient.kcal_per_100g is required`.
- Calorie service `CreateMeal` (`internal/services/calorie/service.go:169-206`), limits and helpers from `internal/food`:
  - name `food.CleanName` (drops invisible characters, collapses whitespace, trims). Empty after cleaning → 400
    `...name is required`; more than `food.MaxNameLength` = 200 characters → 400. **The cleaned name is stored**;
  - `food.ValidKcalPer100g`: 0–900 inclusive;
  - `food.ValidNutrientPer100g` for `fats`, `carbs`, `proteins`: 0–100 inclusive, each optional;
  - duplicate: `food.NormalizeName` (clean + lower case + `ё`→`е`) against all of the user's ingredients → typed
    `DuplicateIngredientError` → handler 409 `{message, item_index, existing_ingredient}`.
- The CreateMeal duplicate check is its own map lookup in the calorie service. It does the same comparison as the
  ingredient service's `checkDuplicateName`.

### Ingredient service (`internal/services/ingredient/`)
- `CreateIngredient(*CreateIngredientRequest)` is used by the **website** `POST /api/ingredients`:
  - `validateCreateRequest` checks only `name != ""` and values `>= 0`. There are no upper limits and no name
    cleaning. This website behavior must not change in this task;
  - then `checkDuplicateName(userID, name, nil)` (normalized, all of the user's ingredients) → `DuplicateIngredientError{Existing}`;
  - then `ingredientRepo.CreateUserIngredient(...)` (`internal/repositories/ingredient.go:357`, insert + re-read);
  - if the write fails, `duplicateAfterWriteError(userID, name, 0, err)` runs. It reports a duplicate (409) **only
    when another ingredient has exactly the written name**, because the unique index `(user_id, name)` compares exact
    text. Otherwise the error stays internal (generic 500).
- `PatchIngredient` + `validatePatchRequest` (`service.go:277-306`) already produce v1-style messages for the same
  limits: `name must not be empty`, `name must be at most 200 characters`, `kcal_per_100g must be between 0 and 900`,
  `<field> must be between 0 and 100 (grams per 100 g)`. They are returned as `*ValidationError` → handler 400.
- `Servicer` interface (`interface.go`) is what the apidata handler uses.

### Database
`user_ingredients` already has `name`, `kcal_per_100g`, `fats`, `carbs`, `proteins` and the unique index
`idx_user_ingredients_unique (user_id, name)`; `CreateUserIngredient` and `QueryInsertUserIngredient` exist.
**No migration is needed.**

### Client (`skills/kkal-tracker-food-log/scripts/kkal_client.py`)
- `request()` returns `(status, payload)`. A non-JSON 2xx answer (an older server falling back to the website page)
  fails with "the server does not support ... (probably an older server version). Nothing was changed."
- `expect_ingredient_change()` (lines 109-115) is used by `edit-ingredient` / `delete-ingredient`:
  - 404/405 **without** the server's `{"message": "Ingredient not found"}` → `fail("HTTP <code>: the server probably
    does not support editing or deleting ingredients yet (an older server version). Nothing was changed.")`;
  - otherwise `expect()`: on a wrong status it prints `HTTP <code>` to stderr and the JSON body to stdout, then exits 1.
    So 400 shows the message naming the field, and 409 shows `existing_ingredient`.
- The module docstring lists the commands. Standard library only.
- On the current server (v5.14.0), `POST /api/v1/ingredients` hits a path that exists only for `GET`. Echo is
  expected to answer `405 Method Not Allowed` (possibly 404). The developer confirms the actual status before adding
  the route (Q-D3).

### Skill text (`skills/kkal-tracker-food-log/SKILL.md`)
- Frontmatter description (about 770 characters now, limit 1024 enforced by
  `internal/skill/skill_test.go:TestDescriptionLengthOfBothVariants` for both variants). It mentions "to fix or delete
  an ingredient in the user's list", but not adding one.
- "Production and dev": ingredients follow the copy whose list they came from (for edit/delete).
- "API basics": command list without a create command.
- Guidance that implies ingredients are created only through `add`:
  - "Fixing mistakes", line 188: "**The right food is not in the ingredient list** → `edit` cannot create ingredients.
    Delete the entry and `add` it again as a new food…";
  - "Correcting the ingredient list", line 217: deletion "cannot be undone through the API (the food would have to be
    created again, with a new ID)", which implies creating again means via `add`;
  - "Error handling", line 231: the older-server line covers "editing or deleting" only.
- The dev variant description (`internal/skill/skill.go:35-42`) treats edits/deletes of an entry or ingredient of the
  dev copy as the only exception to "dev only when named". Creating an ingredient is not an edit, so it follows the
  logging rule (dev only when the message names dev). No change is needed there.

### Skill "version"
- There is **no** version anywhere in the skill: not in the SKILL.md frontmatter, not in `kkal_client.py`, not in
  `internal/skill` or `internal/services/skill`.
- The app binary does not know its own release version at runtime: `deploy.sh` writes `version.txt` in the version
  commit after the image is built, and nothing is stamped via ldflags.
- The download zip is rendered **at request time** from the template embedded into the binary (`skills/embed.go`). So
  the first download after a release always contains the new `SKILL.md` and client. No bump is needed for the bundle
  to "pick up" the changes.
- Constraint: the SKILL.md frontmatter must not get unknown fields. claude.ai validates skill uploads and has already
  rejected such a field once.
- How item 5 is implemented: decision Q1 (option A), see "Decisions".

## Requirements

### Backend

#### `POST /api/v1/ingredients` — create one ingredient of the user's list
- Auth and rate limit: the existing `/api/v1` group (`X-API-Key`, `api-v1` limiter). The route is registered in
  `apidata.Handler.RegisterRoutes` next to the other ingredient routes; no new groups. JWT-only requests → 401.
- Request body (DTO in `internal/handlers/apidata/dto.go`, snake_case like `new_ingredient`):
  ```json
  {"name": "Хлебцы ржаные", "kcal_per_100g": 310, "fats": 2, "carbs": 60, "proteins": 10}
  ```
- Validation: the same rules and limits as `new_ingredient` in `POST /api/v1/food`, from `internal/food` (no copied
  constants). The messages are the field-level messages without the `items[i]: new_ingredient.` prefix, consistent
  with `PUT /api/v1/ingredients/:id`:
  - `name`: cleaned with `food.CleanName`, and the cleaned name is stored and returned. Missing, `null`, empty or only
    whitespace/invisible characters → 400 `name is required`. More than 200 characters after cleaning → 400
    `name must be at most 200 characters`;
  - `kcal_per_100g`: required. Missing or `null` → 400 `kcal_per_100g is required`. Outside 0–900 inclusive → 400
    `kcal_per_100g must be between 0 and 900`;
  - `fats`, `carbs`, `proteins`: optional. Missing or `null` → stored as null (unknown). Outside 0–100 inclusive → 400
    `<field> must be between 0 and 100 (grams per 100 g)`;
  - malformed JSON or wrong types → 400 `Invalid request body`;
  - unknown fields are ignored (as in the other v1 endpoints);
  - every 400 message names the field (and the limit where there is one) and never contains internal error text.
- Duplicate name: the same rule as `add`. The name is normalized with `food.NormalizeName` (case, spacing, invisible
  characters, `ё`/`е`) and compared with all ingredients **of the same user** → 409, nothing created:
  ```json
  {"message": "an ingredient with this name already exists", "existing_ingredient": {"id": 12, "name": "Хлебцы ржаные", "kcal_per_100g": 300, "fats": 2, "carbs": 58, "proteins": 11}}
  ```
  - The body is the existing `IngredientNameConflictResponse` (`existing_ingredient` in the `IngredientInfo` shape,
    no `item_index`), the same as `PUT /api/v1/ingredients/:id`.
  - Another user's ingredient with the same name is not a conflict and is not revealed.
  - **After a failed write**, only an ingredient with **exactly** the written name counts as a duplicate (409). This is
    the existing `duplicateAfterWriteError` behavior: the DB unique index is exact. Any other write failure → logged,
    generic 500. No database error text in any response.
- Success → **201** with the created ingredient exactly as a row of `GET /api/v1/ingredients`: an unwrapped
  `IngredientItem` (decision Q2):
  ```json
  {"id": 131, "name": "Хлебцы ржаные", "kcal_per_100g": 310, "fats": 2, "carbs": 60, "proteins": 10, "times_used": 0}
  ```
  - keys and their order as in the list row: `id`, `name`, `kcal_per_100g`, `fats`, `carbs`, `proteins`,
    `times_used`, `last_used`;
  - macros that were not sent are absent (as in the list);
  - `times_used` / `last_used` are the values the list itself would show for this ingredient. Normally that is
    `times_used: 0` and no `last_used`, but they are not hard-coded: if the diary already has entries with exactly
    this name, the list shows them, and so does the 201;
  - the row returned equals the row of this ingredient in a following `GET /api/v1/ingredients`.
- Side effects: exactly one new row in `user_ingredients`. **No diary entry is created**. The day total and the
  results of `GET /api/v1/data` do not change. The ingredient is immediately usable as `ingredient_id` in
  `POST /api/v1/food` and visible in the website's `GET /api/ingredients` (and on the Food List page after a reload;
  the page caches the list in `sessionStorage`, an existing behavior).
- Logging: Info `User ingredient created` with `user_id` and `ingredient_id`. Unexpected errors: Error with `error` and
  `user_id`, client gets 500 `Failed to create ingredient`, never `err.Error()`.

#### Implementation constraints
- Reuse the ingredient service: the create path (`checkDuplicateName` → `CreateUserIngredient` →
  `duplicateAfterWriteError`) has one implementation. It is shared by the website `CreateIngredient` and the new v1
  create. The v1 validation (cleaning, limits, messages) is shared with `validatePatchRequest`, not copied. Approach:
  Q-D1.
- The website `POST /api/ingredients` keeps its behavior exactly: `>= 0` validation without upper limits, the name
  stored as sent, the 409 for duplicates, the 201 body `models.UserIngredient`.
- `POST /api/v1/food` is unchanged.
- No new SQL is expected. If any is needed (e.g. for the usage statistics of one ingredient, Q-D2), it goes into
  `internal/repositories/queries.go`. **No migration.**
- `Servicer` gets the new method; mocks/fakes of the interface (if any) are updated.

#### Tests
- Service unit tests (`internal/services/ingredient/service_test.go`):
  - v1 create: name cleaning; every limit with boundary values (kcal 0 and 900 pass, -0.1 and 900.1 fail; each macro 0
    and 100 pass, -0.1 and 100.1 fail); missing name / name of only invisible characters; 200 vs 201 characters;
  - duplicate (exact, case, spacing, invisible character, `ё`/`е`) → `DuplicateIngredientError` carrying the existing
    ingredient, nothing written;
  - write failure with an exact-name ingredient present → duplicate; write failure without one → internal error (not
    a duplicate);
  - the website `CreateIngredient` still accepts e.g. 1000 kcal and an uncleaned name (unchanged behavior).
- Integration tests on the fully wired server (`internal/server`, e.g. `ingredient_api_test.go`, helpers `signUp`,
  `listIngredients`, `expectError`):
  - create OK: 201 body equals the expected row (exact keys and order, macros absent when not sent); the ingredient is in
    `GET /api/v1/ingredients` with the same values and in `GET /api/ingredients`; the cleaned name is stored;
  - duplicate name (exact, `гречка` / ` Гречка ` / `ГРЕЧКА` vs `Гречка`, `ежик` vs `Ёжик`) → 409 with
    `existing_ingredient`, and the ingredient count is unchanged;
  - the same name in another user's list → 201;
  - out-of-range and invalid values → 400 with the field named, nothing created: kcal -1 / 900.1 / missing / `null`;
    each macro -0.1 / 100.1; name missing / empty / whitespace only / 201 characters; malformed JSON; `"kcal_per_100g":
    "abc"`;
  - the diary is untouched: the entries of `GET /api/v1/data` for today, and the day total, are the same before and
    after a successful create and after a rejected one;
  - the returned `id` works as `ingredient_id` in `POST /api/v1/food` (calories from the created values);
  - 401 without a key, with an invalid key, and with a JWT only;
  - an ingredient created with the exact name of existing one-off diary entries shows the same `times_used` /
    `last_used` in the 201 as in the list.
- `go vet ./...`, `go test ./...` green (after `make build-frontend`), `go test -race` on the changed packages.

### Skill and client (part of the backend subtask)

#### `kkal_client.py`
- New subcommand:
  ```
  python scripts/kkal_client.py add-ingredient --name NAME --kcal N [--fats N] [--carbs N] [--proteins N]
  ```
  - `--name` and `--kcal` are required. If either is missing, argparse fails (non-zero exit) and no request is sent.
    `--kcal`, `--fats`, `--carbs`, `--proteins` are `type=float`;
  - it sends `POST /ingredients` with `name`, `kcal_per_100g` and only the macros that were passed. A value of `0` is
    sent, not dropped;
  - it expects 201 and prints the response JSON (`print_json`), exit 0;
  - help texts: "add an ingredient to the list without logging anything to the diary"; macros are grams per 100 g.
- Error handling in the style of `edit-ingredient`:
  - 400 → `HTTP 400` on stderr, the JSON body (message naming the field) on stdout, exit 1;
  - 409 → `HTTP 409`, the JSON body with `existing_ingredient`, exit 1;
  - 404 or 405 (an older server without the route; `POST` never returns `Ingredient not found`) → exit 1 with a clear
    message, e.g. `HTTP 405: the server probably does not support adding ingredients yet (an older server version).
    Nothing was created.`;
  - a non-JSON answer, 401, 429, 5xx, network errors → as for every other command.
  - Reuse/parametrize `expect_ingredient_change` rather than copy it. The existing messages of `edit-ingredient` /
    `delete-ingredient` stay correct for those commands.
- Module docstring: add `add-ingredient --name --kcal  add an ingredient to the list without logging it`. Standard
  library only. `API_KEY = "API_KEY_HERE"` stays the only literal placeholder (the render checks it exactly once).
- Client tests in `internal/server/skill_client_test.go` (they run `python3` against the wired test server; QA
  confirms they ran, not skipped):
  - `add-ingredient --name "Хлебцы ржаные" --kcal 310 --fats 0 --proteins 10` → exactly one request
    `POST /api/v1/ingredients` with exactly `{"name", "kcal_per_100g", "fats": 0, "proteins"}`; exit 0; stdout is the
    created row; the ingredient is in the list;
  - without `--kcal` / without `--name` → non-zero exit, no request;
  - duplicate → exit 1, `HTTP 409` on stderr, stdout JSON with `existing_ingredient`;
  - `--kcal 1000` → exit 1, `HTTP 400`, the message names `kcal_per_100g`;
  - older server: 405 (and 404) from a server without the route → exit 1, the "does not support adding ingredients
    yet" message, "Nothing was created". Simulate it per Q-D3.

#### `SKILL.md` (English instructions, Russian examples, existing style)
- **Frontmatter description**: a minimal change that keeps it ≤ 1024 characters (the test enforces it). Recommended:
  "to fix or delete an ingredient in the user's list" → "to add an ingredient to the user's list without logging it
  (e.g. from a label photo), to fix or delete one". No new frontmatter fields.
- **"Production and dev"**: adding an ingredient is routed like logging food. Each message is routed on its own, and
  it goes to dev only when that message names the dev/test tracker.
- **"API basics"** command list: a line such as
  `python scripts/kkal_client.py add-ingredient --name "Хлебцы ржаные" --kcal 310 --fats 2 --carbs 60 --proteins 10   # add an ingredient without logging it`.
- **"Correcting the ingredient list"**, a new part "Add an ingredient without logging it" (wording adapted to the
  section):
  - When: the user asks to add a product to the list without eating it now: "заведи продукт", "добавь в справочник",
    "внеси в список продуктов", "сохрани этот продукт", a label photo with such a request.
  - Label photo intent (decision Q3):
    - photo + an eating/logging statement ("съел", "запиши 50 г", "на обед") → log it with `add` as today
      (`new_ingredient`, or a per-entry override for a known food);
    - photo + "заведи продукт" / "добавь в справочник" / "сохрани продукт" → `add-ingredient`;
    - a photo alone, or an unclear request → ask one short question: "записать в дневник или только добавить в
      справочник?", and act only on the answer.
  - Values per 100 g from the label. If the label gives values per portion, convert them to per 100 g. If the label is
    unreadable or the numbers are missing, ask; do not estimate silently when the user brought a label. Without a label
    the agent may estimate as in step 4 and must say so.
  - Name: short, in the user's language, the way the user calls it (same rule as step 4).
  - Before creating, look at the ingredient list already fetched in this conversation (`ingredients`). If a product
    with the same meaning is already there under another name (e.g. "Ржаной хлеб" vs "Хлеб ржаной"), tell the user
    and ask whether to update that one (`edit-ingredient`, after the user agrees) or to add a separate one (decision
    Q4 ii). The server catches only name variants (case, spacing, `ё`/`е`), not synonyms.
  - **Never** create an ingredient by logging a fake entry (e.g. 1 g) and deleting it.
  - No diary entry is made, so do not run `totals` and do not show the meal/day table.
  - Reply: the created ingredient with its values (name, kcal per 100 g, Ж/У/Б per 100 g, or that macros are unknown),
    and that nothing was logged to the diary.
- **"Fixing mistakes"**, "The right food is not in the ingredient list" (decision Q4 i): replace "Delete the entry and `add` it
  again" with `add-ingredient`, then `edit ENTRY_ID --ingredient-id <new id>`. This keeps the entry, its weight and
  time, and the server recalculates the calories.
- **Deletion paragraph**: "cannot be undone through the API (the food would have to be created again, with a new ID)"
  → it can be created again with `add-ingredient`, with a new ID; old entries are not relinked.
- Remove or reword any other sentence that implies ingredients can be created only through `add`. Step 4 "New foods"
  stays as it is: it is about logging.
- **Limits line** (step 5): the same calorie and macro limits apply to `edit`, `edit-ingredient` **and
  `add-ingredient`**.
- **"Error handling"**:
  - `409` on `add-ingredient` → the product is already in the list (`existing_ingredient`), nothing was created. Tell
    the user its current values. If they differ from the label, offer to update it with `edit-ingredient`, and do it
    only after the user agrees.
  - `400` on `add-ingredient` → the message names the field and the limit; nothing was created.
  - The older-server line covers adding as well: tell the user, do not retry, and **do not fall back** to the
    log-and-delete workaround.
- Item 5 (version): see the next subsection.

#### Skill version (item 5, decision Q1 = A)
- `kkal_client.py` gets one constant `SKILL_VERSION = "2026-10-16"` (the date of the last template change). It
  identifies the whole bundle (SKILL.md + client) and is bumped in every later change of either template file.
- A `version` subcommand prints it and sends no request (it works without a configured key).
- SKILL.md body (not the frontmatter), "API basics": when the user asks which skill version is installed, run
  `version`. The value itself is not repeated in SKILL.md, so there is nothing to keep in sync.
- Nothing about the version enters the frontmatter. `skill.Render`'s anchors (`DEFAULT_BASE_URL`, the `API_KEY` line)
  are unaffected, and the zip still equals the template byte for byte apart from the substituted values.
- Client test: `version` prints the constant, exit 0, no request.
- Module docstring lists `version`.

### Docs (backend developer)
- `docs/api.md`, `/api/v1` list: `POST /api/v1/ingredients`: `{name, kcal_per_100g, fats?, carbs?, proteins?}`, the
  same limits and name cleaning as `new_ingredient`, 201 with the ingredient (list-row shape), 409
  `{message, existing_ingredient}` on a duplicate name (ignoring case, spacing, invisible characters, `ё`/`е`), no diary
  entry.
- `docs/features.md`, "Food Logging API for AI agents": the agent can add an ingredient to the list without logging
  it (e.g. from a label photo), and the skill command `add-ingredient`. The "Ingredient Database" uniqueness bullet also
  names creating through the API. The skill carries `SKILL_VERSION` (shown by the client's `version` command), and
  every download renders the skill from the running release.
- `docs/architecture.md` line 52: the apidata domain covers "ingredient list and its editing (create, update,
  delete)".
- `README.md`, "External API" table: add `POST /api/v1/ingredients`. Also add the missing rows
  `PUT /api/v1/ingredients/:id` and `DELETE /api/v1/ingredients/:id`, a gap from the 2026-10-12 task.
- `AGENTS.md`: no change expected.

### Frontend
None. The website `POST /api/ingredients` and the Food List page are unchanged.

### DevOps
A regular release after acceptance (`deploy` skill). No migration, no new environment variables. After the release,
the user downloads the skill again (Settings → AI) to get `add-ingredient`. Installed copies keep working and show
the "older server" message only against an older server.

## Subtasks

Chain (backend-only, no frontend changes): BA → Backend Dev → Backend Review → Security Review → QA Review →
BA Acceptance → DevOps.

### Backend Developer
1. Read the current Echo v5 docs for the routing/405 behavior (AGENTS.md library rule). Confirm what the current code
   answers to `POST /api/v1/ingredients` before adding the route, and record it in `status.md` (Q-D3).
2. Ingredient service: the v1 create method with the shared validation and the shared create path (Q-D1); the
   `Servicer` interface.
3. `apidata`: request DTO, `POST /ingredients` handler, error mapping (400 / 409 / 500 generic), 201 body = the list
   row with real usage statistics (Q-D2).
4. Service and integration tests (see Tests).
5. `kkal_client.py`: `add-ingredient`, the older-server message, `SKILL_VERSION` and `version`, docstring.
6. `SKILL.md` changes (description, routing, API basics, the new part in "Correcting the ingredient list" with the
   label-photo rule and the synonym check, "Fixing mistakes", deletion paragraph, limits line, error handling, the
   `version` line).
7. Client tests in `skill_client_test.go`; `TestDescriptionLengthOfBothVariants` still passes.
8. Docs: `docs/api.md`, `docs/features.md`, `docs/architecture.md`, `README.md`.
9. `go vet ./...`, `go test ./...`, `go test -race` (changed packages) green; `python3 -m py_compile` on the client;
   `status.md` → `backend-review`.

### Frontend Developer
None.

### DevOps Engineer
Release after BA acceptance and the user's commit (workflow steps 7.5 and 8). Post-release check: `POST
/api/v1/ingredients` without a key → 401; the downloaded zip's client has `add-ingredient`. Do not print keys.

### Reviews
- `backend-code-reviewer`: service reuse (no copied validation or duplicate logic), website create unchanged, handler
  error mapping, client and SKILL.md.
- `security-reviewer`: the new write endpoint is scoped to the key's user (no `user_id` from the body); the 409 reveals
  only the caller's own ingredient; no internal errors (including the unique-index race path); input limits (name
  length, numeric ranges); the rate limit applies; no key in the repository or logs.
- `qa-engineer`: acceptance criteria coverage, boundary values, the race detector, the client tests ran with python3
  (not skipped).

## Decisions (answers of the user, 2026-10-16)
- **Q1, item 5 "bump the skill version" → option A.** `SKILL_VERSION = "2026-10-16"` constant in `kkal_client.py`,
  bumped in every later change of the template (SKILL.md or client); a `version` subcommand prints it without a
  request; the SKILL.md body (API basics) tells the agent to run `version` when asked which version is installed. No
  frontmatter field, and the value is not repeated in SKILL.md. Background: the zip is rendered from the running
  release at every download, so the release itself ships the new files; the constant makes an outdated installed copy
  visible.
- **Q2, 201 body → (a).** Exactly a row of `GET /api/v1/ingredients`, unwrapped: `id, name, kcal_per_100g, fats?,
  carbs?, proteins?, times_used, last_used?`, with the real usage statistics the list shows (normally
  `times_used: 0`; inherited when diary entries with exactly that name already exist).
- **Q3, label photo intent → ask when unclear.** Photo + an eating/logging verb → `add`; photo + "заведи продукт" /
  "добавь в справочник" → `add-ingredient`; photo alone or an unclear request → ask "записать в дневник или только
  добавить в справочник?".
- **Q4 → both.** (i) The right food missing from the list after logging → `add-ingredient` + `edit ENTRY_ID
  --ingredient-id <new id>` instead of delete + re-add (keeps the entry and its time). (ii) Before `add-ingredient`
  the agent checks the fetched list for the same product under another name and asks whether to update that one or
  add a separate one.

## Open Questions

### For the developers (decide during implementation, record in `status.md`)
- **Q-D1. Service shape.**
  - Recommended: a new method, e.g. `AddIngredient(*AddIngredientRequest)` (pointer `KcalPer100g`, so "missing" is
    visible). It validates with helpers extracted from `validatePatchRequest` (same messages, same `internal/food`
    calls), then calls an internal create helper shared with `CreateIngredient` (`checkDuplicateName` →
    `CreateUserIngredient` → `duplicateAfterWriteError`).
  - Alternative: a validation-mode field on `CreateIngredientRequest`.
  - Either way: one create path, website behavior unchanged, typed errors (`ValidationError`,
    `DuplicateIngredientError`).
- **Q-D2. Statistics for the 201.**
  - Reuse `GetIngredientsWithUsage` and pick the created ID. That is simple, and lists are small.
  - Or add a single-ingredient usage query in `queries.go`.
  - Requirement: the 201 row equals the list row.
- **Q-D3. Older-server test.**
  - First confirm the actual status of `POST /api/v1/ingredients` on the current code (expected 405, since
    `GET /ingredients` exists).
  - Then simulate it in `skill_client_test.go`, e.g. a wrapper handler that answers 405 `{"message": "Method Not
    Allowed"}` (and a 404 case) for that method and path, instead of forwarding.
  - Do not put test files into `skills/kkal-tracker-food-log/`.
