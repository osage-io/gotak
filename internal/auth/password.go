package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost
	DefaultCost = 10
	// MinPasswordLength is the minimum password length
	MinPasswordLength = 8
	// MaxPasswordLength is the maximum password length (bcrypt limit)
	MaxPasswordLength = 72
)

var (
	// ErrPasswordTooWeak is returned when password doesn't meet complexity requirements
	ErrPasswordTooWeak = errors.New("password must contain at least one uppercase letter, one lowercase letter, and one number")
	// ErrInvalidPassword is returned when password doesn't match hash
	ErrInvalidPassword = errors.New("invalid password")
)

// PasswordConfig holds password policy configuration
type PasswordConfig struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

// DefaultPasswordConfig returns the default password configuration
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		MinLength:        MinPasswordLength,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSpecial:   false,
	}
}

// HashPassword generates a bcrypt hash from a plain text password
func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password, DefaultPasswordConfig()); err != nil {
		return "", err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// ComparePassword compares a plain text password with a bcrypt hash
func ComparePassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return fmt.Errorf("failed to compare password: %w", err)
	}
	return nil
}

// ValidatePassword validates a password against the configured policy
func ValidatePassword(password string, config *PasswordConfig) error {
	if config == nil {
		config = DefaultPasswordConfig()
	}

	// Check length
	if len(password) < config.MinLength {
		return ErrPasswordTooShort
	}
	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	// Check character requirements
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Validate against requirements
	if config.RequireUppercase && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if config.RequireLowercase && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if config.RequireNumber && !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if config.RequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// GeneratePasswordResetToken generates a secure random token for password reset
func GeneratePasswordResetToken() (string, error) {
	// We'll use a random UUID for simplicity
	// In production, you might want to use crypto/rand for more entropy
	return GenerateSecureToken()
}

// IsPasswordCompromised checks if a password has been compromised
// This is a placeholder - in production, you'd check against haveibeenpwned.com API
func IsPasswordCompromised(password string) (bool, error) {
	// Common weak passwords to reject
	weakPasswords := []string{
		"password", "123456", "password123", "admin", "letmein",
		"qwerty", "abc123", "monkey", "dragon", "master",
	}

	for _, weak := range weakPasswords {
		if password == weak {
			return true, nil
		}
	}

	return false, nil
}

// ValidatePasswordStrength returns a strength score from 0-5
func ValidatePasswordStrength(password string) int {
	score := 0

	// Length scoring
	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if len(password) >= 16 {
		score++
	}

	// Complexity scoring
	if regexp.MustCompile(`[a-z]`).MatchString(password) &&
		regexp.MustCompile(`[A-Z]`).MatchString(password) {
		score++
	}
	if regexp.MustCompile(`[0-9]`).MatchString(password) {
		score++
	}
	if regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
		score++
	}

	// Cap at 5
	if score > 5 {
		score = 5
	}

	return score
}

// GetPasswordStrengthText returns a human-readable strength description
func GetPasswordStrengthText(score int) string {
	switch score {
	case 0, 1:
		return "Very Weak"
	case 2:
		return "Weak"
	case 3:
		return "Fair"
	case 4:
		return "Strong"
	case 5:
		return "Very Strong"
	default:
		return "Unknown"
	}
}

// GenerateSecureToken generates a secure random token
func GenerateSecureToken() (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Add timestamp for uniqueness
	timestamp := time.Now().UnixNano()
	token := fmt.Sprintf("%x-%d", randomBytes, timestamp)

	return token, nil
}
