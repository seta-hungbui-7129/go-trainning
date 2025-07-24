# SETA Training - Development Guide

## 🛠 Development Setup

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Git
- Your favorite IDE (VS Code, GoLand, etc.)

### IDE Setup (VS Code)
Recommended extensions:
```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-json",
    "bradlc.vscode-tailwindcss",
    "GraphQL.vscode-graphql"
  ]
}
```

### Environment Setup
```bash
# Clone repository
git clone <repository-url>
cd seta-training

# Install dependencies
go mod download

# Setup environment
cp .env.example .env
```

## 🏗 Architecture Patterns

### Clean Architecture Layers
1. **API Layer**: GraphQL/REST endpoints
2. **Service Layer**: Business logic
3. **Repository Layer**: Data access
4. **Model Layer**: Data structures

### Dependency Flow
```
API → Service → Repository → Database
```

### Adding New Features

#### 1. Create Model
```go
// internal/models/example.go
type Example struct {
    ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
    Name      string    `json:"name" gorm:"not null"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

#### 2. Create Repository
```go
// internal/repositories/example_repository.go
type ExampleRepository struct {
    db *gorm.DB
}

func NewExampleRepository(db *gorm.DB) *ExampleRepository {
    return &ExampleRepository{db: db}
}

func (r *ExampleRepository) Create(example *models.Example) error {
    return r.db.Create(example).Error
}
```

#### 3. Create Service
```go
// internal/services/example_service.go
type ExampleService struct {
    exampleRepo *repositories.ExampleRepository
}

func NewExampleService(exampleRepo *repositories.ExampleRepository) *ExampleService {
    return &ExampleService{exampleRepo: exampleRepo}
}

func (s *ExampleService) CreateExample(input *CreateExampleInput) (*models.Example, error) {
    // Business logic here
    example := &models.Example{Name: input.Name}
    return example, s.exampleRepo.Create(example)
}
```

#### 4. Create Handler
```go
// internal/handlers/example_handler.go
func (h *ExampleHandler) CreateExample(c *gin.Context) {
    var input CreateExampleInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    example, err := h.exampleService.CreateExample(&input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, example)
}
```

## 📊 GraphQL Development

### Schema Updates
1. Edit `api/graphql/schema.graphql`
2. Run code generation: `go run github.com/99designs/gqlgen@latest generate`
3. Implement resolvers in `api/graphql/resolvers/`

### Adding New Query
```graphql
# In schema.graphql
type Query {
  getExample(id: ID!): Example
}

type Example {
  id: ID!
  name: String!
  createdAt: String!
}
```

### Implementing Resolver
```go
// In resolvers/schema.resolvers.go
func (r *queryResolver) GetExample(ctx context.Context, id string) (*models.Example, error) {
    uuid, err := uuid.Parse(id)
    if err != nil {
        return nil, err
    }
    return r.ExampleService.GetByID(uuid)
}
```

## 🔧 Database Development

### Adding New Model
1. Create model in `internal/models/`
2. Add to migration in `internal/database/database.go`
3. Restart application to auto-migrate

### Migration Example
```go
// In database.go AutoMigrate call
err := d.DB.AutoMigrate(
    &models.User{},
    &models.Team{},
    &models.Example{}, // Add new model here
)
```

### Database Queries
Use GORM for database operations:
```go
// Simple queries
user := &models.User{}
db.Where("email = ?", email).First(user)

// Complex queries with joins
var teams []models.Team
db.Preload("Managers").Preload("Members").Find(&teams)

// Raw SQL when needed
db.Raw("SELECT * FROM users WHERE role = ?", "manager").Scan(&users)
```

## 🧪 Testing

### Unit Tests Structure
```go
// internal/services/user_service_test.go
func TestUserService_CreateUser(t *testing.T) {
    // Setup
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo, nil)
    
    // Test
    input := &CreateUserInput{
        Username: "test",
        Email: "test@example.com",
        Password: "password",
        Role: models.RoleManager,
    }
    
    user, err := service.CreateUser(input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "test", user.Username)
}
```

### Integration Tests
```go
func TestTeamAPI_CreateTeam(t *testing.T) {
    // Setup test database
    db := setupTestDB()
    defer cleanupTestDB(db)
    
    // Setup test server
    router := setupTestRouter(db)
    
    // Test request
    body := `{"teamName": "Test Team"}`
    req := httptest.NewRequest("POST", "/api/v1/teams", strings.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+testToken)
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/services/

# Run with verbose output
go test -v ./...
```

## 🔍 Debugging

### Debug Configuration
```go
// Enable debug mode
gin.SetMode(gin.DebugMode)

// Enable GORM debug logging
db.Debug()
```

### Logging Best Practices
```go
import "log"

// Use structured logging
log.Printf("User created: ID=%s, Email=%s", user.ID, user.Email)

// Error logging
if err != nil {
    log.Printf("Error creating user: %v", err)
    return err
}
```

### Common Debug Points
- Check JWT token validation
- Verify database connections
- Validate input data
- Check middleware execution order

## 📝 Code Style & Standards

### Go Conventions
- Use `gofmt` for formatting
- Follow effective Go guidelines
- Use meaningful variable names
- Add comments for exported functions

### Project Conventions
- Models in `internal/models/`
- Business logic in `internal/services/`
- Data access in `internal/repositories/`
- HTTP handlers in `internal/handlers/`

### Error Handling
```go
// Return errors, don't panic
func (s *Service) DoSomething() error {
    if err := s.repo.Save(); err != nil {
        return fmt.Errorf("failed to save: %w", err)
    }
    return nil
}

// Handle errors at appropriate level
func (h *Handler) HandleRequest(c *gin.Context) {
    if err := h.service.DoSomething(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Internal server error",
        })
        log.Printf("Service error: %v", err)
        return
    }
}
```

## 🚀 Development Workflow

### Daily Development
1. Pull latest changes: `git pull origin main`
2. Start database: `./scripts/start-db.sh`
3. Run application: `./scripts/run.sh`
4. Make changes and test
5. Run tests: `go test ./...`
6. Commit and push changes

### Feature Development
1. Create feature branch: `git checkout -b feature/new-feature`
2. Implement feature following architecture patterns
3. Add tests for new functionality
4. Update documentation if needed
5. Create pull request

### Code Review Checklist
- [ ] Follows project architecture patterns
- [ ] Includes appropriate tests
- [ ] Handles errors properly
- [ ] Updates documentation
- [ ] No hardcoded values
- [ ] Proper input validation
- [ ] Security considerations addressed

## 🔧 Useful Commands

```bash
# Development
go run cmd/server/main.go              # Run application
go build -o bin/server cmd/server/main.go  # Build binary
go mod tidy                            # Clean up dependencies

# GraphQL
go run github.com/99designs/gqlgen@latest generate  # Generate GraphQL code

# Database
./scripts/start-db.sh                 # Start database
docker exec -it seta_training_db psql -U postgres -d seta_training  # Connect to DB

# Testing
go test ./...                          # Run all tests
go test -cover ./...                   # Run with coverage
go test -v ./internal/services/        # Run specific package

# Code Quality
gofmt -w .                            # Format code
go vet ./...                          # Static analysis
golint ./...                          # Linting (if installed)
```

This development guide provides the foundation for contributing to and extending the SETA Training system.
