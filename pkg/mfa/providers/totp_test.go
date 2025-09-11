package providers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dfedick/gotak/pkg/mfa"
)

func TestTOTPProvider_GenerateSecret(t *testing.T) {
	config := &mfa.TOTPConfig{
		Enabled:   true,
		Issuer:    "GoTAK Test",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Skew:      1,
	}

	provider := NewTOTPProvider(config)
	ctx := context.Background()
	userID := uuid.New()
	metadata := map[string]string{
		"account_name": "testuser",
	}

	secret, err := provider.GenerateSecret(ctx, userID, metadata)

	require.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, userID, secret.UserID)
	assert.Equal(t, mfa.MFATypeTOTP, secret.Type)
	assert.NotEmpty(t, secret.Secret)
	assert.NotEmpty(t, secret.QRCode)
	assert.NotEmpty(t, secret.URL)
	assert.Contains(t, secret.URL, "otpauth://totp/")
	assert.Contains(t, secret.URL, "GoTAK")
	assert.Contains(t, secret.URL, "testuser")
	assert.Equal(t, metadata, secret.Metadata)
}

func TestTOTPProvider_ValidateConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		config      *mfa.TOTPConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			config: &mfa.TOTPConfig{
				Enabled:   true,
				Issuer:    "GoTAK",
				Algorithm: "SHA1",
				Digits:    6,
				Period:    30,
				Skew:      1,
			},
			expectError: false,
		},
		{
			name:        "nil configuration",
			config:      nil,
			expectError: true,
			errorMsg:    "configuration cannot be nil",
		},
		{
			name: "disabled provider",
			config: &mfa.TOTPConfig{
				Enabled: false,
				Issuer:  "GoTAK",
			},
			expectError: true,
			errorMsg:    "provider is disabled",
		},
		{
			name: "empty issuer",
			config: &mfa.TOTPConfig{
				Enabled: true,
				Issuer:  "",
			},
			expectError: true,
			errorMsg:    "issuer cannot be empty",
		},
		{
			name: "invalid algorithm",
			config: &mfa.TOTPConfig{
				Enabled:   true,
				Issuer:    "GoTAK",
				Algorithm: "MD5",
			},
			expectError: true,
			errorMsg:    "invalid TOTP algorithm",
		},
		{
			name: "invalid digits",
			config: &mfa.TOTPConfig{
				Enabled: true,
				Issuer:  "GoTAK",
				Digits:  4,
			},
			expectError: true,
			errorMsg:    "digits must be 6 or 8",
		},
		{
			name: "invalid period too low",
			config: &mfa.TOTPConfig{
				Enabled: true,
				Issuer:  "GoTAK",
				Digits:  6,
				Period:  10,
			},
			expectError: true,
			errorMsg:    "period must be between 15 and 300",
		},
		{
			name: "invalid period too high",
			config: &mfa.TOTPConfig{
				Enabled: true,
				Issuer:  "GoTAK",
				Digits:  6,
				Period:  400,
			},
			expectError: true,
			errorMsg:    "period must be between 15 and 300",
		},
		{
			name: "invalid skew",
			config: &mfa.TOTPConfig{
				Enabled: true,
				Issuer:  "GoTAK",
				Digits:  6,
				Period:  30,
				Skew:    10,
			},
			expectError: true,
			errorMsg:    "skew must be between 0 and 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewTOTPProvider(tt.config)
			err := provider.ValidateConfiguration()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTOTPProvider_VerifyEnrollment(t *testing.T) {
	config := &mfa.TOTPConfig{
		Enabled:   true,
		Issuer:    "GoTAK Test",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Skew:      1,
	}

	provider := NewTOTPProvider(config)
	ctx := context.Background()
	userID := uuid.New()

	// Generate a secret for testing
	secret, err := provider.GenerateSecret(ctx, userID, nil)
	require.NoError(t, err)

	// Generate a valid TOTP code for verification
	validCode, err := provider.GetCurrentTOTP(secret.Secret)
	require.NoError(t, err)

	// Test valid verification
	err = provider.VerifyEnrollment(ctx, secret, validCode)
	assert.NoError(t, err)

	// Test invalid verification code
	err = provider.VerifyEnrollment(ctx, secret, "000000")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid TOTP verification code")

	// Test wrong secret type
	wrongSecret := &mfa.MFASecret{
		Type: mfa.MFATypeSMS,
	}
	err = provider.VerifyEnrollment(ctx, wrongSecret, validCode)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid secret type")
}

func TestTOTPProvider_CreateChallenge(t *testing.T) {
	config := &mfa.TOTPConfig{
		Enabled:   true,
		Issuer:    "GoTAK Test",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Skew:      1,
	}

	provider := NewTOTPProvider(config)
	ctx := context.Background()
	
	factor := &mfa.MFAFactor{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Type:   mfa.MFATypeTOTP,
		Status: mfa.MFAStatusActive,
		Secret: "JBSWY3DPEHPK3PXP", // Base32 encoded test secret
	}

	challenge, err := provider.CreateChallenge(ctx, factor)

	require.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.Equal(t, factor.ID, challenge.FactorID)
	assert.Equal(t, mfa.MFATypeTOTP, challenge.Type)
	assert.Equal(t, mfa.ChallengeStatusPending, challenge.Status)
	assert.Equal(t, 0, challenge.Attempts)
	assert.Equal(t, 3, challenge.MaxAttempts)
	assert.Empty(t, challenge.Challenge) // TOTP doesn't use explicit challenge data
	assert.True(t, challenge.ExpiresAt.After(time.Now()))

	// Test invalid factor type
	wrongFactor := &mfa.MFAFactor{
		Type: mfa.MFATypeSMS,
	}
	_, err = provider.CreateChallenge(ctx, wrongFactor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid factor type")
}

func TestTOTPProvider_VerifyTOTPWithFactor(t *testing.T) {
	config := &mfa.TOTPConfig{
		Enabled:   true,
		Issuer:    "GoTAK Test",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Skew:      1,
	}

	provider := NewTOTPProvider(config)
	
	// Create a test factor with a known secret
	testSecret := "JBSWY3DPEHPK3PXP" // Base32: "Hello World!"
	factor := &mfa.MFAFactor{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Type:   mfa.MFATypeTOTP,
		Status: mfa.MFAStatusActive,
		Secret: testSecret,
	}

	// Generate current valid code
	validCode, err := provider.GetCurrentTOTP(testSecret)
	require.NoError(t, err)

	// Test valid verification
	err = provider.VerifyTOTPWithFactor(factor, validCode)
	assert.NoError(t, err)

	// Test invalid code
	err = provider.VerifyTOTPWithFactor(factor, "000000")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid TOTP code")

	// Test wrong factor type
	wrongFactor := &mfa.MFAFactor{
		Type: mfa.MFATypeSMS,
	}
	err = provider.VerifyTOTPWithFactor(wrongFactor, validCode)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid factor type")
}

func TestTOTPProvider_GetTOTPAtTime(t *testing.T) {
	config := &mfa.TOTPConfig{
		Enabled:   true,
		Issuer:    "GoTAK Test",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		Skew:      1,
	}

	provider := NewTOTPProvider(config)
	testSecret := "JBSWY3DPEHPK3PXP"
	
	// Test generating code at specific time
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	code1, err := provider.GetTOTPAtTime(testSecret, testTime)
	require.NoError(t, err)
	assert.Len(t, code1, 6)
	assert.Regexp(t, "^[0-9]{6}$", code1)

	// Same time should produce same code
	code2, err := provider.GetTOTPAtTime(testSecret, testTime)
	require.NoError(t, err)
	assert.Equal(t, code1, code2)

	// Different time should produce different code
	testTime2 := testTime.Add(31 * time.Second) // Move to next time window
	code3, err := provider.GetTOTPAtTime(testSecret, testTime2)
	require.NoError(t, err)
	assert.NotEqual(t, code1, code3)
}

func TestTOTPProvider_ConfigDefaults(t *testing.T) {
	config := &mfa.TOTPConfig{
		Enabled: true,
		Issuer:  "GoTAK Test",
	}

	provider := NewTOTPProvider(config)
	err := provider.ValidateConfiguration()

	require.NoError(t, err)
	assert.Equal(t, "SHA1", config.Algorithm)
	assert.Equal(t, 6, config.Digits)
	assert.Equal(t, 30, config.Period)
	assert.Equal(t, 1, config.Skew)
}

func TestTOTPProvider_GetType(t *testing.T) {
	provider := NewTOTPProvider(&mfa.TOTPConfig{})
	assert.Equal(t, mfa.MFATypeTOTP, provider.GetType())
}

func TestTOTPProvider_generateTOTPURL(t *testing.T) {
	config := &mfa.TOTPConfig{
		Enabled:   true,
		Issuer:    "GoTAK Test",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
	}

	provider := NewTOTPProvider(config)
	secret := "JBSWY3DPEHPK3PXP"
	accountName := "testuser"

	url := provider.generateTOTPURL(accountName, secret)

	assert.Contains(t, url, "otpauth://totp/GoTAK%2520Test:testuser")
	assert.Contains(t, url, "secret="+secret)
	assert.Contains(t, url, "issuer=GoTAK+Test")
	assert.Contains(t, url, "algorithm=SHA1")
	assert.Contains(t, url, "digits=6")
	assert.Contains(t, url, "period=30")
}

func TestTOTPProvider_getAlgorithm(t *testing.T) {
	tests := []struct {
		algorithm string
		expected  string
	}{
		{"SHA1", "SHA1"},
		{"SHA256", "SHA256"}, 
		{"SHA512", "SHA512"},
		{"", "SHA1"}, // Default
		{"unknown", "SHA1"}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.algorithm, func(t *testing.T) {
			config := &mfa.TOTPConfig{
				Algorithm: tt.algorithm,
			}
			provider := NewTOTPProvider(config)
			alg := provider.getAlgorithm()
			assert.Equal(t, tt.expected, alg.String())
		})
	}
}
