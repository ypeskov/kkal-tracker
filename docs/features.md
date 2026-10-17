# Features

- **User Authentication**: JWT-based login/register/activate with bcrypt password hashing and email activation
- **Persistent Sessions**: JWT tokens stored in sessionStorage, survive page reloads. Logout (also the automatic one
  when the current user cannot be loaded) and login clear the client caches (TanStack Query cache, the ingredient
  cache in sessionStorage), so nothing of the previous account is shown or used in the same tab
- **Calorie Tracking**: food entries with name, weight, kcal/100g, auto-calculated totals
- **Dashboard**: today's entries with the total calorie count, weight goal card
- **Weight Management**: weight history over time with charts
- **Weight Goals**: target weight with optional target date and a start date, progress with visual indicators.
  Editing a goal keeps its start date and initial weight unless the start date is changed; the initial weight is the
  latest weight on or before the start date. The report chart shows a trend line: a least-squares fit through the
  initial weight and every weigh-in after the start day (`trend` in the goal progress), extended to the target date
- **Health Metrics**: BMI, BMR, TDEE based on the user profile (age, gender, activity level)
- **Ingredient Database**: global ingredients with multilingual names and nutritional data, copied into the user's
  own list on registration and managed on the Food List page
  - Ingredient names are unique per user ignoring case, spacing, invisible characters and `ё`/`е`: creating or renaming
    an ingredient to such a name on the Food List page (`POST` / `PUT /api/ingredients`) or through the API
    (`/api/v1`: `new_ingredient` of a meal, `POST` / `PUT /api/v1/ingredients`) is rejected with 409 and the existing
    ingredient; the Food List modal stays open with the input kept
    and names the existing ingredient (other save failures show a generic error there). A case-only rename of the ingredient itself is
    allowed; near-duplicates that existed before the rule are kept. Adding a diary entry on the website
    (`POST /api/calories`) still creates or updates its ingredient by the exact name, so it can add a near-duplicate
- **User Profiles**: personal settings, preferences, gender, activity level
- **Reports & Analytics**: calorie and weight trends
  - The Reports page remembers its settings per user on the server (`/api/preferences/reports`), so they are the same
    on every device and browser: the date range (shared by the Charts and Weight History tabs), the Weight, Calories
    and "Show goal projection" checkboxes, the period and the step. The active tab is not remembered: the page opens
    on Charts
  - Range: From is kept as a fixed date; To is kept as "today" when it was today at the time of the change, so it rolls
    forward, otherwise as a fixed date. A To the user did not change stays "today" when it was "today" or not saved
    (also in a tab left open past midnight). Without a saved range the page shows the last 30 days, rolling. A cleared
    date input is not saved
  - The settings are applied once on page entry, before the first report request (prefetched after login); until they
    arrive the page shows "Loading…". If they cannot be loaded the defaults apply. Changes are saved about a second
    after the last one (one request at a time, sent at once when leaving the page, best effort on reload); a failed
    save is silent and the next change sends the whole state again. Several devices: last write wins, another device
    picks the change up on its next visit of the page
  - Adding, editing or deleting a weigh-in refreshes the chart, the trend line, the latest weight and the health
    metrics without a reload. While a new range loads, the previous chart stays on screen
- **AI Insights**: AI-powered nutrition and weight analysis with personalized recommendations
  - OpenAI integration (GPT-4o-mini by default), provider selection in the UI (extensible to more providers)
  - Analysis periods of 7, 14, 30, 90 days, optional specific question
  - Answers in the user's language
  - Rate limited: 2 requests per minute
- **Data Export**: weight and food data as Excel files (download or email delivery)
- **API Key Data Access**: programmatic access to weight/food data
  - Key management UI in Settings (create, revoke, delete)
  - SHA-256 hashed storage, the full key is shown only once at creation
  - Time-limited (N days) or permanent keys
  - `X-API-Key` header authentication for the `/api/v1` endpoints
  - A key gives full access to the user's data, both read and write (there are no scopes)
  - Every skill download (Settings → AI) creates a key named `Claude skill YYYY-MM-DD HH:MM UTC` without expiry;
    earlier keys are never revoked automatically and are deleted on the API Keys tab
- **Food Logging API for AI agents**: an agent (e.g. voice dictation through a skill from `skills/`) stores meals via `/api/v1`
  - Settings → AI downloads the `kkal-tracker-food-log` skill as a zip ready to upload to Claude: the server creates a
    new API key, puts it into the skill's client script `scripts/kkal_client.py` (mode 0600 in the zip) and points the
    skill at its `APP_URL` (`POST /api/skills/food-log/download`). The repository holds only the template; the
    development server (`ENVIRONMENT=development`) generates the `kkal-tracker-food-log-dev` variant at download, so the
    dev and prod skills can be installed side by side
  - The skill is versioned by the `SKILL_VERSION` constant (a date) in its client, shown by the client's `version`
    command and bumped in every change of the template (`SKILL.md` or the client). Every download renders the skill
    from the running release, so a release ships the new skill; an installed copy stays as it was until the user
    downloads it again
  - After logging a meal the skill replies with the items and a table of calories, fats, carbs and proteins (columns
    Ккал, Ж, У, Б) for the meal and for the day. Calories come from the server, the grams from the client command
    `totals`, which adds up the entries of the diary day and checks them against the server's calorie totals
  - The AI tab explains the skill and shows the install steps for claude.ai (code execution and network access to the
    site's own domain, Settings → Capabilities → Skills → Upload skill) and Claude Code (unzip into
    `~/.claude/skills/`), the note that every download adds a key, and a warning that the zip holds a secret key.
    The file name comes from `Content-Disposition`; a 429 shows a "try again in a minute" message
  - The agent reads the whole ingredient list (`GET /api/v1/ingredients`, with usage statistics) and matches foods itself
  - `POST /api/v1/food` takes a meal as a list of items: `ingredient_id` for an existing ingredient or `new_ingredient`
    for a new one. The server never matches or creates ingredients by name on its own
  - A new ingredient whose name equals an existing one (ignoring case, spacing and `ё`/`е`) is rejected with 409 and
    the existing ingredient
  - `one_off: true` writes a new food to the diary without adding it to the ingredient list (restaurant dishes etc.)
  - A meal is atomic (all items or none); calories are calculated on the server
  - Per-entry overrides: an `ingredient_id` item, or `PUT /api/v1/food/:id`, can set `kcal_per_100g`, `fats`, `carbs`
    and `proteins` for this entry only (e.g. values from a label); the ingredient stays as it is
  - `PUT /api/v1/food/:id` corrects an entry (weight, another ingredient, calorie value, macros, time),
    `DELETE /api/v1/food/:id` removes it
  - The agent can add an ingredient to the list without logging anything (`POST /api/v1/ingredients`, the skill
    command `add-ingredient`), e.g. a product from a label photo; the day total does not change
  - The agent can correct ingredients of the list (`PUT /api/v1/ingredients/:id`, only the sent fields change) and
    delete them (`DELETE /api/v1/ingredients/:id`)
  - Diary entries keep their own copy of the food name, calories and macros: editing or deleting an ingredient does not
    change logged entries. `PUT /api/v1/food/:id {"ingredient_id": <same id>}` copies the corrected values into an entry.
    Usage statistics are matched by name, so a renamed ingredient starts from `times_used: 0`
- **Email Service**: activation emails and export delivery
- **Internationalization**: en_US, uk_UA, ru_UA, bg_BG with a language switcher, frontend and backend
- **Integrated Calculator**: built into every numeric input (weight, calories, macros), see the `frontend-style` skill
- **Database Backup**: nightly backups to Google Drive, see [`infrastructure.md`](infrastructure.md)
