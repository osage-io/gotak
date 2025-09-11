package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTokenStorage implements TokenStorage interface for testing
type MockTokenStorage struct {
	tokens map[string]MockTokenInfo
}

type MockTokenInfo struct {
	UserID    string
	Hash      string
	ExpiresAt time.Time
	Revoked   bool
}

func NewMockTokenStorage() *MockTokenStorage {
	return &MockTokenStorage{
		tokens: make(map[string]MockTokenInfo),
	}
}

func (m *MockTokenStorage) StoreRefreshToken(userID string, tokenHash string, expiresAt time.Time) error {
	key := userID + ":" + tokenHash
	m.tokens[key] = MockTokenInfo{
		UserID:    userID,
		Hash:      tokenHash,
		ExpiresAt: expiresAt,
		Revoked:   false,
	}
	return nil
}

func (m *MockTokenStorage) ValidateRefreshToken(userID string, tokenHash string) (bool, error) {
	key := userID + ":" + tokenHash
	token, exists := m.tokens[key]
	if !exists {
		return false, nil
	}
	
	if token.Revoked {
		return false, nil
	}
	
	if time.Now().After(token.ExpiresAt) {
		return false, nil
	}
	
	return true, nil
}

func (m *MockTokenStorage) RevokeRefreshToken(userID string, tokenHash string) error {
	key := userID + ":" + tokenHash
	if token, exists := m.tokens[key]; exists {
		token.Revoked = true
		m.tokens[key] = token
	}
	return nil
}

func (m *MockTokenStorage) RevokeAllUserTokens(userID string) error {
	for key, token := range m.tokens {
		if token.UserID == userID {
			token.Revoked = true
			m.tokens[key] = token
		}
	}
	return nil
}

func TestJWTManager_GenerateTokenPair(t *testing.T) {
	storage := NewMockTokenStorage()
	config := JWTConfig{
		SecretKey:  "test-secret-key-for-jwt-testing",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
		Issuer:     "test-gotak-server",
	}
	
	jwtManager := NewJWTManager(config, storage)
	
	userID := "user-123"
	username := "testuser"
	roles := []string{"operator", "mission_commander"}
	permissions := []string{"missions.read", "cot.send"}
	
	tokenPair, err := jwtManager.GenerateTokenPair(userID, username, roles, permissions)
	require.NoError(t, err)
	require.NotNil(t, tokenPair)
	
	// Verify token pair structure
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, "Bearer", tokenPair.TokenType)
	assert.Equal(t, int64(900), tokenPair.ExpiresIn) // 15 minutes
	assert.Equal(t, int64(86400), tokenPair.RefreshExpiresIn) // 24 hours
}

func TestJWTManager_ValidateAccessToken(t *testing.T) {
	storage := NewMockTokenStorage()
	config := JWTConfig{
		SecretKey:  "test-secret-key-for-jwt-testing",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
		Issuer:     "test-gotak-server",
	}
	
	jwtManager := NewJWTManager(config, storage)
	
	userID := "user-123"
	username := "testuser"
	roles := []string{"operator"}
	permissions := []string{"missions.read"}
	
	// Generate token pair
	tokenPair, err := jwtManager.GenerateTokenPair(userID, username, roles, permissions)
	require.NoError(t, err)
	
	// Validate access token
	claims, err := jwtManager.ValidateToken(tokenPair.AccessToken)
	require.NoError(t, err)
	require.NotNil(t, claims)
	
	// Verify claims
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, "access", claims.TokenType)
	assert.Equal(t, roles, claims.Roles)
	assert.Equal(t, permissions, claims.Permissions)
	assert.NotEmpty(t, claims.SessionID)
	assert.Equal(t, "test-gotak-server", claims.Issuer)
}

func TestJWTManager_ValidateRefreshToken(t *testing.T) {
	storage := NewMockTokenStorage()
	config := JWTConfig{
		SecretKey:  "test-secret-key-for-jwt-testing",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
		Issuer:     "test-gotak-server",
	}
	
	jwtManager := NewJWTManager(config, storage)
	
	userID := "user-123"
	username := "testuser"
	
	// Generate token pair
	tokenPair, err := jwtManager.GenerateTokenPair(userID, username, nil, nil)
	require.NoError(t, err)
	
	// Validate refresh token
	claims, err := jwtManager.ValidateToken(tokenPair.RefreshToken)
	require.NoError(t, err)
	require.NotNil(t, claims)
	
	// Verify claims
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, "refresh", claims.TokenType)
	assert.NotEmpty(t, claims.SessionID)
}

func TestJWTManager_RefreshTokens(t *testing.T) {
	storage := NewMockTokenStorage()
	config := JWTConfig{
		SecretKey:  "test-secret-key-for-jwt-testing",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
		Issuer:     "test-gotak-server",
	}
	
	jwtManager := NewJWTManager(config, storage)
	
	userID := "user-123"
	username := "testuser"
	roles := []string{"operator"}
	permissions := []string{"missions.read"}
	
	// Generate initial token pair
	originalTokenPair, err := jwtManager.GenerateTokenPair(userID, username, roles, permissions)
	require.NoError(t, err)
	
	// Wait a moment to ensure different timestamps
	time.Sleep(time.Millisecond * 100)
	
	// Refresh tokens with updated roles/permissions
	newRoles := []string{"mission_commander"}
	newPermissions := []string{"missions.read", "missions.write"}
	
	newTokenPair, err := jwtManager.RefreshTokens(originalTokenPair.RefreshToken, newRoles, newPermissions)
	require.NoError(t, err)
	require.NotNil(t, newTokenPair)
	
	// Verify new tokens are different
	assert.NotEqual(t, originalTokenPair.AccessToken, newTokenPair.AccessToken)
	assert.NotEqual(t, originalTokenPair.RefreshToken, newTokenPair.RefreshToken)
	
	// Validate new access token has updated claims
	claims, err := jwtManager.ValidateToken(newTokenPair.AccessToken)
	require.NoError(t, err)
	
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, newRoles, claims.Roles)
	assert.Equal(t, newPermissions, claims.Permissions)
	
	// Original refresh token should be revoked (won't validate for refresh)
	_, err = jwtManager.RefreshTokens(originalTokenPair.RefreshToken, roles, permissions)
	assert.Error(t, err)
}

func TestJWTManager_InvalidToken(t *testing.T) {
	storage := NewMockTokenStorage()
	config := JWTConfig{
		SecretKey:  "test-secret-key-for-jwt-testing",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
		Issuer:     "test-gotak-server",
	}
	
	jwtManager := NewJWTManager(config, storage)
	
	// Test invalid token format
	_, err := jwtManager.ValidateToken("invalid-token")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidToken)
	
	// Test token with wrong secret
	wrongSecretManager := NewJWTManager(JWTConfig{
		SecretKey: "wrong-secret",
		AccessTTL: 15 * time.Minute,
	}, nil)
	
	tokenPair, err := jwtManager.GenerateTokenPair("user-123", "testuser", nil, nil)
	require.NoError(t, err)
	
	_, err = wrongSecretManager.ValidateToken(tokenPair.AccessToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	storage := NewMockTokenStorage()
	config := JWTConfig{
		SecretKey:  "test-secret-key-for-jwt-testing",
		AccessTTL:  time.Millisecond * 100, // Very short expiration
		RefreshTTL: time.Millisecond * 200,
		Issuer:     "test-gotak-server",
	}
	
	jwtManager := NewJWTManager(config, storage)
	
	tokenPair, err := jwtManager.GenerateTokenPair("user-123", "testuser", nil, nil)
	require.NoError(t, err)
	
	// Wait for token to expire
	time.Sleep(time.Millisecond * 150)
	
	_, err = jwtManager.ValidateToken(tokenPair.AccessToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestJWTManager_TokenInfo(t *testing.T) {
	storage := NewMockTokenStorage()
	config := JWTConfig{
		SecretKey:  "test-secret-key-for-jwt-testing",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
		Issuer:     "test-gotak-server",
	}
	
	jwtManager := NewJWTManager(config, storage)
	
	userID := "user-123"
	username := "testuser"
	
	tokenPair, err := jwtManager.GenerateTokenPair(userID, username, nil, nil)
	require.NoError(t, err)
	
	// Get access token info
	accessTokenInfo, err := jwtManager.GetTokenInfo(tokenPair.AccessToken)
	require.NoError(t, err)
	require.NotNil(t, accessTokenInfo)
	
	assert.Equal(t, userID, accessTokenInfo.UserID)
	assert.Equal(t, username, accessTokenInfo.Username)
	assert.Equal(t, "access", accessTokenInfo.TokenType)
	assert.False(t, accessTokenInfo.Expired)
	
	// Get refresh token info
	refreshTokenInfo, err := jwtManager.GetTokenInfo(tokenPair.RefreshToken)
	require.NoError(t, err)
	require.NotNil(t, refreshTokenInfo)
	
	assert.Equal(t, userID, refreshTokenInfo.UserID)
	assert.Equal(t, username, refreshTokenInfo.Username)
	assert.Equal(t, "refresh", refreshTokenInfo.TokenType)
	assert.False(t, refreshTokenInfo.Expired)
}
