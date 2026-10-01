# Update & Security Audit Report — Kkal-tracker

**Date**: 2026-10-01
**Audited revision**: `4361c46` (v5.5.2)
**Scope**: dependency/toolchain freshness, backend (Go), frontend (React), Docker, Kubernetes manifests.
**Method**: `govulncheck`, `npm audit`, `go list -m -u`, npm registry comparison, manual read of all handlers, middleware, services, repositories and deployment files. No code was changed.

## Summary

| Area | State |
|------|-------|
| Go toolchain in Docker | **End of life** (`golang:1.25`; supported lines are 1.26 and 1.27) |
| Go modules | 7 direct modules behind; 18 reachable advisories (2 modules + stdlib) |
| npm packages | 16 advisories (1 critical, 10 high); all fixable inside current semver ranges |
| Code findings | 0 critical, 3 high, 6 medium, 9 low |
| Automated tests | None (`*_test.go`: 0, frontend tests: 0) |

Nothing found allows unauthenticated data access with the current code. The highest risks are operational: an unsupported toolchain, a fail-open production configuration, and input validation that is declared but never executed on the auth endpoints.

Two items marked FIXED in `security_report-2026-02-12.md` are not actually in effect: #10 (password max length) and #15 (`.env.sample` in the image). See H2 and H3.

---

## Part 1 — What to update (no business-logic changes)

### 1.1 Toolchain and base images

| Item | Current | Target | Notes |
|------|---------|--------|-------|
| Go (Dockerfile:12, `go.mod`) | `golang:1.25`, `go 1.25.0` | `golang:1.27` (or 1.26) | 1.25 ended at 1.25.14; fixes shipped in 1.26.7+/1.27.x are not backported |
| Local Go | 1.26.2 | 1.26.8 / 1.27.1 | 14 stdlib advisories reachable from this code are fixed in 1.26.3–1.26.6 |
| Node (Dockerfile:4) | `node:22-alpine` | `node:24-alpine` | 22 is maintenance LTS; `@types/node` is already `^25` |
| Runtime image (Dockerfile:32) | `distroless/base-debian12:nonroot` | `distroless/static-debian13:nonroot` | Binary is `CGO_ENABLED=0`, so `static` is enough (no glibc/openssl in the image) |
| Backup image (cronjob-backup.yaml:29) | `rclone/rclone:latest` | pinned version or digest | See M4 |

Related build hygiene:
- `Dockerfile:7` uses `npm install`; use `npm ci` so the image is built from the lockfile.
- `build-and-push.sh:76` uses `--no-cache` but not `--pull`, so stale local base images are reused.

### 1.2 Go modules

Direct dependencies:

| Module | Current | Latest | Advisory |
|--------|---------|--------|----------|
| labstack/echo/v4 | 4.15.0 | 4.16.0 | GO-2026-6293 (reachable, fixed in 4.15.3): `%2F` bypass exposing static files |
| xuri/excelize/v2 | 2.10.0 | 2.11.0 | GO-2026-5960, -6452, -6453 (parser DoS; app only writes files) |
| golang.org/x/crypto | 0.47.0 | 0.57.0 | GO-2026-5005 (ssh/agent, not used) |
| modernc.org/sqlite | 1.44.3 | 1.60.1 | — |
| pressly/goose/v3 | 3.26.0 | 3.28.0 | — |
| go-playground/validator/v10 | 10.30.1 | 10.30.5 | — |
| sashabaranov/go-openai | 1.41.2 | 1.43.0 | — |

Up to date: `golang-jwt/jwt/v5` 5.3.1, `joho/godotenv` 1.5.1.

Indirect modules with advisories: `golang.org/x/net` 0.49.0 → 0.59.0 (GO-2026-4918, -5026 reachable; six more unreachable), `golang.org/x/text` 0.33.0 → 0.42.0 (GO-2026-5970).

All of the above are minor/patch bumps: `go get -u ./... && go mod tidy`. Echo v5 exists but is a major migration and is not needed now.

### 1.3 npm packages

In-range updates (`npm update` / `npm audit fix`, no `--force`) clear all 16 advisories:

| Package | Locked | Latest in range | Why it matters |
|---------|--------|-----------------|----------------|
| dompurify | 3.3.1 | 3.4.16 | **Runtime**. 18 advisories incl. mutation-XSS; it is the only sanitizer in front of `dangerouslySetInnerHTML` |
| @tanstack/react-router | 1.157.18 | 1.170.41 | Pulls `seroval` ≤1.5.2 (critical, deserialization) |
| vite | 7.3.1 | 7.3.x latest | 5 dev-server advisories (file read, `fs.deny` bypass) |
| postcss | 8.5.6 | 8.5.28 | 4 advisories (build time) |
| react / react-dom | 19.2.4 | 19.3.0 | — |
| @tanstack/react-query | 5.90.20 | 5.104.0 | — |
| tailwindcss, @tailwindcss/postcss | 4.1.18 | 4.3.3 | — |
| date-fns | 4.1.0 | 4.4.0 | — |
| i18next / react-i18next | 25.8.0 / 16.5.4 | 25.10.10 / 16.6.6 | — |

Transitive, build/lint-time only: rollup, picomatch, minimatch, brace-expansion, js-yaml, flatted, nanoid, browserslist, ajv, @babel/core.

Major versions available (separate task each, need testing):

| Package | Locked | Latest |
|---------|--------|--------|
| vite / @vitejs/plugin-react | 7.3.1 / 5.1.2 | 8.3.1 / 6.1.1 |
| typescript | 5.9.3 | 7.0.2 |
| eslint / @eslint/js | 9.39.2 | 10.11.0 / 10.0.1 |
| i18next / react-i18next | 25.8.0 / 16.5.4 | 26.4.2 / 17.0.15 |
| lucide-react | 0.563.0 | 1.49.0 |

Cleanup:
- `@types/dompurify` is deprecated (dompurify ships its own types) and sits in `dependencies`; remove it.
- No ESLint config is tracked (`eslint.config.js` missing), so `npm run lint` cannot run.
- Root `package-lock.json` is an empty stray file.

---

## Part 2 — Security findings

### HIGH

#### H1. Unsupported Go toolchain and reachable advisories in the shipped binary
- **Files**: `Dockerfile:12`, `go.mod`
- **Detail**: See 1.1 and 1.2. `govulncheck` reports 18 advisories with call traces from this code (echo static handler, `net/http`, `crypto/tls`, `html/template`, `net/mail`, `net/url`).
- **Fix**: Build with Go 1.27 (or 1.26), bump modules, rebuild with `--pull`.
- **Status**: [ ]

#### H2. Auth endpoints never run validation
- **File**: `internal/handlers/auth/handler.go:50`, `:85`
- **Detail**: `Login` and `Register` call `c.Bind` but not `c.Validate`, so the `validate:"..."` tags on `LoginRequest`/`RegisterRequest` are dead. Every other handler validates (13 binds, 11 validates; the two missing are here).
- **Impact**: Registration accepts an empty or one-character password, any string as email, and any `language_code`. A password over 72 bytes makes bcrypt fail and returns 500. The unvalidated email string is written into the `To:` header and passed to SMTP (`net/smtp` rejects CR/LF, so header injection is blocked, but junk accounts and SMTP errors are not).
- **Fix**: Add `c.Validate(&req)` after both binds. Consider raising `min=6`.
- **Status**: [ ] (regression of 2026-02-12 #10)

#### H3. Production safety checks are fail-open; dev defaults are baked into the image
- **Files**: `Dockerfile:38`, `internal/config/config.go:50-72`, `kubernetes/overlays/prod/.env.sample`
- **Detail**: The image copies `.env.sample` to `/app/.env`, and `godotenv.Load()` reads it. It sets `ENVIRONMENT=development`, `LOG_LEVEL=debug`, `JWT_SECRET=your-jwt-secret-key-change-this-in-production`, `OPENAI_API_KEY=sk-...`. The JWT strength checks only run when `ENVIRONMENT=production`, and the prod env sample does not set `ENVIRONMENT` or `LOG_LEVEL`.
- **Impact**: Unless the real prod ConfigMap sets them, production runs in development mode with debug logs (emails, food names, profile fields). If `JWT_SECRET` is ever missing from the ConfigMap, the app starts with a publicly known secret and only prints a warning; anyone could then forge tokens for any user.
- **Verify** (prints no secret values): `kubectl get configmap -o name | grep kkal-tracker-env` then `kubectl get <name> -o jsonpath='{.data.ENVIRONMENT}{"\n"}{.data.LOG_LEVEL}{"\n"}'`
- **Fix**: Remove `COPY .env.sample .env`; set `ENVIRONMENT=production` via `ENV` in the Dockerfile or the deployment; make the app refuse a weak/placeholder secret unless `ENVIRONMENT=development` is set explicitly.
- **Status**: [ ] (2026-02-12 #15 was marked fixed but the line is still present)

### MEDIUM

#### M1. Weak brute-force and abuse protection on auth routes
- **File**: `internal/server/server.go:163-165`
- **Detail**: One limiter for all `/api/auth/*`: 5 req/s per IP (~430k login attempts per day per IP), no per-account throttling or lockout. `/register` sends an email to any address through the configured SMTP account at the same rate.
- **Impact**: Password guessing; email bombing of third parties and risk of the Gmail account being suspended.
- **Fix**: Separate, much stricter limiters for login and register (per IP and per email); log failed logins at warn level (today they are debug-only and the request logger is disabled).
- **Status**: [ ]

#### M2. Rate-limit key comes from a client-controllable header
- **File**: `internal/server/server.go:102` (no `e.IPExtractor`)
- **Detail**: Without an extractor, Echo's `RealIP()` takes the first `X-Forwarded-For` value. This is safe only while Traefik strips client-supplied forwarding headers (its default); any change there, or direct access to the Service, lets a client rotate the header and bypass all three limiters, including the AI cost limiter.
- **Fix**: `e.IPExtractor = echo.ExtractIPFromXFFHeader(...)` trusting only the cluster network.
- **Status**: [ ]

#### M3. Secrets delivered through a ConfigMap
- **Files**: `kubernetes/overlays/prod/kustomization.yaml`, `kubernetes/base/deployment.yaml:24-26`
- **Detail**: `JWT_SECRET`, `SMTP_PASSWORD`, `OPENAI_API_KEY`, `GDRIVE_OAUTH_TOKEN` are generated into ConfigMap `kkal-tracker-env`. ConfigMaps are not access-controlled or encrypted as Secrets are and appear in `kubectl describe`.
- **Fix**: `secretGenerator` for the sensitive keys, `secretRef` in the pod spec.
- **Status**: [ ] (2026-02-12 #8 was closed as N/A; the gitignore argument does not address in-cluster exposure)

#### M4. Backup job: unpinned third-party image receives every application secret
- **Files**: `kubernetes/base/cronjob-backup.yaml:29,36-38`, `kubernetes/base/configmap-backup.yaml:31,63`
- **Detail**: `rclone/rclone:latest` runs nightly with `envFrom` the full app ConfigMap (JWT secret, SMTP password, OpenAI key), read-write access to the database volume, and installs `sqlite` from the network on each run. The rclone token uses `scope = drive` (full Drive access). Backups are uploaded unencrypted and contain password hashes, API-key hashes and health data.
- **Fix**: Pin the image by digest; pass only `GDRIVE_*`; use `scope = drive.file`; consider an rclone `crypt` remote.
- **Status**: [ ]

#### M5. No HTTP server timeouts or body size limit
- **File**: `internal/server/server.go:213-216`
- **Detail**: `http.Server` has no `ReadHeaderTimeout`/`ReadTimeout`/`IdleTimeout`; no `BodyLimit` middleware. `smtp.SendMail` also has no timeout and runs inside the request.
- **Fix**: Set server timeouts and `echomiddleware.BodyLimit("1M")`.
- **Status**: [ ]

#### M6. No Content-Security-Policy; token readable by any script
- **Files**: `internal/server/server.go:115-120`, `web/src/api/auth.ts`, `web/src/components/ai/AIAnalysisPanel.tsx:62`
- **Detail**: The JWT lives in `sessionStorage`, so any XSS yields account takeover. The single HTML sink is guarded only by DOMPurify 3.3.1 (see 1.3). No CSP is sent as a second layer.
- **Fix**: Update DOMPurify now. Add a CSP later; `web/src/utils/calculator.ts:24` uses `new Function` (input is strictly whitelisted, so it is safe) and would need `'unsafe-eval'` or a small parser.
- **Status**: [ ]

### LOW

| # | Finding | Location |
|---|---------|----------|
| L1 | Pod has no `securityContext` (read-only root FS, drop capabilities, seccomp), no resource limits, no probes | `kubernetes/base/deployment.yaml` |
| L2 | JWT valid 24h with no revocation; user existence / `is_active` not rechecked per request; no `iss`/`aud` | `internal/auth/jwt.go`, `internal/middleware/auth.go` |
| L3 | User enumeration: register returns 409 for existing email; login returns faster for unknown email (no bcrypt) | `internal/handlers/auth/handler.go:96`, `internal/services/auth/service.go:44-56` |
| L4 | `/api/v1` limiter runs after the API-key check, so requests with invalid keys are never limited (one DB lookup each) | `internal/server/server.go:204-205` |
| L5 | Nil-pointer panic when a weight entry is updated without `recorded_at`; the row is already updated, the client gets 500 | `internal/repositories/weight_history.go:172` |
| L6 | SQLite `foreign_keys` pragma is never enabled, so every `ON DELETE CASCADE` in the migrations is inert; registration rollback leaves orphaned `user_ingredients`. No `busy_timeout`/WAL | `internal/database/database.go:19` |
| L7 | Revoke/delete of an unknown API key returns 500 instead of 404 | `internal/handlers/apikey/handler.go:115,131` |
| L8 | Activation tokens stored in plaintext (API keys are hashed) | `internal/repositories/activation_token.go` |
| L9 | No length limits on `food`, ingredient `name`, `first_name`, `last_name` | handler DTOs |

### Checked and found sound
- All SQL is static and parameterized; every update/delete and per-row read is scoped by `user_id`. The 2026-02-12 IDOR fix holds.
- JWT signing method is checked; API keys use `crypto/rand` (256 bit) with SHA-256 at rest and are shown once.
- No internal error text reaches clients in 5xx responses.
- CORS is restricted to `APP_URL`; `X-Frame-Options`, `nosniff`, HSTS are set.
- Excel export escapes formula prefixes; export filenames are built from validated dates.
- Activation email uses `html/template`.
- No secrets found in tracked files or in git history (305 commits scanned for key patterns).
- Runtime image is distroless and non-root.

---

## Part 3 — Non-security defects found on the way

1. **AI system prompt is always empty.** `prompts/system.txt:34-48` references `.TargetWeight`, `.CurrentWeight`, `.TargetDate`, `.GoalProgress`, but `systemPromptData` (`internal/services/ai/openai.go:122-126`) has only `Language`, `Age`, `Height`. `text/template` fails on the missing field, `buildSystemPrompt` logs an error and returns `""`. Reproduced: `can't evaluate field TargetWeight in type systemPromptData`. The model therefore gets no language, format or goal instructions.
2. **PostgreSQL mode does not work.** `database.New` always opens SQLite and no Postgres driver is imported; `DATABASE_TYPE=postgres` would run Postgres-dialect SQL against SQLite.
3. **Export email declares `quoted-printable` but sends the HTML body unencoded** (`internal/services/email/service.go:151`).
4. **No automated tests**, so none of the upgrades above can be verified other than by hand.

---

## Suggested order

1. In-range dependency refresh: `npm update` (web), `go get -u ./... && go mod tidy`, Go 1.27 in Dockerfile/go.mod, `npm ci`, `--pull`. No code changes expected.
2. H2 (two `c.Validate` calls) and H3 (Dockerfile line + `ENVIRONMENT=production`).
3. M1, M2, M5: all in `server.go`.
4. M3, M4, L1: Kubernetes manifests.
5. Major npm upgrades and CSP as separate tasks, after at least smoke tests exist.
