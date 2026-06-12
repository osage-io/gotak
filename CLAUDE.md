# CLAUDE.md

This file provides guidance to Claude (Anthropic's AI assistant) when working with code in this repository.

## Spec-Driven Development Workflow

**GoTAK follows a spec-driven development approach.** All work starts with a GitHub issue.

When you say **"Open a ticket about: X"**, Claude will:
1. **Ask clarifying questions** to understand requirements
2. **Create a GitHub issue** with clear acceptance criteria
3. **Create a feature branch** from main
4. **Implement the solution** following the spec
5. **Commit with issue reference**: `git commit -m "feat: description (closes #N)"`
6. **Push and create PR** referencing the issue
7. **Merge after approval** - issue auto-closes

See **[SPEC_DRIVEN_DEVELOPMENT.md](SPEC_DRIVEN_DEVELOPMENT.md)** for the complete workflow, commands, and examples.


## Project Overview

GoTAK is a high-performance, modern implementation of a TAK (Team Awareness Kit) server written in Go. It provides situational awareness and real-time coordination capabilities for military, first responders, and emergency management teams using the Cursor on Target (CoT) protocol.

## Key Technologies

- **Language**: Go 1.21+ (Note: go.mod shows 1.23, but CI uses 1.21 - standardization needed)
- **Web Framework**: Gin (HTTP), Gorilla Mux & WebSocket
- **Database**: PostgreSQL 15 with sqlx
- **Caching**: Redis 7
- **Messaging**: NATS
- **Authentication**: JWT (v4 and v5 - consolidation needed), WebAuthn, MFA
- **Security**: Vault (transit encryption, secrets), TLS/mTLS
- **Orchestration**: HashiCorp Nomad, Consul Connect, Boundary
- **Frontend**: React (in web/ directory)
- **Testing**: testify, go-sqlmock, integration tests with Docker Compose

## Known Issues (Top 5)

1. **Go Version Inconsistency**: go.mod (1.23) vs CI (1.21) vs README (1.21+)
2. **Duplicate JWT Dependencies**: Both jwt/v4 and jwt/v5 in go.mod
3. **Missing LICENSE File**: Referenced but not present
4. **Fragmented Migrations**: Two migration directories (internal/database/migrations/ and migrations/)
5. **Security Credentials**: Hardcoded dev tokens in CI, multiple .env files

## Architecture

### Directory Structure
```
gotak/
├── cmd/                    # Application entry points
│   ├── gotak-server/       # Main TAK server
│   └── gotak-client/       # Test client
├── internal/               # Private application code
│   ├── auth/              # Authentication & authorization (JWT, MFA, WebAuthn)
│   ├── chat/              # Chat messaging system
│   ├── database/          # Database layer & migrations
│   ├── events/            # Event publishing (NATS)
│   ├── handlers/          # HTTP/WebSocket handlers
│   ├── middleware/        # Auth & logging middleware
│   ├── mission/           # Mission management (objectives, tasks, timeline)
│   └── server/            # Core server implementation
├── pkg/                   # Public library code
│   ├── config/            # Configuration management
│   ├── cot/               # CoT message handling
│   ├── rbac/              # Role-based access control
│   └── tak/               # TAK protocol utilities
├── web/                   # React tactical web interface
├── config/                # YAML configuration files
├── migrations/            # SQL migrations (consolidation needed)
├── nomad/                 # Nomad job specifications
├── hashistack/            # Local dev HashiStack (Consul/Vault/Nomad)
├── scripts/               # Build & deployment scripts
├── tests/                 # Integration & E2E tests
└── docs/                  # Documentation
```

### Core Components

1. **CoT Protocol Engine** (`pkg/cot/`): XML message parsing, validation, generation
2. **Server Core** (`internal/server/`): TCP/UDP/TLS listeners, message broadcasting
3. **Authentication** (`internal/auth/`): JWT tokens, password policies, MFA, WebAuthn
4. **Mission Management** (`internal/mission/`): Objectives, tasks, timeline tracking
5. **Chat System** (`internal/chat/`): Real-time messaging with validation
6. **RBAC** (`pkg/rbac/`): Role-based access control interfaces
7. **Web Interface** (`web/`): React tactical UI with keyboard shortcuts

### Message Flow

1. Clients connect via TCP (8087), UDP (8087), or TLS (8089)
2. CoT XML messages parsed and validated
3. Authentication via JWT or client certificates
4. Messages processed by type (position, chat, system, emergency)
5. Events published to NATS
6. Messages encrypted via Vault transit engine
7. Broadcast to appropriate clients
8. WebSocket updates to web UI

## Development Commands

### Environment Setup
```bash
make dev-up              # Start PostgreSQL, Redis, NATS, Vault, Jaeger
make dev-down            # Stop all services
make dev                 # Start with hot reload (Air)
make dev-web             # Start full stack with web UI
```

### Building
```bash
make build               # Build server binary
make build-all           # Build for all platforms
make docker-build        # Build Docker image
make docker-build-hub    # Build for DockerHub
```

### Testing
```bash
make test                # Run unit tests
make test-coverage       # Generate coverage report
make test-integration    # Run integration tests (requires dev-up)
make lint                # Run golangci-lint
make security            # Run gosec & trivy scans
```

### HashiStack (Local Dev)
```bash
make hashi-up            # Start Consul, Vault, Nomad
make hashi-status        # Check health
make nomad-deploy        # Deploy GoTAK to Nomad
make nomad-stop          # Stop Nomad job
make hashi-down          # Stop everything
```

### Code Quality
```bash
make fmt                 # Format Go code
make precommit-install   # Install pre-commit hooks
make precommit-run       # Run hooks on all files
```

## Configuration

### Default Ports
- **8087**: TCP/UDP for TAK clients
- **8089**: TLS for secure TAK clients
- **8080**: Web interface
- **5432**: PostgreSQL (dev: gotak/gotak_dev)
- **5433**: PostgreSQL Test (gotak/gotak_test)
- **6379**: Redis (password: dev_redis_pass)
- **4222**: NATS (monitoring: 8222)
- **8200**: Vault (dev token: dev-token)
- **16686**: Jaeger UI

### Configuration Files
- `config/server.yaml`: Main server config
- `config/development.yaml`: Dev overrides
- `config/production.yaml`: Production settings
- `config/mfa.yaml`: MFA configuration
- `config/cert_auth.yaml`: Certificate authentication

## Coding Standards

### Go Style
- Follow Effective Go guidelines
- Use `gofmt`, `goimports`, `go vet`
- Pass `golangci-lint` checks
- Maintain >80% test coverage
- Document all public APIs

### Error Handling
```go
// ✅ Good: Wrap errors with context
func processMessage(msg []byte) error {
    event, err := cot.ParseCoT(msg)
    if err != nil {
        return fmt.Errorf("failed to parse CoT: %w", err)
    }
    return nil
}

// ❌ Bad: Ignore or lose error context
func processMessage(msg []byte) {
    event, _ := cot.Parse(msg)  // Don't ignore errors
}
```

### Testing Pattern
```go
func TestFunction(t *testing.T) {
    testCases := []struct {
        name     string
        input    interface{}
        expected interface{}
        wantErr  bool
    }{
        {name: "valid case", input: x, expected: y, wantErr: false},
        {name: "error case", input: z, expected: nil, wantErr: true},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := Function(tc.input)
            if tc.wantErr {
                assert.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tc.expected, result)
        })
    }
}
```

## Security Considerations

### Authentication
- JWT tokens with configurable expiration
- Password policies (length, complexity, history)
- MFA support (TOTP, WebAuthn)
- Client certificate authentication for TLS
- Session management with Redis

### Encryption
- TLS/mTLS for all production connections
- Vault transit engine for message encryption
- Secure credential storage in Vault
- No secrets in code or git history

### Best Practices
- Input validation on all endpoints
- SQL injection prevention (parameterized queries)
- XSS protection in web UI
- CSRF tokens for state-changing operations
- Rate limiting on authentication endpoints
- Audit logging for security events

## Common Tasks

### Adding a New CoT Message Type
1. Define constants in `pkg/cot/cot.go`
2. Add XML struct fields if needed
3. Create helper function `IsTypeXxx()`
4. Add handler in `internal/server/server.go`
5. Add tests in `pkg/cot/cot_test.go`

### Adding a New API Endpoint
1. Define handler in `internal/handlers/`
2. Add route in router setup
3. Add authentication middleware if needed
4. Implement request/response models
5. Add validation
6. Write unit and integration tests
7. Update API documentation

### Database Migration
1. Create migration files in `internal/database/migrations/`
2. Use format: `NNN_description.up.sql` and `NNN_description.down.sql`
3. Test migration up and down
4. Update schema documentation

### Adding a New Service
1. Define interface in appropriate package
2. Implement service with dependency injection
3. Add configuration options
4. Register with Consul (if using service mesh)
5. Add health checks
6. Write comprehensive tests

## Deployment

### Docker Compose (Development)
```bash
docker-compose -f docker-compose.dev.yml up -d
```

### Docker Compose (Production)
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Nomad (Standalone)
```bash
nomad job run nomad/deployments/standalone/gotak-complete.nomad.hcl
```

### Nomad (Consul Connect)
```bash
nomad job run nomad/deployments/consul/gotak-server.nomad.hcl
```

## Troubleshooting

### Common Issues
1. **Port conflicts**: Check if ports 8087, 8089, 8080 are available
2. **Database connection**: Ensure PostgreSQL is running and credentials are correct
3. **Redis connection**: Verify Redis is accessible and password is correct
4. **Vault unsealed**: Dev Vault must be running and unsealed
5. **Migration errors**: Check migration files and database state

### Debug Mode
```bash
./bin/gotak-server -debug -config config/development.yaml
```

### Logs
- Server logs: stdout/stderr or configured log file
- Docker logs: `docker-compose logs -f gotak`
- Nomad logs: `nomad alloc logs <alloc-id>`
- Consul logs: Check Consul UI or `consul monitor`

## Web Interface

### Keyboard Shortcuts
- `Ctrl/Cmd + K`: Command palette
- `/`: Quick search
- `Ctrl + 1-9`: Navigate pages
- `Ctrl + E`: Emergency alert
- `Ctrl + I`: AI Intel Officer
- `Ctrl + L`: Alerts
- `Ctrl + G`: Settings

### Development
```bash
cd web
npm ci                   # Install dependencies
npm run dev              # Start dev server (Vite)
npm run build            # Production build
npm run lint             # ESLint
```

## Related Context Files

- **WARP.md**: Context for Warp terminal
- **BOB-PROMPTS.md**: Infrastructure generation prompts for IBM Bob
- **CONTRIBUTING.md**: Contribution guidelines
- **BUILD.md**: Build and deployment guide
- **DEPLOYMENT.md**: Deployment documentation
- **docs/ARCHITECTURE.md**: Detailed architecture
- **docs/SPRINT_PLAN.md**: Development roadmap

## Quick Reference

### Environment Variables
```bash
# Database
POSTGRES_URL=postgres://gotak:password@localhost:5432/gotak_dev?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=dev_redis_pass

# Vault
VAULT_ADDR=http://localhost:8200
VAULT_TOKEN=dev-token

# NATS
NATS_URL=nats://localhost:4222

# Server
SERVER_PORT=8080
TAK_TCP_PORT=8087
TAK_TLS_PORT=8089
```

### Git Workflow
```bash
# Feature branch
git checkout -b feature/description

# Commit with conventional commits
git commit -m "feat: add new feature"
git commit -m "fix: resolve bug"
git commit -m "docs: update documentation"

# Push and create PR
git push origin feature/description
```

## Notes for Claude

- Always check for the known issues before making changes
- Prefer `apply_diff` for targeted changes over `write_to_file`
- Read related files together (up to 5) for better context
- Run tests after code changes
- Update documentation when adding features
- Follow the existing code patterns and style
- Consider security implications of all changes
- Use the development environment (`make dev-up`) for testing

---

*This context file helps Claude understand the GoTAK project structure, conventions, and best practices for effective code assistance.*