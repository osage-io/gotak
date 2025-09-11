package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/dfedick/gotak/pkg/logger"
)

var (
	ErrPasswordTooShort        = errors.New("password is too short")
	ErrPasswordTooLong         = errors.New("password is too long")
	ErrPasswordMissingUpper    = errors.New("password must contain at least one uppercase letter")
	ErrPasswordMissingLower    = errors.New("password must contain at least one lowercase letter")
	ErrPasswordMissingNumber   = errors.New("password must contain at least one number")
	ErrPasswordMissingSpecial  = errors.New("password must contain at least one special character")
	ErrPasswordCommonPassword  = errors.New("password is too common and easily guessable")
	ErrPasswordContainsUsername = errors.New("password must not contain username")
	ErrPasswordRepeatingChars  = errors.New("password contains too many repeating characters")
	ErrPasswordSequential     = errors.New("password contains sequential characters")
	ErrPasswordRecentlyUsed   = errors.New("password was recently used")
)

// PasswordPolicy defines password security requirements
type PasswordPolicy struct {
	// Length requirements
	MinLength int `mapstructure:"min_length"`
	MaxLength int `mapstructure:"max_length"`
	
	// Character requirements
	RequireUppercase     bool `mapstructure:"require_uppercase"`
	RequireLowercase     bool `mapstructure:"require_lowercase"`
	RequireNumbers       bool `mapstructure:"require_numbers"`
	RequireSpecialChars  bool `mapstructure:"require_special_chars"`
	MinUniqueChars       int  `mapstructure:"min_unique_chars"`
	
	// Pattern restrictions
	ForbidCommonPasswords  bool `mapstructure:"forbid_common_passwords"`
	ForbidUsernameInPassword bool `mapstructure:"forbid_username_in_password"`
	MaxRepeatingChars     int  `mapstructure:"max_repeating_chars"`
	ForbidSequentialChars bool `mapstructure:"forbid_sequential_chars"`
	
	// History and expiration
	PasswordHistorySize   int           `mapstructure:"password_history_size"`
	PasswordMaxAge        time.Duration `mapstructure:"password_max_age"`
	WarnBeforeExpiration  time.Duration `mapstructure:"warn_before_expiration"`
	
	// Account lockout policy
	MaxFailedAttempts     int           `mapstructure:"max_failed_attempts"`
	LockoutDuration       time.Duration `mapstructure:"lockout_duration"`
	LockoutProgressiveDelay bool        `mapstructure:"lockout_progressive_delay"`
	
	// Custom character sets
	AllowedSpecialChars   string `mapstructure:"allowed_special_chars"`
	ForbiddenChars        string `mapstructure:"forbidden_chars"`
}

// DefaultPasswordPolicy returns a military-grade default password policy
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		// Strong length requirements
		MinLength: 12,
		MaxLength: 128,
		
		// Character requirements
		RequireUppercase:    true,
		RequireLowercase:    true,
		RequireNumbers:      true,
		RequireSpecialChars: true,
		MinUniqueChars:      8,
		
		// Pattern restrictions
		ForbidCommonPasswords:    true,
		ForbidUsernameInPassword: true,
		MaxRepeatingChars:       2,
		ForbidSequentialChars:   true,
		
		// History and expiration
		PasswordHistorySize:  5,
		PasswordMaxAge:       90 * 24 * time.Hour, // 90 days
		WarnBeforeExpiration: 7 * 24 * time.Hour,  // 7 days
		
		// Account lockout
		MaxFailedAttempts:       5,
		LockoutDuration:        15 * time.Minute,
		LockoutProgressiveDelay: true,
		
		// Allowed special characters (OWASP recommended)
		AllowedSpecialChars: `!@#$%^&*()_+-=[]{}|;:,.<>?`,
		ForbiddenChars:      `"'` + "`", // Quotes and backticks
	}
}

// PasswordValidator handles password policy validation
type PasswordValidator struct {
	policy PasswordPolicy
	logger *logger.Logger
}

// NewPasswordValidator creates a new password validator
func NewPasswordValidator(policy PasswordPolicy, logger *logger.Logger) *PasswordValidator {
	return &PasswordValidator{
		policy: policy,
		logger: logger,
	}
}

// ValidatePassword validates a password against the security policy
func (pv *PasswordValidator) ValidatePassword(password, username string) error {
	var errors []string
	
	// Length validation
	if len(password) < pv.policy.MinLength {
		errors = append(errors, fmt.Sprintf("password must be at least %d characters long", pv.policy.MinLength))
	}
	
	if len(password) > pv.policy.MaxLength {
		errors = append(errors, fmt.Sprintf("password must be no more than %d characters long", pv.policy.MaxLength))
	}
	
	// Character requirements
	if pv.policy.RequireUppercase && !containsUppercase(password) {
		errors = append(errors, "password must contain at least one uppercase letter")
	}
	
	if pv.policy.RequireLowercase && !containsLowercase(password) {
		errors = append(errors, "password must contain at least one lowercase letter")
	}
	
	if pv.policy.RequireNumbers && !containsNumber(password) {
		errors = append(errors, "password must contain at least one number")
	}
	
	if pv.policy.RequireSpecialChars && !containsSpecialChar(password, pv.policy.AllowedSpecialChars) {
		errors = append(errors, fmt.Sprintf("password must contain at least one special character from: %s", pv.policy.AllowedSpecialChars))
	}
	
	// Unique characters
	if pv.policy.MinUniqueChars > 0 && countUniqueChars(password) < pv.policy.MinUniqueChars {
		errors = append(errors, fmt.Sprintf("password must contain at least %d unique characters", pv.policy.MinUniqueChars))
	}
	
	// Forbidden characters
	if pv.policy.ForbiddenChars != "" && containsForbiddenChars(password, pv.policy.ForbiddenChars) {
		errors = append(errors, fmt.Sprintf("password contains forbidden characters: %s", pv.policy.ForbiddenChars))
	}
	
	// Username check
	if pv.policy.ForbidUsernameInPassword && containsUsername(password, username) {
		errors = append(errors, "password must not contain username")
	}
	
	// Repeating characters
	if pv.policy.MaxRepeatingChars > 0 && hasExcessiveRepeatingChars(password, pv.policy.MaxRepeatingChars) {
		errors = append(errors, fmt.Sprintf("password contains more than %d repeating characters", pv.policy.MaxRepeatingChars))
	}
	
	// Sequential characters
	if pv.policy.ForbidSequentialChars && hasSequentialChars(password) {
		errors = append(errors, "password contains sequential characters (e.g., abc, 123)")
	}
	
	// Common passwords
	if pv.policy.ForbidCommonPasswords && isCommonPassword(password) {
		errors = append(errors, "password is too common and easily guessable")
	}
	
	// If there are validation errors, return them
	if len(errors) > 0 {
		pv.logger.Warn().
			Str("username", username).
			Int("errors", len(errors)).
			Msg("Password validation failed")
		
		return fmt.Errorf("password validation failed: %s", strings.Join(errors, "; "))
	}
	
	pv.logger.Debug().
		Str("username", username).
		Msg("Password validation successful")
	
	return nil
}

// GetPasswordStrengthScore returns a password strength score (0-100)
func (pv *PasswordValidator) GetPasswordStrengthScore(password string) int {
	score := 0
	
	// Length scoring (up to 25 points)
	if len(password) >= 8 {
		score += 5
	}
	if len(password) >= 12 {
		score += 10
	}
	if len(password) >= 16 {
		score += 10
	}
	
	// Character diversity (up to 40 points)
	if containsLowercase(password) {
		score += 10
	}
	if containsUppercase(password) {
		score += 10
	}
	if containsNumber(password) {
		score += 10
	}
	if containsSpecialChar(password, pv.policy.AllowedSpecialChars) {
		score += 10
	}
	
	// Unique characters (up to 15 points)
	uniqueRatio := float64(countUniqueChars(password)) / float64(len(password))
	score += int(uniqueRatio * 15)
	
	// Penalties for weak patterns (up to -20 points)
	if hasExcessiveRepeatingChars(password, 2) {
		score -= 5
	}
	if hasSequentialChars(password) {
		score -= 5
	}
	if isCommonPassword(password) {
		score -= 10
	}
	
	// Bonus for strong passwords (up to 20 points)
	if len(password) >= 20 && countUniqueChars(password) >= 15 {
		score += 20
	}
	
	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score
}

// GetPasswordStrengthLevel returns a human-readable strength level
func (pv *PasswordValidator) GetPasswordStrengthLevel(password string) string {
	score := pv.GetPasswordStrengthScore(password)
	
	switch {
	case score >= 90:
		return "Very Strong"
	case score >= 70:
		return "Strong"
	case score >= 50:
		return "Moderate"
	case score >= 30:
		return "Weak"
	default:
		return "Very Weak"
	}
}

// Helper functions for password validation

func containsUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func containsLowercase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func containsSpecialChar(s, allowedChars string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			if strings.ContainsRune(allowedChars, r) {
				return true
			}
		}
	}
	return false
}

func containsForbiddenChars(s, forbiddenChars string) bool {
	for _, r := range s {
		if strings.ContainsRune(forbiddenChars, r) {
			return true
		}
	}
	return false
}

func countUniqueChars(s string) int {
	unique := make(map[rune]bool)
	for _, r := range s {
		unique[r] = true
	}
	return len(unique)
}

func containsUsername(password, username string) bool {
	if username == "" || len(username) < 3 {
		return false
	}
	
	passwordLower := strings.ToLower(password)
	usernameLower := strings.ToLower(username)
	
	// Check if username is contained in password
	if strings.Contains(passwordLower, usernameLower) {
		return true
	}
	
	// Check if reversed username is contained
	reversedUsername := reverseString(usernameLower)
	if strings.Contains(passwordLower, reversedUsername) {
		return true
	}
	
	return false
}

func hasExcessiveRepeatingChars(password string, maxRepeating int) bool {
	if len(password) < 2 {
		return false
	}
	
	count := 1
	for i := 1; i < len(password); i++ {
		if password[i] == password[i-1] {
			count++
			if count > maxRepeating {
				return true
			}
		} else {
			count = 1
		}
	}
	
	return false
}

func hasSequentialChars(password string) bool {
	if len(password) < 3 {
		return false
	}
	
	passwordLower := strings.ToLower(password)
	
	// Check for common sequential patterns
	sequences := []string{
		"abc", "bcd", "cde", "def", "efg", "fgh", "ghi", "hij", "ijk", "jkl", "klm",
		"lmn", "mno", "nop", "opq", "pqr", "qrs", "rst", "stu", "tuv", "uvw", "vwx", "wxy", "xyz",
		"123", "234", "345", "456", "567", "678", "789", "890",
		"qwe", "wer", "ert", "rty", "tyu", "yui", "uio", "iop",
		"asd", "sdf", "dfg", "fgh", "ghj", "hjk", "jkl",
		"zxc", "xcv", "cvb", "vbn", "bnm",
	}
	
	for _, seq := range sequences {
		if strings.Contains(passwordLower, seq) || strings.Contains(passwordLower, reverseString(seq)) {
			return true
		}
	}
	
	// Check for numeric sequences
	for i := 0; i < len(password)-2; i++ {
		if unicode.IsDigit(rune(password[i])) && unicode.IsDigit(rune(password[i+1])) && unicode.IsDigit(rune(password[i+2])) {
			a, b, c := int(password[i]-'0'), int(password[i+1]-'0'), int(password[i+2]-'0')
			if (b == a+1 && c == b+1) || (b == a-1 && c == b-1) {
				return true
			}
		}
	}
	
	return false
}

func isCommonPassword(password string) bool {
	// List of common passwords (this would be much larger in production)
	commonPasswords := []string{
		"password", "123456", "123456789", "12345678", "12345", "1234567", "password123",
		"admin", "qwerty", "abc123", "Password1", "password1", "123123", "welcome",
		"login", "master", "hello", "guest", "administrator", "root", "toor",
		"pass", "test", "temp", "changeme", "secret", "letmein", "trustno1",
		"dragon", "baseball", "football", "monkey", "696969", "abc123", "mustang",
		"michael", "shadow", "superman", "696969", "123123", "batman", "trustno1",
		"1234", "12345", "123456", "1234567", "12345678", "123456789", "1234567890",
		"qwertyuiop", "asdfghjkl", "zxcvbnm", "qwerty123", "admin123", "root123",
		"user", "demo", "sample", "example", "default", "god", "love", "sex",
	}
	
	passwordLower := strings.ToLower(password)
	
	for _, common := range commonPasswords {
		if passwordLower == strings.ToLower(common) {
			return true
		}
	}
	
	// Check for simple variations (password with numbers at end)
	basePasswords := []string{"password", "admin", "user", "test", "demo"}
	for _, base := range basePasswords {
		if strings.HasPrefix(passwordLower, base) && len(passwordLower) > len(base) {
			suffix := passwordLower[len(base):]
			if isNumericSuffix(suffix) {
				return true
			}
		}
	}
	
	return false
}

func isNumericSuffix(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// PasswordComplexityReport provides detailed feedback on password strength
type PasswordComplexityReport struct {
	Score            int                    `json:"score"`
	Level            string                 `json:"level"`
	IsValid          bool                   `json:"is_valid"`
	ValidationErrors []string               `json:"validation_errors,omitempty"`
	Suggestions      []string               `json:"suggestions,omitempty"`
	Checks           map[string]bool        `json:"checks"`
}

// GeneratePasswordComplexityReport creates a detailed report about password strength
func (pv *PasswordValidator) GeneratePasswordComplexityReport(password, username string) *PasswordComplexityReport {
	report := &PasswordComplexityReport{
		Score:   pv.GetPasswordStrengthScore(password),
		Level:   pv.GetPasswordStrengthLevel(password),
		Checks:  make(map[string]bool),
	}
	
	// Run validation
	err := pv.ValidatePassword(password, username)
	report.IsValid = err == nil
	if err != nil {
		report.ValidationErrors = strings.Split(err.Error(), "; ")
	}
	
	// Individual checks
	report.Checks["min_length"] = len(password) >= pv.policy.MinLength
	report.Checks["max_length"] = len(password) <= pv.policy.MaxLength
	report.Checks["has_uppercase"] = containsUppercase(password)
	report.Checks["has_lowercase"] = containsLowercase(password)
	report.Checks["has_numbers"] = containsNumber(password)
	report.Checks["has_special"] = containsSpecialChar(password, pv.policy.AllowedSpecialChars)
	report.Checks["unique_chars"] = countUniqueChars(password) >= pv.policy.MinUniqueChars
	report.Checks["no_username"] = !containsUsername(password, username)
	report.Checks["no_repeating"] = !hasExcessiveRepeatingChars(password, pv.policy.MaxRepeatingChars)
	report.Checks["no_sequential"] = !hasSequentialChars(password)
	report.Checks["not_common"] = !isCommonPassword(password)
	
	// Generate suggestions
	if !report.Checks["min_length"] {
		report.Suggestions = append(report.Suggestions, fmt.Sprintf("Use at least %d characters", pv.policy.MinLength))
	}
	if !report.Checks["has_uppercase"] {
		report.Suggestions = append(report.Suggestions, "Add uppercase letters")
	}
	if !report.Checks["has_lowercase"] {
		report.Suggestions = append(report.Suggestions, "Add lowercase letters")
	}
	if !report.Checks["has_numbers"] {
		report.Suggestions = append(report.Suggestions, "Add numbers")
	}
	if !report.Checks["has_special"] {
		report.Suggestions = append(report.Suggestions, "Add special characters")
	}
	if !report.Checks["unique_chars"] {
		report.Suggestions = append(report.Suggestions, "Use more unique characters")
	}
	if !report.Checks["no_username"] {
		report.Suggestions = append(report.Suggestions, "Don't include username in password")
	}
	if !report.Checks["no_repeating"] {
		report.Suggestions = append(report.Suggestions, "Avoid repeating characters")
	}
	if !report.Checks["no_sequential"] {
		report.Suggestions = append(report.Suggestions, "Avoid sequential patterns (abc, 123)")
	}
	if !report.Checks["not_common"] {
		report.Suggestions = append(report.Suggestions, "Use a more unique password")
	}
	
	// Add general suggestions based on score
	if report.Score < 50 {
		report.Suggestions = append(report.Suggestions, "Consider using a password manager", "Make your password longer and more complex")
	}
	
	return report
}
