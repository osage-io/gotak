package providers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dfedick/gotak/pkg/mfa"
)

// MockWebAuthnClient implements WebAuthnClient for testing
type MockWebAuthnClient struct {
	beginRegistrationFunc  func(ctx context.Context, user *WebAuthnUser, options *RegistrationOptions) (*CredentialCreation, error)
	finishRegistrationFunc func(ctx context.Context, user *WebAuthnUser, sessionData *SessionData, response *CredentialCreationResponse) (*Credential, error)
	beginLoginFunc         func(ctx context.Context, user *WebAuthnUser, options *LoginOptions) (*CredentialAssertion, error)
	finishLoginFunc        func(ctx context.Context, user *WebAuthnUser, sessionData *SessionData, response *CredentialAssertionResponse) (*Credential, error)
}

func (m *MockWebAuthnClient) BeginRegistration(ctx context.Context, user *WebAuthnUser, options *RegistrationOptions) (*CredentialCreation, error) {
	if m.beginRegistrationFunc != nil {
		return m.beginRegistrationFunc(ctx, user, options)
	}
	return &CredentialCreation{PublicKey: *options}, nil
}

func (m *MockWebAuthnClient) FinishRegistration(ctx context.Context, user *WebAuthnUser, sessionData *SessionData, response *CredentialCreationResponse) (*Credential, error) {
	if m.finishRegistrationFunc != nil {
		return m.finishRegistrationFunc(ctx, user, sessionData, response)
	}
	return &Credential{
		ID:              response.RawID,
		PublicKey:       []byte("mock-public-key"),
		AttestationType: "none",
		Authenticator:   Authenticator{AAGUID: make([]byte, 16), SignCount: 1},
	}, nil
}

func (m *MockWebAuthnClient) BeginLogin(ctx context.Context, user *WebAuthnUser, options *LoginOptions) (*CredentialAssertion, error) {
	if m.beginLoginFunc != nil {
		return m.beginLoginFunc(ctx, user, options)
	}
	return &CredentialAssertion{PublicKey: *options}, nil
}

func (m *MockWebAuthnClient) FinishLogin(ctx context.Context, user *WebAuthnUser, sessionData *SessionData, response *CredentialAssertionResponse) (*Credential, error) {
	if m.finishLoginFunc != nil {
		return m.finishLoginFunc(ctx, user, sessionData, response)
	}
	return &Credential{
		ID:        response.RawID,
		PublicKey: []byte("mock-public-key"),
	}, nil
}

func TestNewWebAuthnProvider(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled:            true,
		RPID:               "example.com",
		RPName:             "Test App",
		Origin:             "https://example.com",
		Timeout:            60 * time.Second,
		RequireResidentKey: false,
		UserVerification:   "preferred",
		Attestation:        "none",
	}

	provider := NewWebAuthnProvider(config)

	assert.NotNil(t, provider)
	assert.Equal(t, config, provider.config)
	assert.NotNil(t, provider.client)
	assert.Equal(t, mfa.MFATypeWebAuthn, provider.GetType())
}

func TestWebAuthnProvider_GenerateSecret(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled:            true,
		RPID:               "example.com",
		RPName:             "Test App",
		Origin:             "https://example.com",
		Timeout:            60 * time.Second,
		RequireResidentKey: false,
		UserVerification:   "preferred",
		Attestation:        "none",
	}
	provider := NewWebAuthnProvider(config)

	userID := uuid.New()
	metadata := map[string]string{
		"username":     "testuser",
		"display_name": "Test User",
	}

	secret, err := provider.GenerateSecret(context.Background(), userID, metadata)

	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, userID, secret.UserID)
	assert.Equal(t, mfa.MFATypeWebAuthn, secret.Type)
	assert.NotEmpty(t, secret.Secret)

	// Verify metadata contains required fields
	assert.Contains(t, secret.Metadata, "registration_options")
	assert.Contains(t, secret.Metadata, "session_data")
	assert.Contains(t, secret.Metadata, "challenge")

	// Verify registration options are valid JSON
	var regOptions RegistrationOptions
	err = json.Unmarshal([]byte(secret.Metadata["registration_options"]), &regOptions)
	assert.NoError(t, err)
	assert.Equal(t, config.RPID, regOptions.RP.ID)
	assert.Equal(t, config.RPName, regOptions.RP.Name)
	assert.Equal(t, "testuser", regOptions.User.Name)
}

func TestWebAuthnProvider_GenerateSecret_InvalidMetadata(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		RPID:    "example.com",
		RPName:  "Test App",
		Timeout: 60 * time.Second,
	}
	provider := NewWebAuthnProvider(config)
	userID := uuid.New()

	tests := []struct {
		name     string
		metadata map[string]string
		wantErr  string
	}{
		{
			name:     "missing username",
			metadata: map[string]string{},
			wantErr:  "username is required for WebAuthn MFA enrollment",
		},
		{
			name:     "empty username",
			metadata: map[string]string{"username": ""},
			wantErr:  "username is required for WebAuthn MFA enrollment",
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

func TestWebAuthnProvider_VerifyEnrollment(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		RPID:    "example.com",
		RPName:  "Test App",
		Timeout: 60 * time.Second,
	}
	provider := NewWebAuthnProvider(config)
	mockClient := &MockWebAuthnClient{}
	provider.SetWebAuthnClient(mockClient)

	// Create a secret with session data
	testUUID := uuid.New()
	sessionData := SessionData{
		Challenge: []byte("test-challenge"),
		UserID:    testUUID[:],
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(60 * time.Second),
	}
	sessionJSON, _ := json.Marshal(sessionData)

	secret := &mfa.MFASecret{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Type:   mfa.MFATypeWebAuthn,
		Metadata: map[string]string{
			"session_data":   string(sessionJSON),
			"username":       "testuser",
			"display_name":   "Test User",
		},
	}

	// Create valid credential creation response
	credResponse := CredentialCreationResponse{
		ID:    "test-credential-id",
		RawID: []byte("test-credential-raw-id"),
		Response: AuthenticatorAttestationResponse{
			ClientDataJSON:    []byte("{}"),
			AttestationObject: []byte("mock-attestation"),
		},
		Type: "public-key",
	}
	credJSON, _ := json.Marshal(credResponse)

	// Test successful verification
	err := provider.VerifyEnrollment(context.Background(), secret, string(credJSON))
	assert.NoError(t, err)

	// Verify credential info was stored
	assert.Contains(t, secret.Metadata, "credential_id")
	assert.Contains(t, secret.Metadata, "public_key")
	assert.Contains(t, secret.Metadata, "attestation_type")
}

func TestWebAuthnProvider_VerifyEnrollment_ExpiredSession(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		RPID:    "example.com",
		RPName:  "Test App",
		Timeout: 60 * time.Second,
	}
	provider := NewWebAuthnProvider(config)

	// Create expired session data
	testUUID := uuid.New()
	sessionData := SessionData{
		Challenge: []byte("test-challenge"),
		UserID:    testUUID[:],
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}
	sessionJSON, _ := json.Marshal(sessionData)

	secret := &mfa.MFASecret{
		Type: mfa.MFATypeWebAuthn,
		Metadata: map[string]string{
			"session_data": string(sessionJSON),
			"username":     "testuser",
		},
	}

	err := provider.VerifyEnrollment(context.Background(), secret, "{}")
	assert.Error(t, err)
	assert.IsType(t, &mfa.MFAError{}, err)
	mfaErr := err.(*mfa.MFAError)
	assert.Equal(t, mfa.ErrTypeChallengeExpired, mfaErr.Type)
}

func TestWebAuthnProvider_CreateChallenge(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		RPID:    "example.com",
		RPName:  "Test App",
		Timeout: 60 * time.Second,
	}
	provider := NewWebAuthnProvider(config)

	factor := &mfa.MFAFactor{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Type:   mfa.MFATypeWebAuthn,
		Metadata: map[string]string{
			"credential_id": "dGVzdC1jcmVkZW50aWFs", // base64 encoded "test-credential"
		},
	}

	challenge, err := provider.CreateChallenge(context.Background(), factor)

	assert.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.Equal(t, factor.ID, challenge.FactorID)
	assert.Equal(t, mfa.MFATypeWebAuthn, challenge.Type)
	assert.Equal(t, mfa.ChallengeStatusPending, challenge.Status)
	assert.NotEmpty(t, challenge.Challenge)
	assert.Equal(t, 0, challenge.Attempts)
	assert.Equal(t, 3, challenge.MaxAttempts)
	assert.True(t, time.Now().Before(challenge.ExpiresAt))

	// Verify challenge contains login options
	var loginOptions LoginOptions
	err = json.Unmarshal([]byte(challenge.Challenge), &loginOptions)
	assert.NoError(t, err)
	assert.Equal(t, config.RPID, loginOptions.RPID)
	assert.Len(t, loginOptions.AllowCredentials, 1)

	// Verify metadata contains session data
	assert.Contains(t, challenge.Metadata, "session_data")
	assert.Contains(t, challenge.Metadata, "challenge")
}

func TestWebAuthnProvider_CreateChallenge_MissingCredential(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		RPID:    "example.com",
		RPName:  "Test App",
		Timeout: 60 * time.Second,
	}
	provider := NewWebAuthnProvider(config)

	factor := &mfa.MFAFactor{
		Type: mfa.MFATypeWebAuthn,
		Metadata: map[string]string{
			// Missing credential_id
		},
	}

	_, err := provider.CreateChallenge(context.Background(), factor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no credential ID found")
}

func TestWebAuthnProvider_VerifyChallenge(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		RPID:    "example.com",
		RPName:  "Test App",
		Timeout: 60 * time.Second,
	}
	provider := NewWebAuthnProvider(config)
	mockClient := &MockWebAuthnClient{}
	provider.SetWebAuthnClient(mockClient)

	// Create session data
	testUUID := uuid.New()
	sessionData := SessionData{
		Challenge: []byte("test-challenge"),
		UserID:    testUUID[:],
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(60 * time.Second),
	}
	sessionJSON, _ := json.Marshal(sessionData)

	challenge := &mfa.MFAChallenge{
		ID:        uuid.New(),
		FactorID:  uuid.New(),
		Type:      mfa.MFATypeWebAuthn,
		Status:    mfa.ChallengeStatusPending,
		Challenge: "{}",
		ExpiresAt: time.Now().Add(60 * time.Second),
		Metadata: map[string]string{
			"session_data": string(sessionJSON),
		},
	}

	// Create valid assertion response
	assertionResponse := CredentialAssertionResponse{
		ID:    "test-credential-id",
		RawID: []byte("test-credential-raw-id"),
		Response: AuthenticatorAssertionResponse{
			ClientDataJSON:    []byte("{}"),
			AuthenticatorData: []byte("mock-auth-data"),
			Signature:         []byte("mock-signature"),
		},
		Type: "public-key",
	}
	responseJSON, _ := json.Marshal(assertionResponse)

	// Test successful verification
	err := provider.VerifyChallenge(context.Background(), challenge, string(responseJSON))
	assert.NoError(t, err)
}

func TestWebAuthnProvider_VerifyChallenge_ExpiredChallenge(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		Timeout: 60 * time.Second,
	}
	provider := NewWebAuthnProvider(config)

	challenge := &mfa.MFAChallenge{
		Type:      mfa.MFATypeWebAuthn,
		ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired
	}

	err := provider.VerifyChallenge(context.Background(), challenge, "{}")
	assert.Error(t, err)
	assert.IsType(t, &mfa.MFAError{}, err)
	mfaErr := err.(*mfa.MFAError)
	assert.Equal(t, mfa.ErrTypeChallengeExpired, mfaErr.Type)
}

func TestWebAuthnProvider_VerifyChallenge_InvalidResponse(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		Timeout: 60 * time.Second,
	}
	provider := NewWebAuthnProvider(config)

	challenge := &mfa.MFAChallenge{
		Type:      mfa.MFATypeWebAuthn,
		ExpiresAt: time.Now().Add(60 * time.Second),
		Metadata:  map[string]string{},
	}

	err := provider.VerifyChallenge(context.Background(), challenge, "invalid-json")
	assert.Error(t, err)
	assert.IsType(t, &mfa.MFAError{}, err)
	mfaErr := err.(*mfa.MFAError)
	assert.Equal(t, mfa.ErrTypeVerificationFailed, mfaErr.Type)
}

func TestWebAuthnProvider_VerifyChallenge_ClientError(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		Timeout: 60 * time.Second,
	}
	provider := NewWebAuthnProvider(config)

	// Mock client that returns error
	mockClient := &MockWebAuthnClient{
		finishLoginFunc: func(ctx context.Context, user *WebAuthnUser, sessionData *SessionData, response *CredentialAssertionResponse) (*Credential, error) {
			return nil, errors.New("authentication failed")
		},
	}
	provider.SetWebAuthnClient(mockClient)

	testUUID := uuid.New()
	sessionData := SessionData{
		Challenge: []byte("test-challenge"),
		UserID:    testUUID[:],
		ExpiresAt: time.Now().Add(60 * time.Second),
	}
	sessionJSON, _ := json.Marshal(sessionData)

	challenge := &mfa.MFAChallenge{
		Type:      mfa.MFATypeWebAuthn,
		ExpiresAt: time.Now().Add(60 * time.Second),
		Metadata: map[string]string{
			"session_data": string(sessionJSON),
		},
	}

	assertionResponse := CredentialAssertionResponse{
		ID:       "test-id",
		RawID:    []byte("test-raw-id"),
		Response: AuthenticatorAssertionResponse{},
		Type:     "public-key",
	}
	responseJSON, _ := json.Marshal(assertionResponse)

	err := provider.VerifyChallenge(context.Background(), challenge, string(responseJSON))
	assert.Error(t, err)
	assert.IsType(t, &mfa.MFAError{}, err)
	mfaErr := err.(*mfa.MFAError)
	assert.Equal(t, mfa.ErrTypeVerificationFailed, mfaErr.Type)
}

func TestWebAuthnProvider_ValidateConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  *mfa.WebAuthnConfig
		wantErr string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: "WebAuthn configuration cannot be nil",
		},
		{
			name: "disabled provider",
			config: &mfa.WebAuthnConfig{
				Enabled: false,
			},
			wantErr: "WebAuthn provider is disabled",
		},
		{
			name: "empty RPID",
			config: &mfa.WebAuthnConfig{
				Enabled: true,
				RPID:    "",
			},
			wantErr: "WebAuthn RPID cannot be empty",
		},
		{
			name: "empty RP name",
			config: &mfa.WebAuthnConfig{
				Enabled: true,
				RPID:    "example.com",
				RPName:  "",
			},
			wantErr: "WebAuthn RP name cannot be empty",
		},
		{
			name: "invalid timeout - too short",
			config: &mfa.WebAuthnConfig{
				Enabled: true,
				RPID:    "example.com",
				RPName:  "Test App",
				Timeout: 15 * time.Second, // Too short
			},
			wantErr: "WebAuthn timeout must be between 30 seconds and 5 minutes",
		},
		{
			name: "invalid timeout - too long",
			config: &mfa.WebAuthnConfig{
				Enabled: true,
				RPID:    "example.com",
				RPName:  "Test App",
				Timeout: 10 * time.Minute, // Too long
			},
			wantErr: "WebAuthn timeout must be between 30 seconds and 5 minutes",
		},
		{
			name: "valid config",
			config: &mfa.WebAuthnConfig{
				Enabled: true,
				RPID:    "example.com",
				RPName:  "Test App",
				Timeout: 60 * time.Second,
			},
			wantErr: "",
		},
		{
			name: "valid config with default timeout",
			config: &mfa.WebAuthnConfig{
				Enabled: true,
				RPID:    "example.com",
				RPName:  "Test App",
				// Timeout is 0, should be set to default
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewWebAuthnProvider(tt.config)
			err := provider.ValidateConfiguration()

			if tt.wantErr == "" {
				assert.NoError(t, err)
				if tt.config != nil && tt.config.Timeout == 0 {
					assert.Equal(t, 60*time.Second, tt.config.Timeout) // Should be set to default
				}
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestWebAuthnProvider_GetType(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		RPID:    "example.com",
		RPName:  "Test App",
	}
	provider := NewWebAuthnProvider(config)

	assert.Equal(t, mfa.MFATypeWebAuthn, provider.GetType())
}

func TestWebAuthnProvider_GetRegistrationOptions(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		RPID:    "example.com",
		RPName:  "Test App",
	}
	provider := NewWebAuthnProvider(config)

	secret := &mfa.MFASecret{
		Type: mfa.MFATypeWebAuthn,
		Metadata: map[string]string{
			"registration_options": `{"challenge":"test"}`,
		},
	}

	options, err := provider.GetRegistrationOptions(context.Background(), secret)
	assert.NoError(t, err)
	assert.Equal(t, `{"challenge":"test"}`, options)
}

func TestWebAuthnProvider_GetRegistrationOptions_InvalidType(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		RPID:    "example.com",
		RPName:  "Test App",
	}
	provider := NewWebAuthnProvider(config)

	secret := &mfa.MFASecret{
		Type: mfa.MFATypeTOTP, // Wrong type
	}

	_, err := provider.GetRegistrationOptions(context.Background(), secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid secret type")
}

func TestWebAuthnProvider_GetLoginOptions(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		RPID:    "example.com",
		RPName:  "Test App",
	}
	provider := NewWebAuthnProvider(config)

	challenge := &mfa.MFAChallenge{
		Type:      mfa.MFATypeWebAuthn,
		Challenge: `{"allowCredentials":[]}`,
	}

	options, err := provider.GetLoginOptions(context.Background(), challenge)
	assert.NoError(t, err)
	assert.Equal(t, `{"allowCredentials":[]}`, options)
}

func TestWebAuthnProvider_GetLoginOptions_InvalidType(t *testing.T) {
	config := &mfa.WebAuthnConfig{
		Enabled: true,
		RPID:    "example.com",
		RPName:  "Test App",
	}
	provider := NewWebAuthnProvider(config)

	challenge := &mfa.MFAChallenge{
		Type: mfa.MFATypeTOTP, // Wrong type
	}

	_, err := provider.GetLoginOptions(context.Background(), challenge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid challenge type")
}
