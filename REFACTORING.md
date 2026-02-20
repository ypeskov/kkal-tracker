# Refactoring Issues

## Status Legend
- [ ] Pending
- [x] Done
- [-] Skipped / Won't Fix

---

## HIGH Priority

### [x] H1 — No service interfaces (Dependency Inversion)
**Files:** All handlers (calories, weight, auth, ingredients, profile, ai, export, apikey, apidata)

Handlers depend on concrete `*service.Service` types, making unit testing impossible and creating tight coupling.

```go
// Current — concrete type
calorieService *calorieservice.Service  // calories/handler.go:16

// Target — interface
type CalorieServicer interface {
    GetEntries(ctx context.Context, userID int, date time.Time) ([]models.CalorieEntry, error)
    // ...
}
```

**Action:** Define an interface in each handler package (or in the service package) and update handler structs to use it.

---

### H2 — No `context.Context` in service methods
**Files:** `internal/services/calorie/service.go`, `weight/service.go`, `auth/service.go`, `ingredient/service.go`, `profile/service.go`, `reports/service.go`, `export/service.go`

Services don't accept `context.Context`, so HTTP request cancellation, timeouts, and deadline propagation never reach the database layer. The AI service is the only exception — it does this correctly.

```go
// Current
func (s *Service) GetEntries(userID int, date time.Time) ([]models.CalorieEntry, error)

// Target
func (s *Service) GetEntries(ctx context.Context, userID int, date time.Time) ([]models.CalorieEntry, error)
```

**Action:** Add `ctx context.Context` as the first parameter to all public service methods, pass it down to repository calls.

---

## MEDIUM Priority

### M1 — Business logic in AI handler
**File:** `internal/handlers/ai/handler.go`

`aggregateNutritionData()` (line ~191) and `convertWeightData()` (line ~220) are handler methods that perform domain-level data transformation. Goal progress calculation (lines ~114–158) also lives in the handler.

**Action:** Move these methods into `internal/services/ai/service.go`. The handler should only call `aiService.Analyze(ctx, req)` and return the response.

---

### M2 — Direct repository access in handlers
**Files:**
- `internal/handlers/export/handler.go:16,52` — injects and calls `userRepo.GetByID()` directly
- `internal/handlers/ai/handler.go:27` — injects `userRepo` directly

Handlers should not interact with repositories. All data access must go through a service.

**Action:**
- Export handler: pass user language through the service request DTO or have the export service fetch the user internally.
- AI handler: move user lookup into `aiService`.

---

### M3 — Duplicate validation (handler struct tags + service manual checks)
**Files:**
- `internal/handlers/calories/dto.go` — struct tag validation
- `internal/services/calorie/service.go:32–49` — duplicate manual range checks
- `internal/services/ingredient/service.go:101–140` — same pattern

The same constraints are enforced twice. If they diverge, behavior becomes inconsistent.

**Action:** Choose one layer for validation. Prefer struct tag validation at the handler boundary; remove the redundant manual checks in services (or keep only cross-field / business-rule validation that can't be expressed in tags).

---

### M4 — Profile service using `db.Begin()` directly
**File:** `internal/services/profile/service.go:96–120`

The service holds a `*sql.DB` reference and manages transactions itself, coupling business logic to the database driver.

**Action:** Move transaction management to the repository layer. Introduce a `WithTx` or `RunInTx` pattern in the repository, or use a Unit of Work object passed to repositories.

---

### M5 — `IngredientRepository` interface too broad
**File:** `internal/repositories/interfaces.go`

`IngredientRepository` has 8 methods mixing global ingredient operations with user-specific ingredient operations. Implementations must satisfy both responsibilities.

**Action:** Split into two interfaces:
```go
type GlobalIngredientRepository interface { ... }
type UserIngredientRepository interface { ... }
```

---

### M6 — Models with embedded business logic
**File:** `internal/models/user.go:32–48`

`SetPassword()` and `CheckPassword()` live on the `User` model struct. Models should be plain data structures; password logic belongs in `internal/services/auth`.

**Action:** Move password hashing/verification to `auth/service.go`. The `User` struct should only hold data fields.

---

## LOW Priority

### L1 — Duplicated ID param parsing (9 occurrences)
**Files:** `calories/handler.go:94,112`, `weight/handler.go:112,156`, `ingredients/handler.go:60,118,162`, `apikey/handler.go:110,126`

```go
// Repeated pattern
id, err := strconv.Atoi(c.Param("id"))
if err != nil {
    return echo.NewHTTPError(http.StatusBadRequest, "Invalid ... ID")
}
```

**Action:** Extract a shared helper, e.g. in `internal/handlers/helpers.go`:
```go
func parseIDParam(c echo.Context) (int, error)
```

---

### L2 — Duplicated bind + validate pattern (24 occurrences)
**Files:** All CRUD handlers (calories, weight, ingredients, auth, profile, apikey, export)

```go
// Repeated in every handler
if err := c.Bind(&req); err != nil {
    return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
}
if err := c.Validate(&req); err != nil {
    return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}
```

**Action:** Extract:
```go
func bindAndValidate[T any](c echo.Context, req *T) error
```

---

### L3 — Duplicated date parsing (5+ occurrences)
**Files:** `weight/handler.go:88,131`, `apidata/handler.go:72,75`, `profile/service.go:128`, `calorie/service.go:119`

Magic string `"2006-01-02"` repeated everywhere.

**Action:** Define a package-level constant and a helper:
```go
const DateFormat = "2006-01-02"

func parseDate(s string) (time.Time, error) {
    return time.Parse(DateFormat, s)
}
```

---

### L4 — `==` instead of `errors.Is()` for sentinel errors
**File:** `internal/handlers/ingredients/handler.go:69,148,171`

```go
// Current — breaks with wrapped errors
if err == sql.ErrNoRows { ... }

// Target
if errors.Is(err, sql.ErrNoRows) { ... }
```

**Action:** Replace all direct sentinel error comparisons with `errors.Is()`.

---

### L5 — Inconsistent constructor naming
**Files:**
- `calories/handler.go` → `New()`
- `auth/handler.go` → `NewHandler()`
- `profile/handler.go` → `NewProfileHandler()`

**Action:** Standardize to `New()` across all handler packages (idiomatic Go — package name provides context).

---

### L6 — Magic date format constant scattered
**Files:** 5+ files (see L3)

Tracked as part of L3 — extract `DateFormat` constant to a shared location.

---

### L7 — Inconsistent DTO ownership
Some DTOs live in the handler package (`calories/dto.go`), others in the service package (`calorie/dto.go`, `profile/types.go`).

**Action:** Decide on one convention:
- **Option A:** Request/response DTOs in handler package; domain types in service package.
- **Option B:** All types in service package; handlers import from service.

---
