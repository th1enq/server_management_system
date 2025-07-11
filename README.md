## Clean Architechture

```
├── cmd/                     # Entry points
│   ├── server/             # Main server application
│   └── migrate/            # Database migration tool
├── internal/
│   ├── delivery/           # Presentation layer
│   │   ├── http/           # HTTP handlers & controllers
│   │   └── middleware/     # Authentication middleware
│   ├── usecases/           # Business logic layer
│   ├── domain/             # Entities & business rules
│   │   ├── entity/         # Core entities
│   │   ├── repository/     # Repository interfaces
│   │   └── services/       # Service interfaces
│   ├── infrastructure/     # External concerns
│   │   ├── database/       # Database implementation
│   │   ├── cache/          # Redis cache implementation
│   │   ├── search/         # Elasticsearch implementation
│   │   └── services/       # External services
│   ├── jobs/               # Background jobs
│   │   ├── scheduler/      # Job scheduler
│   │   └── tasks/          # Task definitions
│   └── wiring/             # Dependency injection
├── configs/                # Configuration files
├── migrations/             # Database migrations
├── scripts/                # Utility scripts
└── docs/                   # API documentation
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