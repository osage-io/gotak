# Sprint 1: Authentication & User Management

## Sprint Overview
- **Duration**: 1 Week
- **Start Date**: Today
- **Goal**: Implement complete authentication system with user management
- **Priority**: Critical (blocks all other features)

## Current State
- Login page exists in UI (`web/src/pages/Login.tsx`)
- No backend authentication implemented
- Using mock localStorage token
- No user database schema
- No session management

## Sprint Goals

### Primary Goals
1. ✅ Working login/logout flow
2. ✅ JWT token generation and validation
3. ✅ User registration
4. ✅ Protected API endpoints
5. ✅ Session management

### Stretch Goals
- Password reset functionality
- Role-based access control (RBAC)
- User profile management
- Multi-factor authentication

## Technical Implementation

### 1. Database Setup

#### Create Migration Files
```bash
# Create new migration
migrate create -ext sql -dir migrations -seq add_users_auth
```

#### SQL Schema
```sql
-- migrations/004_add_users_auth.up.sql

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'operator',
    callsign VARCHAR(50),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);

-- Sessions table for active sessions
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User preferences
CREATE TABLE IF NOT EXISTS user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(20) DEFAULT 'dark',
    language VARCHAR(10) DEFAULT 'en',
    notification_settings JSONB DEFAULT '{}',
    map_preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

### 2. Go Implementation

#### Package Structure
```
internal/
  auth/
    service.go      # Authentication service
    jwt.go          # JWT token handling
    middleware.go   # Auth middleware
    password.go     # Password hashing
  user/
    service.go      # User management service
    repository.go   # Database operations
    models.go       # User models
```

#### Auth Service (`internal/auth/service.go`)
```go
package auth

import (
    "context"
    "time"
    "github.com/google/uuid"
    "github.com/dfedick/gotak/pkg/logger"
)

type Service struct {
    userRepo     UserRepository
    sessionRepo  SessionRepository
    jwtSecret    []byte
    logger       *logger.Logger
}

func NewService(userRepo UserRepository, sessionRepo SessionRepository, jwtSecret string, log *logger.Logger) *Service {
    return &Service{
        userRepo:     userRepo,
        sessionRepo:  sessionRepo,
        jwtSecret:    []byte(jwtSecret),
        logger:       log,
    }
}

func (s *Service) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
    // Implementation
}

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
    // Implementation
}

func (s *Service) Logout(ctx context.Context, token string) error {
    // Implementation
}

func (s *Service) ValidateToken(token string) (*Claims, error) {
    // Implementation
}
```

### 3. API Endpoints

#### Auth Handlers (`internal/handlers/auth.go`)
```go
package handlers

import (
    "net/http"
    "github.com/dfedick/gotak/internal/auth"
)

type AuthHandlers struct {
    authService *auth.Service
    logger      *logger.Logger
}

// POST /api/v1/auth/register
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
    // Parse request
    // Validate input
    // Call service
    // Return response
}

// POST /api/v1/auth/login
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
    // Parse credentials
    // Authenticate
    // Generate token
    // Return token
}

// POST /api/v1/auth/logout
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
    // Get token from header
    // Invalidate session
    // Return success
}

// GET /api/v1/auth/me
func (h *AuthHandlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
    // Get user from context
    // Return user data
}
```

### 4. Middleware

#### Auth Middleware (`internal/auth/middleware.go`)
```go
package auth

import "net/http"

func (s *Service) AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get token from Authorization header
        // Validate token
        // Get user from token
        // Add user to context
        // Call next handler
    })
}
```

### 5. Frontend Integration

#### Update API Client (`web/src/services/apiClient.ts`)
```typescript
class AuthAPI {
    async login(username: string, password: string): Promise<LoginResponse> {
        const response = await fetch('/api/v1/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        
        if (!response.ok) throw new Error('Login failed');
        
        const data = await response.json();
        localStorage.setItem('authToken', data.token);
        return data;
    }
    
    async logout(): Promise<void> {
        const token = localStorage.getItem('authToken');
        await fetch('/api/v1/auth/logout', {
            method: 'POST',
            headers: { 
                'Authorization': `Bearer ${token}`
            }
        });
        localStorage.removeItem('authToken');
    }
    
    async getCurrentUser(): Promise<User> {
        const token = localStorage.getItem('authToken');
        const response = await fetch('/api/v1/auth/me', {
            headers: { 
                'Authorization': `Bearer ${token}`
            }
        });
        return response.json();
    }
}
```

## Testing Plan

### Unit Tests
- [ ] Password hashing and verification
- [ ] JWT token generation and validation
- [ ] User creation and validation
- [ ] Session management

### Integration Tests
- [ ] Complete login flow
- [ ] Registration with validation
- [ ] Token refresh
- [ ] Protected endpoint access
- [ ] Concurrent session handling

### Manual Testing
- [ ] Login via UI
- [ ] Logout functionality
- [ ] Session persistence
- [ ] Token expiration
- [ ] Invalid credentials handling

## Acceptance Criteria

### Must Have
- [ ] Users can register with username, email, password
- [ ] Users can login with username/password
- [ ] JWT tokens are generated on successful login
- [ ] Tokens expire after configurable time (default 24h)
- [ ] All API endpoints require valid token (except auth endpoints)
- [ ] Logout invalidates the session
- [ ] Passwords are hashed using bcrypt
- [ ] Login attempts are rate-limited

### Nice to Have
- [ ] Password strength requirements
- [ ] Email verification on registration
- [ ] Remember me functionality
- [ ] Password reset via email
- [ ] Account lockout after failed attempts
- [ ] Two-factor authentication

## Implementation Steps

### Day 1-2: Database & Models
1. Create and run migrations
2. Implement user models
3. Create repository layer
4. Write unit tests

### Day 3-4: Authentication Service
1. Implement password hashing
2. Create JWT handling
3. Build auth service
4. Add session management
5. Write service tests

### Day 5-6: API & Middleware
1. Create auth handlers
2. Implement auth middleware
3. Update route configuration
4. Add rate limiting
5. Write integration tests

### Day 7: Frontend Integration
1. Update login page
2. Implement auth service client
3. Add auth context/store
4. Update navigation guards
5. Test complete flow

## Configuration

### Environment Variables
```env
# JWT Configuration
JWT_SECRET=your-secret-key-here
JWT_EXPIRES_IN=24h

# Password Policy
MIN_PASSWORD_LENGTH=8
REQUIRE_UPPERCASE=true
REQUIRE_NUMBER=true
REQUIRE_SPECIAL=false

# Rate Limiting
LOGIN_ATTEMPTS_MAX=5
LOGIN_ATTEMPTS_WINDOW=15m
```

### Server Configuration
```yaml
# config/server.yaml
auth:
  jwt:
    secret: ${JWT_SECRET}
    expires_in: 24h
    refresh_expires_in: 7d
  password:
    min_length: 8
    require_uppercase: true
    require_number: true
    require_special: false
  rate_limit:
    enabled: true
    max_attempts: 5
    window: 15m
```

## Success Metrics

### Performance
- Login response time < 200ms
- Token validation < 10ms
- Concurrent users: 1000+

### Security
- All passwords bcrypt hashed
- No plain text passwords in logs
- Tokens use HS256 or RS256
- Rate limiting prevents brute force

### Functionality
- 100% of auth endpoints working
- Zero authentication bypasses
- Seamless token refresh
- Proper error messages

## Rollback Plan

If authentication causes issues:
1. Revert to mock auth (localStorage only)
2. Keep database schema (for next attempt)
3. Document issues encountered
4. Plan fixes for next sprint

## Notes

- Start with simple username/password auth
- Add OAuth2/SAML later if needed
- Consider using existing auth libraries
- Ensure CORS is properly configured
- Add security headers to all responses
- Log all authentication events
- Monitor for suspicious activity

## Dependencies

- `golang-jwt/jwt/v5` - JWT handling
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/google/uuid` - UUID generation
- Existing PostgreSQL database
- Redis for session storage (optional)

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Tests passing (>80% coverage)
- [ ] Code reviewed and approved
- [ ] Documentation updated
- [ ] Deployed to development environment
- [ ] Manual testing completed
- [ ] No critical security issues
- [ ] Performance benchmarks met