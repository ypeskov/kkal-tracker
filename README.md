# Kkal-Tracker

A comprehensive calorie tracking and weight management web application built with Go/Echo backend and React/TanStack frontend.

**🚀 Zero-dependency deployment** - Compiles to a single binary with embedded frontend assets. Uses SQLite by default for maximum portability and minimal infrastructure requirements. Perfect for self-hosting, Docker containers, or any environment where you want a lightweight, production-ready application without external database dependencies.

## Features

### Core Functionality
- **Calorie Tracking**: Add food entries with name, weight, kcal/100g, auto-calculated totals
- **Weight Management**: Track weight history over time with interactive charts
- **Ingredient Database**: Global ingredients with multilingual names and nutritional data
- **User Profiles**: Personal settings and preference management
- **Reports & Analytics**: Data visualization and calorie/weight trends with Chart.js
- **Dashboard**: Real-time view of today's entries with total calorie count
- **Internationalization**: Full i18n support (English, Ukrainian, Russian) with language switcher

### Technical Features
- **Backend**: Go with Echo framework v4.13.4
- **Frontend**: React 18 with TanStack Query & Router, TypeScript, Vite
- **Database**: Dual support - SQLite (default) with PostgreSQL option
- **Authentication**: JWT-based auth (v5.3.0) with bcrypt password hashing, sessionStorage persistence
- **Logging**: Structured logging with slog
- **Build System**: Single binary with embedded frontend assets
- **Live Reload**: Air for development hot-reload
- **Database Backup**: Optional automated backups to Google Drive

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- npm

### Installation

1. Install dependencies:
```bash
make init  # Installs all dependencies and initial setup
```

2. Run database migrations:
```bash
make migrate-up
```

3. Seed database with sample ingredients (optional):
```bash
make seed
```

4. Create a test user:
```bash
# Default (English ingredients)
go run scripts/create_user.go user@example.com password123

# With specific language (en_US, uk_UA, or ru_UA)
go run scripts/create_user.go user@example.com password123 ru_UA
```

## Demo User

For quick testing, there's a demo user created during database seeding:

**Email**: `example@example.com`  
**Password**: `password123`

You can login immediately after running `make seed` without creating additional users.

5. Build and run:
```bash
make run
```

Or for development:
```bash
make dev
```

### Environment Configuration

Copy `.env.sample` to `.env` and adjust values as needed:

```env
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

#### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_TYPE` | `sqlite` | Database provider to use (`sqlite` or `postgres`) |
| `DATABASE_PATH` | `./data/kkal_tracker.db` | Path to SQLite database file (used when `DATABASE_TYPE=sqlite`) |
| `POSTGRES_URL` | _empty_ | PostgreSQL connection string (used when `DATABASE_TYPE=postgres`) |
| `PORT` | `8080` | HTTP server port |
| `JWT_SECRET` | `default-secret-key` | Secret key for JWT token signing (change in production!) |
| `LOG_LEVEL` | `info` | Logging level (`debug`, `info`, `warn`, `error`) |
| `ENVIRONMENT` | `development` | Application environment (`development`, `production`) |

#### Database Provider Selection

The application supports multiple database providers through the repository pattern:

**SQLite (default)**:
```env
DATABASE_TYPE=sqlite
DATABASE_PATH=./data/kkal_tracker.db
```

**PostgreSQL**:
```env
DATABASE_TYPE=postgres
POSTGRES_URL=postgres://username:password@localhost/kkal_tracker?sslmode=disable
```

## API Endpoints

All API routes are prefixed with `/api`:

- `/api/auth/*` - Authentication (login, logout, refresh)
- `/api/calories/*` - Calorie entry CRUD operations
- `/api/ingredients/*` - Ingredient search and management
- `/api/weight/*` - Weight history tracking
- `/api/profile/*` - User profile management
- `/api/reports/*` - Analytics and reporting

## Development

### Make Commands

```bash
# Development
make dev            # Start development with Air live reload
make watch          # Alternative to 'make dev' with Air auto-install prompt
make build          # Build both frontend and backend
make run            # Build and run the application
make clean          # Clean build artifacts
make init           # Install dependencies and initial setup

# Database Migrations
make migrate-up     # Run all pending migrations
make migrate-down   # Rollback last migration
make migrate-status # Check migration status
make migrate-create NAME=migration_name # Create new migration

# Database Seeding
make seed           # Seed database with initial data
make seed-clean     # Clean and re-seed database (dev only)

# Testing & Quality
make test           # Run all Go tests
```

### Live Reload Development

The project includes Air for live reloading during development. Air automatically rebuilds and restarts your application when Go files change.

```bash
make watch          # Start Air live reload
```

Air will:
- Watch for changes in `.go`, `.tsx`, `.ts`, `.html`, `.css`, `.js` files
- Automatically rebuild the application
- Restart the server with the new binary
- Log build errors to `build-errors.log`
- Auto-install Air if not present (with confirmation)

### Database

The application supports multiple database providers through a repository pattern with **Goose migrations**:

- **SQLite**: Default provider, database file created at `DATABASE_PATH`
- **PostgreSQL**: Alternative provider using connection string from `POSTGRES_URL`

The active database provider is controlled by the `DATABASE_TYPE` environment variable.

#### Migration Management

The project uses [Goose](https://github.com/pressly/goose) for database migrations with **manual control only**:

- **Run migrations**: `make migrate-up`
- **Rollback**: `make migrate-down` 
- **Status**: `make migrate-status`
- **Create new**: `make migrate-create NAME=add_new_table`

**Important**: Migrations are **NOT** run automatically on startup. You must run them manually:

```bash
# Before first run or after pulling new migrations
make migrate-up

# Then start the application
make run
```

#### Database Seeding

The project includes a seeding system to populate the database with sample ingredients data:

- **Seed location**: SQL files in `cmd/seed/sql/` directory
- **Seed command**: `make seed` - Executes all SQL files in alphabetical order
- **Clean and reseed**: `make seed-clean` - Removes existing data and re-seeds

The seeding system:
- Reads all `.sql` files from `cmd/seed/sql/`
- Executes them in alphabetical order (name files like `01_ingredients.sql`, `02_categories.sql`)
- Uses transactions for safety
- Supports idempotent operations with `INSERT OR REPLACE`

For production Docker deployments:
```dockerfile
# Option 1: During build
RUN go run cmd/seed/main.go

# Option 2: With environment variable
ENV SEED_DB=true
CMD sh -c "if [ \"$SEED_DB\" = \"true\" ]; then go run cmd/seed/main.go; fi && ./main"
```

### Authentication

- Passwords are hashed using bcrypt
- JWT tokens are valid for 24 hours
- Protected routes require `Authorization: Bearer <token>` header

### User Creation

When creating users with `scripts/create_user.go`:
- Global ingredients are automatically copied to user's personal ingredients
- Language can be specified (en_US, uk_UA, ru_UA) to copy ingredients with appropriate translations
- Default language is en_US if not specified

## Tech Stack

### Backend
- **Echo v4.13.4** - Web framework
- **modernc.org/sqlite v1.38.2** - SQLite driver
- **Goose v3.25.0** - Database migrations
- **JWT-Go v5.3.0** - Authentication
- **golang.org/x/crypto v0.41.0** - bcrypt password hashing
- **go-playground/validator v10.27.0** - Input validation
- **Air** - Live reload for development

### Frontend
- **React 18.3.1** - UI library
- **TanStack Query v5.62.7** - Server state management
- **TanStack Router v1.87.0** - Routing
- **Chart.js v4.5.0** - Data visualization
- **react-i18next v15.7.3** - Internationalization
- **Tailwind CSS v4.1.13** - Styling
- **Vite v6.0.5** - Build tool
- **TypeScript v5.6.2** - Type safety

## Deployment

### Production Build

1. Set `ENVIRONMENT=production` in your environment
2. Change `JWT_SECRET` to a secure random string
3. Configure your database:
   - For SQLite: Set `DATABASE_TYPE=sqlite` and `DATABASE_PATH`
   - For PostgreSQL: Set `DATABASE_TYPE=postgres` and `POSTGRES_URL`
4. Build with `make build`
5. Deploy the binary and serve on your preferred port

The application compiles to a single binary with embedded frontend assets, making it easy to deploy anywhere.

### Kubernetes Deployment

Full Kubernetes configurations are available in the `kubernetes/` directory:
- **Base configurations**: Common resources for all environments
- **Environment overlays**: Specific configs for dev and prod
- **Automated backups**: Includes CronJob for scheduled database backups to Google Drive

### Docker Support

The application is Docker-ready and can be containerized for easy deployment. The single binary architecture ensures minimal container size and fast startup times.