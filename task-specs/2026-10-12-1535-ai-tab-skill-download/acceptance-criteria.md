# Acceptance Criteria: Settings → AI: download the food logging skill as a ready-to-install .zip

Based on the user's decisions D1–D7 of 2026-10-12 (`requirements.md`).

BA acceptance 2026-10-12: every criterion that the code and the automated tests can show is met (checked). Criteria
that need a browser, the dev server or claude.ai are left unchecked and marked **pending manual verification**.
Test results come from the QA run of 2026-10-12: `go vet`, `go test`, `go test -race`, tsc, lint and build all green.

## Backend
### Endpoint
- [x] `POST /api/skills/food-log/download` without an `Authorization` header → 401, and the user's key count is unchanged
  (`TestSkillDownloadRequiresAuth`)
- [x] The same request with an invalid or expired JWT → 401, and no key is created (`TestSkillDownloadRequiresAuth`:
  an empty, an invalid and an expired token; the key count is unchanged)
- [x] With a valid JWT → 200 with these headers (`TestSkillDownloadInstallsAWorkingSkill`,
  `TestSkillDownloadMatchesTheRepository`, handler `TestDownloadFoodLogResponse`):
  - `Content-Type: application/zip`
  - `Content-Disposition: attachment; filename="kkal-tracker-food-log.zip"` (production config) or
    `filename="kkal-tracker-food-log-dev.zip"` (`ENVIRONMENT=development`)
  - `Cache-Control: no-store`
- [x] `GET` on the same path does not create a key (405 or the SPA fallback; no handler runs). Only `POST` is
  registered; the test checks the error status and the unchanged key count
- [x] Two consecutive downloads create two different keys. Both work, and earlier keys of the user stay active (D1).
  The test uses both skill keys and the earlier agent key on `/api/v1/ingredients` → 200
- [x] A failure while building the zip after the key was created → 500 `{"message": "Failed to prepare the skill"}`
  (no internal error text), and the created key no longer exists (verified with a failing zip writer or service fake).
  Service: `TestDownloadFoodLogDeletesTheKeyWhenZippingFails` (failing `writeZip`, no keys left). Handler:
  `TestDownloadFoodLogHidesInternalErrors` (exact generic body, no `Content-Disposition`)
- [x] A failure creating the key → 500 with the same generic message, and nothing is written to the response body.
  `TestDownloadFoodLogKeyCreationFailure` and the handler test: the body is only the generic JSON message, with no zip
  and no attachment header
- [x] Rate limit, if adopted (Q-D3): requests beyond the limit for one user → 429 with the standard message. Another
  user is not affected. Adopted per user (burst 5, then 1/min); `TestSkillDownloadRateLimitPerUser` also checks that a
  limited request creates no key

### Zip contents
- [x] The zip contains exactly 3 files: `<name>/SKILL.md`, `<name>/scripts/kkal_client.py`, `<name>/api_key`, where
  `<name>` is `kkal-tracker-food-log` (production) or `kkal-tracker-food-log-dev` (development). No other entries, and
  none outside `<name>/`. The integration helper `skillKey` compares the sorted entry list exactly; there are no
  directory entries
- [x] `<name>/api_key` holds the raw key plus `\n`. Used as `X-API-Key` on `GET /api/v1/ingredients`, it returns 200
  with the downloading user's ingredients
- [x] The key belongs to the downloading user: user B's `GET /api/api-keys` doesn't list it, and with it `/api/v1/data`
  returns user A's data only.
  **Note:** the test checks the same isolation on `GET /api/v1/ingredients` instead of `/api/v1/data`: A's ingredient
  is visible, B's is not. This meets the intent of the criterion. Every `/api/v1` route takes the user from the key
  through the same `RequireAPIKey` middleware, so the property tested (the key resolves to A and reaches only A's data)
  is the one the criterion asks for. B's key list not showing A's key is tested directly
- [x] The key is listed in `GET /api/api-keys` with name `Claude skill YYYY-MM-DD HH:MM UTC` (the current UTC time) and
  `expires_at: null` (D2). The name is checked against the time window of the request, plus `KeyName` and the service
  test with a non-UTC clock
- [x] `kkal_client.py` in the zip keeps `API_KEY = "XXXXXXXXX"`, and `DEFAULT_BASE_URL` equals `APP_URL`
- [x] SKILL.md in the zip has ``- Base URL: `<APP_URL>` ``
- [x] `APP_URL=https://example.test/` (trailing slash) → `https://example.test` in both files. Covered by
  `TestRenderSubstitutesTheBaseURL` (both variants) and the service test (dev with a trailing `/`). The server itself
  refuses to start with such an `APP_URL` (Echo CORS), a finding recorded in `status.md`
- [x] Production config (`ENVIRONMENT=production`, `APP_URL=https://kcal.peskov.info`) → SKILL.md and `kkal_client.py`
  in the zip are byte-identical to `skills/kkal-tracker-food-log/` in the repository (`TestSkillDownloadMatchesTheRepository`)
- [x] Development config (`ENVIRONMENT=development`, `APP_URL=https://dev-kcal.peskov.info`) → SKILL.md and
  `kkal_client.py` in the zip are byte-identical to the committed `skills/kkal-tracker-food-log-dev/` (same test)
- [x] In both variants, SKILL.md frontmatter: `name` ≤ 64 characters of `[a-z0-9-]` and equal to the folder name
  (unit test). The `description` has no length check (D5). Covered by `TestFrontmatterNameOfBothVariants` and
  `TestValidateName`; `Render` also enforces it at runtime

### Secrets and logging
- [x] The raw key doesn't appear in any log line during a successful or a failed download (a test with a captured
  logger). Only the prefix may be logged, as `CreateKey` already does. Covered by service tests (success and zip
  failure) and the wired-server test with a captured debug logger (checks the key and its non-prefix part)
- [x] The raw key doesn't appear in any response header (`TestSkillDownloadInstallsAWorkingSkill`)
- [x] The binary embeds only `skills/kkal-tracker-food-log/SKILL.md` and `skills/kkal-tracker-food-log/scripts/kkal_client.py`.
  With a real `skills/kkal-tracker-food-log/api_key` present on disk, the zip's `api_key` still holds only the new key,
  and the embed directive lists files explicitly. Covered by `skills/embed.go` (two single-file directives) and
  `TestEmbedListsFilesExplicitly`. QA confirmed that the real `api_key` contents are absent from the built binary
- [x] `.dockerignore` contains `skills/*/api_key`

### Single source of the dev transformation (D4)
- [x] `skills/sync_dev_skill.py` is deleted. The Go generator command (documented in README) regenerates
  `skills/kkal-tracker-food-log-dev/` and leaves an existing dev `api_key` untouched. An optional URL argument works as
  in the old script. Covered by `cmd/skillgen` and `make skill-dev [URL=...]`; `TestWriteDirKeepsOtherFiles`; QA ran
  `make skill-dev` with the dev `api_key` untouched and checked the exit codes by hand
- [x] No references to `sync_dev_skill.py` remain outside `task-specs/`: `README.md` ("Agent skills", formerly line
  244), `docs/architecture.md` (formerly line 38) and any others name the Go generator instead (grep: no matches)
- [x] Running the generator on the current tree produces no diff in `skills/kkal-tracker-food-log-dev/` (QA)
- [x] A test fails if the committed dev copy differs from the generator output for `https://dev-kcal.peskov.info`
  (`TestCommittedDevCopyIsUpToDate`)
- [x] When an expected anchor (frontmatter layout, title, Base URL line, `DEFAULT_BASE_URL`) is missing or appears more
  than once, the transformation returns an error (generator exits non-zero, download → generic 500). Unit tests cover
  this: `TestRenderRequiresEveryAnchorExactlyOnce` (11 cases), and `TestDownloadFoodLogFailsWithoutCreatingAKey`
  (broken source → error, no key). The handler maps any service error to the generic 500

### Skill text (D5, D7)
- [x] The frontmatter `description` of the production SKILL.md is unchanged from `develop` before this task (D5)
  (QA: identical to HEAD)
- [x] SKILL.md "API basics" and the client's "not configured" message name Settings → AI (download the skill) as the way
  to get a configured skill. Creating a key in Settings → API Keys and saving it to `api_key` is mentioned as the
  fallback. The dev copy is regenerated accordingly. Covered in SKILL.md line 24 and the configuration comment and
  `fail` message of `kkal_client.py`; the dev copy matches (golden test)

### Quality
- [x] `go vet ./...` and `go test ./...` pass (after `make build-frontend`), including the race detector (QA). Coverage:
  handlers/skill 100%, services/skill 93.9%, internal/skill 93.2%
- [x] Integration test on the wired server: login → download → unzip → `api_key` works on `/api/v1/ingredients`
  (`TestSkillDownloadInstallsAWorkingSkill`). The smoke test also gained a download step (zip signature and file name).
  Its comment "the downloaded zip carries a working API key" overstates what that step checks; the key is verified in
  the dedicated test. This is cosmetic and not blocking

## Frontend
Code for all items below exists (`Settings.tsx`, `AiTab.tsx`, `api/skills.ts`, `utils/download.ts`, `settings.ai.*`
with the same 15 keys in 4 locales). Browser behavior has not been observed yet.
- [ ] Settings shows three tabs in this order: Export, API Keys, AI. The AI tab has an icon, and Export stays the
  default. **Pending manual verification** (code: order export/apiKeys/ai, `Bot` icon, default `export`)
- [ ] The AI tab shows: **pending manual verification**
  - an explanation of the skill
  - the Download button
  - the note that each download creates a new API key, and that keys from earlier downloads stay active and can be
    deleted on the API Keys tab (D1)
  - the install steps accepted in D6 (code execution and network access to the current site's domain; Settings →
    Capabilities → Skills → Upload skill; tell Claude what you ate), plus the Claude Code line (`~/.claude/skills/`)
  - the amber secret warning
- [ ] Clicking Download sends one `POST /api/skills/food-log/download` with the Bearer token. While it runs, the button
  is disabled and shows the pending label, so a double click sends one request. **Pending manual verification**
- [ ] On success the browser saves a file named from `Content-Disposition` (`kkal-tracker-food-log.zip` on prod,
  `kkal-tracker-food-log-dev.zip` on dev). If the header is missing, the fallback is `kkal-tracker-food-log.zip`.
  **Pending manual verification**
- [ ] On success a success notification appears. Switching to the API Keys tab shows the new key without a page
  reload. **Pending manual verification** (code: `['apiKeys']` invalidated on settle)
- [ ] On a 500 or a network error a localized generic error notification appears, and no file is saved. Server text
  is not shown. **Pending manual verification** (code: fixed `settings.ai.error`, and a non-zip body is rejected)
- [ ] On a 429 (if rate limited) the localized "try again later" message appears. **Pending manual verification**
  (code: `settings.ai.tooManyRequests` on status 429)
- [ ] All AI tab strings are in `settings.ai.*` in `en_US`, `uk_UA`, `ru_UA`, `bg_BG`, with no hard-coded text.
  Switching the language changes every string on the tab. **Pending manual verification** of the language switch
  (QA: the same 15 keys in all 4 locales; no hard-coded strings in `AiTab.tsx`)
- [x] Tailwind utilities only, layout per the `frontend-style` skill, no inline `<script>`, CSP unchanged (`<code>`
  rendered through the `Trans` components map, no `innerHTML`; `security.go` untouched by this task per the security review)
- [x] `npx tsc --noEmit -p tsconfig.app.json`, `npm run lint`, `npm run build` pass (QA, lint 0 errors)

## Integration
- [ ] Dev server (`ENVIRONMENT=development`, `APP_URL=https://dev-kcal.peskov.info`): Settings → AI → Download gives
  `kkal-tracker-food-log-dev.zip`. Unzipped, `python3 kkal-tracker-food-log-dev/scripts/kkal_client.py ingredients`
  (no env vars) lists the user's ingredients from the dev server. **Pending manual verification**
- [ ] The downloaded zip uploads to claude.ai as a skill without validation errors. After that, "I ate two eggs"
  in a chat stores an entry on the server the zip points to (manual check by the user). Known risk: the production
  description is longer than the documented 1024-character limit and was kept by the user's decision (D5). If the upload
  rejects it, record that as a follow-up task, not as a failure of this task. **Pending manual verification**
- [ ] Deleting the downloaded key in Settings → API Keys makes the client from the zip fail with HTTP 401.
  **Pending manual verification**
- [x] Docs updated: `docs/api.md` (route), `docs/features.md` (AI tab), `docs/architecture.md` (skills embedding,
  generator), `docs/infrastructure.md` (Air, `.dockerignore`), README "Agent skills", `AGENTS.md` Secrets rule
  (also `AGENTS.md` Commands: `make skill-dev`)

Also pending manual verification (requirement Backend §9, no separate criterion): Air rebuilds when
`skills/kkal-tracker-food-log/SKILL.md` or `scripts/kkal_client.py` changes (`include_file` is configured in `.air.toml`).
