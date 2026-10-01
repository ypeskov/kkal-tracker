# Update & Security Fix Plan

Source: `reports/security_report-2026-10-01.md` (audit of v5.5.2, commit `4361c46`).
IDs in brackets (H1, M3, L5, ...) refer to findings in that report.

## Status Legend
- [ ] Pending
- [x] Done
- [~] Partly done / done but not fully verified (see note)
- [-] Dropped by the owner's decision (not counted in the totals)

## Progress

| Stage | Done | Total |
|-------|------|-------|
| 1. Dependency & toolchain refresh | 14 | 14 |
| 2. High-priority code fixes | 4 | 4 |
| 3. Server hardening (`server.go`) | 5 | 7 |
| 4. Kubernetes & backup | 5 | 5 |
| 5. Low-priority code fixes | 0 | 0 |
| 6. Non-security defects | 0 | 3 |
| 7. Major upgrades & long-term | 0 | 9 |

---

## Stage 1 — Dependency & toolchain refresh (no code changes) — released as v5.6.0

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
- [x] 1.13 Add `.dockerignore`: `web/node_modules`, `web/dist`, `.git`, `data`, `.env` files and `.deploy.env` are excluded (the last four were sent into the builder stage by `COPY . .`; the final image was not affected by these). Correction to the original note: `web/node_modules` did affect the image, because `COPY web/ ./` overwrote the `npm ci` result with the host's `node_modules`, so the frontend was bundled with whatever was installed locally instead of the lockfile versions
- [x] 1.14 `Dockerfile`: frontend and Go builder stages run on `$BUILDPLATFORM` (Go cross-compiles via `GOARCH=$TARGETARCH`), so building `linux/arm64` on the x86_64 host no longer runs npm/Go under QEMU

Manual (developer machine, not in repo):
- [x] Update local Go 1.26.2 → latest 1.26.x/1.27.x (15 stdlib advisories are reachable when building locally with 1.26.2): the new dev host runs Go 1.27.1

Runtime verification: v5.6.0 is deployed to production; the pod starts cleanly and `/` and `/api/languages` return 200. There are still no automated tests, and the manual smoke test under a user account (login, add entry, weight, export, AI) has not been done.

## Stage 2 — High-priority code fixes — released as v5.6.1

- [x] 2.1 Call `c.Validate(&req)` in `Login` and `Register` (`internal/handlers/auth/handler.go:50,85`) [H2]
- [x] 2.2 Remove `COPY .env.sample .env` from `Dockerfile:38` [H3]
- [x] 2.3 Set `ENVIRONMENT=production` for the production image/deployment [H3]: the image sets it via `ENV`, `kubernetes/overlays/prod/.env.sample` documents it, and the real prod `.env` on the server has it (applied on 2026-10-01)
- [x] 2.4 Make config fail closed: refuse placeholder/short `JWT_SECRET` unless `ENVIRONMENT=development` is set explicitly; an unset `ENVIRONMENT` now means production (`internal/config/config.go`) [H3]

Before deploying 2.2–2.4: confirm the prod ConfigMap provides `JWT_SECRET` (≥32 chars), `ENVIRONMENT`, `LOG_LEVEL`, `PORT`, `DATABASE_PATH`.

Checked on 2026-10-01: the active prod ConfigMap has `LOG_LEVEL=info`, `PORT=8080`, `DATABASE_PATH=/data/app.db`, but `JWT_SECRET` is only 20 characters long and `ENVIRONMENT` is missing. Fixed the same day: `kubernetes/overlays/prod/.env` on the server got a new 64-character `JWT_SECRET` and `ENVIRONMENT=production`, applied with `kubectl apply -k kubernetes/overlays/prod` on image 5.6.0 (all sessions were reset). The pod started cleanly with JSON logs, `/` and `/api/languages` return 200, `/api/auth/me` without a token returns 401. The previous `.env` is kept on the server as `~/kkal-tracker-prod.env.bak-20261001`.

## Stage 3 — Server hardening (`internal/server/server.go`) — released as v5.6.2, login limit removed in v5.6.3

- [~] 3.1 Separate strict rate limiters for `/auth/login` and `/auth/register` [M1]: register is limited to 3 attempts then 1 per 20 min; the group-wide 5 req/s limit stays. The strict login limit (5 attempts then 1 per 20 s) shipped in v5.6.2 and was removed in v5.6.3 by the owner's decision: behind the load balancer it was shared by all users (see 3.7), so anyone could block logins for everybody. Brute-force protection for login is therefore still open until 3.7 is solved. Per-email throttling is deliberately not added (it would let anyone lock a known account out)
- [x] 3.2 Log failed logins at warn level [M1]: email and client IP, never the password; denied rate-limited requests are logged at warn too
- [x] 3.3 Configure `e.IPExtractor` to trust `X-Forwarded-For` only from the cluster network [M2]: new `TRUSTED_PROXIES` env (CIDR list); unset means loopback, link-local and private networks. Prod should set `TRUSTED_PROXIES=10.42.0.0/16` (k3s pod network, Traefik runs there)
- [x] 3.4 `http.Server` timeouts: `ReadHeaderTimeout` 10 s, `ReadTimeout` 30 s, `WriteTimeout` 60 s, `IdleTimeout` 120 s [M5]
- [x] 3.5 `BodyLimit` middleware, 1 MB [M5]
- [x] 3.6 Rate-limit `/api/v1` before the API-key check [L4]
- [ ] 3.7 Infrastructure (outside this repo): make the real client address reach the app. `kcal.peskov.info` resolves to `95.217.168.25`, a load balancer (TCP pass-through) in front of the cluster node (same TLS certificate as on the node), so Traefik and the app see every visitor as `95.217.168.25` and all per-IP limiters (register, AI, `/api/v1`, the auth group limit) are shared by all users. Options: PROXY protocol on the front proxy plus `proxyProtocol.trustedIPs` on the Traefik entry points (affects every app on the cluster, both sides must be switched together), or point DNS straight at the node

## Stage 4 — Kubernetes & backup — closed: items 4.1–4.4 and 4.9 applied to production, 4.5–4.8 dropped

- [x] 4.1 Move `JWT_SECRET`, `SMTP_PASSWORD`, `OPENAI_API_KEY`, `GDRIVE_OAUTH_TOKEN` from ConfigMap to Secret (`secretGenerator` + `secretRef`) [M3]: prod overlay generates Secret `kkal-tracker-secrets` from `.env.secret`; on the server the four keys were moved from `.env` to `.env.secret` (mode 600)
- [x] 4.2 Backup CronJob: pass only `GDRIVE_*` instead of the whole app env [M4]
- [x] 4.3 Pin `rclone/rclone` image by version/digest [M4]: `rclone/rclone:1.75.1@sha256:45401ad7…` is the base of the own backup image
- [x] 4.4 Stop installing `sqlite` from the network on every backup run [M4]: own image `ypeskov/kkal-tracker-backup:1.75.1` (`kubernetes/backup/Dockerfile`) has sqlite3 preinstalled
- [-] 4.5 rclone `scope = drive` → `drive.file` (needs a new OAuth token) [M4]: dropped on 2026-10-01. The token stays with full Drive access and is shared with Orgfin; a narrower scope would need a new browser authorization by the owner, would have to be copied into both projects, and would hide the existing backup folder from rclone
- [-] 4.6 Encrypt backups (rclone `crypt` remote) [M4]: dropped on 2026-10-01, backups stay unencrypted on Google Drive
- [-] 4.7 Deployment `securityContext`: read-only root FS, drop capabilities, no privilege escalation, seccomp [L1]: dropped on 2026-10-01
- [-] 4.8 Deployment resource requests/limits and liveness/readiness probes [L1]: dropped on 2026-10-01
- [x] 4.9 Backups were broken: the nightly job had been failing since 2026-09-14 (last success 2026-09-13). The local snapshot was created, the upload to Google Drive failed with `invalid_grant`: the OAuth token was re-issued for Orgfin around 2026-09-15, which revoked the old one still used here. Fixed on 2026-10-01 by copying the working token from the Orgfin prod ConfigMap into the prod `.env`; a manual run uploaded the snapshot to Google Drive and cleaned up local snapshots older than 7 days. Both projects now share one token, so re-issuing it for one of them breaks the other until it is copied over. Side effect: the job exits before its cleanup step, so local snapshots pile up in `/data/backups` (4 per night: the run plus three retries)

## Stage 5 — Low-priority code fixes — dropped entirely on 2026-10-01 by the owner's decision

- [-] 5.1 Nil-pointer panic on weight update without `recorded_at` (`internal/repositories/weight_history.go:172`) [L5]
- [-] 5.2 Enable SQLite `foreign_keys` and `busy_timeout` pragmas (`internal/database/database.go:19`); check existing data for orphans first [L6]
- [-] 5.3 Return 404 instead of 500 for unknown API key on revoke/delete (`internal/handlers/apikey/handler.go:115,131`) [L7]
- [-] 5.4 Max-length validation for `food`, ingredient `name`, `first_name`, `last_name` [L9]
- [-] 5.5 Constant-time login for unknown emails; decide on register 409 behaviour [L3]
- [-] 5.6 Hash activation tokens at rest [L8]
- [-] 5.7 JWT: recheck user `is_active` per request or shorten lifetime; add `iss` [L2]

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

### 2026-10-01 — v5.6.0 release (items 1.13 partly, 1.14)
- Files changed: `Dockerfile` (build stages on `$BUILDPLATFORM`), `.dockerignore` (new), `version.txt`, `kubernetes/base/deployment.yaml`.
- Why: the first `linux/arm64` build on the x86_64 host failed in `npm run build` with an esbuild host/binary version mismatch (0.27.2 vs 0.28.2), caused by the host's stale `web/node_modules` overwriting the `npm ci` result; the build stages also ran under QEMU for no reason.
- Released: image `ypeskov/kcal-tracker:5.6.0` (`linux/arm64`) pushed to Docker Hub, `develop` merged into `master`, deployed to production with `kubectl apply -k kubernetes/overlays/prod`.
- Checks passed: image builds and is `linux/arm64` in the registry; rollout succeeded, pod runs the pushed digest with 0 restarts; `https://kcal.peskov.info/` and `/api/languages` return 200.
- Not checked: any authenticated flow (login, entries, weight, export, AI).

### 2026-10-01 — Stage 2 (items 1.13, 2.1–2.4), released as v5.6.1
- Files changed: `internal/handlers/auth/handler.go` (validation in `Login`/`Register`), `internal/config/config.go` (fail-closed JWT secret check, unset `ENVIRONMENT` means production), `cmd/web/main.go` (no warning when `.env` is absent), `Dockerfile` (no `.env` in the image, `ENV ENVIRONMENT=production`), `.dockerignore`, `kubernetes/overlays/prod/.env.sample`.
- Tests added (first Go tests in the repo): `internal/config/config_test.go`, `internal/handlers/auth/handler_test.go`.
- Checks passed: `go build ./...`, `go vet ./...`, `go test ./...`, local `docker build` (image is `arm64`, contains no `/app/.env`, has `ENVIRONMENT=production`).
- Behaviour change for local runs: `ENVIRONMENT=development` must be set (the root `.env.sample` already has it), otherwise the app, `cmd/migrate`, `cmd/seed` and `scripts/create_user.go` demand a strong `JWT_SECRET`.
- Not checked: running the server or the container; login of existing users whose stored email would not pass the `email` validator (registration did not validate before).
- Released: image `ypeskov/kcal-tracker:5.6.1` (`linux/arm64`, built natively on the arm64 host) pushed to Docker Hub, `develop` merged into `master`, deployed to production with `kubectl apply -k kubernetes/overlays/prod` after a `kubectl diff` preview (the only change was the image tag).
- Production checks passed: rollout succeeded, pod runs the pushed digest with 0 restarts; `/` and `/api/languages` return 200; `/api/auth/me` without a token returns 401; login with a malformed email returns 400, with unknown credentials 401.
- Still not checked: any authenticated flow (login of a real user, entries, weight, export, AI).

### 2026-10-01 — Stage 3 (items 3.1–3.6), released as v5.6.2
- Files changed: `internal/server/server.go`, `internal/server/security.go` (new: IP extractor, rate limiter helper, limits), `internal/handlers/auth/handler.go` (warn logs, per-route limiters), `internal/config/config.go` (`TRUSTED_PROXIES`, file run through `gofmt`), `go.mod` (`golang.org/x/time` became a direct dependency, same version), `.env.sample`, `kubernetes/overlays/prod/.env.sample`, `CLAUDE.md`.
- Tests added: `internal/server/security_test.go` (IP extraction, rate limits, body limit, timeouts, `/api/v1` limiter order, spoofed `X-Forwarded-For`), more cases in the config and auth handler tests.
- Checks passed: `go build ./...`, `go vet ./...`, `go test -race ./...`, local `docker build`.
- Not checked: running the server; that Traefik passes the real client address in `X-Forwarded-For` (Traefik has no access log, so this has to be confirmed after deploy from the `remote_ip` of a failed-login warning). If Traefik saw every client as one cluster address, the login limit would be shared by all users.
- Not covered (outside the plan items): `smtp.SendMail` still has no timeout (M5 detail); the frontend shows the generic "Login failed" for HTTP 429.
- Released: image `ypeskov/kcal-tracker:5.6.2` (`linux/arm64`) pushed to Docker Hub, `develop` merged into `master`, `TRUSTED_PROXIES=10.42.0.0/16` added to the prod `.env`, deployed with `kubectl apply -k kubernetes/overlays/prod` after a `kubectl diff` preview (image tag and the regenerated ConfigMap only; `JWT_SECRET` unchanged, sessions kept).
- Production checks passed: rollout succeeded, pod runs the pushed digest with 0 restarts; `/` and `/api/languages` return 200; `/api/auth/me` without a token 401; failed login 401 and logged at warn; second `/api/v1/data` request without a key 429; 2 MB body 413; a spoofed `X-Forwarded-For: 1.2.3.4` did not end up in the logged `remote_ip`.
- Client address check: a request sent straight to the node is logged with the real client address, so Traefik and the extractor work as intended. A request through the public DNS name is logged as `95.217.168.25` (the front proxy) — see item 3.7.
- Still not checked: any authenticated flow (login of a real user, entries, weight, export, AI).

### 2026-10-01 — v5.6.3: strict login rate limit removed
- Why: the load balancer hides client addresses (item 3.7), so the per-IP login limit was one bucket for all users. The owner decided to remove it; the register limit, warn logging of failed logins and the rest of Stage 3 stay.
- Files changed: `internal/server/server.go`, `internal/server/security.go`, `internal/handlers/auth/handler.go`, `internal/server/security_test.go`, `CLAUDE.md`.
- Checks passed: `go build ./...`, `go vet ./...`, `go test -race ./...`.

### 2026-10-01 — Stage 4, items 4.1–4.4 and 4.9
- Files changed: `kubernetes/base/deployment.yaml` (`secretRef`), `kubernetes/base/cronjob-backup.yaml` (own pinned image, explicit `GDRIVE_*` env instead of the whole ConfigMap), `kubernetes/base/configmap-backup.yaml` (no `apk add`), `kubernetes/overlays/prod/kustomization.yaml` (`secretGenerator`), `kubernetes/overlays/prod/.env.sample` and new `.env.secret.sample`, new `kubernetes/backup/Dockerfile`, `.gitignore`, `.dockerignore`, `CLAUDE.md`.
- Checks passed: `kustomize build` (v5.6.0, same as on the server) of the prod overlay with the sample env files — ConfigMap holds only non-sensitive keys, Secret holds the four secrets, name references in the Deployment and CronJob are rewritten; the backup image builds for `arm64` (rclone 1.75.1, sqlite 3.53.4); the backup script run in that image against a throwaway database produced a snapshot that restores and passes `integrity_check`.
- Found and fixed on the way: production backups had been failing since 2026-09-14, see item 4.9.
- Applied to production: `ypeskov/kkal-tracker-backup:1.75.1` pushed to Docker Hub (new repository), the prod `.env` split into `.env` and `.env.secret`, `kubectl apply -k kubernetes/overlays/prod` after a `kubectl diff` preview. The app image is unchanged (5.6.3), `JWT_SECRET` is unchanged, sessions were kept.
- Production checks passed: the app pod restarted cleanly with both env sources; `/` and `/api/languages` return 200, `/api/auth/me` without a token and a failed login return 401; a manual backup run on the new image received only the `GDRIVE_*`, `DATABASE_PATH`, `BACKUP_PATH` and `RETENTION_DAYS` variables, installed nothing from the network, and uploaded the snapshot to Google Drive.
- Not checked: the scheduled nightly run (02:00 Europe/Sofia) on the new setup.
- Not done: removal of the old generated `kkal-tracker-env-*` ConfigMaps, which still hold previous secret values (kustomize does not prune them).
