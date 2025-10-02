package auth

import (
	"sync"
	"time"
)

// InMemoryTokenStorage provides in-memory token storage (for development)
type InMemoryTokenStorage struct {
	mu            sync.RWMutex
	refreshTokens map[string]RefreshTokenData
}

// RefreshTokenData holds refresh token data
type RefreshTokenData struct {
	UserID    string
	TokenHash string
	ExpiresAt time.Time
}

// NewInMemoryTokenStorage creates a new in-memory token storage
func NewInMemoryTokenStorage() *InMemoryTokenStorage {
	return &InMemoryTokenStorage{
		refreshTokens: make(map[string]RefreshTokenData),
	}
}

// StoreRefreshToken stores a refresh token
func (s *InMemoryTokenStorage) StoreRefreshToken(userID string, tokenHash string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refreshTokens[tokenHash] = RefreshTokenData{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	return nil
}

// ValidateRefreshToken validates a refresh token
func (s *InMemoryTokenStorage) ValidateRefreshToken(userID string, tokenHash string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.refreshTokens[tokenHash]
	if !exists {
		return false, nil
	}

	// Check if token belongs to user and hasn't expired
	return data.UserID == userID && time.Now().Before(data.ExpiresAt), nil
}

// RevokeRefreshToken revokes a specific refresh token
func (s *InMemoryTokenStorage) RevokeRefreshToken(userID string, tokenHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.refreshTokens, tokenHash)
	return nil
}

// RevokeAllUserTokens revokes all tokens for a user
func (s *InMemoryTokenStorage) RevokeAllUserTokens(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for hash, data := range s.refreshTokens {
		if data.UserID == userID {
			delete(s.refreshTokens, hash)
		}
	}
	return nil
}

// CleanupExpiredTokens removes expired tokens (should be called periodically)
func (s *InMemoryTokenStorage) CleanupExpiredTokens() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for hash, data := range s.refreshTokens {
		if now.After(data.ExpiresAt) {
			delete(s.refreshTokens, hash)
		}
	}
}