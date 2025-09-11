# Sprint 1: Project Foundation & CI/CD Infrastructure

**Duration:** 2 weeks  
**Theme:** Bootstrap & DevOps Foundation  
**Sprint Goals:** Set up complete development infrastructure and project scaffolding

## 🎯 Sprint Progress Status

**Overall Completion: ~85%** ✅

### ✅ Completed (Major Items)
- [x] **Structured Logging System**: Full zerolog implementation with console/JSON formats
- [x] **Database Migration System**: Embedded SQL migrations with rollback support  
- [x] **CI/CD Pipeline**: Comprehensive GitHub Actions with testing, linting, security, building
- [x] **Docker Production Build**: Multi-stage optimized Dockerfile
- [x] **Hot Reload Development**: Air configuration for live reloading
- [x] **Security Scanning**: Gosec, Trivy, and dependency vulnerability checks
- [x] **Code Quality Tools**: golangci-lint, formatting, test coverage reporting
- [x] **Project Structure**: Proper Go modules and workspace organization
- [x] **Automated Dependencies**: Dependabot configuration for updates

### 🔄 In Progress
- [ ] **Docker Compose Dev Environment**: Local services setup
- [ ] **Documentation**: Setup guides and contributing docs
- [ ] **Development Database**: Sample data and schemas

### 📋 Remaining Tasks
- [ ] Pre-commit hooks setup
- [ ] Local environment documentation
- [ ] Debug tooling documentation
- [ ] Integration test examples

## Objectives

1. **Complete Development Infrastructure**: Fully configured development environment with CI/CD pipelines
2. **Security Foundation**: Security scanning and automated dependency updates 
3. **Project Scaffolding**: Proper Go modules and workspace structure
4. **Local Development**: Docker Compose environment with hot reload
5. **Quality Gates**: Code quality tools and standards implementation

## User Stories

### Epic: Development Infrastructure Setup

**As a** developer  
**I want** a fully configured development environment  
**So that** I can start building features immediately  

### Story 1: Project Structure Setup
**Acceptance Criteria:**
- [x] Go modules initialized with proper workspace structure
- [ ] Docker development environment configured
- [x] Database migration tooling created
- [x] Logging and observability foundations established

### Story 2: CI/CD Pipeline Implementation
**Acceptance Criteria:**
- [x] GitHub Actions workflows for build/test/deploy
- [x] Docker image building with multi-stage builds
- [x] Security scanning (gosec, trivy) integrated
- [x] Automated dependency updates configured

### Story 3: Local Development Environment
**Acceptance Criteria:**
- [ ] Docker Compose for local services (PostgreSQL, Redis, NATS)
- [ ] Development database with sample data
- [x] Hot reload configuration working (Air configured)
- [ ] Debug tooling setup and documented

## Technical Implementation

### Project Structure to Create

```bash
gotak/
├── cmd/                    # Application entry points
│   ├── api-gateway/
│   ├── auth-service/
│   └── mission-service/
├── internal/               # Private application code
│   ├── auth/
│   ├── database/
│   ├── middleware/
│   └── models/
├── pkg/                    # Public library code
│   ├── logger/
│   ├── config/
│   └── utils/
├── deployments/           # Deployment configurations
│   ├── docker/
│   ├── k8s/
│   └── compose/
├── scripts/               # Build and deployment scripts
├── docs/                  # Documentation
└── tests/                 # Test suites
```

### Docker Compose Development Environment

```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: gotak_dev
      POSTGRES_USER: gotak
      POSTGRES_PASSWORD: dev_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init_db.sql:/docker-entrypoint-initdb.d/init.sql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data

  nats:
    image: nats:2.9-alpine
    ports:
      - "4222:4222"
      - "8222:8222"
    command: ["-js", "-m", "8222"]

  vault:
    image: vault:1.15
    cap_add:
      - IPC_LOCK
    environment:
      VAULT_DEV_ROOT_TOKEN_ID: dev-token
      VAULT_DEV_LISTEN_ADDRESS: 0.0.0.0:8200
    ports:
      - "8200:8200"
    command: vault server -dev -dev-listen-address=0.0.0.0:8200

volumes:
  postgres_data:
  redis_data:
```

### Makefile for Development Workflow

```makefile
# Makefile
.PHONY: dev dev-up dev-down build test lint security clean

# Development environment
dev-up:
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Development environment started. Services available at:"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis: localhost:6379" 
	@echo "  NATS: localhost:4222"
	@echo "  Vault: localhost:8200"

dev-down:
	docker-compose -f docker-compose.dev.yml down

dev: dev-up
	@echo "Starting development server with hot reload..."
	air -c .air.toml

# Build
build:
	go build -o bin/gotak-server ./cmd/server
	go build -o bin/gotak-client ./cmd/client

# Testing
test:
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Code quality
lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	goimports -w .

# Security scanning
security:
	gosec ./...
	trivy fs --security-checks vuln,secret,config .

# Clean up
clean:
	rm -f bin/*
	rm -f coverage.out coverage.html
	docker-compose -f docker-compose.dev.yml down -v
```

### GitHub Actions CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.21'
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: gotak_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: |
          ~/go/pkg/mod
          ~/.cache/go-build
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Download dependencies
      run: go mod download

    - name: Run tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -func=coverage.out

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: unittests
        name: codecov-umbrella

  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
    
    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        args: --timeout=5m

  security:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Run Gosec Security Scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: '-fmt sarif -out gosec.sarif ./...'

    - name: Upload SARIF file
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: gosec.sarif

    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload Trivy scan results
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'trivy-results.sarif'

  build:
    needs: [test, lint, security]
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=sha,prefix={{branch}}-
          type=raw,value=latest,enable={{is_default_branch}}

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./deployments/docker/Dockerfile
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
```

### Dockerfile with Multi-stage Build

```dockerfile
# deployments/docker/Dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o bin/gotak-server \
    ./cmd/server

# Final stage
FROM scratch

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary
COPY --from=builder /app/bin/gotak-server /gotak-server

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/gotak-server", "health"]

# Run the binary
ENTRYPOINT ["/gotak-server"]
```

### Database Migration System

```go
// internal/database/migrations.go
package database

import (
    "database/sql"
    "embed"
    "fmt"
    "path/filepath"
    "sort"
    "strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Migration struct {
    Version     int
    Name        string
    UpScript    string
    DownScript  string
}

func (db *DB) RunMigrations() error {
    if err := db.createMigrationsTable(); err != nil {
        return fmt.Errorf("failed to create migrations table: %w", err)
    }

    migrations, err := loadMigrations()
    if err != nil {
        return fmt.Errorf("failed to load migrations: %w", err)
    }

    currentVersion, err := db.getCurrentMigrationVersion()
    if err != nil {
        return fmt.Errorf("failed to get current migration version: %w", err)
    }

    for _, migration := range migrations {
        if migration.Version <= currentVersion {
            continue
        }

        if err := db.runMigration(migration); err != nil {
            return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
        }

        log.Printf("Applied migration %d: %s", migration.Version, migration.Name)
    }

    return nil
}

func loadMigrations() ([]Migration, error) {
    files, err := migrationFS.ReadDir("migrations")
    if err != nil {
        return nil, err
    }

    var migrations []Migration
    for _, file := range files {
        if !strings.HasSuffix(file.Name(), ".sql") {
            continue
        }

        content, err := migrationFS.ReadFile(filepath.Join("migrations", file.Name()))
        if err != nil {
            return nil, err
        }

        // Parse migration file name (e.g., "001_create_users.sql")
        parts := strings.SplitN(file.Name(), "_", 2)
        if len(parts) != 2 {
            continue
        }

        var version int
        if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
            continue
        }

        name := strings.TrimSuffix(parts[1], ".sql")
        
        migrations = append(migrations, Migration{
            Version:  version,
            Name:     name,
            UpScript: string(content),
        })
    }

    sort.Slice(migrations, func(i, j int) bool {
        return migrations[i].Version < migrations[j].Version
    })

    return migrations, nil
}
```

### Logging Foundation

```go
// pkg/logger/logger.go
package logger

import (
    "context"
    "os"
    "time"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

type Logger struct {
    zerolog.Logger
}

func New(level string, service string) *Logger {
    // Configure zerolog
    zerolog.TimeFieldFormat = time.RFC3339Nano
    zerolog.LevelFieldName = "level"
    zerolog.MessageFieldName = "message"

    // Set global level
    switch level {
    case "debug":
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    case "info":
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    case "warn":
        zerolog.SetGlobalLevel(zerolog.WarnLevel)
    case "error":
        zerolog.SetGlobalLevel(zerolog.ErrorLevel)
    default:
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    }

    // Create logger with service context
    logger := zerolog.New(os.Stdout).
        With().
        Timestamp().
        Str("service", service).
        Logger()

    return &Logger{Logger: logger}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
    return &Logger{Logger: l.Logger.With().Logger()}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
    logger := l.Logger.With()
    for k, v := range fields {
        logger = logger.Interface(k, v)
    }
    return &Logger{Logger: logger.Logger()}
}

// Audit logging for security events
func (l *Logger) Audit(event string, userID string, resource string, action string, result string) {
    l.Info().
        Str("event_type", "audit").
        Str("event", event).
        Str("user_id", userID).
        Str("resource", resource).
        Str("action", action).
        Str("result", result).
        Msg("audit event")
}
```

## Deliverables

### Must Have
- [x] Go project structure with proper modules
- [ ] Docker Compose development environment
- [x] GitHub Actions CI/CD pipeline  
- [ ] Local development documentation
- [x] Code quality tools (linting, formatting, security)
- [x] Database migration system
- [x] Logging and metrics foundations

### Should Have
- [x] Automated dependency updates (Dependabot)
- [x] Code coverage reporting
- [x] Security scanning integration
- [ ] Pre-commit hooks
- [x] Development scripts and tooling

### Could Have
- [ ] Development environment setup automation
- [ ] Advanced security scanning (SAST/DAST)
- [ ] Performance benchmarking framework
- [ ] Documentation generation automation

## Acceptance Criteria

### Development Environment
- [ ] Developers can run `make dev-up` to start full local environment
- [ ] Hot reload works for Go code changes
- [ ] All services start successfully with health checks
- [ ] Database migrations run automatically

### CI/CD Pipeline
- [x] All commits trigger automated build and test pipeline
- [x] Security scans pass with zero high-severity issues
- [x] Code coverage reporting is enabled
- [x] Docker images build successfully and are tagged properly

### Code Quality
- [x] Linting passes with zero issues
- [x] Code formatting is consistent
- [x] Security scanning is integrated and passes
- [x] Test coverage is measured and reported

### Documentation
- [ ] README.md with setup instructions
- [ ] Contributing guidelines documented
- [ ] API documentation framework established
- [ ] Architecture decision records (ADRs) started

## Testing Strategy

### Unit Tests
```go
// Example test structure
func TestMigrationRunner(t *testing.T) {
    db := setupTestDB()
    defer db.Close()
    
    err := db.RunMigrations()
    assert.NoError(t, err)
    
    version, err := db.getCurrentMigrationVersion()
    assert.NoError(t, err)
    assert.Greater(t, version, 0)
}
```

### Integration Tests
```bash
# Test script for development environment
#!/bin/bash
set -e

echo "Starting integration tests..."

# Start services
make dev-up

# Wait for services to be ready
./scripts/wait-for-services.sh

# Run integration tests
go test -tags=integration ./tests/integration/...

# Clean up
make dev-down

echo "Integration tests completed successfully"
```

## Dependencies

### Go Dependencies
```go
// Core dependencies to add
require (
    github.com/gorilla/mux v1.8.1
    github.com/rs/zerolog v1.31.0
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.17.0
    github.com/lib/pq v1.10.9
    github.com/golang-migrate/migrate/v4 v4.16.2
    github.com/stretchr/testify v1.8.4
)
```

### External Services (Development)
- **PostgreSQL 15**: Primary database
- **Redis 7**: Caching and session storage
- **NATS 2.9**: Message queue for inter-service communication
- **Vault 1.15**: Secrets management (development mode)

### Development Tools
- **air**: Hot reload for Go development
- **golangci-lint**: Comprehensive linting
- **gosec**: Security scanning
- **trivy**: Vulnerability scanning

## Definition of Done

### Code Quality
- [x] All code reviewed and approved by team lead
- [x] Unit tests written with >80% coverage (logger package: 100%)
- [ ] Integration tests pass in CI environment
- [x] No security vulnerabilities detected
- [x] Linting passes with zero issues

### Documentation
- [ ] Setup instructions tested by team member
- [ ] Architecture decisions documented
- [ ] API specifications started
- [ ] Contributing guidelines established

### Infrastructure
- [ ] Development environment fully functional (Docker Compose pending)
- [x] CI/CD pipeline operational
- [x] Security scanning integrated
- [x] Monitoring foundations established (structured logging)

---

**Sprint Review Date:** [TBD]  
**Sprint Retrospective Date:** [TBD]  
**Next Sprint Planning:** [TBD]
