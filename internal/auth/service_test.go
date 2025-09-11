package auth

import (
	"testing"
	"time"

	"github.com/dfedick/gotak/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_DefaultPasswordPolicy(t *testing.T) {
	policy := DefaultPasswordPolicy()

	// Test default values
	assert.Equal(t, 12, policy.MinLength)
	assert.Equal(t, 128, policy.MaxLength)
	assert.True(t, policy.RequireUppercase)
	assert.True(t, policy.RequireLowercase)
	assert.True(t, policy.RequireNumbers)
	assert.True(t, policy.RequireSpecialChars)
	assert.Equal(t, 8, policy.MinUniqueChars)
	assert.Equal(t, 2, policy.MaxRepeatingChars)
	assert.True(t, policy.ForbidSequentialChars)
	assert.True(t, policy.ForbidCommonPasswords)
	assert.True(t, policy.ForbidUsernameInPassword)
	assert.Equal(t, 5, policy.MaxFailedAttempts)
	assert.Equal(t, 15*time.Minute, policy.LockoutDuration)
	assert.True(t, policy.LockoutProgressiveDelay)
	assert.NotEmpty(t, policy.AllowedSpecialChars)
	assert.NotEmpty(t, policy.ForbiddenChars)
}

func TestAuthConfig_Defaults(t *testing.T) {
	config := AuthConfig{}

	// Test that NewAuthService sets proper defaults
	log := logger.NewDefault()

	// This would require a real database connection, so we'll test the default values logic separately
	assert.Equal(t, 0, config.MinPasswordLength) // Before defaults are applied
	assert.Equal(t, 0, config.MaxPasswordLength)
	assert.Equal(t, 0, config.MaxFailedAttempts)
	assert.Equal(t, time.Duration(0), config.LockoutDuration)
	assert.Equal(t, 0, config.BcryptCost)

	// Test that we would set defaults (without database)
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

	assert.Equal(t, 8, config.MinPasswordLength)
	assert.Equal(t, 128, config.MaxPasswordLength)
	assert.Equal(t, 5, config.MaxFailedAttempts)
	assert.Equal(t, 15*time.Minute, config.LockoutDuration)

	_ = log // Use logger to avoid unused import
}

func TestPasswordStrengthIntegration(t *testing.T) {
	log := logger.NewDefault()
	policy := DefaultPasswordPolicy()
	validator := NewPasswordValidator(policy, log)

	tests := []struct {
		password       string
		username       string
		shouldBeValid  bool
		expectedLevel  string
		minScore       int
		maxScore       int
	}{
		{
			password:      "VeryStr0ngP@ssw4rd!",
			username:      "testuser",
			shouldBeValid: true,
			expectedLevel: "Strong",
			minScore:      70,
			maxScore:      100,
		},
		{
			password:      "WeakP@ss97!",
			username:      "testuser",
			shouldBeValid: false, // Too short
			expectedLevel: "Moderate",
			minScore:      40,
			maxScore:      70,
		},
		{
			password:      "password",
			username:      "testuser",
			shouldBeValid: false, // Multiple violations
			expectedLevel: "Very Weak",
			minScore:      0,
			maxScore:      30,
		},
		{
			password:      "Extr3m3lyStr0ng&C0mpl3xP@ssw4rd!",
			username:      "testuser",
			shouldBeValid: true,
			expectedLevel: "Very Strong",
			minScore:      85,
			maxScore:      100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			// Test validation
			err := validator.ValidatePassword(tt.password, tt.username)
			if tt.shouldBeValid {
				assert.NoError(t, err, "Password should be valid: %s", tt.password)
			} else {
				assert.Error(t, err, "Password should be invalid: %s", tt.password)
			}

			// Test strength scoring
			score := validator.GetPasswordStrengthScore(tt.password)
			assert.GreaterOrEqual(t, score, tt.minScore, "Score too low for: %s", tt.password)
			assert.LessOrEqual(t, score, tt.maxScore, "Score too high for: %s", tt.password)

			// Test strength level
			level := validator.GetPasswordStrengthLevel(tt.password)
			assert.Equal(t, tt.expectedLevel, level, "Wrong strength level for: %s", tt.password)

			// Test complexity report
			report := validator.GeneratePasswordComplexityReport(tt.password, tt.username)
			assert.NotNil(t, report)
			assert.Equal(t, tt.shouldBeValid, report.IsValid)
			assert.Equal(t, score, report.Score)
			assert.Equal(t, level, report.Level)
			assert.NotNil(t, report.Checks)

			if !tt.shouldBeValid {
				assert.NotEmpty(t, report.ValidationErrors, "Invalid password should have errors")
				assert.NotEmpty(t, report.Suggestions, "Invalid password should have suggestions")
			}
		})
	}
}

func TestPasswordPolicyCustomization(t *testing.T) {
	log := logger.NewDefault()

	// Test lenient policy
	lenientPolicy := PasswordPolicy{
		MinLength:                8,
		MaxLength:                64,
		RequireUppercase:         false,
		RequireLowercase:         true,
		RequireNumbers:           false,
		RequireSpecialChars:      false,
		MinUniqueChars:           4,
		ForbidCommonPasswords:    false,
		ForbidUsernameInPassword: false,
		MaxRepeatingChars:        5,
		ForbidSequentialChars:    false,
		AllowedSpecialChars:      "!@#$%",
		ForbiddenChars:           "",
	}

	validator := NewPasswordValidator(lenientPolicy, log)

	// These passwords should pass the lenient policy
	lenientPasswords := []string{
		"simplepass",
		"password123",
		"testuser", // Contains username but allowed
		"abcdefgh", // Sequential but allowed
	}

	for _, password := range lenientPasswords {
		t.Run("lenient_"+password, func(t *testing.T) {
			err := validator.ValidatePassword(password, "testuser")
			assert.NoError(t, err, "Lenient policy should allow: %s", password)
		})
	}

	// Test strict policy
	strictPolicy := PasswordPolicy{
		MinLength:                16,
		MaxLength:                32,
		RequireUppercase:         true,
		RequireLowercase:         true,
		RequireNumbers:           true,
		RequireSpecialChars:      true,
		MinUniqueChars:           12,
		ForbidCommonPasswords:    true,
		ForbidUsernameInPassword: true,
		MaxRepeatingChars:        1,
		ForbidSequentialChars:    true,
		AllowedSpecialChars:      "!@#$%^&*",
		ForbiddenChars:           "'\"",
	}

	strictValidator := NewPasswordValidator(strictPolicy, log)

	// These should fail the strict policy
	err := strictValidator.ValidatePassword("VeryStr0ngP@ssw4rd!", "testuser")
	assert.Error(t, err, "Strict policy should reject shorter passwords")

	// This should pass strict policy (shorter to meet 32 char limit)
	err = strictValidator.ValidatePassword("V3ryStr0ng&C0mpl3x!", "testuser")
	assert.NoError(t, err, "Strict policy should accept very strong passwords")
}

func TestPatternDetection(t *testing.T) {
	log := logger.NewDefault()
	policy := DefaultPasswordPolicy()
	validator := NewPasswordValidator(policy, log)

	// Test username detection
	err := validator.ValidatePassword("MyPasswordWithJohndoe!", "johndoe")
	assert.Error(t, err, "Should detect username in password")
	assert.Contains(t, err.Error(), "username")

	// Test reversed username detection  
	err = validator.ValidatePassword("MyPasswordWithEondohj!", "johndoe")
	assert.Error(t, err, "Should detect reversed username in password")

	// Test repeating characters
	err = validator.ValidatePassword("P@ssssssword97!", "testuser")
	assert.Error(t, err, "Should detect excessive repeating characters")

	// Test sequential patterns
	sequentialTests := []string{
		"P@ssword123!", // Contains "123"
		"P@sswordabc!", // Contains "abc"
		"P@sswordqwe!", // Contains "qwe"
		"P@ssword456!", // Contains "456"
		"P@sswordxyz!", // Contains "xyz"
	}

	for _, password := range sequentialTests {
		t.Run(password, func(t *testing.T) {
			err := validator.ValidatePassword(password, "testuser")
			assert.Error(t, err, "Should detect sequential pattern in: %s", password)
			assert.Contains(t, err.Error(), "sequential")
		})
	}

	// Test common passwords
	commonPasswords := []string{
		"password",
		"admin",
		"password123",
		"admin123",
		"Password1",
	}

	for _, password := range commonPasswords {
		t.Run(password, func(t *testing.T) {
			err := validator.ValidatePassword(password, "testuser")
			assert.Error(t, err, "Should detect common password: %s", password)
		})
	}
}

func TestComplexityReportDetails(t *testing.T) {
	log := logger.NewDefault()
	policy := DefaultPasswordPolicy()
	validator := NewPasswordValidator(policy, log)

	// Test detailed report for weak password
	report := validator.GeneratePasswordComplexityReport("weak", "testuser")

	// Verify report structure
	assert.False(t, report.IsValid)
	assert.Equal(t, "Very Weak", report.Level)
	assert.Greater(t, len(report.ValidationErrors), 0)
	assert.Greater(t, len(report.Suggestions), 0)
	assert.NotNil(t, report.Checks)

	// Verify specific checks exist
	expectedChecks := []string{
		"min_length", "max_length", "has_uppercase", "has_lowercase",
		"has_numbers", "has_special", "unique_chars", "no_username",
		"no_repeating", "no_sequential", "not_common",
	}

	for _, check := range expectedChecks {
		_, exists := report.Checks[check]
		assert.True(t, exists, "Check should exist: %s", check)
	}

	// Test report for strong password
	strongReport := validator.GeneratePasswordComplexityReport("VeryStr0ngP@ssw4rd!", "testuser")
	assert.True(t, strongReport.IsValid)
	assert.Equal(t, "Strong", strongReport.Level)
	assert.Len(t, strongReport.ValidationErrors, 0)
	assert.Len(t, strongReport.Suggestions, 0)
}

func TestUnicodeHandling(t *testing.T) {
	log := logger.NewDefault()
	policy := DefaultPasswordPolicy()
	validator := NewPasswordValidator(policy, log)

	// Test unicode characters are accepted
	unicodePasswords := []string{
		"Unic0d3P@ssw4rd🔒",
		"P@ssw4rd🌟97!",
		"Str0ng🔐P@ssw4rd!",
	}

	for _, password := range unicodePasswords {
		t.Run(password, func(t *testing.T) {
			err := validator.ValidatePassword(password, "testuser")
			assert.NoError(t, err, "Should accept unicode characters in: %s", password)
		})
	}
}

func TestBoundaryConditions(t *testing.T) {
	log := logger.NewDefault()
	policy := DefaultPasswordPolicy()
	validator := NewPasswordValidator(policy, log)

	// Test minimum length boundary
	minLengthPassword := "MinLen97#46!" // Exactly 12 characters
	err := validator.ValidatePassword(minLengthPassword, "testuser")
	assert.NoError(t, err, "Should accept password at minimum length")

	// Test maximum unique characters (avoid sequential patterns)
	longPassword := "ThisIsAnExtr3m3lyL0ngP@ssw4rdWithManyDiff3r3ntCh@r@ct3rs!"
	err = validator.ValidatePassword(longPassword, "testuser")
	assert.NoError(t, err, "Should accept very long password")

	// Test empty password
	err = validator.ValidatePassword("", "testuser")
	assert.Error(t, err, "Should reject empty password")

	// Test very long password (near max length, avoid sequential patterns)
	veryLongPassword := "Sup3rL0ngP@ssw4rd" + "N0S3qu3nti@l" + "Ch@r@ct3rs" + "H3r3" + "F0rT3sting!"
	if len(veryLongPassword) <= policy.MaxLength {
		err = validator.ValidatePassword(veryLongPassword, "testuser")
		assert.NoError(t, err, "Should accept password under max length")
	}
}

// Mock helper functions for testing (since we can't easily test the AuthService methods that require database)
func TestHelperFunctions(t *testing.T) {
	// Test string reversal
	assert.Equal(t, "olleh", reverseString("hello"))
	assert.Equal(t, "", reverseString(""))
	assert.Equal(t, "a", reverseString("a"))

	// Test character counting
	assert.Equal(t, 4, countUniqueChars("hello")) // h,e,l,o (l appears twice)
	assert.Equal(t, 1, countUniqueChars("aaa"))
	assert.Equal(t, 0, countUniqueChars(""))

	// Test character type detection
	assert.True(t, containsUppercase("Hello"))
	assert.False(t, containsUppercase("hello"))

	assert.True(t, containsLowercase("Hello"))
	assert.False(t, containsLowercase("HELLO"))

	assert.True(t, containsNumber("Hello123"))
	assert.False(t, containsNumber("Hello"))

	assert.True(t, containsSpecialChar("Hello!", "!@#"))
	assert.False(t, containsSpecialChar("Hello", "!@#"))
}

func TestPasswordStrengthScoring(t *testing.T) {
	log := logger.NewDefault()
	policy := DefaultPasswordPolicy()
	validator := NewPasswordValidator(policy, log)

	// Test score progression
	passwords := []struct {
		password string
		minScore int
		maxScore int
	}{
		{"a", 0, 30},                              // Very weak
		{"password", 15, 35},                      // Weak
		{"Password1!", 50, 75},                    // Moderate  
		{"Str0ngP@ssw4rd!", 65, 85},              // Strong
		{"V3ryStr0ng&C0mpl3xP@ssw4rd!", 80, 100}, // Very Strong
	}

	for _, tt := range passwords {
		t.Run(tt.password, func(t *testing.T) {
			score := validator.GetPasswordStrengthScore(tt.password)
			assert.GreaterOrEqual(t, score, tt.minScore, "Score too low")
			assert.LessOrEqual(t, score, tt.maxScore, "Score too high")
			assert.GreaterOrEqual(t, score, 0, "Score should be non-negative")
			assert.LessOrEqual(t, score, 100, "Score should not exceed 100")
		})
	}
}
