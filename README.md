# SETA Training - User, Team & Asset Management System

A microservices-based system built with Go for managing users, teams, and digital assets with role-based access control.

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- Docker and Docker Compose
- Git

### 1. Clone and Setup
```bash
git clone <repository-url>
cd seta-training
cp .env.example .env  # Edit as needed
```

### 2. Start Database
```bash
./scripts/start-db.sh
```

### 3. Run Application
```bash
./scripts/run.sh
```

The application will be available at:
- **GraphQL Playground**: http://localhost:8080/playground
- **Health Check**: http://localhost:8080/health
- **REST API**: http://localhost:8080/api/v1

## 📁 Project Structure

```
seta-training/
├── api/
│   ├── graphql/          # GraphQL schema, resolvers, generated code
│   └── rest/             # REST API handlers (future)
├── cmd/
│   └── server/           # Main application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── database/         # Database connection and migrations
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # Authentication and other middleware
│   ├── models/           # Database models
│   ├── repositories/     # Data access layer
│   └── services/         # Business logic layer
├── pkg/
│   ├── auth/             # JWT and password utilities
│   └── utils/            # Shared utilities
├── docker/               # Docker configuration
├── scripts/              # Build and deployment scripts
└── docs/                 # Documentation
    ├── PROJECT_STRUCTURE.md # Complete project structure
    ├── API_DOCUMENTATION.md # API endpoints and examples
    ├── DEPLOYMENT_GUIDE.md  # Deployment instructions
    └── DEVELOPMENT_GUIDE.md # Development guidelines
```

## 📚 Documentation

- **[Project Structure](docs/PROJECT_STRUCTURE.md)** - Complete directory structure and architecture overview
- **[API Documentation](docs/API_DOCUMENTATION.md)** - GraphQL and REST API endpoints with examples
- **[Deployment Guide](docs/DEPLOYMENT_GUIDE.md)** - Local development and production deployment instructions
- **[Development Guide](docs/DEVELOPMENT_GUIDE.md)** - Development workflow, patterns, and best practices

## 🎯 Features Implemented

### ✅ **Core Infrastructure**
- Microservices architecture with clean separation of concerns
- PostgreSQL database with automatic migrations
- JWT authentication with proper token validation
- Role-based access control (RBAC) for managers and members
- Docker setup for easy database deployment

### ✅ **GraphQL Service (User Management)**
- User creation with username, email, password, and role
- User authentication (login/logout) with JWT tokens
- User listing (fetchUsers query)
- Password hashing with bcrypt
- Email and username uniqueness validation

### ✅ **REST API (Team Management)**
- Team creation by managers
- Team member management (add/remove members)
- Team manager management (add/remove managers)
- Proper authorization - only managers can manage teams
- Automatic creator assignment as team manager

### ✅ **Database Models & Relationships**
- User model with roles and authentication
- Team model with many-to-many relationships
- Folder and Note models with sharing capabilities (ready for implementation)
- Proper foreign key constraints and indexes

### ✅ **Security & Middleware**
- JWT middleware for authentication
- Role-based middleware for authorization
- Password encryption and validation
- Input validation and error handling

## 🚀 API Examples

### GraphQL (User Management)

#### Create User
```graphql
mutation {
  createUser(input: {
    username: "manager1"
    email: "manager@example.com"
    password: "password123"
    role: manager
  }) {
    id username email role
  }
}
```

#### Login
```graphql
mutation {
  login(input: {
    email: "manager@example.com"
    password: "password123"
  }) {
    user { id username email role }
    token
  }
}
```

### REST API (Team Management)

#### Create Team
```bash
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "teamName": "Development Team",
    "managers": [],
    "members": []
  }'
```

## 🔧 Development Commands

```bash
# Start database
./scripts/start-db.sh

# Run application
./scripts/run.sh

# Build application
./scripts/build.sh

# Run tests
go test ./...

# Generate GraphQL code
go run github.com/99designs/gqlgen@latest generate
```

## 🏗 Architecture Highlights

1. **Clean Architecture**: Separated into layers (handlers, services, repositories, models)
2. **Dependency Injection**: Proper service initialization and dependency management
3. **Error Handling**: Comprehensive error handling with appropriate HTTP status codes
4. **Database Migrations**: Automatic schema creation and updates
5. **Configuration Management**: Environment-based configuration with defaults
6. **Logging**: Structured logging with GORM query logging
7. **Testing Ready**: Structure supports easy unit and integration testing

## 🔄 Future Enhancements

The foundation is solid and ready for:
- **Asset Management**: Folder and note CRUD operations (models already created)
- **Sharing System**: Implement folder/note sharing with access control
- **Manager Asset Viewing**: Team asset visibility for managers
- **Unit Tests**: Comprehensive test suite
- **API Documentation**: Swagger/OpenAPI documentation
- **Monitoring**: Loki + Grafana + Promtail integration

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [Development Guide](docs/DEVELOPMENT_GUIDE.md) for detailed development instructions.

## 📄 License

This project is part of SETA golang/nodejs training program.
# 🏗 Training Exercise: User, Team & Asset Management

## 🎯 Objective

Build a microservices-based system to manage users, teams, and digital assets—avoiding a monolithic design:

- Users can have roles: **manager** or **member**.
- Managers can create teams, add/remove members or other managers.
- Users can manage and share digital assets (folders & notes) with access control.

---

## ⚙ Proposed Microservice Architecture

- ✅ **GraphQL service**: For user management: create user, login, logout, fetch users, assign roles.
- ✅ **REST API**: For team management & asset management (folders, notes, sharing).

---

## 🧩 Functional Requirements

### 🔹 Auth & User Management Service (GraphQL)

- Create user:
  - `userId` (auto-generated)
  - `username`
  - `email` (unique)
  - `role`: "manager" or "member"
- Authentication:
  - Login, logout (JWT or session-based)
- User listing & query:
  - `fetchUsers` to get list of users
- Role assignment:
  - Manager: can create teams, manage users in teams
  - Member: can only be added to teams, no team management

---

### 🔹 Team Management Service (REST)

- Managers can:
  - Create teams
  - Add/remove members
  - Add/remove other managers (only main manager can do this)

Each team:

- `teamId`
- `teamName`
- `managers` (list)
- `members` (list)

---

### 🔹 Asset Management & Sharing (REST API)

- **Folders**: owned by users, contain notes
- **Notes**: belong to folders, have content
- Users can:
  - Share folders or individual notes with other users (read or write access)
  - Revoke access at any time
- When sharing a folder → all notes inside are also shared

**Managers**:

- Can view (read-only) all assets their team members have or can access
- Cannot edit unless explicitly shared with write access

---

## 🔑 Key Rules & Permissions

- Only authenticated users can use APIs.
- Managers can only manage users within their own teams.
- Members cannot create/manage teams.
- Only asset owners can manage sharing.

---

## 🛠 API Endpoints

### 📌 GraphQL: User Management

| Query/Mutation                      | Description             |
| ----------------------------------- | ----------------------- |
| `createUser(username, email, role)` | Create a new user       |
| `login(email, password)`            | Login and receive token |
| `logout()`                          | Logout current user     |
| `fetchUsers()`                      | List all users          |

---

### 📌 REST API: Team Management

| Method | Path                                 | Description        |
| ------ | ------------------------------------ | ------------------ |
| POST   | /teams                               | Create a team      |
| POST   | /teams/{teamId}/members              | Add member to team |
| DELETE | /teams/{teamId}/members/{memberId}   | Remove member      |
| POST   | /teams/{teamId}/managers             | Add manager        |
| DELETE | /teams/{teamId}/managers/{managerId} | Remove manager     |

#### ✅ Create team – request body:

```json
{
  "teamName": "string",
  "managers": [{"managerId": "string", "managerName": "string"}],
  "members": [{"memberId": "string", "memberName": "string"}]
}
```

---

### 📌 REST API: Asset Management

#### 🔹 Folder Management

| Method | Path                | Description                    |
| ------ | ------------------- | ------------------------------ |
| POST   | /folders            | Create new folder              |
| GET    | /folders/\:folderId | Get folder details             |
| PUT    | /folders/\:folderId | Update folder (name, metadata) |
| DELETE | /folders/\:folderId | Delete folder and its notes    |

#### 🔹 Note Management

| Method | Path                      | Description               |
| ------ | ------------------------- | ------------------------- |
| POST   | /folders/\:folderId/notes | Create note inside folder |
| GET    | /notes/\:noteId           | View note                 |
| PUT    | /notes/\:noteId           | Update note               |
| DELETE | /notes/\:noteId           | Delete note               |

#### 🔹 Sharing API

| Method | Path                               | Description                         |
| ------ | ---------------------------------- | ----------------------------------- |
| POST   | /folders/\:folderId/share          | Share folder with user (read/write) |
| DELETE | /folders/\:folderId/share/\:userId | Revoke folder sharing               |
| POST   | /notes/\:noteId/share              | Share single note                   |
| DELETE | /notes/\:noteId/share/\:userId     | Revoke note sharing                 |

#### 🔹 Manager-only APIs

| Method | Path                   | Description                                         |
| ------ | ---------------------- | --------------------------------------------------- |
| GET    | /teams/\:teamId/assets | View all assets that team members own or can access |
| GET    | /users/\:userId/assets | View all assets owned by or shared with user        |

---

## 🧩 Database Design Suggestion (PostgreSQL)

- Users: `userId`, `username`, `email`, `role`, `passwordHash`
- Teams: `teamId`, `teamName`
- team\_members, team\_managers: mapping tables
- Folders: `folderId`, `name`, `ownerId`
- Notes: `noteId`, `title`, `body`, `folderId`, `ownerId`
- folder\_shares, note\_shares: `userId`, `access` ("read" or "write")

---

## ✅ Development Requirements

- **Authentication**
  - Use JWT for authentication.
  - Validate and decode tokens, including expiration checks and claims verification—not limited to extracting user ID.

- **Authorization (RBAC)**
  - Enforce role validation before permitting team creation or manager assignment.
  - Restrict actions based on defined user roles.

- **Error Handling**
  - Handle duplicate email registration attempts.
  - Validate roles and flag unauthorized operations.
  - Return appropriate HTTP status codes and error messages.

- **Data Modeling**
  - Define models for the following entities:
    - `User`
    - `Team`
    - `Folder`
    - `Note`

- **Technology Stack**
  - Use the Go framework (Gin + GORM).
  - Apply clean architecture and separation of concerns for maintainability.

- **Logging and Monitoring**
  - Implement centralized logging.
  - Suggested stack: Loki + Grafana + Promtail (alternatives acceptable if justified).
  - Ensure visibility into API performance and system 

---
