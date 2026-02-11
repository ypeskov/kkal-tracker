# Claude Code Context

## Project Overview
A Go/Echo web application for calorie tracking with React/TanStack frontend. Built using `/Users/ypeskov/Projects/Go/qr-generator` as structural reference. Features calorie tracking, weight management, ingredient database, and reporting capabilities.

## Architecture
- **Backend**: Go 1.25 with Echo framework v4.15.0
- **Frontend**: React 19 with TanStack Query & Router, TypeScript, Vite
- **Database**: Dual support - SQLite (default) with PostgreSQL option, Goose migrations v3.26.0
- **Auth**: JWT (v5.3.1) with bcrypt password hashing, sessionStorage persistence, email activation
- **I18n**: react-i18next (frontend) + custom translator (backend) with en_US, uk_UA, ru_UA, bg_BG locales
- **Logging**: Structured logging with slog
- **Dev Tools**: Air for live reload
- **Deployment**: Kubernetes-ready with backup integration

## Key Commands

### Development
```bash
make dev              # Start development with Air live reload
make watch            # Alternative to 'make dev' with Air auto-install prompt
make build            # Build both frontend and backend
make build-frontend   # Build only the React frontend
make run              # Build and run the application
make clean            # Clean build artifacts
make init             # Install dependencies and initial setup
make install-deps     # Install Go and Node dependencies
```

### Database Migrations
```bash
make migrate-up           # Run all pending migrations
make migrate-down         # Rollback last migration
make migrate-status       # Check migration status
make migrate-create NAME=migration_name  # Create new migration
```

### Database Seeding
```bash
make seed         # Seed database with initial data
make seed-clean   # Clean and re-seed database (dev only)
```

### Testing & Quality
```bash
make test         # Run all Go tests
go test ./...     # Run Go tests (alternative)
go vet ./...      # Go static analysis
go mod tidy       # Clean up dependencies
```

## Project Structure

### Directory Layout
```
cmd/                    # Application entry points
├── web/                # Main web server
├── migrate/            # Migration runner CLI
└── seed/               # Database seeding (with sql/ subdirectory for seed files)

internal/               # Core application code (clean architecture)
├── auth/               # JWT token management
├── config/             # Configuration loading from env
├── database/           # Database connection & dialect support (SQLite/PostgreSQL)
├── handlers/           # HTTP handlers — one subdirectory per domain
│   ├── ai/             # AI analysis endpoints (handler.go, dto.go)
│   ├── auth/           # Authentication (login, register, activate)
│   ├── calories/       # Calorie entry CRUD
│   ├── export/         # Data export (Excel download/email)
│   ├── ingredients/    # Ingredient management
│   ├── languages/      # Supported languages endpoint
│   ├── metrics/        # Health metrics (BMI, BMR, TDEE)
│   ├── profile/        # User profile & weight goals
│   ├── reports/        # Analytics & reporting
│   ├── static/         # Static file serving (embedded frontend)
│   └── weight/         # Weight history tracking
├── i18n/               # Backend translator with embedded locale files
├── logger/             # Structured logging setup (slog wrapper)
├── middleware/          # Echo middleware (auth, logger, validator)
├── models/             # Data models (user, calorie_entry, ingredient, weight_history, activation_token)
├── repositories/       # Data access layer — repository pattern with interfaces
│                       # Key files: interfaces.go, queries.go, errors.go, SqlLoader.go
└── services/           # Business logic — one subdirectory per domain
    ├── ai/             # Multi-provider AI integration (OpenAI, extensible)
    ├── auth/           # Authentication & password hashing
    ├── calorie/        # Calorie entry logic
    ├── email/          # Email sending (activation, export delivery)
    ├── export/         # Excel generation for weight/food data
    ├── ingredient/     # Ingredient management with translations
    ├── metrics/        # Health metrics calculation (BMI, BMR, TDEE)
    ├── profile/        # User profile & preferences
    ├── reports/        # Analytics & data aggregation
    └── weight/         # Weight tracking

kubernetes/             # K8s deployment configs
├── base/               # Base resources (deployment, service, ingress, PV/PVC, backup CronJob)
└── overlays/           # Environment-specific overlays (dev/, prod/)

migrations/             # Goose SQL migration files (run with `make migrate-*`)

scripts/                # Utility scripts (e.g., create_user.go)

web/                    # React frontend (Vite + TypeScript)
├── src/
│   ├── api/            # API service classes (one per domain)
│   ├── components/     # React components (flat + subdirs: ai/, reports/, settings/)
│   ├── hooks/          # Custom React hooks
│   ├── i18n/           # i18next config + locales/ (en_US, uk_UA, ru_UA, bg_BG)
│   ├── pages/          # Page components (Dashboard, FoodList, Profile, Report, Settings, AIInsights, etc.)
│   ├── styles/         # Global CSS styles
│   ├── types/          # TypeScript type definitions
│   └── utils/          # Utility functions (calculator, numericInput)
├── public/             # Static assets (favicon, icons)
├── dist/               # Built frontend assets (generated, git-ignored)
└── embed.go            # Go embedded filesystem for serving dist/

bin/                    # Compiled binaries (generated)
data/                   # SQLite database files (app.db)
tmp/                    # Air temporary files
```

### Conventions
- **Handlers**: Each domain has its own directory with `handler.go` and optionally `dto.go` for request/response types
- **Services**: Each domain has `service.go`, optionally `types.go` for DTOs and `errors.go` for domain errors
- **Repositories**: One file per entity, plus shared `interfaces.go`, `queries.go`, `errors.go`
- **Frontend API**: One `.ts` file per backend domain (e.g., `calories.ts`, `weight.ts`, `ai.ts`)
- **Locale files**: JSON, named by locale code (e.g., `bg_BG.json`), both in `web/src/i18n/locales/` and `internal/i18n/locales/`

## Environment Variables
```
# Server Configuration
PORT=8080
JWT_SECRET=your-jwt-secret-key-change-this-in-production
LOG_LEVEL=debug

# Database Configuration
DATABASE_TYPE=sqlite                    # Options: sqlite, postgres
DATABASE_PATH=./data/app.db            # For SQLite
# POSTGRES_URL=postgres://user:password@localhost/kkal_tracker?sslmode=disable  # For PostgreSQL

# AI Configuration (providers are activated when API keys are set)
OPENAI_API_KEY=sk-...                  # OpenAI API key
OPENAI_BASE_URL=                        # Optional: custom base URL for proxy/Azure
# ANTHROPIC_API_KEY=sk-ant-...         # Anthropic API key (future)
# ANTHROPIC_BASE_URL=                   # Optional: custom base URL
# OLLAMA_BASE_URL=http://localhost:11434  # Ollama local server URL

# Google Drive Backup (Optional)
# Use rclone to generate OAuth2 token: https://rclone.org/drive/
GDRIVE_OAUTH_TOKEN={"access_token":"...","token_type":"Bearer","refresh_token":"..."}
GDRIVE_FOLDER_PATH=/services/kkal-tracker/backups
```

## Important Notes
- **No auto-migrations**: Migrations must be run manually for production safety
- **Embedded assets**: Frontend dist files are embedded in Go binary via `web/embed.go`
- **API routes**: All API endpoints under `/api` prefix
- **Authentication**: JWT tokens stored in sessionStorage, persist across page reloads
- **Internationalization**: Support for English (en_US), Ukrainian (uk_UA), Russian in Ukraine (ru_UA), and Bulgarian (bg_BG)
- **Development**: Uses Air for live reload, excludes `web/dist` from watching
- **Production**: JSON logging format, structured error handling
- **Logging**: NEVER import `log/slog` directly - use the centralized logger passed from main.go through dependency injection
- **Comments**: ALWAYS write comments in English only. Never use Russian, Ukrainian, or any other language for code comments. This ensures international accessibility and consistency across the codebase
- **Server Management**:
  # ⚠️  CRITICAL WARNING - DO NOT START DEVELOPMENT SERVER ⚠️
  # NEVER RUN `make dev` OR ANY SERVER COMMANDS WITHOUT EXPLICIT USER APPROVAL
  # THE USER MANAGES THE DEVELOPMENT SERVER (Air) IN A SEPARATE TERMINAL SESSION
  # ONLY BUILD COMMANDS (`make build`, `make build-frontend`) ARE ALLOWED

## Frontend Style Standards

The frontend follows strict Tailwind CSS utility-first approach with standardized layout patterns. **ALWAYS** adhere to these guidelines when creating or modifying pages:

### Page Layout Standards

#### Standard Content Pages (Dashboard, Food List, Profile, etc.)
Use **constrained width** layout for better readability and consistency:

```tsx
<div className="max-w-screen-xl mx-auto px-4 py-2 md:px-6 lg:px-8">
  {/* Page content */}
</div>
```

- **Width**: `max-w-screen-xl` (1280px maximum)
- **Centering**: `mx-auto` (horizontally centered)
- **Padding**: Responsive padding that increases on larger screens
  - Mobile: `px-4 py-2`
  - Medium: `md:px-6`
  - Large: `lg:px-8`

#### Data Visualization Pages (Reports, Charts, Dashboards with graphs)
Use **full-width** layout to maximize chart visibility:

```tsx
<div className="px-4 py-2 md:px-6 lg:px-8">
  {/* Page content with charts */}
</div>
```

- **No max-width**: Allows content to use full viewport width
- **Same padding**: Maintains consistent spacing with other pages

### Card/Section Styling

All card components should use the **Dashboard card pattern** for visual consistency:

```tsx
<div className="bg-white rounded-lg shadow-md p-4">
  {/* Card content */}
</div>
```

- **Background**: `bg-white`
- **Border radius**: `rounded-lg` (8px)
- **Shadow**: `shadow-md` (medium shadow for depth)
- **Padding**: `p-4` or custom padding like `p-lg`, `pl-lg`, `pr-lg` if defined

### Page Header Styling

For page titles and headers:

```tsx
<div className="mb-6">
  <h2 className="text-3xl font-semibold text-gray-800">{t('page.title')}</h2>
</div>
```

- **Title size**: `text-3xl` (1.875rem)
- **Weight**: `font-semibold`
- **Color**: `text-gray-800`
- **Bottom margin**: `mb-6` for spacing

### ⚠️ Critical Rules

1. **NEVER use undefined CSS classes**: Do not use semantic classes like `.page`, `.card`, `.page-header`, etc. These were removed during Tailwind migration
2. **ONLY use Tailwind utilities**: All styling must use Tailwind CSS utility classes
3. **Follow existing patterns**: Reference `DashboardPage.tsx`, `FoodList.tsx`, `Profile.tsx`, and `Report.tsx` for layout patterns
4. **Responsive design**: Always include responsive breakpoints (`md:`, `lg:`) for padding and layout
5. **Consistency over customization**: Prefer standardized patterns over unique layouts for each page

### Form Input Styling

Standard input field styling:

```tsx
<input
  type="text"
  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-500"
/>
```

Numeric input with integrated calculator:

```tsx
import CalculatorInput from '@/components/CalculatorInput';

<CalculatorInput
  id="weight"
  value={weight}
  onChange={setWeight}
  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
  required
/>
```

**Calculator Features:**
- Replaces standard numeric inputs for weight, calories, fats, carbs, proteins
- Calculator icon button appears on the right side of input
- Clicking icon shows/hides calculator keypad below input
- Input field acts as calculator display (no separate display)
- Supports keyboard input and calculator buttons simultaneously
- Evaluates expressions with `=` button or Enter key
- Special buttons: `C` (clear), `←` (backspace)
- Focus remains on input when using calculator buttons
- Used in: AddFoodEntryForm, EditEntryModal, AddIngredientModal, EditIngredientModal

### Button Styling

Primary action buttons:

```tsx
<button className="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors">
  {t('common.save')}
</button>
```

Secondary/outline buttons:

```tsx
<button className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors">
  {t('common.cancel')}
</button>
```

### Migration from Old CSS

If you encounter old semantic CSS classes during refactoring:

| Old Class | Replacement |
|-----------|-------------|
| `.page` | `max-w-screen-xl mx-auto px-4 py-2 md:px-6 lg:px-8` (or full-width variant) |
| `.card` | `bg-white rounded-lg shadow-md` |
| `.page-header` | `mb-6` |
| `.page-title` | `text-3xl font-semibold text-gray-800` |

## Dependencies

### Backend
- Echo v4.15.0 (web framework)
- modernc.org/sqlite v1.44.3 (SQLite driver)
- Goose v3.26.0 (migrations)
- JWT-Go v5.3.1 (authentication)
- golang.org/x/crypto v0.47.0 (bcrypt)
- go-playground/validator v10.30.1 (validation)
- sashabaranov/go-openai v1.41.2 (AI integration)
- xuri/excelize v2.10.0 (Excel export)
- Air (live reload - auto-installed via make watch)

### Frontend
- React 19.2.4 + React DOM
- TanStack Query v5.90.20 (server state)
- TanStack Router v1.157.18 (routing)
- Chart.js v4.5.1 + react-chartjs-2 v5.3.1 (charts)
- date-fns v4.1.0 (date utilities)
- react-i18next v16.5.4 (internationalization)
- lucide-react v0.563.0 (icons)
- Tailwind CSS v4.1.18 (styling)
- Vite v7.3.1 (build tool)
- TypeScript v5.9.3

## Build Process
1. Frontend builds to `web/dist/` (via Vite)
2. Go embeds `web/dist` files via `web/embed.go` using `//go:embed`
3. Single binary (`main`) serves both API and static files
4. Air watches Go, TS, TSX, JS, HTML, CSS files and triggers rebuilds
5. Air excludes `web/dist`, `web/node_modules`, `tmp`, `bin`, `data` from watching
6. Build errors logged to `tmp/build-errors.log`
7. Docker images versioned via `version.txt` (current: v5.2.1)
8. **IMPORTANT**: `build-and-push.sh` always builds without cache (`--no-cache`) and removes `web/dist/` before building

## Features
- **User Authentication**: JWT-based login/register/activate with bcrypt password hashing and email activation
- **Calorie Tracking**: Add food entries with name, weight, kcal/100g, auto-calculated totals
- **Weight Management**: Track weight history over time with charts and weight goals
- **Weight Goals**: Set target weight with target date, track progress with visual indicators
- **Health Metrics**: BMI, BMR, TDEE calculations based on user profile (age, gender, activity level)
- **Ingredient Database**: Global ingredients with multilingual names and nutritional data
- **User Profiles**: Personal settings, preferences, gender, activity level management
- **Reports & Analytics**: Data visualization and calorie/weight trends
- **AI Insights**: AI-powered nutrition and weight analysis with personalized recommendations
  - OpenAI integration (GPT-4o-mini by default)
  - Provider selection in UI (extensible architecture for multiple providers)
  - Customizable analysis periods (7, 14, 30, 90 days)
  - Optional specific questions for targeted advice
  - Multilingual responses based on user's language preference
  - Rate limited: 2 requests per minute
- **Data Export**: Export weight and food data as Excel files (download or email delivery)
- **Email Service**: Activation emails and export delivery
- **Dashboard**: View today's entries with total calorie count
- **Internationalization**: Full i18n support (en_US, uk_UA, ru_UA, bg_BG) with language switcher, both frontend and backend
- **Persistent Sessions**: JWT tokens stored in sessionStorage
- **Responsive UI**: Clean, modern interface with proper form validation
- **Database Backup**: Automated backups to Google Drive via OAuth2
- **Integrated Calculator**: Built-in calculator for all numeric input fields (weight, calories, macros)
  - Calculator icon button on all numeric inputs
  - Input field serves as calculator display
  - Support for mathematical expressions (+, -, *, /, parentheses)
  - Real-time expression evaluation
  - Keyboard and button input modes
  - Auto-evaluation on Enter key or blur
  - Collapsible calculator panel with numeric keypad

## Service Architecture

The application follows a clean architecture pattern with separation of concerns:

### Domain Services (`internal/services/`)
- **ai**: Multi-provider AI integration (OpenAI, extensible for others)
- **auth**: Authentication, JWT management, password hashing
- **calorie**: Calorie entry business logic and calculations
- **email**: Email sending (activation, export delivery)
- **export**: Excel generation for weight/food data export
- **ingredient**: Global ingredient management with translations
- **metrics**: Health metrics calculation (BMI, BMR, TDEE)
- **profile**: User profile, preferences, and weight goals
- **reports**: Analytics, trends, and data aggregation
- **weight**: Weight tracking and history management

### Data Layer (`internal/repositories/`)
- Repository pattern for data access
- Interface-based design for testability
- Separate repositories for each domain entity (user, calorie_entry, ingredient, weight_history, activation_token)
- Centralized query management in `queries.go`
- Custom error types in `errors.go` for consistent error handling
- SQL file loader utility in `SqlLoader.go` for loading SQL from files

### API Endpoints
All API routes are prefixed with `/api`:
- `GET /api/languages` - Supported languages (public, no auth)
- `/api/auth/*` - Authentication (rate limited: 5 req/sec)
  - `POST /api/auth/login` - User login
  - `POST /api/auth/register` - User registration
  - `GET /api/auth/activate/:token` - Activate account via email link
  - `GET /api/auth/me` - Get current user (requires auth)
- `/api/calories/*` - Calorie entry CRUD (GET, POST, PUT /:id, DELETE /:id)
- `/api/ingredients/*` - Ingredient management (GET, GET /:id, POST, PUT /:id, DELETE /:id)
- `/api/weight/*` - Weight history (GET, POST, PUT /:id, DELETE /:id)
- `/api/profile` - User profile (GET, PUT)
- `/api/profile/goal` - Weight goal (GET progress, PUT set, DELETE clear)
- `/api/metrics` - Health metrics (GET)
- `/api/reports/data` - Aggregated report data (GET, with from/to query params)
- `/api/ai/*` - AI analysis (rate limited: 2 req/min)
  - `GET /api/ai/status` - AI service status
  - `POST /api/ai/analyze` - Perform AI analysis
- `/api/export` - Data export (POST, weight/food as Excel or email)

## Deployment

### Docker
- **Dockerfile**: Multi-stage build for optimized image size
- **Build script**: `build-and-push.sh` - Automated Docker build and push
  - **Always builds without cache** (`--no-cache`) to ensure fresh builds
  - **Automatically cleans `web/dist/`** before build
  - Supports custom tags and platform selection (e.g., `linux/amd64`)
  - Optional push to Docker registry
  - Usage: `./build-and-push.sh [--platform=PLATFORM] [push] [TAG]`
- **Image**: `ypeskov/kcal-tracker` (versioned with `version.txt`)

### Kubernetes
Full Kubernetes deployment configuration in `kubernetes/` directory:
- **Base configurations**: `kubernetes/base/` - Common resources
  - Deployment, Service, Ingress
  - Persistent Volume & Persistent Volume Claim
  - ConfigMap for backup configuration
  - CronJob for automated backups
- **Environment overlays**: `kubernetes/overlays/{dev,prod}/` - Environment-specific configs
- Includes CronJob for automated database backups to Google Drive

### Database Management
- **Migrations**: Manual control for production safety
- **Dual database support**: SQLite for development, PostgreSQL for production
- **Seeding**: `cmd/seed/` for initial data (ingredients database)
- **Backup strategy**: Automated backups to Google Drive

## Error Handling Patterns
- Service-level errors in `internal/services/*/errors.go`
- Repository errors in `internal/repositories/errors.go`
- Structured error responses with proper HTTP status codes
- Consistent error format across all endpoints
- **CRITICAL: NEVER expose internal error details to clients.** In HTTP handlers, always return generic error messages in responses. Log the actual `err` for debugging, but send only a safe message to the client:
  ```go
  // CORRECT:
  h.logger.Error("Export failed", "error", err, "user_id", userID)
  return echo.NewHTTPError(http.StatusInternalServerError, "Export failed")

  // WRONG — security vulnerability, leaks internal details:
  return echo.NewHTTPError(http.StatusInternalServerError, "Export failed: "+err.Error())
  ```
  This applies to all `StatusInternalServerError` responses. Validation errors (`StatusBadRequest`) from `c.Validate()` are acceptable to pass through since they only contain field names and validation rules.

## Production Build & Deployment

### Prerequisites
- Code tested locally (`make dev`)
- All changes committed to develop branch
- Docker authenticated for push to Docker Hub

### Build Steps

#### 1. Commit Code Changes
```bash
git add .
git commit -m "description of changes"
git push origin develop
```

#### 2. Build and Push Docker Image
Use the `build-and-push.sh` script:
```bash
./build-and-push.sh VERSION --push
```

Example:
```bash
./build-and-push.sh 5.1.1 --push
```

**What the script does:**
- Removes `web/dist/` for a clean build
- Builds image without cache (`--no-cache`)
- Tags as both `VERSION` and `latest`
- Pushes to Docker Hub (`ypeskov/kcal-tracker:VERSION`)
- Writes version to `version.txt`

**IMPORTANT:** Do NOT specify `--platform` flag, so the image builds for the host machine architecture (arm64 for Apple Silicon servers).

#### 3. Commit Version Update
```bash
git add version.txt
git commit -m "docker vVERSION"
git push origin develop
```

#### 4. Update Kubernetes Deployment
Edit `kubernetes/base/deployment.yaml`:
```yaml
image: ypeskov/kcal-tracker:VERSION
```

Commit:
```bash
git add kubernetes/base/deployment.yaml
git commit -m "update deployment to vVERSION"
git push origin develop
```

#### 5. Deploy to Server
On the Kubernetes server:
```bash
# Option 1: Apply manifests
kubectl apply -k kubernetes/overlays/prod/

# Option 2: Update image directly
kubectl set image deployment/kkal-tracker kkal-tracker=ypeskov/kcal-tracker:VERSION

# Option 3: Restart deployment (if tag didn't change)
kubectl rollout restart deployment kkal-tracker
```

#### 6. Verify Deployment
```bash
kubectl get pods
kubectl logs POD_NAME
kubectl logs POD_NAME --previous  # logs from previous run if crashed
```

### Common Errors
| Error | Cause | Solution |
|-------|-------|----------|
| `exec format error` | Image built for wrong architecture | Rebuild without `--platform` flag |
| `CrashLoopBackOff` | Application error | Check `kubectl logs --previous` |
| Pod doesn't update | Kubernetes caches image with same tag | Use a new version tag |