# API Endpoints

All routes are under `/api`.

- `GET /api/languages` - supported languages (public, no auth)
- `/api/auth/*` - authentication (rate limited: 5 req/sec per IP)
  - `POST /api/auth/login` - login
  - `POST /api/auth/register` - registration (additionally: 3 attempts, then 1 per 20 min per IP)
  - `GET /api/auth/activate/:token` - activate the account from the email link
  - `GET /api/auth/me` - current user (requires auth)
- `/api/calories/*` - calorie entry CRUD (GET, POST, PUT /:id, DELETE /:id)
- `/api/ingredients/*` - ingredient management (GET, GET /:id, POST, PUT /:id, DELETE /:id); `PUT` replaces all
  values (nutrients that are not sent are cleared). `POST` and `PUT /:id` return 409
  `{message, existing_ingredient}` (`existing_ingredient` in the shape of `GET /:id`) when the user already has another
  ingredient with the same name, ignoring case, spacing, invisible characters and `ё`/`е`
- `/api/weight/*` - weight history (GET, POST, PUT /:id, DELETE /:id)
- `/api/profile` - user profile (GET, PUT)
- `/api/profile/goal` - weight goal (GET progress, PUT set, DELETE clear)
- `/api/metrics` - health metrics (GET)
- `/api/reports/data` - aggregated report data (GET, `from`/`to` query params)
- `/api/preferences/:key` - per-user UI preferences (JWT auth only: an `X-API-Key` gives 401). The only key is
  `reports`, any other → 404 `{"message": "Unknown preference"}`
  - `GET` - always 200 with the full object. Nothing saved → the defaults `{"date_from": null, "date_to": null,
    "show_weight": true, "show_calories": true, "show_goal_projection": true, "period": "daily", "step_interval": 1}`.
    A stored value that cannot be read gives the defaults; an invalid stored field gives that field's default (an
    invalid range resets both dates to `null`)
  - `PUT` - replaces the whole value (body: the same object as `GET`) and returns the stored object (200). Fields of
    `reports`:
    - `date_from`: `YYYY-MM-DD` from `1900-01-01` to `2100-12-31`, or `null`; `date_to`: the same, the literal
      `"today"` (the client's current day, resolved by the client), or `null`. Both `null` (no saved range: the client
      default) or both set; `date_from` after `date_to` is accepted. A missing date counts as `null`
    - `show_weight`, `show_calories`, `show_goal_projection`: booleans, required
    - `period`: `daily` | `weekly` | `monthly` | `yearly`, required; `step_interval`: `1` | `5` | `10` | `15` (a JSON
      integer), required
    - unknown fields are ignored and not stored
  - Errors of `PUT`: 400 `{"message": "Invalid request body"}` for malformed JSON or a body that is not an object; 400
    with a message that starts with the field name for a missing or invalid field (`period must be one of daily,
    weekly, monthly, yearly`, `date_from and date_to must both be set or both be null`, ...), nothing is stored; 413
    for a body over 4 KB; 500 `Failed to load preferences` / `Failed to save preferences` on a database failure
- `/api/ai/*` - AI analysis (rate limited: 2 req/min)
  - `GET /api/ai/status` - AI service status
  - `POST /api/ai/analyze` - run an analysis
- `/api/export` - data export (POST, weight/food as Excel or email)
- `/api/api-keys/*` - API key management (JWT auth)
  - `POST /api/api-keys` - create a key (returns the full key once)
  - `GET /api/api-keys` - list the user's keys (prefixes only)
  - `POST /api/api-keys/:id/revoke` - revoke a key
  - `DELETE /api/api-keys/:id` - delete a key
- `POST /api/skills/food-log/download` - the food logging agent skill as a ready-to-install zip (JWT auth, no request
  body). **Every call creates a new API key** of the user: name `Claude skill YYYY-MM-DD HH:MM UTC`, no expiry; earlier
  keys stay active. Rate limited per user: 5 downloads at once, then 1 per minute
  - 200: body is the zip; headers `Content-Type: application/zip`,
    `Content-Disposition: attachment; filename="<skill>.zip"`, `Cache-Control: no-store`. `<skill>` is
    `kkal-tracker-food-log`, or `kkal-tracker-food-log-dev` on a server with `ENVIRONMENT=development`
  - Zip: one folder `<skill>/` with exactly two files, both pointed at `APP_URL`: `SKILL.md` (mode 0644) and
    `scripts/kkal_client.py` (mode 0600), which holds the new key in place of the `API_KEY_HERE` placeholder
  - 401 `{"message": ...}` without or with an invalid JWT (no key is created); 429
    `{"message": "Too many requests, please try again later"}` over the limit; 500
    `{"message": "Failed to prepare the skill"}` on any internal failure (a key created for a failed download is deleted)
- `/api/v1/*` - external API (`X-API-Key` header, rate limited per IP: burst of 20, then 1 req/sec)
  - `GET /api/v1/data?type=weight|food|both&from=YYYY-MM-DD&to=YYYY-MM-DD`
  - `GET /api/v1/ingredients` - all user ingredients with `times_used` and `last_used`
  - `POST /api/v1/ingredients` - add an ingredient to the list without a diary entry: `{name, kcal_per_100g, fats?, carbs?,
    proteins?}`. The same limits and name cleaning as `new_ingredient` of `POST /food` (`kcal_per_100g` required, 0–900;
    macros 0–100, `null` or not sent = unknown; name up to 200 characters after cleaning). 400 names the field
    (`name is required`, `kcal_per_100g is required`, `<field> must be between ...`). Returns 201 with the ingredient
    as a row of `GET /ingredients` (`{id, name, kcal_per_100g, fats?, carbs?, proteins?, times_used, last_used?}`; the
    statistics are matched by name, so diary entries that already have exactly this name count); 409
    `{message, existing_ingredient}` when the user already has an ingredient with the same name (ignoring case,
    spacing, invisible characters and `ё`/`е`). No diary entry is created
  - `POST /api/v1/food` - store a meal: `{meal_datetime?, items: [{ingredient_id, weight, kcal_per_100g?, fats?, carbs?, proteins?} | {new_ingredient: {name, kcal_per_100g, fats?, carbs?, proteins?}, weight, one_off?}]}`.
    On an `ingredient_id` item, `kcal_per_100g` (0–900) and `fats`/`carbs`/`proteins` (grams per 100 g, 0–100) override
    the ingredient's values for this entry only; values that are not sent are copied from the ingredient (macros may be
    null). The ingredient does not change. Out of range → 400 `items[<i>]: <field> must be between 0 and 100 (grams per
    100 g)`; any of these fields at the top level of a `new_ingredient` item → 400 `items[<i>]: <field> of a new food
    belongs inside new_ingredient`. A rejected item rejects the whole meal
  - `PUT /api/v1/food/:id` - change a food entry: `{weight?, ingredient_id?, kcal_per_100g?, fats?, carbs?, proteins?, meal_datetime?}`;
    fields that are not sent (or `null`) keep their values, calories are recalculated (macros do not affect them).
    `ingredient_id` first copies the name, calories and all three macros from that ingredient, then the overrides sent
    in the same request (`kcal_per_100g`, `fats`, `carbs`, `proteins`) are applied on top. A macro cannot be reset to
    null, `0` means none. Same limits as `POST /food`; 400 with nothing changed on a violation or when nothing is sent.
    Returns `{entry, day, day_total_calories?}`; the ingredient does not change
  - `DELETE /api/v1/food/:id` - delete a food entry
  - `PUT /api/v1/ingredients/:id` - change an ingredient of the list: `{name?, kcal_per_100g?, fats?, carbs?, proteins?}`;
    fields that are not sent (or `null`) keep their values, the sent ones are checked against the limits of `POST /food`
    (`kcal_per_100g` 0–900, macros 0–100, name up to 200 characters after cleaning). Returns `{ingredient}`;
    409 `{message, existing_ingredient}` when another ingredient of the user has the same name (ignoring case, spacing,
    invisible characters and `ё`/`е`), 404 for a missing or foreign ingredient. Diary entries do not change
  - `DELETE /api/v1/ingredients/:id` - delete an ingredient of the list (204); diary entries stay
