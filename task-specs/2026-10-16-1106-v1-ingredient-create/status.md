# Task Status: Create an ingredient through the external API (`POST /api/v1/ingredients`) and the `add-ingredient` skill command

## Current Phase
**Phase**: `devops`
**Assigned to**: `devops-engineer`
**Updated**: 2026-10-16

Chain (backend-only, no frontend changes): BA → Backend Dev → Backend Review → Security Review → QA Review →
BA Acceptance → DevOps.
Q1–Q4 answered by the user (2026-10-16), all as recommended; recorded in `requirements.md`, "Decisions", and folded
into `acceptance-criteria.md`. The user approved the final specs (workflow step 2.5) on 2026-10-16.

## Developer decisions (Q-D1–Q-D3)
- **Q-D1, service shape → as recommended.** New `AddIngredient(*AddIngredientRequest) (*models.UserIngredientUsage,
  error)` with pointer fields, so a missing `name` / `kcal_per_100g` is visible. Validation `validateAddRequest` shares
  `cleanName` and `validateValues` with `validatePatchRequest` (extracted from it; the same messages and the same
  `internal/food` calls). One create path `createIngredient` (`checkDuplicateName` → `CreateUserIngredient` →
  `duplicateAfterWriteError`, Info `User ingredient created`) is used by the website `CreateIngredient` and by
  `AddIngredient`. The website keeps `validateCreateRequest` (`>= 0`, the name as sent).
- **Q-D2, statistics of the 201 → reuse `GetUserIngredientsWithUsage`** and pick the created ID, so the row is the list
  row by construction; no new SQL. The ingredient is already stored when the statistics are read, so a failure of that
  read is logged (Error, `user_id`, `ingredient_id`) and the 201 carries the ingredient with `times_used: 0` and no
  `last_used`, instead of a 500 that would make the client retry a create that succeeded (the retry would get 409).
- **Q-D3, older server.** Checked on the current code (v5.14.0, Echo v5.4.0) with a temporary test before the route
  was added: `POST /api/v1/ingredients` with a valid key → **404 `{"message":"Not Found"}`** (not 405); without a key
  or with only a JWT → 401 `API key required` (the group middleware runs for the unmatched route too). The client
  treats 404/405 without `Ingredient not found` as an older server. `skill_client_test.go` simulates both 405
  `{"message":"Method Not Allowed"}` and 404 `{"message":"Not Found"}` with an override in the test server wrapper
  (`answerAs`); no test files in `skills/`.

## Phase History
| Date       | Phase       | Agent            | Notes                                                                                   |
|------------|-------------|------------------|-----------------------------------------------------------------------------------------|
| 2026-10-16 | analysis    | business-analyst | Specs written: backend + client + SKILL.md, no frontend, no migration (verified); Q1–Q4 for the user |
| 2026-10-16 | analysis    | business-analyst | Q1 = A (SKILL_VERSION + `version`), Q2 = a (list row), Q3 ask when unclear, Q4 both; specs finalized |
| 2026-10-16 | backend-dev | business-analyst | First phase set; waiting for the user's approval of the specs                           |
| 2026-10-16 | backend-dev | backend-developer | `POST /api/v1/ingredients`, client `add-ingredient` / `version` / `SKILL_VERSION`, SKILL.md, tests, docs; Q-D1–Q-D3 above |
| 2026-10-16 | backend-review | backend-developer | Submitted for review: `go vet`, `go test ./...`, `go test -race` (changed packages) green; client tests ran with python3, none skipped |
| 2026-10-16 | security-review | backend-code-reviewer | Backend review clean (no report): vet/build/gofmt, fresh `go test ./...` and `-race` on changed packages green; client tests ran |
| 2026-10-16 | qa-review | security-reviewer | Security review clean (no report): key-scoped create, user-scoped 409 and usage stats, generic 500, limits from `internal/food`, api-v1 limiter; no secrets in the diff; no dependency changes |
| 2026-10-16 | ba-acceptance | qa-engineer | QA review clean (no report): `go vet`, `go test -count=1 ./...`, `-race` (ingredient, apidata, server, skill, skills) green; coverage of the new code 100% except the concurrent-delete fallback of `withUsage`; client tests ran with python3 3.14.4, none skipped; every automatable criterion mapped to a test |
| 2026-10-16 | backend-review | backend-developer | QA suggestions fixed (tests only, no production code): `TestSkillClientAddIngredient/duplicate` creates its own existing ingredient through the website API and checks nothing was created, `invalid value` checks one POST and an unchanged ingredient count; `TestAddIngredientUsage` covers the concurrent-delete fallback of `withUsage` (now 100%). Every subtest passes alone; `go vet ./...`, `go test -count=1 ./internal/server/ ./internal/services/ingredient/` green |
| 2026-10-16 | security-review | backend-code-reviewer | Backend review iteration 2 clean (no report): no production code changed since iteration 1; test changes reviewed; `go vet ./...`, `go test -count=1 ./internal/server/ ./internal/services/ingredient/` green; each `TestSkillClientAddIngredient` subtest passes alone |
| 2026-10-16 | qa-review | security-reviewer | Security re-review (iteration 2) clean: only `skill_client_test.go` and `service_test.go` changed; no keys, client runs isolated from `KKAL_*` env and use the local test server |
| 2026-10-16 | ba-acceptance | qa-engineer | QA re-review (iteration 2) clean: QA suggestions 1–3 resolved; each `TestSkillClientAddIngredient` subtest passes alone; `go vet`, `go test -count=1 ./...`, `-race` (server, ingredient) green; `withUsage` 100% covered |
| 2026-10-16 | devops | business-analyst | Acceptance passed: every criterion verifiable in code and tests is met (endpoint, service reuse, client, version, SKILL.md, docs, README); 1 manual post-release check pending (download + `add-ingredient` on the live server, Food List and day total) |

**Pending manual verification after release**: download the skill, run `add-ingredient` on the server, see the
ingredient on the Food List page after a reload, and check that the dashboard day total is unchanged.

## Review Iterations
- **Iteration 1**: 2026-10-16 — backend review clean. Non-blocking suggestions: mention the `SKILL_VERSION` bump rule in the AGENTS.md "Agent skill" rule; ignore `__pycache__/` (py_compile leaves it in `skills/.../scripts/`); the 400/409 mapping of the v1 ingredient create and update handlers could share a helper.
- **Iteration 1 (security)**: 2026-10-16 — security review clean, no findings. Informational only: no per-user cap on the number of ingredients (pre-existing; `POST /api/v1/food` already creates up to 50 per request).
- **Iteration 1 (QA)**: 2026-10-16 — QA review clean. Non-blocking suggestions: the `duplicate` subtest of `TestSkillClientAddIngredient` relies on the ingredient made by the `created` subtest (run alone with `-run 'TestSkillClientAddIngredient/^duplicate$'` it fails, because it creates the ingredient instead of getting a 409); the `invalid value` subtest does not assert one request / nothing created; the concurrent-delete fallback of `withUsage` (`service.go`, after the loop) has no test.
- **Iteration 2**: 2026-10-16 — backend review clean (tests only; AGENTS.md `SKILL_VERSION` rule and `.gitignore` `__pycache__/` checked).
- **Iteration 2 (security)**: 2026-10-16 — clean; test-only changes, no production code changed since iteration 1.
- **Iteration 2 (QA)**: 2026-10-16 — QA review clean. All 6 `TestSkillClientAddIngredient` subtests pass when run alone (python3, none skipped); `duplicate` is self-contained, `invalid value` checks one request and an unchanged count; the `withUsage` concurrent-delete fallback is tested (100% coverage).
- **BA acceptance**: 2026-10-16 — passed; status → `devops` (one manual check pending after release).
