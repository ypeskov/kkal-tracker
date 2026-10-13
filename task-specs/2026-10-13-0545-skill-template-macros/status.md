# Task Status: Food logging skill: template-only skills/, key in the client, macros step restored, per-entry macro overrides

## Current Phase
**Phase**: `done`
**Assigned to**: `-`
**Updated**: 2026-10-13

Chain (backend-only, no frontend changes): BA → Backend Dev → Backend Review → Security Review → QA Review →
BA Acceptance → DevOps.
User decisions: Parts A (1–7), B (1–3, implemented as the client command `totals`), C (include). Recorded in
`requirements.md`.
Q-D1, Q-D2 and Q-D3 are decided: see "Developer decisions".

## Developer decisions
- **Q-D1 (testing `totals`)**: Go tests in `internal/server/skill_client_test.go`. They serve the wired test server
  over a real `httptest` listener, download the skill, unpack it into `t.TempDir()` (with a stray `api_key` file next
  to `SKILL.md`, which must be ignored) and run `python3 scripts/kkal_client.py` with `KKAL_BASE_URL` set and
  `KKAL_API_KEY` unset, so the embedded key is used. Every request is recorded, so "exactly one request" and "no
  request" are asserted. `TestSkillClientTemplateWithoutKey` runs the unrendered repository template (placeholder →
  exit 1, Settings -> AI hint). The tests skip only when `python3` is not on `PATH`; on the dev VM (Python 3.14.4)
  they ran and passed, with no skip.
- **Q-D2 (key substitution API)**:
  - `Render` checks the placeholder without replacing it: `API_KEY_HERE` must occur exactly once in the client, in
    the line `API_KEY = "API_KEY_HERE"`, and never in `SKILL.md`; otherwise `ErrSourceLayout`. This runs before any
    key exists.
  - `(*Skill).WithAPIKey(rawKey)` returns a copy with the key substituted. A key that is not `^[0-9a-f]{64}$` gives
    `ErrInvalidAPIKey`; a key line that is missing or duplicated gives `ErrSourceLayout`. Error texts never contain
    the key.
  - The service runs render → `CreateKey` → `WithAPIKey` → zip. A failure after key creation deletes the key (the
    existing `deleteKey`), and the handler answers 500 "Failed to prepare the skill". The service gives
    `scripts/kkal_client.py` mode 0600 (it holds the key) and `SKILL.md` 0644.
- **Q-D3**: `ProductionURL` stays the anchor constant. `DevelopmentURL` is removed; the tests use a local constant.
  `WriteDir` and its test are removed (there is no caller left).
- Other decisions:
  - `totals` without `--day` is an argparse usage error (exit 2, non-zero). A malformed `--day`, including
    non-zero-padded dates such as `2026-3-1`, gives exit 1 with a message. `--expect-meal-kcal` without `--meal` gives
    exit 1 before any request.
  - `meal.entry_ids` lists the `--meal` IDs found on that day, in `--meal` order, without duplicates. The missing
    IDs go to `missing_meal_ids`.
  - Grams are summed as `Decimal` built from the JSON numbers' text, then rounded once, halves up. This avoids
    binary float error before the rounding.
  - The client's "not configured" message names "Settings -> AI" (ASCII arrow, like the other client messages). It
    does not mention `KKAL_API_KEY`. That variable is described as a manual-check override in the configuration
    comment and in SKILL.md.
  - In POST, the macro overrides are validated in the `ingredient_id` branch next to `kcal_per_100g`. In PUT they are
    applied after the ingredient re-copy and the kcal override. A shared helper (`overrideNutrients`) validates all
    three values before applying any of them.
  - The Python helper `load_api_key` is renamed `load_key`, so a grep for the old `api_key` file finds only the
    `.gitignore`/`.dockerignore` safety nets and the explicit-embed comments.
  - No path references to fix inside this folder after the rename: none pointed to the old `...-HHMM-...` name.

## Phase History
| Date       | Phase    | Agent            | Notes                                                                 |
|------------|----------|------------------|-----------------------------------------------------------------------|
| 2026-10-13 | analysis | business-analyst | Specs written: Parts A–C, backend-only; no migration and no frontend changes (verified) |
| 2026-10-13 | backend-dev | backend-developer | Specs approved by the user on 2026-10-13; implementation of Parts A–C started |
| 2026-10-13 | backend-review | backend-code-reviewer | Parts A–C implemented; go vet, go test, go test -race (changed packages) green; Python client tests ran |
| 2026-10-13 | backend-review | backend-developer | security suggestion: never-print-key line added to SKILL.md (API basics) |
| 2026-10-13 | security-review | security-reviewer | Backend review clean (no report); `totals` exit 2 without --day and the --expect-meal-kcal/--meal rule accepted |
| 2026-10-13 | qa-review | qa-engineer | Security review clean (no report): key format validated before insertion, key deleted on failure after creation, no key in logs/errors/headers, client 0600, explicit embed, no real keys in tracked files or skills/ history; Part C ownership and 0–100 validation OK; dependencies unchanged |
| 2026-10-13 | security-review | security-reviewer | Re-review clean: SKILL.md key-secrecy line in API basics; frontmatter unchanged, no API_KEY_HERE; skill/server/skills tests green |
| 2026-10-13 | qa-review | qa-engineer | Security re-check clean (no report): the SKILL.md key-secrecy line addresses the optional suggestion; only SKILL.md changed since the security review |
| 2026-10-13 | ba-acceptance | business-analyst | QA review clean (no report): go vet, go test -count=1 ./..., go test -race (changed packages) green; skill client tests ran on Python 3.14.4 (9 subtests + template test, no skip); every acceptance criterion mapped to tests except the manual claude.ai / dev-server checks; frontend tsc OK, no web/ changes |

| 2026-10-13 | devops | business-analyst | Acceptance passed: every criterion that code and tests can show is met (Parts A–C, SKILL.md content, docs, AGENTS.md); 2 integration criteria left for manual verification after release (dev zip run; real Claude conversation: table rendering, day row after edit/delete, prod download and key listing, claude.ai upload) |
| 2026-10-13 | done | orchestrator | Released v5.14.0 with deploy.sh from the main session (the devops agent declined a relayed confirmation); pod 1/1 Running on 5.14.0, 0 restarts; site 200; download without auth 401; released binary contains the template placeholder and the macros step, no api_key files in the image |

**Pending manual verification after release**:
- the downloaded dev/prod zip runs `ingredients` and `totals` with no environment variables;
- in a real Claude conversation:
  - the Ккал/Ж/У/Б table renders with rows "Приём"/"За день";
  - the day row is shown after edit and after delete;
- the prod zip has two files, and its key is listed as `Claude skill … UTC`;
- the claude.ai upload succeeds.

QA non-blocking notes:
- `totals` is not tested against 429/5xx (it uses the same `expect` helper as the other commands);
- the "edit macros" subtest changes entry B and runs last.

## Review Iterations
- **Iteration 1**: 2026-10-13 — backend review clean; go vet, go build, go test, go test -race (changed packages) green; skill client tests ran (not skipped); security review clean
- **Iteration 2**: 2026-10-13 — re-review of the SKILL.md key-secrecy line: clean
- **Iteration 3**: 2026-10-13 — QA review clean; status → ba-acceptance
- **Iteration 3**: 2026-10-13 — BA acceptance passed; status → devops (manual checks pending after release)
