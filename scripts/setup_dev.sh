#!/bin/bash

# Development Setup Script for Server Management System
# This script helps developers set up the project quickly

set -e

echo "🚀 Setting up Server Management System Development Environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.19 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go $GO_VERSION is installed"
}

# Check if Docker is installed and running
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker."
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker is not running. Please start Docker."
        exit 1
    fi
    
    print_success "Docker is installed and running"
}

# Check if Docker Compose is installed
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose."
        exit 1
    fi
    
    print_success "Docker Compose is installed"
}

# Install Go tools
install_tools() {
    print_status "Installing Go development tools..."
    
    # Wire for dependency injection
    if ! command -v wire &> /dev/null; then
        print_status "Installing Wire..."
        go install github.com/google/wire/cmd/wire@latest
    fi
    
    # Swag for Swagger documentation
    if ! command -v swag &> /dev/null; then
        print_status "Installing Swag..."
        go install github.com/swaggo/swag/cmd/swag@latest
    fi
    
    # Golangci-lint for linting
    if ! command -v golangci-lint &> /dev/null; then
        print_status "Installing golangci-lint..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
    fi
    
    # Goimports for import formatting
    if ! command -v goimports &> /dev/null; then
        print_status "Installing goimports..."
        go install golang.org/x/tools/cmd/goimports@latest
    fi
    
    # Gosec for security scanning
    if ! command -v gosec &> /dev/null; then
        print_status "Installing gosec..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    fi
    
    print_success "Go tools installed successfully"
}

# Download dependencies
install_dependencies() {
    print_status "Installing Go dependencies..."
    go mod tidy
    go mod download
    print_success "Dependencies installed"
}

# Generate Wire code
generate_wire() {
    print_status "Generating Wire dependency injection code..."
    go generate ./internal/wiring
    print_success "Wire code generated"
}

# Generate Swagger docs
generate_swagger() {
    print_status "Generating Swagger documentation..."
    if [ -f "./scripts/gen_swagger.sh" ]; then
        chmod +x ./scripts/gen_swagger.sh
        ./scripts/gen_swagger.sh
    else
        swag init -g cmd/server/main.go -o docs/
    fi
    print_success "Swagger documentation generated"
}

# Start Docker services
start_docker() {
    print_status "Starting Docker services..."
    docker-compose up -d
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 10
    
    # Check if PostgreSQL is ready
    until docker-compose exec -T postgres pg_isready -h localhost -p 5432 -U postgres; do
        print_status "Waiting for PostgreSQL..."
        sleep 2
    done
    
    print_success "Docker services started successfully"
}

# Run database migrations
run_migrations() {
    print_status "Running database migrations..."
    go run cmd/migrate/main.go up
    print_success "Database migrations completed"
}

# Build application
build_app() {
    print_status "Building application..."
    make build
    print_success "Application built successfully"
}

# Run tests
run_tests() {
    print_status "Running tests..."
    go test ./...
    print_success "All tests passed"
}

# Create .env file if it doesn't exist
create_env() {
    if [ ! -f ".env" ]; then
        print_status "Creating .env file..."
        cat > .env << EOF
# Development Environment Variables
APP_ENV=development
LOG_LEVEL=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=server_management

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Elasticsearch
ES_HOST=localhost
ES_PORT=9200

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Email (for testing)
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
EOF
        print_success ".env file created"
    else
        print_warning ".env file already exists"
    fi
}

# Setup complete development environment
setup_dev() {
    print_status "Setting up complete development environment..."
    
    check_go
    check_docker
    check_docker_compose
    
    create_env
    install_tools
    install_dependencies
    generate_wire
    generate_swagger
    start_docker
    run_migrations
    build_app
    
    print_success "🎉 Development environment setup completed!"
    echo ""
    echo "📋 Next steps:"
    echo "   1. Start the server: make run"
    echo "   2. Open Swagger UI: http://localhost:8080/swagger/index.html"
    echo "   3. Check health: http://localhost:8080/health"
    echo ""
    echo "🔧 Available commands:"
    echo "   make help          - Show all available commands"
    echo "   make dev           - Start development environment"
    echo "   make test          - Run tests"
    echo "   make clean         - Clean build artifacts"
    echo ""
}

# Main execution
case "${1:-setup}" in
    "setup")
        setup_dev
        ;;
    "check")
        check_go
        check_docker
        check_docker_compose
        ;;
    "tools")
        install_tools
        ;;
    "deps")
        install_dependencies
        ;;
    "wire")
        generate_wire
        ;;
    "swagger")
        generate_swagger
        ;;
    "docker")
        start_docker
        ;;
    "migrate")
        run_migrations
        ;;
    "build")
        build_app
        ;;
    "test")
        run_tests
        ;;
    "env")
        create_env
        ;;
    *)
        echo "Usage: $0 {setup|check|tools|deps|wire|swagger|docker|migrate|build|test|env}"
        echo ""
        echo "Commands:"
        echo "  setup    - Complete development environment setup (default)"
        echo "  check    - Check prerequisites"
        echo "  tools    - Install Go development tools"
        echo "  deps     - Install Go dependencies"
        echo "  wire     - Generate Wire code"
        echo "  swagger  - Generate Swagger documentation"
        echo "  docker   - Start Docker services"
        echo "  migrate  - Run database migrations"
        echo "  build    - Build application"
        echo "  test     - Run tests"
        echo "  env      - Create .env file"
        exit 1
        ;;
esac
