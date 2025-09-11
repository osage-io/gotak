# GoTAK Authentication & Security System

## Overview

The GoTAK authentication system provides military-grade security features including JWT token management, comprehensive password policies, account lockout protection, and role-based access control (RBAC). The system is designed for enterprise and government environments requiring high security standards.

## Table of Contents

1. [Architecture](#architecture)
2. [Password Security Policies](#password-security-policies)
3. [JWT Token Management](#jwt-token-management)
4. [Account Lockout Protection](#account-lockout-protection)
5. [User Management](#user-management)
6. [Role-Based Access Control](#role-based-access-control)
7. [Configuration](#configuration)
8. [API Reference](#api-reference)
9. [Testing](#testing)
10. [Security Best Practices](#security-best-practices)

## Architecture

The authentication system consists of several key components:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Password      │    │   JWT Token     │    │   Account       │
│   Policy        │    │   Management    │    │   Lockout       │
│   Validator     │    │                 │    │   Protection    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────────┐
         │              AuthService                            │
         │  - User Registration & Authentication               │
         │  - Password Validation & Management                 │
         │  - Token Lifecycle Management                       │
         │  - Role & Permission Queries                        │
         └─────────────────────────────────────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────────┐
         │              Database Layer                         │
         │  - Users, Roles, Permissions                        │
         │  - Token Storage & Revocation                       │
         │  - Audit Logs                                       │
         └─────────────────────────────────────────────────────┘
```

### Core Components

#### 1. **PasswordValidator** (`internal/auth/password_policy.go`)
- Comprehensive password complexity validation
- Real-time strength scoring (0-100 scale)
- Pattern detection (sequential chars, common passwords, username inclusion)
- Configurable policies for different environments

#### 2. **JWTManager** (`internal/auth/jwt.go`)
- Secure JWT token generation and validation
- Access token (short-lived) and refresh token (long-lived) pairs
- Token revocation and cleanup
- Claims-based user identity and permissions

#### 3. **AuthService** (`internal/auth/service.go`)
- User registration, login, and password management
- Integration with password policies and JWT tokens
- Account lockout and security event logging
- Database integration for persistent storage

#### 4. **TokenStorage** (`internal/auth/storage.go`)
- PostgreSQL-backed token storage
- Refresh token lifecycle management
- Secure token revocation and cleanup

## Password Security Policies

### Default Policy Configuration

The default military-grade password policy enforces:

```go
MinLength: 12                    // Minimum 12 characters
MaxLength: 128                   // Maximum 128 characters
RequireUppercase: true           // At least one A-Z
RequireLowercase: true           // At least one a-z  
RequireNumbers: true             // At least one 0-9
RequireSpecialChars: true        // At least one special character
MinUniqueChars: 8               // Minimum 8 unique characters
MaxRepeatingChars: 2            // Maximum 2 consecutive identical chars
ForbidSequentialChars: true     // Block abc, 123, qwerty patterns
ForbidCommonPasswords: true     // Block password, admin, etc.
ForbidUsernameInPassword: true  // Username cannot appear in password
PasswordMaxAge: 90 days         // Password expires after 90 days
PasswordHistorySize: 5          // Last 5 passwords cannot be reused
MaxFailedAttempts: 5            // Lock after 5 failed attempts
LockoutDuration: 15 minutes     // Account locked for 15 minutes
```

### Pattern Detection

The system detects and blocks various weak password patterns:

**Sequential Characters:**
- Alphabetic: `abc`, `def`, `xyz`
- Numeric: `123`, `456`, `789`
- Keyboard: `qwe`, `asd`, `zxc`
- Reversed: `cba`, `321`, `ewq`

**Common Passwords:**
- Dictionary words: `password`, `admin`, `user`
- Common variations: `password123`, `admin456`, `Password1`
- Base words with numbers: `test123`, `demo456`

**Username Patterns:**
- Direct inclusion: password contains username
- Reversed inclusion: password contains reversed username
- Case-insensitive detection

### Password Strength Scoring

The system provides real-time password strength scoring on a 0-100 scale:

- **0-29: Very Weak** - Missing multiple requirements
- **30-49: Weak** - Basic requirements met, patterns detected
- **50-69: Moderate** - Good complexity, room for improvement
- **70-89: Strong** - Excellent complexity and length
- **90-100: Very Strong** - Maximum security, ideal for sensitive accounts

Scoring factors:
- Length (up to 25 points)
- Character diversity (up to 40 points)  
- Unique character ratio (up to 15 points)
- Bonus for exceptional length/complexity (up to 20 points)
- Penalties for weak patterns (up to -20 points)

### Custom Policies

Password policies can be customized for different environments:

```go
// Lenient policy for development
lenientPolicy := PasswordPolicy{
    MinLength: 8,
    RequireUppercase: false,
    RequireNumbers: false,
    RequireSpecialChars: false,
    ForbidSequentialChars: false,
    MaxFailedAttempts: 10,
}

// Strict policy for high-security environments
strictPolicy := PasswordPolicy{
    MinLength: 16,
    MinUniqueChars: 12,
    MaxRepeatingChars: 1,
    PasswordMaxAge: 30 * 24 * time.Hour, // 30 days
    MaxFailedAttempts: 3,
    LockoutDuration: 30 * time.Minute,
}
```

## JWT Token Management

### Token Architecture

GoTAK uses a dual-token system for secure authentication:

**Access Tokens:**
- Short-lived (15 minutes default)
- Contains user identity and permissions
- Used for API authorization
- Automatically expires for security

**Refresh Tokens:**
- Long-lived (24 hours default)
- Used to generate new access tokens
- Stored securely in database
- Can be revoked instantly

### Token Claims

JWT tokens contain comprehensive user information:

```json
{
  "sub": "user-uuid",           // User ID
  "username": "johndoe",        // Username
  "roles": ["admin", "user"],   // User roles
  "permissions": [              // User permissions
    "system.admin",
    "user.create",
    "user.read"
  ],
  "token_type": "access",       // access|refresh
  "iat": 1640995200,           // Issued at
  "exp": 1640995200,           // Expires at
  "jti": "token-uuid"          // Token ID for revocation
}
```

### Token Lifecycle

1. **Generation**: User authenticates → JWT pair created → Refresh token stored
2. **Usage**: Access token validates API requests → Claims extracted
3. **Refresh**: Access token expires → Refresh token generates new pair
4. **Revocation**: User logout/password change → All tokens invalidated
5. **Cleanup**: Expired tokens automatically removed

### Security Features

- **Secure signing**: RSA-256 or HMAC-SHA256 algorithms
- **Token revocation**: Instant invalidation via database blacklist
- **Automatic cleanup**: Expired tokens removed from storage
- **Replay protection**: Unique token IDs prevent reuse
- **Scope limitation**: Tokens contain minimal necessary claims

## Account Lockout Protection

### Lockout Logic

The system protects against brute force attacks with progressive lockout:

1. Track failed login attempts per user
2. Lock account after threshold reached (5 attempts default)
3. Lockout duration increases with repeated violations
4. Automatic unlock after timeout period
5. Reset counter on successful authentication

### Lockout Information

Administrators can query detailed lockout status:

```go
lockInfo, err := authService.GetAccountLockInfo(userID)
// Returns:
// - Current failed attempts
// - Maximum attempts allowed
// - Remaining attempts before lockout
// - Lock status and expiration time
```

### Progressive Lockout

When enabled, lockout durations increase with repeated violations:
- 1st lockout: 15 minutes
- 2nd lockout: 30 minutes  
- 3rd lockout: 1 hour
- 4th+ lockout: 24 hours

### Manual Override

Administrators can manually unlock accounts:

```go
err := authService.unlockAccount(userID)
```

## User Management

### User Registration

New user registration with comprehensive validation:

```go
registerReq := &RegisterRequest{
    Username:  "johndoe",
    Email:     "john@example.com",
    Password:  "SecureP@ssw4rd97!",
    FirstName: &firstName,
    LastName:  &lastName,
}

user, err := authService.RegisterUser(registerReq)
```

Validation includes:
- Username uniqueness and format
- Email format and uniqueness  
- Password policy compliance
- Optional profile information

### User Authentication

Login process with security checks:

```go
loginReq := &LoginRequest{
    Username:   "johndoe",
    Password:   "SecureP@ssw4rd97!",
    AuthMethod: "local",
    MFACode:    "123456", // If MFA enabled
}

tokenPair, err := authService.Authenticate(loginReq)
```

Authentication flow:
1. User existence verification
2. Account status check (active, not locked)
3. Password verification
4. MFA validation (if enabled)
5. Token generation
6. Failed attempt reset
7. Security event logging

### Password Management

Secure password change with validation:

```go
err := authService.ChangePassword(userID, currentPassword, newPassword)
```

Password change process:
1. Current password verification
2. New password policy validation
3. Password history check (prevent reuse)
4. Secure hash generation
5. Database update
6. Token revocation (force re-auth)
7. Audit logging

## Role-Based Access Control

### Role Hierarchy

GoTAK implements a flexible RBAC system:

```
System Administrator
├── User Manager
│   ├── User Creator
│   └── User Reader
├── Configuration Manager
└── Audit Viewer

Operator
├── Position Reporter
└── Message Sender

Standard User
└── Basic Access
```

### Permission System

Permissions follow a resource.action pattern:

```
system.admin          # Full system access
user.create          # Create users
user.read            # Read user information  
user.update          # Update users
user.delete          # Delete users
config.read          # Read configuration
config.update        # Update configuration
message.send         # Send messages
position.report      # Report positions
audit.read           # View audit logs
```

### Role Assignment

Users can have multiple roles with combined permissions:

```sql
-- Example role assignments
INSERT INTO user_roles (user_id, role_id) VALUES
  ('user-uuid', 'admin-role-uuid'),
  ('user-uuid', 'operator-role-uuid');
```

### Permission Checks

The system validates permissions for all operations:

```go
// In JWT middleware
if !hasPermission(claims.Permissions, "user.create") {
    return ErrInsufficientPermissions
}
```

## Configuration

### Server Configuration

Configure authentication in `config/server.yaml`:

```yaml
auth:
  # Password Policy
  min_password_length: 12
  max_password_length: 128
  require_uppercase: true
  require_lowercase: true
  require_numbers: true
  require_special_chars: true
  
  # Account Lockout
  max_failed_attempts: 5
  lockout_duration: "15m"
  
  # JWT Configuration
  jwt:
    secret_key: "${JWT_SECRET_KEY}"
    access_ttl: "15m"
    refresh_ttl: "24h"
    issuer: "gotak-server"
    
  # Password Security
  bcrypt_cost: 12
```

### Environment Variables

Security-sensitive values use environment variables:

```bash
# Required
JWT_SECRET_KEY="your-secure-secret-key-here"
DB_PASSWORD="your-database-password"

# Optional
BCRYPT_COST=12
TOKEN_CLEANUP_INTERVAL="1h"
```

### Database Configuration

PostgreSQL connection for authentication storage:

```yaml
database:
  host: "localhost"
  port: 5432
  dbname: "gotak"
  user: "gotak"
  password: "${DB_PASSWORD}"
  sslmode: "require"
```

## API Reference

### Authentication Endpoints

#### POST `/auth/register`
Register a new user account.

**Request:**
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "SecureP@ssw4rd97!",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "id": "user-uuid",
  "username": "johndoe",
  "email": "john@example.com",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### POST `/auth/login`
Authenticate user and return JWT tokens.

**Request:**
```json
{
  "username": "johndoe",
  "password": "SecureP@ssw4rd97!",
  "auth_method": "local"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_at": "2024-01-01T00:15:00Z",
  "user": {
    "id": "user-uuid",
    "username": "johndoe",
    "email": "john@example.com",
    "roles": ["user"]
  }
}
```

#### POST `/auth/refresh`
Refresh access token using refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_at": "2024-01-01T00:30:00Z"
}
```

#### POST `/auth/logout`
Revoke user tokens and logout.

**Headers:**
```
Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
```

**Response:**
```json
{
  "message": "Successfully logged out"
}
```

#### POST `/auth/change-password`
Change user password with validation.

**Request:**
```json
{
  "current_password": "OldP@ssw4rd97!",
  "new_password": "NewSecureP@ssw4rd24!"
}
```

**Response:**
```json
{
  "message": "Password changed successfully"
}
```

### Password Validation Endpoints

#### POST `/auth/validate-password`
Validate password strength and policy compliance.

**Request:**
```json
{
  "password": "TestP@ssw4rd97!",
  "username": "johndoe"
}
```

**Response:**
```json
{
  "is_valid": true,
  "score": 85,
  "level": "Strong",
  "validation_errors": [],
  "suggestions": [],
  "checks": {
    "min_length": true,
    "has_uppercase": true,
    "has_lowercase": true,
    "has_numbers": true,
    "has_special": true,
    "unique_chars": true,
    "no_username": true,
    "no_repeating": true,
    "no_sequential": true,
    "not_common": true
  }
}
```

### User Management Endpoints

#### GET `/users/me`
Get current user profile.

**Headers:**
```
Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
```

**Response:**
```json
{
  "id": "user-uuid",
  "username": "johndoe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "is_active": true,
  "mfa_enabled": false,
  "roles": ["user"],
  "permissions": ["user.read", "position.report"],
  "last_login": "2024-01-01T12:00:00Z",
  "password_expires_at": "2024-04-01T00:00:00Z"
}
```

#### GET `/users/{id}/lock-info`
Get account lockout information (admin only).

**Response:**
```json
{
  "user_id": "user-uuid",
  "username": "johndoe",
  "failed_attempts": 2,
  "max_attempts": 5,
  "remaining_attempts": 3,
  "is_locked": false,
  "locked_until": null
}
```

## Testing

### Unit Tests

The authentication system includes comprehensive unit tests:

```bash
# Run all auth tests
go test ./internal/auth -v

# Run specific test suites
go test ./internal/auth -run TestJWTManager
go test ./internal/auth -run TestPasswordValidator
go test ./internal/auth -run TestAuthService
```

### Test Coverage

Current test coverage includes:

- **JWT Token Management**: Token generation, validation, refresh, revocation
- **Password Policies**: All validation rules, strength scoring, pattern detection
- **Authentication Service**: User registration, login, password changes
- **Account Lockout**: Failed attempts, lockout logic, automatic unlock
- **Helper Functions**: String manipulation, character detection, pattern matching

### Integration Tests

Integration tests require a PostgreSQL database:

```bash
# Run integration tests (requires database)
go test -tags=integration ./internal/auth -v

# Set up test database
createdb gotak_test
psql -d gotak_test -f migrations/schema.sql
```

### Performance Tests

Load testing for high-concurrency scenarios:

```bash
# Password validation performance
go test -bench=BenchmarkPasswordValidation ./internal/auth

# Token generation performance  
go test -bench=BenchmarkTokenGeneration ./internal/auth
```

### Demo Application

Interactive demonstration of password policies:

```bash
# Run password policy demonstration
go run examples/password_policy_demo.go
```

This demo showcases:
- Policy configuration examples
- Real-time password validation
- Strength scoring and feedback
- Custom policy scenarios
- Security best practices

## Security Best Practices

### Password Security

1. **Use Strong Defaults**: Military-grade 12+ character requirements
2. **Enable All Validations**: Character types, patterns, common passwords
3. **Regular Updates**: Force password changes every 90 days
4. **History Prevention**: Block reuse of last 5 passwords
5. **Account Lockout**: Enable progressive lockout protection

### Token Security

1. **Short Access Tokens**: 15-minute expiration maximum
2. **Secure Storage**: Store refresh tokens securely in database
3. **Regular Cleanup**: Remove expired tokens automatically
4. **Immediate Revocation**: Revoke tokens on logout/password change
5. **Unique Secrets**: Use strong, unique JWT signing keys

### Database Security

1. **Encrypted Storage**: Use bcrypt for password hashing (cost 12+)
2. **Secure Connections**: Always use TLS for database connections
3. **Access Control**: Limit database permissions to minimum required
4. **Audit Logging**: Log all authentication events
5. **Regular Backups**: Maintain secure, encrypted backups

### Network Security

1. **HTTPS Only**: Never transmit credentials over HTTP
2. **Secure Headers**: Implement security headers (HSTS, CSP, etc.)
3. **Rate Limiting**: Protect endpoints from abuse
4. **IP Restrictions**: Limit admin access to known networks
5. **Monitoring**: Monitor for suspicious authentication patterns

### Operational Security

1. **Environment Variables**: Store secrets in environment variables
2. **Log Security**: Ensure logs don't contain sensitive information
3. **Error Handling**: Don't leak system information in error messages
4. **Updates**: Keep dependencies and system components updated
5. **Incident Response**: Have procedures for security incidents

### Compliance Considerations

The GoTAK authentication system supports various compliance requirements:

- **FIPS 140-2**: Cryptographic standards for government systems
- **Common Criteria**: Security evaluation standards
- **NIST 800-63B**: Digital identity guidelines
- **DoD 8500/8570**: Information assurance requirements
- **GDPR/CCPA**: Privacy and data protection regulations

### Deployment Checklist

Before deploying to production:

- [ ] Generate strong JWT secret keys
- [ ] Configure TLS certificates
- [ ] Set up secure database connections
- [ ] Enable audit logging
- [ ] Configure backup procedures
- [ ] Test account lockout functionality
- [ ] Verify password policy enforcement
- [ ] Review security headers
- [ ] Test token revocation
- [ ] Monitor authentication metrics

---

This authentication system provides enterprise-grade security suitable for military, government, and high-security commercial environments. The comprehensive password policies, token management, and account protection features ensure robust defense against common attack vectors while maintaining usability for legitimate users.
