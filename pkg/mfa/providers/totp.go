// Package providers contains implementations of MFA providers for different authentication methods
package providers

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"

	"github.com/dfedick/gotak/pkg/mfa"
)

// TOTPProvider implements the MFAProvider interface for Time-based One-Time Password authentication
type TOTPProvider struct {
	config *mfa.TOTPConfig
}

// NewTOTPProvider creates a new TOTP provider with the given configuration
func NewTOTPProvider(config *mfa.TOTPConfig) *TOTPProvider {
	return &TOTPProvider{
		config: config,
	}
}

// GetType returns the MFA type this provider handles
func (p *TOTPProvider) GetType() mfa.MFAType {
	return mfa.MFATypeTOTP
}

// GenerateSecret creates a new TOTP secret for enrollment
func (p *TOTPProvider) GenerateSecret(ctx context.Context, userID uuid.UUID, metadata map[string]string) (*mfa.MFASecret, error) {
	// Generate cryptographically secure random key
	key := make([]byte, 20) // 160-bit key as recommended by RFC 6238
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	// Encode key as base32 (required by TOTP standard)
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(key)

	// Get account name from metadata or use user ID
	accountName := userID.String()
	if name, exists := metadata["account_name"]; exists && name != "" {
		accountName = name
	}

	// Create TOTP URL for QR code generation
	totpURL := p.generateTOTPURL(accountName, secret)

	// Generate QR code image
	qrCode, err := qrcode.Encode(totpURL, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return &mfa.MFASecret{
		ID:       uuid.New(),
		UserID:   userID,
		Type:     mfa.MFATypeTOTP,
		Secret:   secret,
		Metadata: metadata,
		QRCode:   qrCode,
		URL:      totpURL,
	}, nil
}

// VerifyEnrollment validates that enrollment is complete and working
func (p *TOTPProvider) VerifyEnrollment(ctx context.Context, secret *mfa.MFASecret, verificationCode string) error {
	if secret.Type != mfa.MFATypeTOTP {
		return fmt.Errorf("invalid secret type for TOTP provider: %s", secret.Type)
	}

	// Validate the verification code
	valid, err := p.validateTOTP(secret.Secret, verificationCode)
	if err != nil {
		return fmt.Errorf("failed to validate TOTP code: %w", err)
	}

	if !valid {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "invalid TOTP verification code", nil)
	}

	return nil
}

// CreateChallenge generates a new authentication challenge
func (p *TOTPProvider) CreateChallenge(ctx context.Context, factor *mfa.MFAFactor) (*mfa.MFAChallenge, error) {
	if factor.Type != mfa.MFATypeTOTP {
		return nil, fmt.Errorf("invalid factor type for TOTP provider: %s", factor.Type)
	}

	// TOTP doesn't require server-side challenge generation
	// The challenge is implicit - the current time window
	challenge := &mfa.MFAChallenge{
		ID:          uuid.New(),
		FactorID:    factor.ID,
		Type:        mfa.MFATypeTOTP,
		Status:      mfa.ChallengeStatusPending,
		Challenge:   "", // No explicit challenge for TOTP
		Attempts:    0,
		MaxAttempts: 3, // Allow 3 attempts per challenge
		ExpiresAt:   time.Now().Add(5 * time.Minute), // 5 minute challenge window
		CreatedAt:   time.Now(),
	}

	return challenge, nil
}

// VerifyChallenge validates a user's response to an authentication challenge
func (p *TOTPProvider) VerifyChallenge(ctx context.Context, challenge *mfa.MFAChallenge, response string) error {
	if challenge.Type != mfa.MFATypeTOTP {
		return fmt.Errorf("invalid challenge type for TOTP provider: %s", challenge.Type)
	}

	// Get the factor to retrieve the secret
	// Note: This would typically be done by the MFA manager, but for completeness
	// we assume the secret is available through some mechanism
	// In practice, the manager would pass the factor or secret to this method

	// Note: In practice, the MFA manager would provide the factor's secret
	// For now, we'll return an error indicating that this method needs the factor
	return fmt.Errorf("TOTP challenge verification requires factor secret - use VerifyTOTPWithFactor method")
}

// ValidateConfiguration checks if the provider configuration is valid
func (p *TOTPProvider) ValidateConfiguration() error {
	if p.config == nil {
		return fmt.Errorf("TOTP configuration cannot be nil")
	}

	if !p.config.Enabled {
		return fmt.Errorf("TOTP provider is disabled")
	}

	if p.config.Issuer == "" {
		return fmt.Errorf("TOTP issuer cannot be empty")
	}

	// Validate algorithm
	switch p.config.Algorithm {
	case "SHA1", "SHA256", "SHA512":
		// Valid algorithms
	case "":
		p.config.Algorithm = "SHA1" // Default
	default:
		return fmt.Errorf("invalid TOTP algorithm: %s", p.config.Algorithm)
	}

	// Validate digits
	if p.config.Digits == 0 {
		p.config.Digits = 6 // Default
	}
	if p.config.Digits != 6 && p.config.Digits != 8 {
		return fmt.Errorf("TOTP digits must be 6 or 8, got: %d", p.config.Digits)
	}

	// Validate period
	if p.config.Period == 0 {
		p.config.Period = 30 // Default 30 seconds
	}
	if p.config.Period < 15 || p.config.Period > 300 {
		return fmt.Errorf("TOTP period must be between 15 and 300 seconds, got: %d", p.config.Period)
	}

	// Validate skew
	if p.config.Skew == 0 {
		p.config.Skew = 1 // Default skew of 1 time step
	}
	if p.config.Skew < 0 || p.config.Skew > 5 {
		return fmt.Errorf("TOTP skew must be between 0 and 5, got: %d", p.config.Skew)
	}

	return nil
}

// Helper methods

// generateTOTPURL creates a TOTP URL for QR code generation according to the Key URI Format
// https://github.com/google/google-authenticator/wiki/Key-Uri-Format
func (p *TOTPProvider) generateTOTPURL(accountName, secret string) string {
	// Create URL with proper escaping
	u := url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   "/" + url.PathEscape(p.config.Issuer+":"+accountName),
	}

	// Add query parameters
	values := url.Values{}
	values.Set("secret", secret)
	values.Set("issuer", p.config.Issuer)
	values.Set("algorithm", p.config.Algorithm)
	values.Set("digits", fmt.Sprintf("%d", p.config.Digits))
	values.Set("period", fmt.Sprintf("%d", p.config.Period))

	u.RawQuery = values.Encode()

	return u.String()
}

// validateTOTP validates a TOTP code against a secret
func (p *TOTPProvider) validateTOTP(secret, code string) (bool, error) {
	// Configure TOTP options based on our settings
	opts := totp.ValidateOpts{
		Period:    uint(p.config.Period),
		Skew:      uint(p.config.Skew),
		Digits:    otp.Digits(p.config.Digits),
		Algorithm: p.getAlgorithm(),
	}

	// Validate the code
	valid, err := totp.ValidateCustom(code, secret, time.Now().UTC(), opts)
	if err != nil {
		return false, fmt.Errorf("TOTP validation failed: %w", err)
	}

	return valid, nil
}

// getAlgorithm converts string algorithm to otp.Algorithm
func (p *TOTPProvider) getAlgorithm() otp.Algorithm {
	switch p.config.Algorithm {
	case "SHA256":
		return otp.AlgorithmSHA256
	case "SHA512":
		return otp.AlgorithmSHA512
	default:
		return otp.AlgorithmSHA1
	}
}

// VerifyTOTPWithFactor is a helper method for the MFA manager to verify TOTP with factor access
func (p *TOTPProvider) VerifyTOTPWithFactor(factor *mfa.MFAFactor, code string) error {
	if factor.Type != mfa.MFATypeTOTP {
		return fmt.Errorf("invalid factor type for TOTP verification: %s", factor.Type)
	}

	valid, err := p.validateTOTP(factor.Secret, code)
	if err != nil {
		return err
	}

	if !valid {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "invalid TOTP code", nil)
	}

	return nil
}

// GetCurrentTOTP generates the current TOTP code for testing purposes
// This should only be used in development/testing environments
func (p *TOTPProvider) GetCurrentTOTP(secret string) (string, error) {
	// Generate current TOTP code
	code, err := totp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP code: %w", err)
	}

	return code, nil
}

// GetTOTPAtTime generates TOTP code for a specific time (useful for testing)
func (p *TOTPProvider) GetTOTPAtTime(secret string, t time.Time) (string, error) {
	code, err := totp.GenerateCode(secret, t.UTC())
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP code for time %v: %w", t, err)
	}

	return code, nil
}
