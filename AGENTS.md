# Kkal Tracker

Calorie and weight tracking web application: Go/Echo backend with a React/TanStack frontend embedded in a single binary.
Calorie diary, weight history and goals, ingredient database, health metrics, reports, AI insights, Excel export,
API keys for external access (including food logging by AI agents). Details: [`docs/features.md`](docs/features.md).

## Map
- Go/Echo + SQLite backend in `internal/`: handler → service → repository, all SQL in `internal/repositories/queries.go`
- React/TanStack/Tailwind frontend in `web/`, built to `web/dist` and embedded into the single binary
- `skills/` is a product feature (skills for external agents that log food), not the agent harness
- Stack, directory layout, conventions: [`docs/architecture.md`](docs/architecture.md)

## Commands
```bash
make watch                 # dev server with Air live reload (see the dev-server skill)
make build                 # frontend + backend; make build-frontend builds only web/dist
make migrate-up            # also: migrate-down, migrate-status, migrate-create NAME=...
make seed                  # seed ingredients; seed-clean re-seeds (dev only)
go vet ./... && go test ./...                          # tests need web/dist: run make build-frontend first
cd web && npx tsc --noEmit -p tsconfig.app.json && npm run lint
```
Go tests include an end-to-end smoke test (`internal/server/smoke_test.go`): the fully wired server on a temporary
migrated SQLite database with a fake SMTP server (register → activation → login → calories → weight → export).

## Rules
- **Never expose internal errors to clients.** Log `err`, return a generic message:
  ```go
  h.logger.Error("Export failed", "error", err, "user_id", userID)
  return echo.NewHTTPError(http.StatusInternalServerError, "Export failed")   // never "...: "+err.Error()
  ```
  Validation errors from `c.Validate()` (400) may pass through: they contain only field names and rules
- **Logging**: never import `log/slog` directly, use the injected logger
- **Comments, commit messages, docs**: English only
- **Git commits**: no "Co-Authored-By" or any other AI/assistant attribution, no mentions of AI tools
- **Libraries**: before writing code against a library, fetch its current docs for the version in `go.mod` /
  `package.json` (Context7 MCP server when available), do not rely on memory
- **Migrations** are never run automatically: neither the server nor the deploy applies them (see the `deploy` skill)
- **CSP** (`internal/server/security.go`): scripts only from the own origin, no inline scripts, no `eval`. Do not add
  inline `<script>`, external script/style/font hosts or eval-based code without updating the policy
- **Secrets**: API keys never go to the repository (skills hold the `XXXXXXXXX` placeholder, real keys live in
  git-ignored `api_key` files); `.env` / `.env.secret` exist only locally and on the server
- **Frontend**: Tailwind utilities only, every user-facing string through `t()` in all 4 locales, units too
  (`common.kg`, `common.kcalPerDay`). Layout and component patterns: the `frontend-style` skill
- **API**: all routes under `/api`; list in [`docs/api.md`](docs/api.md)
- **Dependencies**: TypeScript stays on 6.0 (typescript-eslint supports < 6.1 only)
- **Production** runs only what is on `master` and is released only with `deploy.sh` (the `deploy` skill)

## Docs
- [`docs/architecture.md`](docs/architecture.md) — stack, directory layout, conventions
- [`docs/features.md`](docs/features.md) — features and their behavior
- [`docs/api.md`](docs/api.md) — API endpoints
- [`docs/infrastructure.md`](docs/infrastructure.md) — Docker, Kubernetes, backups, environment variables

## Skills
Procedures live in `.agents/skills/` (`.claude/skills` is a symlink to it): `deploy`, `dev-server`, `frontend-style`,
`prod-db-snapshot`, `prod-db-from-backup`, `sync-subagents`, `workflow`.

## Subagents
Subagents are maintained in `.claude/agents/` (source of truth, Claude Code format): `business-analyst`,
`backend-developer`, `backend-code-reviewer`, `frontend-developer`, `frontend-code-reviewer`, `security-reviewer`,
`qa-engineer`, `devops-engineer`. Change them only there. How they work together on a task (BA → dev → reviews →
QA → acceptance → DevOps, `task-specs/`, `agent-reviews/`): the `workflow` skill.

**Every task that changes code (feature, bug, refactoring) runs through the `workflow` skill by default.** If a task
looks small, propose to the user to skip the workflow and do it in the main session; skip it only when they agree.

If you are not Claude Code and your runtime supports subagents: at session start run the check from the
`sync-subagents` skill; if your runtime's subagents are missing or stale, offer the user to regenerate them.
