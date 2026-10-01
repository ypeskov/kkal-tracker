# Update & Security Fix Plan

Source: `reports/security_report-2026-10-01.md` (audit of v5.5.2, commit `4361c46`).
IDs in brackets (H1, M3, L5, ...) refer to findings in that report.

## Status Legend
- [ ] Pending
- [x] Done
- [~] Partly done / done but not fully verified (see note)

## Progress

| Stage | Done | Total |
|-------|------|-------|
| 1. Dependency & toolchain refresh | 12 | 13 |
| 2. High-priority code fixes | 0 | 4 |
| 3. Server hardening (`server.go`) | 0 | 6 |
| 4. Kubernetes & backup | 0 | 8 |
| 5. Low-priority code fixes | 0 | 7 |
| 6. Non-security defects | 0 | 3 |
| 7. Major upgrades & long-term | 0 | 9 |

---

## Stage 1 — Dependency & toolchain refresh (no code changes)

### Go
- [x] 1.1 Bump direct modules: echo 4.15.0 → 4.16.0, excelize 2.10.0 → 2.11.0, x/crypto 0.47.0 → 0.57.0, modernc sqlite 1.44.3 → 1.60.1, goose 3.26.0 → 3.28.0, validator 10.30.1 → 10.30.5, go-openai 1.41.2 → 1.43.0 [H1]
- [x] 1.2 Bump indirect modules with advisories: x/net 0.49.0 → 0.59.0, x/text 0.33.0 → 0.42.0 [H1]
- [x] 1.3 Move off EOL Go 1.25: `go.mod` now `go 1.26.0` (minimum required by the updated modules), `Dockerfile` builds with `golang:1.27` [H1]
- [x] 1.4 `govulncheck ./...` reports no reachable advisories (with Go 1.27.1; with local Go 1.26.2 only stdlib advisories remain, see manual item below)

### npm (`web/`)
- [x] 1.5 In-range update of all packages, incl. dompurify 3.3.1 → 3.4.16, @tanstack/react-router 1.157.18 → 1.170.41, vite 7.3.1 → 7.3.6, postcss 8.5.6 → 8.5.28, react 19.2.4 → 19.3.0; version floors in `package.json` raised to match [M6]
- [x] 1.6 `npm audit` reports 0 vulnerabilities (was 16)
- [x] 1.7 Remove deprecated `@types/dompurify`
- [x] 1.8 Frontend builds (`npm run build`, includes `tsc -b`)

### Docker / build scripts
- [x] 1.9 `Dockerfile`: `npm install` → `npm ci`
- [x] 1.10 `Dockerfile`: `node:22-alpine` → `node:24-alpine`
- [x] 1.11 `Dockerfile`: `distroless/base-debian12` → `distroless/static-debian13` (image 49.7 MB → 21.8 MB)
- [x] 1.12 `build-and-push.sh`: add `--pull`
- [ ] 1.13 Add `.dockerignore` (found during stage 1: there is none, so `COPY . .` sends `.git`, `data/`, local `.env`, `web/node_modules` into the builder stage; the final image is not affected)

Manual (developer machine, not in repo):
- [ ] Update local Go 1.26.2 → latest 1.26.x/1.27.x (15 stdlib advisories are reachable when building locally with 1.26.2)

Not verified in stage 1: runtime behaviour. There are no automated tests and the server was not started (project rule). Do a manual smoke test (login, add entry, weight, export, AI) before deploying.

## Stage 2 — High-priority code fixes

- [ ] 2.1 Call `c.Validate(&req)` in `Login` and `Register` (`internal/handlers/auth/handler.go:50,85`) [H2]
- [ ] 2.2 Remove `COPY .env.sample .env` from `Dockerfile:38` [H3]
- [ ] 2.3 Set `ENVIRONMENT=production` for the production image/deployment [H3]
- [ ] 2.4 Make config fail closed: refuse placeholder/short `JWT_SECRET` unless `ENVIRONMENT=development` is set explicitly (`internal/config/config.go:50-72`) [H3]

Before deploying 2.2–2.4: confirm the prod ConfigMap provides `JWT_SECRET` (≥32 chars), `ENVIRONMENT`, `LOG_LEVEL`, `PORT`, `DATABASE_PATH`.

## Stage 3 — Server hardening (`internal/server/server.go`)

- [ ] 3.1 Separate strict rate limiters for `/auth/login` and `/auth/register` [M1]
- [ ] 3.2 Log failed logins at warn level [M1]
- [ ] 3.3 Configure `e.IPExtractor` to trust `X-Forwarded-For` only from the cluster network [M2]
- [ ] 3.4 `http.Server` timeouts: `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` [M5]
- [ ] 3.5 `BodyLimit` middleware [M5]
- [ ] 3.6 Rate-limit `/api/v1` before the API-key check [L4]

## Stage 4 — Kubernetes & backup

- [ ] 4.1 Move `JWT_SECRET`, `SMTP_PASSWORD`, `OPENAI_API_KEY`, `GDRIVE_OAUTH_TOKEN` from ConfigMap to Secret (`secretGenerator` + `secretRef`) [M3]
- [ ] 4.2 Backup CronJob: pass only `GDRIVE_*` instead of the whole app env [M4]
- [ ] 4.3 Pin `rclone/rclone` image by version/digest [M4]
- [ ] 4.4 Stop installing `sqlite` from the network on every backup run [M4]
- [ ] 4.5 rclone `scope = drive` → `drive.file` (needs a new OAuth token) [M4]
- [ ] 4.6 Encrypt backups (rclone `crypt` remote) [M4]
- [ ] 4.7 Deployment `securityContext`: read-only root FS, drop capabilities, no privilege escalation, seccomp [L1]
- [ ] 4.8 Deployment resource requests/limits and liveness/readiness probes [L1]

## Stage 5 — Low-priority code fixes

- [ ] 5.1 Nil-pointer panic on weight update without `recorded_at` (`internal/repositories/weight_history.go:172`) [L5]
- [ ] 5.2 Enable SQLite `foreign_keys` and `busy_timeout` pragmas (`internal/database/database.go:19`); check existing data for orphans first [L6]
- [ ] 5.3 Return 404 instead of 500 for unknown API key on revoke/delete (`internal/handlers/apikey/handler.go:115,131`) [L7]
- [ ] 5.4 Max-length validation for `food`, ingredient `name`, `first_name`, `last_name` [L9]
- [ ] 5.5 Constant-time login for unknown emails; decide on register 409 behaviour [L3]
- [ ] 5.6 Hash activation tokens at rest [L8]
- [ ] 5.7 JWT: recheck user `is_active` per request or shorten lifetime; add `iss` [L2]

## Stage 6 — Non-security defects

- [ ] 6.1 AI system prompt is always empty: add goal fields to `systemPromptData` (`internal/services/ai/openai.go:122-126`)
- [ ] 6.2 Export email: body declared `quoted-printable` but sent unencoded (`internal/services/email/service.go:151`)
- [ ] 6.3 PostgreSQL mode: implement (driver + `database.New`) or remove the option and docs

## Stage 7 — Major upgrades & long-term (separate tasks)

- [ ] 7.1 Smoke tests for auth, calories, weight, export (prerequisite for the rest)
- [ ] 7.2 Add ESLint config so `npm run lint` works
- [ ] 7.3 vite 7 → 8, @vitejs/plugin-react 5 → 6
- [ ] 7.4 TypeScript 5.9 → 7.0
- [ ] 7.5 eslint 9 → 10 (npm now marks the 9.x line as no longer supported)
- [ ] 7.6 i18next 25 → 26, react-i18next 16 → 17
- [ ] 7.7 lucide-react 0.563 → 1.x
- [ ] 7.8 Content-Security-Policy (requires replacing `new Function` in `web/src/utils/calculator.ts:24`) [M6]
- [ ] 7.9 Remove empty root `package-lock.json`

---

## Change log

### 2026-10-01 — Stage 1 (items 1.1–1.12)
- Files changed: `go.mod`, `go.sum`, `web/package.json`, `web/package-lock.json`, `Dockerfile`, `build-and-push.sh`, `CLAUDE.md` (version list only). No application code touched.
- Checks passed: `go build ./...` and `go vet ./...` (Go 1.26.2 and 1.27.1), `govulncheck` (0 reachable with 1.27.1, was 18), `npm audit` (0, was 16), `npm run build`, `docker build --pull` of the full image; resulting binary is statically linked and built with go1.27.1.
- Not checked: running the server or the container.
