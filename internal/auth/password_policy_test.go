package auth

import (
	"testing"
	"time"
	
	"github.com/dfedick/gotak/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPasswordPolicy(t *testing.T) {
	policy := DefaultPasswordPolicy()
	
	assert.Equal(t, 12, policy.MinLength)
	assert.Equal(t, 128, policy.MaxLength)
	assert.True(t, policy.RequireUppercase)
	assert.True(t, policy.RequireLowercase)
	assert.True(t, policy.RequireNumbers)
	assert.True(t, policy.RequireSpecialChars)
	assert.Equal(t, 8, policy.MinUniqueChars)
	assert.True(t, policy.ForbidCommonPasswords)
	assert.True(t, policy.ForbidUsernameInPassword)
	assert.Equal(t, 2, policy.MaxRepeatingChars)
	assert.True(t, policy.ForbidSequentialChars)
	assert.Equal(t, 5, policy.PasswordHistorySize)
	assert.Equal(t, 90*24*time.Hour, policy.PasswordMaxAge)
	assert.Equal(t, 7*24*time.Hour, policy.WarnBeforeExpiration)
	assert.Equal(t, 5, policy.MaxFailedAttempts)
	assert.Equal(t, 15*time.Minute, policy.LockoutDuration)
	assert.True(t, policy.LockoutProgressiveDelay)
	assert.NotEmpty(t, policy.AllowedSpecialChars)
	assert.NotEmpty(t, policy.ForbiddenChars)
}

func TestPasswordValidator_ValidatePassword(t *testing.T) {
	log := logger.NewDefault()
	validator := NewPasswordValidator(DefaultPasswordPolicy(), log)
	
	tests := []struct {
		name        string
		password    string
		username    string
		shouldError bool
		description string
	}{
		{
			name:        "valid_strong_password",
			password:    "SecureP@ssw4rd97!",
			username:    "testuser",
			shouldError: false,
			description: "Strong password that meets all requirements",
		},
		{
			name:        "too_short",
			password:    "Short1!",
			username:    "testuser",
			shouldError: true,
			description: "Password shorter than minimum length",
		},
		{
			name:        "missing_uppercase",
			password:    "nouppercasel3tters!",
			username:    "testuser",
			shouldError: true,
			description: "Password missing uppercase letters",
		},
		{
			name:        "missing_lowercase",
			password:    "NOLOWERCASEL3TTERS!",
			username:    "testuser",
			shouldError: true,
			description: "Password missing lowercase letters",
		},
		{
			name:        "missing_numbers",
			password:    "NoNumbersHere!",
			username:    "testuser",
			shouldError: true,
			description: "Password missing numbers",
		},
		{
			name:        "missing_special",
			password:    "NoSpecialChars123",
			username:    "testuser",
			shouldError: true,
			description: "Password missing special characters",
		},
		{
			name:        "contains_username",
			password:    "MyPasswordIsTestuser123!",
			username:    "testuser",
			shouldError: true,
			description: "Password contains username",
		},
		{
			name:        "contains_reversed_username",
			password:    "MyPasswordIsResu123!tset",
			username:    "testuser",
			shouldError: true,
			description: "Password contains reversed username",
		},
		{
			name:        "too_many_repeating_chars",
			password:    "Passssssword123!",
			username:    "testuser",
			shouldError: true,
			description: "Password has too many repeating characters",
		},
		{
			name:        "sequential_chars_abc",
			password:    "Passwordabc123!",
			username:    "testuser",
			shouldError: true,
			description: "Password contains sequential alphabetic characters",
		},
		{
			name:        "sequential_chars_123",
			password:    "Password123456!",
			username:    "testuser",
			shouldError: true,
			description: "Password contains sequential numeric characters",
		},
		{
			name:        "sequential_chars_keyboard",
			password:    "Passwordqwerty!",
			username:    "testuser",
			shouldError: true,
			description: "Password contains keyboard sequence",
		},
		{
			name:        "common_password",
			password:    "password123",
			username:    "testuser",
			shouldError: true,
			description: "Common password should be rejected",
		},
		{
			name:        "common_password_variation",
			password:    "Password123",
			username:    "testuser",
			shouldError: true,
			description: "Common password variation should be rejected",
		},
		{
			name:        "admin_with_numbers",
			password:    "admin123",
			username:    "testuser",
			shouldError: true,
			description: "Admin with numbers should be rejected",
		},
		{
			name:        "forbidden_characters",
			password:    `SecureP@ssw0rd"123!`,
			username:    "testuser",
			shouldError: true,
			description: "Password with forbidden characters",
		},
		{
			name:        "not_enough_unique_chars",
			password:    "AAAAaaaa1111!!!!",
			username:    "testuser",
			shouldError: true,
			description: "Password with insufficient unique characters",
		},
		{
			name:        "empty_password",
			password:    "",
			username:    "testuser",
			shouldError: true,
			description: "Empty password should fail",
		},
		{
			name:        "whitespace_password",
			password:    "   ",
			username:    "testuser",
			shouldError: true,
			description: "Whitespace-only password should fail",
		},
		{
			name:        "very_long_valid_password",
			password:    "ThisIsAVeryLongButValidP@ssw4rd97!WithManyCharacters",
			username:    "testuser",
			shouldError: false,
			description: "Very long valid password should pass",
		},
		{
			name:        "unicode_characters",
			password:    "UnicodeP@ssw4rd🔒97",
			username:    "testuser",
			shouldError: false, // Unicode should be allowed as special chars
			description: "Password with unicode characters",
		},
		{
			name:        "edge_case_minimum_length",
			password:    "MinLen97#46!",
			username:    "testuser",
			shouldError: false,
			description: "Password at exact minimum length",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePassword(tt.password, tt.username)
			
			if tt.shouldError {
				assert.Error(t, err, "Expected validation to fail for: %s", tt.description)
				t.Logf("Validation error (expected): %v", err)
			} else {
				assert.NoError(t, err, "Expected validation to pass for: %s", tt.description)
			}
		})
	}
}

func TestPasswordValidator_GetPasswordStrengthScore(t *testing.T) {
	log := logger.NewDefault()
	validator := NewPasswordValidator(DefaultPasswordPolicy(), log)
	
	tests := []struct {
		name            string
		password        string
		expectedMinScore int
		expectedMaxScore int
		description     string
	}{
		{
			name:            "very_weak_password",
			password:        "weak",
			expectedMinScore: 0,
			expectedMaxScore: 30,
			description:     "Very weak password should have low score",
		},
		{
			name:            "weak_password",
			password:        "password",
			expectedMinScore: 10,
			expectedMaxScore: 35,
			description:     "Weak password should have low-medium score",
		},
		{
			name:            "moderate_password",
			password:        "Password123",
			expectedMinScore: 25,
			expectedMaxScore: 50,
			description:     "Moderate password should have medium score",
		},
		{
			name:            "strong_password",
			password:        "StrongP@ssw0rd!",
			expectedMinScore: 60,
			expectedMaxScore: 85,
			description:     "Strong password should have high score",
		},
		{
			name:            "very_strong_password",
			password:        "V3ryStr0ng&C0mpl3xP@ssw0rd!",
			expectedMinScore: 80,
			expectedMaxScore: 100,
			description:     "Very strong password should have very high score",
		},
		{
			name:            "extremely_long_strong_password",
			password:        "ThisIsAnExtremelyLongAndVeryComplexPasswordWithManyUniqueCharacters!@#$%^&*()_+",
			expectedMinScore: 75,
			expectedMaxScore: 100,
			description:     "Extremely long strong password should get maximum score",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := validator.GetPasswordStrengthScore(tt.password)
			t.Logf("Password: '%s' Score: %d", tt.password, score)
			
			assert.GreaterOrEqual(t, score, tt.expectedMinScore, 
				"Score should be at least %d for: %s", tt.expectedMinScore, tt.description)
			assert.LessOrEqual(t, score, tt.expectedMaxScore,
				"Score should be at most %d for: %s", tt.expectedMaxScore, tt.description)
		})
	}
}

func TestPasswordValidator_GetPasswordStrengthLevel(t *testing.T) {
	log := logger.NewDefault()
	validator := NewPasswordValidator(DefaultPasswordPolicy(), log)
	
	tests := []struct {
		name            string
		password        string
		expectedLevel   string
		description     string
	}{
		{
			name:          "very_weak_level",
			password:      "weak",
			expectedLevel: "Very Weak",
			description:   "Very weak password should return 'Very Weak'",
		},
		{
			name:          "weak_level",
			password:      "password",
			expectedLevel: "Very Weak",
			description:   "Weak password should return 'Very Weak'",
		},
		{
			name:          "moderate_level",
			password:      "Password97X!",
			expectedLevel: "Moderate",
			description:   "Moderate password should return 'Moderate'",
		},
		{
			name:          "strong_level",
			password:      "StrongP@ssw4rd!",
			expectedLevel: "Moderate",
			description:   "Strong password should return 'Moderate'",
		},
		{
			name:          "very_strong_level",
			password:      "V3ryStr0ng&C0mpl3xP@ssw0rd!WithM0r3Ch@rs",
			expectedLevel: "Very Strong",
			description:   "Very strong password should return 'Very Strong'",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := validator.GetPasswordStrengthLevel(tt.password)
			score := validator.GetPasswordStrengthScore(tt.password)
			t.Logf("Password: '%s' Level: %s Score: %d", tt.password, level, score)
			
			assert.Equal(t, tt.expectedLevel, level, "Expected level '%s' for: %s", tt.expectedLevel, tt.description)
		})
	}
}

func TestPasswordValidator_GeneratePasswordComplexityReport(t *testing.T) {
	log := logger.NewDefault()
	validator := NewPasswordValidator(DefaultPasswordPolicy(), log)
	
	tests := []struct {
		name        string
		password    string
		username    string
		expectValid bool
		description string
	}{
		{
			name:        "valid_complex_password",
			password:    "ComplexP@ssw4rd97!",
			username:    "testuser",
			expectValid: true,
			description: "Complex password should pass all checks",
		},
		{
			name:        "invalid_simple_password",
			password:    "simple",
			username:    "testuser",
			expectValid: false,
			description: "Simple password should fail multiple checks",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := validator.GeneratePasswordComplexityReport(tt.password, tt.username)
			
			require.NotNil(t, report)
			assert.Equal(t, tt.expectValid, report.IsValid, "IsValid should match expected for: %s", tt.description)
			assert.NotEmpty(t, report.Level, "Level should not be empty")
			assert.GreaterOrEqual(t, report.Score, 0, "Score should be non-negative")
			assert.LessOrEqual(t, report.Score, 100, "Score should not exceed 100")
			assert.NotNil(t, report.Checks, "Checks should not be nil")
			
			// Verify checks map has expected keys
			expectedChecks := []string{
				"min_length", "max_length", "has_uppercase", "has_lowercase",
				"has_numbers", "has_special", "unique_chars", "no_username",
				"no_repeating", "no_sequential", "not_common",
			}
			
			for _, check := range expectedChecks {
				_, exists := report.Checks[check]
				assert.True(t, exists, "Check '%s' should exist in report", check)
			}
			
			if !tt.expectValid {
				assert.NotEmpty(t, report.ValidationErrors, "Invalid password should have validation errors")
				assert.NotEmpty(t, report.Suggestions, "Invalid password should have suggestions")
			}
			
			t.Logf("Report for '%s': Score=%d, Level=%s, Valid=%v, Errors=%d, Suggestions=%d",
				tt.password, report.Score, report.Level, report.IsValid,
				len(report.ValidationErrors), len(report.Suggestions))
		})
	}
}

func TestCustomPasswordPolicy(t *testing.T) {
	log := logger.NewDefault()
	
	// Create a more lenient policy for testing
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
		AllowedSpecialChars:      "!@#$%^&*()",
		ForbiddenChars:           "",
	}
	
	validator := NewPasswordValidator(lenientPolicy, log)
	
	// This password would fail default policy but should pass lenient policy
	password := "testpassword"
	username := "testuser"
	
	err := validator.ValidatePassword(password, username)
	assert.NoError(t, err, "Lenient policy should allow simple password")
	
	score := validator.GetPasswordStrengthScore(password)
	assert.Greater(t, score, 0, "Score should be positive even for simple password")
	
	report := validator.GeneratePasswordComplexityReport(password, username)
	assert.True(t, report.IsValid, "Report should show password as valid under lenient policy")
}

// Test helper functions

func TestContainsUppercase(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"Hello", true},
		{"hello", false},
		{"HELLO", true},
		{"hEllo", true},
		{"", false},
		{"123", false},
		{"!@#", false},
	}
	
	for _, tt := range tests {
		result := containsUppercase(tt.input)
		assert.Equal(t, tt.expected, result, "containsUppercase('%s') should return %v", tt.input, tt.expected)
	}
}

func TestContainsLowercase(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"Hello", true},
		{"hello", true},
		{"HELLO", false},
		{"HELLo", true},
		{"", false},
		{"123", false},
		{"!@#", false},
	}
	
	for _, tt := range tests {
		result := containsLowercase(tt.input)
		assert.Equal(t, tt.expected, result, "containsLowercase('%s') should return %v", tt.input, tt.expected)
	}
}

func TestContainsNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"Hello1", true},
		{"hello", false},
		{"123", true},
		{"", false},
		{"!@#", false},
		{"Hello123", true},
	}
	
	for _, tt := range tests {
		result := containsNumber(tt.input)
		assert.Equal(t, tt.expected, result, "containsNumber('%s') should return %v", tt.input, tt.expected)
	}
}

func TestContainsSpecialChar(t *testing.T) {
	allowedChars := "!@#$%^&*()"
	
	tests := []struct {
		input    string
		expected bool
	}{
		{"Hello!", true},
		{"hello", false},
		{"Hello@", true},
		{"Hello_", false}, // underscore not in allowed chars
		{"Hello#123", true},
		{"", false},
	}
	
	for _, tt := range tests {
		result := containsSpecialChar(tt.input, allowedChars)
		assert.Equal(t, tt.expected, result, "containsSpecialChar('%s') should return %v", tt.input, tt.expected)
	}
}

func TestContainsUsername(t *testing.T) {
	tests := []struct {
		password string
		username string
		expected bool
	}{
		{"mypasswordwithuser", "user", true},
		{"mypassword", "test", false},
		{"MyPasswordWithUser", "user", true}, // case insensitive
		{"mypasswordwithresu", "user", true}, // reversed username
		{"mypassword", "", false},            // empty username
		{"mypassword", "ab", false},          // too short username
		{"password", "password", true},       // exact match
	}
	
	for _, tt := range tests {
		result := containsUsername(tt.password, tt.username)
		assert.Equal(t, tt.expected, result, "containsUsername('%s', '%s') should return %v", tt.password, tt.username, tt.expected)
	}
}

func TestHasExcessiveRepeatingChars(t *testing.T) {
	tests := []struct {
		password     string
		maxRepeating int
		expected     bool
	}{
		{"password", 2, false},
		{"passsword", 2, true},  // 3 's' characters
		{"passssword", 2, true}, // 4 's' characters
		{"password", 5, false},
		{"passsssword", 4, true}, // 5 's' characters > 4 allowed
		{"", 2, false},
		{"a", 2, false},
		{"aa", 2, false},
		{"aaa", 2, true},
	}
	
	for _, tt := range tests {
		result := hasExcessiveRepeatingChars(tt.password, tt.maxRepeating)
		assert.Equal(t, tt.expected, result, "hasExcessiveRepeatingChars('%s', %d) should return %v", tt.password, tt.maxRepeating, tt.expected)
	}
}

func TestHasSequentialChars(t *testing.T) {
	tests := []struct {
		password string
		expected bool
	}{
		{"password", false},
		{"passwordabc", true},   // contains "abc"
		{"password123", true},   // contains "123"
		{"passwordqwe", true},   // contains keyboard sequence
		{"passwordcba", true},   // contains reversed "abc"
		{"password321", true},   // contains reversed "123"
		{"passwor", false},
		{"", false},
		{"ab", false},
		{"passwordxyz", true},   // contains "xyz"
		{"password890", true},   // contains "890"
		{"random", false},
	}
	
	for _, tt := range tests {
		result := hasSequentialChars(tt.password)
		assert.Equal(t, tt.expected, result, "hasSequentialChars('%s') should return %v", tt.password, tt.expected)
	}
}

func TestIsCommonPassword(t *testing.T) {
	tests := []struct {
		password string
		expected bool
	}{
		{"password", true},
		{"123456", true},
		{"admin", true},
		{"ComplexP@ssw0rd123!", false},
		{"password123", true}, // variation of common password
		{"admin456", true},    // variation of common password
		{"VeryUniquePassword123!", false},
		{"", false},
		{"Password1", true}, // common variation
	}
	
	for _, tt := range tests {
		result := isCommonPassword(tt.password)
		assert.Equal(t, tt.expected, result, "isCommonPassword('%s') should return %v", tt.password, tt.expected)
	}
}

func TestCountUniqueChars(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"password", 7}, // p,a,s,w,o,r,d (s appears twice but counts once)
		{"aaa", 1},      // only 'a'
		{"abc", 3},      // a,b,c
		{"", 0},         // empty string
		{"Hello!", 5},   // H,e,l,o,! (l appears twice)
	}
	
	for _, tt := range tests {
		result := countUniqueChars(tt.input)
		assert.Equal(t, tt.expected, result, "countUniqueChars('%s') should return %d", tt.input, tt.expected)
	}
}

func TestReverseString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "olleh"},
		{"abc", "cba"},
		{"", ""},
		{"a", "a"},
		{"12345", "54321"},
	}
	
	for _, tt := range tests {
		result := reverseString(tt.input)
		assert.Equal(t, tt.expected, result, "reverseString('%s') should return '%s'", tt.input, tt.expected)
	}
}
