# Claude Code Context

## Project Overview
A Go/Echo web application for calorie tracking with React/TanStack frontend. Built using `/Users/ypeskov/Projects/Go/qr-generator` as structural reference.

## Architecture
- **Backend**: Go with Echo framework v4.13.4
- **Frontend**: React 18 with TanStack Query, TypeScript, Vite
- **Database**: SQLite with Goose migrations v3.25.0
- **Auth**: JWT with bcrypt password hashing, sessionStorage persistence
- **I18n**: react-i18next with en_US, uk_UA, ru_UA locales
- **Logging**: Structured logging with slog
- **Dev Tools**: Air for live reload

## Key Commands

### Development
```bash
make dev          # Start development with Air live reload
make build        # Build both frontend and backend
make clean        # Clean build artifacts
```

### Database Migrations
```bash
make migrate-up           # Run all pending migrations
make migrate-down         # Rollback last migration
make migrate-status       # Check migration status
make migrate-create NAME=migration_name  # Create new migration
```

### Testing & Quality
```bash
go test ./...     # Run Go tests
go vet ./...      # Go static analysis
go mod tidy       # Clean up dependencies
```

## Project Structure
```
├── cmd/web/main.go              # Application entry point
├── internal/
│   ├── config/config.go         # Configuration management
│   ├── logger/logger.go         # Structured logging setup
│   ├── database/database.go     # Database connection & migrations
│   ├── server/server.go         # Echo server setup
│   ├── handlers/                # HTTP handlers
│   ├── middleware/              # Custom middleware
│   ├── routes/api/              # API route handlers
│   ├── services/                # Business logic services
│   └── models/                  # Data models & repositories
├── migrations/                  # Goose SQL migrations
├── web/                         # React frontend
│   ├── src/
│   │   ├── api/                 # API service classes
│   │   ├── components/          # React components
│   │   └── i18n/                # Internationalization
│   ├── dist/                    # Built frontend assets
│   ├── embed.go                 # Go embedded filesystem
│   └── package.json
├── .air.toml                    # Air live reload config
└── Makefile                     # Build automation
```

## Environment Variables
```
PORT=8080
ENV=development
LOG_LEVEL=debug
DATABASE_PATH=./data/app.db
JWT_SECRET=your-secret-key
```

## Important Notes
- **No auto-migrations**: Migrations must be run manually for production safety
- **Embedded assets**: Frontend dist files are embedded in Go binary via `web/embed.go`
- **API routes**: All API endpoints under `/api` prefix
- **Authentication**: JWT tokens stored in sessionStorage, persist across page reloads
- **Internationalization**: Support for English (en_US), Ukrainian (uk_UA), and Russian in Ukraine (ru_UA)
- **Development**: Uses Air for live reload, excludes `web/dist` from watching
- **Production**: JSON logging format, structured error handling

## Dependencies
- Echo v4.13.4 (web framework)
- SQLite driver v1.14.32
- Goose v3.25.0 (migrations)
- JWT-Go v5.3.0 (authentication)
- Air v1.61.5 (live reload)
- React 18 + TanStack Query (frontend)
- react-i18next (internationalization)

## Build Process
1. Frontend builds to `web/dist/`
2. Go embeds `web/dist` files via `web/embed.go`
3. Single binary serves both API and static files
4. Air watches Go files and triggers rebuilds (excludes `web/dist`)

## Features
- **User Authentication**: JWT-based login/logout with bcrypt password hashing
- **Calorie Tracking**: Add food entries with name, weight, kcal/100g, auto-calculated totals
- **Dashboard**: View today's entries with total calorie count
- **Internationalization**: Full i18n support with language switcher
- **Persistent Sessions**: JWT tokens stored in sessionStorage
- **Responsive UI**: Clean, modern interface with proper form validation