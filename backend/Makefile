.PHONY: help db-generate db-migrate db-migrate-up db-migrate-down build test clean

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Database operations
db-generate: ## Generate Go code from SQL using sqlc
	@echo "Generating repository code with sqlc..."
	@for dir in queries/*/; do \
		if [ -f "$$dir/sqlc.yaml" ]; then \
			echo "Generating from $$dir"; \
			cd "$$dir" && sqlc generate && cd ../..; \
		fi; \
	done
	@echo "✅ Repository code generated successfully"

db-migrate: ## Show migration status
	@echo "Current migration status:"
	@goose -dir migrations postgres "$(DATABASE_URL)" status

db-migrate-up: ## Run all pending migrations
	@echo "Running migrations..."
	@goose -dir migrations postgres "$(DATABASE_URL)" up
	@echo "✅ Migrations completed"

db-migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@goose -dir migrations postgres "$(DATABASE_URL)" down
	@echo "✅ Migration rolled back"

db-migrate-reset: ## Reset all migrations (WARNING: This will drop all data)
	@echo "⚠️  WARNING: This will reset all migrations and drop all data!"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@goose -dir migrations postgres "$(DATABASE_URL)" reset
	@echo "✅ Database reset completed"

# Development operations
build: ## Build the application
	@echo "Building application..."
	@mkdir -p tmp
	@go build -o tmp/api ./cmd/api
	@go build -o tmp/worker ./cmd/worker
	@echo "✅ Build completed"

test: ## Run tests
	@echo "Running tests..."
	@go test ./...
	@echo "✅ Tests completed"

test-crypto: ## Run encryption/decryption tests
	@echo "Testing encryption functionality..."
	@go test ./internal/crypto -v
	@echo "✅ Crypto tests completed"

# Code quality
lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run
	@echo "✅ Linting completed"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatting completed"

# Clean up
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf tmp/
	@echo "✅ Cleanup completed"

# Setup commands
setup: ## Install dependencies and tools
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

# Full regeneration workflow
regen: db-generate fmt ## Regenerate database code and format
	@echo "✅ Full regeneration completed"

# Development workflow
dev-setup: setup db-migrate-up db-generate ## Complete development setup
	@echo "✅ Development environment setup completed"

# Environment validation
validate-env: ## Validate required environment variables
	@echo "Validating environment variables..."
	@test -n "$(DATABASE_URL)" || (echo "❌ DATABASE_URL is required" && exit 1)
	@test -n "$(ENCRYPTION_SECRET_KEY)" || (echo "❌ ENCRYPTION_SECRET_KEY is required" && exit 1)
	@test $$(echo -n "$(ENCRYPTION_SECRET_KEY)" | wc -c) -eq 32 || (echo "❌ ENCRYPTION_SECRET_KEY must be exactly 32 bytes" && exit 1)
	@echo "✅ Environment validation passed"

# Quick development commands
quick-test: db-generate test ## Quick test cycle: generate code and run tests
	@echo "✅ Quick test cycle completed"

all: validate-env db-migrate-up db-generate test build ## Full pipeline: validate, migrate, generate, test, build
	@echo "✅ Full pipeline completed successfully"
