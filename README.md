# Calorie Tracker

A modern calorie tracking web application built with Go/Echo backend and React/TanStack frontend.

## Features

- **Backend**: Go with Echo framework
- **Frontend**: React with TanStack Query for state management
- **Database**: SQLite with migrations
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

3. Create a test user:
```bash
go run scripts/create_user.go user@example.com password123
```

4. Build and run:
```bash
make run
```

Or for development:
```bash
make dev
```

### Environment Configuration

Copy `.env` and adjust values as needed:

```env
DATABASE_PATH=./data/kkal_tracker.db
PORT=8080
JWT_SECRET=your-jwt-secret-key-change-this-in-production
LOG_LEVEL=debug
```

## API Endpoints

### Authentication
- `POST /api/auth/login` - Login user
- `GET /api/auth/me` - Get current user (requires auth)

### Calorie Entries
- `GET /api/calories` - Get calorie entries (requires auth)
- `POST /api/calories` - Create calorie entry (requires auth)
- `DELETE /api/calories/:id` - Delete calorie entry (requires auth)

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

The application uses SQLite with **Goose migrations**. The database file is created at the path specified in `DATABASE_PATH` environment variable.

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

### Authentication

- Passwords are hashed using bcrypt
- JWT tokens are valid for 24 hours
- Protected routes require `Authorization: Bearer <token>` header

## Tech Stack

### Backend
- **Echo**: HTTP web framework (v4.13.4)
- **SQLite**: Database with mattn/go-sqlite3 driver (v1.14.32)
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
3. Configure your database path appropriately
4. Build with `make build`
5. Deploy the binary and serve on your preferred port