# Sprint 8: Persistence Layer & Audit Logging

**Duration:** 2 weeks  
**Theme:** Enterprise Data Management & Compliance  
**Sprint Goals:** Implement production-grade data persistence and comprehensive audit logging

## Objectives

1. **Storage Abstraction**: Database-agnostic storage interface with PostgreSQL implementation
2. **Data Migration**: Schema migration system and deployment tooling
3. **Audit Logging**: Comprehensive audit trail for compliance and security
4. **Admin APIs**: Management endpoints for system administration
5. **Performance Optimization**: Database tuning for production scale

## User Stories

### Epic: Enterprise Data Management Platform

**US-8.1: Production Database Layer**
```
As a system administrator
I want a robust database layer that can scale to production workloads
So that the system can handle thousands of users and millions of messages
```

**Acceptance Criteria:**
- Storage abstraction layer supporting multiple database backends
- PostgreSQL implementation with connection pooling and optimization
- Database migration system with rollback capabilities
- Performance monitoring and query optimization
- Backup and restore procedures

**US-8.2: Comprehensive Audit Logging**
```
As a compliance officer
I want detailed audit logs of all system activities
So that I can meet regulatory requirements and investigate security incidents
```

**Acceptance Criteria:**
- Audit trail for authentication, authorization, and data access
- Structured logging with configurable retention policies
- Search and filtering capabilities for audit logs
- Export functionality for compliance reporting
- Real-time audit monitoring and alerting

**US-8.3: System Administration APIs**
```
As a system administrator
I want comprehensive admin APIs for system management
So that I can monitor, configure, and maintain the system effectively
```

**Acceptance Criteria:**
- User management endpoints (create, update, deactivate)
- System configuration management
- Health check and status monitoring APIs
- Performance metrics and analytics endpoints
- Bulk operations and data import/export

**US-8.4: Data Archival and Retention**
```
As a data management officer
I want automated data archival and retention policies
So that the system maintains performance while preserving historical data
```

**Acceptance Criteria:**
- Configurable data retention policies
- Automated archival of old messages and events
- Data compression and storage optimization
- Archive search and retrieval capabilities
- Compliance with data protection regulations

## Technical Implementation

### Storage Abstraction Layer

**Storage Interface**
```go
// pkg/storage/interface.go
package storage

import (
    "context"
    "time"
    
    "github.com/google/uuid"
)

type Storage interface {
    // User management
    UserStorage
    
    // Mission and task management
    MissionStorage
    
    // Communication and messaging
    MessageStorage
    
    // Position and entity tracking
    PositionStorage
    
    // Audit and compliance
    AuditStorage
    
    // System management
    SystemStorage
    
    // Health and monitoring
    Health(ctx context.Context) error
    Close() error
}

type UserStorage interface {
    CreateUser(ctx context.Context, user *User) error
    GetUser(ctx context.Context, id uuid.UUID) (*User, error)
    GetUserByUsername(ctx context.Context, username string) (*User, error)
    UpdateUser(ctx context.Context, id uuid.UUID, updates UserUpdates) error
    ListUsers(ctx context.Context, filter UserFilter) ([]*User, int, error)
    DeleteUser(ctx context.Context, id uuid.UUID) error
    
    // Authentication
    CreateSession(ctx context.Context, session *Session) error
    GetSession(ctx context.Context, token string) (*Session, error)
    RevokeSession(ctx context.Context, token string) error
    CleanupExpiredSessions(ctx context.Context) error
}

type MissionStorage interface {
    CreateMission(ctx context.Context, mission *Mission) error
    GetMission(ctx context.Context, id uuid.UUID) (*Mission, error)
    UpdateMission(ctx context.Context, id uuid.UUID, updates MissionUpdates) error
    ListMissions(ctx context.Context, filter MissionFilter) ([]*Mission, int, error)
    DeleteMission(ctx context.Context, id uuid.UUID) error
    
    // Tasks
    CreateTask(ctx context.Context, task *Task) error
    GetTasksByMission(ctx context.Context, missionID uuid.UUID) ([]*Task, error)
    UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status TaskStatus) error
}

type MessageStorage interface {
    CreateMessage(ctx context.Context, msg *Message) error
    GetMessages(ctx context.Context, filter MessageFilter) ([]*Message, int, error)
    GetMessagesByRoom(ctx context.Context, roomID uuid.UUID, pagination Pagination) ([]*Message, error)
    UpdateMessage(ctx context.Context, id uuid.UUID, updates MessageUpdates) error
    DeleteMessage(ctx context.Context, id uuid.UUID) error
    
    // Chat rooms
    CreateRoom(ctx context.Context, room *Room) error
    GetRoom(ctx context.Context, id uuid.UUID) (*Room, error)
    ListRooms(ctx context.Context, filter RoomFilter) ([]*Room, error)
}

type AuditStorage interface {
    CreateAuditEvent(ctx context.Context, event *AuditEvent) error
    GetAuditEvents(ctx context.Context, filter AuditFilter) ([]*AuditEvent, int, error)
    SearchAuditEvents(ctx context.Context, query AuditQuery) ([]*AuditEvent, error)
    ArchiveOldEvents(ctx context.Context, olderThan time.Time) (int, error)
}
```

**PostgreSQL Implementation**
```go
// pkg/storage/postgres/postgres.go
package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    "github.com/lib/pq"
    _ "github.com/lib/pq"
    
    "github.com/dfedick/gotak/pkg/storage"
)

type PostgresStorage struct {
    db     *sql.DB
    config *Config
    logger Logger
}

type Config struct {
    Host            string        `yaml:"host"`
    Port            int           `yaml:"port"`
    Database        string        `yaml:"database"`
    Username        string        `yaml:"username"`
    Password        string        `yaml:"password"`
    SSLMode         string        `yaml:"ssl_mode"`
    
    // Connection pool settings
    MaxOpenConns    int           `yaml:"max_open_conns"`
    MaxIdleConns    int           `yaml:"max_idle_conns"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
    ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
    
    // Performance settings
    StatementTimeout time.Duration `yaml:"statement_timeout"`
    QueryTimeout     time.Duration `yaml:"query_timeout"`
}

func NewPostgresStorage(config *Config, logger Logger) (*PostgresStorage, error) {
    connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        config.Host, config.Port, config.Username, config.Password,
        config.Database, config.SSLMode)
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(config.MaxOpenConns)
    db.SetMaxIdleConns(config.MaxIdleConns)
    db.SetConnMaxLifetime(config.ConnMaxLifetime)
    db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
    
    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    storage := &PostgresStorage{
        db:     db,
        config: config,
        logger: logger,
    }
    
    return storage, nil
}

func (ps *PostgresStorage) CreateUser(ctx context.Context, user *storage.User) error {
    query := `
        INSERT INTO users (id, username, email, first_name, last_name, 
                          password_hash, roles, groups, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
    
    _, err := ps.db.ExecContext(ctx, query,
        user.ID, user.Username, user.Email, user.FirstName, user.LastName,
        user.PasswordHash, pq.Array(user.Roles), pq.Array(user.Groups),
        user.CreatedAt, user.UpdatedAt)
    
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    return nil
}

func (ps *PostgresStorage) GetUser(ctx context.Context, id uuid.UUID) (*storage.User, error) {
    query := `
        SELECT id, username, email, first_name, last_name, password_hash,
               roles, groups, active, last_login, created_at, updated_at
        FROM users WHERE id = $1`
    
    user := &storage.User{}
    var roles, groups pq.StringArray
    
    err := ps.db.QueryRowContext(ctx, query, id).Scan(
        &user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
        &user.PasswordHash, &roles, &groups, &user.Active, &user.LastLogin,
        &user.CreatedAt, &user.UpdatedAt)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, storage.ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    user.Roles = []string(roles)
    user.Groups = []string(groups)
    
    return user, nil
}
```

### Audit Logging System

**Audit Event Structure**
```go
// pkg/audit/audit.go
package audit

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/google/uuid"
)

type AuditLogger struct {
    storage  AuditStorage
    config   *Config
    logger   Logger
}

type Config struct {
    Enabled         bool          `yaml:"enabled"`
    Level           Level         `yaml:"level"`
    RetentionDays   int           `yaml:"retention_days"`
    BufferSize      int           `yaml:"buffer_size"`
    FlushInterval   time.Duration `yaml:"flush_interval"`
    IncludePayload  bool          `yaml:"include_payload"`
    SensitiveFields []string      `yaml:"sensitive_fields"`
}

type Level string
const (
    LevelInfo    Level = "info"
    LevelWarning Level = "warning"
    LevelError   Level = "error"
    LevelCritical Level = "critical"
)

type EventType string
const (
    EventTypeAuth           EventType = "auth"
    EventTypeUser           EventType = "user"
    EventTypeMission        EventType = "mission"
    EventTypeMessage        EventType = "message"
    EventTypeSystem         EventType = "system"
    EventTypePosition       EventType = "position"
    EventTypeConfiguration  EventType = "configuration"
)

type AuditEvent struct {
    ID          uuid.UUID              `json:"id"`
    Type        EventType              `json:"type"`
    Action      string                 `json:"action"`
    Level       Level                  `json:"level"`
    UserID      *uuid.UUID             `json:"user_id,omitempty"`
    Username    string                 `json:"username,omitempty"`
    ResourceID  *uuid.UUID             `json:"resource_id,omitempty"`
    ResourceType string                `json:"resource_type,omitempty"`
    
    // Request context
    IPAddress   string                 `json:"ip_address,omitempty"`
    UserAgent   string                 `json:"user_agent,omitempty"`
    RequestID   string                 `json:"request_id,omitempty"`
    
    // Event details
    Message     string                 `json:"message"`
    Details     map[string]interface{} `json:"details,omitempty"`
    Error       string                 `json:"error,omitempty"`
    
    // Timing
    Timestamp   time.Time              `json:"timestamp"`
    Duration    *time.Duration         `json:"duration,omitempty"`
    
    // Classification
    Classification string              `json:"classification,omitempty"`
    Sensitivity    string              `json:"sensitivity,omitempty"`
}

func (al *AuditLogger) LogAuthentication(ctx context.Context, userID uuid.UUID, username, action string, success bool) {
    level := LevelInfo
    message := fmt.Sprintf("User %s: %s", username, action)
    
    if !success {
        level = LevelWarning
        message += " (failed)"
    }
    
    event := &AuditEvent{
        ID:       uuid.New(),
        Type:     EventTypeAuth,
        Action:   action,
        Level:    level,
        UserID:   &userID,
        Username: username,
        Message:  message,
        Details: map[string]interface{}{
            "success": success,
            "action":  action,
        },
        Timestamp: time.Now(),
    }
    
    al.logEvent(ctx, event)
}

func (al *AuditLogger) LogMissionAccess(ctx context.Context, userID uuid.UUID, missionID uuid.UUID, action string) {
    event := &AuditEvent{
        ID:           uuid.New(),
        Type:         EventTypeMission,
        Action:       action,
        Level:        LevelInfo,
        UserID:       &userID,
        ResourceID:   &missionID,
        ResourceType: "mission",
        Message:      fmt.Sprintf("Mission %s: %s", missionID.String(), action),
        Timestamp:    time.Now(),
    }
    
    al.logEvent(ctx, event)
}

func (al *AuditLogger) LogSystemEvent(ctx context.Context, action, message string, details map[string]interface{}) {
    event := &AuditEvent{
        ID:        uuid.New(),
        Type:      EventTypeSystem,
        Action:    action,
        Level:     LevelInfo,
        Message:   message,
        Details:   details,
        Timestamp: time.Now(),
    }
    
    al.logEvent(ctx, event)
}

func (al *AuditLogger) logEvent(ctx context.Context, event *AuditEvent) {
    // Add request context if available
    if userID := getUserIDFromContext(ctx); userID != "" {
        if event.UserID == nil {
            if uid, err := uuid.Parse(userID); err == nil {
                event.UserID = &uid
            }
        }
    }
    
    if reqID := getRequestIDFromContext(ctx); reqID != "" {
        event.RequestID = reqID
    }
    
    if ip := getIPFromContext(ctx); ip != "" {
        event.IPAddress = ip
    }
    
    // Filter sensitive data
    if event.Details != nil {
        event.Details = al.filterSensitiveData(event.Details)
    }
    
    // Store audit event
    if err := al.storage.CreateAuditEvent(ctx, event); err != nil {
        al.logger.Error("Failed to store audit event", "error", err, "event_id", event.ID)
    }
    
    // Log to structured logger as well
    al.logger.Info("Audit event", 
        "event_id", event.ID,
        "type", event.Type,
        "action", event.Action,
        "user_id", event.UserID,
        "message", event.Message)
}
```

### Database Migration System

**Migration Manager**
```go
// pkg/storage/migration/manager.go
package migration

import (
    "context"
    "database/sql"
    "fmt"
    "io/fs"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "time"
)

type Manager struct {
    db         *sql.DB
    logger     Logger
    migrations []Migration
}

type Migration struct {
    Version     int
    Name        string
    UpSQL       string
    DownSQL     string
    Filepath    string
}

type MigrationRecord struct {
    Version     int       `db:"version"`
    Name        string    `db:"name"`
    AppliedAt   time.Time `db:"applied_at"`
    Checksum    string    `db:"checksum"`
}

func NewManager(db *sql.DB, logger Logger) *Manager {
    return &Manager{
        db:     db,
        logger: logger,
    }
}

func (m *Manager) LoadMigrationsFromFS(migrationFS fs.FS) error {
    err := fs.WalkDir(migrationFS, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        
        if d.IsDir() || !strings.HasSuffix(path, ".sql") {
            return nil
        }
        
        migration, err := m.parseMigrationFile(migrationFS, path)
        if err != nil {
            return fmt.Errorf("failed to parse migration %s: %w", path, err)
        }
        
        m.migrations = append(m.migrations, migration)
        return nil
    })
    
    if err != nil {
        return err
    }
    
    // Sort migrations by version
    sort.Slice(m.migrations, func(i, j int) bool {
        return m.migrations[i].Version < m.migrations[j].Version
    })
    
    return nil
}

func (m *Manager) Migrate(ctx context.Context) error {
    // Create migration table if it doesn't exist
    if err := m.createMigrationTable(ctx); err != nil {
        return fmt.Errorf("failed to create migration table: %w", err)
    }
    
    // Get applied migrations
    applied, err := m.getAppliedMigrations(ctx)
    if err != nil {
        return fmt.Errorf("failed to get applied migrations: %w", err)
    }
    
    appliedSet := make(map[int]bool)
    for _, record := range applied {
        appliedSet[record.Version] = true
    }
    
    // Apply pending migrations
    for _, migration := range m.migrations {
        if appliedSet[migration.Version] {
            m.logger.Debug("Skipping already applied migration", 
                "version", migration.Version, "name", migration.Name)
            continue
        }
        
        m.logger.Info("Applying migration", 
            "version", migration.Version, "name", migration.Name)
        
        if err := m.applyMigration(ctx, migration); err != nil {
            return fmt.Errorf("failed to apply migration %d (%s): %w", 
                migration.Version, migration.Name, err)
        }
    }
    
    return nil
}

func (m *Manager) applyMigration(ctx context.Context, migration Migration) error {
    tx, err := m.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to start transaction: %w", err)
    }
    defer tx.Rollback()
    
    // Execute migration SQL
    if _, err := tx.ExecContext(ctx, migration.UpSQL); err != nil {
        return fmt.Errorf("failed to execute migration SQL: %w", err)
    }
    
    // Record migration
    checksum := calculateChecksum(migration.UpSQL)
    _, err = tx.ExecContext(ctx, 
        `INSERT INTO schema_migrations (version, name, applied_at, checksum) 
         VALUES ($1, $2, $3, $4)`,
        migration.Version, migration.Name, time.Now(), checksum)
    
    if err != nil {
        return fmt.Errorf("failed to record migration: %w", err)
    }
    
    return tx.Commit()
}
```

### Admin API Layer

**Admin Service**
```go
// internal/admin/service.go
package admin

import (
    "context"
    "fmt"
    "time"
    
    "github.com/dfedick/gotak/pkg/storage"
    "github.com/dfedick/gotak/pkg/audit"
)

type AdminService struct {
    storage     storage.Storage
    audit       *audit.AuditLogger
    logger      Logger
    config      *Config
}

type Config struct {
    EnableUserManagement bool     `yaml:"enable_user_management"`
    EnableSystemControl  bool     `yaml:"enable_system_control"`
    AllowedRoles        []string `yaml:"allowed_roles"`
    RequiredPermissions []string `yaml:"required_permissions"`
}

type SystemStatus struct {
    Version        string                 `json:"version"`
    Uptime         time.Duration          `json:"uptime"`
    DatabaseHealth string                 `json:"database_health"`
    
    // Resource usage
    CPUUsage       float64                `json:"cpu_usage"`
    MemoryUsage    int64                  `json:"memory_usage"`
    DiskUsage      int64                  `json:"disk_usage"`
    
    // Application metrics
    ActiveUsers    int                    `json:"active_users"`
    TotalUsers     int                    `json:"total_users"`
    ActiveMissions int                    `json:"active_missions"`
    MessagesSent   int64                  `json:"messages_sent"`
    
    // Performance metrics
    AvgResponseTime time.Duration         `json:"avg_response_time"`
    RequestsPerSec  float64               `json:"requests_per_sec"`
    ErrorRate       float64               `json:"error_rate"`
    
    // Health checks
    HealthChecks   map[string]HealthCheck `json:"health_checks"`
}

type HealthCheck struct {
    Status      string        `json:"status"`
    LastCheck   time.Time     `json:"last_check"`
    Duration    time.Duration `json:"duration"`
    Error       string        `json:"error,omitempty"`
}

func (as *AdminService) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
    userID := getUserIDFromContext(ctx)
    
    // Audit system access
    if userID != "" {
        as.audit.LogSystemEvent(ctx, "get_status", "System status accessed", nil)
    }
    
    status := &SystemStatus{
        Version:      getVersion(),
        Uptime:       getUptime(),
        HealthChecks: make(map[string]HealthCheck),
    }
    
    // Database health check
    dbHealth := as.checkDatabaseHealth(ctx)
    status.DatabaseHealth = dbHealth.Status
    status.HealthChecks["database"] = dbHealth
    
    // Get user statistics
    if stats, err := as.getUserStatistics(ctx); err == nil {
        status.ActiveUsers = stats.ActiveUsers
        status.TotalUsers = stats.TotalUsers
    }
    
    // Get mission statistics
    if stats, err := as.getMissionStatistics(ctx); err == nil {
        status.ActiveMissions = stats.ActiveMissions
    }
    
    return status, nil
}

func (as *AdminService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    adminUserID := getUserIDFromContext(ctx)
    
    // Validate admin permissions
    if !as.hasPermission(ctx, "user.create") {
        as.audit.LogSystemEvent(ctx, "create_user_denied", 
            "User creation denied - insufficient permissions", 
            map[string]interface{}{"target_username": req.Username})
        return nil, ErrInsufficientPermissions
    }
    
    user := &User{
        ID:           uuid.New(),
        Username:     req.Username,
        Email:        req.Email,
        FirstName:    req.FirstName,
        LastName:     req.LastName,
        PasswordHash: hashPassword(req.Password),
        Roles:        req.Roles,
        Groups:       req.Groups,
        Active:       true,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    
    if err := as.storage.CreateUser(ctx, user); err != nil {
        as.audit.LogSystemEvent(ctx, "create_user_failed",
            "User creation failed",
            map[string]interface{}{
                "target_username": req.Username,
                "error": err.Error(),
            })
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    // Audit successful creation
    as.audit.LogSystemEvent(ctx, "create_user_success",
        fmt.Sprintf("User created: %s", req.Username),
        map[string]interface{}{
            "target_user_id": user.ID,
            "target_username": req.Username,
            "admin_user_id": adminUserID,
        })
    
    // Remove sensitive data before returning
    user.PasswordHash = ""
    
    return user, nil
}

func (as *AdminService) GetAuditEvents(ctx context.Context, filter AuditFilter) ([]*AuditEvent, int, error) {
    // Validate admin permissions
    if !as.hasPermission(ctx, "audit.read") {
        return nil, 0, ErrInsufficientPermissions
    }
    
    events, total, err := as.storage.GetAuditEvents(ctx, filter)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to get audit events: %w", err)
    }
    
    // Audit the audit access (meta!)
    as.audit.LogSystemEvent(ctx, "audit_access",
        fmt.Sprintf("Audit events accessed (count: %d)", len(events)),
        map[string]interface{}{
            "filter": filter,
            "result_count": len(events),
        })
    
    return events, total, nil
}
```

## Database Schema

```sql
-- Enhanced user management
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    roles TEXT[] DEFAULT '{}',
    groups TEXT[] DEFAULT '{}',
    active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    password_changed_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- User sessions
CREATE TABLE user_sessions (
    token VARCHAR(255) PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Comprehensive audit log
CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    level VARCHAR(20) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    username VARCHAR(255),
    resource_id UUID,
    resource_type VARCHAR(50),
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(255),
    message TEXT NOT NULL,
    details JSONB,
    error TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    duration INTERVAL,
    classification VARCHAR(50),
    sensitivity VARCHAR(50)
);

-- Schema migration tracking
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP NOT NULL DEFAULT NOW(),
    checksum VARCHAR(64) NOT NULL
);

-- System configuration
CREATE TABLE system_config (
    key VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    category VARCHAR(100),
    read_only BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Performance optimization indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(active);

CREATE INDEX idx_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_sessions_expires ON user_sessions(expires_at);

CREATE INDEX idx_audit_type_action ON audit_events(type, action);
CREATE INDEX idx_audit_user_time ON audit_events(user_id, timestamp DESC);
CREATE INDEX idx_audit_timestamp ON audit_events(timestamp DESC);
CREATE INDEX idx_audit_resource ON audit_events(resource_type, resource_id);

-- Partitioning for audit table (PostgreSQL 10+)
CREATE TABLE audit_events_y2025m01 PARTITION OF audit_events
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE audit_events_y2025m02 PARTITION OF audit_events
FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Add more partitions as needed
```

## API Specifications

### Admin API Endpoints
```
GET    /api/v1/admin/status              # System status and health
GET    /api/v1/admin/metrics             # Performance metrics
POST   /api/v1/admin/users               # Create user
GET    /api/v1/admin/users               # List users
GET    /api/v1/admin/users/{id}          # Get user
PUT    /api/v1/admin/users/{id}          # Update user  
DELETE /api/v1/admin/users/{id}          # Deactivate user
POST   /api/v1/admin/users/{id}/unlock   # Unlock user account
```

### Audit API Endpoints
```
GET    /api/v1/admin/audit               # List audit events
GET    /api/v1/admin/audit/search        # Search audit events
GET    /api/v1/admin/audit/export        # Export audit data
GET    /api/v1/admin/audit/statistics    # Audit statistics
POST   /api/v1/admin/audit/archive       # Archive old events
```

### System Management API
```
GET    /api/v1/admin/config              # System configuration
PUT    /api/v1/admin/config/{key}        # Update configuration
POST   /api/v1/admin/maintenance/start   # Start maintenance mode
POST   /api/v1/admin/maintenance/stop    # Stop maintenance mode
POST   /api/v1/admin/backup              # Create backup
GET    /api/v1/admin/backup/status       # Backup status
```

## Testing Strategy

### Unit Tests
```go
func TestUserStorage_CreateUser(t *testing.T) {
    storage := setupTestStorage()
    
    user := &User{
        ID:       uuid.New(),
        Username: "testuser",
        Email:    "test@example.com",
        PasswordHash: "hashed_password",
        Roles:    []string{"user"},
        Groups:   []string{"default"},
        Active:   true,
    }
    
    err := storage.CreateUser(context.Background(), user)
    assert.NoError(t, err)
    
    // Verify user was created
    retrieved, err := storage.GetUser(context.Background(), user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Username, retrieved.Username)
}

func TestAuditLogger_LogAuthentication(t *testing.T) {
    logger := setupTestAuditLogger()
    userID := uuid.New()
    
    logger.LogAuthentication(context.Background(), userID, "testuser", "login", true)
    
    // Verify audit event was created
    events, _, err := logger.storage.GetAuditEvents(context.Background(), AuditFilter{
        Type: EventTypeAuth,
        UserID: &userID,
    })
    
    assert.NoError(t, err)
    assert.Len(t, events, 1)
    assert.Equal(t, "login", events[0].Action)
}
```

### Integration Tests
```go
func TestMigrationSystem(t *testing.T) {
    db := setupTestDatabase()
    manager := migration.NewManager(db, testLogger)
    
    // Load test migrations
    err := manager.LoadMigrationsFromFS(testMigrationFS)
    assert.NoError(t, err)
    
    // Apply migrations
    err = manager.Migrate(context.Background())
    assert.NoError(t, err)
    
    // Verify tables were created
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
    assert.NoError(t, err)
    assert.True(t, count > 0)
}
```

## Acceptance Criteria

### Storage Layer
- [ ] Storage abstraction supports multiple backends
- [ ] PostgreSQL implementation handles production load
- [ ] Database connections properly pooled and managed
- [ ] Query performance optimized for common operations
- [ ] Transaction handling and rollback working correctly

### Migration System
- [ ] Database migrations apply successfully
- [ ] Rollback functionality working
- [ ] Migration history tracked and validated
- [ ] Schema changes deployed safely
- [ ] Backup and restore procedures tested

### Audit Logging
- [ ] All user actions captured in audit log
- [ ] Audit events structured and searchable
- [ ] Sensitive data properly filtered
- [ ] Log retention policies enforced
- [ ] Export functionality working for compliance

### Admin APIs
- [ ] User management endpoints functional
- [ ] System status and metrics accessible
- [ ] Configuration management working
- [ ] Proper authorization for admin functions
- [ ] Bulk operations perform efficiently

### Performance
- [ ] Database queries complete within SLA (< 100ms)
- [ ] Audit logging adds minimal overhead (< 10ms)
- [ ] Connection pooling prevents resource exhaustion
- [ ] Large result sets properly paginated
- [ ] Memory usage stable under load

## Dependencies

### Backend Dependencies
```go
require (
    github.com/lib/pq v1.10.9              // PostgreSQL driver
    github.com/golang-migrate/migrate/v4 v4.16.2 // Database migrations
    github.com/jmoiron/sqlx v1.3.5          // SQL extensions
    github.com/jackc/pgx/v5 v5.4.3         // PostgreSQL driver (alternative)
)
```

### Database Extensions
```sql
-- Enable useful PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
```

## Definition of Done

### Code Quality
- [ ] All code reviewed and approved
- [ ] Unit tests with 85%+ coverage
- [ ] Integration tests for database operations
- [ ] Performance benchmarks meet requirements
- [ ] Security review completed

### Functionality
- [ ] All user stories completed and accepted
- [ ] Database operations working reliably
- [ ] Audit logging capturing all required events
- [ ] Admin APIs functional and secure
- [ ] Migration system tested in production-like environment

### Performance & Reliability
- [ ] Database handles expected production load
- [ ] Backup and restore procedures verified
- [ ] Failover and recovery scenarios tested
- [ ] Monitoring and alerting configured
- [ ] Documentation complete for operations team

---

**Sprint Review Date:** [TBD]  
**Sprint Retrospective Date:** [TBD]  
**Next Sprint Planning:** [TBD]
