# Security Audit Report — Kkal-tracker

**Date**: 2026-02-12
**Overall Risk Level**: HIGH

## Findings Overview

| Severity | Count |
|----------|-------|
| CRITICAL | 2 |
| HIGH | 3 |
| NORMAL | 6 |
| LOW | 4 |

---

## CRITICAL

### 1. Stored XSS via AI Analysis Response (`dangerouslySetInnerHTML`)

- **File**: `web/src/components/ai/AIAnalysisPanel.tsx:59-62`
- **Description**: AI analysis result is rendered as raw HTML via `dangerouslySetInnerHTML` without sanitization. An attacker could craft a food name containing JavaScript (e.g., `<img src=x onerror=alert(document.cookie)>`), which might be reflected in the AI response and executed in the browser.
- **Impact**: Session hijacking, account takeover, data exfiltration.
- **Fix**: Install `dompurify` and sanitize: `DOMPurify.sanitize(result.analysis)`
- **Status**: [x] FIXED

### 2. Insecure Default JWT Secret Key

- **File**: `internal/config/config.go:53`
- **Description**: JWT secret defaults to `"default-secret-key"` if `JWT_SECRET` env var is not set. Anyone reading source code can forge JWT tokens for any user.
- **Impact**: Complete authentication bypass.
- **Fix**: Refuse to start in production with default secret; require minimum key length.
- **Status**: [x] FIXED

---

## HIGH

### 3. Overly Permissive CORS Configuration (`AllowOrigins: *`)

- **File**: `internal/server/server.go:110`
- **Description**: `echomiddleware.CORS()` without config allows requests from any origin. Any website can make authenticated cross-origin requests.
- **Impact**: Cross-site request forgery, data theft.
- **Fix**: Restrict `AllowOrigins` to the application's own domain.
- **Status**: [x] FIXED

### 4. Calorie Entry IDOR — Data Leak via Update Return Value

- **File**: `internal/repositories/calorie_entry.go:173-191`
- **Query**: `QueryGetCalorieEntryByID` in `internal/repositories/queries.go:274-283`
- **Description**: After UPDATE (correctly scoped by `user_id`), `GetByID(id)` returns the entry without `user_id` filter. If UPDATE affects 0 rows (wrong user), the response still contains the other user's data.
- **Impact**: Attacker can read any user's calorie entry details.
- **Fix**: Check `rowsAffected == 0` and return error; add `user_id` to `GetByID` query.
- **Status**: [x] FIXED

### 5. Panic on Short Activation Token Input

- **Files**: `internal/handlers/auth/handler.go:133`, `internal/services/auth/service.go:183`, `internal/repositories/activation_token.go:87`
- **Description**: `token[:8]+"..."` without length check causes panic (index out of range) for tokens shorter than 8 chars.
- **Impact**: DoS via crafted request to `GET /api/auth/activate/abc`.
- **Fix**: Check `len(token) < 8` before slicing.
- **Status**: [x] FIXED

---

## NORMAL

### 6. Missing Input Validation in Calories Handler

- **File**: `internal/handlers/calories/handler.go:52,105`
- **Description**: `c.Validate(&req)` is not called after `c.Bind()` in CreateEntry and UpdateEntry, bypassing validation tags.
- **Fix**: Add `c.Validate(&req)` after binding.
- **Status**: [x] FIXED

### 7. Excel Formula Injection in Exported Food Names

- **File**: `internal/services/export/excel.go:148`
- **Description**: Food names starting with `=`, `+`, `-`, `@` are written directly to Excel cells and may be interpreted as formulas.
- **Fix**: Prefix formula-triggering values with a single quote.
- **Status**: [x] FIXED

### 8. Sensitive Data in Kubernetes ConfigMap Instead of Secret

- **File**: `kubernetes/base/deployment.yaml:24-26`
- **Description**: JWT_SECRET, SMTP_PASSWORD, API keys stored in ConfigMap (plain text) instead of Kubernetes Secrets.
- **Fix**: Split sensitive values into a K8s Secret resource.
- **Status**: [ ] N/A — .env already gitignored, secrets managed outside repo

### 9. Hardcoded SQLite Query in Ingredient Repository

- **File**: `internal/repositories/ingredient.go:107-111`
- **Description**: Inline SQL with `?` placeholders will fail on PostgreSQL.
- **Fix**: Move to `queries.go` with dialect-aware variants.
- **Status**: [x] FIXED

### 10. No Password Maximum Length (bcrypt 72-byte Truncation)

- **File**: `internal/handlers/auth/handler.go:22`
- **Description**: Only `min=6` validation. bcrypt silently truncates at 72 bytes.
- **Fix**: Add `max=72` to password validation tag.
- **Status**: [x] FIXED

### 11. In-Memory Rate Limiter (Not Multi-Instance Safe)

- **File**: `internal/server/server.go:151,179,192`
- **Description**: Rate limiters use in-memory store; ineffective if scaled to multiple replicas.
- **Fix**: Document limitation or switch to Redis-backed limiter if scaling.
- **Status**: [x] FIXED (documented)

---

## LOW

### 12. GetCurrentUser Returns Raw Model Instead of DTO

- **File**: `internal/handlers/auth/handler.go:128`
- **Description**: Returns full `models.User` struct. `PasswordHash` is excluded via `json:"-"`, but future fields may leak.
- **Fix**: Use a dedicated response DTO.
- **Status**: [ ] Deferred — `json:"-"` protects currently, full DTO refactor is a future task

### 13. Missing Security Headers

- **File**: `internal/server/server.go`
- **Description**: No X-Content-Type-Options, X-Frame-Options, HSTS, or CSP headers.
- **Fix**: Add Echo's Secure middleware.
- **Status**: [x] FIXED

### 14. Content-Disposition Header Injection in Export

- **File**: `internal/handlers/export/handler.go:91`
- **Description**: Date strings used in filename without format validation.
- **Fix**: Add `datetime=2006-01-02` validation to export DTO.
- **Status**: [x] FIXED

### 15. `.env.sample` Copied into Production Docker Image

- **File**: `Dockerfile:38`
- **Description**: Reveals configuration structure and default values.
- **Fix**: Remove `COPY .env.sample .env` from Dockerfile.
- **Status**: [x] FIXED

---

## Positive Observations

- All SQL queries are parameterized (no SQL injection)
- bcrypt with DefaultCost for password hashing
- JWT algorithm verification (HMAC check prevents algorithm confusion)
- API keys: crypto/rand + SHA-256 hashing, raw key shown once
- `json:"-"` on PasswordHash and KeyHash fields
- Distroless Docker image + non-root user
- No `err.Error()` leakage in HTTP 500 responses
- Proper `.gitignore` for secrets and build artifacts
