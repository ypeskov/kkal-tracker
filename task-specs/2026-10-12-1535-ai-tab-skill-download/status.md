# Task Status: Settings → AI: download the food logging skill as a ready-to-install .zip

## Current Phase
**Phase**: `done`
**Assigned to**: —
**Updated**: 2026-10-12

Chain (full-stack): BA → Backend Dev → Backend Review → Frontend Dev → Frontend Review → Security Review → QA Review →
BA Acceptance → DevOps.
The user answered Q1–Q7 (recorded as D1–D7 in `requirements.md`) and approved the specs on 2026-10-12.
Q-D1–Q-D4: decided by the backend developer (see Developer decisions).
BA acceptance passed: all criteria that code and tests can show are met. **Pending manual verification** (12 checks,
unchecked in `acceptance-criteria.md`):
- Frontend: tab order/icon/default, AI tab content, button pending state, saved file name, success message and API Keys
  refresh, error message, 429 message, 4 languages
- Integration: the dev zip running `kkal_client.py ingredients`, claude.ai upload and logging (known risk: description
  length, D5), the deleted key giving 401
- Air rebuild on skill file changes

Notes: the isolation criterion is satisfied through `/api/v1/ingredients` instead of `/api/v1/data` (same
`RequireAPIKey` user resolution). The smoke test comment overstates its download step; this is cosmetic.

## Developer decisions
- **Q-D1 (layout)**: `skills/embed.go` (package `skills`, two `//go:embed` file directives, string vars; a test
  checks the directives); transformation in `internal/skill` (`Render`, `WriteDir`, name validation); service
  `internal/services/skill`; handler `internal/handlers/skill`; generator `cmd/skillgen` + `make skill-dev [URL=...]`.
- **Q-D2 (cleanup)**: render → create key → zip in memory → respond. A zip failure deletes the key (a delete failure is
  logged at error level); render/URL failures happen before any key exists. Zip has no directory entries; `api_key`
  is mode 0600, the other files 0644.
- **Q-D3 (rate limit)**: adopted, per user (`user_id` after `RequireAuth`): burst 5, then 1 per minute; 429 with the
  standard "Too many requests, please try again later". Shared helper with the per-IP limiter in `security.go`.
- **Q-D4 (Air)**: `include_file` with the two embedded files (exact root-relative paths, checked in the Air source);
  `md`/`py` stay out of `include_ext`.
- Finding: an `APP_URL` with a trailing `/` already makes the server panic at startup (Echo CORS rejects an origin
  with a path), so the trim in the skill only matters for the generator and unit tests. Not changed here.

## Phase History
| Date       | Phase    | Agent            | Notes                                                        |
|------------|----------|------------------|--------------------------------------------------------------|
| 2026-10-12 | analysis | business-analyst | Specs written, open questions Q1–Q7 for the user, Q-D1–Q-D4 for the developers |
| 2026-10-12 | analysis | business-analyst | Answers applied: defaults for Q1–Q4, Q6, Q7 accepted; Q5 rejected (description stays as is, recorded as a known risk) |
| 2026-10-12 | backend-dev | backend-developer | Specs approved by the user; started implementation |
| 2026-10-12 | backend-review | backend-developer | Backend done (endpoint, embed, Go generator, docs, tests); go vet, go test, go test -race green |
| 2026-10-12 | frontend-dev | backend-code-reviewer | Backend review clean: no issues; go vet, go test, race on affected packages green |
| 2026-10-12 | frontend-review | frontend-developer | Settings AI tab (AiTab.tsx, api/skills.ts, downloadBlob moved to utils/download.ts), settings.ai.* in 4 locales, docs/features.md; tsc, lint (0 errors, no new warnings), build green; dev server not run (stopped by the user) |
| 2026-10-12 | security-review | frontend-code-reviewer | Frontend review clean: no issues; tsc, lint (0 errors), build green; Trans renders <code> via a components map (no innerHTML), CSP unchanged |
| 2026-10-12 | qa-review | security-reviewer | Security review clean: no issues; JWT-only route (no cookie auth, API key rejected), per-user limiter after auth and fail-closed, key only in the zip body, explicit 2-file embed, api_key excluded from the build context, no dependency changes |
| 2026-10-12 | ba-acceptance | qa-engineer | QA review clean: no issues. go vet, go test, go test -race (-count=1) green; coverage handlers/skill 100%, services/skill 93.9%, internal/skill 93.2%, server 100%, cmd/skillgen 0% (exit codes checked by hand: bad URL 1, two args 2, missing dir 1); `make skill-dev` idempotent, dev api_key untouched; real api_key files absent from the built binary; tsc, lint (0 errors, changed files clean), build green. Browser and claude.ai checks left for manual verification |
| 2026-10-12 | devops | business-analyst | Acceptance passed: all code/test-verifiable criteria met; 12 browser/dev-server/claude.ai checks pending manual verification |
| 2026-10-12 | done | devops-engineer | Released v5.13.0 with deploy.sh (image ypeskov/kcal-tracker:5.13.0 from master); prod ENVIRONMENT=production, APP_URL=https://kcal.peskov.info; build context and image contain no api_key; pod Running, site 200, AI tab strings (4 locales) in the served JS, POST /api/skills/food-log/download without auth → 401. The 12 manual checks are still pending |

## Review Iterations
- **Iteration 1**: 2026-10-12 — backend review clean (no report)
- **Iteration 1**: 2026-10-12 — frontend review clean (no report)
- **Iteration 1**: 2026-10-12 — security review clean (no report)
- **Iteration 1**: 2026-10-12 — QA review clean (no report)
- **Iteration 1**: 2026-10-12 — BA acceptance passed; on to `devops` (12 manual checks pending)
