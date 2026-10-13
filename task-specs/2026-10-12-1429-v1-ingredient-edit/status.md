# Task Status: Edit and delete ingredients through the external API (`/api/v1`), duplicate-name protection on the website

## Current Phase
**Phase**: `done`
**Assigned to**: —
**Updated**: 2026-10-12

Chain (full-stack): BA → Backend Dev → Backend Review → Frontend Dev → Frontend Review → Security Review → QA Review →
BA Acceptance → DevOps.
Q1–Q3 answered by the user; specs approved by the user (2026-10-12). Backend review passed (iteration 2); frontend review passed (iteration 1); security review passed (iteration 1); QA review passed (iteration 1); BA acceptance passed.
**Not verified** (user chose to release without manual verification, 2026-10-12; the dev API key was invalid on the dev
server): 7 browser checks of the Food List modals (Frontend section) and 2 Integration checks with the live dev server,
marked in `acceptance-criteria.md`. All other criteria are met.

### Backend decisions
- **Q-D1**: new package `internal/food` (`food.go`): `MaxKcalPer100g` / `MaxNutrientPer100g` / `MaxNameLength`,
  `ValidKcalPer100g`, `ValidNutrientPer100g`, `NameTooLong`, `CleanName`, `NormalizeName`. The calorie service
  (`POST /api/v1/food`, `PUT /api/v1/food/:id`) and the ingredient service both use it; the unexported copies in
  `calorie/service.go` are removed (the item weight limit stays in `calorie`, it is not shared).
- **Q-D2**: a separate service method `PatchIngredient(*PatchIngredientRequest)` with pointer fields;
  `UpdateIngredient` keeps the full-replace semantics. Both read the ingredient first (ownership → `sql.ErrNoRows`)
  and write through the shared `saveIngredient` (duplicate check + `UpdateUserIngredient`).
- Duplicate rule: one implementation, `checkDuplicateName` in the ingredient service (create, full and partial update);
  typed `DuplicateIngredientError{Existing}`. Race / unique-index failure: after a failed write the service re-checks
  for another ingredient with exactly the written name (the index compares exact text) and reports it as the
  duplicate (409), otherwise the error stays internal (500 generic);
  driver-independent, no error-string matching.
- Also changed: the dev skill description in `skills/sync_dev_skill.py` names ingredients next to entries as the
  dev-routing exception (consistent with the new "ingredient IDs are per server" rule).

### Frontend decisions
- `web/src/api/ingredients.ts`: `DuplicateIngredientError{existing: Ingredient}` on a 409 with a well-formed
  `existing_ingredient`; any other non-OK response (incl. a malformed 409) → `IngredientRequestError{status}` with a
  fixed message; the response text is never read into an error.
- `FoodList.tsx` derives the inline message from `createMutation.error` / `updateMutation.error`
  (`foodList.duplicateName` with `existing.name`, otherwise `foodList.saveError`) and passes it to the modals
  (`errorMessage`, `onNameChange`). Reset: `mutation.reset()` on open, close (Cancel, ×, backdrop) and on a name edit
  while an error is shown; a new submit clears it (TanStack Query sets `error: null` on `pending`).
- `docs/features.md`: the Food List behavior on a duplicate name.

## Phase History
| Date       | Phase    | Agent            | Notes                                                                      |
|------------|----------|------------------|----------------------------------------------------------------------------|
| 2026-10-12 | analysis | business-analyst | Specs written, open questions Q1–Q3 for the user                           |
| 2026-10-12 | analysis | business-analyst | Q1/Q2 defaults accepted, Q3 in scope: website 409 + frontend, full-stack   |
| 2026-10-12 | backend-dev | backend-developer | Specs approved by the user (2026-10-12); started implementation         |
| 2026-10-12 | backend-review | backend-code-reviewer | Backend, client, skill, docs done; vet/test/race green; submitted    |
| 2026-10-12 | backend-dev | backend-code-reviewer | Review 1: race handling reports any write error as 409 when a near-duplicate exists |
| 2026-10-12 | backend-review | backend-developer | Review 1 fixed: after a failed write only an exact name match is a 409; test added; features.md clarified |
| 2026-10-12 | frontend-dev | backend-code-reviewer | Review 2 clean: vet/test/race green; report deleted |
| 2026-10-12 | frontend-review | frontend-developer | Typed 409 error, inline duplicate/generic save errors in both modals, 2 keys × 4 locales; tsc/lint/build pass |
| 2026-10-12 | security-review | frontend-code-reviewer | Frontend review clean: tsc/lint (0 errors)/build pass, 409 contract matches, 4 locales; no report |
| 2026-10-12 | qa-review | security-reviewer | Security review clean: ownership in SQL (404 for foreign IDs), 409 only discloses own ingredients, no internal errors, React-escaped messages, no secrets, no dependency changes; no report |
| 2026-10-12 | ba-acceptance | qa-engineer | QA review clean: vet, test (90 pass), race, tsc, lint (0 errors), build pass; coverage of the changed code 84.0% incl. integration tests; client checked against a local mock server; no report |
| 2026-10-12 | devops | business-analyst | Acceptance passed: all code/test-verifiable criteria met; 9 browser/dev-server criteria not verified (user chose to release without manual checks) |
| 2026-10-12 | done | devops-engineer | Released v5.12.0 with deploy.sh (no migrations): image ypeskov/kcal-tracker:5.12.0, rollout OK, pod Running 0 restarts, prod serves the new bundle; /api/v1/ingredients/:id PUT/DELETE without a key → 401 |

## Review Iterations
- **Iteration 1**: 2026-10-12 — backend review: 1 improvement required (`duplicateAfterWriteError` must match the exact name), 1 doc suggestion; back to `backend-dev`
- **Iteration 2**: 2026-10-12 — backend review clean; on to `frontend-dev`
- **Iteration 3**: 2026-10-12 — frontend review clean; on to `security-review`
- **Iteration 4**: 2026-10-12 — security review clean; on to `qa-review`
- **Iteration 5**: 2026-10-12 — QA review clean; on to `ba-acceptance` (browser and live dev-skill checks left for manual verification)
- **Iteration 6**: 2026-10-12 — BA acceptance passed; on to `devops` (browser and live dev-server criteria not verified by the user's decision)
