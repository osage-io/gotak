// Package mfa provides multi-factor authentication interfaces and implementations
// for the GoTAK tactical awareness platform.
package mfa

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MFAType represents the type of multi-factor authentication method
type MFAType string

const (
	MFATypeTOTP     MFAType = "totp"     // Time-based One-Time Password
	MFATypeSMS      MFAType = "sms"      // SMS-based verification
	MFATypeEmail    MFAType = "email"    // Email-based verification
	MFATypeWebAuthn MFAType = "webauthn" // WebAuthn/FIDO2 hardware tokens
	MFATypeBackup   MFAType = "backup"   // Backup recovery codes
)

// MFAStatus represents the current status of an MFA method
type MFAStatus string

const (
	MFAStatusPending   MFAStatus = "pending"   // Awaiting enrollment completion
	MFAStatusActive    MFAStatus = "active"    // Active and ready for use
	MFAStatusDisabled  MFAStatus = "disabled"  // Temporarily disabled
	MFAStatusRevoked   MFAStatus = "revoked"   // Permanently revoked
	MFAStatusSuspended MFAStatus = "suspended" // Suspended due to security concerns
)

// ChallengeStatus represents the status of an MFA challenge
type ChallengeStatus string

const (
	ChallengeStatusPending   ChallengeStatus = "pending"   // Awaiting user response
	ChallengeStatusVerified  ChallengeStatus = "verified"  // Successfully verified
	ChallengeStatusFailed    ChallengeStatus = "failed"    // Verification failed
	ChallengeStatusExpired   ChallengeStatus = "expired"   // Challenge expired
	ChallengeStatusCanceled  ChallengeStatus = "canceled"  // Challenge canceled
)

// MFAProvider defines the interface that all MFA providers must implement
type MFAProvider interface {
	// GetType returns the MFA type this provider handles
	GetType() MFAType

	// GenerateSecret creates a new MFA secret for enrollment
	GenerateSecret(ctx context.Context, userID uuid.UUID, metadata map[string]string) (*MFASecret, error)

	// VerifyEnrollment validates that enrollment is complete and working
	VerifyEnrollment(ctx context.Context, secret *MFASecret, verificationCode string) error

	// CreateChallenge generates a new authentication challenge
	CreateChallenge(ctx context.Context, factor *MFAFactor) (*MFAChallenge, error)

	// VerifyChallenge validates a user's response to an authentication challenge
	VerifyChallenge(ctx context.Context, challenge *MFAChallenge, response string) error

	// ValidateConfiguration checks if the provider configuration is valid
	ValidateConfiguration() error
}

// MFAManager defines the main interface for managing multi-factor authentication
type MFAManager interface {
	// RegisterProvider registers a new MFA provider
	RegisterProvider(provider MFAProvider) error

	// GetProvider returns a specific MFA provider by type
	GetProvider(mfaType MFAType) (MFAProvider, error)

	// ListProviders returns all registered providers
	ListProviders() []MFAProvider

	// EnrollFactor initiates enrollment for a new MFA factor
	EnrollFactor(ctx context.Context, userID uuid.UUID, mfaType MFAType, metadata map[string]string) (*EnrollmentSession, error)

	// CompleteFactor completes enrollment and activates the MFA factor
	CompleteFactor(ctx context.Context, sessionID uuid.UUID, verificationCode string) (*MFAFactor, error)

	// ListUserFactors returns all MFA factors for a user
	ListUserFactors(ctx context.Context, userID uuid.UUID) ([]*MFAFactor, error)

	// DisableFactor temporarily disables an MFA factor
	DisableFactor(ctx context.Context, factorID uuid.UUID) error

	// RevokeFactor permanently revokes an MFA factor
	RevokeFactor(ctx context.Context, factorID uuid.UUID) error

	// CreateAuthChallenge creates a new authentication challenge
	CreateAuthChallenge(ctx context.Context, userID uuid.UUID, requestedTypes []MFAType) (*AuthChallenge, error)

	// VerifyAuthChallenge verifies a user's response to an authentication challenge
	VerifyAuthChallenge(ctx context.Context, challengeID uuid.UUID, factorID uuid.UUID, response string) error

	// IsAuthChallengeComplete checks if all required factors are satisfied
	IsAuthChallengeComplete(ctx context.Context, challengeID uuid.UUID) (bool, error)
}

// MFASecret represents the secret data for an MFA enrollment
type MFASecret struct {
	ID       uuid.UUID         `json:"id"`
	UserID   uuid.UUID         `json:"user_id"`
	Type     MFAType           `json:"type"`
	Secret   string            `json:"secret"`           // Encrypted secret data
	Metadata map[string]string `json:"metadata"`         // Provider-specific metadata
	QRCode   []byte            `json:"qr_code,omitempty"` // QR code image for enrollment
	URL      string            `json:"url,omitempty"`     // Provider-specific URL (e.g., otpauth://)
}

// MFAFactor represents a completed and active MFA factor for a user
type MFAFactor struct {
	ID           uuid.UUID         `json:"id"`
	UserID       uuid.UUID         `json:"user_id"`
	Type         MFAType           `json:"type"`
	Name         string            `json:"name"`         // User-friendly name
	Status       MFAStatus         `json:"status"`
	Secret       string            `json:"secret"`       // Encrypted secret data
	Metadata     map[string]string `json:"metadata"`     // Provider-specific metadata
	BackupCodes  []string          `json:"backup_codes"` // Encrypted backup codes
	LastUsedAt   *time.Time        `json:"last_used_at,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// EnrollmentSession represents an ongoing MFA enrollment process
type EnrollmentSession struct {
	ID         uuid.UUID         `json:"id"`
	UserID     uuid.UUID         `json:"user_id"`
	Type       MFAType           `json:"type"`
	Secret     *MFASecret        `json:"secret"`
	Metadata   map[string]string `json:"metadata"`
	ExpiresAt  time.Time         `json:"expires_at"`
	CreatedAt  time.Time         `json:"created_at"`
}

// MFAChallenge represents a single MFA challenge for one factor
type MFAChallenge struct {
	ID         uuid.UUID       `json:"id"`
	FactorID   uuid.UUID       `json:"factor_id"`
	Type       MFAType         `json:"type"`
	Status     ChallengeStatus `json:"status"`
	Challenge  string          `json:"challenge,omitempty"` // Challenge data (e.g., SMS code)
	Attempts   int             `json:"attempts"`
	MaxAttempts int            `json:"max_attempts"`
	ExpiresAt  time.Time       `json:"expires_at"`
	CreatedAt  time.Time       `json:"created_at"`
	VerifiedAt *time.Time      `json:"verified_at,omitempty"`
}

// AuthChallenge represents a complete authentication challenge that may require multiple factors
type AuthChallenge struct {
	ID                uuid.UUID                 `json:"id"`
	UserID            uuid.UUID                 `json:"user_id"`
	RequiredTypes     []MFAType                 `json:"required_types"`
	Challenges        map[uuid.UUID]*MFAChallenge `json:"challenges"` // Factor ID -> Challenge
	Status            ChallengeStatus           `json:"status"`
	CompletedFactors  []uuid.UUID               `json:"completed_factors"`
	ExpiresAt         time.Time                 `json:"expires_at"`
	CreatedAt         time.Time                 `json:"created_at"`
	CompletedAt       *time.Time                `json:"completed_at,omitempty"`
}

// MFAConfig defines the configuration for MFA policies and providers
type MFAConfig struct {
	// Global MFA settings
	Enabled              bool          `yaml:"enabled" json:"enabled"`
	RequiredTypes        []MFAType     `yaml:"required_types" json:"required_types"`
	ChallengeLifetime    time.Duration `yaml:"challenge_lifetime" json:"challenge_lifetime"`
	EnrollmentLifetime   time.Duration `yaml:"enrollment_lifetime" json:"enrollment_lifetime"`
	MaxFailedAttempts    int           `yaml:"max_failed_attempts" json:"max_failed_attempts"`
	LockoutDuration      time.Duration `yaml:"lockout_duration" json:"lockout_duration"`

	// Backup codes
	BackupCodesEnabled   bool `yaml:"backup_codes_enabled" json:"backup_codes_enabled"`
	BackupCodesCount     int  `yaml:"backup_codes_count" json:"backup_codes_count"`
	BackupCodesLength    int  `yaml:"backup_codes_length" json:"backup_codes_length"`

	// Provider configurations
	TOTP     TOTPConfig     `yaml:"totp" json:"totp"`
	SMS      SMSConfig      `yaml:"sms" json:"sms"`
	Email    EmailConfig    `yaml:"email" json:"email"`
	WebAuthn WebAuthnConfig `yaml:"webauthn" json:"webauthn"`
}

// TOTPConfig defines configuration for Time-based One-Time Password
type TOTPConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	Issuer     string `yaml:"issuer" json:"issuer"`
	Algorithm  string `yaml:"algorithm" json:"algorithm"`     // SHA1, SHA256, SHA512
	Digits     int    `yaml:"digits" json:"digits"`           // 6 or 8
	Period     int    `yaml:"period" json:"period"`           // Time step in seconds
	Skew       int    `yaml:"skew" json:"skew"`               // Number of time steps to allow
}

// SMSConfig defines configuration for SMS-based MFA
type SMSConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	Provider   string `yaml:"provider" json:"provider"`     // twilio, aws-sns, etc.
	Template   string `yaml:"template" json:"template"`     // SMS message template
	CodeLength int    `yaml:"code_length" json:"code_length"` // Length of SMS code
	CodeTTL    time.Duration `yaml:"code_ttl" json:"code_ttl"`
}

// EmailConfig defines configuration for Email-based MFA
type EmailConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	Provider   string `yaml:"provider" json:"provider"`     // smtp, aws-ses, etc.
	Template   string `yaml:"template" json:"template"`     // Email template
	CodeLength int    `yaml:"code_length" json:"code_length"` // Length of email code
	CodeTTL    time.Duration `yaml:"code_ttl" json:"code_ttl"`
}

// WebAuthnConfig defines configuration for WebAuthn/FIDO2 authentication
type WebAuthnConfig struct {
	Enabled         bool     `yaml:"enabled" json:"enabled"`
	RelyingPartyID  string   `yaml:"relying_party_id" json:"relying_party_id"`
	RelyingPartyName string  `yaml:"relying_party_name" json:"relying_party_name"`
	Origin          string   `yaml:"origin" json:"origin"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	AuthenticatorSelection json.RawMessage `yaml:"authenticator_selection" json:"authenticator_selection"`
}

// MFAStorage defines the interface for persisting MFA data
type MFAStorage interface {
	// Factor management
	CreateFactor(ctx context.Context, factor *MFAFactor) error
	GetFactor(ctx context.Context, factorID uuid.UUID) (*MFAFactor, error)
	GetUserFactors(ctx context.Context, userID uuid.UUID) ([]*MFAFactor, error)
	UpdateFactor(ctx context.Context, factor *MFAFactor) error
	DeleteFactor(ctx context.Context, factorID uuid.UUID) error

	// Enrollment sessions
	CreateEnrollmentSession(ctx context.Context, session *EnrollmentSession) error
	GetEnrollmentSession(ctx context.Context, sessionID uuid.UUID) (*EnrollmentSession, error)
	DeleteEnrollmentSession(ctx context.Context, sessionID uuid.UUID) error

	// Authentication challenges
	CreateAuthChallenge(ctx context.Context, challenge *AuthChallenge) error
	GetAuthChallenge(ctx context.Context, challengeID uuid.UUID) (*AuthChallenge, error)
	UpdateAuthChallenge(ctx context.Context, challenge *AuthChallenge) error
	DeleteAuthChallenge(ctx context.Context, challengeID uuid.UUID) error

	// MFA challenges
	CreateMFAChallenge(ctx context.Context, challenge *MFAChallenge) error
	GetMFAChallenge(ctx context.Context, challengeID uuid.UUID) (*MFAChallenge, error)
	UpdateMFAChallenge(ctx context.Context, challenge *MFAChallenge) error
	DeleteMFAChallenge(ctx context.Context, challengeID uuid.UUID) error

	// Backup codes
	StoreBackupCodes(ctx context.Context, factorID uuid.UUID, codes []string) error
	ValidateBackupCode(ctx context.Context, factorID uuid.UUID, code string) (bool, error)
	InvalidateBackupCode(ctx context.Context, factorID uuid.UUID, code string) error

	// Audit and security
	RecordMFAEvent(ctx context.Context, event *MFAEvent) error
	GetMFAEvents(ctx context.Context, userID uuid.UUID, limit int) ([]*MFAEvent, error)
}

// MFAEvent represents an audit event for MFA operations
type MFAEvent struct {
	ID          uuid.UUID         `json:"id"`
	UserID      uuid.UUID         `json:"user_id"`
	FactorID    *uuid.UUID        `json:"factor_id,omitempty"`
	ChallengeID *uuid.UUID        `json:"challenge_id,omitempty"`
	EventType   string            `json:"event_type"` // enrollment, challenge, verification, etc.
	Result      string            `json:"result"`     // success, failure, error
	Metadata    map[string]string `json:"metadata"`
	IPAddress   string            `json:"ip_address"`
	UserAgent   string            `json:"user_agent"`
	CreatedAt   time.Time         `json:"created_at"`
}

// MFAError represents an MFA-specific error with additional context
type MFAError struct {
	Type    string            `json:"type"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
	Cause   error             `json:"-"`
}

func (e *MFAError) Error() string {
	return e.Message
}

func (e *MFAError) Unwrap() error {
	return e.Cause
}

// Common MFA error types
const (
	ErrTypeInvalidProvider      = "invalid_provider"
	ErrTypeProviderNotFound     = "provider_not_found"
	ErrTypeFactorNotFound       = "factor_not_found"
	ErrTypeInvalidChallenge     = "invalid_challenge"
	ErrTypeChallengeExpired     = "challenge_expired"
	ErrTypeVerificationFailed   = "verification_failed"
	ErrTypeEnrollmentFailed     = "enrollment_failed"
	ErrTypeRateLimitExceeded    = "rate_limit_exceeded"
	ErrTypeInvalidConfiguration = "invalid_configuration"
	ErrTypeStorageError         = "storage_error"
)

// NewMFAError creates a new MFA error with the specified type and message
func NewMFAError(errorType, message string, cause error) *MFAError {
	return &MFAError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}
