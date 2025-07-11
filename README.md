## Clean Architecture

```
├── cmd/                        # Entry points
│   ├── server/                 # Main server application
│   └── migrate/                # Database migration tool
├── internal/                   # Private application code
│   ├── app/                    # Application layer
│   ├── configs/                # Configuration layer
│   ├── delivery/               # Presentation layer
│   │   ├── http/               # HTTP delivery
│   │   │   ├── controllers/    # HTTP controllers
│   │   │   ├── presenters/     # Response presenters
│   │   ├── middleware/         # HTTP middleware
│   ├── domain/                 # Domain layer (Business entities & rules)
│   │   ├── entity/             # Core entities
│   │   ├── repository/         # Repository interfaces
│   │   ├── services/           # Domain services interfaces
│   │   ├── errors/             # Domain errors
│   │   ├── query/              # Query objects
│   │   ├── report/             # Report domain
│   │   ├── scope/              # Permission scopes
│   │   └── response.go         # Standard API response
│   ├── dto/                    # Data Transfer Objects
│   ├── usecases/               # Use cases layer (Business logic)
│   ├── infrastructure/         # Infrastructure layer
│   │   ├── database/           # Database implementation
│   │   ├── cache/              # Cache implementation
│   │   ├── search/             # Search implementation
│   │   ├── repository/         # Repository implementations
│   │   ├── services/           # Service implementations
│   │   ├── models/             # Database models
│   │   ├── validator/          # Custom validators
│   ├── jobs/                   # Background jobs
│   │   ├── scheduler/          # Job scheduler
│   │   ├── tasks/              # Task definitions
│   ├── utils/                  # Utility functions
│   └── wiring/                 # Dependency injection
├── configs/                    # Configuration files
├── migrations/                 # Database migrations
├── scripts/                    # Utility scripts
├── docs/                       # API documentation
├── tests/                      # Test files
│   └── unit/                   # Unit tests
│       └── services/           # Service tests
├── template/                   # Templates
├── logs/                       # Log files
├── exports/                    # Export files
├── logstash/                   # Logstash configuration
├── docker-compose.yml          # Docker compose configuration
├── Makefile                    # Build automation
├── go.mod                      # Go module dependencies
├── go.sum                      # Go module checksums
└── README.md                   # Project documentation
```

### How to run

**Start docker and server**
```bash
make run
```

### Config: `configs/config.dev.yaml`

```yaml
server:
  name: VCS-SMS
  env: development
  port: 8080

database:
  host: localhost
  port: 5432
  user: postgres
  password: password
  dbname: vcs_sms
  max_idle_conns: 10
  max_open_conns: 100

cache:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10

elasticsearch:
  url: http://localhost:9200

jwt:
  secret: your-jwt-secret-key
  expiration: 24h

log:
  level: info
```
## API Documentation

### Swagger UI
Swagger UI : `http://localhost:8080/swagger/index.html`

### Auth
JWT Bearer Token:
```
Authorization: Bearer <your-jwt-token>
```

### Endpoint API

#### Authentication
```
POST /api/v1/auth/login      
POST /api/v1/auth/register   
POST /api/v1/auth/refresh    
POST /api/v1/auth/logout     
```

#### Server Management
```
GET    /api/v1/servers           
POST   /api/v1/servers           
PUT    /api/v1/servers/{id}      
DELETE /api/v1/servers/{id}      
POST   /api/v1/servers/import    
GET    /api/v1/servers/export    
```

#### User Management
```
GET    /api/v1/users             
POST   /api/v1/users             
PUT    /api/v1/users/{id}        
DELETE /api/v1/users/{id}        
GET    /api/v1/users/profile     
PUT    /api/v1/users/profile     
```

#### Reports
```
POST /api/v1/reports/daily    
POST /api/v1/reports          
```

#### Jobs Monitoring
```
GET /api/v1/jobs         
GET /api/v1/jobs/status  
```