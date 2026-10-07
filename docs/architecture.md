# Architecture

## Stack
- **Backend**: Go 1.26+ (Docker builds with 1.27), Echo v5 (handlers take `*echo.Context`)
- **Frontend**: React 19, TanStack Query & Router, TypeScript, Vite, Tailwind CSS v4, Chart.js, date-fns, lucide-react
- **Database**: SQLite only (modernc.org/sqlite), Goose migrations. `queries.go` still has PostgreSQL-dialect variants;
  they are not wired up: there is no driver and no configuration option
- **Auth**: JWT + bcrypt, token in sessionStorage, email activation; `X-API-Key` auth for `/api/v1`
- **I18n**: react-i18next (frontend) + own translator (backend) with embedded locale files; `en_US`, `uk_UA`, `ru_UA`, `bg_BG`
- **Logging**: slog, created in `main.go` and passed through dependency injection; JSON format in production
- **AI**: multi-provider service (OpenAI now, providers are activated by their API keys)
- Versions: `go.mod`, `web/package.json`. TypeScript stays on 6.0: typescript-eslint supports TypeScript < 6.1 only

## Directory layout
```
cmd/
  web/                  # main web server
  migrate/              # migration runner CLI
  seed/                 # database seeding (seed files in sql/)
internal/
  auth/                 # JWT token management
  config/               # configuration from the environment
  database/             # SQLite connection, migration helpers
  handlers/<domain>/    # HTTP handlers
  i18n/                 # backend translator with embedded locale files
  logger/               # slog setup
  middleware/           # auth, API key, logger, validator
  models/               # user, calorie_entry, ingredient, weight_history, activation_token, api_key
  repositories/         # data access, one file per entity
  server/               # wiring, routes, security (CSP, rate limits), end-to-end tests
  services/<domain>/    # business logic
migrations/             # Goose SQL migrations
scripts/                # utility scripts (create_user.go, install-local-db.sh)
skills/                 # product feature: skills for external agents that log food through /api/v1
  kkal-tracker-food-log/      # production skill (SKILL.md + scripts/kkal_client.py)
  kkal-tracker-food-log-dev/  # generated copy for the dev server, do not edit by hand
  sync_dev_skill.py           # regenerates the dev skill from the production one
web/
  src/api/              # API service classes, one per backend domain
  src/components/       # components (flat + ai/, reports/, settings/)
  src/hooks/ src/types/ src/utils/ src/styles/
  src/i18n/             # i18next config + locales/
  src/pages/            # Dashboard, FoodList, Profile, Report, Settings, AIInsights, ...
  public/               # favicon, icons
  dist/                 # build output (generated, git-ignored), embedded by web/embed.go
kubernetes/             # base/, overlays/{dev,prod}/, backup/ (backup image)
docs/                   # reference docs
data/                   # local SQLite database (git-ignored)
tmp/                    # Air output, logs (git-ignored)
```
Domains (handlers and services): ai, apidata (external `/api/v1`: data export, ingredient list, food logging),
apikey, auth, calorie(s), email, export, ingredient(s), languages, metrics, profile (incl. weight goals), reports,
static (embedded frontend), weight.

## Conventions
- **Handlers**: `handlers/<domain>/handler.go`, request/response types in `dto.go`
- **Services**: `services/<domain>/service.go`, DTOs in `types.go`, domain errors in `errors.go`
- **Repositories**: one file per entity plus shared `interfaces.go` (interfaces for testability), `queries.go` (all
  SQL, by dialect), `errors.go`, `SqlLoader.go`
- **Error handling**: domain errors in services, repository errors in `repositories/errors.go`; handlers map them to
  status codes and never pass internal error text to the client
- **Frontend API**: one `.ts` file per backend domain (`calories.ts`, `weight.ts`, `ai.ts`, ...)
- **Locale files**: JSON named by locale code, in `web/src/i18n/locales/` and `internal/i18n/locales/`
