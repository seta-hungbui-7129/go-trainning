# SETA Training - Project Structure

## 📁 Complete Directory Structure

```
seta-training/
├── .env                          # Environment variables (local development)
├── .env.example                  # Environment variables template
├── .gitignore                    # Git ignore rules
├── README.md                     # Main project documentation
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── gqlgen.yml                    # GraphQL code generation config
│
├── api/                          # API layer
│   ├── graphql/                  # GraphQL API
│   │   ├── generated/            # Generated GraphQL code
│   │   ├── model/                # GraphQL models
│   │   ├── resolvers/            # GraphQL resolvers
│   │   ├── scalars/              # Custom scalar types
│   │   └── schema.graphql        # GraphQL schema definition
│   └── rest/                     # REST API (future expansion)
│
├── bin/                          # Compiled binaries
│   └── server                    # Main server binary
│
├── cmd/                          # Application entry points
│   └── server/                   # Main server application
│       └── main.go               # Application entry point
│
├── docker/                       # Docker configuration
│   ├── docker-compose.yml        # Database services
│   └── init.sql                  # Database initialization
│
├── docs/                         # Documentation
│   └── PROJECT_STRUCTURE.md      # This file
│
├── internal/                     # Private application code
│   ├── config/                   # Configuration management
│   │   └── config.go             # Environment config loader
│   ├── database/                 # Database connection & migrations
│   │   └── database.go           # Database setup and migrations
│   ├── handlers/                 # HTTP request handlers
│   │   └── team_handler.go       # Team management endpoints
│   ├── middleware/               # HTTP middleware
│   │   └── auth.go               # Authentication middleware
│   ├── models/                   # Database models
│   │   ├── user.go               # User model and relationships
│   │   ├── team.go               # Team model and relationships
│   │   ├── folder.go             # Folder model and sharing
│   │   └── note.go               # Note model and sharing
│   ├── repositories/             # Data access layer
│   │   ├── user_repository.go    # User data operations
│   │   ├── team_repository.go    # Team data operations
│   │   ├── folder_repository.go  # Folder data operations
│   │   └── note_repository.go    # Note data operations
│   └── services/                 # Business logic layer
│       ├── user_service.go       # User business logic
│       ├── team_service.go       # Team business logic
│       ├── folder_service.go     # Folder business logic
│       └── note_service.go       # Note business logic
│
├── migrations/                   # Database migrations (future)
│
├── pkg/                          # Public/shared packages
│   ├── auth/                     # Authentication utilities
│   │   ├── jwt.go                # JWT token management
│   │   └── password.go           # Password hashing utilities
│   └── utils/                    # Shared utilities (future)
│
└── scripts/                      # Build and deployment scripts
    ├── build.sh                  # Build application
    ├── run.sh                    # Run application
    └── start-db.sh               # Start database services
```

## 🏗 Architecture Overview

### **Clean Architecture Layers**

1. **API Layer** (`api/`)
   - GraphQL schema and resolvers
   - REST endpoints (future)
   - Input validation and serialization

2. **Application Layer** (`cmd/`)
   - Main application entry point
   - Dependency injection setup
   - Server configuration

3. **Business Logic Layer** (`internal/services/`)
   - Core business rules
   - Use case implementations
   - Service orchestration

4. **Data Access Layer** (`internal/repositories/`)
   - Database operations
   - Query implementations
   - Data mapping

5. **Infrastructure Layer** (`internal/`)
   - Database connections
   - External service integrations
   - Configuration management

### **Key Design Patterns**

- **Repository Pattern**: Data access abstraction
- **Service Pattern**: Business logic encapsulation
- **Dependency Injection**: Loose coupling between layers
- **Middleware Pattern**: Cross-cutting concerns (auth, logging)
- **Clean Architecture**: Separation of concerns

## 📋 File Descriptions

### **Core Application Files**
- `cmd/server/main.go`: Application bootstrap and dependency wiring
- `internal/config/config.go`: Environment-based configuration
- `internal/database/database.go`: Database connection and auto-migration

### **Authentication & Security**
- `pkg/auth/jwt.go`: JWT token generation and validation
- `pkg/auth/password.go`: Password hashing with bcrypt
- `internal/middleware/auth.go`: Authentication and authorization middleware

### **Data Models**
- `internal/models/user.go`: User entity with roles and relationships
- `internal/models/team.go`: Team entity with manager/member relationships
- `internal/models/folder.go`: Folder entity with sharing capabilities
- `internal/models/note.go`: Note entity with sharing capabilities

### **Business Logic**
- `internal/services/user_service.go`: User management and authentication
- `internal/services/team_service.go`: Team creation and member management
- `internal/services/folder_service.go`: Folder CRUD and sharing
- `internal/services/note_service.go`: Note CRUD and sharing

### **API Implementation**
- `api/graphql/schema.graphql`: GraphQL schema definition
- `api/graphql/resolvers/`: GraphQL query and mutation resolvers
- `internal/handlers/team_handler.go`: REST API endpoints for teams

### **Development Tools**
- `scripts/`: Development and deployment automation
- `docker/`: Local development database setup
- `gqlgen.yml`: GraphQL code generation configuration

## 🔄 Data Flow

1. **Request** → API Layer (GraphQL/REST)
2. **Validation** → Middleware (Auth, RBAC)
3. **Business Logic** → Services Layer
4. **Data Access** → Repository Layer
5. **Database** → PostgreSQL with GORM
6. **Response** → JSON/GraphQL response

This structure follows Go best practices and provides a scalable foundation for the microservices-based system.
