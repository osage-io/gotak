package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
	"github.com/dfedick/gotak/pkg/logger"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrAccountLocked     = errors.New("account locked due to too many failed attempts")
	ErrUserNotActive     = errors.New("user account is not active")
)

// User represents a user in the system
type User struct {
	ID              string     `json:"id" db:"id"`
	Username        string     `json:"username" db:"username"`
	Email           string     `json:"email" db:"email"`
	PasswordHash    string     `json:"-" db:"password_hash"`
	FirstName       *string    `json:"first_name" db:"first_name"`
	LastName        *string    `json:"last_name" db:"last_name"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	MFAEnabled      bool       `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret       *string    `json:"-" db:"mfa_secret"`
	LastLogin       *time.Time `json:"last_login" db:"last_login"`
	FailedAttempts  int        `json:"-" db:"failed_attempts"`
	LockedUntil     *time.Time `json:"-" db:"locked_until"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// Role represents a system role
type Role struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	IsSystemRole bool      `json:"is_system_role" db:"is_system_role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Permission represents a system permission
type Permission struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username   string `json:"username" validate:"required"`
	Password   string `json:"password" validate:"required"`
	MFACode    string `json:"mfa_code,omitempty"`
	AuthMethod string `json:"auth_method" validate:"required,oneof=local vault_oidc certificate"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username  string  `json:"username" validate:"required,min=3,max=50"`
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=8"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

// AccountLockInfo represents account lockout information
type AccountLockInfo struct {
	UserID             string     `json:"user_id"`
	Username           string     `json:"username"`
	FailedAttempts     int        `json:"failed_attempts"`
	MaxAttempts        int        `json:"max_attempts"`
	RemainingAttempts  int        `json:"remaining_attempts"`
	IsLocked           bool       `json:"is_locked"`
	LockedUntil        *time.Time `json:"locked_until,omitempty"`
}

// PasswordChangeRequest represents a password change request
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required"`
}

// AuthService handles authentication operations
type AuthService struct {
	db                *sql.DB
	jwtManager        *JWTManager
	tokenStorage      *PostgreSQLTokenStorage
	passwordValidator *PasswordValidator
	logger            *logger.Logger
	config            AuthConfig
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// Password policy
	MinPasswordLength    int           `mapstructure:"min_password_length"`
	MaxPasswordLength    int           `mapstructure:"max_password_length"`
	RequireUppercase     bool          `mapstructure:"require_uppercase"`
	RequireLowercase     bool          `mapstructure:"require_lowercase"`
	RequireNumbers       bool          `mapstructure:"require_numbers"`
	RequireSpecialChars  bool          `mapstructure:"require_special_chars"`
	
	// Account lockout
	MaxFailedAttempts    int           `mapstructure:"max_failed_attempts"`
	LockoutDuration      time.Duration `mapstructure:"lockout_duration"`
	
	// Password hashing
	BcryptCost           int           `mapstructure:"bcrypt_cost"`
	
	// JWT configuration
	JWT                  JWTConfig     `mapstructure:"jwt"`
}

// NewAuthService creates a new authentication service
func NewAuthService(db *sql.DB, config AuthConfig, logger *logger.Logger) (*AuthService, error) {
	// Set defaults
	if config.MinPasswordLength == 0 {
		config.MinPasswordLength = 8
	}
	if config.MaxPasswordLength == 0 {
		config.MaxPasswordLength = 128
	}
	if config.MaxFailedAttempts == 0 {
		config.MaxFailedAttempts = 5
	}
	if config.LockoutDuration == 0 {
		config.LockoutDuration = 15 * time.Minute
	}
	if config.BcryptCost == 0 {
		config.BcryptCost = bcrypt.DefaultCost
	}

	// Create token storage
	tokenStorage := NewPostgreSQLTokenStorage(db, logger)
	
	// Create JWT manager
	jwtManager := NewJWTManager(config.JWT, tokenStorage)
	
	// Create password policy from config
	passwordPolicy := PasswordPolicy{
		MinLength:                config.MinPasswordLength,
		MaxLength:                config.MaxPasswordLength,
		RequireUppercase:         config.RequireUppercase,
		RequireLowercase:         config.RequireLowercase,
		RequireNumbers:           config.RequireNumbers,
		RequireSpecialChars:      config.RequireSpecialChars,
		MinUniqueChars:           8, // Default to 8
		ForbidCommonPasswords:    true,
		ForbidUsernameInPassword: true,
		MaxRepeatingChars:        3,
		ForbidSequentialChars:    true,
		PasswordHistorySize:      5,
		PasswordMaxAge:           90 * 24 * time.Hour,
		WarnBeforeExpiration:     7 * 24 * time.Hour,
		MaxFailedAttempts:        config.MaxFailedAttempts,
		LockoutDuration:          config.LockoutDuration,
		LockoutProgressiveDelay:  true,
		AllowedSpecialChars:      `!@#$%^&*()_+-=[]{}|;:,.<>?`,
		ForbiddenChars:           `"'` + "`",
	}
	
	// Create password validator
	passwordValidator := NewPasswordValidator(passwordPolicy, logger)
	
	return &AuthService{
		db:                db,
		jwtManager:        jwtManager,
		tokenStorage:      tokenStorage,
		passwordValidator: passwordValidator,
		logger:            logger,
		config:            config,
	}, nil
}

// Authenticate authenticates a user and returns a token pair
func (a *AuthService) Authenticate(req *LoginRequest) (*TokenPair, error) {
	// Get user by username
	user, err := a.GetUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			a.logger.Warn().
				Str("username", req.Username).
				Str("auth_method", req.AuthMethod).
				Msg("Authentication attempt for non-existent user")
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		a.logger.Warn().
			Str("user_id", user.ID).
			Str("username", user.Username).
			Msg("Authentication attempt for inactive user")
		return nil, ErrUserNotActive
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		a.logger.Warn().
			Str("user_id", user.ID).
			Str("username", user.Username).
			Time("locked_until", *user.LockedUntil).
			Msg("Authentication attempt for locked account")
		return nil, ErrAccountLocked
	}

	// Verify password
	if !a.verifyPassword(req.Password, user.PasswordHash) {
		// Increment failed attempts
		if err := a.incrementFailedAttempts(user.ID); err != nil {
			a.logger.Error().
				Err(err).
				Str("user_id", user.ID).
				Msg("Failed to increment failed attempts")
		}
		
		a.logger.Warn().
			Str("user_id", user.ID).
			Str("username", user.Username).
			Msg("Invalid password")
		return nil, ErrInvalidCredentials
	}

	// TODO: Add MFA verification if enabled
	if user.MFAEnabled && req.MFACode == "" {
		return nil, errors.New("MFA code required")
	}

	// Reset failed attempts and update last login
	if err := a.resetFailedAttempts(user.ID); err != nil {
		a.logger.Warn().
			Err(err).
			Str("user_id", user.ID).
			Msg("Failed to reset failed attempts")
	}

	// Get user roles and permissions
	roles, err := a.GetUserRoles(user.ID)
	if err != nil {
		a.logger.Error().
			Err(err).
			Str("user_id", user.ID).
			Msg("Failed to get user roles")
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	permissions, err := a.GetUserPermissions(user.ID)
	if err != nil {
		a.logger.Error().
			Err(err).
			Str("user_id", user.ID).
			Msg("Failed to get user permissions")
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// Generate token pair
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	permissionNames := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionNames[i] = perm.Name
	}

	tokenPair, err := a.jwtManager.GenerateTokenPair(user.ID, user.Username, roleNames, permissionNames)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	a.logger.Info().
		Str("user_id", user.ID).
		Str("username", user.Username).
		Str("auth_method", req.AuthMethod).
		Msg("User authenticated successfully")

	return tokenPair, nil
}

// RefreshTokens refreshes an access token using a refresh token
func (a *AuthService) RefreshTokens(refreshToken string) (*TokenPair, error) {
	// Validate the refresh token and get user info
	claims, err := a.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}

	// Get updated user roles and permissions
	roles, err := a.GetUserRoles(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	permissions, err := a.GetUserPermissions(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	permissionNames := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionNames[i] = perm.Name
	}

	return a.jwtManager.RefreshTokens(refreshToken, roleNames, permissionNames)
}

// ValidateToken validates a JWT token and returns claims
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	return a.jwtManager.ValidateToken(tokenString)
}

// RevokeToken revokes a specific token
func (a *AuthService) RevokeToken(tokenString string) error {
	return a.jwtManager.RevokeToken(tokenString)
}

// RevokeAllUserTokens revokes all tokens for a user
func (a *AuthService) RevokeAllUserTokens(userID string) error {
	return a.jwtManager.RevokeAllUserTokens(userID)
}

// RegisterUser registers a new user
func (a *AuthService) RegisterUser(req *RegisterRequest) (*User, error) {
	// Check if user already exists
	if _, err := a.GetUserByUsername(req.Username); err == nil {
		return nil, ErrUserAlreadyExists
	}

	// Check email uniqueness
	if _, err := a.GetUserByEmail(req.Email); err == nil {
		return nil, errors.New("email already exists")
	}

	// Validate password policy
	if err := a.validatePassword(req.Password, req.Username); err != nil {
		a.logger.Warn().
			Str("username", req.Username).
			Err(err).
			Msg("Password validation failed during registration")
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash password
	passwordHash, err := a.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	userID := uuid.New().String()
	query := `
		INSERT INTO gotak.users (id, username, email, password_hash, first_name, last_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`

	_, err = a.db.Exec(query, userID, req.Username, req.Email, passwordHash, req.FirstName, req.LastName, true)
	if err != nil {
		a.logger.Error().
			Err(err).
			Str("username", req.Username).
			Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	a.logger.Info().
		Str("user_id", userID).
		Str("username", req.Username).
		Msg("User registered successfully")

	return a.GetUserByID(userID)
}

// GetUserByID retrieves a user by ID
func (a *AuthService) GetUserByID(userID string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, 
		       is_active, mfa_enabled, mfa_secret, last_login, 
		       failed_attempts, locked_until, created_at, updated_at
		FROM gotak.users
		WHERE id = $1
	`

	user := &User{}
	err := a.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.IsActive, &user.MFAEnabled,
		&user.MFASecret, &user.LastLogin, &user.FailedAttempts,
		&user.LockedUntil, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (a *AuthService) GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, 
		       is_active, mfa_enabled, mfa_secret, last_login, 
		       failed_attempts, locked_until, created_at, updated_at
		FROM gotak.users
		WHERE username = $1
	`

	user := &User{}
	err := a.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.IsActive, &user.MFAEnabled,
		&user.MFASecret, &user.LastLogin, &user.FailedAttempts,
		&user.LockedUntil, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (a *AuthService) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, 
		       is_active, mfa_enabled, mfa_secret, last_login, 
		       failed_attempts, locked_until, created_at, updated_at
		FROM gotak.users
		WHERE email = $1
	`

	user := &User{}
	err := a.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.IsActive, &user.MFAEnabled,
		&user.MFASecret, &user.LastLogin, &user.FailedAttempts,
		&user.LockedUntil, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetUserRoles retrieves roles for a user
func (a *AuthService) GetUserRoles(userID string) ([]Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.is_system_role, r.created_at
		FROM gotak.roles r
		INNER JOIN gotak.user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND ur.is_active = true
	`

	rows, err := a.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystemRole, &role.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetUserPermissions retrieves permissions for a user through their roles
func (a *AuthService) GetUserPermissions(userID string) ([]Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM gotak.permissions p
		INNER JOIN gotak.role_permissions rp ON p.id = rp.permission_id
		INNER JOIN gotak.user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND ur.is_active = true
	`

	rows, err := a.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var perm Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// verifyPassword verifies a password against its hash
func (a *AuthService) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// hashPassword hashes a password using bcrypt
func (a *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), a.config.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// validatePassword validates a password against comprehensive security policy
func (a *AuthService) validatePassword(password, username string) error {
	return a.passwordValidator.ValidatePassword(password, username)
}

// GetPasswordComplexityReport generates a detailed password complexity report
func (a *AuthService) GetPasswordComplexityReport(password, username string) *PasswordComplexityReport {
	return a.passwordValidator.GeneratePasswordComplexityReport(password, username)
}

// CheckPasswordExpiration checks if a user's password needs to be changed
func (a *AuthService) CheckPasswordExpiration(userID string) (bool, time.Time, error) {
	query := `
		SELECT COALESCE(password_changed_at, created_at)
		FROM gotak.users
		WHERE id = $1
	`
	
	var passwordChangedAt time.Time
	err := a.db.QueryRow(query, userID).Scan(&passwordChangedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, time.Time{}, ErrUserNotFound
		}
		return false, time.Time{}, fmt.Errorf("failed to get password change date: %w", err)
	}
	
	maxAge := a.passwordValidator.policy.PasswordMaxAge
	expirationDate := passwordChangedAt.Add(maxAge)
	isExpired := time.Now().After(expirationDate)
	
	return isExpired, expirationDate, nil
}

// ShouldWarnPasswordExpiration checks if user should be warned about password expiration
func (a *AuthService) ShouldWarnPasswordExpiration(userID string) (bool, time.Time, error) {
	isExpired, expirationDate, err := a.CheckPasswordExpiration(userID)
	if err != nil {
		return false, time.Time{}, err
	}
	
	if isExpired {
		return true, expirationDate, nil
	}
	
	warnPeriod := a.passwordValidator.policy.WarnBeforeExpiration
	warnDate := expirationDate.Add(-warnPeriod)
	shouldWarn := time.Now().After(warnDate)
	
	return shouldWarn, expirationDate, nil
}

// incrementFailedAttempts increments failed login attempts and locks account if needed
func (a *AuthService) incrementFailedAttempts(userID string) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current failed attempts
	var failedAttempts int
	err = tx.QueryRow("SELECT failed_attempts FROM gotak.users WHERE id = $1", userID).Scan(&failedAttempts)
	if err != nil {
		return err
	}

	failedAttempts++
	
	// Check if account should be locked
	var lockedUntil *time.Time
	if failedAttempts >= a.config.MaxFailedAttempts {
		lockTime := time.Now().Add(a.config.LockoutDuration)
		lockedUntil = &lockTime
	}

	// Update user
	_, err = tx.Exec(`
		UPDATE gotak.users 
		SET failed_attempts = $1, locked_until = $2, updated_at = NOW()
		WHERE id = $3
	`, failedAttempts, lockedUntil, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// resetFailedAttempts resets failed login attempts after successful login
func (a *AuthService) resetFailedAttempts(userID string) error {
	_, err := a.db.Exec(`
		UPDATE gotak.users 
		SET failed_attempts = 0, locked_until = NULL, last_login = NOW(), updated_at = NOW()
		WHERE id = $1
	`, userID)
	return err
}

// IsAccountLocked checks if a user account is currently locked
func (a *AuthService) IsAccountLocked(userID string) (bool, error) {
	query := `
		SELECT locked_until, failed_attempts
		FROM gotak.users
		WHERE id = $1
	`
	
	var lockedUntil *time.Time
	var failedAttempts int
	err := a.db.QueryRow(query, userID).Scan(&lockedUntil, &failedAttempts)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrUserNotFound
		}
		return false, fmt.Errorf("failed to check account lock status: %w", err)
	}
	
	// Check if account is locked and lock period hasn't expired
	if lockedUntil != nil && time.Now().Before(*lockedUntil) {
		return true, nil
	}
	
	// If lock period has expired, automatically unlock the account
	if lockedUntil != nil && time.Now().After(*lockedUntil) {
		if err := a.unlockAccount(userID); err != nil {
			a.logger.Error().
				Str("user_id", userID).
				Err(err).
				Msg("Failed to automatically unlock expired account")
		}
	}
	
	return false, nil
}

// unlockAccount manually unlocks a user account
func (a *AuthService) unlockAccount(userID string) error {
	_, err := a.db.Exec(`
		UPDATE gotak.users 
		SET failed_attempts = 0, locked_until = NULL, updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to unlock account: %w", err)
	}
	
	a.logger.Info().
		Str("user_id", userID).
		Msg("Account unlocked successfully")
	
	return nil
}

// GetAccountLockInfo returns detailed information about account lock status
func (a *AuthService) GetAccountLockInfo(userID string) (*AccountLockInfo, error) {
	query := `
		SELECT username, failed_attempts, locked_until
		FROM gotak.users
		WHERE id = $1
	`
	
	var username string
	var failedAttempts int
	var lockedUntil *time.Time
	
	err := a.db.QueryRow(query, userID).Scan(&username, &failedAttempts, &lockedUntil)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get account lock info: %w", err)
	}
	
	lockInfo := &AccountLockInfo{
		UserID:           userID,
		Username:         username,
		FailedAttempts:   failedAttempts,
		MaxAttempts:      a.config.MaxFailedAttempts,
		IsLocked:         false,
		RemainingAttempts: a.config.MaxFailedAttempts - failedAttempts,
	}
	
	if lockedUntil != nil {
		lockInfo.LockedUntil = lockedUntil
		lockInfo.IsLocked = time.Now().Before(*lockedUntil)
	}
	
	return lockInfo, nil
}

// ChangePassword changes a user's password with validation
func (a *AuthService) ChangePassword(userID, currentPassword, newPassword string) error {
	// Get user to verify current password
	user, err := a.GetUserByID(userID)
	if err != nil {
		return err
	}
	
	// Verify current password
	if !a.verifyPassword(currentPassword, user.PasswordHash) {
		a.logger.Warn().
			Str("user_id", userID).
			Msg("Incorrect current password during password change attempt")
		return ErrInvalidCredentials
	}
	
	// Validate new password
	if err := a.validatePassword(newPassword, user.Username); err != nil {
		a.logger.Warn().
			Str("user_id", userID).
			Err(err).
			Msg("New password validation failed")
		return fmt.Errorf("password validation failed: %w", err)
	}
	
	// Check if new password is same as current password
	if a.verifyPassword(newPassword, user.PasswordHash) {
		return errors.New("new password must be different from current password")
	}
	
	// Hash new password
	newPasswordHash, err := a.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}
	
	// Update password
	_, err = a.db.Exec(`
		UPDATE gotak.users 
		SET password_hash = $1, password_changed_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	
	a.logger.Info().
		Str("user_id", userID).
		Str("username", user.Username).
		Msg("Password changed successfully")
	
	// Revoke all existing tokens to force re-authentication
	if err := a.RevokeAllUserTokens(userID); err != nil {
		a.logger.Warn().
			Str("user_id", userID).
			Err(err).
			Msg("Failed to revoke tokens after password change")
		// Don't fail the password change for this
	}
	
	return nil
}
