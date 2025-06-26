# Server Management System

## 🏗️ **KIẾN TRÚC PROJECT**

Project này sử dụng **Clean Architecture** với các layer được tách biệt rõ ràng:

```
server_management_system/
├── cmd/                          # Entry points
│   ├── server/                   # Main HTTP server
│   └── migrate/                  # Database migration tool
├── configs/                      # Configuration files
│   ├── config.yaml               # Main config file
│   └── configs.go                # Config constants
├── internal/                     # Private application code
│   ├── app/                      # Application layer
│   ├── configs/                  # Configuration structs
│   ├── controller/               # HTTP Controllers (Presentation layer)
│   ├── dataaccess/              # Data access layer
│   │   ├── cache/               # Redis cache
│   │   ├── database/            # PostgreSQL database
│   │   └── elasticsearch/       # Elasticsearch client
│   ├── handler/                  # Request handlers
│   │   ├── http/                # HTTP route handlers
│   │   └── jobs/                # Background job handlers
│   ├── models/                   # Domain models
│   ├── repositories/             # Data repositories (Data layer)
│   ├── services/                 # Business logic (Domain layer)
│   ├── utils/                    # Utility functions
│   ├── wiring/                   # Dependency injection (Wire)
│   └── worker/                   # Background workers
├── docs/                         # API documentation (Swagger)
├── exports/                      # Export files
├── logs/                         # Log files
├── migrations/                   # Database migrations
├── scripts/                      # Utility scripts
└── template/                     # Email templates
```

## 📋 **LAYER DESCRIPTIONS**

### **1. Presentation Layer (Controllers)**
- `internal/controller/`: REST API controllers
- `internal/handler/http/`: HTTP route configuration
- Responsible for HTTP request/response handling

### **2. Domain Layer (Services)**
- `internal/services/`: Business logic implementation
- `internal/models/`: Domain entities and value objects
- Contains core business rules

### **3. Data Layer (Repositories)**
- `internal/repositories/`: Data access interfaces and implementations
- `internal/dataaccess/`: Database, cache, and external service clients
- Handles data persistence and retrieval

### **4. Infrastructure Layer**
- `internal/configs/`: Configuration management
- `internal/utils/`: Shared utilities
- `internal/wiring/`: Dependency injection setup

## 🔧 **DEPENDENCY INJECTION**

Project sử dụng Google Wire cho dependency injection:
- Wire configuration: `internal/wiring/wire.go`
- Generated code: `internal/wiring/wire_gen.go`
- Individual wiresets in each package

## 🚀 **COMMANDS**

### Development
```bash
# Start server
go run cmd/server/main.go

# Run migrations
go run cmd/migrate/main.go up

# Generate Wire DI
go generate ./internal/wiring

# Generate Swagger docs
./scripts/gen_swagger.sh
```

### Docker
```bash
# Start infrastructure
docker-compose up -d

# Start ELK stack
./scripts/start_elk.sh
```

## 📝 **API DOCUMENTATION**

Swagger UI available at: `http://localhost:8080/swagger/index.html`

## 🔧 **CONFIGURATION**

Configuration is managed through:
- `configs/config.yaml`: Main configuration file
- Environment variables (override config file)
- Default values embedded in code

## 🧪 **TESTING**

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 **MONITORING**

- **Logging**: Structured logging with Zap
- **Metrics**: Elasticsearch integration
- **Health checks**: `/health` endpoint
- **Reports**: Daily automated reports

## 🔒 **SECURITY**

- JWT authentication
- CORS configuration
- Input validation
- SQL injection prevention (GORM ORM)
