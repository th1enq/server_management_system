# Server Management System Makefile

.PHONY: help run reset gen_swagger migrate-down migrate-up

# Default target
help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build
run: ## Build the application
	@./scripts/run_server.sh

# Run
reset: ## Run the server
	@echo "Shutting down docker..."
	@docker-compose down -v

gen-swagger: ## Run database migrations up
	@./scripts/gen_swagger.sh

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@go run ./cmd/migrate down

migrate-up: ## Rollback database migrations
	@echo "Migrations..."
	@go run ./cmd/migrate up

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverpkg=./internal/... -coverprofile=coverage.out ./tests/unit/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
