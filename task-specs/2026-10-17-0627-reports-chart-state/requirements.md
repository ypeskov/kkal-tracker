# Reports page: remember the chart settings (per user, on the server) and keep the trend line up to date

## Overview
On the Reports page (`/reports`), Charts tab, the user picks a date range and the chart checkboxes (Weight, Calories,
Show goal projection). Today all of them are lost on a page reload (and also when the user goes to another page of
the app and comes back): the page always opens with "last 30 days" and every checkbox on.

The trend line also goes stale. After the user adds, edits or deletes a weight entry, the chart shows the new weight
point but keeps the old trend line. Only a reload refreshes it, and the reload resets the selected range. So the user
cannot have the selected range and an up-to-date trend at the same time.

The user asked for three things:
1. Save the selected range and apply it when the page is opened.
2. Redraw the chart (including the trend line) whenever its parameters or the underlying data change.
3. Also remember the checkboxes (weight, kcal, goal).

The settings are stored on the server per user, so they follow the user across devices and browsers (decision Q1).

Business value: the user sets the report view once (e.g. "since my goal started") and sees it on every visit and on
every device. The trend line always reflects the current weigh-ins, so it can be trusted without reloading the page.

## Decisions (answers of the user, 2026-10-17)
- **Q1 → B, on the server per user.** The settings are synced across devices, which makes this a full-stack task
  with a migration.
- **Q2 → as recommended.** From is saved as a fixed date. To is saved as "today" when it equals today at the moment
  of the change, so it rolls forward; otherwise it is saved as a fixed date. Without a saved range the default stays
  "last 30 days", rolling.
- **Q3 → keep the trend "since the goal start".** Only the refresh after weight changes is fixed.
- **Q4 → also remember Period and Step**, but not the active tab.
- **Q5 → one shared range** for the Charts and Weight History tabs, as now (BA recommendation, not asked).
- **Q6 → no new frontend test framework.** The backend gets Go unit tests and integration tests on the wired test
  server, like the other endpoints (decided by the user).

### Requirement changes by the user (2026-10-17, during development, after the frontend review)
- **C1. Logout and login clear the client caches.** This replaces the earlier "out of scope" observation. The
  frontend review found that a re-login in the same tab could show or save U1's settings for U2.
  - `App.tsx` `endSession()` serves both logout paths: the menu logout, and the automatic logout when
    `GET /api/auth/me` fails. In this order, it:
    1. removes the token;
    2. runs `queryClient.clear()` and `ingredientService.clearCache()`;
    3. shows the login screen.
  - `startSession()` (login) clears the same caches again before the user is loaded. This covers a mutation of the
    previous session that completed after the logout.
  - A pending Reports save carries the token of the session that made the change. It is dropped when the token
    differs at send time. Consequence: a change made less than ~0.8 s (the debounce) before a logout is not saved.
  - The ingredient `sessionStorage` cache was not cleared on logout before. Now it is.
- **C2. Midnight rule for To.**
  - Only From changes, and the saved To was `"today"` or not saved (`null`, the rolling default) → To stays
    `"today"`, even when the page was opened on an earlier day. The To input then moves to today's date, so the page
    shows what is saved.
  - A saved fixed To stays fixed.
  - When the user changes To, the original rule applies: `"today"` only when it equals today.

## Current State

### Page state (`web/src/pages/Report.tsx`)
- All filter state is local `useState` with hard-coded defaults (lines 19-32):
  - `activeTab` = `'chart'`;
  - `showWeight`, `showCalories`, `showGoalProjection` = `true`;
  - `period` = `'daily'`, `stepInterval` = `'1'`;
  - `dateFrom` = today − 30 days, `dateTo` = today (`format(..., 'yyyy-MM-dd')`, local time).
- Nothing is persisted, so all of this is lost on every reload and on every unmount: navigating to another route and
  back remounts `Report`.
- One date range is shared by both tabs. `ReportFilters` shows the From/To inputs on both tabs, and `WeightHistory`
  receives `dateFrom`/`dateTo` (line 295). Period, step interval and the checkboxes are shown only on the Charts tab.
- Controls (`web/src/components/ReportFilters.tsx`):
  - two `<input type="date">` (From, To), no presets;
  - a Period select (daily/weekly/monthly/yearly);
  - a Step select (1/5/10/15 days), shown only for daily;
  - checkboxes Weight, Calories and "Show goal projection" (`report.show_goal_projection`). The last one is shown only
    when the goal has a target date (`hasGoalWithDate`, `Report.tsx:241`). A cleared date input gives `''`. The
    backend then falls back to the last 30 days (`internal/services/reports/service.go:33-36`).

### Data fetching (TanStack Query v5, `new QueryClient()` with defaults in `web/src/main.tsx:24`)
- `['reportData', dateFrom, dateTo]` → `GET /api/reports/data?from&to` (`Report.tsx:35-38`). The key contains the
  dates, so a range change fetches new data. While a new range loads, `isLoading` replaces the chart with "Loading…"
  (`Report.tsx:269-273`).
- `['weightGoalProgress']` → `GET /api/profile/goal` (`Report.tsx:41-44`). It is the only source of the trend line,
  the target line and `target_date`.

### How the trend line is computed
- On the server: `internal/services/profile/service.go:226-266` (`weightTrend`), called from `GetWeightGoalProgress`
  at line 356. It is a least-squares fit through the initial weight at the goal start and every weigh-in after the
  goal start day. It does not depend on the selected report range (as documented in `docs/features.md`, lines 10-11).
- On the client: `web/src/components/reports/WeightCaloriesChart.tsx:82-92` evaluates
  `startWeight + kgPerDay × days since start` on each chart date. It is drawn only when `showWeight` is on (lines
  97-124), and so is the target line (lines 127-138).
- "Show goal projection" extends the x axis by one point per day from the last date that has data to the goal's target
  date, up to 730 days (`WeightCaloriesChart.tsx:60-75`).

### Root causes

**(a) The range and the checkboxes are not remembered.**
- `Report.tsx:19-32` keeps them in component state only. There is no persistence of any kind.

**(b1) The trend line does not update after a weight is added, edited or deleted. This is a bug.**
- The weight mutations in `web/src/components/reports/WeightHistory.tsx` (create, lines 111-113; update, 126-128;
  delete, 140-142) invalidate only `['weightHistory']` and `['reportData']`.
- They do not invalidate `['weightGoalProgress']`, which holds the server-fitted `trend`.
- Switching tabs changes only `activeTab` (`Report.tsx:19, 246-296`), so the `['weightGoalProgress']` observer stays
  mounted and no refetch-on-mount happens. With the default QueryClient, the stale query is refetched only on window
  refocus, on reconnect or on a reload.
- Result: the new weight point is drawn, but the trend line is the old one. The user reloads to fix it, and the reload
  resets the range: this ties (b) to (a).
- `['latestWeight']` (`Profile.tsx:58-61`) and `['healthMetrics']` (the BMI uses the latest weight,
  `internal/services/metrics/service.go:36-46`) are not invalidated either. They live on other pages and refetch when
  those pages mount.

**(b2) The trend line looks the same after a range change. This is by design (decision Q3).**
- The chart is redrawn on a range change, but the trend is fitted since the goal start. Only its visible segment
  changes.
- With the projection on, the x axis always runs to the target date, so a range change alters only a small part of
  the chart.

### Per-user settings and storage today
- Per-user settings are columns of `users`:
  - `language` and `activity_level`;
  - the weight goal (`target_weight`, `target_date`, `goal_set_at`, `initial_weight_at_goal`, migration
    `20260129123416_add_weight_goal.sql`).
- `QueryGetUserByID` selects every column explicitly (`internal/repositories/queries.go:101-112`). The user model is
  scanned in `GetByID`/`GetByEmail` (`internal/repositories/user.go`), and `/api/auth/me` returns it.
- There is no preferences table and no preferences endpoint. In the browser, `localStorage` holds only the i18next
  language cache, and `sessionStorage` holds the JWT and the ingredient cache.
- Newer per-user tables use their own table with `user_id ... REFERENCES users(id) ON DELETE CASCADE` (`api_keys`,
  `20260211193105_add_api_keys.sql`), a repository per entity, SQL in `queries.go` with SQLite and (unwired) Postgres
  variants.
- Note: `PRAGMA foreign_keys` is not enabled (`internal/database/database.go`), so `ON DELETE CASCADE` is not enforced
  in SQLite. This is pre-existing and the same for all tables.
- Migrations are Goose SQL files named `YYYYMMDDHHMMSS_<name>.sql` (`make migrate-create NAME=...`). They are
  **never applied automatically**. The `deploy` skill stops the release when there are new migrations and asks the
  user how to apply them to production.
- JWT-protected routes are registered in `internal/server/server.go:162-201` (`apiGroup.Group(<prefix>,
  authMiddleware.RequireAuth)`). There is a global body limit of 1 MiB (`maxRequestBodySize`,
  `internal/server/security.go:23`).
- Integration tests run on the fully wired server with a temporary migrated SQLite database
  (`newSmokeEnvironment`, `signUp`, `client.expect` in `internal/server/smoke_test.go`; example:
  `weight_goal_test.go`).
- The frontend loads the current user (`['user']`, `GET /api/auth/me`) before the router renders (`App.tsx:61-96`).
  The user, with `id`, is available through `useRouteContext({ from: '__root__' })` (`DashboardPage.tsx:19`).
- Logout (`App.tsx:85-88`) only drops the token. The TanStack Query cache is not cleared.

## Storage design (developer decision, recommended)
Recommended: **a generic per-user UI preferences table, key → JSON, with a server-side allow-list of keys and a typed,
strictly validated struct per key.** Only one key, `reports`, exists in this task.

```sql
CREATE TABLE user_preferences (
    user_id    INTEGER NOT NULL,
    pref_key   TEXT    NOT NULL,
    value      TEXT    NOT NULL,          -- JSON written by the server, never the raw request body
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, pref_key),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

Reasons:
- **Not columns on `users`.** Seven UI-only columns would be added to every `users` SELECT and scan, which serve auth
  and the profile, and to `/api/auth/me`. The goal columns are domain data. These are view settings.
- **Not a dedicated `report_settings` table.** Typed columns would add nothing: the values are never filtered or
  joined in SQL, and the server validates them in Go anyway. The next page that wants to remember its filters (e.g.
  the Dashboard date filter, which resets too) would need yet another migration and table.
- **Generic store with per-key validation.** One migration now, and future pages add a key and a Go struct without a
  migration. Strictness is the same as with typed columns:
  - the server decodes the request into the key's struct and validates every field;
  - it stores its own re-encoding of the normalized struct, never the client's JSON;
  - unknown keys are rejected.
- It is the developer's call to deviate (e.g. a dedicated table), with the reason recorded in `status.md` (Q-D1). The
  API and the behavior below do not depend on it.

## Requirements

### Backend

#### B1. Migration
- One new Goose migration, created with `make migrate-create NAME=add_user_preferences`. It creates the table above
  (or the alternative from Q-D1), with a working `Down` that drops it.
- Applying it with `make migrate-up` and rolling it back with `make migrate-down` works on a copy of a current
  database. No existing table changes. No data migration.
- The server does not apply it (AGENTS.md).

#### B2. Endpoints (JWT, `authMiddleware.RequireAuth`)
- New domain `preferences` (handler `internal/handlers/preferences/`, service `internal/services/preferences/`,
  repository `internal/repositories/user_preference.go`, model, SQL in `queries.go`), wired in `server.go`.
- `GET /api/preferences/:key` and `PUT /api/preferences/:key`. The only allowed key is `reports`. Any other key → 404
  `{"message": "Unknown preference"}`.
- The user is always the JWT user. No user id in the path or the body. An API key (`X-API-Key`) does not authorize
  these routes: 401, as for every JWT route.
- Body/response of the `reports` key. JSON keys in snake_case, all fields always present in responses:
  ```json
  {
    "date_from": "2026-09-01",
    "date_to": "today",
    "show_weight": true,
    "show_calories": false,
    "show_goal_projection": true,
    "period": "weekly",
    "step_interval": 1
  }
  ```
  - `date_from`: a valid calendar date `YYYY-MM-DD` between `1900-01-01` and `2100-12-31`, or `null`.
  - `date_to`: such a date, the literal `"today"`, or `null`.
  - The range is all or nothing. Either both `date_from` and `date_to` are `null` (no saved range, so the client uses
    the rolling default), or both are set. `date_from` cannot be `"today"`.
  - `date_from` later than `date_to` is accepted. It is not validated today either, and "today" cannot be compared on
    the server.
  - `show_weight`, `show_calories`, `show_goal_projection`: JSON booleans, required.
  - `period`: one of `daily`, `weekly`, `monthly`, `yearly`, required.
  - `step_interval`: an integer, one of `1`, `5`, `10`, `15`, required.
  - Unknown fields are ignored. This is consistent with the other APIs, and an open tab of an older frontend after a
    release keeps saving.
- `GET`:
  - always 200 with the full object;
  - nothing stored → the defaults: `date_from`/`date_to` `null`, three `true`, `daily`, `1`;
  - a stored value that cannot be decoded, or has an invalid field, → 200 with the defaults for the invalid fields.
    The range is reset as a whole. The server logs a Warn with `user_id` and `key`, never the raw value.
- `PUT`:
  - replaces the whole stored value of the key: an upsert of one row per user and key;
  - 200 with the stored (normalized) object;
  - the server stores its own re-encoding of the validated struct, plus a schema version (e.g. `"v": 1`). The schema
    version is not part of the API object.
- Errors:
  - malformed JSON, or a body that is not a JSON object (`null`, `[]`, a scalar, empty) → 400 `Invalid request body`;
  - a missing or invalid field, including a wrong type (e.g. `"step_interval": "5"`, `"show_calories": "yes"`) or
    `null` for a required field → 400 with a message that names the field, e.g. `period must be one of daily, weekly,
    monthly, yearly`, `step_interval must be one of 1, 5, 10, 15`, `show_calories must be true or false`,
    `date_from and date_to must both be set or both be null`. This was aligned at acceptance with the criteria and the
    implementation; the first version of this spec said wrong types → `Invalid request body`;
  - a missing `date_from` / `date_to` counts as `null`. The other fields are required. Field names match exactly
    (case-sensitive);
  - a body over **4 KiB** → 413 (route-level `BodyLimit`, below the global 1 MiB);
  - a database failure → logged with the injected logger (`user_id`, `key`, `error`) and a generic 500
    (`Failed to load preferences` / `Failed to save preferences`), never `err.Error()`.
- No rate limiter is required. A write is an upsert of one bounded row per user and key, so it cannot grow the
  database. The security reviewer may say otherwise.

#### B3. Tests
- Service unit tests:
  - every validation rule with boundary values: `2026-02-28` passes, `2026-02-30` / `2026-13-01` / `26-09-01` /
    `""` fail; `1900-01-01` and `2100-12-31` pass, `1899-12-31` and `2101-01-01` fail;
  - `"today"` is accepted for `date_to` only;
  - a range with one bound `null` fails;
  - every `period` / `step_interval` value, plus `hourly`, `7` and `"5"` (a string);
  - a missing boolean fails;
  - normalization: what is stored is the server's encoding;
  - read fallbacks: an undecodable stored value gives all defaults; an invalid stored field gives that field's
    default; an invalid stored range resets the range; the Warn is logged without the value;
  - an unknown key is rejected.
- Integration tests on the wired server (e.g. `internal/server/preferences_test.go`):
  - a new user's GET → the defaults;
  - PUT → 200 with the normalized object, then GET returns the same; a second PUT replaces the first;
  - isolation: U1 PUTs, then U2's GET → the defaults, and U2's PUT does not change U1's settings;
  - 401 without a JWT and with only an `X-API-Key`;
  - 404 for an unknown key (GET and PUT);
  - each 400 case above, after which the stored value is unchanged;
  - 413 over 4 KiB, after which the stored value is unchanged;
  - unknown fields are ignored;
  - a corrupt stored value (written directly into the test database) → GET 200 with the defaults.
- The temporary database of the test environment is created by running the migrations, so the new migration is
  exercised by every integration test.
- `go vet ./...`, `go test ./...` (after `make build-frontend`), and `go test -race` on the changed packages pass.

### Frontend

#### F1. Loading the settings before the first report request
- An API service `web/src/api/preferences.ts` with `getReportPreferences()` / `saveReportPreferences(...)`, following
  the existing service classes.
- Query key: `['preferences', 'reports', user.id]`. The user id comes from the router context. It is a second guard,
  besides the cache reset on logout and login (C1).
- The Reports page first resolves the settings, then renders the filters and the data. Until the settings query
  succeeds or fails, the page shows the existing "Loading…" state. Neither the filters nor the default values appear.
  The report data query is `enabled` only after that.
- The first `GET /api/reports/data` uses the restored range. There is no request with the default range before it.
- Recommended: prefetch the settings once after login, at the app shell level, so the Reports page usually opens
  without waiting. This is the developer's choice (Q-D2).
- The client validates the GET response field by field with the same rules as the server, and falls back to the
  defaults per field. A `GET` failure (network, 500) → all defaults: the page works, with no error shown.
- The settings are applied to the page state **once per mount** (on page entry). A background refetch while the user
  is on the page (e.g. on window focus) does not change the controls.

#### F2. Range semantics (Q2)
- Defaults, when nothing is saved (range `null`) or the saved range is invalid: From = today − 30 days, To = today.
  They are computed on each visit, so they roll forward daily, as now.
- When the user changes a date:
  - **From** is saved as the absolute date;
  - **To** is saved as `"today"` when it equals today's date at that moment, otherwise as the absolute date;
  - when only From changes and the saved To was `"today"` or `null`, To stays `"today"`, and the To input shows
    today (C2: the midnight case). A saved fixed To stays fixed.
- Changing only a checkbox, the period or the step keeps a `null` range `null`: the default keeps rolling.
- A cleared date input (`''`) is not saved. The last saved range stays. The page keeps its current in-memory behavior
  for the empty input.
- "Today" is the browser's local date (`format(new Date(), 'yyyy-MM-dd')`). The server stores `"today"` literally and
  never resolves dates.
- The resolution (saved settings + today → From/To) is a pure function.

#### F3. Saving
- Every user change of a persisted field:
  1. updates the page state at once;
  2. updates the query cache at once (`setQueryData`), so returning to the page in the same session shows the latest
     values even if the save has not finished;
  3. schedules a save of the **whole** settings object.
- Debounce: one PUT after about 1 s without further changes (the developer picks 500–1000 ms). Several quick changes
  (typing a date, toggling several checkboxes) produce one request with the final state.
- At most one PUT is in flight. A change made meanwhile is sent after it completes, so an older state never overwrites
  a newer one on the server.
- Leaving the page (unmount on SPA navigation) flushes a pending save immediately.
- Best effort on reload/close: on `pagehide` a pending save is sent with `fetch(..., { keepalive: true })`. A change
  followed by a reload within the debounce window may still be lost, which is acceptable. A change followed by a
  reload after the debounce must be kept.
- A failed save (network, 4xx/5xx):
  - the UI keeps working with the in-memory state;
  - no toast, banner or alert, and no automatic retry loop;
  - at most one `console.warn` per failed request;
  - the next change sends the full current state again.
  - A 401 is handled as by the other requests today, with no special flow.
- Multiple devices/tabs: last write wins. Another open tab or device picks up the change on its next entry to the page.
  No live sync.
- `showGoalProjection` is saved and restored even while its checkbox is hidden (goal without a target date).
- The active tab is not saved (Q4). The page always opens on Charts.

#### F4. Trend line and goal data refresh after weight changes (fixes b1)
- A successful create, update or delete of a weight entry (`WeightHistory.tsx`) also invalidates
  `['weightGoalProgress']`, `['latestWeight']` and `['healthMetrics']`, besides the existing `['weightHistory']` and
  `['reportData']`.
- Back on the Charts tab, the trend line is the one for the new data, without a reload or a window refocus.

#### F5. Dynamic redraw
- A change of From/To, the period, the step or a checkbox redraws the chart in place, with the new data and the trend
  line evaluated on the new dates.
- While a new range is loading, the previous chart stays on screen (TanStack Query v5 `placeholderData:
  keepPreviousData` on `['reportData', …]`). "Loading…" is shown only on the first load. No stale dataset stays
  visible after the new data arrives.

#### F6. Code placement and conventions
- The settings logic (load, validate, resolve the range, debounced/serialized save, flush) goes into a typed hook,
  e.g. `web/src/hooks/useReportPreferences.ts`. `Report.tsx` only consumes it.
- Read the current docs of TanStack Query v5 (`keepPreviousData`, `setQueryData`, `invalidateQueries`, `enabled`),
  `react-chartjs-2` 5.x and `fetch` `keepalive` before coding (AGENTS.md library rule).
- No new user-facing strings are expected. Any new string goes through `t()` in all 4 locales. Tailwind only. No CSP
  change: same-origin `fetch` only.

### Docs
- `docs/api.md`: `GET` / `PUT /api/preferences/:key` (JWT; key `reports` only, 404 otherwise), the body fields and
  rules, `"today"`, the defaults on GET, 400/413, and that the request replaces the whole value.
- `docs/features.md`, "Reports & Analytics":
  - the Reports page remembers the range ("To = today" rolls forward), the chart checkboxes, the period and the step,
    per user on the server, so they are the same on every device;
  - the trend line refreshes after weight changes.
- `docs/features.md`, "Persistent Sessions": logout and login clear the client caches (C1).
- `docs/architecture.md`:
  - add `preferences` to the domains list, with a note "per-user UI preferences (key → JSON validated per key)";
  - add the new model to `models/`.
- `AGENTS.md`: no change expected.

### DevOps
- The release contains a **new migration**. Per the `deploy` skill, stop before the release and agree with the user
  how it is applied to the production database (record the agreement in `status.md`).
- Recommended order:
  1. take a snapshot of the production database (`prod-db-snapshot` skill);
  2. apply the migration to production with `cmd/migrate` (`up`);
  3. release.
- If the new image runs before the migration, `/api/preferences/reports` returns 500. The Reports page then still
  works with the defaults and silent failed saves (F1, F3), so the order is forgiving, but the feature is inactive
  until the migration is applied.
- No new environment variables.
- Post-release check:
  - on one device, change the range and a checkbox and reload: the settings are kept;
  - open the page on another device or browser: the same settings;
  - `GET /api/preferences/reports` without a JWT → 401.

## Subtasks

Chain (full-stack): BA → Backend Dev → Backend Review → Frontend Dev → Frontend Review → Security Review →
QA Review → BA Acceptance → DevOps.

### Backend Developer
1. Read the current Echo v5 docs (route-level `BodyLimit`, binding) and Goose docs (AGENTS.md library rule).
2. Migration (B1). Decide Q-D1 and record it in `status.md`.
3. Model, repository (get / upsert, SQL in `queries.go`, SQLite and Postgres variants like the neighbors), service
   (allow-list of keys, the `reports` struct, validation, normalization, read fallbacks), handler (B2), wiring in
   `server.go` with `RequireAuth` and the 4 KiB route limit.
4. Unit and integration tests (B3).
5. Docs: `docs/api.md`, `docs/architecture.md`. The `features.md` part can wait for the frontend developer.
6. `go vet ./...`, `go test ./...`, `go test -race` (changed packages) green. Check `make migrate-up` /
   `make migrate-down` on a local copy. Set `status.md` → `backend-review`.

### Frontend Developer
1. Read the current docs of TanStack Query v5, `react-chartjs-2` 5.x and `fetch` `keepalive`.
2. `web/src/api/preferences.ts`. A hook with the load/validate/resolve/save logic (F1–F3, F6), with a pure range
   resolution function.
3. `Report.tsx`: gate on the settings, initialize the state once per mount, save on every change, and use
   `keepPreviousData` for the report data (F5).
4. `WeightHistory.tsx`: the additional invalidations (F4).
5. Optional prefetch after login (Q-D2).
6. Manual check in the dev server (`dev-server` skill) of the frontend and integration scenarios in
   `acceptance-criteria.md`, including the Network tab checks.
7. Docs: `docs/features.md`.
8. From `web/`: `npx tsc --noEmit -p tsconfig.app.json`, `npm run lint`, `npm run build` pass. Go tests stay green.
   Set `status.md` → `frontend-review`.

### DevOps Engineer
Release after BA acceptance and the user's commit (workflow steps 7.5 and 8), with the migration step agreed with the
user first (see "DevOps"). Post-release checks as listed there.

### Reviews
- `backend-code-reviewer`:
  - layering (handler → service → repository, SQL only in `queries.go`);
  - the stored value is the server's own encoding;
  - read fallbacks;
  - generic 500s and the injected logger;
  - the migration with `Down`;
  - test coverage of B3.
- `frontend-code-reviewer`:
  - no default flash and no default-range request;
  - state applied once per mount;
  - the debounce, the single in-flight save and the flush on unmount/pagehide;
  - silent failure;
  - the query key with the user id;
  - the invalidations;
  - no settings access outside the hook.
- `security-reviewer`:
  - the routes are JWT-only, user-scoped by the token, and not reachable with an API key;
  - only allowed keys;
  - strict validation and the 4 KiB limit;
  - no raw client JSON stored or logged;
  - no internal errors in responses;
  - no cross-user leak through the client cache (the user id in the query key);
  - whether a rate limit is needed.
- `qa-engineer`:
  - the Go tests (race, coverage of the new packages);
  - the frontend scenarios verified manually in the browser on the dev server and recorded (no frontend test runner,
    Q6).

## Open Questions

### For the user
None blocking. At release time the DevOps engineer asks how the migration is applied to production (recommended:
snapshot, `cmd/migrate up`, then the release).

### For the developers (decide during implementation, record in `status.md`)
- **Q-D1. Storage shape.** Recommended: the generic `user_preferences` key → JSON table with per-key typed validation
  (see "Storage design"). The alternative is a dedicated `report_settings` table. The API and the behavior stay the
  same.
- **Q-D2. Prefetch.** Recommended: prefetch `['preferences', 'reports', user.id]` once after login (app shell), so the
  page usually opens without the loading state. Without it, the page waits for one small GET on entry.
- **Q-D3. Chart updates.** Whether `react-chartjs-2` needs `redraw` / a `key` to avoid stale datasets, or whether
  in-place updates are enough (datasets are matched by `label`).

### Observations outside the scope (not to be changed in this task)
- ~~Logout does not clear the TanStack Query cache~~: brought into scope by the user (C1).
- Residual, pre-existing, same tab only (frontend and security reviews):
  - a profile or language mutation of the previous session that is still in flight after the next login can write
    its response into the new session's cache through `onSuccess` `setQueryData`;
  - the reset at login narrows this window but cannot cancel a mutation.
- `PRAGMA foreign_keys` is not enabled, so `ON DELETE CASCADE` does not run in SQLite (all tables).
- The daily chart uses a category x axis with only the dates that have data. A straight-in-time trend line can look
  bent where weigh-ins are unevenly spaced.
- The projection extends from the last data date even when the selected To is in the past.
- The Dashboard date filter is not remembered either. It can reuse `user_preferences` with a new key later.
