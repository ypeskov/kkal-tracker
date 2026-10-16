# Acceptance Criteria: Create an ingredient through the external API (`POST /api/v1/ingredients`) and the `add-ingredient` skill command

User U1 has `Гречка` (id G, 330 kcal) and `Ёжик`. User U2 has `Овсянка`. Requests use U1's valid `X-API-Key` unless
stated otherwise. The user's decisions Q1–Q4 (2026-10-16) are recorded in `requirements.md`, "Decisions".

## Backend: `POST /api/v1/ingredients`

### Success
- [x] `{"name": "Хлебцы ржаные", "kcal_per_100g": 310, "fats": 2, "carbs": 60, "proteins": 10}` → 201
      `{"id": N, "name": "Хлебцы ржаные", "kcal_per_100g": 310, "fats": 2, "carbs": 60, "proteins": 10, "times_used": 0}`
      (unwrapped list row, keys exactly in this order, no `last_used`)
- [x] `GET /api/v1/ingredients` then contains a row equal to the 201 body; the website's `GET /api/ingredients` lists
      the ingredient with the same name and values
- [x] `{"name": "Соус", "kcal_per_100g": 120}` → 201 without `fats`, `carbs`, `proteins` keys; stored as null (the list
      row has no macro keys either)
- [x] `"fats": null` (or `carbs`/`proteins` null) is treated as not sent → stored as null, 201
- [x] `{"name": "  Хлебцы​   с отрубями ", ...}` (extra spaces, a zero-width space) → stored and returned as
      `Хлебцы с отрубями`
- [x] Boundaries pass: `kcal_per_100g` 0 and 900; `fats`/`carbs`/`proteins` 0 and 100; a 200-character name → 201
- [x] The returned `id` used as `ingredient_id` in `POST /api/v1/food` with weight 100 → 201, entry calories 310
- [x] When U1 already has one-off diary entries named exactly `Кофе раф` (2 entries, last on D), creating
      `Кофе раф` → 201 with `times_used: 2`, `last_used: D`, equal to its row in the list

### Validation (each → 400, `{"message": ...}`, nothing created: U1's ingredient count unchanged)
- [x] `kcal_per_100g` -1 or 900.1 → `kcal_per_100g must be between 0 and 900`
- [x] `kcal_per_100g` missing or `null` → `kcal_per_100g is required`
- [x] `fats`, `carbs`, `proteins` -0.1 or 100.1 → `<field> must be between 0 and 100 (grams per 100 g)`, the field named
- [x] `name` missing, `null`, `""`, `"   "`, or only invisible characters → `name is required`
- [x] 201-character name (after cleaning) → `name must be at most 200 characters`
- [x] Malformed JSON, `"kcal_per_100g": "abc"`, `"name": 5` → `Invalid request body`
- [x] No 400 message contains internal error text

### Duplicate names (each → 409, nothing created)
- [x] `{"name": "Гречка", "kcal_per_100g": 340}` → 409
      `{"message": "an ingredient with this name already exists", "existing_ingredient": {"id": G, "name": "Гречка",
      "kcal_per_100g": 330, ...}}` (`IngredientInfo` shape, no `item_index`); G unchanged
- [x] `гречка`, ` Гречка `, `ГРЕЧКА`, `Греч​ка` with a zero-width space → 409 with G
- [x] `ежик` → 409 with `Ёжик` as `existing_ingredient`
- [x] `Овсянка` (exists only in U2's list) → 201 for U1; U2's list unchanged; the response reveals nothing about U2
- [x] Write failure path (service test): another ingredient with exactly the written name → `DuplicateIngredientError`
      (409); a write failure while only a near-duplicate exists, or with no matching ingredient → internal error →
      500 `Failed to create ingredient`, logged, no DB text in the body

### Diary untouched
- [x] U1 logs a meal today. Then a successful `POST /api/v1/ingredients` and a rejected one (400, 409). After them,
      `GET /api/v1/data?type=food&from=<today>&to=<today>` returns the same entries, and the day total (from `totals`
      or the sum of the entries) is unchanged. No diary entry with the new name exists

### Auth and routing
- [x] Without `X-API-Key`, with an invalid or revoked key → 401; with only a JWT `Authorization` header → 401
- [x] The route is in the `/api/v1` group (the `api-v1` rate limiter applies). Existing `GET /api/v1/ingredients` and
      `PUT`/`DELETE /api/v1/ingredients/:id` behave as before (their tests still pass)
- [x] The website `POST /api/ingredients` is unchanged: `kcalPer100g` 1000 is still accepted (201), the name is stored as
      sent, a duplicate still gives 409 with `existing_ingredient` in the `models.UserIngredient` shape

### Code
- [x] One create path in the ingredient service (`checkDuplicateName` → `CreateUserIngredient` →
      `duplicateAfterWriteError`), used by both the website create and the v1 create; the v1 validation shares its
      helpers/messages with `PatchIngredient`; limits only from `internal/food` (no copied constants or functions)
- [x] Unexpected errors are logged with the injected logger (`user_id`), not with a global `slog` call; no
      `err.Error()` in responses
- [x] No new migration file
- [x] Service unit tests and integration tests cover the criteria above; `go vet ./...`, `go test ./...` and
      `go test -race` on the changed packages pass

## Skill and client

### `kkal_client.py add-ingredient`
- [x] `add-ingredient --name "Хлебцы ржаные" --kcal 310 --fats 0 --proteins 10` sends exactly one request,
      `POST /api/v1/ingredients`, with the body exactly `{"name": "Хлебцы ржаные", "kcal_per_100g": 310, "fats": 0,
      "proteins": 10}` (`0` is sent, `carbs` is absent); exit 0; stdout is the created ingredient as JSON
- [x] Without `--name` or without `--kcal` → non-zero exit, a usage message, no request
- [x] Duplicate name → exit 1, `HTTP 409` on stderr, stdout JSON containing `existing_ingredient`
- [x] `--kcal 1000` → exit 1, `HTTP 400` on stderr, stdout JSON whose message names `kcal_per_100g`
- [x] Against a server without the route (405, and 404) → exit 1 with a message that the server probably does not
      support adding ingredients yet and that nothing was created; no "Ingredient not found" interpretation
- [x] `edit-ingredient` / `delete-ingredient` keep their current messages and behavior
- [x] Module docstring lists `add-ingredient`; standard library only; `API_KEY_HERE` occurs exactly once in the template
      client
- [x] These cases are automated in `internal/server/skill_client_test.go` and ran with python3 on the dev VM (not
      skipped)

### Version (decision Q1)
- [x] `kkal_client.py` has one constant `SKILL_VERSION = "2026-10-16"`; `version` prints `2026-10-16`, exit 0, no
      request, also with the unconfigured template key (automated in `skill_client_test.go`); the docstring lists
      `version`
- [x] SKILL.md body ("API basics") tells the agent to run `version` when asked for the installed version; the value is
      not repeated in SKILL.md; no new frontmatter field
- [x] The downloaded zip's client contains the same `SKILL_VERSION`; the zip still equals the template byte for byte
      apart from the substituted base URL / key / dev naming

### `SKILL.md` (checked by reading)
- [x] Frontmatter: only `name` and `description`; the description mentions adding an ingredient to the list (minimal
      change); `TestDescriptionLengthOfBothVariants` passes (both variants ≤ 1024)
- [x] "Production and dev": adding an ingredient is routed like logging (dev only when the message names dev)
- [x] "API basics" lists `add-ingredient` with an example
- [x] "Correcting the ingredient list" has the add-without-logging part:
      - triggers ("заведи продукт", "добавь в справочник", a label photo with such a request);
      - values per 100 g from the label (convert per-portion values; ask when unreadable);
      - a short name in the user's language;
      - check the fetched list for the same product under another name (e.g. "Ржаной хлеб" vs "Хлеб ржаной") and ask
        whether to update that one or add a separate one;
      - never log-and-delete to create an ingredient;
      - no `totals`/table;
      - the reply gives the created values and says nothing was logged
- [x] Label-photo intent: photo + an eating/logging verb → `add`; photo + "заведи продукт" / "добавь в справочник" →
      `add-ingredient`; photo alone or an unclear request → the agent asks "записать в дневник или только добавить в
      справочник?" and acts only on the answer
- [x] "Fixing mistakes": the right food missing from the list → `add-ingredient` + `edit ENTRY_ID --ingredient-id
      <new id>`, which keeps the entry, its weight and time; the "Delete the entry and `add` it again" instruction is gone
- [x] The deletion paragraph no longer implies that ingredients can only be re-created through `add`; no remaining
      sentence implies that
- [x] The limits line includes `add-ingredient`
- [x] Error handling covers `409` and `400` on `add-ingredient`, and the older-server message for adding (no retry, no
      log-and-delete fallback)
- [x] Everything else is kept: logging steps 1–7, totals, edit/delete rules, routing, key secrecy

## Docs
- [x] `docs/api.md`: `POST /api/v1/ingredients` with body, limits, name cleaning, 201 shape, 409, "no diary entry"
- [x] `docs/features.md`: adding an ingredient without logging (API + skill `add-ingredient`); the uniqueness bullet
      covers creating through the API; the skill's `SKILL_VERSION` and the client's `version` command, and that every
      download renders the skill from the running release
- [x] `docs/architecture.md`: the apidata domain mentions creating ingredients
- [x] `README.md` "External API" table lists `POST /api/v1/ingredients`, `PUT /api/v1/ingredients/:id`,
      `DELETE /api/v1/ingredients/:id`

## Integration
- [x] `go vet ./...`, `go test ./...` (after `make build-frontend`), `go test -race` on the changed packages are green
- [x] No files under `web/` changed
- [ ] Manual, after release: download the skill, run `add-ingredient` against the server, see the ingredient on the Food
      List page after a reload, and see that the day total on the dashboard did not change — **pending manual
      verification after release**; the same flow is automated against the wired test server
      (`TestSkillClientAddIngredient`, `TestIngredientAPICreateKeepsDiary`)

## Acceptance notes (2026-10-16, business-analyst)
- Verified by reading the uncommitted code, SKILL.md, the client, the docs and the tests. The test results are the
  ones QA recorded in `status.md`, iteration 2: `go vet`, `go test -count=1 ./...`, `-race` green; the python3 client
  tests ran (3.14.4) and none were skipped; each `TestSkillClientAddIngredient` subtest passes alone.
- Where each part is:
  - endpoint: `internal/handlers/apidata/handler.go` `CreateIngredient` and `dto.go` `CreateIngredientRequest`;
  - service: `internal/services/ingredient/service.go`, with `AddIngredient`, `validateAddRequest`, `cleanName` and
    `validateValues` shared with `validatePatchRequest`, the one create path `createIngredient`, and `withUsage`;
  - tests: `ingredient_api_test.go` (`TestIngredientAPICreate*`), `service_test.go` (`TestAddIngredient*`,
    `TestCreateIngredientKeepsWebsiteRules`), `handler_test.go` (500 without the internal text),
    `skill_client_test.go` (`TestSkillClientAddIngredient`, `TestSkillClientVersion`).
- The 201 body is asserted literally, keys in order, and equals the list row, including the inherited
  `times_used: 2` / `last_used`.
- Accepted design detail (Q-D2): if the statistics cannot be read after a successful insert, the 201 carries the
  ingredient with `times_used: 0`, so the client does not retry a create that succeeded.
- Older server: v5.14.0 answers 404 `Not Found` (checked by the developer). The client test simulates both 404 and
  405.
- "No files under `web/`": checked by modification time (every `web/` file is older than this task's changes) and
  by the review reports. Git was not run by the BA.
- Outside the spec, coordinator's edits:
  - `AGENTS.md`, rule "Agent skill": every template change bumps `SKILL_VERSION`;
  - `.gitignore` ignores `__pycache__/`.
  Both match the task.
