# Claude Code Context

## Project Overview
A Go/Echo web application for calorie tracking with React/TanStack frontend. Built using `/Users/ypeskov/Projects/Go/qr-generator` as structural reference. Features calorie tracking, weight management, ingredient database, and reporting capabilities.

## Architecture
- **Backend**: Go with Echo framework v4.13.4
- **Frontend**: React 18 with TanStack Query & Router, TypeScript, Vite
- **Database**: Dual support - SQLite (default) with PostgreSQL option, Goose migrations v3.25.0
- **Auth**: JWT (v5.3.0) with bcrypt password hashing, sessionStorage persistence
- **I18n**: react-i18next with en_US, uk_UA, ru_UA locales
- **Logging**: Structured logging with slog
- **Dev Tools**: Air for live reload
- **Deployment**: Kubernetes-ready with backup integration

## Key Commands

### Development
```bash
make dev          # Start development with Air live reload
make watch        # Alternative to 'make dev' with Air auto-install prompt
make build        # Build both frontend and backend
make run          # Build and run the application
make clean        # Clean build artifacts
make init         # Install dependencies and initial setup
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
```
├── cmd/
│   ├── web/main.go              # Application entry point
│   ├── migrate/main.go          # Migration runner
│   └── seed/                    # Database seeding
│       ├── main.go
│       └── sql/                 # SQL seed files
├── internal/
│   ├── auth/jwt.go              # JWT token management
│   ├── config/config.go         # Configuration management
│   ├── database/database.go     # Database connection & migrations
│   ├── handlers/                # HTTP handlers by domain
│   │   ├── auth/                # Authentication endpoints
│   │   ├── calories/            # Calorie tracking
│   │   ├── ingredients/         # Ingredient management
│   │   ├── profile/             # User profile
│   │   ├── reports/             # Analytics & reports
│   │   ├── static/              # Static file serving
│   │   └── weight/              # Weight tracking
│   ├── logger/logger.go         # Structured logging setup
│   ├── middleware/              # Custom middleware
│   │   ├── auth.go              # JWT authentication
│   │   ├── logger.go            # Request logging
│   │   └── validator.go         # Input validation
│   ├── models/                  # Data models
│   │   ├── user.go
│   │   ├── calorie_entry.go
│   │   ├── ingredient.go
│   │   └── weight_history.go
│   ├── repositories/            # Data access layer
│   │   ├── interfaces.go        # Repository interfaces
│   │   ├── queries.go           # Centralized SQL queries
│   │   ├── errors.go            # Repository-level errors
│   │   ├── SqlLoader.go         # SQL file loader utility
│   │   ├── user.go
│   │   ├── calorie_entry.go
│   │   ├── ingredient.go
│   │   └── weight_history.go
│   ├── server/server.go         # Echo server setup
│   └── services/                # Business logic by domain
│       ├── auth/                # Authentication service
│       ├── calorie/             # Calorie management
│       ├── ingredient/          # Ingredient service
│       ├── profile/             # Profile management
│       ├── reports/             # Reporting service
│       └── weight/              # Weight tracking
├── kubernetes/                  # K8s deployment configs
│   ├── base/                    # Base configurations
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── ingress.yaml
│   │   ├── pv.yaml
│   │   ├── pvc.yaml
│   │   ├── configmap-backup.yaml
│   │   ├── cronjob-backup.yaml
│   │   └── kustomization.yaml
│   └── overlays/                # Environment overlays
│       ├── dev/
│       └── prod/
├── migrations/                  # Goose SQL migrations
│   ├── 20250907200000_initial_schema.sql
│   ├── 20250907200001_create_indexes.sql
│   ├── 20250907200002_add_ingredients_table.sql
│   ├── 20250907200003_add_unique_constraint_user_ingredients.sql
│   ├── 20250907200004_add_user_profile_fields.sql
│   ├── 20250907200005_add_weight_to_users.sql
│   └── 20250919055016_remove_weight_from_users.sql
├── scripts/                     # Utility scripts
│   └── create_user.go           # User creation utility
├── web/                         # React frontend
│   ├── src/
│   │   ├── api/                 # API service classes
│   │   ├── components/          # React components
│   │   │   ├── Calculator.tsx   # Calculator popup component
│   │   │   ├── CalculatorInput.tsx  # Input with integrated calculator
│   │   │   └── ...              # Other components
│   │   ├── hooks/               # Custom React hooks
│   │   ├── i18n/                # Internationalization
│   │   ├── pages/               # Page components
│   │   ├── styles/              # Global styles and CSS
│   │   ├── utils/               # Utility functions
│   │   │   ├── calculator.ts    # Math expression evaluation utilities
│   │   │   ├── numericInput.ts  # Numeric input validation
│   │   │   └── ...              # Other utilities
│   │   ├── App.tsx              # Root application component
│   │   ├── main.tsx             # Application entry point
│   │   └── router.tsx           # Router configuration
│   ├── public/                  # Static assets
│   │   ├── favicon.ico
│   │   ├── icon.png
│   │   ├── icon.svg
│   │   ├── steak-icon.png
│   │   └── steak-icon.svg
│   ├── dist/                    # Built frontend assets
│   ├── embed.go                 # Go embedded filesystem
│   ├── index.html               # HTML template
│   ├── package.json             # Node dependencies
│   ├── vite.config.ts           # Vite configuration
│   ├── tailwind.config.js       # Tailwind CSS config
│   ├── postcss.config.js        # PostCSS config
│   └── tsconfig.json            # TypeScript config
├── bin/                         # Compiled binaries
├── data/                        # SQLite database files
│   └── app.db
├── tmp/                         # Air temporary files
│   └── build-errors.log
├── .air.toml                    # Air live reload config
├── .env                         # Local environment variables (git-ignored)
├── .env.sample                  # Environment variables template
├── .gitignore                   # Git ignore patterns
├── Dockerfile                   # Container image definition
├── build-and-push.sh            # Docker build and push script
├── version.txt                  # Application version
├── Makefile                     # Build automation
├── CLAUDE.md                    # Claude Code context (this file)
└── README.md                    # Project documentation
```

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
- **Internationalization**: Support for English (en_US), Ukrainian (uk_UA), and Russian in Ukraine (ru_UA)
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
- Echo v4.13.4 (web framework)
- modernc.org/sqlite v1.38.2 (SQLite driver)
- Goose v3.25.0 (migrations)
- JWT-Go v5.3.0 (authentication)
- golang.org/x/crypto v0.41.0 (bcrypt)
- go-playground/validator v10.27.0 (validation)
- Air (live reload - auto-installed via make watch)

### Frontend
- React 18.3.1 + React DOM
- TanStack Query v5.62.7 (server state)
- TanStack Router v1.87.0 (routing)
- Chart.js v4.5.0 + react-chartjs-2 v5.3.0 (charts)
- date-fns v4.1.0 (date utilities)
- react-i18next v15.7.3 (internationalization)
- lucide-react v0.544.0 (icons)
- Tailwind CSS v4.1.13 (styling)
- Vite v6.0.5 (build tool)
- TypeScript v5.6.2

## Build Process
1. Frontend builds to `web/dist/` (via Vite)
2. Go embeds `web/dist` files via `web/embed.go` using `//go:embed`
3. Single binary (`main`) serves both API and static files
4. Air watches Go, TS, TSX, JS, HTML, CSS files and triggers rebuilds
5. Air excludes `web/dist`, `web/node_modules`, `tmp`, `bin`, `data` from watching
6. Build errors logged to `tmp/build-errors.log`
7. Docker images versioned via `version.txt` (current: v1.2.12)
8. **IMPORTANT**: `build-and-push.sh` always builds without cache (`--no-cache`) and removes `web/dist/` before building

## Features
- **User Authentication**: JWT-based login/logout with bcrypt password hashing
- **Calorie Tracking**: Add food entries with name, weight, kcal/100g, auto-calculated totals
- **Weight Management**: Track weight history over time with charts
- **Ingredient Database**: Global ingredients with multilingual names and nutritional data
- **User Profiles**: Personal settings and preference management
- **Reports & Analytics**: Data visualization and calorie/weight trends
- **AI Insights**: AI-powered nutrition and weight analysis with personalized recommendations
  - OpenAI integration (GPT-4o-mini by default)
  - Provider selection in UI (extensible architecture for multiple providers)
  - Customizable analysis periods (7, 14, 30, 90 days)
  - Optional specific questions for targeted advice
  - Multilingual responses based on user's language preference
- **Dashboard**: View today's entries with total calorie count
- **Internationalization**: Full i18n support (en_US, uk_UA, ru_UA) with language switcher
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
- **auth**: Authentication, JWT management, password hashing
- **calorie**: Calorie entry business logic and calculations
- **ingredient**: Global ingredient management with translations
- **weight**: Weight tracking and history management
- **profile**: User profile and preferences
- **reports**: Analytics, trends, and data aggregation

### Data Layer (`internal/repositories/`)
- Repository pattern for data access
- Interface-based design for testability
- Separate repositories for each domain entity
- Centralized query management in `queries.go`
- Custom error types in `errors.go` for consistent error handling
- SQL file loader utility in `SqlLoader.go` for loading SQL from files

### API Endpoints
All API routes are prefixed with `/api`:
- `/api/auth/*` - Authentication (login, logout, refresh)
- `/api/calories/*` - Calorie entry CRUD operations
- `/api/ingredients/*` - Ingredient search and management
- `/api/weight/*` - Weight history tracking
- `/api/profile/*` - User profile management
- `/api/reports/*` - Analytics and reporting
- `/api/ai/*` - AI analysis endpoints
  - `GET /api/ai/providers` - List available AI providers
  - `POST /api/ai/analyze` - Perform AI analysis

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