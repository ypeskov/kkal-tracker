---
name: workflow
description: Task lifecycle for the agent team - BA → backend/frontend dev → code reviews → security → QA → BA acceptance → DevOps, with task-specs/, status.md and review reports in agent-reviews/. Use when the user asks to run a task (feature, bug, refactoring) through the workflow or the agents, and in every subagent working on a task from task-specs/.
---

# Agent Workflow

The task lifecycle and the interaction flow between the project agents (`.claude/agents/`). The orchestrator (the
main session) runs the steps by delegating each one to its subagent with the means of the current runtime, and
relays questions between the subagents and the user: subagents do not talk to the user directly.

## Task lifecycle
```
User → BA → Backend Dev → Backend Review → Frontend Dev → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps → Done
```
Steps that do not apply are skipped (see Skipping steps). After any review step, work can return to the previous
developer for fixes; after fixes, the **full review chain repeats from that point forward**.

## Steps

### 1. Task input — User
Describes a task (bug, feature, refactoring) in any form (text, images, ...).

### 2. Business analysis — `business-analyst`
1. Reads `AGENTS.md` and the relevant code.
2. Returns clarifying questions to the orchestrator when requirements are unclear (the orchestrator asks the user).
3. Creates `task-specs/YYYY-MM-DD-HHMM-<task-name>/` (date and time of creation).
4. Writes `requirements.md` (subtasks per agent) and `acceptance-criteria.md` (checklists).
5. Creates `status.md` with the first applicable phase.

### 2.5. User approval — orchestrator
Presents `requirements.md` and `acceptance-criteria.md` to the user and gets explicit confirmation before any
development. Feedback goes back into the specs, then confirm again.

### 3. Backend development — `backend-developer` (skip without backend changes)
1. Reads `task-specs/<task>/requirements.md`.
2. Implements handlers, services, repositories, models; creates and applies migrations if needed.
3. Writes tests; `go vet ./...` and `go test ./...` are green.
4. Updates `AGENTS.md` / `docs/` if they became outdated.
5. `status.md` → `backend-review`.

### 4. Backend code review — `backend-code-reviewer` (skip without backend changes)
1. Reviews all backend changes, writes `agent-reviews/backend-review.md`.
2. Issues → `status.md` → `backend-dev`. Clean → deletes the report, `status.md` → `frontend-dev` (or
   `security-review` without frontend changes).

### 4.5. Frontend development — `frontend-developer` (skip without frontend changes)
1. Reads `task-specs/<task>/requirements.md`.
2. Implements pages, components, hooks, API services, styles, i18n.
3. From `web/`: `npx tsc --noEmit -p tsconfig.app.json`, `npm run lint`, `npm run build` pass.
4. Updates `AGENTS.md` / `docs/` / the `frontend-style` skill if they became outdated.
5. `status.md` → `frontend-review`.

### 4.6. Frontend code review — `frontend-code-reviewer` (skip without frontend changes)
1. Reviews all frontend changes, writes `agent-reviews/frontend-review.md`.
2. Issues → `status.md` → `frontend-dev`. Clean → deletes the report, `status.md` → `security-review`.

### 5. Security review — `security-reviewer`
1. Reviews all changes for vulnerabilities; audits dependencies if they changed.
2. Writes `agent-reviews/security-review.md`.
3. Issues → `status.md` → the responsible developer; after the fix the chain repeats from that developer's review.
   Clean → deletes the report, `status.md` → `qa-review`.

### 6. QA review — `qa-engineer`
1. Reads `requirements.md` and `acceptance-criteria.md`, the changed source and test files.
2. Runs the tests, the race detector and coverage; checks coverage of the criteria, error paths, edge cases.
3. Writes `agent-reviews/qa-review.md`.
4. Issues → `status.md` → the responsible developer; after the fix the chain repeats from that developer's review.
   Clean → deletes the report, `status.md` → `ba-acceptance`.

### 7. BA acceptance — `business-analyst`
1. Verifies every criterion of `acceptance-criteria.md` against the code and the test results, checks the boxes.
2. A criterion not met → `status.md` → the responsible developer; the chain repeats from that point.
   All met → `status.md` → `devops`.

### 7.5. Commit — orchestrator
Shows the user the result and, once they agree, commits on `develop` (one commit per logical change, rules from
`AGENTS.md`). The release needs a clean, committed tree.

### 8. DevOps — `devops-engineer`
1. Decides what is needed (usually a release).
2. **Asks the user for confirmation** before anything that changes production.
3. Releases with the `deploy` skill (`deploy.sh`: image from `master`, version commit, rollout) and verifies it.
4. `status.md` → `done`.

## Return flow (fixes)
1. The reviewer writes a report and sets the `status.md` phase back to the developer.
2. The developer fixes the issues.
3. The full chain repeats from the review step that follows the developer:
   - Backend fix → Backend Review → Frontend Dev → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps
   - Frontend fix → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps
   - Security or QA fix in the backend → Backend Review → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps
   - Security or QA fix in the frontend → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps

## Status tracking
Every task has `task-specs/<task>/status.md`:
```markdown
# Task Status: <Task Title>

## Current Phase
**Phase**: `<phase>`
**Assigned to**: `<agent name>`
**Updated**: YYYY-MM-DD

## Phase History
| Date       | Phase          | Agent                 | Notes                  |
|------------|----------------|-----------------------|------------------------|
| YYYY-MM-DD | backend-dev    | backend-developer     | Started implementation |
| YYYY-MM-DD | backend-review | backend-code-reviewer | Submitted for review   |

## Review Iterations
- **Iteration 1**: <date> — <outcome>
```
Phases: `analysis`, `backend-dev`, `backend-review`, `frontend-dev`, `frontend-review`, `security-review`,
`qa-review`, `ba-acceptance`, `devops`, `done`.

## Review reports
`agent-reviews/<kind>-review.md` (`backend`, `frontend`, `security`, `qa`). A report exists while its issues are open
and is deleted when the review is clean: no report means the step has passed. Create `agent-reviews/` when missing.

## Skipping steps
- **Backend-only**: BA → Backend Dev → Backend Review → Security Review → QA Review → BA Acceptance → DevOps
- **Frontend-only**: BA → Frontend Dev → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps
- **Full-stack**: the whole chain
- **DevOps-only**: BA → DevOps
- **Bug fix**: the applicable subset

## Parallelism
- **Allowed**: several developers on different files without conflicts; entirely separate tasks
- **Not allowed**: reviews of the same task at the same time (backend + security + QA), any dependent steps of the
  same task
