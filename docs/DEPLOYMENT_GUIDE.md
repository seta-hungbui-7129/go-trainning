# SETA Training - Deployment Guide

## 🚀 Quick Start (Local Development)

### Prerequisites
- Go 1.21 or higher
- Docker and Docker Compose
- Git

### 1. Clone Repository
```bash
git clone <repository-url>
cd seta-training
```

### 2. Environment Setup
```bash
# Copy environment template
cp .env.example .env

# Edit environment variables as needed
nano .env
```

### 3. Start Database
```bash
# Start PostgreSQL with Docker
./scripts/start-db.sh

# Wait for database to be ready (script handles this)
```

### 4. Run Application
```bash
# Option 1: Run directly
./scripts/run.sh

# Option 2: Build and run binary
./scripts/build.sh
./bin/server

# Option 3: Run with Go
go run cmd/server/main.go
```

### 5. Verify Installation
```bash
# Check health endpoint
curl http://localhost:8080/health

# Access GraphQL Playground
open http://localhost:8080/playground
```

## 🐳 Docker Deployment

### Database Only (Current Setup)
```bash
cd docker
docker-compose up -d postgres
```

### Full Application (Future)
Create `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/bin/server .
COPY --from=builder /app/.env .

CMD ["./server"]
```

## ☁️ Production Deployment

### Environment Variables
```bash
# Database Configuration
DB_HOST=your-postgres-host
DB_PORT=5432
DB_USER=your-db-user
DB_PASSWORD=your-secure-password
DB_NAME=seta_training_prod
DB_SSLMODE=require

# JWT Configuration
JWT_SECRET=your-super-secure-jwt-secret-key
JWT_EXPIRY_HOURS=24

# Server Configuration
SERVER_PORT=8080
GIN_MODE=release

# GraphQL Configuration
GRAPHQL_PLAYGROUND=false

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
```

### Database Setup
```sql
-- Create production database
CREATE DATABASE seta_training_prod;
CREATE USER seta_app WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE seta_training_prod TO seta_app;

-- Enable UUID extension
\c seta_training_prod;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

### Build for Production
```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server cmd/server/main.go

# Or with build flags
go build -ldflags="-w -s" -o bin/server cmd/server/main.go
```

## 🔧 Configuration Options

### Environment Variables Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | Database host |
| `DB_PORT` | 5432 | Database port |
| `DB_USER` | postgres | Database username |
| `DB_PASSWORD` | password | Database password |
| `DB_NAME` | seta_training | Database name |
| `DB_SSLMODE` | disable | SSL mode (disable/require) |
| `JWT_SECRET` | default-secret | JWT signing secret |
| `JWT_EXPIRY_HOURS` | 24 | Token expiry time |
| `SERVER_PORT` | 8080 | Server port |
| `GIN_MODE` | debug | Gin mode (debug/release) |
| `GRAPHQL_PLAYGROUND` | true | Enable GraphQL playground |
| `LOG_LEVEL` | info | Log level |
| `LOG_FORMAT` | json | Log format |

### Database Migration
The application automatically runs migrations on startup. For manual migration:

```bash
# The migrations are handled by GORM AutoMigrate
# No separate migration files needed currently
```

## 📊 Monitoring & Logging

### Health Checks
```bash
# Application health
curl http://localhost:8080/health

# Database connectivity is checked in health endpoint
```

### Logging
The application uses structured logging. In production:
- Set `LOG_FORMAT=json` for structured logs
- Set `LOG_LEVEL=info` or `warn` for production
- Set `GIN_MODE=release` to reduce verbose logging

### Metrics (Future)
Planned integration with:
- Prometheus for metrics collection
- Grafana for visualization
- Loki for log aggregation

## 🔒 Security Considerations

### Production Security Checklist
- [ ] Use strong JWT secret (minimum 32 characters)
- [ ] Enable SSL/TLS for database connections (`DB_SSLMODE=require`)
- [ ] Disable GraphQL playground in production (`GRAPHQL_PLAYGROUND=false`)
- [ ] Use environment variables for secrets (never commit to code)
- [ ] Set up proper firewall rules
- [ ] Use HTTPS in production
- [ ] Implement rate limiting (future enhancement)
- [ ] Set up proper CORS policies (future enhancement)

### Database Security
```bash
# Create dedicated database user with limited privileges
CREATE USER seta_app WITH PASSWORD 'secure_random_password';
GRANT CONNECT ON DATABASE seta_training_prod TO seta_app;
GRANT USAGE ON SCHEMA public TO seta_app;
GRANT CREATE ON SCHEMA public TO seta_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO seta_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO seta_app;
```

## 🚨 Troubleshooting

### Common Issues

#### Database Connection Failed
```bash
# Check if database is running
docker ps | grep postgres

# Check database logs
docker logs seta_training_db

# Test connection manually
psql -h localhost -p 5432 -U postgres -d seta_training
```

#### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080

# Kill process if needed
kill -9 <PID>
```

#### Migration Errors
```bash
# Check database permissions
# Ensure user has CREATE privileges
# Check database exists and is accessible
```

### Debug Mode
```bash
# Run with debug logging
GIN_MODE=debug LOG_LEVEL=debug go run cmd/server/main.go
```

## 📈 Scaling Considerations

### Horizontal Scaling
- Application is stateless and can be scaled horizontally
- Use load balancer (nginx, HAProxy) for multiple instances
- Database connection pooling is configured

### Database Scaling
- Use read replicas for read-heavy workloads
- Consider connection pooling (PgBouncer)
- Monitor database performance and optimize queries

### Caching (Future)
- Redis for session storage
- Application-level caching for frequently accessed data
- CDN for static assets

This deployment guide provides a foundation for both development and production deployments of the SETA Training system.
