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
│   │   ├── hooks/               # Custom React hooks
│   │   ├── i18n/                # Internationalization
│   │   ├── pages/               # Page components
│   │   ├── styles/              # Global styles and CSS
│   │   ├── utils/               # Utility functions
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

## Features
- **User Authentication**: JWT-based login/logout with bcrypt password hashing
- **Calorie Tracking**: Add food entries with name, weight, kcal/100g, auto-calculated totals
- **Weight Management**: Track weight history over time with charts
- **Ingredient Database**: Global ingredients with multilingual names and nutritional data
- **User Profiles**: Personal settings and preference management
- **Reports & Analytics**: Data visualization and calorie/weight trends
- **Dashboard**: View today's entries with total calorie count
- **Internationalization**: Full i18n support (en_US, uk_UA, ru_UA) with language switcher
- **Persistent Sessions**: JWT tokens stored in sessionStorage
- **Responsive UI**: Clean, modern interface with proper form validation
- **Database Backup**: Automated backups to Google Drive via OAuth2

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

## Deployment

### Docker
- **Dockerfile**: Multi-stage build for optimized image size
- **Build script**: `build-and-push.sh` - Automated Docker build and push
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