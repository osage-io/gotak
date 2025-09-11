package providers

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/mail"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"

	"github.com/dfedick/gotak/pkg/mfa"
)

// EmailProvider implements the MFAProvider interface for Email-based authentication
type EmailProvider struct {
	config *mfa.EmailConfig
	smtp   SMTPClient
}

// SMTPClient interface allows for testing with mock SMTP clients
type SMTPClient interface {
	SendMail(addr string, auth smtp.Auth, from string, to []string, msg []byte) error
}

// DefaultSMTPClient wraps the standard library smtp package
type DefaultSMTPClient struct{}

func (c *DefaultSMTPClient) SendMail(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, auth, from, to, msg)
}

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	From     string `yaml:"from" json:"from"`
	UseTLS   bool   `yaml:"use_tls" json:"use_tls"`
}

// EmailMessage represents an email message to be sent
type EmailMessage struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Code    string `json:"-"` // Not included in JSON, used for template
}

// NewEmailProvider creates a new Email provider with the given configuration
func NewEmailProvider(config *mfa.EmailConfig) *EmailProvider {
	return &EmailProvider{
		config: config,
		smtp:   &DefaultSMTPClient{},
	}
}

// GetType returns the MFA type this provider handles
func (p *EmailProvider) GetType() mfa.MFAType {
	return mfa.MFATypeEmail
}

// GenerateSecret creates a new Email MFA secret for enrollment
func (p *EmailProvider) GenerateSecret(ctx context.Context, userID uuid.UUID, metadata map[string]string) (*mfa.MFASecret, error) {
	// For Email, the "secret" is the email address
	emailAddress, exists := metadata["email_address"]
	if !exists || emailAddress == "" {
		return nil, fmt.Errorf("email address is required for Email MFA enrollment")
	}

	// Validate email address format
	if !p.isValidEmailAddress(emailAddress) {
		return nil, fmt.Errorf("invalid email address format")
	}

	return &mfa.MFASecret{
		ID:       uuid.New(),
		UserID:   userID,
		Type:     mfa.MFATypeEmail,
		Secret:   emailAddress, // Store the email address as the secret
		Metadata: metadata,
	}, nil
}

// VerifyEnrollment validates Email enrollment by sending a verification code
func (p *EmailProvider) VerifyEnrollment(ctx context.Context, secret *mfa.MFASecret, verificationCode string) error {
	if secret.Type != mfa.MFATypeEmail {
		return fmt.Errorf("invalid secret type for Email provider: %s", secret.Type)
	}

	// For enrollment, we need to generate and send a code, then verify it
	// Similar to SMS, this assumes the verification code was already sent
	storedCode, exists := secret.Metadata["verification_code"]
	if !exists {
		return fmt.Errorf("no verification code found for Email enrollment")
	}

	if storedCode != verificationCode {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "invalid Email verification code", nil)
	}

	return nil
}

// CreateChallenge generates a new Email authentication challenge
func (p *EmailProvider) CreateChallenge(ctx context.Context, factor *mfa.MFAFactor) (*mfa.MFAChallenge, error) {
	if factor.Type != mfa.MFATypeEmail {
		return nil, fmt.Errorf("invalid factor type for Email provider: %s", factor.Type)
	}

	// Generate random verification code
	code, err := p.generateVerificationCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification code: %w", err)
	}

	// Send Email with the verification code
	emailAddress := factor.Secret
	if err := p.sendEmail(ctx, emailAddress, code); err != nil {
		return nil, fmt.Errorf("failed to send Email verification code: %w", err)
	}

	challenge := &mfa.MFAChallenge{
		ID:          uuid.New(),
		FactorID:    factor.ID,
		Type:        mfa.MFATypeEmail,
		Status:      mfa.ChallengeStatusPending,
		Challenge:   code, // Store the code for verification (should be encrypted in practice)
		Attempts:    0,
		MaxAttempts: 3,
		ExpiresAt:   time.Now().Add(p.config.CodeTTL),
		CreatedAt:   time.Now(),
	}

	return challenge, nil
}

// VerifyChallenge validates a user's response to an Email authentication challenge
func (p *EmailProvider) VerifyChallenge(ctx context.Context, challenge *mfa.MFAChallenge, response string) error {
	if challenge.Type != mfa.MFATypeEmail {
		return fmt.Errorf("invalid challenge type for Email provider: %s", challenge.Type)
	}

	// Check if challenge has expired
	if time.Now().After(challenge.ExpiresAt) {
		return mfa.NewMFAError(mfa.ErrTypeChallengeExpired, "Email challenge has expired", nil)
	}

	// Verify the response against the challenge code
	if challenge.Challenge != response {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "invalid Email verification code", nil)
	}

	return nil
}

// ValidateConfiguration checks if the provider configuration is valid
func (p *EmailProvider) ValidateConfiguration() error {
	if p.config == nil {
		return fmt.Errorf("Email configuration cannot be nil")
	}

	if !p.config.Enabled {
		return fmt.Errorf("Email provider is disabled")
	}

	if p.config.Provider == "" {
		return fmt.Errorf("Email provider type cannot be empty")
	}

	if p.config.Template == "" {
		return fmt.Errorf("Email template cannot be empty")
	}

	// Validate code length
	if p.config.CodeLength == 0 {
		p.config.CodeLength = 6 // Default
	}
	if p.config.CodeLength < 4 || p.config.CodeLength > 8 {
		return fmt.Errorf("Email code length must be between 4 and 8, got: %d", p.config.CodeLength)
	}

	// Validate code TTL
	if p.config.CodeTTL == 0 {
		p.config.CodeTTL = 5 * time.Minute // Default
	}
	if p.config.CodeTTL < time.Minute || p.config.CodeTTL > 30*time.Minute {
		return fmt.Errorf("Email code TTL must be between 1 and 30 minutes, got: %v", p.config.CodeTTL)
	}

	return nil
}

// Helper methods

// generateVerificationCode creates a random numeric verification code
func (p *EmailProvider) generateVerificationCode() (string, error) {
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

// sendEmail sends an email message using the configured provider
func (p *EmailProvider) sendEmail(ctx context.Context, emailAddress, code string) error {
	// Render the email template
	subject, body, err := p.renderTemplate(code)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Send email based on configured provider
	switch strings.ToLower(p.config.Provider) {
	case "smtp":
		return p.sendSMTPEmail(ctx, emailAddress, subject, body)
	case "aws-ses":
		return p.sendAWSSESEmail(ctx, emailAddress, subject, body)
	case "mock":
		return p.sendMockEmail(ctx, emailAddress, subject, body)
	default:
		return fmt.Errorf("unsupported email provider: %s", p.config.Provider)
	}
}

// renderTemplate renders the email template with the verification code
func (p *EmailProvider) renderTemplate(code string) (string, string, error) {
	tmpl, err := template.New("email").Parse(p.config.Template)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse email template: %w", err)
	}

	data := struct {
		Code string
	}{
		Code: code,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute email template: %w", err)
	}

	content := buf.String()

	// Extract subject and body from template
	// Expected format: "Subject: ...\n\n..."
	parts := strings.SplitN(content, "\n\n", 2)
	if len(parts) != 2 {
		return "GoTAK Verification Code", content, nil // Fallback
	}

	subject := strings.TrimPrefix(parts[0], "Subject: ")
	body := parts[1]

	return subject, body, nil
}

// sendSMTPEmail sends email using SMTP
func (p *EmailProvider) sendSMTPEmail(ctx context.Context, toEmail, subject, body string) error {
	// In a real implementation, these would come from configuration
	smtpConfig := SMTPConfig{
		Host:     "smtp.gmail.com",
		Port:     587,
		Username: "your-email@gmail.com",
		Password: "your-app-password",
		From:     "your-email@gmail.com",
		UseTLS:   true,
	}

	// Create SMTP auth
	auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)

	// Construct email message
	msg := p.buildEmailMessage(smtpConfig.From, toEmail, subject, body)

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", smtpConfig.Host, smtpConfig.Port)

	// Send email
	if err := p.smtp.SendMail(addr, auth, smtpConfig.From, []string{toEmail}, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send SMTP email: %w", err)
	}

	return nil
}

// buildEmailMessage constructs a proper email message with headers
func (p *EmailProvider) buildEmailMessage(from, to, subject, body string) string {
	return fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)
}

// sendAWSSESEmail sends email using AWS SES (placeholder implementation)
func (p *EmailProvider) sendAWSSESEmail(ctx context.Context, toEmail, subject, body string) error {
	// This would implement AWS SES email sending
	return fmt.Errorf("AWS SES email provider not yet implemented")
}

// sendMockEmail is a mock implementation for testing
func (p *EmailProvider) sendMockEmail(ctx context.Context, toEmail, subject, body string) error {
	// In test/dev mode, just log the email instead of sending
	fmt.Printf("MOCK EMAIL to %s:\nSubject: %s\nBody: %s\n", toEmail, subject, body)
	return nil
}

// isValidEmailAddress performs basic email validation
func (p *EmailProvider) isValidEmailAddress(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// SendEnrollmentEmail sends an email verification code for enrollment
func (p *EmailProvider) SendEnrollmentEmail(ctx context.Context, emailAddress string) (string, error) {
	code, err := p.generateVerificationCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate enrollment code: %w", err)
	}

	if err := p.sendEmail(ctx, emailAddress, code); err != nil {
		return "", fmt.Errorf("failed to send enrollment email: %w", err)
	}

	return code, nil
}

// VerifyEmailWithCode is a helper method to verify email codes directly
func (p *EmailProvider) VerifyEmailWithCode(expectedCode, providedCode string) error {
	if expectedCode != providedCode {
		return mfa.NewMFAError(mfa.ErrTypeVerificationFailed, "invalid email verification code", nil)
	}
	return nil
}

// SetSMTPClient allows injection of a custom SMTP client for testing
func (p *EmailProvider) SetSMTPClient(client SMTPClient) {
	p.smtp = client
}
