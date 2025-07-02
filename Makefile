# Server Management System Makefile

.PHONY: help build run test clean docker-up docker-down migrate-up migrate-down swagger wire fmt lint

# Default target
help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build
build: ## Build the application
	@echo "Building server..."
	@go build -o bin/server cmd/server/main.go
	@echo "Building migrate tool..."
	@go build -o bin/migrate cmd/migrate/main.go
	@echo "Build completed!"

# Run
run: ## Run the server
	@echo "Starting server..."
	@go run cmd/server/main.go

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@go run cmd/migrate/main.go up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@go run cmd/migrate/main.go down

# Development
dev: docker-up migrate-up run ## Start development environment

# Testing
test: ## Run tests
	@echo "Running tests..."
	@go test ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

# Wire and Swagger
wire: ## Generate Wire dependency injection code
	@echo "Generating Wire code..."
	@go generate ./internal/wiring

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger docs..."
	@swag init -g cmd/server/main.go -o docs


# Docker
docker-up: ## Start Docker services
	@echo "Starting Docker services..."
	@docker-compose up -d

docker-down: ## Stop Docker services
	@echo "Stopping Docker services..."
	@docker-compose down

docker-logs: ## View Docker logs
	@docker-compose logs -f

# ELK Stack
elk-up: ## Start ELK stack
	@echo "Starting ELK stack..."
	@./scripts/start_elk.sh

elk-setup: ## Setup Kibana
	@echo "Setting up Kibana..."
	@./scripts/setup_kibana.sh

# Fake servers for testing
fake-servers-start: ## Start fake servers for testing
	@echo "Starting fake servers..."
	@./scripts/start_fake_servers.sh

fake-servers-stop: ## Stop fake servers
	@echo "Stopping fake servers..."
	@./scripts/stop_fake_servers.sh

# Clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean completed!"

# Install dependencies
deps: ## Install/update dependencies
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Security check
security: ## Run security checks
	@echo "Running security checks..."
	@gosec ./...

# All checks before commit
check: fmt lint test security ## Run all checks (format, lint, test, security)

# Production build
build-prod: ## Build for production
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server cmd/server/main.go
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/migrate cmd/migrate/main.go
	@echo "Production build completed!"

# Database
db-reset: migrate-down migrate-up ## Reset database

# Export/Import
export-servers: ## Export servers to Excel
	@echo "Exporting servers..."
	@curl -X GET "http://localhost:8080/api/v1/servers/export" -H "Authorization: Bearer YOUR_TOKEN" --output servers_export.xlsx
	@echo "Servers exported to servers_export.xlsx"

# Monitoring
logs: ## Tail application logs
	@tail -f logs/app.log

# Setup development environment
setup: deps wire swagger docker-up migrate-up ## Setup complete development environment
	@echo "Development environment setup completed!"
	@echo "Server will be available at: http://localhost:8080"
	@echo "Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "Kibana: http://localhost:5601"
