package providers

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"

	"github.com/dfedick/gotak/pkg/mfa"
)

// SMSProvider implements the MFAProvider interface for SMS-based authentication
type SMSProvider struct {
	config *mfa.SMSConfig
	client HTTPClient
}

// HTTPClient interface allows for testing with mock HTTP clients
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// TwilioSMSProvider is a concrete implementation using Twilio API
type TwilioSMSProvider struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	BaseURL    string
}

// SMSMessage represents an SMS message to be sent
type SMSMessage struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Body    string `json:"body"`
	Code    string `json:"-"` // Not included in JSON, used for template
}

// NewSMSProvider creates a new SMS provider with the given configuration
func NewSMSProvider(config *mfa.SMSConfig) *SMSProvider {
	return &SMSProvider{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetType returns the MFA type this provider handles
func (p *SMSProvider) GetType() mfa.MFAType {
	return mfa.MFATypeSMS
}

// GenerateSecret creates a new SMS MFA secret for enrollment
func (p *SMSProvider) GenerateSecret(ctx context.Context, userID uuid.UUID, metadata map[string]string) (*mfa.MFASecret, error) {
	// For SMS, the "secret" is really just the phone number and configuration
	phoneNumber, exists := metadata["phone_number"]
	if !exists || phoneNumber == "" {
		return nil, fmt.Errorf("phone number is required for SMS MFA enrollment")
	}

	// Validate phone number format (basic validation)
	if !p.isValidPhoneNumber(phoneNumber) {
		return nil, fmt.Errorf("invalid phone number format")
	}

	return &mfa.MFASecret{
		ID:       uuid.New(),
		UserID:   userID,
		Type:     mfa.MFATypeSMS,
		Secret:   phoneNumber, // Store the phone number as the secret
		Metadata: metadata,
	}, nil
}

// VerifyEnrollment validates SMS enrollment by sending a verification code
func (p *SMSProvider) VerifyEnrollment(ctx context.Context, secret *mfa.MFASecret, verificationCode string) error {
	if secret.Type != mfa.MFATypeSMS {
		return fmt.Errorf("invalid secret type for SMS provider: %s", secret.Type)
	}

	// For enrollment, we need to generate and send a code, then verify it
	// In practice, this would involve a two-step process:
	// 1. GenerateEnrollmentCode() - sends SMS with code
	// 2. VerifyEnrollmentCode() - validates the code
	
	// For this implementation, we'll assume the verification code was already sent
	// and stored in the enrollment session, and we're just validating it here
	
	storedCode, exists := secret.Metadata["verification_code"]
	if !exists {
		return fmt.Errorf("no verification code found for SMS enrollment")
	}

	if storedCode != verificationCode {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "invalid SMS verification code", nil)
	}

	return nil
}

// CreateChallenge generates a new SMS authentication challenge
func (p *SMSProvider) CreateChallenge(ctx context.Context, factor *mfa.MFAFactor) (*mfa.MFAChallenge, error) {
	if factor.Type != mfa.MFATypeSMS {
		return nil, fmt.Errorf("invalid factor type for SMS provider: %s", factor.Type)
	}

	// Generate random verification code
	code, err := p.generateVerificationCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification code: %w", err)
	}

	// Send SMS with the verification code
	phoneNumber := factor.Secret
	if err := p.sendSMS(ctx, phoneNumber, code); err != nil {
		return nil, fmt.Errorf("failed to send SMS verification code: %w", err)
	}

	challenge := &mfa.MFAChallenge{
		ID:          uuid.New(),
		FactorID:    factor.ID,
		Type:        mfa.MFATypeSMS,
		Status:      mfa.ChallengeStatusPending,
		Challenge:   code, // Store the code for verification (should be encrypted in practice)
		Attempts:    0,
		MaxAttempts: 3,
		ExpiresAt:   time.Now().Add(p.config.CodeTTL),
		CreatedAt:   time.Now(),
	}

	return challenge, nil
}

// VerifyChallenge validates a user's response to an SMS authentication challenge
func (p *SMSProvider) VerifyChallenge(ctx context.Context, challenge *mfa.MFAChallenge, response string) error {
	if challenge.Type != mfa.MFATypeSMS {
		return fmt.Errorf("invalid challenge type for SMS provider: %s", challenge.Type)
	}

	// Check if challenge has expired
	if time.Now().After(challenge.ExpiresAt) {
		return mfa.NewMFAError(mfa.ErrTypeChallengeExpired, "SMS challenge has expired", nil)
	}

	// Verify the response against the challenge code
	if challenge.Challenge != response {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "invalid SMS verification code", nil)
	}

	return nil
}

// ValidateConfiguration checks if the provider configuration is valid
func (p *SMSProvider) ValidateConfiguration() error {
	if p.config == nil {
		return fmt.Errorf("SMS configuration cannot be nil")
	}

	if !p.config.Enabled {
		return fmt.Errorf("SMS provider is disabled")
	}

	if p.config.Provider == "" {
		return fmt.Errorf("SMS provider type cannot be empty")
	}

	if p.config.Template == "" {
		return fmt.Errorf("SMS template cannot be empty")
	}

	// Validate code length
	if p.config.CodeLength == 0 {
		p.config.CodeLength = 6 // Default
	}
	if p.config.CodeLength < 4 || p.config.CodeLength > 8 {
		return fmt.Errorf("SMS code length must be between 4 and 8, got: %d", p.config.CodeLength)
	}

	// Validate code TTL
	if p.config.CodeTTL == 0 {
		p.config.CodeTTL = 5 * time.Minute // Default
	}
	if p.config.CodeTTL < time.Minute || p.config.CodeTTL > 15*time.Minute {
		return fmt.Errorf("SMS code TTL must be between 1 and 15 minutes, got: %v", p.config.CodeTTL)
	}

	return nil
}

// Helper methods

// generateVerificationCode creates a random numeric verification code
func (p *SMSProvider) generateVerificationCode() (string, error) {
	max := int64(1)
	for i := 0; i < p.config.CodeLength; i++ {
		max *= 10
	}
	max-- // Make it so max digits, e.g., 999999 for 6 digits

	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", err
	}

	// Format with leading zeros to ensure correct length
	format := fmt.Sprintf("%%0%dd", p.config.CodeLength)
	return fmt.Sprintf(format, n.Int64()), nil
}

// sendSMS sends an SMS message using the configured provider
func (p *SMSProvider) sendSMS(ctx context.Context, phoneNumber, code string) error {
	// Render the message template
	message, err := p.renderTemplate(code)
	if err != nil {
		return fmt.Errorf("failed to render SMS template: %w", err)
	}

	// Send SMS based on configured provider
	switch strings.ToLower(p.config.Provider) {
	case "twilio":
		return p.sendTwilioSMS(ctx, phoneNumber, message)
	case "aws-sns":
		return p.sendAWSSNSSMS(ctx, phoneNumber, message)
	case "mock":
		return p.sendMockSMS(ctx, phoneNumber, message)
	default:
		return fmt.Errorf("unsupported SMS provider: %s", p.config.Provider)
	}
}

// renderTemplate renders the SMS message template with the verification code
func (p *SMSProvider) renderTemplate(code string) (string, error) {
	tmpl, err := template.New("sms").Parse(p.config.Template)
	if err != nil {
		return "", fmt.Errorf("failed to parse SMS template: %w", err)
	}

	data := struct {
		Code string
	}{
		Code: code,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute SMS template: %w", err)
	}

	return buf.String(), nil
}

// sendTwilioSMS sends SMS using Twilio API
func (p *SMSProvider) sendTwilioSMS(ctx context.Context, phoneNumber, message string) error {
	// In a real implementation, you would get these from configuration or environment
	accountSID := "your_twilio_account_sid"
	authToken := "your_twilio_auth_token"
	fromNumber := "your_twilio_phone_number"

	// Twilio API endpoint
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	// Prepare form data
	data := url.Values{}
	data.Set("To", phoneNumber)
	data.Set("From", fromNumber)
	data.Set("Body", message)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create Twilio request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(accountSID, authToken)

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Twilio SMS: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Twilio API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// sendAWSSNSSMS sends SMS using AWS SNS (placeholder implementation)
func (p *SMSProvider) sendAWSSNSSMS(ctx context.Context, phoneNumber, message string) error {
	// This would implement AWS SNS SMS sending
	// For now, return a placeholder implementation
	return fmt.Errorf("AWS SNS SMS provider not yet implemented")
}

// sendMockSMS is a mock implementation for testing
func (p *SMSProvider) sendMockSMS(ctx context.Context, phoneNumber, message string) error {
	// In test/dev mode, just log the SMS instead of sending
	fmt.Printf("MOCK SMS to %s: %s\n", phoneNumber, message)
	return nil
}

// isValidPhoneNumber performs basic phone number validation
func (p *SMSProvider) isValidPhoneNumber(phone string) bool {
	// Remove common formatting characters
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")

	// Check if it starts with + (international format)
	if strings.HasPrefix(cleaned, "+") {
		cleaned = cleaned[1:]
	}

	// Must be all digits and reasonable length
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return false
	}

	for _, r := range cleaned {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

// SendEnrollmentSMS sends an SMS verification code for enrollment
func (p *SMSProvider) SendEnrollmentSMS(ctx context.Context, phoneNumber string) (string, error) {
	code, err := p.generateVerificationCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate enrollment code: %w", err)
	}

	if err := p.sendSMS(ctx, phoneNumber, code); err != nil {
		return "", fmt.Errorf("failed to send enrollment SMS: %w", err)
	}

	return code, nil
}

// VerifySMSWithCode is a helper method to verify SMS codes directly
func (p *SMSProvider) VerifySMSWithCode(expectedCode, providedCode string) error {
	if expectedCode != providedCode {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "invalid SMS verification code", nil)
	}
	return nil
}
