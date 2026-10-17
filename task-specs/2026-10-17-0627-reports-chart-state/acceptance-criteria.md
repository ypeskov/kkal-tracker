# Acceptance Criteria: Reports page: remember the chart settings (per user, on the server) and keep the trend line up to date

Decisions Q1–Q6 and the user's changes C1 (logout/login clear the client caches) and C2 (midnight rule for To), all of
2026-10-17, are recorded in `requirements.md`, "Decisions".

Setup:
- Today is T (browser local date).
- User U1 has a goal with a start date S, a target date and a target weight. U1 has weigh-ins after S and calorie
  entries in the last 60 days.
- User U2 is another account.
- Frontend checks run on the dev server in a desktop browser; mobile layout is unchanged.

`R` below is the request/response object of `/api/preferences/reports`.

## Backend

### Migration
- [x] Exactly one new Goose migration creates the preferences storage (Q-D1, recorded in `status.md`). Its `Down`
      drops it. No existing table changes
- [x] `make migrate-up` then `make migrate-down` then `make migrate-up` succeed on a local copy of a current database
- [x] The server does not apply migrations on start (unchanged)

### `GET /api/preferences/reports`
- [x] A new user → 200 `{"date_from": null, "date_to": null, "show_weight": true, "show_calories": true,
      "show_goal_projection": true, "period": "daily", "step_interval": 1}`
- [x] After a successful PUT → 200 with exactly the object the PUT returned
- [x] A stored value that cannot be decoded (written directly into the test database) → 200 with all defaults; a Warn
      log with `user_id` and the key, without the raw value
- [x] A stored value with one invalid field (e.g. `period: "hourly"`) → 200 with that field defaulted, the others
      kept. An invalid stored range → both dates `null`

### `PUT /api/preferences/reports`
- [x] `{"date_from": "2026-09-01", "date_to": "today", "show_weight": true, "show_calories": false,
      "show_goal_projection": true, "period": "weekly", "step_interval": 1}` → 200 with the same object; a following
      GET returns it
- [x] A second PUT with other values replaces the first completely
- [x] `date_from`/`date_to` both `null` with other fields set → 200, and GET returns `null` dates
- [x] Unknown extra fields (e.g. `"foo": 1`) → 200, ignored, not returned
- [x] The stored row holds the server's own JSON encoding with a schema version, not the request body verbatim
      (checked in the test or by reading the code)
- [x] Each of the following → 400 with a message naming the field, and the stored value is unchanged:
  - [x] `date_from` `"2026-02-30"`, `"2026-13-01"`, `"26-09-01"`, `""`, `"today"`, `"1899-12-31"`, `"2101-01-01"`
  - [x] `date_to` `"tomorrow"`, `"2026-02-30"`
  - [x] only one of `date_from` / `date_to` is `null`
  - [x] `period` `"hourly"` or missing
  - [x] `step_interval` `7`, `"5"` (a string) or missing
  - [x] a checkbox field missing, or `"yes"` instead of a boolean
- [x] Boundaries pass: `1900-01-01`, `2100-12-31`, `2026-02-28` and `"today"` as `date_to`
- [x] `date_from` later than `date_to` (both dates) → 200 (accepted, as the UI allows it)
- [x] Malformed JSON, or a body that is not a JSON object (`null`, `[]`, a scalar, empty) → 400
      `Invalid request body`
- [x] A body over 4 KiB → 413, and the stored value is unchanged

### Access and isolation
- [x] Without a JWT → 401; with only an `X-API-Key` → 401 (GET and PUT)
- [x] U1 PUTs. U2's GET → the defaults. U2's PUT does not change U1's GET
- [x] `GET` / `PUT /api/preferences/<anything other than reports>` → 404 `{"message": "Unknown preference"}`
- [x] A database failure → 500 with the generic message (`Failed to load preferences` / `Failed to save preferences`),
      logged with the injected logger. No internal error text in any response

### Code and tests
- [x] Handler → service → repository layering. SQL only in `internal/repositories/queries.go`. No direct `log/slog`
      use for logging (the injected logger)
- [x] Unit tests (validation boundaries, normalization, read fallbacks, unknown key) and integration tests on the wired
      server cover the criteria above
- [x] `go vet ./...`, `go test ./...` (after `make build-frontend`), `go test -race` on the changed packages pass

## Frontend

### Loading
- [x] Opening `/reports` (reload, a direct URL or menu navigation) never shows the default range or default checkboxes
      before the saved ones. Until the settings arrive, the page shows "Loading…"
- [x] Network tab: the first `GET /api/reports/data` after opening the page carries the restored `from`/`to`. No
      request with the default 30-day range precedes it
- [x] `GET /api/preferences/reports` failing (simulated, e.g. by stopping the backend route or by offline mode for that
      request) → the page opens with the defaults and works. No error message, no blank page
- [x] A response with an invalid field (e.g. intercepted to `period: "hourly"`) → that field uses its default, the
      others apply
- [x] While the user is on the page, a window refocus does not change the controls, even if the settings were changed
      on another device meanwhile

### Defaults and range persistence (Q2, C2)
- [x] U1 with nothing saved: From = T − 30 days, To = T, all three checkboxes on, daily, step 1, Charts tab, as today
- [x] U1 sets From = 2026-09-01 (To stays T) and reloads → From = 2026-09-01, To = T. The PUT body had
      `"date_to": "today"`
- [x] Same saved state, a later day (simulated by changing the system date) → From = 2026-09-01, To = that day
- [x] U1 sets From = 2026-08-01, To = 2026-08-31 and reloads → exactly 2026-08-01 … 2026-08-31, also on a later day
- [x] With nothing saved, changing only a checkbox → the PUT has `null` dates. After the reload the range is the
      rolling default
- [x] Clearing a date input does not save. After the reload the last complete range is shown
- [x] The Weight History tab uses the same restored range (Q5)
- [x] SPA navigation: change the range, go to Dashboard via the menu immediately (within the debounce window), come
      back → the change is kept, and it reached the server (Network tab)
- [x] C2: a page opened on day D with the saved To `"today"` (or nothing saved) is left open past midnight. Changing
      only From on D+1 saves `"date_to": "today"`, and the To input shows D+1. With a saved fixed To, the fixed To
      stays

### Checkboxes, period, step, tab (Q4)
- [x] Calories off, reload → Calories off, no calorie line or right axis
- [x] Weight off, reload → Weight off, no weight, trend or target lines
- [x] Both off, reload → the "select at least one metric" message
- [x] "Show goal projection" off, reload → off, and the x axis ends at the last data date
- [x] A goal without a target date hides the projection checkbox, and its saved value is kept. Once a target date is
      set again, the checkbox shows the saved value
- [x] Period weekly, reload → weekly. Daily with step 10, reload → daily, 10
- [x] Switch to Weight History and reload → the page opens on Charts (the tab is not saved)

### Saving behavior
- [x] Toggling three checkboxes quickly, or typing a date, produces one PUT with the final state about 1 s after the
      last change (Network tab)
- [x] Never two PUTs in flight at once. With a throttled network, the final server state equals the final UI state
- [x] A change, then a reload after more than the debounce interval → kept
- [x] PUT failing (simulated 500 or offline) → the controls and the chart keep working, no toast/alert, at most one
      console warning per failed request. The next change sends the full current state again and succeeds once the
      backend is back

### Cross-device and per user (C1)
- [x] U1 changes the settings in browser A. Opening `/reports` in browser B (or a private window) as U1 shows the same
      settings
- [ ] U1 changes the settings, logs out, U2 logs in **in the same tab** → U2 sees U2's own settings (or the defaults),
      never U1's, not even briefly. U1 logs in again → U1's settings — **pending manual check by the user** (verified
      in code, see the notes)
- [x] C1: both logout paths (the menu, and the automatic one when `/api/auth/me` fails) remove the token, then clear
      the TanStack Query cache and the ingredient `sessionStorage` cache, then show the login screen. Login clears both
      caches again
- [x] C1: a Reports change still pending at logout is not sent for the next account (the save is bound to the token
      of the session that made it)

### Trend line refresh (F4)
- [x] On the Weight History tab, add a weigh-in dated after S with a value well off the trend. Switch to Charts → the
      new point is drawn and the trend line changed (equal to what a full reload shows), with no reload and no
      refocus
- [x] The same for editing and deleting a weigh-in after S
- [x] Each of the three mutations triggers `GET /api/profile/goal` (Network tab) while the Reports page is open
- [x] After a weight change, opening the Profile page shows the new latest weight and the recalculated BMI

### Dynamic redraw (F5)
- [x] Changing From or To keeps the previous chart visible while loading, then shows the new range with the trend line
      evaluated on its dates. "Loading…" appears only on the first load of the page
- [x] Changing Period, Step or a checkbox redraws immediately. No dataset of the previous state stays visible
- [x] The statistics card matches the range shown

### Code
- [x] An API service `web/src/api/preferences.ts`. The settings logic is in one typed hook that `Report.tsx` consumes.
      The range resolution is a pure function
- [x] The query key contains the user id
- [x] No new user-facing strings without `t()` in all 4 locales; Tailwind only; no CSP change
- [x] `npx tsc --noEmit -p tsconfig.app.json`, `npm run lint`, `npm run build` pass in `web/`

## Docs
- [x] `docs/api.md` describes `GET` / `PUT /api/preferences/:key`:
  - the `reports` key only;
  - the fields and their rules, including `"today"`;
  - the defaults on GET;
  - 400 / 404 / 413;
  - full replacement on PUT
- [x] `docs/features.md`, "Reports & Analytics":
  - the remembered settings (range with the rolling To, checkboxes, period, step);
  - per user on the server, the same on every device;
  - the trend refresh after weight changes
- [x] `docs/features.md`, "Persistent Sessions": logout and login clear the client caches (C1)
- [x] `docs/architecture.md`: the `preferences` domain and the new model are listed

## Integration
- [x] End to end on the dev server:
  1. U1 logs in and opens Reports.
  2. U1 sets From = S, turns Calories off and sets Period = weekly.
  3. U1 adds a weigh-in on the Weight History tab and returns to Charts. The new point and the updated trend line are
     shown, with no reload.
  4. U1 reloads. From = S, To = today, Calories off, weekly, and the same trend.
  5. U1 opens the page in another browser. The same settings.
- [ ] Release: the migration was applied to production in the way agreed with the user before the feature is
      checked. Recorded in `status.md` — **pending (DevOps, with the user's agreement)**
- [ ] After the release, on production: change the range and a checkbox, reload → kept; another device → the same
      — **pending, post-release**
- [ ] After the release, on production: `GET /api/preferences/reports` without a JWT → 401 — **pending, post-release**

## Acceptance notes (2026-10-17, business-analyst)
- **Sources**:
  - the code, read by the BA;
  - QA's results in `status.md`: `go vet`, `go test -count=1 ./...` and `-race` green; coverage of the service 95.7%
    and the handler 88.9%; migrate up → down → up OK;
  - the developers' and reviewers' notes;
  - the user's browser check on the dev server: the settings persist across reload, the trend redraws after a weight
    add/edit/delete, and the previous chart stays while loading.
  - The BA ran no commands.
- **Backend, where each part is**:
  - migration `migrations/20261017065036_add_user_preferences.sql` (Down drops the table);
  - `internal/services/preferences/` (`service.go` allow-list `definitions`, `reports.go` `readReports`, `fields.go`
    exact-name reader with the date bounds);
  - `internal/handlers/preferences/handler.go` (generic 500s, 404 / 400 / 413 mapping);
  - `internal/repositories/user_preference.go` + `queries.go` (upsert);
  - wiring `internal/server/server.go:190-193` (`RequireAuth` + `BodyLimit(4 KiB)`).
- **Backend tests**: `internal/server/preferences_test.go`, `internal/services/preferences/service_test.go` and
  `internal/handlers/preferences/handler_test.go` cover every backend criterion, including:
  - the literal stored encoding with `"v":1`;
  - 4096 bytes → 200, 4097 → 413;
  - corrupt stored values;
  - U1/U2 isolation;
  - 401 with a real API key.
- **Frontend, where each part is**:
  - `web/src/hooks/useReportPreferences.ts`:
    - `parseReportPreferences`, `resolveReportSettings` and `toSavedRange` are pure functions;
    - the module-level saver: 800 ms debounce, one PUT in flight, flush on unmount / `visibilitychange` / `pagehide`,
      `keepalive`, one `console.warn` per failure, the token check;
    - the query key `['preferences','reports',userId]`;
  - `Report.tsx` mounts `ReportContent` (and with it the `['reportData', …]` query) only when the settings are there,
    so no default-range request can exist;
  - `keepPreviousData` on the report data;
  - the prefetch in `Root.tsx`;
  - `WeightHistory.tsx` `invalidateWeightQueries` (`weightHistory`, `reportData`, `weightGoalProgress`,
    `latestWeight`, `healthMetrics`);
  - `App.tsx` `endSession` / `startSession` (C1).
- **Criteria verified in code only.** Beyond what the user checked in the browser, these were verified by reading the
  code and by QA's code review; there is no frontend test runner (Q6):
  - the later-day and midnight (C2) cases: `resolveReportSettings`, and `toSavedRange` with `keepsToday`;
  - the failure and throttling cases of the saver;
  - the refocus case: state applied once per mount, and `queryFn` keeps the cached value while a save is pending;
  - the cross-browser case: server storage plus the GET-after-PUT integration test.
  The developer also exercised the pure functions and the saver with a throwaway Node script.
- **Open (unchecked)**:
  1. the same-tab logout → U2 login → U1 login check in the browser. In code: `endSession` removes the token before
     `queryClient.clear()`, `startSession` clears again, the query key carries the user id, and the saver drops saves
     of another token. Pending the user's manual check;
  2. applying the migration to production by agreement with the user (DevOps);
  3. the two post-release production checks.
- **Accepted design details**:
  - wrong types → 400 naming the field (B2 aligned);
  - a missing date counts as `null`;
  - a change made less than ~0.8 s before a logout is not saved (C1);
  - the residual in-flight-mutation case after a re-login is pre-existing and out of scope.
