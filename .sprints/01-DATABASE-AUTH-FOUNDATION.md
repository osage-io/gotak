# Sprint 1: Database Abstraction & Authentication Foundation

**Duration:** 2 weeks  
**Sprint Goal:** Implement embedded-first database layer and zero-trust authentication system

## Sprint Objectives

1. Create database abstraction layer supporting SQLite → PostgreSQL/TimescaleDB
2. Implement OIDC authentication with Vault integration
3. Build fallback authentication system
4. Add comprehensive audit logging
5. Create user and role management system

## User Stories

### Epic 1: Database Abstraction Layer

**As a** system administrator  
**I want** the server to run with embedded SQLite for demos and scale to PostgreSQL for production  
**So that** deployment complexity is minimized while supporting enterprise requirements

#### Story 1.1: Database Interface Design
- [ ] Create `pkg/database` package with common interface
- [ ] Define CRUD operations for all data types
- [ ] Design migration system supporting both backends
- [ ] Add connection pooling and health checks

#### Story 1.2: SQLite Implementation
- [ ] Implement SQLite backend with JSON columns for time-series simulation
- [ ] Add automatic schema creation and migrations
- [ ] Create indexes for performance optimization
- [ ] Add backup and restore capabilities

#### Story 1.3: PostgreSQL/TimescaleDB Implementation
- [ ] Implement PostgreSQL backend with proper table design
- [ ] Add TimescaleDB extension support for time-series data
- [ ] Create migration scripts from SQLite
- [ ] Add connection pooling and failover support

#### Story 1.4: Configuration-Driven Selection
- [ ] Auto-detect database backend based on configuration
- [ ] Add connection string validation
- [ ] Implement graceful fallback to SQLite
- [ ] Add database health endpoints

### Epic 2: Zero-Trust Authentication System

**As a** military operator  
**I want** secure authentication through Vault OIDC with TLS certificate fallback  
**So that** access is properly controlled and audited

#### Story 2.1: Vault OIDC Integration
- [ ] Implement Vault OIDC authentication flow
- [ ] Add JWT token validation and refresh
- [ ] Create user profile management
- [ ] Add group-based claims processing

#### Story 2.2: TLS Certificate Authentication
- [ ] Implement mTLS client certificate validation
- [ ] Add certificate-based user identification
- [ ] Create certificate management utilities
- [ ] Add certificate revocation checking

#### Story 2.3: Fallback Authentication
- [ ] Implement local user/password authentication
- [ ] Add password policy enforcement (DoD requirements)
- [ ] Create password reset functionality
- [ ] Add account lockout protection

#### Story 2.4: Session Management
- [ ] Implement secure session handling
- [ ] Add session timeout and renewal
- [ ] Create logout and session invalidation
- [ ] Add concurrent session limits

### Epic 3: Role-Based Access Control (RBAC)

**As a** mission commander  
**I want** fine-grained role-based permissions  
**So that** users only access appropriate functionality and data

#### Story 3.1: Role Definition System
- [ ] Create role hierarchy (Admin, Commander, Operator, Observer)
- [ ] Define permission system with granular controls
- [ ] Add group-based permissions
- [ ] Create role assignment interface

#### Story 3.2: Authorization Middleware
- [ ] Implement authorization middleware for all endpoints
- [ ] Add resource-level permissions (missions, groups, etc.)
- [ ] Create permission checking utilities
- [ ] Add authorization caching

#### Story 3.3: User Management
- [ ] Create user CRUD operations
- [ ] Add user profile management
- [ ] Implement group membership management
- [ ] Add user activity tracking

### Epic 4: Comprehensive Audit Logging

**As a** security administrator  
**I want** complete audit trails of all system activities  
**So that** DoD compliance requirements are met

#### Story 4.1: Audit Framework
- [ ] Create audit logging interface
- [ ] Define audit event types and schemas
- [ ] Add structured logging with correlation IDs
- [ ] Implement audit log retention policies

#### Story 4.2: Authentication Auditing
- [ ] Log all login/logout attempts with details
- [ ] Track failed authentication attempts
- [ ] Monitor session management events
- [ ] Add suspicious activity detection

#### Story 4.3: Authorization Auditing
- [ ] Log all permission checks and results
- [ ] Track role and permission changes
- [ ] Monitor data access patterns
- [ ] Add privilege escalation detection

#### Story 4.4: Data Classification
- [ ] Implement data classification labels (UNCLASSIFIED, RESTRICTED, etc.)
- [ ] Add automatic classification based on content
- [ ] Create classification enforcement rules
- [ ] Add declassification workflows

## Technical Tasks

### Database Schema Design

```sql
-- Users table (works in both SQLite and PostgreSQL)
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT, -- NULL for OIDC-only users
    first_name TEXT,
    last_name TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Roles table
CREATE TABLE roles (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    permissions TEXT, -- JSON array in SQLite, JSONB in PostgreSQL
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User roles junction
CREATE TABLE user_roles (
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    role_id TEXT REFERENCES roles(id) ON DELETE CASCADE,
    granted_by TEXT REFERENCES users(id),
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- Audit logs (time-series optimized)
CREATE TABLE audit_logs (
    id TEXT PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id TEXT REFERENCES users(id),
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    details TEXT, -- JSON
    classification TEXT DEFAULT 'UNCLASSIFIED'
);
```

### Configuration Structure

```yaml
# Database configuration
database:
  # Auto-detection: if external_url is set, use PostgreSQL, otherwise SQLite
  external_url: "" # postgresql://user:pass@host:port/db
  embedded_path: "./data/gotak.db"
  max_connections: 25
  connection_timeout: 30s

# Authentication configuration  
auth:
  # Primary: Vault OIDC
  vault:
    enabled: false
    address: ""
    oidc_path: "auth/oidc"
    role: "gotak-users"
  
  # Secondary: TLS certificates
  tls:
    enabled: false
    ca_file: ""
    require_cert: false
  
  # Fallback: Local authentication
  local:
    enabled: true
    password_policy:
      min_length: 12
      require_uppercase: true
      require_lowercase: true
      require_numbers: true
      require_symbols: true
  
  # Session settings
  session:
    timeout: 24h
    refresh_threshold: 1h
    max_concurrent: 5

# Audit configuration
audit:
  enabled: true
  retention_days: 2555 # 7 years for DoD compliance
  classification:
    auto_classify: true
    default_level: "UNCLASSIFIED"
```

### API Endpoints to Implement

```yaml
# Authentication endpoints
POST   /api/v1/auth/login           # Local authentication
POST   /api/v1/auth/oidc/callback   # OIDC callback
POST   /api/v1/auth/logout          # Logout
POST   /api/v1/auth/refresh         # Token refresh
GET    /api/v1/auth/profile         # User profile
PUT    /api/v1/auth/profile         # Update profile

# User management endpoints  
GET    /api/v1/users                # List users (Admin only)
POST   /api/v1/users                # Create user (Admin only)
GET    /api/v1/users/{id}           # Get user details
PUT    /api/v1/users/{id}           # Update user
DELETE /api/v1/users/{id}           # Delete user (Admin only)

# Role management endpoints
GET    /api/v1/roles                # List roles
POST   /api/v1/roles                # Create role (Admin only)
GET    /api/v1/roles/{id}           # Get role details
PUT    /api/v1/roles/{id}           # Update role (Admin only)
DELETE /api/v1/roles/{id}           # Delete role (Admin only)

# Assignment endpoints
POST   /api/v1/users/{id}/roles     # Assign role to user
DELETE /api/v1/users/{id}/roles/{role_id} # Remove role from user

# Audit endpoints
GET    /api/v1/audit/logs           # View audit logs (Admin only)
GET    /api/v1/audit/users/{id}     # User activity logs
POST   /api/v1/audit/search         # Search audit logs
```

## Acceptance Criteria

### Database Layer
- [ ] Server starts with SQLite when no external DB configured
- [ ] Server connects to PostgreSQL when connection string provided
- [ ] Migrations run automatically on startup
- [ ] Database health check endpoint returns status
- [ ] Connection pooling handles load appropriately

### Authentication
- [ ] OIDC login redirects to Vault and handles callback
- [ ] TLS certificate authentication works with valid client certs
- [ ] Local login enforces password policy
- [ ] Failed logins are rate-limited and logged
- [ ] Session tokens are properly validated and refreshed

### Authorization
- [ ] Role permissions are enforced on all endpoints
- [ ] Users can only access resources they have permission for
- [ ] Admin users can manage users and roles
- [ ] Permission changes take effect immediately

### Audit Logging
- [ ] All authentication events are logged with full context
- [ ] Authorization failures are logged and monitored
- [ ] Data access is logged with user and resource details
- [ ] Audit logs cannot be modified by non-admin users
- [ ] Logs include proper classification labels

### Performance & Reliability
- [ ] Database operations complete within 100ms for 95% of requests
- [ ] Authentication completes within 500ms
- [ ] System handles 1000 concurrent authentication requests
- [ ] Graceful degradation when external services unavailable

## Development Tasks

### Week 1: Database & Core Auth
- [ ] Create database abstraction layer
- [ ] Implement SQLite backend
- [ ] Add basic user/role schema
- [ ] Implement local authentication
- [ ] Add session management

### Week 2: External Services & Audit
- [ ] Add PostgreSQL backend support
- [ ] Implement Vault OIDC integration
- [ ] Add TLS certificate authentication
- [ ] Create audit logging framework
- [ ] Add comprehensive testing

## Testing Strategy

### Unit Tests
- [ ] Database interface implementations
- [ ] Authentication flows (all methods)
- [ ] Authorization middleware
- [ ] Audit logging functionality

### Integration Tests
- [ ] End-to-end authentication flows
- [ ] Database backend switching
- [ ] External service integration (Vault)
- [ ] Session management

### Security Tests
- [ ] SQL injection prevention
- [ ] Authentication bypass attempts
- [ ] Authorization escalation attempts
- [ ] Session hijacking prevention

## Dependencies

### External Services (Optional)
- Vault server for OIDC authentication
- PostgreSQL/TimescaleDB for production database

### Go Packages to Add
- `github.com/golang-jwt/jwt/v5` - JWT handling
- `modernc.org/sqlite` - Pure Go SQLite driver
- `github.com/lib/pq` - PostgreSQL driver
- `golang.org/x/crypto` - Password hashing
- `github.com/casbin/casbin/v2` - Authorization engine

---

## Sprint Retrospective Template

### What went well?
- 

### What could be improved?
- 

### Action items for next sprint:
- 

### Blockers encountered:
- 

### Technical debt created:
- 
