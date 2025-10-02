package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"` // Never expose password hash in JSON
	Role         string     `json:"role" db:"role"`
	Callsign     string     `json:"callsign,omitempty" db:"callsign"`
	Active       bool       `json:"active" db:"active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastLogin    *time.Time `json:"last_login,omitempty" db:"last_login"`
}

// Session represents an active user session
type Session struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	TokenHash string    `json:"-" db:"token_hash"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IPAddress string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent string    `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UserPreferences represents user-specific preferences
type UserPreferences struct {
	UserID               uuid.UUID       `json:"user_id" db:"user_id"`
	Theme                string          `json:"theme" db:"theme"`
	Language             string          `json:"language" db:"language"`
	NotificationSettings map[string]interface{} `json:"notification_settings" db:"notification_settings"`
	MapPreferences       map[string]interface{} `json:"map_preferences" db:"map_preferences"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	TokenHash string    `json:"-" db:"token_hash"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// LoginAttempt tracks login attempts for rate limiting
type LoginAttempt struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Username    *string   `json:"username,omitempty" db:"username"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	Success     bool      `json:"success" db:"success"`
	AttemptedAt time.Time `json:"attempted_at" db:"attempted_at"`
}

// UserRole represents possible user roles
type UserRole string

const (
	// RoleAdmin has full system access
	RoleAdmin UserRole = "admin"
	// RoleOperator can manage operations
	RoleOperator UserRole = "operator"
	// RoleViewer has read-only access
	RoleViewer UserRole = "viewer"
)

// IsValid checks if a role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case RoleAdmin, RoleOperator, RoleViewer:
		return true
	default:
		return false
	}
}

// HasPermission checks if a role has a specific permission
func (r UserRole) HasPermission(permission string) bool {
	switch r {
	case RoleAdmin:
		return true // Admin has all permissions
	case RoleOperator:
		// Operator permissions
		switch permission {
		case "read", "write", "update", "execute":
			return true
		default:
			return false
		}
	case RoleViewer:
		// Viewer permissions
		return permission == "read"
	default:
		return false
	}
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required"`
	Callsign string `json:"callsign,omitempty" validate:"max=50"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Role     *string `json:"role,omitempty"`
	Callsign *string `json:"callsign,omitempty" validate:"omitempty,max=50"`
	Active   *bool   `json:"active,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UserFilter represents filters for querying users
type UserFilter struct {
	Active   *bool
	Role     *string
	SearchTerm *string // Searches username, email, callsign
	Limit    int
	Offset   int
	OrderBy  string
	OrderDir string // ASC or DESC
}

// Sanitize returns a sanitized version of the user for API responses
func (u *User) Sanitize() *User {
	return &User{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		Callsign:  u.Callsign,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		LastLogin: u.LastLogin,
		// PasswordHash is intentionally omitted
	}
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return UserRole(u.Role) == RoleAdmin
}

// IsOperator checks if the user has operator role
func (u *User) IsOperator() bool {
	return UserRole(u.Role) == RoleOperator
}

// IsViewer checks if the user has viewer role
func (u *User) IsViewer() bool {
	return UserRole(u.Role) == RoleViewer
}

// CanWrite checks if the user can perform write operations
func (u *User) CanWrite() bool {
	return u.IsAdmin() || u.IsOperator()
}

// CanRead checks if the user can perform read operations
func (u *User) CanRead() bool {
	return u.Active // All active users can read
}