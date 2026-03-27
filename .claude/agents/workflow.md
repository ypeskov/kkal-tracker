# Agent Workflow

This document defines the task lifecycle and interaction flow between all project agents.
Every agent must read and follow this workflow.

## Task Lifecycle

```
User → BA → Backend Dev → Backend Review → Frontend Dev → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps → Done
```

Steps are skipped when not applicable (e.g., infrastructure-only tasks skip backend dev/review; frontend-only tasks skip backend steps; backend-only tasks skip frontend steps).
After any review step, work can return to the previous developer for fixes. When fixes are made, the **full review chain repeats from that point forward**.

## Detailed Flow

### Step 1: Task Input
- **Who**: User
- **What**: Describes a task (bug, feature, or refactoring) in any format (text, images, etc.)

### Step 2: Business Analysis
- **Who**: Business Analyst
- **What**:
  1. Reads `CLAUDE.md` at the project root.
  2. Asks the user clarifying questions if requirements are unclear.
  3. Reads relevant codebase to understand the current state.
  4. Creates task directory: `task-specs/YYYY-MM-DD-HHMM-<task-name>/` (date + time of creation).
  5. Writes `requirements.md` with decomposed subtasks per agent.
  6. Writes `acceptance-criteria.md` with checklists.
  7. Creates `status.md` and sets phase to `backend-dev` (or first applicable phase).
  8. Hands off to the first applicable developer.
- **Can ask**: User, any agent (for technical clarification).

### Step 2.5: User Approval
- **Who**: Orchestrator
- **What**: After BA completes, the orchestrator MUST present the generated `requirements.md` and `acceptance-criteria.md` to the user and get explicit confirmation before proceeding. If the user has feedback or changes, update the specs accordingly and re-confirm.

### Step 3: Backend Development
- **Who**: Backend Developer
- **Skip if**: Task does not involve backend changes.
- **What**:
  1. Reads task spec from `task-specs/<task>/requirements.md`.
  2. Implements changes (handlers, services, repositories, models).
  3. Creates and applies migrations if needed.
  4. Writes tests for new functionality.
  5. Runs all tests — ensures everything is green.
  6. Runs `go vet` — ensures no errors.
  7. Updates `CLAUDE.md` if architecture changed.
  8. Updates `status.md` → phase: `backend-review`.
- **Can ask**: Business Analyst (requirements clarification). BA may escalate to User.

### Step 4: Backend Code Review
- **Who**: Backend Code Reviewer
- **Skip if**: Task does not involve backend changes.
- **What**:
  1. Reviews all backend changes against review criteria.
  2. Writes report to `agent-reviews/backend-review.md`.
  3. If issues found → updates `status.md` → phase: `backend-dev` (return for fixes).
  4. If clean → deletes report, updates `status.md` → phase: `frontend-dev` (or `security-review` if no frontend changes).

### Step 4.5: Frontend Development
- **Who**: Frontend Developer
- **Skip if**: Task does not involve frontend changes.
- **What**:
  1. Reads task spec from `task-specs/<task>/requirements.md`.
  2. Implements changes (pages, components, hooks, API services, styles, i18n).
  3. Runs `npm run lint` and `npm run build` from `web/` — ensures no errors.
  4. Updates `CLAUDE.md` if architecture changed.
  5. Updates `status.md` → phase: `frontend-review`.
- **Can ask**: Business Analyst (requirements clarification). BA may escalate to User.

### Step 4.6: Frontend Code Review
- **Who**: Frontend Code Reviewer
- **Skip if**: Task does not involve frontend changes.
- **What**:
  1. Reviews all frontend changes against review criteria.
  2. Writes report to `agent-reviews/frontend-review.md`.
  3. If issues found → updates `status.md` → phase: `frontend-dev` (return for fixes).
  4. If clean → deletes report, updates `status.md` → phase: `security-review`.

### Step 5: Security Review
- **Who**: Security Reviewer
- **What**:
  1. Reviews all changes for security vulnerabilities.
  2. Runs dependency audits if dependencies changed.
  3. Writes report to `agent-reviews/security-review.md`.
  4. If issues found → updates `status.md` → phase: `backend-dev`. After fix, the full review chain repeats from Step 4 (backend review).
  5. If clean → deletes report, updates `status.md` → phase: `qa-review`.

### Step 6: QA Review
- **Who**: QA Engineer
- **What**:
  1. Reads task spec (`requirements.md` and `acceptance-criteria.md`).
  2. Reads changed source and test files.
  3. Runs `go test ./...`, `go test -race ./...`, and coverage report.
  4. Checks test coverage against acceptance criteria, error paths, edge cases, and mock quality.
  5. Writes report to `agent-reviews/qa-review.md`.
  6. If issues found → updates `status.md` → phase: `backend-dev`. After fix, the full review chain repeats from Step 4 (backend review).
  7. If clean → deletes report, updates `status.md` → phase: `ba-acceptance`.

### Step 7: BA Acceptance Check
- **Who**: Business Analyst
- **What**:
  1. Reviews the implemented changes against `acceptance-criteria.md`.
  2. Reads the code and test results to verify each criterion.
  3. Checks each acceptance criterion checkbox.
  4. If any criteria not met → updates `status.md` → phase back to the responsible developer. Full review chain repeats from that point.
  5. If all criteria met → updates `status.md` → phase: `devops`.

### Step 8: DevOps
- **Who**: DevOps Engineer
- **What**:
  1. Evaluates what is needed (usually: build Docker image + deploy to Kubernetes).
  2. Builds and pushes Docker image via `build-and-push.sh`.
  3. **MANDATORY**: Asks the user for confirmation before deploying to production.
  4. After user confirmation → deploys to production via `kubectl`.
  5. Verifies rollout status.
  6. Updates `status.md` → phase: `done`.

## Return Flow (Fixes)

When any review step finds issues:

1. The reviewer writes a report and sets `status.md` phase back to the developer.
2. The developer fixes the issues.
3. The **full review chain repeats from the review step that follows the developer**:
   - Backend dev fixes → Backend Review → Frontend Dev → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps
   - Frontend dev fixes → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps
   - Security fix (backend) → Backend Review → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps
   - Security fix (frontend) → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps
   - QA fix (backend) → Backend Review → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps
   - QA fix (frontend) → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps

## Status Tracking

Every task has a `status.md` file in its `task-specs/<task>/` directory.

### status.md Format

```markdown
# Task Status: <Task Title>

## Current Phase

**Phase**: `<phase>`
**Assigned to**: `<agent name>`
**Updated**: YYYY-MM-DD

## Phase History

| Date       | Phase            | Agent               | Notes                     |
|------------|------------------|----------------------|---------------------------|
| YYYY-MM-DD | backend-dev      | Backend Developer    | Started implementation    |
| YYYY-MM-DD | backend-review   | Backend Code Reviewer| Submitted for review      |
| ...        | ...              | ...                  | ...                       |

## Review Iterations

- **Iteration 1**: <date> — <outcome>
- **Iteration 2**: <date> — <outcome>
```

### Valid Phases

- `analysis` — BA is processing the task
- `backend-dev` — Backend Developer is working
- `backend-review` — Backend Code Reviewer is reviewing
- `frontend-dev` — Frontend Developer is working
- `frontend-review` — Frontend Code Reviewer is reviewing
- `security-review` — Security Reviewer is reviewing
- `qa-review` — QA Engineer is reviewing test coverage and quality
- `ba-acceptance` — BA is verifying acceptance criteria
- `devops` — DevOps is building/deploying
- `done` — Task is complete

## Skipping Steps

- **Backend-only task**: BA → Backend Dev → Backend Review → Security Review → QA Review → BA Acceptance → DevOps.
- **Frontend-only task**: BA → Frontend Dev → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps.
- **Full-stack task**: BA → Backend Dev → Backend Review → Frontend Dev → Frontend Review → Security Review → QA Review → BA Acceptance → DevOps.
- **DevOps-only task**: BA → DevOps.
- **Bug fix**: Follow the applicable subset.

## Parallelism

- **Allowed**: multiple Backend Dev agents for different files without conflicts; entirely separate tasks.
- **Not allowed**: Backend Review + Security Review + QA Review simultaneously; any dependent steps of the same task.
