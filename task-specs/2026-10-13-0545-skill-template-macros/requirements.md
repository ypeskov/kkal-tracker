# Food logging skill: template-only skills/, key in the client, macros step restored, per-entry macro overrides

## Overview
Three linked changes to the `kkal-tracker-food-log` skill and the food API.

- **A. `skills/` is the template only.** The repository keeps one skill, with placeholders. The download fills in
  everything that changes: base URL, the dev naming, and now the API key. The key goes straight into
  `kkal_client.py`, so there is no `api_key` file any more. This removes:
  - the committed dev copy and its generator;
  - the local `api_key` files that hold real keys on the dev VM;
  - the habit of editing installed copies.
  Fewer places for a skill to drift, and no secret files lying next to the sources.
- **B. Restore the macros step (regression).** The user wants proteins, fats and carbs next to the calories, for the
  meal and for the day. The step existed only in a hand-edited installed copy (2026-10-04) and was lost when the
  2026-10-12 download replaced it. This time it goes into the repository. The summing becomes a client command, not
  arithmetic done by the agent.
- **C. Per-entry macro overrides.** When the user reads label values for a known food, the agent can store kcal and
  macros for that entry together. Today only kcal can be overridden, so the entry ends up with the label's calories
  and the ingredient's macros.

## Current State

### Skill sources and download
- `skills/kkal-tracker-food-log/`:
  - `SKILL.md`: frontmatter on lines 1–4, the title, and ``- Base URL: `https://kcal.peskov.info` ``. "API basics"
    describes the key lookup `KKAL_API_KEY` → `api_key` file → constant, and the manual fallback "create a key in
    Settings → API Keys and save it to the `api_key` file". The rate limit line says "a normal logging takes two
    calls".
  - `scripts/kkal_client.py`:
    - `API_KEY = "XXXXXXXXX"` and `API_KEY_PLACEHOLDER = "XXXXXXXXX"`;
    - `API_KEY_FILE` points to `<skill dir>/api_key`;
    - `load_api_key()` reads the env variable, then the file, then the constant;
    - the "not configured" message mentions the file.
- `skills/kkal-tracker-food-log-dev/`: a committed generated copy (SKILL.md and client). Both skill folders also hold a
  git-ignored `api_key` file with a **real key** on the dev VM. Never print these files.
- `skills/embed.go`: embeds exactly the two template files, as string vars. `skills/embed_test.go` checks the
  directives and that the embedded content equals the files on disk.
- `internal/skill/skill.go`:
  - `Render(src, development, baseURL)`:
    - normalizes and validates the URL (`normalizeBaseURL`);
    - for dev: frontmatter `name`/`description`, ` (DEV)` in the title and `devBanner`;
    - replaces the Base URL line and `DEFAULT_BASE_URL`, each anchor exactly once (`replaceOnce`, `ErrSourceLayout`);
    - checks the frontmatter name.
  - `WriteDir` is used only by `cmd/skillgen`.
  - Constants: `ProductionURL`, `DevelopmentURL` (used by the generator and tests).
- `internal/skill/skill_test.go`:
  - `TestCommittedDevCopyIsUpToDate` (the drift test) and `TestWriteDirKeepsOtherFiles`;
  - `TestDescriptionLengthOfBothVariants` (≤ 1024, keep);
  - assertions on the `XXXXXXXXX` placeholder.
- `cmd/skillgen/main.go`, `make skill-dev` (`Makefile` lines 4 and 97–99).
- `internal/services/skill/service.go`, `DownloadFoodLog`:
  1. render (before any key exists, so layout and URL failures create no key);
  2. `CreateKey` with the name `Claude skill YYYY-MM-DD HH:MM UTC`;
  3. build the zip: `<name>/SKILL.md` and `<name>/scripts/kkal_client.py` with mode 0644, and `<name>/api_key` (raw key
     plus `\n`) with mode 0600;
  4. if zipping fails, delete the key.

  It logs only `api_key_id`.
- `internal/server/skill_api_test.go`:
  - `skillKey` reads the key from `<name>/api_key` and expects exactly 3 zip entries;
  - `TestSkillDownloadInstallsAWorkingSkill` expects `API_KEY = "XXXXXXXXX"` in the client;
  - `TestSkillDownloadMatchesTheRepository` compares the zip with `skills/<name>/`, including the committed dev copy.
- Docs that mention the dev copy, the generator or `api_key`:
  - `AGENTS.md`: line 19 (`make skill-dev`), lines 41–43 (Secrets);
  - `README.md`: lines 245–247, 256;
  - `docs/architecture.md`: lines 20, 34, 38–40;
  - `docs/features.md`: lines 38–41;
  - `docs/api.md`: lines 36–37;
  - `docs/infrastructure.md`: line 13.
- `.gitignore:38` and `.dockerignore:10` hold `skills/*/api_key`. `.air.toml` `include_file` lists the two template
  files, which keep their paths.
- Frontend (`AiTab.tsx`, `settings.ai.*`) says nothing about an `api_key` file. The warning "the zip contains a secret
  key" stays true. **No frontend changes are needed.**

### Diary day
- `CreateMeal` / `UpdateMealEntry` convert `meal_datetime` to UTC. `day` in the response is that UTC date, and
  `day_total_calories` is the sum over that UTC date. `GET /api/v1/data?from&to` filters by the UTC date of
  `meal_datetime`. The web dashboard also uses the UTC date (`toISOString`).
- `kkal_client.py entries` defaults to today's UTC date.
- So the server's diary day is always the UTC date. The reason to pass the day explicitly to the macros command:
  - the `day` of the response differs from "today" when `meal_datetime` names another day ("вчера вечером");
  - UTC midnight can pass between `add` and the follow-up call;
  - only the response's `day` is guaranteed to be the day `day_total_calories` was counted for.

  (The user's wording "wrong diary day after midnight in the user's timezone" describes the symptom only in part.
  The requirement stands.)

### Food API (Part C)
- `internal/handlers/apidata/dto.go`:
  - `FoodItemRequest`: `ingredient_id`, `new_ingredient`, `one_off`, `weight`, `kcal_per_100g` (an override, only for
    `ingredient_id` items);
  - `UpdateFoodRequest`: `weight`, `ingredient_id`, `kcal_per_100g`, `meal_datetime`.
- `internal/services/calorie/service.go`:
  - `CreateMeal`: an `ingredient_id` item copies `Name`, `KcalPer100g`, `Fats`, `Carbs`, `Proteins` from the
    ingredient; `kcal_per_100g` is validated with `food.ValidKcalPer100g`; `kcal_per_100g` on a `new_ingredient` item →
    400 "belongs inside new_ingredient".
  - `UpdateMealEntry`:
    - empty request → `EntryValidationError` "nothing to change…";
    - `ingredient_id` re-copies name, kcal and all three macros;
    - then `kcal_per_100g` overrides kcal;
    - weight; time.
- Calorie entries already store their own nullable `fats`, `carbs`, `proteins` (`migrations/20250907200000_initial_schema.sql`).
  `CalorieEntryRepository.Update` and `CreateMeal` write them. **No migration is needed** (verified).
- `food.ValidNutrientPer100g` (0–100) already exists and is used for `new_ingredient` and `PUT /ingredients/:id`.
- `GET /api/v1/data` and the `add`/`edit` responses already return `fats`/`carbs`/`proteins` per 100 g per entry,
  omitted when null.

### Regression source (reference)
The lost installed copy (`~/.claude/skills/.trash/1791830534115-2285-ZED3Wu/kkal-tracker-food-log/SKILL.md`) had:
- "### 6. Add up the macros": `entries --from <day>` piped into an inline Python snippet;
- "### 7. Reply to the user": a table with columns Б-Ж-У.

Its text is a reference for wording only. Its column order and the inline snippet are replaced by the decisions below.
It predates the edit-ingredient / delete-ingredient sections and the download text, which stay.

## Requirements

### Part A: template-only `skills/` and the key in the client (backend)
1. **Placeholder.** In `skills/kkal-tracker-food-log/scripts/kkal_client.py` the key line becomes
   `API_KEY = "API_KEY_HERE"`. The token `API_KEY_HERE` occurs **exactly once** in the template client (in that line)
   and nowhere in `SKILL.md`. The client must recognize an unconfigured key without a second literal occurrence of
   the token (e.g. build the comparison string by concatenation). The rendered client must not contain `API_KEY_HERE`
   at all.
2. **Key substitution at download.** The service order becomes:
   1. render, which checks every anchor, including the key line, exactly once, before any key exists;
   2. create the key;
   3. substitute the key;
   4. zip;
   5. respond.

   - The substitution is validated like the base URL. The raw key must match `^[0-9a-f]{64}$`, and the anchor
     `API_KEY = "API_KEY_HERE"` must occur exactly once. Otherwise an error is returned (`ErrSourceLayout` or a
     dedicated error).
   - A failure after key creation (substitution or zip) deletes the key (the existing cleanup path) and gives the
     generic 500 `Failed to prepare the skill`.
   - The raw key never appears in error messages, logs or headers.
   - The API shape (e.g. a method on the rendered skill vs a parameter of `Render`) is up to the developer (Q-D2).
3. **Zip layout.** It has exactly two entries, `<name>/SKILL.md` and `<name>/scripts/kkal_client.py`, and **no
   `api_key`**. `kkal_client.py` now holds the secret, so its mode is **0600** (the protection `api_key` had), and
   `SKILL.md` stays 0644. Everything else in the response is unchanged: headers, the key name `Claude skill … UTC`,
   no expiry, the rate limit.
4. **Client key lookup.**
   - Order: `KKAL_API_KEY` (non-empty after strip) → the embedded constant.
   - Remove `API_KEY_FILE` and all reading of an `api_key` file.
   - `KKAL_BASE_URL` stays as it is.
   - Update the configuration comment and the "not configured" message: download the skill again in Kkal Tracker
     (Settings → AI) and install it in place of this one. `KKAL_API_KEY` is mentioned only as an override for manual
     checks.
   - No mention of `api_key` or of Settings → API Keys as a way to configure the skill.
5. **SKILL.md "API basics".**
   - Auth: the downloaded skill already contains its key in `scripts/kkal_client.py`. `KKAL_API_KEY` overrides it for
     manual checks.
   - If the script says the key is not configured, or on `401`, ask the user to download the skill again in Kkal
     Tracker (Settings → AI) and install it in place of this one.
   - The rate limit line changes as in Part B.
6. **Delete the dev copy and the generator:**
   - `skills/kkal-tracker-food-log-dev/` (the tracked files);
   - `cmd/skillgen/`;
   - the Makefile target `skill-dev` (and its `.PHONY` entry);
   - `TestCommittedDevCopyIsUpToDate`;
   - `WriteDir` and its test, if unused after the generator is gone;
   - `DevelopmentURL`, if it is only used by tests.

   Grep the repo, excluding `task-specs/`, for `skillgen`, `skill-dev`, `kkal-tracker-food-log-dev/` and `api_key`
   (the skill file, not the DB table), and update every reference.
7. **The dev variant stays, generated at download time.** On `ENVIRONMENT=development` the download renders the dev
   name `kkal-tracker-food-log-dev`, the DEV description and the ` (DEV)` title with the `[DEV]` banner, exactly as
   today. This keeps decision D3 of the 2026-10-12 task: the dev and prod skills can be installed side by side. Dev
   differs from prod only by generated values, so it is not stored.
8. **Keep** `TestDescriptionLengthOfBothVariants`: both descriptions are ≤ 1024 characters. Keep the frontmatter name
   test and the anchor exactly-once tests. Add the key anchor to the exactly-once tests.
9. **Local key files.** The developer deletes `skills/kkal-tracker-food-log/api_key` and
   `skills/kkal-tracker-food-log-dev/api_key` from the working tree without printing or copying their content.
   `.gitignore` and `.dockerignore` keep `skills/*/api_key` as a safety net. The explicit two-file embed and its test
   stay.
10. **AGENTS.md rule** (Rules section):
    - The food logging skill is edited only in `skills/kkal-tracker-food-log/` in the repository. Never edit an
      installed copy (e.g. `~/.claude/skills/…`): changes there are lost on the next download, which is how the
      macros step was lost.
    - The repository holds the template only. Base URL, dev naming and the API key are filled in by the download.
    - Also update the Secrets rule: the skill template holds the `API_KEY_HERE` placeholder, and the key is put into
      the client only in the download zip. Remove the `make skill-dev` line from Commands.

### Part B: macros step (skill text + client command)
1. **New client command `totals`** in `kkal_client.py`:
   ```
   python scripts/kkal_client.py totals --day YYYY-MM-DD [--meal ID ...] [--expect-meal-kcal N] [--expect-day-kcal N]
   ```
   - `--day` is **required** and validated as `YYYY-MM-DD` before any request (exit 1 with a message).
   - One request: `GET /api/v1/data?type=food&from=<day>&to=<day>`. HTTP and network errors are handled like the
     other commands (exit 1).
   - Per entry, grams of a nutrient = `value_per_100g × weight / 100`. A missing (null/absent) nutrient counts as 0.
   - `kcal` of a row is the sum of the entries' server `calories` (integers).
   - The **meal row** includes only the entries whose `id` is in `--meal`. It is `null` when `--meal` is not given.
     The **day row** includes all entries of the day.
   - Grams are summed unrounded and rounded once at the end to whole grams, halves up (not Python's banker's
     `round`).
   - `without_macros` lists the entries with all three nutrients missing, with `in_meal` true/false. An entry with
     some nutrients is not listed.
   - `missing_meal_ids` lists the `--meal` IDs not found on that day.
   - Checks, each `null` when its `--expect-*` is not given:
     - `check.meal` is `"ok"` when the meal kcal equals `--expect-meal-kcal` and `missing_meal_ids` is empty,
       otherwise `"mismatch"`;
     - `check.day` is `"ok"` when the day kcal equals `--expect-day-kcal`, otherwise `"mismatch"`.
   - Output: pretty JSON (`ensure_ascii=False`), keys in this order (К-Ж-У-Б, like the reply table):
     ```json
     {
       "day": "2026-10-13",
       "meal": {"entry_ids": [3622, 3623, 3624], "kcal": 888, "fats": 59, "carbs": 35, "proteins": 54},
       "day_total": {"entries": 7, "kcal": 1640, "fats": 87, "carbs": 96, "proteins": 112},
       "missing_meal_ids": [],
       "without_macros": [{"id": 3624, "food": "Кофе", "in_meal": true}],
       "check": {"meal": "ok", "day": "ok"}
     }
     ```
   - The exit code is 0 when the data was fetched, whatever the checks say. The agent reads `check`.
   - Add it to the module docstring and the argparse help.
2. **SKILL.md, new step "### 6. Add up the macros"** (after "5. Send the meal"):
   - Why: the user wants fats, carbs and proteins next to the calories, for the meal and for the day. The server totals
     only calories. Never add up macros by hand; run `totals`.
   - Command: `totals --day <day from the add response> --meal <ids of the entries just stored>
     --expect-meal-kcal <total_calories> --expect-day-kcal <day_total_calories>`. Omit `--expect-day-kcal` when the
     response has no `day_total_calories`; the day kcal then comes from `totals`.
   - Always pass the `day` from the response, never "today": it is the diary day the server counted
     `day_total_calories` for. It differs from today when `meal_datetime` names another day, and `entries`/`totals`
     cannot guess it.
   - `check.meal == "mismatch"`: show the meal's server calories and say the meal macros could not be calculated.
     `check.day == "mismatch"`: show the server's `day_total_calories` and say the day macros could not be calculated.
     Do not show numbers that do not belong together.
   - Foods in `without_macros`: name them in the reply. The totals are understated by them.
   - "How much today?" questions: `totals --day <today's UTC date, the day the website shows>` without `--meal`; the day
     row only.
   - Update the response description in step 5: entries also carry the nutrients per 100 g, plus the diary `day`.
3. **SKILL.md, "### 7. Reply to the user"** (replaces the current step 6):
   - A short per-item list: food, grams, kcal.
   - Then a **real markdown table, not in a code block**, with the rows "Приём" and "За день" and the columns in the
     order **К-Ж-У-Б**: Ккал, Ж, г, У, г, Б, г.
     - Use the same order in every language, with headers in the user's language.
     - State explicitly that it is not the usual Б-Ж-У order.
     - Calories come from the server (`total_calories`, `day_total_calories`), grams from `totals`, whole numbers.
     - A row whose check failed shows the server kcal and `—` in the gram cells.
   - Example (the item list may be shown in a code block as an illustration; the table must be shown as it should be
     written):
     ```
     |         | Ккал  | Ж, г | У, г | Б, г |
     |---------|-------|------|------|------|
     | Приём   | 888   | 59   | 35   | 54   |
     | За день | 1 640 | 87   | 96   | 112  |
     ```
   - Then notes:
     - assumed weights;
     - the chosen ingredient when there was a choice;
     - new ingredients with their values;
     - one-off items;
     - foods without macros;
     - any per-entry override. When only kcal was overridden, say that the macros still come from the ingredient.
4. **After `edit` / `delete`**: show the entry as it is now and the "За день" row again.
   - After `edit`: `totals --day <day from the edit response> --expect-day-kcal <day_total_calories>`; kcal from the
     response.
   - After `delete` the server returns nothing, so `totals --day <the entry's day>` is the only source. The entry's
     day is the `day` of the add/edit response that produced it, or the UTC date of its `meal_datetime` in `entries`.
5. **Rate limit line**: "A normal logging takes three calls: `ingredients`, `add` and `totals`."
6. Update the commands list at the top of SKILL.md with `totals` and the new `edit` flags.
7. Keep everything else in the current SKILL.md: dev/prod routing, the edit-ingredient / delete-ingredient sections,
   error handling, the download text.

### Part C: per-entry macro overrides (backend)
1. **`POST /api/v1/food`**: `ingredient_id` items accept optional `fats`, `carbs`, `proteins` (grams per 100 g) next to
   `kcal_per_100g`.
   - Each sent field overrides the value copied from the ingredient for **this entry only**. Fields not sent are
     copied from the ingredient as today, and may be null.
   - Each is validated with `food.ValidNutrientPer100g` (0–100 inclusive). On failure: 400
     `items[<i>]: <field> must be between 0 and 100 (grams per 100 g)`, and nothing is saved (the meal stays atomic).
   - On a `new_ingredient` item, a top-level `fats`/`carbs`/`proteins` → 400
     `items[<i>]: <field> of a new food belongs inside new_ingredient` (like `kcal_per_100g` today).
   - The ingredient itself never changes: the `GET /api/v1/ingredients` values stay the same.
2. **`PUT /api/v1/food/:id`** accepts optional `fats`, `carbs`, `proteins`.
   - Fields not sent (or `null`) keep the entry's current values.
   - With `ingredient_id` in the same request, the name, kcal and all three macros are first re-copied from the new
     ingredient. Then every override sent in the same request (`kcal_per_100g`, `fats`, `carbs`, `proteins`) is
     applied on top. Overrides not sent come from the new ingredient.
   - Without `ingredient_id`, sent overrides replace the entry's current values.
   - Same 0–100 validation, 400 `<field> must be between 0 and 100 (grams per 100 g)`, nothing changed.
   - A request with only `fats`/`carbs`/`proteins` is a valid change. The "nothing to change" message lists all
     fields: `weight, ingredient_id, kcal_per_100g, fats, carbs, proteins or meal_datetime`.
   - A macro cannot be reset to null through the API; `0` is the way to say "none" (like `PUT /ingredients/:id`).
   - Calories are recalculated as today. Macros do not affect calories.
   - The ingredient never changes.
3. Service DTOs (`MealItem`, `UpdateMealEntryRequest`) get the three pointer fields. The handler maps them. The response
   shape is unchanged (entries already show `fats`/`carbs`/`proteins`).
4. **No migration.** The `/api/calories` (web) routes and the frontend are unchanged.
5. **Client**:
   - `edit` gets `--fats`, `--carbs`, `--proteins` (float, grams per 100 g, "for this entry only"), and the "Nothing to
     change" message lists them.
   - `add` needs no code change (the JSON passes through).
6. **SKILL.md**:
   - When the user gives label values for a known food ("на упаковке 260 ккал, Ж 3, У 48, Б 9"), keep the
     `ingredient_id` and override kcal and macros together on the item (`kcal_per_100g`, `fats`, `carbs`, `proteins`),
     or with `edit --kcal … --fats … --carbs … --proteins …`.
   - If the user gives only kcal, override only kcal and say in the reply that the macros still come from the
     ingredient.
   - The JSON example in step 5 shows an item with a full override.
   - The limits line mentions the macro overrides (0–100).
   - The "edit" examples and the "Correcting the ingredient list" wording (override the entry vs edit the ingredient)
     cover macros too.
   - The rule "offer to update the ingredient list, change it only after the user agrees" stays.

### Docs (backend developer)
- `docs/api.md`:
  - the download zip has two files, the key is in `scripts/kkal_client.py` (mode 0600), no `api_key`;
  - `POST /api/v1/food` item fields `fats?`, `carbs?`, `proteins?` and their semantics;
  - `PUT /api/v1/food/:id` fields and the ingredient-switch semantics.
- `docs/features.md`:
  - lines 38–41: the key goes into the skill's client, the dev variant is generated at download;
  - the per-entry kcal/macro overrides;
  - the skill's macros table (meal and day).
- `docs/architecture.md`: remove `cmd/skillgen` and the dev copy lines; `skills/` holds the template only;
  `internal/skill` renders the variant, the base URL and the key.
- `README.md` "Agent skills" and Project Structure: remove the dev copy and generator; describe the template and key
  substitution; `PUT /api/v1/food/:id` fields (line 238); remove `skillgen/` from the tree.
- `docs/infrastructure.md`: the `.dockerignore` entry is a safety net (the repository no longer uses `api_key` files).
- `AGENTS.md`: Part A §10.

### Frontend
None. Confirmed: the AI tab text does not mention `api_key`; "the zip contains a secret key" stays true; the Claude Code
line (unzip into `~/.claude/skills/`) stays valid.

### DevOps
Release only, with the `deploy` skill. There are no migrations. Before building, make sure the working tree has no
`skills/*/api_key`, and that `.dockerignore` still excludes it.

## Subtasks

### Backend Developer
1. Read the current docs where needed (AGENTS.md library rule): Python `argparse` (`nargs`/`action="extend"` for
   `--meal`) and `archive/zip` modes. Nothing new is expected in `go.mod`.
2. **Part A, server**:
   - placeholder and key anchor in the template;
   - key substitution with validation (Q-D2), and the service order render → key → substitute → zip, with key cleanup
     on any failure after creation;
   - zip without `api_key`, client file 0600.
3. **Part A, cleanup**:
   - delete `skills/kkal-tracker-food-log-dev/`, `cmd/skillgen/`, `make skill-dev`, the drift test, and `WriteDir`
     plus `DevelopmentURL` if unused;
   - delete the two local `skills/*/api_key` files (never print them);
   - grep and fix every reference.
4. **Part A, client**: key lookup env → constant, messages and comments.
5. **Part C**: DTOs, handler mapping, service validation and semantics (POST, PUT), errors as specified.
6. **Client**: the `totals` command (Part B §1), `edit --fats/--carbs/--proteins`.
7. **SKILL.md**: Part A §5, Part B §2–7, Part C §6. Write it in the existing style and language (English instructions,
   Russian examples). Keep the description unchanged unless it must change; it must stay ≤ 1024 characters in both
   variants.
8. **Tests**:
   - `internal/skill`:
     - the key anchor exactly once (missing, twice → error, no key created at the service level);
     - an invalid key (empty, 63/65 chars, uppercase, with a quote) → error, and the error text does not contain the
       key;
     - the rendered client contains `API_KEY = "<key>"` and no `API_KEY_HERE`;
     - the dev variant still has name, description, banner and URL;
     - remove the drift and `WriteDir` tests; keep `TestDescriptionLengthOfBothVariants`.
   - `internal/services/skill`:
     - exactly two zip entries, modes (client 0600, SKILL.md 0644), the key in the client;
     - a substitution failure after key creation deletes the key;
     - the raw key is not logged.
   - `internal/server/skill_api_test.go`:
     - `skillKey` extracts the key from the client (exactly two entries, `API_KEY = "<64 hex>"`);
     - production: the zip files equal the repository template byte for byte after putting the placeholder back in
       place of the key;
     - development: the zip has the dev name/description/banner/URL and otherwise equals the template;
     - the key from the client works on `/api/v1` (existing test, adapted).
   - Part C, unit tests (service, handler) and integration (`food_api_test.go`):
     - a POST override with all three fields, and with one field (the others from the ingredient);
     - out of range (-1, 100.01) → 400 with the item index, nothing saved;
     - top-level macros on a `new_ingredient` item → 400;
     - a PUT with only macros; a PUT with `ingredient_id` and an override in the same request; a PUT with
       `ingredient_id` only (re-copy); a PUT out of range → 400, unchanged;
     - the ingredient stays unchanged in every case.
   - `totals`, automated (Q-D1): the meal row and day row sums with a missing nutrient; `without_macros`;
     `missing_meal_ids`; both checks ok/mismatch/null; halves-up rounding; `--day` validation; key lookup without an
     `api_key` file.
   - `go vet ./...`, `go test ./...`, `go test -race` on the changed packages.
9. **Docs**: see "Docs" above, and AGENTS.md (Part A §10).

### Frontend Developer
None.

### DevOps Engineer
Release with the `deploy` skill after acceptance and commit.

Post-release checks:
- prod: Settings → AI download → `kkal-tracker-food-log.zip` with exactly `SKILL.md` and `scripts/kkal_client.py`,
  base URL `https://kcal.peskov.info`, the key in the client;
- dev server: `kkal-tracker-food-log-dev.zip` with `https://dev-kcal.peskov.info`.

Do not print keys in the release notes or logs. Remind the user to reinstall the skill from a fresh download: installed
copies from earlier downloads keep working (they read their `api_key` file), but they do not have the macros step.

### Reviews
`backend-code-reviewer` → `security-reviewer` (the secret moves from a separate 0600 file into the script; key
validation; no key in errors and logs; the local key files removed) → `qa-engineer` → BA acceptance. The frontend steps
are skipped.

## Open Questions

### For the developers (decide during implementation, record in `status.md`)
- **Q-D1, how to test `totals`.** The repository has no Python tests, and `go test` is the only test runner.
  - Recommended: a Go test in `internal/server`. It downloads the skill from the wired test server, unpacks it to
    `t.TempDir()`, and runs `python3 scripts/kkal_client.py totals …` against the server (`KKAL_BASE_URL` = the
    `httptest` URL; the key from the embedded constant, `KKAL_API_KEY` unset). It covers key embedding, the client and
    `totals` end to end. Skip with `t.Skip` when `python3` is not on `PATH`. QA confirms that it ran, not skipped, on
    the dev VM.
  - Pure-logic cases (rounding, missing nutrients) can be covered through crafted entries in the same test.
  - Do not put test files into `skills/kkal-tracker-food-log/` (template only) unless the explicit embed and the zip
    are guaranteed to exclude them.
- **Q-D2, the key substitution API.** A method on the rendered skill (e.g. `(*Skill).WithAPIKey(key)`) or a separate
  function, under these constraints:
  - the anchor is checked before the key is created;
  - the substitution validates the key format;
  - errors carry no key text.
- **Q-D3**: whether `ProductionURL` stays the anchor constant (yes, expected), and whether `DevelopmentURL`/`WriteDir`
  are removed or kept for tests.

### For the user (informational, no blocker)
- The stated reason for passing the day ("UTC date, wrong day after midnight in the user's timezone") is only partly
  accurate. The whole app, including the website, treats the diary day as the UTC date. Passing `day` from the response
  is still required for the reasons in Current State. The skill text will give those reasons. Changing the diary day
  to the user's local date would be a separate task.
- Out of scope, but worth a decision: the old installed copy in `~/.claude/skills/.trash/…` contains a real hard-coded
  API key. Delete that trash folder and, if that key is no longer needed, delete it in Settings → API Keys.
