# Calorie Tracker

A modern calorie tracking web application built with Go/Echo backend and React/TanStack frontend.

**🚀 Zero-dependency deployment** - Compiles to a single binary with embedded frontend assets. Uses SQLite by default for maximum portability and minimal infrastructure requirements. Perfect for self-hosting, Docker containers, or any environment where you want a lightweight, production-ready application without external database dependencies.

## Features

- **Backend**: Go with Echo framework
- **Frontend**: React with TanStack Query for state management
- **Database**: Multi-provider support (SQLite/PostgreSQL) with migrations ⚠️ *PostgreSQL not yet implemented*
- **Authentication**: JWT-based auth with bcrypt password hashing
- **Logging**: Structured logging with slog
- **Build System**: Embedded file system for serving frontend assets

## Project Structure

```
├── assets/                 # Embedded static assets
├── cmd/web/               # Application entry point
├── internal/
│   ├── auth/              # JWT authentication
│   ├── config/            # Configuration management
│   ├── database/          # Database connection and migrations
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Data models and repositories
│   ├── routes/
│   │   ├── api/           # API route handlers
│   │   └── web/           # Web route handlers
│   └── server/            # Server setup
├── web/                   # React frontend
│   ├── src/
│   │   ├── api/           # API client code
│   │   ├── components/    # React components
│   │   └── ...
├── scripts/               # Utility scripts
├── Makefile              # Build automation
└── .env                  # Environment variables
```

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- npm

### Installation

1. Install dependencies:
```bash
make install-deps
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
# Database Configuration
DATABASE_TYPE=sqlite
DATABASE_PATH=./data/kkal_tracker.db
# POSTGRES_URL=postgres://user:password@localhost/kkal_tracker?sslmode=disable

# Server Configuration
PORT=8080
JWT_SECRET=your-jwt-secret-key-change-this-in-production
LOG_LEVEL=debug
```

#### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_TYPE` | `sqlite` | Database provider to use (`sqlite` or `postgres`) |
| `DATABASE_PATH` | `./data/kkal_tracker.db` | Path to SQLite database file (used when `DATABASE_TYPE=sqlite`) |
| `POSTGRES_URL` | _empty_ | PostgreSQL connection string (used when `DATABASE_TYPE=postgres`) ⚠️ *Not yet implemented* |
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

**PostgreSQL** ⚠️ *Not yet implemented*:
```env
DATABASE_TYPE=postgres
POSTGRES_URL=postgres://username:password@localhost/kkal_tracker?sslmode=disable
```

## Frontend Features

- Login form with validation
- Dashboard with calorie entry form
- Today's entries list with total calories
- TanStack Query for efficient data fetching and caching

## Development

### Make Commands

```bash
make build          # Build the application
make run            # Build and run
make dev            # Run in development mode
make watch          # Live reload with Air (recommended for development)
make build-frontend # Build frontend only
make install-deps   # Install all dependencies
make clean          # Clean build artifacts
make test           # Run tests

# Database Migrations
make migrate-up     # Run pending migrations
make migrate-down   # Rollback last migration
make migrate-status # Check migration status
make migrate-create NAME=migration_name # Create new migration

# Database Seeding  
make seed           # Seed database with sample data
make seed-clean     # Clean and re-seed database
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
- **PostgreSQL**: Alternative provider using connection string from `POSTGRES_URL` ⚠️ *Not yet implemented*

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
- **Echo**: HTTP web framework (v4.13.4)
- **Database**: Multi-provider architecture with SQLite driver (v1.14.32) and PostgreSQL support ⚠️ *PostgreSQL not yet implemented*
- **Goose**: Database migrations with pressly/goose/v3 (v3.25.0)
- **JWT**: golang-jwt/jwt/v5 for authentication (v5.3.0)
- **Crypto**: golang.org/x/crypto for password hashing (v0.41.0)
- **Config**: godotenv for environment management (v1.5.1)
- **Logging**: Standard library slog

### Frontend
- **React 18**: UI library
- **TanStack Query**: Server state management
- **Vite**: Build tool and dev server
- **TypeScript**: Type safety

## Production Deployment

1. Set `ENVIRONMENT=production` in your environment
2. Change `JWT_SECRET` to a secure random string
3. Configure your database:
   - For SQLite: Set `DATABASE_TYPE=sqlite` and `DATABASE_PATH`
   - For PostgreSQL: Set `DATABASE_TYPE=postgres` and `POSTGRES_URL` ⚠️ *Not yet implemented*
4. Build with `make build`
5. Deploy the binary and serve on your preferred port