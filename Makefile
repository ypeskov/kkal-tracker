BINARY_NAME=kkal-tracker
WEB_DIR=web

.PHONY: build run clean dev install-deps build-frontend watch migrate-up migrate-down migrate-status migrate-create seed

build: build-frontend
	@echo "Building..."
	@go build -o main cmd/web/main.go

run: build
	@echo "Running..."
	@./main

dev:
	@echo "Running in development mode..."
	@go run cmd/web/main.go

build-frontend:
	@echo "Building frontend..."
	@cd $(WEB_DIR) && npm install --silent && npm run build

install-deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@cd $(WEB_DIR) && npm install

clean:
	@echo "Cleaning..."
	@go clean
	@rm -f main
	@rm -rf bin/
	@rm -rf $(WEB_DIR)/dist/
	@rm -rf $(WEB_DIR)/node_modules/
	@rm -rf tmp/

test:
	@echo "Testing..."
	@go test -v ./...

init: install-deps
	@echo "Creating initial user..."
	@echo "You'll need to manually insert a user into the database or create a registration endpoint"

# Live Reload with Air
watch:
	@if command -v air > /dev/null; then \
	    air; \
	    echo "Watching...";\
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        go install github.com/air-verse/air@latest; \
	        air; \
	        echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

# Database Migrations with Goose
migrate-up:
	@echo "Running migrations..."
	@go run cmd/migrate/main.go -cmd=up

migrate-down:
	@echo "Rolling back migrations..."
	@go run cmd/migrate/main.go -cmd=down

migrate-status:
	@echo "Checking migration status..."
	@go run cmd/migrate/main.go -cmd=status

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@if command -v goose > /dev/null; then \
		goose -dir migrations create $(NAME) sql; \
	else \
		echo "Installing goose..."; \
		go install github.com/pressly/goose/v3/cmd/goose@latest; \
		goose -dir migrations create $(NAME) sql; \
	fi

# Database Seeding
seed:
	@echo "Seeding database..."
	@go run cmd/seed/main.go

seed-clean:
	@echo "Cleaning and re-seeding database..."
	@sqlite3 ./data/app.db "DELETE FROM global_ingredients; DELETE FROM global_ingredient_names;"
	@go run cmd/seed/main.go

.DEFAULT_GOAL := build