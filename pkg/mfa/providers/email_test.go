package providers

import (
	"context"
	"errors"
	"net/smtp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dfedick/gotak/pkg/mfa"
)

// MockSMTPClient implements SMTPClient for testing
type MockSMTPClient struct {
	sendMailFunc func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error
	sentEmails   []SentEmail
}

type SentEmail struct {
	Addr string
	Auth smtp.Auth
	From string
	To   []string
	Msg  string
}

func (m *MockSMTPClient) SendMail(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	m.sentEmails = append(m.sentEmails, SentEmail{
		Addr: addr,
		Auth: auth,
		From: from,
		To:   to,
		Msg:  string(msg),
	})

	if m.sendMailFunc != nil {
		return m.sendMailFunc(addr, auth, from, to, msg)
	}
	return nil
}

func (m *MockSMTPClient) GetSentEmails() []SentEmail {
	return m.sentEmails
}

func (m *MockSMTPClient) Reset() {
	m.sentEmails = nil
}

func TestNewEmailProvider(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "mock",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}

	provider := NewEmailProvider(config)

	assert.NotNil(t, provider)
	assert.Equal(t, config, provider.config)
	assert.NotNil(t, provider.smtp)
	assert.Equal(t, mfa.MFATypeEmail, provider.GetType())
}

func TestEmailProvider_GenerateSecret(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "mock",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	userID := uuid.New()
	metadata := map[string]string{
		"email_address": "test@example.com",
	}

	secret, err := provider.GenerateSecret(context.Background(), userID, metadata)

	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, userID, secret.UserID)
	assert.Equal(t, mfa.MFATypeEmail, secret.Type)
	assert.Equal(t, "test@example.com", secret.Secret)
	assert.Equal(t, metadata, secret.Metadata)
}

func TestEmailProvider_GenerateSecret_InvalidEmail(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "mock",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	userID := uuid.New()

	tests := []struct {
		name     string
		metadata map[string]string
		wantErr  string
	}{
		{
			name:     "missing email address",
			metadata: map[string]string{},
			wantErr:  "email address is required for Email MFA enrollment",
		},
		{
			name:     "empty email address",
			metadata: map[string]string{"email_address": ""},
			wantErr:  "email address is required for Email MFA enrollment",
		},
		{
			name:     "invalid email format",
			metadata: map[string]string{"email_address": "invalid-email"},
			wantErr:  "invalid email address format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.GenerateSecret(context.Background(), userID, tt.metadata)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestEmailProvider_VerifyEnrollment(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "mock",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	secret := &mfa.MFASecret{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Type:   mfa.MFATypeEmail,
		Secret: "test@example.com",
		Metadata: map[string]string{
			"verification_code": "123456",
		},
	}

	// Valid verification
	err := provider.VerifyEnrollment(context.Background(), secret, "123456")
	assert.NoError(t, err)

	// Invalid verification code
	err = provider.VerifyEnrollment(context.Background(), secret, "654321")
	assert.Error(t, err)
	assert.IsType(t, &mfa.MFAError{}, err)
	mfaErr := err.(*mfa.MFAError)
	assert.Equal(t, mfa.ErrTypeVerificationFailed, mfaErr.Type)

	// Wrong secret type
	wrongTypeSecret := &mfa.MFASecret{Type: mfa.MFATypeTOTP}
	err = provider.VerifyEnrollment(context.Background(), wrongTypeSecret, "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid secret type")

	// Missing verification code
	secretWithoutCode := &mfa.MFASecret{
		Type:     mfa.MFATypeEmail,
		Metadata: map[string]string{},
	}
	err = provider.VerifyEnrollment(context.Background(), secretWithoutCode, "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no verification code found")
}

func TestEmailProvider_CreateChallenge(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "smtp", // Use SMTP so we can mock it
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	mockSMTP := &MockSMTPClient{}
	provider.SetSMTPClient(mockSMTP)

	factor := &mfa.MFAFactor{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Type:   mfa.MFATypeEmail,
		Secret: "test@example.com",
	}

	challenge, err := provider.CreateChallenge(context.Background(), factor)

	assert.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.Equal(t, factor.ID, challenge.FactorID)
	assert.Equal(t, mfa.MFATypeEmail, challenge.Type)
	assert.Equal(t, mfa.ChallengeStatusPending, challenge.Status)
	assert.Len(t, challenge.Challenge, config.CodeLength)
	assert.Equal(t, 0, challenge.Attempts)
	assert.Equal(t, 3, challenge.MaxAttempts)
	assert.True(t, time.Now().Before(challenge.ExpiresAt))

	// Verify email was sent
	sentEmails := mockSMTP.GetSentEmails()
	assert.Len(t, sentEmails, 1)
	assert.Contains(t, sentEmails[0].To, "test@example.com")
}

func TestEmailProvider_CreateChallenge_WrongType(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "mock",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	factor := &mfa.MFAFactor{
		Type: mfa.MFATypeTOTP,
	}

	_, err := provider.CreateChallenge(context.Background(), factor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid factor type")
}

func TestEmailProvider_VerifyChallenge(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "mock",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	challenge := &mfa.MFAChallenge{
		ID:        uuid.New(),
		FactorID:  uuid.New(),
		Type:      mfa.MFATypeEmail,
		Status:    mfa.ChallengeStatusPending,
		Challenge: "123456",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	// Valid verification
	err := provider.VerifyChallenge(context.Background(), challenge, "123456")
	assert.NoError(t, err)

	// Invalid code
	err = provider.VerifyChallenge(context.Background(), challenge, "654321")
	assert.Error(t, err)
	assert.IsType(t, &mfa.MFAError{}, err)
	mfaErr := err.(*mfa.MFAError)
	assert.Equal(t, mfa.ErrTypeVerificationFailed, mfaErr.Type)

	// Expired challenge
	expiredChallenge := &mfa.MFAChallenge{
		Type:      mfa.MFATypeEmail,
		Challenge: "123456",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	err = provider.VerifyChallenge(context.Background(), expiredChallenge, "123456")
	assert.Error(t, err)
	assert.IsType(t, &mfa.MFAError{}, err)
	mfaErr = err.(*mfa.MFAError)
	assert.Equal(t, mfa.ErrTypeChallengeExpired, mfaErr.Type)

	// Wrong challenge type
	wrongTypeChallenge := &mfa.MFAChallenge{Type: mfa.MFATypeTOTP}
	err = provider.VerifyChallenge(context.Background(), wrongTypeChallenge, "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid challenge type")
}

func TestEmailProvider_ValidateConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  *mfa.EmailConfig
		wantErr string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: "Email configuration cannot be nil",
		},
		{
			name: "disabled provider",
			config: &mfa.EmailConfig{
				Enabled: false,
			},
			wantErr: "Email provider is disabled",
		},
		{
			name: "empty provider type",
			config: &mfa.EmailConfig{
				Enabled:  true,
				Provider: "",
			},
			wantErr: "Email provider type cannot be empty",
		},
		{
			name: "empty template",
			config: &mfa.EmailConfig{
				Enabled:  true,
				Provider: "mock",
				Template: "",
			},
			wantErr: "Email template cannot be empty",
		},
		{
			name: "invalid code length",
			config: &mfa.EmailConfig{
				Enabled:    true,
				Provider:   "mock",
				Template:   "test",
				CodeLength: 10,
			},
			wantErr: "Email code length must be between 4 and 8",
		},
		{
			name: "invalid code TTL",
			config: &mfa.EmailConfig{
				Enabled:    true,
				Provider:   "mock",
				Template:   "test",
				CodeLength: 6,
				CodeTTL:    time.Hour,
			},
			wantErr: "Email code TTL must be between 1 and 30 minutes",
		},
		{
			name: "valid config",
			config: &mfa.EmailConfig{
				Enabled:    true,
				Provider:   "mock",
				Template:   "Your code: {{.Code}}",
				CodeLength: 6,
				CodeTTL:    5 * time.Minute,
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewEmailProvider(tt.config)
			err := provider.ValidateConfiguration()

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestEmailProvider_generateVerificationCode(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "mock",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	// Test multiple code generations
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := provider.generateVerificationCode()
		assert.NoError(t, err)
		assert.Len(t, code, config.CodeLength)
		assert.Regexp(t, `^\d{6}$`, code) // Should be 6 digits
		codes[code] = true
	}

	// Should generate different codes (at least some variation)
	assert.Greater(t, len(codes), 80) // Allow some duplicates but expect variety
}

func TestEmailProvider_SendEnrollmentEmail(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "smtp",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	mockSMTP := &MockSMTPClient{}
	provider.SetSMTPClient(mockSMTP)

	code, err := provider.SendEnrollmentEmail(context.Background(), "test@example.com")

	assert.NoError(t, err)
	assert.Len(t, code, config.CodeLength)

	// Verify email was sent
	sentEmails := mockSMTP.GetSentEmails()
	assert.Len(t, sentEmails, 1)
	assert.Contains(t, sentEmails[0].To, "test@example.com")
}

func TestEmailProvider_SMTPError(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "smtp",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	// Mock SMTP client that returns error
	mockSMTP := &MockSMTPClient{
		sendMailFunc: func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
			return errors.New("SMTP connection failed")
		},
	}
	provider.SetSMTPClient(mockSMTP)

	factor := &mfa.MFAFactor{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Type:   mfa.MFATypeEmail,
		Secret: "test@example.com",
	}

	_, err := provider.CreateChallenge(context.Background(), factor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send Email verification code")
}

func TestEmailProvider_VerifyEmailWithCode(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "mock",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	// Valid code
	err := provider.VerifyEmailWithCode("123456", "123456")
	assert.NoError(t, err)

	// Invalid code
	err = provider.VerifyEmailWithCode("123456", "654321")
	assert.Error(t, err)
	assert.IsType(t, &mfa.MFAError{}, err)
	mfaErr := err.(*mfa.MFAError)
	assert.Equal(t, mfa.ErrTypeVerificationFailed, mfaErr.Type)
}

func TestEmailProvider_renderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		code     string
		wantSub  string
		wantBody string
		wantErr  bool
	}{
		{
			name:     "simple template",
			template: "Subject: Verification Code\n\nYour code: {{.Code}}",
			code:     "123456",
			wantSub:  "Verification Code",
			wantBody: "Your code: 123456",
			wantErr:  false,
		},
		{
			name:     "template without subject",
			template: "Your verification code is: {{.Code}}",
			code:     "999888",
			wantSub:  "GoTAK Verification Code",
			wantBody: "Your verification code is: 999888",
			wantErr:  false,
		},
		{
			name:     "invalid template",
			template: "Your code: {{.InvalidField}}",
			code:     "123456",
			wantSub:  "",
			wantBody: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &mfa.EmailConfig{
				Template: tt.template,
			}
			provider := NewEmailProvider(config)

			subject, body, err := provider.renderTemplate(tt.code)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSub, subject)
				assert.Equal(t, tt.wantBody, body)
			}
		})
	}
}

func TestEmailProvider_isValidEmailAddress(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "mock",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user+tag@domain.co.uk", true},
		{"invalid-email", false},
		{"@domain.com", false},
		{"test@", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := provider.isValidEmailAddress(tt.email)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestEmailProvider_MockProvider(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:    true,
		Provider:   "mock",
		Template:   "Your code: {{.Code}}",
		CodeLength: 6,
		CodeTTL:    5 * time.Minute,
	}
	provider := NewEmailProvider(config)

	factor := &mfa.MFAFactor{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Type:   mfa.MFATypeEmail,
		Secret: "test@example.com",
	}

	// Mock provider should work without SMTP client
	challenge, err := provider.CreateChallenge(context.Background(), factor)
	assert.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.Equal(t, factor.ID, challenge.FactorID)
	assert.Equal(t, mfa.MFATypeEmail, challenge.Type)
	assert.Len(t, challenge.Challenge, config.CodeLength)
}

func TestEmailProvider_DefaultConfiguration(t *testing.T) {
	config := &mfa.EmailConfig{
		Enabled:  true,
		Provider: "mock",
		Template: "Your code: {{.Code}}",
		// Leave CodeLength and CodeTTL as 0 to test defaults
	}
	provider := NewEmailProvider(config)

	err := provider.ValidateConfiguration()
	assert.NoError(t, err)
	assert.Equal(t, 6, config.CodeLength)    // Should be set to default
	assert.Equal(t, 5*time.Minute, config.CodeTTL) // Should be set to default
}
