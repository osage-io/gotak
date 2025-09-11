package mfa

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// DefaultMFAManager provides a concrete implementation of the MFAManager interface
type DefaultMFAManager struct {
	config    *MFAConfig
	storage   MFAStorage
	providers map[MFAType]MFAProvider
	logger    zerolog.Logger
	mu        sync.RWMutex
}

// NewMFAManager creates a new MFA manager with the given configuration and storage
func NewMFAManager(config *MFAConfig, storage MFAStorage, logger zerolog.Logger) *DefaultMFAManager {
	return &DefaultMFAManager{
		config:    config,
		storage:   storage,
		providers: make(map[MFAType]MFAProvider),
		logger:    logger,
	}
}

// RegisterProvider registers a new MFA provider
func (m *DefaultMFAManager) RegisterProvider(provider MFAProvider) error {
	if provider == nil {
		return NewMFAError(ErrTypeInvalidProvider, "provider cannot be nil", nil)
	}

	if err := provider.ValidateConfiguration(); err != nil {
		return NewMFAError(ErrTypeInvalidConfiguration, 
			fmt.Sprintf("provider configuration validation failed: %v", err), err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	mfaType := provider.GetType()
	m.providers[mfaType] = provider

	m.logger.Info().
		Str("mfa_type", string(mfaType)).
		Msg("MFA provider registered")

	return nil
}

// GetProvider returns a specific MFA provider by type
func (m *DefaultMFAManager) GetProvider(mfaType MFAType) (MFAProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[mfaType]
	if !exists {
		return nil, NewMFAError(ErrTypeProviderNotFound, 
			fmt.Sprintf("provider not found for type: %s", mfaType), nil)
	}

	return provider, nil
}

// ListProviders returns all registered providers
func (m *DefaultMFAManager) ListProviders() []MFAProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]MFAProvider, 0, len(m.providers))
	for _, provider := range m.providers {
		providers = append(providers, provider)
	}

	return providers
}

// EnrollFactor initiates enrollment for a new MFA factor
func (m *DefaultMFAManager) EnrollFactor(ctx context.Context, userID uuid.UUID, mfaType MFAType, metadata map[string]string) (*EnrollmentSession, error) {
	// Get the provider for this MFA type
	provider, err := m.GetProvider(mfaType)
	if err != nil {
		return nil, err
	}

	// Generate a secret for enrollment
	secret, err := provider.GenerateSecret(ctx, userID, metadata)
	if err != nil {
		m.recordMFAEvent(ctx, userID, nil, nil, "enrollment_started", "error", 
			map[string]string{"error": err.Error()})
		return nil, NewMFAError(ErrTypeEnrollmentFailed, 
			fmt.Sprintf("failed to generate secret: %v", err), err)
	}

	// Create enrollment session
	session := &EnrollmentSession{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      mfaType,
		Secret:    secret,
		Metadata:  metadata,
		ExpiresAt: time.Now().Add(m.config.EnrollmentLifetime),
		CreatedAt: time.Now(),
	}

	// Store the enrollment session
	if err := m.storage.CreateEnrollmentSession(ctx, session); err != nil {
		m.recordMFAEvent(ctx, userID, nil, nil, "enrollment_started", "error", 
			map[string]string{"error": err.Error()})
		return nil, NewMFAError(ErrTypeStorageError, 
			fmt.Sprintf("failed to store enrollment session: %v", err), err)
	}

	m.recordMFAEvent(ctx, userID, nil, &session.ID, "enrollment_started", "success", nil)

	m.logger.Info().
		Str("user_id", userID.String()).
		Str("session_id", session.ID.String()).
		Str("mfa_type", string(mfaType)).
		Msg("MFA enrollment started")

	return session, nil
}

// CompleteFactor completes enrollment and activates the MFA factor
func (m *DefaultMFAManager) CompleteFactor(ctx context.Context, sessionID uuid.UUID, verificationCode string) (*MFAFactor, error) {
	// Get enrollment session
	session, err := m.storage.GetEnrollmentSession(ctx, sessionID)
	if err != nil {
		return nil, NewMFAError(ErrTypeStorageError, 
			fmt.Sprintf("failed to get enrollment session: %v", err), err)
	}

	// Check if session is still valid
	if time.Now().After(session.ExpiresAt) {
		m.storage.DeleteEnrollmentSession(ctx, sessionID)
		m.recordMFAEvent(ctx, session.UserID, nil, &sessionID, "enrollment_completed", "expired", nil)
		return nil, NewMFAError(ErrTypeChallengeExpired, "enrollment session expired", nil)
	}

	// Get provider and verify enrollment
	provider, err := m.GetProvider(session.Type)
	if err != nil {
		return nil, err
	}

	if err := provider.VerifyEnrollment(ctx, session.Secret, verificationCode); err != nil {
		m.recordMFAEvent(ctx, session.UserID, nil, &sessionID, "enrollment_completed", "verification_failed", 
			map[string]string{"error": err.Error()})
		return nil, NewMFAError(ErrTypeVerificationFailed, 
			fmt.Sprintf("enrollment verification failed: %v", err), err)
	}

	// Create backup codes if enabled
	var backupCodes []string
	if m.config.BackupCodesEnabled {
		backupCodes, err = m.generateBackupCodes()
		if err != nil {
			return nil, NewMFAError(ErrTypeEnrollmentFailed, 
				fmt.Sprintf("failed to generate backup codes: %v", err), err)
		}
	}

	// Create MFA factor
	factor := &MFAFactor{
		ID:          uuid.New(),
		UserID:      session.UserID,
		Type:        session.Type,
		Name:        m.getDefaultFactorName(session.Type),
		Status:      MFAStatusActive,
		Secret:      session.Secret.Secret,
		Metadata:    session.Metadata,
		BackupCodes: backupCodes,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store the factor
	if err := m.storage.CreateFactor(ctx, factor); err != nil {
		m.recordMFAEvent(ctx, session.UserID, &factor.ID, &sessionID, "enrollment_completed", "error", 
			map[string]string{"error": err.Error()})
		return nil, NewMFAError(ErrTypeStorageError, 
			fmt.Sprintf("failed to store MFA factor: %v", err), err)
	}

	// Store backup codes if generated
	if len(backupCodes) > 0 {
		if err := m.storage.StoreBackupCodes(ctx, factor.ID, backupCodes); err != nil {
			m.logger.Warn().
				Str("factor_id", factor.ID.String()).
				Err(err).
				Msg("Failed to store backup codes")
		}
	}

	// Clean up enrollment session
	m.storage.DeleteEnrollmentSession(ctx, sessionID)

	m.recordMFAEvent(ctx, session.UserID, &factor.ID, &sessionID, "enrollment_completed", "success", nil)

	m.logger.Info().
		Str("user_id", session.UserID.String()).
		Str("factor_id", factor.ID.String()).
		Str("mfa_type", string(session.Type)).
		Msg("MFA enrollment completed")

	return factor, nil
}

// ListUserFactors returns all MFA factors for a user
func (m *DefaultMFAManager) ListUserFactors(ctx context.Context, userID uuid.UUID) ([]*MFAFactor, error) {
	factors, err := m.storage.GetUserFactors(ctx, userID)
	if err != nil {
		return nil, NewMFAError(ErrTypeStorageError, 
			fmt.Sprintf("failed to get user factors: %v", err), err)
	}

	return factors, nil
}

// DisableFactor temporarily disables an MFA factor
func (m *DefaultMFAManager) DisableFactor(ctx context.Context, factorID uuid.UUID) error {
	factor, err := m.storage.GetFactor(ctx, factorID)
	if err != nil {
		return NewMFAError(ErrTypeFactorNotFound, 
			fmt.Sprintf("factor not found: %v", err), err)
	}

	factor.Status = MFAStatusDisabled
	factor.UpdatedAt = time.Now()

	if err := m.storage.UpdateFactor(ctx, factor); err != nil {
		return NewMFAError(ErrTypeStorageError, 
			fmt.Sprintf("failed to update factor: %v", err), err)
	}

	m.recordMFAEvent(ctx, factor.UserID, &factorID, nil, "factor_disabled", "success", nil)

	m.logger.Info().
		Str("factor_id", factorID.String()).
		Str("user_id", factor.UserID.String()).
		Msg("MFA factor disabled")

	return nil
}

// RevokeFactor permanently revokes an MFA factor
func (m *DefaultMFAManager) RevokeFactor(ctx context.Context, factorID uuid.UUID) error {
	factor, err := m.storage.GetFactor(ctx, factorID)
	if err != nil {
		return NewMFAError(ErrTypeFactorNotFound, 
			fmt.Sprintf("factor not found: %v", err), err)
	}

	factor.Status = MFAStatusRevoked
	factor.UpdatedAt = time.Now()

	if err := m.storage.UpdateFactor(ctx, factor); err != nil {
		return NewMFAError(ErrTypeStorageError, 
			fmt.Sprintf("failed to update factor: %v", err), err)
	}

	m.recordMFAEvent(ctx, factor.UserID, &factorID, nil, "factor_revoked", "success", nil)

	m.logger.Info().
		Str("factor_id", factorID.String()).
		Str("user_id", factor.UserID.String()).
		Msg("MFA factor revoked")

	return nil
}

// CreateAuthChallenge creates a new authentication challenge
func (m *DefaultMFAManager) CreateAuthChallenge(ctx context.Context, userID uuid.UUID, requestedTypes []MFAType) (*AuthChallenge, error) {
	// Get user's active factors
	factors, err := m.ListUserFactors(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filter for active factors
	activeFactors := make([]*MFAFactor, 0)
	for _, factor := range factors {
		if factor.Status == MFAStatusActive {
			activeFactors = append(activeFactors, factor)
		}
	}

	if len(activeFactors) == 0 {
		return nil, NewMFAError(ErrTypeFactorNotFound, "no active MFA factors found for user", nil)
	}

	// Determine required types (use config default if none specified)
	requiredTypes := requestedTypes
	if len(requiredTypes) == 0 {
		requiredTypes = m.config.RequiredTypes
	}

	// Create auth challenge
	challenge := &AuthChallenge{
		ID:               uuid.New(),
		UserID:           userID,
		RequiredTypes:    requiredTypes,
		Challenges:       make(map[uuid.UUID]*MFAChallenge),
		Status:           ChallengeStatusPending,
		CompletedFactors: make([]uuid.UUID, 0),
		ExpiresAt:        time.Now().Add(m.config.ChallengeLifetime),
		CreatedAt:        time.Now(),
	}

	// Create individual challenges for available factors
	for _, factor := range activeFactors {
		// Skip if this type is not required
		typeRequired := len(requiredTypes) == 0 // If no specific types required, accept any
		for _, reqType := range requiredTypes {
			if factor.Type == reqType {
				typeRequired = true
				break
			}
		}

		if !typeRequired {
			continue
		}

		provider, err := m.GetProvider(factor.Type)
		if err != nil {
			m.logger.Warn().
				Str("mfa_type", string(factor.Type)).
				Err(err).
				Msg("Provider not available for factor type")
			continue
		}

		mfaChallenge, err := provider.CreateChallenge(ctx, factor)
		if err != nil {
			m.logger.Warn().
				Str("factor_id", factor.ID.String()).
				Str("mfa_type", string(factor.Type)).
				Err(err).
				Msg("Failed to create MFA challenge")
			continue
		}

		// Store the MFA challenge
		if err := m.storage.CreateMFAChallenge(ctx, mfaChallenge); err != nil {
			m.logger.Warn().
				Str("challenge_id", mfaChallenge.ID.String()).
				Err(err).
				Msg("Failed to store MFA challenge")
			continue
		}

		challenge.Challenges[factor.ID] = mfaChallenge
	}

	if len(challenge.Challenges) == 0 {
		return nil, NewMFAError(ErrTypeInvalidChallenge, "no valid challenges could be created", nil)
	}

	// Store auth challenge
	if err := m.storage.CreateAuthChallenge(ctx, challenge); err != nil {
		return nil, NewMFAError(ErrTypeStorageError, 
			fmt.Sprintf("failed to store auth challenge: %v", err), err)
	}

	m.recordMFAEvent(ctx, userID, nil, &challenge.ID, "auth_challenge_created", "success", 
		map[string]string{"challenge_count": fmt.Sprintf("%d", len(challenge.Challenges))})

	m.logger.Info().
		Str("challenge_id", challenge.ID.String()).
		Str("user_id", userID.String()).
		Int("challenge_count", len(challenge.Challenges)).
		Msg("Authentication challenge created")

	return challenge, nil
}

// VerifyAuthChallenge verifies a user's response to an authentication challenge
func (m *DefaultMFAManager) VerifyAuthChallenge(ctx context.Context, challengeID uuid.UUID, factorID uuid.UUID, response string) error {
	// Get auth challenge
	challenge, err := m.storage.GetAuthChallenge(ctx, challengeID)
	if err != nil {
		return NewMFAError(ErrTypeStorageError, 
			fmt.Sprintf("failed to get auth challenge: %v", err), err)
	}

	// Check if challenge is still valid
	if time.Now().After(challenge.ExpiresAt) {
		challenge.Status = ChallengeStatusExpired
		m.storage.UpdateAuthChallenge(ctx, challenge)
		m.recordMFAEvent(ctx, challenge.UserID, &factorID, &challengeID, "auth_challenge_verified", "expired", nil)
		return NewMFAError(ErrTypeChallengeExpired, "authentication challenge expired", nil)
	}

	// Get the specific MFA challenge
	mfaChallenge, exists := challenge.Challenges[factorID]
	if !exists {
		return NewMFAError(ErrTypeInvalidChallenge, "challenge not found for factor", nil)
	}

	// Get provider and verify response
	provider, err := m.GetProvider(mfaChallenge.Type)
	if err != nil {
		return err
	}

	if err := provider.VerifyChallenge(ctx, mfaChallenge, response); err != nil {
		// Update challenge attempts
		mfaChallenge.Attempts++
		if mfaChallenge.Attempts >= mfaChallenge.MaxAttempts {
			mfaChallenge.Status = ChallengeStatusFailed
		}
		m.storage.UpdateMFAChallenge(ctx, mfaChallenge)

		m.recordMFAEvent(ctx, challenge.UserID, &factorID, &challengeID, "auth_challenge_verified", "verification_failed", 
			map[string]string{"attempts": fmt.Sprintf("%d", mfaChallenge.Attempts)})

		return NewMFAError(ErrTypeVerificationFailed, 
			fmt.Sprintf("challenge verification failed: %v", err), err)
	}

	// Mark challenge as verified
	now := time.Now()
	mfaChallenge.Status = ChallengeStatusVerified
	mfaChallenge.VerifiedAt = &now
	m.storage.UpdateMFAChallenge(ctx, mfaChallenge)

	// Update factor last used time
	factor, err := m.storage.GetFactor(ctx, factorID)
	if err == nil {
		factor.LastUsedAt = &now
		m.storage.UpdateFactor(ctx, factor)
	}

	// Add to completed factors if not already present
	factorCompleted := false
	for _, completedID := range challenge.CompletedFactors {
		if completedID == factorID {
			factorCompleted = true
			break
		}
	}
	if !factorCompleted {
		challenge.CompletedFactors = append(challenge.CompletedFactors, factorID)
	}

	m.storage.UpdateAuthChallenge(ctx, challenge)

	m.recordMFAEvent(ctx, challenge.UserID, &factorID, &challengeID, "auth_challenge_verified", "success", nil)

	m.logger.Info().
		Str("challenge_id", challengeID.String()).
		Str("factor_id", factorID.String()).
		Str("user_id", challenge.UserID.String()).
		Msg("MFA challenge verified successfully")

	return nil
}

// IsAuthChallengeComplete checks if all required factors are satisfied
func (m *DefaultMFAManager) IsAuthChallengeComplete(ctx context.Context, challengeID uuid.UUID) (bool, error) {
	challenge, err := m.storage.GetAuthChallenge(ctx, challengeID)
	if err != nil {
		return false, NewMFAError(ErrTypeStorageError, 
			fmt.Sprintf("failed to get auth challenge: %v", err), err)
	}

	// Check if challenge has expired
	if time.Now().After(challenge.ExpiresAt) {
		return false, NewMFAError(ErrTypeChallengeExpired, "authentication challenge expired", nil)
	}

	// If no specific types required, need at least one factor completed
	if len(challenge.RequiredTypes) == 0 {
		complete := len(challenge.CompletedFactors) > 0
		if complete {
			challenge.Status = ChallengeStatusVerified
			now := time.Now()
			challenge.CompletedAt = &now
			m.storage.UpdateAuthChallenge(ctx, challenge)
		}
		return complete, nil
	}

	// Check if all required types are satisfied
	requiredTypesSatisfied := make(map[MFAType]bool)
	for _, reqType := range challenge.RequiredTypes {
		requiredTypesSatisfied[reqType] = false
	}

	// Check completed factors against required types
	for _, factorID := range challenge.CompletedFactors {
		factor, err := m.storage.GetFactor(ctx, factorID)
		if err != nil {
			continue
		}
		if _, required := requiredTypesSatisfied[factor.Type]; required {
			requiredTypesSatisfied[factor.Type] = true
		}
	}

	// Check if all required types are satisfied
	allSatisfied := true
	for _, satisfied := range requiredTypesSatisfied {
		if !satisfied {
			allSatisfied = false
			break
		}
	}

	if allSatisfied {
		challenge.Status = ChallengeStatusVerified
		now := time.Now()
		challenge.CompletedAt = &now
		m.storage.UpdateAuthChallenge(ctx, challenge)

		m.recordMFAEvent(ctx, challenge.UserID, nil, &challengeID, "auth_challenge_completed", "success", nil)
	}

	return allSatisfied, nil
}

// Helper methods

func (m *DefaultMFAManager) generateBackupCodes() ([]string, error) {
	codes := make([]string, m.config.BackupCodesCount)
	
	for i := 0; i < m.config.BackupCodesCount; i++ {
		bytes := make([]byte, m.config.BackupCodesLength)
		if _, err := rand.Read(bytes); err != nil {
			return nil, err
		}
		codes[i] = base64.RawURLEncoding.EncodeToString(bytes)[:m.config.BackupCodesLength]
	}

	return codes, nil
}

func (m *DefaultMFAManager) getDefaultFactorName(mfaType MFAType) string {
	switch mfaType {
	case MFATypeTOTP:
		return "Authenticator App"
	case MFATypeSMS:
		return "SMS"
	case MFATypeEmail:
		return "Email"
	case MFATypeWebAuthn:
		return "Security Key"
	case MFATypeBackup:
		return "Backup Codes"
	default:
		return string(mfaType)
	}
}

func (m *DefaultMFAManager) recordMFAEvent(ctx context.Context, userID uuid.UUID, factorID *uuid.UUID, challengeID *uuid.UUID, eventType, result string, metadata map[string]string) {
	event := &MFAEvent{
		ID:          uuid.New(),
		UserID:      userID,
		FactorID:    factorID,
		ChallengeID: challengeID,
		EventType:   eventType,
		Result:      result,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
	}

	// Try to get IP and User Agent from context
	if ipAddr, ok := ctx.Value("ip_address").(string); ok {
		event.IPAddress = ipAddr
	}
	if userAgent, ok := ctx.Value("user_agent").(string); ok {
		event.UserAgent = userAgent
	}

	if err := m.storage.RecordMFAEvent(ctx, event); err != nil {
		m.logger.Warn().
			Err(err).
			Str("event_type", eventType).
			Str("user_id", userID.String()).
			Msg("Failed to record MFA event")
	}
}
