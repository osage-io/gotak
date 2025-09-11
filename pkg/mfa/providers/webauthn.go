package providers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dfedick/gotak/pkg/mfa"
)

// WebAuthnProvider implements the MFAProvider interface for WebAuthn/FIDO2 authentication
type WebAuthnProvider struct {
	config *mfa.WebAuthnConfig
	client WebAuthnClient
}

// WebAuthnClient interface allows for testing with mock WebAuthn clients
type WebAuthnClient interface {
	BeginRegistration(ctx context.Context, user *WebAuthnUser, options *RegistrationOptions) (*CredentialCreation, error)
	FinishRegistration(ctx context.Context, user *WebAuthnUser, sessionData *SessionData, response *CredentialCreationResponse) (*Credential, error)
	BeginLogin(ctx context.Context, user *WebAuthnUser, options *LoginOptions) (*CredentialAssertion, error)
	FinishLogin(ctx context.Context, user *WebAuthnUser, sessionData *SessionData, response *CredentialAssertionResponse) (*Credential, error)
}

// WebAuthnUser represents a user for WebAuthn operations
type WebAuthnUser struct {
	ID          []byte `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// RegistrationOptions contains options for credential registration
type RegistrationOptions struct {
	Challenge              []byte                      `json:"challenge"`
	RP                     RelyingParty               `json:"rp"`
	User                   WebAuthnUser               `json:"user"`
	PubKeyCredParams       []PubKeyCredParam          `json:"pubKeyCredParams"`
	AuthenticatorSelection *AuthenticatorSelection    `json:"authenticatorSelection,omitempty"`
	Timeout                int                        `json:"timeout"`
	Attestation            string                     `json:"attestation"`
	Extensions             map[string]interface{}     `json:"extensions,omitempty"`
}

// LoginOptions contains options for credential assertion
type LoginOptions struct {
	Challenge        []byte                     `json:"challenge"`
	Timeout          int                        `json:"timeout"`
	RPID             string                     `json:"rpId"`
	AllowCredentials []PublicKeyCredDescriptor  `json:"allowCredentials"`
	UserVerification string                     `json:"userVerification"`
	Extensions       map[string]interface{}     `json:"extensions,omitempty"`
}

// CredentialCreation represents the credential creation options
type CredentialCreation struct {
	PublicKey RegistrationOptions `json:"publicKey"`
}

// CredentialAssertion represents the credential assertion options
type CredentialAssertion struct {
	PublicKey LoginOptions `json:"publicKey"`
}

// CredentialCreationResponse represents the client's response to credential creation
type CredentialCreationResponse struct {
	ID       string                        `json:"id"`
	RawID    []byte                       `json:"rawId"`
	Response AuthenticatorAttestationResponse `json:"response"`
	Type     string                       `json:"type"`
}

// CredentialAssertionResponse represents the client's response to credential assertion
type CredentialAssertionResponse struct {
	ID       string                       `json:"id"`
	RawID    []byte                      `json:"rawId"`
	Response AuthenticatorAssertionResponse `json:"response"`
	Type     string                      `json:"type"`
}

// AuthenticatorAttestationResponse contains the authenticator's attestation response
type AuthenticatorAttestationResponse struct {
	ClientDataJSON    []byte `json:"clientDataJSON"`
	AttestationObject []byte `json:"attestationObject"`
}

// AuthenticatorAssertionResponse contains the authenticator's assertion response
type AuthenticatorAssertionResponse struct {
	ClientDataJSON    []byte `json:"clientDataJSON"`
	AuthenticatorData []byte `json:"authenticatorData"`
	Signature         []byte `json:"signature"`
	UserHandle        []byte `json:"userHandle,omitempty"`
}

// Supporting data structures

// RelyingParty represents the WebAuthn Relying Party
type RelyingParty struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PubKeyCredParam represents supported public key algorithms
type PubKeyCredParam struct {
	Type string `json:"type"`
	Alg  int    `json:"alg"`
}

// AuthenticatorSelection specifies requirements for authenticator selection
type AuthenticatorSelection struct {
	AuthenticatorAttachment string `json:"authenticatorAttachment,omitempty"`
	RequireResidentKey      bool   `json:"requireResidentKey"`
	UserVerification        string `json:"userVerification"`
}

// PublicKeyCredDescriptor describes a credential
type PublicKeyCredDescriptor struct {
	Type       string   `json:"type"`
	ID         []byte   `json:"id"`
	Transports []string `json:"transports,omitempty"`
}

// SessionData stores session data for WebAuthn operations
type SessionData struct {
	Challenge   []byte    `json:"challenge"`
	UserID      []byte    `json:"userId"`
	Credentials [][]byte  `json:"credentials,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// Credential represents a WebAuthn credential
type Credential struct {
	ID              []byte    `json:"id"`
	PublicKey       []byte    `json:"publicKey"`
	AttestationType string    `json:"attestationType"`
	Authenticator   Authenticator `json:"authenticator"`
	Flags           UserVerificationFlags `json:"flags"`
}

// Authenticator represents authenticator information
type Authenticator struct {
	AAGUID       []byte `json:"aaguid"`
	SignCount    uint32 `json:"signCount"`
	CloneWarning bool   `json:"cloneWarning"`
}

// UserVerificationFlags represents user verification flags
type UserVerificationFlags struct {
	UserPresent    bool `json:"userPresent"`
	UserVerified   bool `json:"userVerified"`
	BackupEligible bool `json:"backupEligible"`
	BackupState    bool `json:"backupState"`
}

// DefaultWebAuthnClient provides a basic WebAuthn client implementation
type DefaultWebAuthnClient struct {
	config *mfa.WebAuthnConfig
}

// NewWebAuthnProvider creates a new WebAuthn provider with the given configuration
func NewWebAuthnProvider(config *mfa.WebAuthnConfig) *WebAuthnProvider {
	return &WebAuthnProvider{
		config: config,
		client: &DefaultWebAuthnClient{config: config},
	}
}

// GetType returns the MFA type this provider handles
func (p *WebAuthnProvider) GetType() mfa.MFAType {
	return mfa.MFATypeWebAuthn
}

// GenerateSecret creates a new WebAuthn MFA secret for enrollment
func (p *WebAuthnProvider) GenerateSecret(ctx context.Context, userID uuid.UUID, metadata map[string]string) (*mfa.MFASecret, error) {
	// For WebAuthn, the "secret" contains the user information and credential options
	username, exists := metadata["username"]
	if !exists || username == "" {
		return nil, fmt.Errorf("username is required for WebAuthn MFA enrollment")
	}

	displayName, exists := metadata["display_name"]
	if !exists || displayName == "" {
		displayName = username
	}

	// Create WebAuthn user
	webAuthnUser := &WebAuthnUser{
		ID:          userID[:], // Convert UUID to bytes
		Name:        username,
		DisplayName: displayName,
	}

	// Generate registration options
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}

	registrationOptions := &RegistrationOptions{
		Challenge: challenge,
		RP: RelyingParty{
			ID:   p.config.RPID,
			Name: p.config.RPName,
		},
		User: *webAuthnUser,
		PubKeyCredParams: []PubKeyCredParam{
			{Type: "public-key", Alg: -7},  // ES256
			{Type: "public-key", Alg: -257}, // RS256
		},
		AuthenticatorSelection: &AuthenticatorSelection{
			RequireResidentKey: false,
			UserVerification:   "preferred",
		},
		Timeout:     int(p.config.Timeout.Milliseconds()),
		Attestation: "none",
	}

	// Serialize the registration options
	optionsJSON, err := json.Marshal(registrationOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize registration options: %w", err)
	}

	// Create session data
	sessionData := &SessionData{
		Challenge: challenge,
		UserID:    webAuthnUser.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(p.config.Timeout),
	}

	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize session data: %w", err)
	}

	// Store both registration options and session data in metadata
	enrichedMetadata := make(map[string]string)
	for k, v := range metadata {
		enrichedMetadata[k] = v
	}
	enrichedMetadata["registration_options"] = string(optionsJSON)
	enrichedMetadata["session_data"] = string(sessionJSON)
	enrichedMetadata["challenge"] = base64.URLEncoding.EncodeToString(challenge)

	return &mfa.MFASecret{
		ID:       uuid.New(),
		UserID:   userID,
		Type:     mfa.MFATypeWebAuthn,
		Secret:   base64.URLEncoding.EncodeToString(challenge), // Store challenge as secret
		Metadata: enrichedMetadata,
	}, nil
}

// VerifyEnrollment validates WebAuthn enrollment by processing the credential creation response
func (p *WebAuthnProvider) VerifyEnrollment(ctx context.Context, secret *mfa.MFASecret, verificationData string) error {
	if secret.Type != mfa.MFATypeWebAuthn {
		return fmt.Errorf("invalid secret type for WebAuthn provider: %s", secret.Type)
	}

	// Parse the credential creation response
	var credResponse CredentialCreationResponse
	if err := json.Unmarshal([]byte(verificationData), &credResponse); err != nil {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "failed to parse credential creation response", err)
	}

	// Get session data from metadata
	sessionDataJSON, exists := secret.Metadata["session_data"]
	if !exists {
		return fmt.Errorf("no session data found for WebAuthn enrollment")
	}

	var sessionData SessionData
	if err := json.Unmarshal([]byte(sessionDataJSON), &sessionData); err != nil {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "failed to parse session data", err)
	}

	// Check session expiration
	if time.Now().After(sessionData.ExpiresAt) {
		return mfa.NewMFAError(mfa.ErrTypeChallengeExpired, "WebAuthn enrollment session has expired", nil)
	}

	// Create WebAuthn user from metadata
	webAuthnUser := &WebAuthnUser{
		ID:          sessionData.UserID,
		Name:        secret.Metadata["username"],
		DisplayName: secret.Metadata["display_name"],
	}

	// Finish registration using the WebAuthn client
	credential, err := p.client.FinishRegistration(ctx, webAuthnUser, &sessionData, &credResponse)
	if err != nil {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "WebAuthn registration verification failed", err)
	}

	// Store the credential information in metadata for later use
	secret.Metadata["credential_id"] = base64.URLEncoding.EncodeToString(credential.ID)
	secret.Metadata["public_key"] = base64.URLEncoding.EncodeToString(credential.PublicKey)
	secret.Metadata["attestation_type"] = credential.AttestationType

	return nil
}

// CreateChallenge generates a new WebAuthn authentication challenge
func (p *WebAuthnProvider) CreateChallenge(ctx context.Context, factor *mfa.MFAFactor) (*mfa.MFAChallenge, error) {
	if factor.Type != mfa.MFATypeWebAuthn {
		return nil, fmt.Errorf("invalid factor type for WebAuthn provider: %s", factor.Type)
	}

	// Parse stored credential information
	credentialIDStr, exists := factor.Metadata["credential_id"]
	if !exists {
		return nil, fmt.Errorf("no credential ID found for WebAuthn factor")
	}

	credentialID, err := base64.URLEncoding.DecodeString(credentialIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode credential ID: %w", err)
	}

	// Generate authentication challenge
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}

	// Create login options
	loginOptions := &LoginOptions{
		Challenge: challenge,
		Timeout:   int(p.config.Timeout.Milliseconds()),
		RPID:      p.config.RPID,
		AllowCredentials: []PublicKeyCredDescriptor{
			{
				Type: "public-key",
				ID:   credentialID,
				Transports: []string{"usb", "nfc", "ble", "internal"},
			},
		},
		UserVerification: "preferred",
	}

	// Serialize login options
	optionsJSON, err := json.Marshal(loginOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize login options: %w", err)
	}

	// Create session data
	sessionData := &SessionData{
		Challenge: challenge,
		UserID:    factor.UserID[:],
		Credentials: [][]byte{credentialID},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(p.config.Timeout),
	}

	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize session data: %w", err)
	}

	// Create MFA challenge
	mfaChallenge := &mfa.MFAChallenge{
		ID:          uuid.New(),
		FactorID:    factor.ID,
		Type:        mfa.MFATypeWebAuthn,
		Status:      mfa.ChallengeStatusPending,
		Challenge:   string(optionsJSON), // Store login options as challenge
		Attempts:    0,
		MaxAttempts: 3,
		ExpiresAt:   time.Now().Add(p.config.Timeout),
		CreatedAt:   time.Now(),
		Metadata: map[string]string{
			"session_data": string(sessionJSON),
			"challenge":    base64.URLEncoding.EncodeToString(challenge),
		},
	}

	return mfaChallenge, nil
}

// VerifyChallenge validates a user's response to a WebAuthn authentication challenge
func (p *WebAuthnProvider) VerifyChallenge(ctx context.Context, challenge *mfa.MFAChallenge, response string) error {
	if challenge.Type != mfa.MFATypeWebAuthn {
		return fmt.Errorf("invalid challenge type for WebAuthn provider: %s", challenge.Type)
	}

	// Check if challenge has expired
	if time.Now().After(challenge.ExpiresAt) {
		return mfa.NewMFAError(mfa.ErrTypeChallengeExpired, "WebAuthn challenge has expired", nil)
	}

	// Parse the credential assertion response
	var assertionResponse CredentialAssertionResponse
	if err := json.Unmarshal([]byte(response), &assertionResponse); err != nil {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "failed to parse credential assertion response", err)
	}

	// Get session data from challenge metadata
	sessionDataJSON, exists := challenge.Metadata["session_data"]
	if !exists {
		return fmt.Errorf("no session data found for WebAuthn challenge")
	}

	var sessionData SessionData
	if err := json.Unmarshal([]byte(sessionDataJSON), &sessionData); err != nil {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "failed to parse session data", err)
	}

	// Create WebAuthn user (we need to get user info from somewhere)
	// In a real implementation, you'd get this from the user database
	webAuthnUser := &WebAuthnUser{
		ID:          sessionData.UserID,
		Name:        "user", // This should come from user lookup
		DisplayName: "User", // This should come from user lookup
	}

	// Finish login using the WebAuthn client
	_, err := p.client.FinishLogin(ctx, webAuthnUser, &sessionData, &assertionResponse)
	if err != nil {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "WebAuthn authentication verification failed", err)
	}

	return nil
}

// ValidateConfiguration checks if the provider configuration is valid
func (p *WebAuthnProvider) ValidateConfiguration() error {
	if p.config == nil {
		return fmt.Errorf("WebAuthn configuration cannot be nil")
	}

	if !p.config.Enabled {
		return fmt.Errorf("WebAuthn provider is disabled")
	}

	if p.config.RPID == "" {
		return fmt.Errorf("WebAuthn RPID cannot be empty")
	}

	if p.config.RPName == "" {
		return fmt.Errorf("WebAuthn RP name cannot be empty")
	}

	if p.config.Timeout == 0 {
		p.config.Timeout = 60 * time.Second // Default timeout
	}

	if p.config.Timeout < 30*time.Second || p.config.Timeout > 5*time.Minute {
		return fmt.Errorf("WebAuthn timeout must be between 30 seconds and 5 minutes, got: %v", p.config.Timeout)
	}

	return nil
}

// SetWebAuthnClient allows injection of a custom WebAuthn client for testing
func (p *WebAuthnProvider) SetWebAuthnClient(client WebAuthnClient) {
	p.client = client
}

// Helper methods for the default WebAuthn client implementation

func (c *DefaultWebAuthnClient) BeginRegistration(ctx context.Context, user *WebAuthnUser, options *RegistrationOptions) (*CredentialCreation, error) {
	// This is a simplified implementation
	// In production, you'd use a proper WebAuthn library like github.com/go-webauthn/webauthn
	return &CredentialCreation{
		PublicKey: *options,
	}, nil
}

func (c *DefaultWebAuthnClient) FinishRegistration(ctx context.Context, user *WebAuthnUser, sessionData *SessionData, response *CredentialCreationResponse) (*Credential, error) {
	// This is a simplified implementation
	// In production, you'd validate the attestation and create a proper credential
	return &Credential{
		ID:              response.RawID,
		PublicKey:       []byte("mock-public-key"),
		AttestationType: "none",
		Authenticator: Authenticator{
			AAGUID:    make([]byte, 16),
			SignCount: 1,
		},
	}, nil
}

func (c *DefaultWebAuthnClient) BeginLogin(ctx context.Context, user *WebAuthnUser, options *LoginOptions) (*CredentialAssertion, error) {
	// This is a simplified implementation
	return &CredentialAssertion{
		PublicKey: *options,
	}, nil
}

func (c *DefaultWebAuthnClient) FinishLogin(ctx context.Context, user *WebAuthnUser, sessionData *SessionData, response *CredentialAssertionResponse) (*Credential, error) {
	// This is a simplified implementation
	// In production, you'd validate the signature against the stored public key
	return &Credential{
		ID:        response.RawID,
		PublicKey: []byte("mock-public-key"),
	}, nil
}

// GetRegistrationOptions returns the WebAuthn registration options for enrollment
func (p *WebAuthnProvider) GetRegistrationOptions(ctx context.Context, secret *mfa.MFASecret) (string, error) {
	if secret.Type != mfa.MFATypeWebAuthn {
		return "", fmt.Errorf("invalid secret type for WebAuthn provider")
	}

	registrationOptions, exists := secret.Metadata["registration_options"]
	if !exists {
		return "", fmt.Errorf("no registration options found in secret metadata")
	}

	return registrationOptions, nil
}

// GetLoginOptions returns the WebAuthn login options for authentication
func (p *WebAuthnProvider) GetLoginOptions(ctx context.Context, challenge *mfa.MFAChallenge) (string, error) {
	if challenge.Type != mfa.MFATypeWebAuthn {
		return "", fmt.Errorf("invalid challenge type for WebAuthn provider")
	}

	// The challenge already contains the login options
	return challenge.Challenge, nil
}
