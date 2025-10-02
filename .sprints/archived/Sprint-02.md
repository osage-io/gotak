# Sprint 2: Authentication & Security Foundation

**Duration:** 2 weeks  
**Theme:** Zero Trust Security Implementation  
**Sprint Goals:** Implement core authentication service and security infrastructure

## Objectives

1. **Authentication Service**: JWT token management with refresh tokens and multi-factor support
2. **Authorization System**: RBAC with Casbin policy engine and hierarchical permissions
3. **Vault Integration**: HashiCorp Vault OIDC integration with fallback authentication
4. **Audit Logging**: Comprehensive audit trail infrastructure for security events
5. **Security Middleware**: Authentication and authorization middleware for all services

## User Stories

### Epic: Authentication System

**As a** military operator  
**I want** secure authentication with appropriate access controls  
**So that** only authorized personnel can access operational data  

### Story 1: User Authentication
**Acceptance Criteria:**
- [ ] Login with username/password authentication
- [ ] Multi-factor authentication support (TOTP)
- [ ] Session management with secure JWT tokens
- [ ] Password policy enforcement with complexity requirements
- [ ] Account lockout after failed attempts

### Story 2: HashiCorp Vault Integration
**Acceptance Criteria:**
- [ ] Vault OIDC authentication integration
- [ ] Dynamic secret management for database credentials
- [ ] Automatic token rotation and renewal
- [ ] Fallback authentication when Vault unavailable
- [ ] Vault policy configuration for service accounts

### Story 3: Role-Based Access Control
**Acceptance Criteria:**
- [ ] Define military roles and permissions hierarchy
- [ ] Implement Casbin RBAC policy engine
- [ ] Create authorization middleware for all endpoints
- [ ] Support for hierarchical permissions inheritance
- [ ] Role assignment and management interface

### Story 4: Audit Logging System
**Acceptance Criteria:**
- [ ] Log all authentication attempts (success/failure)
- [ ] Track authorization decisions and policy violations
- [ ] Secure audit trail storage with integrity protection
- [ ] Real-time security monitoring and alerting
- [ ] Audit log search and reporting capabilities

## Technical Implementation

### Authentication Service Architecture

```go
// internal/auth/service.go
type AuthService struct {
    db           database.DB
    vault        vault.Client
    logger       logger.Logger
    jwtSecret    []byte
    enforcer     *casbin.Enforcer
    pwdPolicy    PasswordPolicy
}

type User struct {
    ID              uuid.UUID `json:"id" db:"id"`
    Username        string    `json:"username" db:"username"`
    Email           string    `json:"email" db:"email"`
    PasswordHash    string    `json:"-" db:"password_hash"`
    FirstName       string    `json:"first_name" db:"first_name"`
    LastName        string    `json:"last_name" db:"last_name"`
    IsActive        bool      `json:"is_active" db:"is_active"`
    MFAEnabled      bool      `json:"mfa_enabled" db:"mfa_enabled"`
    MFASecret       string    `json:"-" db:"mfa_secret"`
    LastLogin       time.Time `json:"last_login" db:"last_login"`
    FailedAttempts  int       `json:"-" db:"failed_attempts"`
    LockedUntil     *time.Time `json:"-" db:"locked_until"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
    TokenType    string `json:"token_type"`
}

type LoginRequest struct {
    Username    string `json:"username" validate:"required"`
    Password    string `json:"password" validate:"required"`
    MFACode     string `json:"mfa_code,omitempty"`
    AuthMethod  string `json:"auth_method" validate:"required,oneof=local oidc cert"`
}
```

### JWT Token Management

```go
// internal/auth/jwt.go
type JWTManager struct {
    secret        []byte
    accessTTL     time.Duration
    refreshTTL    time.Duration
    issuer        string
}

type Claims struct {
    UserID      string   `json:"user_id"`
    Username    string   `json:"username"`
    Roles       []string `json:"roles"`
    Permissions []string `json:"permissions"`
    TokenType   string   `json:"token_type"` // "access" or "refresh"
    jwt.RegisteredClaims
}

func (j *JWTManager) GenerateTokenPair(user *User, roles []Role) (*TokenPair, error) {
    // Generate access token
    accessClaims := &Claims{
        UserID:    user.ID.String(),
        Username:  user.Username,
        Roles:     extractRoleNames(roles),
        Permissions: extractPermissions(roles),
        TokenType: "access",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTTL)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    j.issuer,
            Subject:   user.ID.String(),
        },
    }

    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, err := accessToken.SignedString(j.secret)
    if err != nil {
        return nil, err
    }

    // Generate refresh token
    refreshClaims := &Claims{
        UserID:    user.ID.String(),
        Username:  user.Username,
        TokenType: "refresh",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTTL)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    j.issuer,
            Subject:   user.ID.String(),
        },
    }

    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshTokenString, err := refreshToken.SignedString(j.secret)
    if err != nil {
        return nil, err
    }

    return &TokenPair{
        AccessToken:  accessTokenString,
        RefreshToken: refreshTokenString,
        ExpiresIn:    int64(j.accessTTL.Seconds()),
        TokenType:    "Bearer",
    }, nil
}

func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return j.secret, nil
    })

    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }

    return claims, nil
}
```

### HashiCorp Vault Integration

```go
// internal/auth/vault.go
type VaultClient struct {
    client   *vault.Client
    config   VaultConfig
    logger   logger.Logger
}

type VaultConfig struct {
    Address    string `mapstructure:"address"`
    Token      string `mapstructure:"token"`
    RoleID     string `mapstructure:"role_id"`
    SecretID   string `mapstructure:"secret_id"`
    MountPath  string `mapstructure:"mount_path"`
    Namespace  string `mapstructure:"namespace"`
}

func NewVaultClient(config VaultConfig, logger logger.Logger) (*VaultClient, error) {
    vaultConfig := vault.DefaultConfig()
    vaultConfig.Address = config.Address

    client, err := vault.NewClient(vaultConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create vault client: %w", err)
    }

    if config.Namespace != "" {
        client.SetNamespace(config.Namespace)
    }

    vc := &VaultClient{
        client: client,
        config: config,
        logger: logger,
    }

    // Authenticate with Vault
    if err := vc.authenticate(); err != nil {
        return nil, fmt.Errorf("failed to authenticate with vault: %w", err)
    }

    return vc, nil
}

func (vc *VaultClient) authenticate() error {
    // Try token authentication first
    if vc.config.Token != "" {
        vc.client.SetToken(vc.config.Token)
        return nil
    }

    // Fall back to AppRole authentication
    if vc.config.RoleID != "" && vc.config.SecretID != "" {
        data := map[string]interface{}{
            "role_id":   vc.config.RoleID,
            "secret_id": vc.config.SecretID,
        }

        resp, err := vc.client.Logical().Write("auth/approle/login", data)
        if err != nil {
            return err
        }

        if resp == nil || resp.Auth == nil {
            return errors.New("no auth info returned from vault")
        }

        vc.client.SetToken(resp.Auth.ClientToken)
        vc.logger.Info().Msg("Successfully authenticated with Vault using AppRole")
        return nil
    }

    return errors.New("no valid vault authentication method configured")
}

func (vc *VaultClient) VerifyOIDCToken(token string) (*OIDCUserInfo, error) {
    data := map[string]interface{}{
        "jwt": token,
    }

    resp, err := vc.client.Logical().Write(vc.config.MountPath+"/login", data)
    if err != nil {
        return nil, fmt.Errorf("failed to verify OIDC token: %w", err)
    }

    if resp == nil || resp.Auth == nil {
        return nil, errors.New("invalid token response from vault")
    }

    // Extract user info from response
    userInfo := &OIDCUserInfo{
        Subject:   resp.Auth.DisplayName,
        Email:     extractString(resp.Auth.Metadata, "email"),
        Name:      extractString(resp.Auth.Metadata, "name"),
        Groups:    extractStringSlice(resp.Auth.Metadata, "groups"),
        VaultToken: resp.Auth.ClientToken,
    }

    return userInfo, nil
}
```

### RBAC with Casbin

```go
// internal/auth/rbac.go
type RBACManager struct {
    enforcer *casbin.Enforcer
    db       database.DB
    logger   logger.Logger
}

func NewRBACManager(db database.DB, logger logger.Logger) (*RBACManager, error) {
    // Load model from embedded file
    model, err := model.NewModelFromString(rbacModel)
    if err != nil {
        return nil, fmt.Errorf("failed to load RBAC model: %w", err)
    }

    // Create adapter for database policy storage
    adapter, err := gormadapter.NewAdapterByDB(db.GetGormDB())
    if err != nil {
        return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
    }

    // Create enforcer
    enforcer, err := casbin.NewEnforcer(model, adapter)
    if err != nil {
        return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
    }

    // Load policies from database
    if err := enforcer.LoadPolicy(); err != nil {
        return nil, fmt.Errorf("failed to load policies: %w", err)
    }

    return &RBACManager{
        enforcer: enforcer,
        db:       db,
        logger:   logger,
    }, nil
}

// RBAC Model Configuration
const rbacModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
`

func (rm *RBACManager) Enforce(userID string, resource string, action string) (bool, error) {
    allowed, err := rm.enforcer.Enforce(userID, resource, action)
    if err != nil {
        rm.logger.Error().
            Err(err).
            Str("user_id", userID).
            Str("resource", resource).
            Str("action", action).
            Msg("RBAC enforcement error")
        return false, err
    }

    rm.logger.Debug().
        Str("user_id", userID).
        Str("resource", resource).
        Str("action", action).
        Bool("allowed", allowed).
        Msg("RBAC enforcement decision")

    return allowed, nil
}

func (rm *RBACManager) AddRoleForUser(userID string, role string) error {
    added, err := rm.enforcer.AddRoleForUser(userID, role)
    if err != nil {
        return err
    }

    if added {
        rm.logger.Info().
            Str("user_id", userID).
            Str("role", role).
            Msg("Added role for user")
    }

    return rm.enforcer.SavePolicy()
}

func (rm *RBACManager) GetRolesForUser(userID string) ([]string, error) {
    return rm.enforcer.GetRolesForUser(userID)
}
```

### Authentication Middleware

```go
// internal/middleware/auth.go
type AuthMiddleware struct {
    jwtManager   *auth.JWTManager
    rbacManager  *auth.RBACManager
    logger       logger.Logger
    skipPaths    []string
}

func NewAuthMiddleware(jwtManager *auth.JWTManager, rbacManager *auth.RBACManager, logger logger.Logger) *AuthMiddleware {
    return &AuthMiddleware{
        jwtManager:  jwtManager,
        rbacManager: rbacManager,
        logger:      logger,
        skipPaths:   []string{"/health", "/metrics", "/v1/auth/login", "/v1/auth/refresh"},
    }
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip authentication for certain paths
        for _, path := range m.skipPaths {
            if strings.HasPrefix(r.URL.Path, path) {
                next.ServeHTTP(w, r)
                return
            }
        }

        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }

        tokenParts := strings.SplitN(authHeader, " ", 2)
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
            return
        }

        // Validate JWT token
        claims, err := m.jwtManager.ValidateToken(tokenParts[1])
        if err != nil {
            m.logger.Warn().
                Err(err).
                Str("ip", r.RemoteAddr).
                Str("user_agent", r.UserAgent()).
                Msg("Invalid token")
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Check if token is access token
        if claims.TokenType != "access" {
            http.Error(w, "Invalid token type", http.StatusUnauthorized)
            return
        }

        // Add user info to request context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "username", claims.Username)
        ctx = context.WithValue(ctx, "roles", claims.Roles)
        ctx = context.WithValue(ctx, "permissions", claims.Permissions)

        // Continue to next handler
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (m *AuthMiddleware) Authorize(resource string, action string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := r.Context().Value("user_id").(string)
            
            allowed, err := m.rbacManager.Enforce(userID, resource, action)
            if err != nil {
                m.logger.Error().
                    Err(err).
                    Str("user_id", userID).
                    Str("resource", resource).
                    Str("action", action).
                    Msg("Authorization error")
                http.Error(w, "Authorization error", http.StatusInternalServerError)
                return
            }

            if !allowed {
                m.logger.Warn().
                    Str("user_id", userID).
                    Str("resource", resource).
                    Str("action", action).
                    Str("ip", r.RemoteAddr).
                    Msg("Access denied")
                http.Error(w, "Access denied", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### Database Schema

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    mfa_enabled BOOLEAN DEFAULT false,
    mfa_secret VARCHAR(255),
    last_login TIMESTAMP,
    failed_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- User roles junction table
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Refresh tokens table
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    revoked_at TIMESTAMP,
    CONSTRAINT unique_active_token UNIQUE (user_id, token_hash)
);

-- Audit logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(255) NOT NULL,
    resource VARCHAR(255),
    resource_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    details JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Password reset tokens
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Casbin rules (for RBAC policies)
CREATE TABLE casbin_rules (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(255) NOT NULL,
    v0 VARCHAR(255),
    v1 VARCHAR(255),
    v2 VARCHAR(255),
    v3 VARCHAR(255),
    v4 VARCHAR(255),
    v5 VARCHAR(255)
);

-- Indexes for performance
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);

-- Insert default roles
INSERT INTO roles (name, description) VALUES
    ('system_admin', 'Full system administration access'),
    ('mission_commander', 'Mission planning and personnel management'),
    ('operator', 'Operational access for field personnel'),
    ('observer', 'Read-only access to operational picture');
```

### API Endpoints Implementation

```go
// internal/handlers/auth.go
type AuthHandler struct {
    authService *auth.Service
    logger      logger.Logger
    validator   *validator.Validate
}

// POST /v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req auth.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    if err := h.validator.Struct(req); err != nil {
        http.Error(w, "Validation failed", http.StatusBadRequest)
        return
    }

    // Attempt authentication
    tokenPair, err := h.authService.Authenticate(r.Context(), &req)
    if err != nil {
        h.logger.Warn().
            Err(err).
            Str("username", req.Username).
            Str("auth_method", req.AuthMethod).
            Str("ip", r.RemoteAddr).
            Msg("Authentication failed")

        // Don't reveal specific error details to prevent enumeration
        http.Error(w, "Authentication failed", http.StatusUnauthorized)
        return
    }

    h.logger.Info().
        Str("username", req.Username).
        Str("auth_method", req.AuthMethod).
        Str("ip", r.RemoteAddr).
        Msg("User authenticated successfully")

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tokenPair)
}

// POST /v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
    var req struct {
        RefreshToken string `json:"refresh_token" validate:"required"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    if err := h.validator.Struct(req); err != nil {
        http.Error(w, "Validation failed", http.StatusBadRequest)
        return
    }

    tokenPair, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
    if err != nil {
        h.logger.Warn().
            Err(err).
            Str("ip", r.RemoteAddr).
            Msg("Token refresh failed")
        http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tokenPair)
}

// POST /v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    if err := h.authService.RevokeRefreshToken(r.Context(), userID, req.RefreshToken); err != nil {
        h.logger.Warn().Err(err).Str("user_id", userID).Msg("Failed to revoke refresh token")
    }

    h.logger.Info().
        Str("user_id", userID).
        Str("ip", r.RemoteAddr).
        Msg("User logged out")

    w.WriteHeader(http.StatusNoContent)
}
```

## API Specifications

### Authentication Endpoints
```yaml
# Authentication Service API
POST   /v1/auth/login              # User login
POST   /v1/auth/logout             # User logout  
POST   /v1/auth/refresh            # Refresh JWT token
GET    /v1/auth/profile            # Get user profile
PUT    /v1/auth/profile            # Update user profile
POST   /v1/auth/change-password    # Change password

# Authorization endpoints
GET    /v1/auth/permissions        # List user permissions
POST   /v1/auth/authorize          # Check authorization
GET    /v1/auth/roles              # List available roles

# MFA endpoints
POST   /v1/auth/mfa/setup          # Setup MFA
POST   /v1/auth/mfa/verify         # Verify MFA code
DELETE /v1/auth/mfa/disable        # Disable MFA
```

## Deliverables

### Must Have
- [ ] Authentication service with JWT token management
- [ ] User registration and profile management  
- [ ] RBAC system with Casbin integration
- [ ] Password policy enforcement
- [ ] Audit logging system
- [ ] Authentication middleware for all services
- [ ] HashiCorp Vault integration

### Should Have
- [ ] MFA support framework (TOTP)
- [ ] Account lockout after failed attempts
- [ ] Password reset functionality
- [ ] Session management and revocation
- [ ] Security event monitoring

### Could Have
- [ ] Social login integration
- [ ] LDAP/Active Directory integration
- [ ] Advanced password policies
- [ ] Behavioral analysis for anomaly detection

## Acceptance Criteria

### Authentication
- [ ] Users can successfully login and receive JWT tokens
- [ ] Token refresh mechanism works correctly
- [ ] MFA can be enabled and verified
- [ ] Password policies are enforced
- [ ] Account lockout prevents brute force attacks

### Authorization
- [ ] RBAC permissions are enforced on all endpoints
- [ ] Role assignment and management works correctly
- [ ] Hierarchical permissions are inherited properly
- [ ] Access denied responses are logged and audited

### Vault Integration
- [ ] OIDC authentication works with Vault
- [ ] Dynamic secrets are retrieved and rotated
- [ ] Fallback authentication works when Vault unavailable
- [ ] Service accounts authenticate with AppRole

### Audit & Security
- [ ] All auth events are logged to audit system
- [ ] Security events trigger real-time alerts
- [ ] Audit logs are tamper-evident
- [ ] Failed authentication attempts are tracked

## Testing Strategy

### Unit Tests
```go
func TestJWTTokenGeneration(t *testing.T) {
    jwtManager := auth.NewJWTManager([]byte("test-secret"), time.Hour, time.Hour*24)
    user := &auth.User{
        ID:       uuid.New(),
        Username: "testuser",
    }
    
    tokenPair, err := jwtManager.GenerateTokenPair(user, nil)
    assert.NoError(t, err)
    assert.NotEmpty(t, tokenPair.AccessToken)
    assert.NotEmpty(t, tokenPair.RefreshToken)
    
    claims, err := jwtManager.ValidateToken(tokenPair.AccessToken)
    assert.NoError(t, err)
    assert.Equal(t, user.ID.String(), claims.UserID)
}

func TestRBACEnforcement(t *testing.T) {
    rbacManager := setupTestRBAC()
    
    // Add user role
    err := rbacManager.AddRoleForUser("user1", "operator")
    assert.NoError(t, err)
    
    // Test permission enforcement
    allowed, err := rbacManager.Enforce("user1", "/api/v1/missions", "read")
    assert.NoError(t, err)
    assert.True(t, allowed)
    
    denied, err := rbacManager.Enforce("user1", "/api/v1/admin", "write")
    assert.NoError(t, err)
    assert.False(t, denied)
}
```

### Integration Tests
```go
func TestAuthenticationFlow(t *testing.T) {
    server := setupTestServer()
    
    // Test login
    loginReq := auth.LoginRequest{
        Username:   "testuser",
        Password:   "testpass123!",
        AuthMethod: "local",
    }
    
    resp := testRequest(t, server, "POST", "/v1/auth/login", loginReq)
    assert.Equal(t, http.StatusOK, resp.Code)
    
    var tokenResponse auth.TokenPair
    json.Unmarshal(resp.Body.Bytes(), &tokenResponse)
    assert.NotEmpty(t, tokenResponse.AccessToken)
    
    // Test protected endpoint access
    req := httptest.NewRequest("GET", "/v1/auth/profile", nil)
    req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)
    
    resp = httptest.NewRecorder()
    server.ServeHTTP(resp, req)
    assert.Equal(t, http.StatusOK, resp.Code)
}
```

## Dependencies

### Go Dependencies
```go
require (
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/casbin/casbin/v2 v2.82.0
    github.com/casbin/gorm-adapter/v3 v3.20.0
    github.com/hashicorp/vault/api v1.10.0
    github.com/pquerna/otp v1.4.0
    golang.org/x/crypto v0.17.0
    github.com/go-playground/validator/v10 v10.16.0
)
```

### External Services
- **HashiCorp Vault**: OIDC authentication and secrets management
- **PostgreSQL**: User and audit data storage
- **Redis**: Session storage and token blacklisting

## Definition of Done

### Code Quality
- [ ] All code reviewed and approved
- [ ] Unit tests with >85% coverage
- [ ] Integration tests pass
- [ ] Security scanning passes
- [ ] RBAC policies tested thoroughly

### Security
- [ ] Vulnerability assessment completed
- [ ] Security controls tested
- [ ] Audit logging verified
- [ ] Token security validated
- [ ] Password policies enforced

### Documentation
- [ ] API documentation complete
- [ ] Security architecture documented
- [ ] RBAC model documented
- [ ] Integration guides written

---

**Sprint Review Date:** [TBD]  
**Sprint Retrospective Date:** [TBD]  
**Next Sprint Planning:** [TBD]
