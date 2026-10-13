# Acceptance Criteria: Food logging skill: template-only skills/, key in the client, macros step restored, per-entry macro overrides

## Backend: Part A (template and download)
- [x] `skills/` contains only `kkal-tracker-food-log/SKILL.md`, `kkal-tracker-food-log/scripts/kkal_client.py`,
      `embed.go` and `embed_test.go`. `skills/kkal-tracker-food-log-dev/` does not exist, and no `skills/*/api_key`
      exists in the working tree.
- [x] The template client contains the line `API_KEY = "API_KEY_HERE"`. The token `API_KEY_HERE` occurs exactly once
      in the template client and zero times in `SKILL.md`.
- [x] `cmd/skillgen/` and the `skill-dev` Make target (including `.PHONY`) are gone.
      `grep -rn "skillgen\|skill-dev\|kkal-tracker-food-log-dev/" --exclude-dir=task-specs` finds nothing outside the Go
      dev variant constant and its tests and docs describing the generated variant.
- [x] The repository has no references to an `api_key` skill file outside `.gitignore`, `.dockerignore`, the
      explicit-embed comment and docs that call those entries a safety net.
- [x] `.gitignore` and `.dockerignore` still contain `skills/*/api_key`. `skills/embed.go` still lists the two files
      explicitly, and its test passes.
- [x] A download with `ENVIRONMENT=production`, `APP_URL=https://kcal.peskov.info` returns a zip with exactly two
      entries:
      - `kkal-tracker-food-log/SKILL.md` with mode 0644;
      - `kkal-tracker-food-log/scripts/kkal_client.py` with mode 0600.
- [x] The client in that zip contains `API_KEY = "<64 lowercase hex>"` and no `API_KEY_HERE`. Putting
      `API_KEY_HERE` back in place of the key gives the repository template byte for byte, and `SKILL.md` equals the
      repository file byte for byte.
- [x] A download with `ENVIRONMENT=development`, `APP_URL=https://dev-kcal.peskov.info` returns
      `kkal-tracker-food-log-dev.zip`, with two entries under `kkal-tracker-food-log-dev/`:
      - the frontmatter `name: kkal-tracker-food-log-dev` and the DEV description;
      - the title `# Kkal Tracker: food diary logging (DEV)` followed by the `[DEV]` banner;
      - the base URL `https://dev-kcal.peskov.info` in both files;
      - the key in the client.
- [x] The key from the client works as `X-API-Key` on `GET /api/v1/ingredients` (200, only the downloading user's
      data). It is listed in `GET /api/api-keys` as `Claude skill YYYY-MM-DD HH:MM UTC` without expiry.
- [x] Key substitution:
      - a raw key that is not `^[0-9a-f]{64}$` (empty, 63 or 65 characters, uppercase, containing `"`) → error, and the
        error text does not contain the key;
      - the anchor missing or present twice → `ErrSourceLayout`.
- [x] A template whose key anchor is missing or duplicated fails at render time, and no API key is created.
- [x] A failure after the key is created (substitution or zip) → 500 `{"message": "Failed to prepare the skill"}`,
      and the key is deleted.
- [x] The raw key never appears in log output, error messages or response headers.
- [x] `TestDescriptionLengthOfBothVariants` exists and passes: both descriptions are ≤ 1024 characters. The frontmatter
      name test passes for both variants.
- [x] Client key lookup:
      - `KKAL_API_KEY` set → used;
      - not set → the embedded constant;
      - an `api_key` file next to `SKILL.md` is ignored;
      - the placeholder constant with no env variable → exit 1 with a message that names Settings → AI and contains
        neither `api_key` nor "API Keys".

## Backend: Part B (`totals` command)
- [x] `totals` without `--day`, or with `--day 2026-13-01` / `--day 13.10.2026` → non-zero exit, and no HTTP request
      is made.
- [x] `totals --day D` makes exactly one request, `GET /api/v1/data?type=food&from=D&to=D`, and prints JSON with the
      keys `day`, `meal`, `day_total`, `missing_meal_ids`, `without_macros`, `check`, in that order. Inside the rows the
      order is `kcal`, `fats`, `carbs`, `proteins`.
- [x] Example. Day entries:
      - A: 200 g, 534 kcal, F 22.5, C null, P 16;
      - B: 120 g, 24 kcal, F 0.2, C 4.2, P 0.6;
      - C: 150 g, 330 kcal, F 9, C 20, P 14;
      - D: 100 g, 50 kcal, no macros.

      With `--meal A B C`:
      - `meal`: kcal 888, fats round(45 + 0.24 + 13.5) = 59, carbs round(0 + 5.04 + 30) = 35,
        proteins round(32 + 0.72 + 21) = 54;
      - `day_total`: entries 4, kcal 938, same grams;
      - `without_macros` = [{D, in_meal: false}].
- [x] Rounding is halves up: an entry giving 12.5 g → 13, 0.5 g → 1.
- [x] `--expect-meal-kcal 888 --expect-day-kcal 938` → `check: {"meal": "ok", "day": "ok"}`.
      `--expect-day-kcal 1000` → `"day": "mismatch"`. Without the `--expect-*` flags, the checks are `null`.
- [x] `--meal` with an ID not on that day → the ID is in `missing_meal_ids`, and with `--expect-meal-kcal`,
      `check.meal` is `"mismatch"`.
- [x] Without `--meal`, `meal` is `null`.
- [x] An empty day gives a `day_total` of 0 entries and 0 for all values.
- [x] Exit code 0 whenever the data was fetched (also on a mismatch). HTTP errors (401, 429, 5xx) and network errors
      → exit 1 with the status, like the other commands.
- [x] The `totals` behavior is covered by an automated test that runs on the dev VM (not skipped), per Q-D1.

## Backend: Part C (per-entry macro overrides)
- [x] `POST /api/v1/food`. Ingredient 94 has kcal 250, F 1, C 50, P 8. The item
      `{"ingredient_id": 94, "weight": 100, "kcal_per_100g": 260, "fats": 3, "carbs": 48, "proteins": 9}` → 201, and
      the entry has kcal_per_100g 260, fats 3, carbs 48, proteins 9, calories 260. `GET /api/v1/ingredients` still shows
      ingredient 94 with 250/1/50/8.
- [x] `POST` with `{"ingredient_id": 94, "weight": 100, "fats": 3}` → the entry has fats 3, carbs 50, proteins 8,
      kcal_per_100g 250.
- [x] `POST` where an ingredient has null carbs and the item sends `carbs: 0` → the entry has carbs 0.
- [x] `POST` with `fats: -1` or `proteins: 100.01` on item 1 of 2 → 400
      `items[1]: <field> must be between 0 and 100 (grams per 100 g)`, and no entry of the meal is saved. `0` and `100`
      are accepted.
- [x] `POST` with top-level `fats` on a `new_ingredient` item → 400
      `items[<i>]: fats of a new food belongs inside new_ingredient`, and nothing is saved.
- [x] `PUT /api/v1/food/:id {"proteins": 20}` → 200: only proteins changes, everything else including calories stays,
      and the response has `day` and `day_total_calories`.
- [x] `PUT {"ingredient_id": 76}` → name, kcal and all three macros are re-copied from ingredient 76 (existing
      behavior).
- [x] `PUT {"ingredient_id": 76, "fats": 5}` → name, kcal, carbs and proteins come from 76, fats is 5.
- [x] `PUT {"ingredient_id": 76, "kcal_per_100g": 300, "carbs": 10}` → kcal 300, carbs 10, fats and proteins from 76.
- [x] `PUT {"carbs": 101}` → 400 `carbs must be between 0 and 100 (grams per 100 g)`, and the entry is unchanged.
- [x] `PUT {}` → 400, with a message listing `weight, ingredient_id, kcal_per_100g, fats, carbs, proteins or
      meal_datetime`.
- [x] `PUT {"fats": null}` alone → 400 nothing to change. With other fields, `null` keeps the current fats.
- [x] In no case does a POST or PUT change the ingredient's row.
- [x] Another user's entry or ingredient → 404 / `ingredient N not found` as today.
- [x] No new migration file. The `/api/calories` routes are unchanged.
- [x] The client command `edit 123 --fats 3 --carbs 48 --proteins 9` sends exactly those fields. `edit 123` with no
      options → exit 1 with a message listing `--fats`, `--carbs`, `--proteins` among the options.

## Backend: SKILL.md content (checked by reading)
- [x] "API basics":
      - the key is inside the downloaded skill's client;
      - `KKAL_API_KEY` is for manual checks;
      - on "not configured" or `401`, download the skill again in Settings → AI;
      - no `api_key` file, no "save the key".
- [x] The rate limit line says a normal logging takes three calls: `ingredients`, `add`, `totals`.
- [x] The commands list includes `totals` and the `edit` macro flags.
- [x] Step "Add up the macros":
      - uses `totals --day <day from the add response> --meal <ids> --expect-meal-kcal … --expect-day-kcal …`;
      - explains why the day is passed explicitly;
      - says never to add up by hand;
      - covers mismatch handling for both rows, foods without macros, and a missing `day_total_calories`.
- [x] Step "Reply to the user":
      - a per-item list, then a real markdown table (not in a code block) with rows "Приём" and "За день";
      - columns in the order Ккал, Ж, У, Б (К-Ж-У-Б), stated explicitly, the same order in every language;
      - calories from the server, grams from `totals`, rounded to whole numbers;
      - a failed row shows `—` in the gram cells;
      - notes: assumed weights, chosen ingredient, new ingredients, one-off items, foods without macros, overrides
        (kcal-only override → the macros come from the ingredient).
- [x] After `edit` and after `delete`, the skill shows the "За день" row recomputed with `totals` for that entry's day;
      after `delete`, `totals` is stated to be the only source.
- [x] Label values for a known food → kcal and macros are overridden together on the item or with `edit`. The
      ingredient is changed only after the user agrees, as before.
- [x] Dev/prod routing, edit-ingredient / delete-ingredient, error handling and the download text are kept.
- [x] The description is unchanged, or both variants are ≤ 1024 characters.

## Docs
- [x] `AGENTS.md`:
      - a rule that the skill is edited only in `skills/kkal-tracker-food-log/` in the repository, never in an installed
        copy, with the reason;
      - the Secrets rule mentions the `API_KEY_HERE` placeholder and key substitution at download;
      - the `make skill-dev` line is removed.
- [x] `docs/api.md`:
      - the two-file zip with the key in the client (0600);
      - the `POST /api/v1/food` item macro fields;
      - the `PUT /api/v1/food/:id` macro fields and the ingredient-switch semantics.
- [x] `docs/features.md`: the key in the client, the dev variant generated at download, per-entry macro overrides, the
      skill's meal/day macros table.
- [x] `docs/architecture.md`, `README.md` ("Agent skills", line 238, Project Structure) and `docs/infrastructure.md`
      have no dev copy, no generator and no `api_key` file as a mechanism.

## Integration
- [x] `go vet ./...` and `go test ./...` are green (after `make build-frontend`), and `go test -race` on the changed
      packages is green.
- [ ] Manual check on the dev server: download the dev zip, unpack it, and run
      `python3 scripts/kkal_client.py ingredients`; it works with no environment variables.
      `python3 scripts/kkal_client.py totals --day <today UTC>` prints the JSON above.
      **Pending manual verification after release.** The same path is automated against the wired test server
      (`TestSkillClientTotals`, `TestSkillClientTemplateWithoutKey`).
- [ ] Manual check in Claude: logging a meal with the downloaded skill ends with the per-item list and a rendered table
      in the order Ккал, Ж, У, Б, with rows "Приём" and "За день". The "За день" kcal equals the website's day total.
      **Pending manual verification after release.** This also covers the day row after edit/delete in a real
      conversation, the prod download contents and key listing, and the claude.ai upload. The SKILL.md text for all of
      them is verified above.
- [x] No frontend files changed.

## Acceptance notes (2026-10-13, business-analyst)
- How each criterion was verified:
  - Part A: `internal/skill` (`Render`, `checkKeyPlaceholder`, `WithAPIKey`), `internal/services/skill/service.go`
    (render → key → substitute → zip, `deleteKey` on failure, client 0600); tests `skill_test.go`,
    `service_test.go`, `skill_api_test.go` (two entries, modes, template byte for byte, dev variant).
  - Part B: `kkal_client.py` `cmd_totals`; `skill_client_test.go`, which ran on Python 3.14.4 per QA. The spec's
    example numbers (888/59/35/54, day 938) are asserted literally.
  - Part C: `services/calorie/service.go` (`overrideNutrients`, `firstSet`), handler mapping;
    `TestFoodAPIOverridesMacrosOfAnEntry`, `TestFoodAPIEditsMacrosOfAnEntry`, the isolation test.
- Accepted deviations:
  - `totals` without `--day` exits 2 (an argparse usage error). The criterion only requires a non-zero exit.
  - The client message uses the ASCII "Settings -> AI", the same style as its other messages.
- The SKILL.md description is unchanged from v5.13.1.
- Remaining `api_key` mentions are the safety nets plus tests that check that a stray `api_key` file is ignored.
