# Task Status: Reports page: remember the chart settings (per user, on the server) and keep the trend line up to date

## Current Phase
**Phase**: `devops`
**Assigned to**: `devops-engineer`
**Updated**: 2026-10-17

Chain (full-stack): BA → Backend Dev → Backend Review → Frontend Dev → Frontend Review → Security Review →
QA Review → BA Acceptance → DevOps.
The user answered Q1–Q6 on 2026-10-17; the answers and the user's later changes C1 (logout/login clear the client
caches) and C2 (midnight rule for To) are recorded in `requirements.md`, "Decisions".

**Release note for DevOps — the release contains a migration**: `migrations/20261017065036_add_user_preferences.sql`
(creates `user_preferences`; Down drops it).
- Neither the deploy nor the server applies it. Per the `deploy` skill, stop before the release and **agree with the
  user how it is applied to production**, then record the agreement here.
- Recommended order:
  1. a snapshot with `prod-db-snapshot`;
  2. `cmd/migrate up` against the production database;
  3. the release.
- If the image runs before the migration, `/api/preferences/reports` answers 500 and the Reports page falls back to
  the defaults with silent failed saves. The feature is inactive until the migration is applied.
- Post-release checks:
  - change the range and a checkbox, reload: kept;
  - another device: the same settings;
  - `GET /api/preferences/reports` without a JWT → 401.

**Pending manual verification by the user**: logout → login as U2 in the same tab (no U1 settings visible, not even
briefly) → login as U1 again (U1's settings). Verified in code only.

## Developer decisions (Q-D1–Q-D3)

### Backend (2026-10-17)
- **Q-D1 → the recommended generic store.** Table `user_preferences (user_id, pref_key, value, updated_at)`, primary
  key `(user_id, pref_key)`, migration `migrations/20261017065036_add_user_preferences.sql` (Down drops the table).
  Nothing in the code argued for a dedicated table: the values are never queried in SQL, and the Dashboard filter can
  reuse the store with a new key. The allow-list is `definitions` in `internal/services/preferences/service.go`; a new
  key is a `definition` (version, defaults, field reader) without a migration.
- **Stored value**: the server's re-encoding of the normalized value plus the schema version, keys sorted:
  `{"date_from":…,"date_to":…,"period":…,"show_calories":…,"show_goal_projection":…,"show_weight":…,"step_interval":…,"v":1}`.
  A request's own `v` or other unknown fields never reach the database. A stored value without `"v": 1` (or not a JSON
  object) is read as all defaults with a Warn (`user_id`, `key`); invalid stored fields get their defaults with a
  Warn that lists the field names only, never the value.
- **One reader for requests and stored values** (`readReports`): every invalid field keeps its default and is
  reported. A request with any reported field → 400 with the first message in the order `date_from`, `date_to`, the
  range, `show_weight`, `show_calories`, `show_goal_projection`, `period`, `step_interval`; a stored value → the
  fallbacks above. The range is valid or reset as a whole.
- **Type errors name the field.** `"step_interval": "5"` or `"show_calories": "yes"` → 400 with the field's message
  (`step_interval must be one of 1, 5, 10, 15`, `show_calories must be true or false`), as the acceptance criteria
  require. `Invalid request body` is for malformed JSON and bodies that are not a JSON object (`null`, `[]`, a scalar,
  empty, trailing data). `null` for a boolean, period or step is "missing" → 400.
- **A missing `date_from`/`date_to` counts as `null`** (both missing = no range); the other fields are required.
- **Field names match exactly** (no case-insensitive matching as with `encoding/json` structs): `"Show_Weight"` is an
  unknown field, so `show_weight` is missing → 400.
- **Body limit**: `maxPreferencesBodySize = 4 << 10` in `internal/server/security.go`, a `BodyLimit` on the
  `/api/preferences` group after `RequireAuth` (GET has no body). Echo rejects a larger `Content-Length` before the
  handler; for a body without `Content-Length` the handler returns the 413 the limited reader yields while reading.
- **The handler reads the raw body** (`io.ReadAll`) instead of `c.Bind`: the service decodes it per key. The
  `Content-Type` is not checked (the JWT is a header, not a cookie, so there is no CSRF angle).
- **Migration file** created with the goose CLI pinned to the go.mod version
  (`go run github.com/pressly/goose/v3/cmd/goose@v3.28.0 -dir migrations create add_user_preferences sql`): `goose`
  is not installed and `make migrate-create` would install `@latest`. `go.mod`/`go.sum` are unchanged.
- **Checked**: `migrate up → down → up` on a copy of `data/app.db`; the local dev database `data/app.db` is migrated
  (`make migrate-up`), so the frontend can use the endpoint on the dev server.
- `docs/features.md` is left to the frontend developer (it describes page behavior that does not exist yet).

### Frontend (2026-10-17)
- **Library docs**: Context7 was not available in the session; the APIs were checked in the installed sources
  (`@tanstack/query-core`/`react-query` 5.104.0, `react-router` 1.170.41, `react-chartjs-2` 5.3.1). In 5.104
  `prefetchQuery`/`ensureQueryData` are deprecated in favor of `queryClient.query(options)`; `QueryFunctionContext`
  carries `client`; `keepPreviousData` returns the observer's last defined data while the new key is pending.
- **Q-D2 → prefetch in the app shell.** `Root.tsx` calls `prefetchReportPreferences(queryClient, user.id)` in an effect
  keyed by the user id, i.e. once per login / app load (`queryClient.query(...).catch(noop)`; a concurrent fetch of the
  page is deduplicated by query-core). The page usually opens from the cache without "Loading…".
- **Q-D3 → no `redraw` / `key`.** react-chartjs-2 5.3.1 matches datasets by `label`, rebuilds the dataset list from
  the new props (datasets that are gone are dropped), replaces the labels and assigns the options, then calls
  `chart.update()`. All datasets of `WeightCaloriesChart` have distinct labels, so in-place updates are enough.
- **Gate by structure.** `Report` renders the header and "Loading…" until `useReportPreferences()` returns settings, then
  mounts `ReportContent`, which holds the report queries. The `['reportData', …]` query does not exist before the
  settings, so no request with the default range is possible (stronger than `enabled`).
- **Applied once per mount**: the hook copies the settings into its own state the first time they are available (from
  the cache in the `useState` initializer, or by a guarded state update during render when the GET arrives later; a
  failed GET → defaults). Later refetches (mount refetch, window focus) only update the cache, which is what the next
  entry of the page shows ("another device picks it up on its next entry": with a warm cache, on the entry after the
  refetch, or on a reload). `retry: 1` on the settings GET, so a single blip does not open the page with defaults.
- **Range state**: the hook keeps the page values (resolved dates) and the saved range separately. Only a date change
  recomputes the saved range (`toSavedRange`: From fixed, To `"today"` when it equals the browser's today). A cleared
  input or a date the server would reject (e.g. a 5-digit year) keeps the last saved range and sends nothing.
  A change that does not alter the saved object (e.g. clearing a date) sends nothing.
- **Saver**: module-level (one per tab), so a save flushed on unmount completes after the page is gone and a later visit
  queues behind it. Debounce 800 ms; at most one PUT in flight, a newer state goes out after it completes (or after
  its own debounce); unmount → `flush()`; `pagehide` → forced send (the only case where two PUTs can overlap: the page
  is going away); `visibilitychange` → hidden also flushes (mobile browsers may never fire `pagehide`). Every PUT uses
  `fetch` `keepalive`, so a save in flight survives a reload. A failure → one `console.warn`, no retry, the next change
  sends the whole state.
- **Per-user safety of saves**: a pending save remembers the token of the session that made the change and is dropped
  if the token differs when it is sent (logout, or another login in the tab). Consequence: a change made less than
  0.8 s before a logout is not saved.
- **GET vs. pending save**: the settings `queryFn` returns the cached value instead of the response when a change was
  made while the GET was out or a save is still pending/in flight, so an older server state never replaces a newer
  local one in the cache.
- **Validation of the GET**: `preferencesService.getReportPreferences()` returns `unknown`; the hook's
  `parseReportPreferences` applies the server rules field by field (range all or nothing).
- ~~Out of scope, unchanged: logout does not clear the query cache~~ — superseded by the user's change below.

### Requirements change by the user (2026-10-17, after the frontend review)
1. **Logout now clears the client caches** (the user changed the earlier "out of scope" decision). `App.tsx`:
   - `endSession()` serves both logout paths: the menu logout (`onLogout` in the router context) and the automatic
     one when `['user']` (`GET /api/auth/me`) fails, e.g. an expired token. There is no other 401 handling in the
     frontend. Order: `authService.logout()` (the token goes first, so a pending Reports save of this session is
     dropped by the saver's token check, as before) → `queryClient.clear()` + `ingredientService.clearCache()` →
     `setIsAuthenticated(false)`.
   - `startSession()` (login, `Login`'s `onLogin`) resets the same caches again before the user is loaded.
   - Checked in query-core 5.104.0: `clear()` removes every query and mutation; removing a query destroys it and
     cancels its fetch, and a late response of a cancelled fetch is ignored, so queries in flight cannot repopulate
     the cache. A mutation in flight is not cancelled and its `onSuccess` may still write (e.g.
     `setQueryData(['profile'], …)`), so the reset at login covers that path.
   - After the reset `['user']` is empty, so after a login the app shows "Loading…" until `/api/auth/me` returns the
     new user (before, a cached previous user could render first). The settings key keeps the user id as a second
     guard.
   - The ingredient sessionStorage cache was **not** cleared on logout before (`clearCache()` was only called after
     a new ingredient was created); it is now.
2. **Midnight case.** `toSavedRange(dateFrom, dateTo, previousRange, toChanged, today)`: when only From changes and
   the saved To was `"today"` or not saved (`null`, rolling default), To stays `"today"` even if the page was opened
   on an earlier day; the page's To then moves to today's date as well, so the page shows what is saved. A saved fixed
   To stays fixed; a To the user changes follows the old rule (`"today"` only when it equals today). Checkbox, period
   and step changes do not touch the saved range (verified: `null` stays `null`, `"today"` stays `"today"`).
- **Not done in this step**: the browser checks on the dev server (the orchestrator asked not to start it). The pure
  functions and the saver were exercised with a throwaway Node script (bundled with rolldown, not committed): date
  rules and boundaries, field fallbacks, range resolution, debounce, single in-flight, queued latest state, one warning
  per failure, dropped cross-session save, forced `pagehide` send. The acceptance scenarios with the Network tab are
  left to QA.

## Phase History
| Date       | Phase       | Agent            | Notes                                                                                          |
|------------|-------------|------------------|------------------------------------------------------------------------------------------------|
| 2026-10-17 | analysis    | business-analyst | Specs written. Root causes:<br>(a) `Report.tsx:19-32` keeps all state in `useState` only.<br>(b1) weight mutations in `WeightHistory.tsx` do not invalidate `['weightGoalProgress']`.<br>(b2) the trend is fitted since the goal start by design.<br>Q1–Q6 for the user |
| 2026-10-17 | analysis    | business-analyst | Q1 = B (server, per user), Q2/Q4 as recommended, Q3 keep the trend, Q5/Q6 per recommendation. Specs revised for full-stack: `user_preferences` key → JSON (Q-D1), `GET`/`PUT /api/preferences/:key`, migration |
| 2026-10-17 | backend-dev | business-analyst | First phase set; waiting for the user's approval of the revised specs                          |
| 2026-10-17 | backend-dev | backend-developer | Specs approved by the user. Migration `20261017065036_add_user_preferences`, model, repository, service (`reports` key, validation, normalization, read fallbacks), handler `GET`/`PUT /api/preferences/:key` (JWT, 4 KB limit), unit + integration tests, `docs/api.md`, `docs/architecture.md`. Q-D1: generic store |
| 2026-10-17 | backend-review | backend-developer | Submitted for review. `go vet ./...`, `go test ./...`, `go test -race` (changed packages) green |
| 2026-10-17 | frontend-dev | backend-code-reviewer | Backend review clean, no report. `go build`, `go vet`, `go test ./...`, `go test -race` (changed packages) green; migrate down → up → down → up on a copy of `data/app.db` OK. The type-error deviation (field message instead of `Invalid request body`) follows `acceptance-criteria.md` and is accepted |
| 2026-10-17 | frontend-review | frontend-developer | `api/preferences.ts`, `types/preferences.ts`, `hooks/useReportPreferences.ts` (validate, resolve, debounced single-flight save, flush), `Report.tsx` gated on the settings with `keepPreviousData`, prefetch in `Root.tsx` (Q-D2), weight mutations invalidate goal progress / latest weight / health metrics, `docs/features.md`. No new strings. tsc, lint (0 warnings in the changed files), build, `go vet`, `go test ./...` green; dev server not started (by instruction) |
| 2026-10-17 | security-review | frontend-code-reviewer | Iteration 1: Frontend review clean, no report. tsc, lint (no warnings in the changed files), build green. Saver (debounce, single in-flight, queued latest state, unmount/visibilitychange/pagehide flush, token check), once-per-mount apply, GET/pending-save reconciliation, structural gate (no default-range request), retry once → defaults, invalidation keys and `queryClient.query()` (query-core 5.104.0) verified |
| 2026-10-17 | frontend-review | frontend-developer | Review passed; two changes requested by the user: logout/login clear the query cache and the ingredient cache (`App.tsx`), To stays "today" past midnight when only From changes (`useReportPreferences.ts`). `docs/features.md` updated. tsc, lint (27 pre-existing warnings, 0 errors), build, `go vet`, `go test ./...` green |
| 2026-10-17 | security-review | frontend-code-reviewer | Iteration 2 clean, no report. `endSession()` order (token → `queryClient.clear()` + ingredient cache → login screen) for both logout paths, `startSession()` clear on login, the saver drops a pending save of the old session, in-flight queries are cancelled by `clear()`; To kept "today" past midnight when only From changes. tsc, lint (no new warnings), build green |
| 2026-10-17 | qa-review | security-reviewer | Security review clean, no report. Routes JWT-only (`RequireAuth` reads only the `Authorization` header; `X-API-Key` → 401), user only from the token, allow-list before any DB access, strict per-field validation, 4 KiB limit after auth (Content-Length and streamed bodies), stored value is the server's re-encoding with `v`, no stored values or internal errors in responses/logs, SQL parameterized. No CSRF angle (no cookie auth, CORS limited to `AppURL`). No rate limit needed (bounded upsert, as other JWT write routes). Frontend: user id in the query key, `clear()` + ingredient cache on logout/login, saves dropped on token change, nothing new in local/sessionStorage, CSP unchanged. Dependencies unchanged. Dev server probed unauthenticated only: 401 for GET/PUT, API key, unknown key |
| 2026-10-17 | ba-acceptance | qa-engineer | QA review clean, no report. `go vet ./...`, `go test -count=1 ./...`, `go test -race` (preferences service/handler, `internal/server`, repositories) green; coverage: service 95.7%, handler 88.9% (only unreachable marshal/DB-load branches uncovered); migrate up → down → up on a copy of `data/app.db` OK. Every backend criterion maps to `services/preferences/service_test.go`, `handlers/preferences/handler_test.go` or `server/preferences_test.go` (boundaries incl. 2024-02-29 / 2026-02-29, "today" only for To, half range, wrong types, 4096 → 200 / 4097 → 413 unchanged, unknown key/fields, corrupt stored values, U1/U2 isolation, 401 without JWT and with a real API key only, generic 500). Frontend criteria verified by reading the code; tsc, lint (no new warnings), build green. Dev server probed unauthenticated only: 401. **Manual browser check still open**: logout → login as U2 → login as U1 in the same tab (made after the user's browser check) |
| 2026-10-17 | devops | business-analyst | Acceptance passed. Every criterion verified against the code and QA's results is met; the user's browser check confirmed persistence across reload, the trend redraw after weight add/edit/delete and the previous chart kept while loading. `requirements.md` updated: C1 (cache reset on logout/login) and C2 (midnight To) recorded as decisions, B2 aligned (wrong types → 400 naming the field). Unchecked: the same-tab logout/login browser check (pending, user), the production migration (DevOps, by agreement with the user), 2 post-release checks |

## Review Iterations
- **Backend iteration 1**: 2026-10-17 — clean (no issues). Notes for later steps: `requirements.md` B2 still says wrong types → `Invalid request body`, while `acceptance-criteria.md` and the implementation name the field (BA to align at acceptance)
- **Frontend iteration 1**: 2026-10-17 — clean (no issues). Optional suggestions returned to the orchestrator: keep `"today"` for To when only From changes after midnight in a tab opened the day before
- **Frontend iteration 2**: 2026-10-17 — clean (no issues). Logout/login cache reset and the midnight To rule verified; "U2 never sees U1's settings" now holds in code (QA to confirm in the browser). Residual, not blocking: a profile/language mutation of U1 still in flight after U2 has logged in could merge U1's data into U2's `['user']` cache through its `onSuccess` `setQueryData`
- **Security iteration 1**: 2026-10-17 — clean (no issues). Residual, not blocking (pre-existing, same tab and same device only): a mutation of the previous session still in flight after the next login can write its response into the new session's cache through `onSuccess` `setQueryData`
- **QA iteration 1**: 2026-10-17 — clean (no issues). Non-blocking suggestions: an integration test for a real database failure (e.g. drop `user_preferences` through `client.db`, expect the generic 500 without `no such table`); unit cases `1900-02-29` (not leap), `2000-02-29`, `date_to` `1899-12-31`, `"Today"`. Still manual: the same-tab logout/login scenario, and the midnight To rule (verified in code only)
- **BA acceptance**: 2026-10-17 — passed; status → `devops`. Open: the same-tab logout/login manual check, the
  production migration by agreement, the post-release checks.
