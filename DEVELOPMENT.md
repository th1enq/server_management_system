# 🛠️ Development Guide

## 📁 **CẤU TRÚC PROJECT SAU REFACTOR**

Sau khi refactor, project đã được tổ chức theo **Clean Architecture** với cấu trúc rõ ràng:

```
server_management_system/
├── 📁 cmd/                       # Entry points
│   ├── server/main.go           # HTTP server entrypoint  
│   └── migrate/main.go          # Database migration tool
│
├── 📁 configs/                   # Configuration files
│   ├── config.yaml              # Main configuration
│   └── configs.go               # Default config constants
│
├── 📁 internal/                  # Private application code
│   ├── 📁 app/                   # Application orchestration
│   │   ├── standalone_server.go # Main app server
│   │   └── wireset.go           # Wire DI configuration
│   │
│   ├── 📁 configs/               # Configuration management
│   │   ├── config.go            # Config structures
│   │   ├── *.go                 # Individual config types
│   │   └── wireset.go           # Config wire setup
│   │
│   ├── 📁 controller/            # 🎯 HTTP Controllers (Presentation Layer)
│   │   ├── server_handler.go    # Server CRUD endpoints
│   │   ├── report_handler.go    # Report generation endpoints
│   │   └── wireset.go           # Controller wire setup
│   │
│   ├── 📁 dataaccess/           # 💾 Data Access Layer
│   │   ├── cache/               # Redis cache client
│   │   ├── database/            # PostgreSQL & PgxPool
│   │   ├── elasticsearch/       # Elasticsearch client
│   │   └── wireset.go           # Data access wire setup
│   │
│   ├── 📁 handler/              # 🔧 Infrastructure Handlers
│   │   ├── http/                # HTTP server & routing
│   │   ├── jobs/                # Background job handlers
│   │   └── wireset.go           # Handler wire setup
│   │
│   ├── 📁 models/               # 📦 Domain Models
│   │   ├── server.go            # Server entity
│   │   ├── user.go              # User entity
│   │   ├── monitor.go           # Monitoring models
│   │   └── response.go          # API response models
│   │
│   ├── 📁 repositories/         # 🗃️ Data Repositories (Data Layer)
│   │   ├── server_repository.go # Server data operations
│   │   ├── user_repository.go   # User data operations
│   │   └── wireset.go           # Repository wire setup
│   │
│   ├── 📁 services/             # 💼 Business Services (Domain Layer)
│   │   ├── server_service.go    # Server business logic
│   │   ├── report_service.go    # Report business logic
│   │   ├── user_service.go      # User business logic
│   │   └── wireset.go           # Service wire setup
│   │
│   ├── 📁 utils/                # 🔧 Utilities
│   │   ├── log.go               # Logging utilities
│   │   ├── excelize.go          # Excel generation
│   │   ├── signal.go            # Signal handling
│   │   └── wireset.go           # Utils wire setup
│   │
│   ├── 📁 wiring/               # ⚡ Dependency Injection
│   │   ├── wire.go              # Wire configuration
│   │   └── wire_gen.go          # Generated DI code
│   │
│   └── 📁 worker/               # 🏃 Background Workers
│       └── monitoring.go        # Server monitoring worker
│
├── 📁 docs/                     # API Documentation
├── 📁 exports/                  # Generated export files
├── 📁 logs/                     # Application logs
├── 📁 migrations/               # Database migrations
├── 📁 scripts/                  # Utility scripts
├── 📁 template/                 # Email templates
│
├── 📄 Makefile                  # Development commands
├── 📄 README.md                 # Project documentation
├── 📄 .gitignore               # Git ignore rules
├── 📄 docker-compose.yml        # Docker services
├── 📄 go.mod                   # Go dependencies
└── 📄 go.sum                   # Dependency checksums
```

## 🔧 **CÁC VẤN ĐỀ ĐÃ ĐƯỢC SỬA**

### ✅ **1. Package Naming Inconsistency**
- **Trước**: `package handler` được import như `controller`
- **Sau**: Đồng nhất `package controller` trong `internal/controller/`

### ✅ **2. Import Path Confusion**
- **Trước**: `internal/config` vs `internal/configs`
- **Sau**: Chỉ sử dụng `internal/configs` với function `Load()` compatibility

### ✅ **3. Typo trong Thư mục**
- **Trước**: `internal/dataacess/`
- **Sau**: `internal/dataaccess/`

### ✅ **4. Wire Configuration**
- **Trước**: Wire configuration rải rác và không consistent
- **Sau**: Clean wire setup với regenerated code

### ✅ **5. Build Errors**
- **Trước**: Syntax errors trong `worker/monitoring.go`
- **Sau**: Fixed tất cả compilation errors

## 🚀 **HƯỚNG DẪN SỬ DỤNG**

### **Quick Start**
```bash
# 1. Setup development environment
./scripts/setup_dev.sh

# 2. Start development
make dev

# 3. View API docs
open http://localhost:8080/swagger/index.html
```

### **Development Commands**
```bash
# Build project
make build

# Run server
make run

# Run tests
make test

# Generate Wire DI
make wire

# Generate Swagger docs
make swagger

# Database migrations
make migrate-up
make migrate-down

# Docker services
make docker-up
make docker-down

# Code quality
make fmt
make lint
make check
```

### **Environment Setup**
```bash
# Install development tools
./scripts/setup_dev.sh tools

# Start infrastructure only
./scripts/setup_dev.sh docker

# Check prerequisites
./scripts/setup_dev.sh check
```

## 📊 **CLEAN ARCHITECTURE LAYERS**

### **🎯 Presentation Layer**
```
internal/controller/ → HTTP request/response handling
internal/handler/http/ → Route configuration
```

### **💼 Domain Layer**
```
internal/services/ → Business logic
internal/models/ → Domain entities
```

### **🗃️ Data Layer**
```
internal/repositories/ → Data access interfaces
internal/dataaccess/ → Database/cache clients
```

### **🔧 Infrastructure Layer**
```
internal/configs/ → Configuration
internal/utils/ → Shared utilities
internal/wiring/ → Dependency injection
```

## 🔄 **DEPENDENCY FLOW**

```
Controller → Service → Repository → DataAccess
    ↓         ↓           ↓           ↓
   HTTP    Business    Data       Database
 Handling   Logic     Access      Client
```

## 📝 **NAMING CONVENTIONS**

### **Files**
- `snake_case.go` cho file names
- `PascalCase` cho types và functions
- `camelCase` cho variables

### **Packages**
- `lowercase` package names
- Descriptive và concise
- No underscores

### **Interfaces**
- Suffix với `Interface` nếu cần thiết
- Hoặc suffix với `-er` (e.g., `ServerRepository`)

## 🧪 **TESTING STRATEGY**

### **Unit Tests**
- `*_test.go` files in same package
- Test business logic in services
- Mock dependencies với interfaces

### **Integration Tests**
- Test complete workflows
- Use test database
- Test API endpoints

### **Coverage**
```bash
make test-coverage
open coverage.html
```

## 🔒 **SECURITY CONSIDERATIONS**

- ✅ JWT authentication implemented
- ✅ Input validation với Gin binding
- ✅ SQL injection prevention với GORM
- ✅ CORS configuration
- ✅ Environment variable security

## 📈 **PERFORMANCE OPTIMIZATIONS**

- ✅ Connection pooling (pgxpool)
- ✅ Redis caching
- ✅ Background job processing
- ✅ Elasticsearch for search
- ✅ Concurrent server monitoring

## 🔍 **MONITORING & OBSERVABILITY**

- ✅ Structured logging với Zap
- ✅ Health check endpoint
- ✅ Metrics collection
- ✅ Error tracking
- ✅ Performance monitoring

## 📚 **NEXT STEPS**

1. **Add Authentication Middleware**
2. **Implement Rate Limiting**
3. **Add Metrics Collection**
4. **Implement Circuit Breaker**
5. **Add API Versioning**
6. **Implement Graceful Shutdown**
7. **Add Configuration Validation**
8. **Implement Request Tracing**

---

**📞 Support**: Nếu có vấn đề gì, check logs hoặc run `make help` để xem available commands.
