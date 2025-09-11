# Sprint 10: Security & Compliance Framework

**Duration:** 2 weeks  
**Theme:** Enterprise Security & Regulatory Compliance  
**Sprint Goals:** Implement comprehensive security framework meeting government and enterprise compliance requirements

## Objectives

1. **Advanced Authentication**: Multi-factor authentication and certificate-based authentication
2. **Authorization Framework**: Role-based access control with fine-grained permissions
3. **Data Protection**: Encryption at rest and in transit with key management
4. **Compliance Standards**: FISMA, NIST, and other regulatory compliance
5. **Security Monitoring**: Intrusion detection and security event monitoring

## User Stories

### Epic: Enterprise Security & Compliance

**US-10.1: Multi-Factor Authentication**
```
As a security administrator
I want users to authenticate using multiple factors
So that account security is enhanced beyond just passwords
```

**Acceptance Criteria:**
- Support for TOTP (Time-based One-Time Password) authentication
- SMS and email-based second factor options
- Hardware token support (FIDO2/WebAuthn)
- Backup codes for account recovery
- Administrative enforcement of MFA policies

**US-10.2: Certificate-Based Authentication**
```
As a government user
I want to authenticate using my CAC/PIV card
So that I can use standard government authentication methods
```

**Acceptance Criteria:**
- X.509 certificate authentication support
- CAC/PIV card integration
- Certificate validation and revocation checking
- Mutual TLS (mTLS) for client authentication
- Certificate-to-user mapping

**US-10.3: Fine-Grained Access Control**
```
As a security officer
I want to control exactly what users can access and do
So that data access follows the principle of least privilege
```

**Acceptance Criteria:**
- Role-based access control (RBAC) system
- Attribute-based access control (ABAC) for complex policies
- Resource-level permissions
- Dynamic permission evaluation
- Policy management interface

**US-10.4: Data Encryption and Key Management**
```
As a compliance officer
I want all sensitive data encrypted with proper key management
So that data is protected according to security standards
```

**Acceptance Criteria:**
- AES-256 encryption for data at rest
- TLS 1.3 for data in transit
- Hardware Security Module (HSM) integration
- Key rotation and lifecycle management
- Secure key distribution for federation

**US-10.5: Security Monitoring and Alerting**
```
As a security operations center analyst
I want real-time security monitoring and alerting
So that I can detect and respond to security threats quickly
```

**Acceptance Criteria:**
- Security event correlation and analysis
- Intrusion detection and prevention
- Anomaly detection for user behavior
- Real-time security alerting
- Integration with SIEM systems

## Technical Implementation

### Multi-Factor Authentication

**MFA Manager**
```go
// pkg/auth/mfa/manager.go
package mfa

import (
    "context"
    "crypto/rand"
    "encoding/base32"
    "fmt"
    "time"
    
    "github.com/pquerna/otp/totp"
    "github.com/skip2/go-qrcode"
    "github.com/google/uuid"
)

type MFAManager struct {
    storage     MFAStorage
    smsProvider SMSProvider
    emailProvider EmailProvider
    config      *Config
    logger      Logger
}

type Config struct {
    TOTPIssuer        string        `yaml:"totp_issuer"`
    TOTPSkew          int           `yaml:"totp_skew"`
    BackupCodeLength  int           `yaml:"backup_code_length"`
    BackupCodeCount   int           `yaml:"backup_code_count"`
    
    // SMS settings
    SMSEnabled        bool          `yaml:"sms_enabled"`
    SMSProvider       string        `yaml:"sms_provider"`
    SMSTemplate       string        `yaml:"sms_template"`
    
    // Email settings
    EmailEnabled      bool          `yaml:"email_enabled"`
    EmailTemplate     string        `yaml:"email_template"`
    
    // Security settings
    MaxAttempts       int           `yaml:"max_attempts"`
    LockoutDuration   time.Duration `yaml:"lockout_duration"`
    CodeLifetime      time.Duration `yaml:"code_lifetime"`
}

type MFAMethod string
const (
    MFAMethodTOTP      MFAMethod = "totp"
    MFAMethodSMS       MFAMethod = "sms"
    MFAMethodEmail     MFAMethod = "email"
    MFAMethodBackup    MFAMethod = "backup"
    MFAMethodWebAuthn  MFAMethod = "webauthn"
)

type UserMFA struct {
    UserID        uuid.UUID    `json:"user_id"`
    Method        MFAMethod    `json:"method"`
    Secret        string       `json:"secret,omitempty"`
    PhoneNumber   string       `json:"phone_number,omitempty"`
    Email         string       `json:"email,omitempty"`
    BackupCodes   []string     `json:"backup_codes,omitempty"`
    Enabled       bool         `json:"enabled"`
    CreatedAt     time.Time    `json:"created_at"`
    LastUsed      time.Time    `json:"last_used,omitempty"`
}

type MFAChallenge struct {
    ID            uuid.UUID    `json:"id"`
    UserID        uuid.UUID    `json:"user_id"`
    Method        MFAMethod    `json:"method"`
    Code          string       `json:"code,omitempty"`
    ExpiresAt     time.Time    `json:"expires_at"`
    Attempts      int          `json:"attempts"`
    Verified      bool         `json:"verified"`
    CreatedAt     time.Time    `json:"created_at"`
}

func NewMFAManager(storage MFAStorage, config *Config, logger Logger) *MFAManager {
    return &MFAManager{
        storage: storage,
        config:  config,
        logger:  logger,
    }
}

func (m *MFAManager) SetupTOTP(ctx context.Context, userID uuid.UUID, username string) (*TOTPSetup, error) {
    // Generate secret key
    key := make([]byte, 20)
    if _, err := rand.Read(key); err != nil {
        return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
    }
    
    secret := base32.StdEncoding.EncodeToString(key)
    
    // Create TOTP URL
    url, err := totp.Generate(totp.GenerateOpts{
        Issuer:      m.config.TOTPIssuer,
        AccountName: username,
        Secret:      key,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to generate TOTP URL: %w", err)
    }
    
    // Generate QR code
    qrCode, err := qrcode.Encode(url.String(), qrcode.Medium, 256)
    if err != nil {
        return nil, fmt.Errorf("failed to generate QR code: %w", err)
    }
    
    // Store MFA setup (not yet enabled)
    userMFA := &UserMFA{
        UserID:    userID,
        Method:    MFAMethodTOTP,
        Secret:    secret,
        Enabled:   false,
        CreatedAt: time.Now(),
    }
    
    if err := m.storage.CreateMFA(ctx, userMFA); err != nil {
        return nil, fmt.Errorf("failed to store MFA setup: %w", err)
    }
    
    return &TOTPSetup{
        Secret:     secret,
        QRCode:     qrCode,
        BackupURL:  url.String(),
        BackupCodes: m.generateBackupCodes(),
    }, nil
}

func (m *MFAManager) VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) error {
    userMFA, err := m.storage.GetMFAByUser(ctx, userID, MFAMethodTOTP)
    if err != nil {
        return fmt.Errorf("failed to get user MFA: %w", err)
    }
    
    if !userMFA.Enabled {
        return ErrMFANotEnabled
    }
    
    // Validate TOTP code
    valid := totp.Validate(code, userMFA.Secret)
    if !valid {
        // Try with skew
        for i := 1; i <= m.config.TOTPSkew; i++ {
            pastValid := totp.ValidateCustom(code, userMFA.Secret, time.Now().Add(-time.Duration(i)*30*time.Second), totp.ValidateOpts{})
            futureValid := totp.ValidateCustom(code, userMFA.Secret, time.Now().Add(time.Duration(i)*30*time.Second), totp.ValidateOpts{})
            
            if pastValid || futureValid {
                valid = true
                break
            }
        }
    }
    
    if !valid {
        return ErrInvalidMFACode
    }
    
    // Update last used timestamp
    userMFA.LastUsed = time.Now()
    if err := m.storage.UpdateMFA(ctx, userMFA); err != nil {
        m.logger.Warn("Failed to update MFA last used", "error", err)
    }
    
    return nil
}

func (m *MFAManager) SendSMSChallenge(ctx context.Context, userID uuid.UUID) (*MFAChallenge, error) {
    userMFA, err := m.storage.GetMFAByUser(ctx, userID, MFAMethodSMS)
    if err != nil {
        return nil, fmt.Errorf("failed to get user SMS MFA: %w", err)
    }
    
    if !userMFA.Enabled || userMFA.PhoneNumber == "" {
        return nil, ErrSMSMFANotEnabled
    }
    
    // Generate 6-digit code
    code := m.generateNumericCode(6)
    
    challenge := &MFAChallenge{
        ID:        uuid.New(),
        UserID:    userID,
        Method:    MFAMethodSMS,
        Code:      code,
        ExpiresAt: time.Now().Add(m.config.CodeLifetime),
        CreatedAt: time.Now(),
    }
    
    // Store challenge
    if err := m.storage.CreateChallenge(ctx, challenge); err != nil {
        return nil, fmt.Errorf("failed to store SMS challenge: %w", err)
    }
    
    // Send SMS
    message := fmt.Sprintf(m.config.SMSTemplate, code)
    if err := m.smsProvider.SendSMS(userMFA.PhoneNumber, message); err != nil {
        return nil, fmt.Errorf("failed to send SMS: %w", err)
    }
    
    // Don't return the actual code in the response
    challenge.Code = ""
    return challenge, nil
}

func (m *MFAManager) VerifyBackupCode(ctx context.Context, userID uuid.UUID, code string) error {
    userMFA, err := m.storage.GetMFAByUser(ctx, userID, MFAMethodTOTP)
    if err != nil {
        return fmt.Errorf("failed to get user MFA: %w", err)
    }
    
    // Find and validate backup code
    for i, backupCode := range userMFA.BackupCodes {
        if backupCode == code {
            // Remove used backup code
            userMFA.BackupCodes = append(userMFA.BackupCodes[:i], userMFA.BackupCodes[i+1:]...)
            userMFA.LastUsed = time.Now()
            
            if err := m.storage.UpdateMFA(ctx, userMFA); err != nil {
                return fmt.Errorf("failed to update MFA after backup code use: %w", err)
            }
            
            return nil
        }
    }
    
    return ErrInvalidBackupCode
}

type TOTPSetup struct {
    Secret      string   `json:"secret"`
    QRCode      []byte   `json:"qr_code"`
    BackupURL   string   `json:"backup_url"`
    BackupCodes []string `json:"backup_codes"`
}

func (m *MFAManager) generateBackupCodes() []string {
    codes := make([]string, m.config.BackupCodeCount)
    for i := range codes {
        codes[i] = m.generateAlphanumericCode(m.config.BackupCodeLength)
    }
    return codes
}
```

### Certificate-Based Authentication

**Certificate Authentication Manager**
```go
// pkg/auth/cert/manager.go
package cert

import (
    "context"
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "time"
    
    "github.com/google/uuid"
)

type CertAuthManager struct {
    storage      CertStorage
    validator    *CertValidator
    config       *Config
    logger       Logger
}

type Config struct {
    TrustedCAs           []string      `yaml:"trusted_cas"`
    RequireClientCert    bool          `yaml:"require_client_cert"`
    CRLCheckEnabled      bool          `yaml:"crl_check_enabled"`
    OCSPCheckEnabled     bool          `yaml:"ocsp_check_enabled"`
    CertCacheTimeout     time.Duration `yaml:"cert_cache_timeout"`
    
    // CAC/PIV specific
    CACEnabled           bool          `yaml:"cac_enabled"`
    PIVEnabled           bool          `yaml:"piv_enabled"`
    RequireHardwareToken bool          `yaml:"require_hardware_token"`
    
    // Certificate field mapping
    UserIDField          string        `yaml:"user_id_field"`
    UsernameField        string        `yaml:"username_field"`
    EmailField           string        `yaml:"email_field"`
    RoleField            string        `yaml:"role_field"`
}

type UserCertificate struct {
    UserID           uuid.UUID         `json:"user_id"`
    Serial           string            `json:"serial"`
    Subject          string            `json:"subject"`
    Issuer           string            `json:"issuer"`
    Fingerprint      string            `json:"fingerprint"`
    ValidFrom        time.Time         `json:"valid_from"`
    ValidTo          time.Time         `json:"valid_to"`
    KeyUsage         []string          `json:"key_usage"`
    ExtendedKeyUsage []string          `json:"extended_key_usage"`
    CertificateData  []byte            `json:"certificate_data"`
    Status           CertStatus        `json:"status"`
    CreatedAt        time.Time         `json:"created_at"`
    UpdatedAt        time.Time         `json:"updated_at"`
}

type CertStatus string
const (
    CertStatusActive   CertStatus = "active"
    CertStatusRevoked  CertStatus = "revoked"
    CertStatusExpired  CertStatus = "expired"
    CertStatusSuspended CertStatus = "suspended"
)

type CertValidator struct {
    trustedCAs   []*x509.Certificate
    crlCache     map[string]*x509.RevocationList
    ocspCache    map[string]*OCSPResponse
    config       *Config
    logger       Logger
}

func NewCertAuthManager(storage CertStorage, config *Config, logger Logger) (*CertAuthManager, error) {
    validator, err := NewCertValidator(config, logger)
    if err != nil {
        return nil, fmt.Errorf("failed to create certificate validator: %w", err)
    }
    
    return &CertAuthManager{
        storage:   storage,
        validator: validator,
        config:    config,
        logger:    logger,
    }, nil
}

func (cam *CertAuthManager) AuthenticateWithCertificate(ctx context.Context, certPEM []byte) (*AuthResult, error) {
    // Parse certificate
    block, _ := pem.Decode(certPEM)
    if block == nil {
        return nil, ErrInvalidCertificateFormat
    }
    
    cert, err := x509.ParseCertificate(block.Bytes)
    if err != nil {
        return nil, fmt.Errorf("failed to parse certificate: %w", err)
    }
    
    // Validate certificate
    if err := cam.validator.ValidateCertificate(ctx, cert); err != nil {
        return nil, fmt.Errorf("certificate validation failed: %w", err)
    }
    
    // Extract user information from certificate
    userInfo, err := cam.extractUserInfo(cert)
    if err != nil {
        return nil, fmt.Errorf("failed to extract user info from certificate: %w", err)
    }
    
    // Check if certificate is registered
    userCert, err := cam.storage.GetCertificateByFingerprint(ctx, userInfo.Fingerprint)
    if err != nil {
        if err == ErrCertificateNotFound {
            // Auto-register certificate if configured
            if cam.config.AutoRegisterCerts {
                userCert, err = cam.registerCertificate(ctx, cert, userInfo)
                if err != nil {
                    return nil, fmt.Errorf("failed to auto-register certificate: %w", err)
                }
            } else {
                return nil, ErrCertificateNotRegistered
            }
        } else {
            return nil, fmt.Errorf("failed to lookup certificate: %w", err)
        }
    }
    
    // Check certificate status
    if userCert.Status != CertStatusActive {
        return nil, fmt.Errorf("certificate status is %s", userCert.Status)
    }
    
    // Update last used
    userCert.UpdatedAt = time.Now()
    if err := cam.storage.UpdateCertificate(ctx, userCert); err != nil {
        cam.logger.Warn("Failed to update certificate last used", "error", err)
    }
    
    return &AuthResult{
        UserID:      userCert.UserID,
        Username:    userInfo.Username,
        Email:       userInfo.Email,
        Roles:       userInfo.Roles,
        AuthMethod:  "certificate",
        CertSerial:  userCert.Serial,
        Fingerprint: userCert.Fingerprint,
    }, nil
}

func (cv *CertValidator) ValidateCertificate(ctx context.Context, cert *x509.Certificate) error {
    // Check certificate validity period
    now := time.Now()
    if now.Before(cert.NotBefore) {
        return ErrCertificateNotYetValid
    }
    if now.After(cert.NotAfter) {
        return ErrCertificateExpired
    }
    
    // Verify certificate chain
    roots := x509.NewCertPool()
    for _, caCert := range cv.trustedCAs {
        roots.AddCert(caCert)
    }
    
    opts := x509.VerifyOptions{
        Roots: roots,
    }
    
    if _, err := cert.Verify(opts); err != nil {
        return fmt.Errorf("certificate chain verification failed: %w", err)
    }
    
    // Check CRL if enabled
    if cv.config.CRLCheckEnabled {
        if err := cv.checkCRL(ctx, cert); err != nil {
            return fmt.Errorf("CRL check failed: %w", err)
        }
    }
    
    // Check OCSP if enabled
    if cv.config.OCSPCheckEnabled {
        if err := cv.checkOCSP(ctx, cert); err != nil {
            return fmt.Errorf("OCSP check failed: %w", err)
        }
    }
    
    return nil
}

func (cam *CertAuthManager) RegisterCertificate(ctx context.Context, userID uuid.UUID, certPEM []byte) error {
    // Parse and validate certificate
    block, _ := pem.Decode(certPEM)
    if block == nil {
        return ErrInvalidCertificateFormat
    }
    
    cert, err := x509.ParseCertificate(block.Bytes)
    if err != nil {
        return fmt.Errorf("failed to parse certificate: %w", err)
    }
    
    // Validate certificate
    if err := cam.validator.ValidateCertificate(ctx, cert); err != nil {
        return fmt.Errorf("certificate validation failed: %w", err)
    }
    
    // Create user certificate record
    userCert := &UserCertificate{
        UserID:          userID,
        Serial:          cert.SerialNumber.String(),
        Subject:         cert.Subject.String(),
        Issuer:          cert.Issuer.String(),
        Fingerprint:     calculateFingerprint(cert),
        ValidFrom:       cert.NotBefore,
        ValidTo:         cert.NotAfter,
        KeyUsage:        keyUsageToStrings(cert.KeyUsage),
        ExtendedKeyUsage: extKeyUsageToStrings(cert.ExtKeyUsage),
        CertificateData: cert.Raw,
        Status:          CertStatusActive,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }
    
    // Store certificate
    if err := cam.storage.CreateCertificate(ctx, userCert); err != nil {
        return fmt.Errorf("failed to store certificate: %w", err)
    }
    
    cam.logger.Info("Certificate registered", 
        "user_id", userID, 
        "serial", userCert.Serial,
        "subject", userCert.Subject)
    
    return nil
}
```

### Role-Based Access Control (RBAC)

**RBAC Manager**
```go
// pkg/auth/rbac/manager.go
package rbac

import (
    "context"
    "fmt"
    "strings"
    "sync"
    "time"
    
    "github.com/google/uuid"
)

type RBACManager struct {
    storage     RBACStorage
    evaluator   *PolicyEvaluator
    cache       *PermissionCache
    config      *Config
    logger      Logger
    mu          sync.RWMutex
}

type Config struct {
    DefaultRole         string        `yaml:"default_role"`
    CacheTimeout        time.Duration `yaml:"cache_timeout"`
    HierarchicalRoles   bool          `yaml:"hierarchical_roles"`
    AttributeBasedRules bool          `yaml:"attribute_based_rules"`
    PolicyLanguage      string        `yaml:"policy_language"`
}

type Role struct {
    ID          uuid.UUID   `json:"id"`
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Permissions []string    `json:"permissions"`
    ParentRoles []uuid.UUID `json:"parent_roles,omitempty"`
    Attributes  map[string]interface{} `json:"attributes,omitempty"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
}

type Permission struct {
    ID          uuid.UUID `json:"id"`
    Resource    string    `json:"resource"`
    Action      string    `json:"action"`
    Scope       string    `json:"scope,omitempty"`
    Conditions  []string  `json:"conditions,omitempty"`
    Description string    `json:"description"`
}

type UserRole struct {
    UserID      uuid.UUID              `json:"user_id"`
    RoleID      uuid.UUID              `json:"role_id"`
    Scope       string                 `json:"scope,omitempty"`
    Conditions  map[string]interface{} `json:"conditions,omitempty"`
    ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
    AssignedBy  uuid.UUID              `json:"assigned_by"`
    AssignedAt  time.Time              `json:"assigned_at"`
}

type AccessRequest struct {
    UserID     uuid.UUID              `json:"user_id"`
    Resource   string                 `json:"resource"`
    Action     string                 `json:"action"`
    Scope      string                 `json:"scope,omitempty"`
    Context    map[string]interface{} `json:"context,omitempty"`
    Timestamp  time.Time              `json:"timestamp"`
}

type AccessDecision struct {
    Allowed    bool                   `json:"allowed"`
    Reason     string                 `json:"reason,omitempty"`
    AppliedRoles []string             `json:"applied_roles,omitempty"`
    AppliedPermissions []string       `json:"applied_permissions,omitempty"`
    Conditions map[string]interface{} `json:"conditions,omitempty"`
    TTL        time.Duration          `json:"ttl,omitempty"`
}

func NewRBACManager(storage RBACStorage, config *Config, logger Logger) *RBACManager {
    return &RBACManager{
        storage:   storage,
        evaluator: NewPolicyEvaluator(config),
        cache:     NewPermissionCache(config.CacheTimeout),
        config:    config,
        logger:    logger,
    }
}

func (rm *RBACManager) CheckAccess(ctx context.Context, req *AccessRequest) (*AccessDecision, error) {
    // Check cache first
    cacheKey := rm.buildCacheKey(req)
    if decision := rm.cache.Get(cacheKey); decision != nil {
        return decision, nil
    }
    
    // Get user roles
    userRoles, err := rm.storage.GetUserRoles(ctx, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user roles: %w", err)
    }
    
    // Expand roles (include parent roles if hierarchical)
    expandedRoles, err := rm.expandRoles(ctx, userRoles)
    if err != nil {
        return nil, fmt.Errorf("failed to expand roles: %w", err)
    }
    
    // Collect all permissions from roles
    allPermissions := make(map[string]*Permission)
    appliedRoles := make([]string, 0)
    
    for _, role := range expandedRoles {
        appliedRoles = append(appliedRoles, role.Name)
        
        rolePermissions, err := rm.storage.GetRolePermissions(ctx, role.ID)
        if err != nil {
            rm.logger.Warn("Failed to get role permissions", "role_id", role.ID, "error", err)
            continue
        }
        
        for _, perm := range rolePermissions {
            permKey := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
            allPermissions[permKey] = perm
        }
    }
    
    // Evaluate permissions
    decision := &AccessDecision{
        Allowed:            false,
        AppliedRoles:       appliedRoles,
        AppliedPermissions: make([]string, 0),
    }
    
    for _, perm := range allPermissions {
        if rm.matchesResourceAndAction(perm, req.Resource, req.Action) {
            decision.AppliedPermissions = append(decision.AppliedPermissions, 
                fmt.Sprintf("%s:%s", perm.Resource, perm.Action))
            
            // Check conditions
            if len(perm.Conditions) > 0 {
                conditionsMet, err := rm.evaluator.EvaluateConditions(perm.Conditions, req.Context)
                if err != nil {
                    rm.logger.Warn("Failed to evaluate conditions", "permission_id", perm.ID, "error", err)
                    continue
                }
                if !conditionsMet {
                    continue
                }
            }
            
            // Check scope
            if perm.Scope != "" && req.Scope != "" {
                if !rm.matchesScope(perm.Scope, req.Scope) {
                    continue
                }
            }
            
            decision.Allowed = true
            decision.Reason = fmt.Sprintf("Granted by permission %s:%s", perm.Resource, perm.Action)
            break
        }
    }
    
    if !decision.Allowed {
        decision.Reason = "No matching permissions found"
    }
    
    // Cache decision
    decision.TTL = rm.config.CacheTimeout
    rm.cache.Set(cacheKey, decision, rm.config.CacheTimeout)
    
    return decision, nil
}

func (rm *RBACManager) AssignRole(ctx context.Context, userID, roleID uuid.UUID, assignedBy uuid.UUID, scope string) error {
    // Verify role exists
    role, err := rm.storage.GetRole(ctx, roleID)
    if err != nil {
        return fmt.Errorf("failed to get role: %w", err)
    }
    
    // Check if user already has this role
    userRoles, err := rm.storage.GetUserRoles(ctx, userID)
    if err != nil {
        return fmt.Errorf("failed to get user roles: %w", err)
    }
    
    for _, ur := range userRoles {
        if ur.RoleID == roleID && ur.Scope == scope {
            return ErrRoleAlreadyAssigned
        }
    }
    
    // Create role assignment
    userRole := &UserRole{
        UserID:     userID,
        RoleID:     roleID,
        Scope:      scope,
        AssignedBy: assignedBy,
        AssignedAt: time.Now(),
    }
    
    if err := rm.storage.AssignRole(ctx, userRole); err != nil {
        return fmt.Errorf("failed to assign role: %w", err)
    }
    
    // Clear user's permission cache
    rm.cache.ClearUser(userID)
    
    rm.logger.Info("Role assigned", 
        "user_id", userID, 
        "role_id", roleID, 
        "role_name", role.Name,
        "assigned_by", assignedBy)
    
    return nil
}

func (rm *RBACManager) CreateRole(ctx context.Context, name, description string, permissions []string) (*Role, error) {
    // Validate permissions exist
    for _, permStr := range permissions {
        parts := strings.Split(permStr, ":")
        if len(parts) != 2 {
            return nil, fmt.Errorf("invalid permission format: %s", permStr)
        }
        
        resource, action := parts[0], parts[1]
        exists, err := rm.storage.PermissionExists(ctx, resource, action)
        if err != nil {
            return nil, fmt.Errorf("failed to check permission existence: %w", err)
        }
        if !exists {
            return nil, fmt.Errorf("permission does not exist: %s", permStr)
        }
    }
    
    role := &Role{
        ID:          uuid.New(),
        Name:        name,
        Description: description,
        Permissions: permissions,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    if err := rm.storage.CreateRole(ctx, role); err != nil {
        return nil, fmt.Errorf("failed to create role: %w", err)
    }
    
    rm.logger.Info("Role created", "role_id", role.ID, "name", role.Name)
    
    return role, nil
}

func (rm *RBACManager) matchesResourceAndAction(perm *Permission, resource, action string) bool {
    // Exact match
    if perm.Resource == resource && perm.Action == action {
        return true
    }
    
    // Wildcard matching
    if perm.Resource == "*" || perm.Action == "*" {
        return true
    }
    
    // Pattern matching (e.g., "mission:*", "*:read")
    if strings.HasSuffix(perm.Resource, "*") {
        prefix := strings.TrimSuffix(perm.Resource, "*")
        if strings.HasPrefix(resource, prefix) && perm.Action == action {
            return true
        }
    }
    
    if strings.HasSuffix(perm.Action, "*") {
        prefix := strings.TrimSuffix(perm.Action, "*")
        if strings.HasPrefix(action, prefix) && perm.Resource == resource {
            return true
        }
    }
    
    return false
}
```

### Data Encryption and Key Management

**Encryption Manager**
```go
// pkg/security/encryption/manager.go
package encryption

import (
    "context"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "time"
    
    "golang.org/x/crypto/hkdf"
    "github.com/google/uuid"
)

type EncryptionManager struct {
    keyManager   KeyManager
    hsm          HSMInterface
    config       *Config
    logger       Logger
}

type Config struct {
    Algorithm        string        `yaml:"algorithm"`
    KeySize          int           `yaml:"key_size"`
    KeyRotationDays  int           `yaml:"key_rotation_days"`
    HSMEnabled       bool          `yaml:"hsm_enabled"`
    HSMProvider      string        `yaml:"hsm_provider"`
    HSMConfig        interface{}   `yaml:"hsm_config"`
    
    // Database encryption
    DatabaseKeyID    string        `yaml:"database_key_id"`
    
    // File encryption
    FileKeyID        string        `yaml:"file_key_id"`
    
    // Transport encryption
    TLSMinVersion    string        `yaml:"tls_min_version"`
    CipherSuites     []string      `yaml:"cipher_suites"`
}

type EncryptionKey struct {
    ID          string    `json:"id"`
    Algorithm   string    `json:"algorithm"`
    KeySize     int       `json:"key_size"`
    Purpose     string    `json:"purpose"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    RotatedAt   time.Time `json:"rotated_at,omitempty"`
    ExpiresAt   time.Time `json:"expires_at,omitempty"`
    Version     int       `json:"version"`
}

type EncryptedData struct {
    KeyID       string `json:"key_id"`
    Algorithm   string `json:"algorithm"`
    IV          string `json:"iv"`
    Ciphertext  string `json:"ciphertext"`
    Tag         string `json:"tag,omitempty"`
    Version     int    `json:"version"`
}

func NewEncryptionManager(keyManager KeyManager, config *Config, logger Logger) (*EncryptionManager, error) {
    em := &EncryptionManager{
        keyManager: keyManager,
        config:     config,
        logger:     logger,
    }
    
    // Initialize HSM if enabled
    if config.HSMEnabled {
        hsm, err := NewHSM(config.HSMProvider, config.HSMConfig)
        if err != nil {
            return nil, fmt.Errorf("failed to initialize HSM: %w", err)
        }
        em.hsm = hsm
    }
    
    return em, nil
}

func (em *EncryptionManager) Encrypt(ctx context.Context, data []byte, keyID string) (*EncryptedData, error) {
    key, err := em.keyManager.GetKey(ctx, keyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get encryption key: %w", err)
    }
    
    if key.Status != "active" {
        return nil, fmt.Errorf("key %s is not active", keyID)
    }
    
    var ciphertext, iv, tag []byte
    
    if em.hsm != nil {
        // Use HSM for encryption
        result, err := em.hsm.Encrypt(ctx, keyID, data)
        if err != nil {
            return nil, fmt.Errorf("HSM encryption failed: %w", err)
        }
        ciphertext = result.Ciphertext
        iv = result.IV
        tag = result.Tag
    } else {
        // Software encryption
        keyData, err := em.keyManager.GetKeyData(ctx, keyID)
        if err != nil {
            return nil, fmt.Errorf("failed to get key data: %w", err)
        }
        
        switch key.Algorithm {
        case "AES-256-GCM":
            ciphertext, iv, tag, err = em.encryptAESGCM(data, keyData)
            if err != nil {
                return nil, fmt.Errorf("AES-GCM encryption failed: %w", err)
            }
        default:
            return nil, fmt.Errorf("unsupported algorithm: %s", key.Algorithm)
        }
    }
    
    return &EncryptedData{
        KeyID:      keyID,
        Algorithm:  key.Algorithm,
        IV:         base64.StdEncoding.EncodeToString(iv),
        Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
        Tag:        base64.StdEncoding.EncodeToString(tag),
        Version:    key.Version,
    }, nil
}

func (em *EncryptionManager) Decrypt(ctx context.Context, encData *EncryptedData) ([]byte, error) {
    key, err := em.keyManager.GetKey(ctx, encData.KeyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get decryption key: %w", err)
    }
    
    // Check if key version matches
    if key.Version != encData.Version {
        // Try to get the specific version
        key, err = em.keyManager.GetKeyVersion(ctx, encData.KeyID, encData.Version)
        if err != nil {
            return nil, fmt.Errorf("failed to get key version %d: %w", encData.Version, err)
        }
    }
    
    iv, err := base64.StdEncoding.DecodeString(encData.IV)
    if err != nil {
        return nil, fmt.Errorf("failed to decode IV: %w", err)
    }
    
    ciphertext, err := base64.StdEncoding.DecodeString(encData.Ciphertext)
    if err != nil {
        return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
    }
    
    var tag []byte
    if encData.Tag != "" {
        tag, err = base64.StdEncoding.DecodeString(encData.Tag)
        if err != nil {
            return nil, fmt.Errorf("failed to decode tag: %w", err)
        }
    }
    
    if em.hsm != nil {
        // Use HSM for decryption
        return em.hsm.Decrypt(ctx, encData.KeyID, ciphertext, iv, tag)
    } else {
        // Software decryption
        keyData, err := em.keyManager.GetKeyData(ctx, encData.KeyID)
        if err != nil {
            return nil, fmt.Errorf("failed to get key data: %w", err)
        }
        
        switch encData.Algorithm {
        case "AES-256-GCM":
            return em.decryptAESGCM(ciphertext, keyData, iv, tag)
        default:
            return nil, fmt.Errorf("unsupported algorithm: %s", encData.Algorithm)
        }
    }
}

func (em *EncryptionManager) encryptAESGCM(data, key []byte) (ciphertext, iv, tag []byte, err error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, nil, nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, nil, nil, err
    }
    
    // Generate random IV
    iv = make([]byte, gcm.NonceSize())
    if _, err := rand.Read(iv); err != nil {
        return nil, nil, nil, err
    }
    
    // Encrypt data
    sealed := gcm.Seal(nil, iv, data, nil)
    
    // Split ciphertext and tag
    ciphertext = sealed[:len(sealed)-gcm.Overhead()]
    tag = sealed[len(sealed)-gcm.Overhead():]
    
    return ciphertext, iv, tag, nil
}

func (em *EncryptionManager) GenerateKey(ctx context.Context, purpose, algorithm string, keySize int) (*EncryptionKey, error) {
    keyID := uuid.New().String()
    
    key := &EncryptionKey{
        ID:        keyID,
        Algorithm: algorithm,
        KeySize:   keySize,
        Purpose:   purpose,
        Status:    "active",
        CreatedAt: time.Now(),
        Version:   1,
    }
    
    // Set expiration if configured
    if em.config.KeyRotationDays > 0 {
        key.ExpiresAt = time.Now().AddDate(0, 0, em.config.KeyRotationDays)
    }
    
    if em.hsm != nil {
        // Generate key in HSM
        if err := em.hsm.GenerateKey(ctx, keyID, algorithm, keySize); err != nil {
            return nil, fmt.Errorf("failed to generate key in HSM: %w", err)
        }
    } else {
        // Generate key in software
        keyData := make([]byte, keySize/8)
        if _, err := rand.Read(keyData); err != nil {
            return nil, fmt.Errorf("failed to generate key data: %w", err)
        }
        
        if err := em.keyManager.StoreKey(ctx, key, keyData); err != nil {
            return nil, fmt.Errorf("failed to store key: %w", err)
        }
    }
    
    if err := em.keyManager.CreateKey(ctx, key); err != nil {
        return nil, fmt.Errorf("failed to create key metadata: %w", err)
    }
    
    em.logger.Info("Encryption key generated", "key_id", keyID, "algorithm", algorithm, "purpose", purpose)
    
    return key, nil
}

func (em *EncryptionManager) RotateKey(ctx context.Context, keyID string) (*EncryptionKey, error) {
    // Get current key
    currentKey, err := em.keyManager.GetKey(ctx, keyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get current key: %w", err)
    }
    
    // Create new version
    newKey := &EncryptionKey{
        ID:        keyID,
        Algorithm: currentKey.Algorithm,
        KeySize:   currentKey.KeySize,
        Purpose:   currentKey.Purpose,
        Status:    "active",
        CreatedAt: time.Now(),
        Version:   currentKey.Version + 1,
    }
    
    if em.config.KeyRotationDays > 0 {
        newKey.ExpiresAt = time.Now().AddDate(0, 0, em.config.KeyRotationDays)
    }
    
    if em.hsm != nil {
        // Rotate key in HSM
        if err := em.hsm.RotateKey(ctx, keyID, newKey.Version); err != nil {
            return nil, fmt.Errorf("failed to rotate key in HSM: %w", err)
        }
    } else {
        // Generate new key data
        keyData := make([]byte, newKey.KeySize/8)
        if _, err := rand.Read(keyData); err != nil {
            return nil, fmt.Errorf("failed to generate new key data: %w", err)
        }
        
        if err := em.keyManager.StoreKey(ctx, newKey, keyData); err != nil {
            return nil, fmt.Errorf("failed to store new key: %w", err)
        }
    }
    
    // Update key metadata
    if err := em.keyManager.CreateKey(ctx, newKey); err != nil {
        return nil, fmt.Errorf("failed to create new key metadata: %w", err)
    }
    
    // Mark old version as deprecated
    currentKey.Status = "deprecated"
    currentKey.RotatedAt = time.Now()
    if err := em.keyManager.UpdateKey(ctx, currentKey); err != nil {
        em.logger.Warn("Failed to update old key status", "key_id", keyID, "error", err)
    }
    
    em.logger.Info("Key rotated", "key_id", keyID, "old_version", currentKey.Version, "new_version", newKey.Version)
    
    return newKey, nil
}
```

## Database Schema

```sql
-- Multi-factor authentication
CREATE TABLE user_mfa (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    method VARCHAR(50) NOT NULL,
    secret TEXT,
    phone_number VARCHAR(20),
    email VARCHAR(255),
    backup_codes TEXT[],
    enabled BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_used TIMESTAMP,
    
    UNIQUE(user_id, method)
);

CREATE TABLE mfa_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    method VARCHAR(50) NOT NULL,
    code_hash VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    attempts INTEGER DEFAULT 0,
    verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Certificate authentication
CREATE TABLE user_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    serial VARCHAR(255) NOT NULL,
    subject TEXT NOT NULL,
    issuer TEXT NOT NULL,
    fingerprint VARCHAR(255) UNIQUE NOT NULL,
    valid_from TIMESTAMP NOT NULL,
    valid_to TIMESTAMP NOT NULL,
    key_usage TEXT[],
    extended_key_usage TEXT[],
    certificate_data BYTEA NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- RBAC system
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    permissions TEXT[],
    parent_roles UUID[],
    attributes JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    scope VARCHAR(255),
    conditions TEXT[],
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(resource, action, scope)
);

CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    scope VARCHAR(255),
    conditions JSONB,
    expires_at TIMESTAMP,
    assigned_by UUID REFERENCES users(id),
    assigned_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id, role_id, scope)
);

-- Encryption keys
CREATE TABLE encryption_keys (
    id VARCHAR(255) PRIMARY KEY,
    algorithm VARCHAR(100) NOT NULL,
    key_size INTEGER NOT NULL,
    purpose VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    rotated_at TIMESTAMP,
    expires_at TIMESTAMP
);

-- Security events
CREATE TABLE security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    source_ip INET,
    user_agent TEXT,
    description TEXT NOT NULL,
    metadata JSONB,
    timestamp TIMESTAMP DEFAULT NOW(),
    processed BOOLEAN DEFAULT false
);

-- Performance indexes
CREATE INDEX idx_user_mfa_user ON user_mfa(user_id);
CREATE INDEX idx_mfa_challenges_expires ON mfa_challenges(expires_at);
CREATE INDEX idx_user_certificates_user ON user_certificates(user_id);
CREATE INDEX idx_user_certificates_fingerprint ON user_certificates(fingerprint);
CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
CREATE INDEX idx_security_events_type_time ON security_events(event_type, timestamp DESC);
CREATE INDEX idx_security_events_user_time ON security_events(user_id, timestamp DESC);
```

## API Specifications

### Authentication API
```
POST   /api/v1/auth/mfa/setup             # Setup MFA
POST   /api/v1/auth/mfa/verify            # Verify MFA code
POST   /api/v1/auth/mfa/challenge         # Request MFA challenge
POST   /api/v1/auth/certificate           # Certificate authentication
GET    /api/v1/auth/certificate/register  # Register user certificate
```

### Authorization API
```
POST   /api/v1/auth/check-access          # Check user permissions
GET    /api/v1/roles                      # List roles
POST   /api/v1/roles                      # Create role
GET    /api/v1/roles/{id}                 # Get role details
PUT    /api/v1/roles/{id}                 # Update role
DELETE /api/v1/roles/{id}                 # Delete role
POST   /api/v1/users/{id}/roles           # Assign role to user
```

### Security Management API
```
GET    /api/v1/security/events            # List security events
POST   /api/v1/security/keys/generate     # Generate encryption key
POST   /api/v1/security/keys/rotate       # Rotate encryption key
GET    /api/v1/security/audit             # Security audit report
POST   /api/v1/security/compliance        # Run compliance check
```

## Testing Strategy

### Unit Tests
```go
func TestMFAManager_SetupTOTP(t *testing.T) {
    manager := setupTestMFAManager()
    userID := uuid.New()
    
    setup, err := manager.SetupTOTP(context.Background(), userID, "testuser")
    assert.NoError(t, err)
    assert.NotEmpty(t, setup.Secret)
    assert.NotEmpty(t, setup.QRCode)
    assert.Len(t, setup.BackupCodes, 10)
}

func TestRBACManager_CheckAccess(t *testing.T) {
    manager := setupTestRBACManager()
    userID := uuid.New()
    
    // Assign role with specific permission
    roleID := createTestRole(manager, "test_role", []string{"mission:read"})
    err := manager.AssignRole(context.Background(), userID, roleID, uuid.New(), "")
    assert.NoError(t, err)
    
    // Check access
    decision, err := manager.CheckAccess(context.Background(), &AccessRequest{
        UserID:   userID,
        Resource: "mission",
        Action:   "read",
    })
    
    assert.NoError(t, err)
    assert.True(t, decision.Allowed)
}
```

### Integration Tests
```go
func TestCertificateAuthentication(t *testing.T) {
    manager := setupTestCertAuthManager()
    
    // Generate test certificate
    cert := generateTestCertificate()
    certPEM := encodeCertificateToPEM(cert)
    
    // Register certificate
    userID := uuid.New()
    err := manager.RegisterCertificate(context.Background(), userID, certPEM)
    assert.NoError(t, err)
    
    // Authenticate with certificate
    result, err := manager.AuthenticateWithCertificate(context.Background(), certPEM)
    assert.NoError(t, err)
    assert.Equal(t, userID, result.UserID)
}
```

## Acceptance Criteria

### Multi-Factor Authentication
- [ ] TOTP authentication working with standard apps
- [ ] SMS and email second factor delivery
- [ ] Hardware token support (FIDO2/WebAuthn)
- [ ] Backup codes for recovery
- [ ] Administrative MFA policy enforcement

### Certificate Authentication
- [ ] X.509 certificate validation
- [ ] CAC/PIV card integration
- [ ] Certificate revocation checking
- [ ] Mutual TLS client authentication
- [ ] Certificate-to-user mapping

### Role-Based Access Control
- [ ] Hierarchical role system
- [ ] Fine-grained permissions
- [ ] Dynamic permission evaluation
- [ ] Attribute-based access control
- [ ] Policy management interface

### Data Encryption
- [ ] AES-256 encryption for data at rest
- [ ] TLS 1.3 for data in transit
- [ ] Hardware Security Module integration
- [ ] Automated key rotation
- [ ] Secure key distribution

### Security Monitoring
- [ ] Real-time security event detection
- [ ] Anomaly detection algorithms
- [ ] Security incident alerting
- [ ] SIEM integration
- [ ] Compliance reporting

## Dependencies

### Backend Dependencies
```go
require (
    github.com/pquerna/otp v1.4.0                    // TOTP implementation
    github.com/skip2/go-qrcode v0.0.0-20200617195104  // QR code generation
    golang.org/x/crypto v0.14.0                      // Cryptographic functions
    github.com/go-webauthn/webauthn v0.8.6           // WebAuthn support
    github.com/miekg/pkcs11 v1.1.1                   // HSM integration
)
```

### Infrastructure Dependencies
- Certificate Authority for certificate validation
- Hardware Security Module (optional)
- SIEM system for security monitoring
- SMS provider for second factor authentication

## Definition of Done

### Code Quality
- [ ] All code reviewed and approved
- [ ] Unit tests with 90%+ coverage
- [ ] Security testing completed
- [ ] Performance benchmarks meet requirements
- [ ] Vulnerability assessment passed

### Functionality
- [ ] All user stories completed and accepted
- [ ] Multi-factor authentication working
- [ ] Certificate authentication integrated
- [ ] RBAC system operational
- [ ] Encryption and key management functional

### Security & Compliance
- [ ] Security framework meets FISMA requirements
- [ ] Compliance audit completed
- [ ] Penetration testing passed
- [ ] Security monitoring operational
- [ ] Documentation complete for auditors

---

**Sprint Review Date:** [TBD]  
**Sprint Retrospective Date:** [TBD]  
**Next Sprint Planning:** [TBD]
