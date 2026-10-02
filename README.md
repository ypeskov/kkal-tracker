# Kkal-Tracker

A calorie tracking and weight management web application: Go/Echo backend, React/TanStack frontend, SQLite database.

The application compiles to a single binary with the frontend embedded. It needs no external services to run: the database is one SQLite file. SMTP (activation emails, export by email) and OpenAI (AI insights) are optional integrations.

Current version: see [`version.txt`](version.txt).

## Features

- **Food diary**: entries with name, weight and kcal per 100 g; calories and macros are calculated automatically
- **Ingredient list**: every user has a personal ingredient list, copied on registration from the global one in the user's language
- **Weight tracking**: weight history with charts, a weight goal with a target date and progress
- **Health metrics**: BMI, BMR and TDEE from the profile (age, height, gender, activity level)
- **Reports**: calorie and weight trends for any period (Chart.js)
- **AI insights**: nutrition and weight analysis by an OpenAI model for the last 7, 14, 30 or 90 days, with an optional question, answered in the user's language
- **Data export**: weight and food data as an Excel file, downloaded or sent by email
- **API keys and external API**: programmatic access to the user's data, including food logging by AI agents (see [External API](#external-api-apiv1))
- **Accounts**: registration with email activation, JWT authentication
- **Internationalization**: English, Ukrainian, Russian and Bulgarian (`en_US`, `uk_UA`, `ru_UA`, `bg_BG`), both in the UI and in emails
- **Built-in calculator** in every numeric field (weight, calories, macros): `150+80` is evaluated in place
- **Backups**: a Kubernetes CronJob copies the database to Google Drive

## Tech Stack

| | |
|---|---|
| Backend | Go 1.26+, Echo v5, `modernc.org/sqlite` (pure Go, no CGO), Goose migrations, `golang-jwt` v5, bcrypt, `go-playground/validator`, `go-openai`, `excelize` |
| Frontend | React 19, TypeScript, Vite, TanStack Query and Router, Tailwind CSS v4, Chart.js, i18next, lucide-react |
| Tooling | Air (live reload), ESLint, Docker (distroless image), Kubernetes with Kustomize |

Exact versions are in [`go.mod`](go.mod) and [`web/package.json`](web/package.json).

## Getting Started

### Prerequisites

- Go 1.26+
- Node.js 20.19+ or 22.12+ (required by Vite; Docker builds use Node 24) and npm
- Optional: the `sqlite3` CLI (used by `make seed-clean`)

### First run

All commands are run from the repository root.

```bash
# 1. Configuration
cp .env.sample .env

# 2. Go and npm dependencies
make install-deps

# 3. Database schema
make migrate-up

# 4. Global ingredient list (optional, but do it before creating users)
make seed

# 5. A user that can log in right away
go run scripts/create_user.go user@example.com password123          # English ingredients
go run scripts/create_user.go user@example.com password123 ru_UA    # en_US, uk_UA or ru_UA

# 6. Build the frontend and the binary, then start the server
make run
```

The application is available at <http://localhost:8080>.

Notes:

- Keep `ENVIRONMENT=development` in `.env` for local work. Without it the application runs as production and refuses to start with a placeholder or short `JWT_SECRET`. The same applies to `make migrate-*`, `make seed` and `scripts/create_user.go`.
- A user receives a copy of the global ingredients when the account is created, so seed the database first. The seed has ingredient names in `en_US`, `uk_UA` and `ru_UA`.
- `scripts/create_user.go` creates an active user without sending an email. Registration through the UI sends an activation link (valid for 24 hours), so it needs working SMTP settings.
- `make seed` also inserts a sample account `example@example.com` with a month of diary entries. It is created inactive and cannot log in as is.

### Development

```bash
make watch    # Air: rebuilds the frontend and the backend on every change, restarts the server
make dev      # go run without rebuilding the frontend (web/dist must already exist)
```

Air watches `.go`, `.ts`, `.tsx`, `.js`, `.html` and `.css` files and writes build errors to `tmp/build-errors.log`. `make watch` offers to install Air if it is missing.

### Make commands

```bash
make build            # Build the frontend and the backend (binary: ./main)
make build-frontend   # Build only the frontend into web/dist
make run              # Build and run
make watch            # Live reload with Air
make dev              # go run cmd/web/main.go
make clean            # Remove build artifacts, web/dist, web/node_modules and tmp
make install-deps     # go mod tidy and npm install
make test             # go test -v ./...

make migrate-up                      # Apply all pending migrations
make migrate-down                    # Roll back the last migration
make migrate-status                  # Show migration status
make migrate-create NAME=add_table   # Create a new migration file

make seed             # Run the SQL files from cmd/seed/sql in alphabetical order
make seed-clean       # Delete the global ingredients and seed again (dev only, needs sqlite3)
```

### Tests and checks

```bash
make build-frontend      # Go tests need web/dist: the frontend is embedded into the binary
go test ./...
go vet ./...
cd web && npm run lint
```

The Go tests include an end-to-end smoke test (`internal/server/smoke_test.go`). It starts the fully wired server on a temporary migrated database with a fake SMTP server and walks through registration, the activation email, login, calories, weight and export.

## Configuration

The application reads its settings from the environment. A `.env` file in the working directory is loaded when it exists; variables that are already set take precedence.

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `production` | `development` or `production`. Unset means production: JSON logs and a strict `JWT_SECRET` check |
| `JWT_SECRET` | none | Signs the JWT tokens. Outside `development` it must be at least 32 characters and not a known placeholder, e.g. `openssl rand -hex 32` |
| `PORT` | `8080` | HTTP port |
| `DATABASE_PATH` | `./data/kkal_tracker.db` | SQLite database file (`.env.sample` sets `./data/app.db`) |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `APP_URL` | `http://localhost:8080` | Public URL of the application: base of the activation links and the allowed CORS origin |
| `TRUSTED_PROXIES` | loopback, link-local and private networks | Comma-separated CIDR ranges whose `X-Forwarded-For` header is trusted. The client IP is the key of the rate limiters |
| `SMTP_HOST` | `smtp.gmail.com` | SMTP server for activation emails and export by email |
| `SMTP_PORT` | `587` | |
| `SMTP_USER` | none | |
| `SMTP_PASSWORD` | none | For Gmail: an App Password |
| `SMTP_FROM` | `noreply@kkal-tracker.com` | Sender address |
| `OPENAI_API_KEY` | none | Enables AI insights |
| `OPENAI_BASE_URL` | none | Custom base URL (proxy) |
| `OPENAI_MODEL` | `gpt-5.2` | Model used for the analysis |
| `AI_USE_MAX_TOKENS` | `false` | Limit the length of the AI answer |
| `AI_MAX_TOKENS` | `2000` | The limit, when `AI_USE_MAX_TOKENS` is `true` |

`GDRIVE_OAUTH_TOKEN`, `GDRIVE_FOLDER_PATH` and `GDRIVE_CLEANUP` are read only by the Kubernetes backup job, see [Backups](#backups).

## Database

SQLite is the only supported database. The schema is managed by [Goose](https://github.com/pressly/goose) migrations in `migrations/`.

**Migrations are never applied automatically.** The server does not touch the schema on startup. Run `make migrate-up` before the first start and after pulling changes that add migrations.

The `make migrate-*` targets run the migration CLI `cmd/migrate` with `go run`:

| Command | What it does |
|---------|--------------|
| `make migrate-up` | Applies all pending migrations |
| `make migrate-down` | Rolls back the last applied migration (one per call) |
| `make migrate-status` | Lists the migrations as applied or pending |
| `make migrate-create NAME=add_table` | Creates an empty timestamped SQL file in `migrations/` with the `goose` CLI (installed on first use) |

How the CLI works:

- It loads the same configuration as the server (`.env`, then the environment) and migrates the database at `DATABASE_PATH`. The `JWT_SECRET` check applies here too, so either set `ENVIRONMENT=development` or provide a real secret.
- It reads the migration files from the `migrations/` directory relative to the current directory, so run it from the repository root.
- Applied versions are recorded by Goose in the `goose_db_version` table of the same database.
- `scripts/create_user.go` applies all pending migrations as well before creating the user.

For another database file set `DATABASE_PATH`:

```bash
DATABASE_PATH=/path/to/app.db make migrate-up
```

In production the migrations are applied the same way: from a checkout of the repository with Go installed, with `DATABASE_PATH` pointing to the production database file. The Docker image has no migration runner and `deploy.sh` does not run migrations. In the Kubernetes setup the database is on a `hostPath` volume, so the file is reachable on the server itself.

## API

All endpoints are under `/api`. Unknown `/api` paths answer with 404, they never fall back to the frontend page.

### Application API (JWT)

`POST /api/auth/login` returns a token valid for 24 hours. Protected endpoints take it in the `Authorization: Bearer <token>` header.

| Endpoint | Description |
|----------|-------------|
| `GET /api/languages` | Supported languages (public) |
| `POST /api/auth/login` | Log in |
| `POST /api/auth/register` | Register; sends an activation email |
| `GET /api/auth/activate/:token` | Activate an account |
| `GET /api/auth/me` | Current user |
| `GET, POST /api/calories`, `PUT, DELETE /api/calories/:id` | Food diary entries (`dateFrom`, `dateTo` filters) |
| `GET, POST /api/ingredients`, `GET, PUT, DELETE /api/ingredients/:id` | Ingredient list of the user |
| `GET, POST /api/weight`, `PUT, DELETE /api/weight/:id` | Weight history (`from`, `to` filters) |
| `GET, PUT /api/profile` | Profile |
| `GET, PUT, DELETE /api/profile/goal` | Weight goal: progress, set, clear |
| `GET /api/metrics` | BMI, BMR, TDEE |
| `GET /api/reports/data?from=&to=` | Aggregated data for reports |
| `GET /api/ai/status`, `POST /api/ai/analyze` | AI insights |
| `POST /api/export` | Export to Excel: download or email |
| `GET, POST /api/api-keys`, `POST /api/api-keys/:id/revoke`, `DELETE /api/api-keys/:id` | API key management |

Rate limits per IP: `/api/auth/*` 5 requests per second; registration 3 attempts, then one per 20 minutes; `/api/ai/*` 2 requests per minute.

### External API (`/api/v1`)

For scripts and AI agents. Authentication is an API key in the `X-API-Key` header. Keys are created in the UI (Settings → API Keys), can be time-limited or permanent, and are shown in full only once; the server stores a SHA-256 hash. A key gives full read and write access to the user's data, there are no scopes.

Rate limit per IP: a burst of 20 requests, then 1 request per second.

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/data?type=weight\|food\|both&from=YYYY-MM-DD&to=YYYY-MM-DD` | Weight entries, food entries or both |
| `GET /api/v1/ingredients` | All ingredients of the user with `times_used` and `last_used` |
| `POST /api/v1/food` | Store a meal |
| `PUT /api/v1/food/:id` | Change a food entry |
| `DELETE /api/v1/food/:id` | Delete a food entry |

```bash
curl -H "X-API-Key: $KEY" "http://localhost:8080/api/v1/data?type=both&from=2026-10-01&to=2026-10-02"
```

A meal is a list of items. Each item refers to an existing ingredient by `ingredient_id` or describes a new one in `new_ingredient`:

```json
{
  "meal_datetime": "2026-10-02T13:30:00+03:00",
  "items": [
    {"ingredient_id": 75, "weight": 150},
    {"new_ingredient": {"name": "Hummus", "kcal_per_100g": 240, "fats": 17, "carbs": 12, "proteins": 8}, "weight": 60},
    {"new_ingredient": {"name": "Restaurant pasta", "kcal_per_100g": 180}, "weight": 350, "one_off": true}
  ]
}
```

- The server never matches or creates ingredients by name on its own: the client reads the ingredient list and decides.
- A new ingredient whose name equals an existing one (ignoring case, spacing and `ё`/`е`) is rejected with 409, the response contains the existing ingredient.
- `one_off: true` writes the food to the diary without adding it to the ingredient list.
- `meal_datetime` is optional and takes RFC 3339 with a time zone offset.
- A meal is atomic: all items are stored or none. Calories are calculated on the server.
- `PUT /api/v1/food/:id` takes `weight`, `ingredient_id`, `kcal_per_100g`, `meal_datetime`; fields that are not sent keep their values, calories are recalculated.

### Agent skills

`skills/` contains ready-made skills that let an AI agent log food through the external API, for example from voice dictation.

- `kkal-tracker-food-log/` is the production skill: `SKILL.md` and the client `scripts/kkal_client.py` (Python standard library only).
- `kkal-tracker-food-log-dev/` is the same skill pointed at the development server. It is generated, do not edit it by hand: change the production skill and run `python3 skills/sync_dev_skill.py`.
- The client takes the key from the `KKAL_API_KEY` variable or from the git-ignored `api_key` file in the skill folder; `KKAL_BASE_URL` overrides the server address. The scripts in the repository hold only the `XXXXXXXXX` placeholder.

## Project Structure

```
cmd/
├── web/           # Web server
├── migrate/       # Migration runner
└── seed/          # Database seeding, SQL files in sql/
internal/
├── auth/          # JWT tokens
├── config/        # Configuration from the environment
├── database/      # SQLite connection and Goose calls
├── handlers/      # HTTP handlers, one directory per domain
├── i18n/          # Backend translations
├── logger/        # slog setup
├── middleware/    # JWT auth, API key auth, validator
├── models/        # Data models
├── repositories/  # Data access
├── server/        # Wiring, routes, security settings, smoke tests
└── services/      # Business logic, one directory per domain
kubernetes/        # Kustomize base, dev and prod overlays, backup image
migrations/        # Goose SQL migrations
scripts/           # create_user.go
skills/            # Agent skills for food logging
web/               # React frontend; web/dist is embedded into the binary
```

## Security

- Passwords are hashed with bcrypt; JWT tokens live for 24 hours and are kept in `sessionStorage`.
- Strict Content-Security-Policy (`internal/server/security.go`): scripts only from the own origin, no inline scripts, no `eval`.
- HSTS, `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`; request bodies are limited to 1 MB.
- Rate limiters are in-memory and per instance: the deployment runs a single replica.
- Internal error details are logged, never returned to the client.

## Deployment

### Binary

```bash
make build    # ./main with the frontend embedded
```

Set `JWT_SECRET` (32+ characters), `DATABASE_PATH`, `APP_URL` and the SMTP settings in the environment, apply the migrations and start the binary. Leave `ENVIRONMENT` unset or set it to `production`.

### Docker

`build-and-push.sh` builds the image `ypeskov/kcal-tracker`:

```bash
./build-and-push.sh 5.9.0                           # Build, tag as 5.9.0 and latest
./build-and-push.sh 5.9.0 --push                    # Build and push the version tag to Docker Hub
./build-and-push.sh 5.9.0 --push --platform=linux/amd64
```

The script removes `web/dist`, always builds without cache and writes `v<TAG>` to `version.txt`. Without `--platform` the image is built for the architecture of the build machine.

The image is distroless and runs as a non-root user. It contains the server binary and the migration files, with no `.env` file and no migration runner: all settings come from the environment, and `ENVIRONMENT=production` is preset.

```bash
docker run -p 8080:8080 \
  -v "$PWD/data:/data" \
  -e DATABASE_PATH=/data/app.db \
  -e JWT_SECRET="$(openssl rand -hex 32)" \
  ypeskov/kcal-tracker:5.9.0
```

The database file must already be migrated and writable by the container user (UID 65532).

### Automated deployment

`deploy.sh` runs the whole release from the `develop` branch with a clean working tree:

```bash
./deploy.sh --tag=5.9.0
```

1. Builds and pushes the Docker image
2. Sets the image tag in `kubernetes/base/deployment.yaml`
3. Commits `version.txt` and the manifest as `v<TAG>`, pushes `develop`
4. Merges `develop` into `master` and pushes it
5. Connects to the server over SSH, pulls the repository and runs `kubectl apply -k kubernetes/overlays/prod`

The server address and the repository path on it come from `.deploy.env` (git-ignored, see `.deploy.env.sample`). Options: `--skip-build`, `--skip-k8s`, `--skip-deploy`, `--dry-run`, `--platform=...`; `./deploy.sh --help` describes them.

### Kubernetes

```
kubernetes/
├── base/                     # Deployment, Service, Ingress, PV and PVC, backup CronJob and its script
├── backup/Dockerfile         # Backup image: pinned rclone with sqlite3
└── overlays/
    ├── dev/                  # The base as is
    └── prod/                 # Config and secret generators, Traefik ingress with TLS
```

The production overlay generates the ConfigMap `kkal-tracker-env` from `.env` and the Secret `kkal-tracker-secrets` from `.env.secret` (`JWT_SECRET`, `SMTP_PASSWORD`, `OPENAI_API_KEY`, `GDRIVE_OAUTH_TOKEN`). Both files live only on the server; create them from `.env.sample` and `.env.secret.sample` in `kubernetes/overlays/prod/`.

```bash
kubectl kustomize kubernetes/overlays/prod      # Preview
kubectl apply -k kubernetes/overlays/prod       # Apply
kubectl rollout status deployment/kkal-tracker
kubectl logs -f deployment/kkal-tracker
```

Details of the setup:

- The image tag is set in `kubernetes/base/deployment.yaml` (`deploy.sh` updates it).
- The database is on a `hostPath` PersistentVolume mounted at `/data`, `DATABASE_PATH=/data/app.db`.
- The production ingress uses the cluster-wide Traefik controller and a cert-manager certificate; the host is set in `overlays/prod/ingress-patch.yaml`.
- Set `TRUSTED_PROXIES` to the pod network of the cluster (k3s default: `10.42.0.0/16`), so that rate limiting sees real client addresses.

Rollback:

```bash
kubectl rollout undo deployment/kkal-tracker
```

or set the previous image tag in `deployment.yaml` and apply again.

### Backups

The CronJob `kkal-tracker-backup` runs every day at 02:00 (Europe/Sofia). It makes a consistent copy of the database with `sqlite3 .backup`, compresses it, stores it in `/data/backups` (7 days) and uploads it to Google Drive with rclone.

- `GDRIVE_OAUTH_TOKEN` (in `.env.secret`): OAuth2 token generated with [rclone](https://rclone.org/drive/). Without it the upload is skipped and only the local copy is kept.
- `GDRIVE_FOLDER_PATH` (in `.env`): target folder on Google Drive.
- `GDRIVE_CLEANUP=true` (in `.env`): also delete backups older than 7 days from Google Drive.

The job image `ypeskov/kkal-tracker-backup` is built from `kubernetes/backup/Dockerfile`; its tag follows the rclone version.
