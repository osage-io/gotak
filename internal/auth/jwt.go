package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidTokenType = errors.New("invalid token type")
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secret        []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
	tokenStorage  TokenStorage // Interface for token storage/blacklisting
}

// TokenStorage interface for managing token storage and validation
type TokenStorage interface {
	StoreRefreshToken(userID string, tokenHash string, expiresAt time.Time) error
	ValidateRefreshToken(userID string, tokenHash string) (bool, error)
	RevokeRefreshToken(userID string, tokenHash string) error
	RevokeAllUserTokens(userID string) error
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	TokenType   string   `json:"token_type"` // "access" or "refresh"
	SessionID   string   `json:"session_id,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshExpiresIn int64 `json:"refresh_expires_in,omitempty"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey    string        `mapstructure:"secret_key"`
	AccessTTL    time.Duration `mapstructure:"access_ttl"`
	RefreshTTL   time.Duration `mapstructure:"refresh_ttl"`
	Issuer       string        `mapstructure:"issuer"`
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(config JWTConfig, storage TokenStorage) *JWTManager {
	// Set defaults if not provided
	if config.AccessTTL == 0 {
		config.AccessTTL = 15 * time.Minute
	}
	if config.RefreshTTL == 0 {
		config.RefreshTTL = 24 * 7 * time.Hour // 7 days
	}
	if config.Issuer == "" {
		config.Issuer = "gotak-server"
	}
	if config.SecretKey == "" {
		// Generate a random secret for development (should be configured in production)
		secret := make([]byte, 32)
		rand.Read(secret)
		config.SecretKey = hex.EncodeToString(secret)
	}

	return &JWTManager{
		secret:       []byte(config.SecretKey),
		accessTTL:    config.AccessTTL,
		refreshTTL:   config.RefreshTTL,
		issuer:       config.Issuer,
		tokenStorage: storage,
	}
}

// GenerateTokenPair creates a new access and refresh token pair
func (j *JWTManager) GenerateTokenPair(userID string, username string, roles []string, permissions []string) (*TokenPair, error) {
	now := time.Now()
	sessionID := uuid.New().String()

	// Generate access token
	accessClaims := &Claims{
		UserID:      userID,
		Username:    username,
		Roles:       roles,
		Permissions: permissions,
		TokenType:   "access",
		SessionID:   sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	accessToken, err := j.generateToken(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: "refresh",
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	refreshToken, err := j.generateToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token if storage is available
	if j.tokenStorage != nil {
		tokenHash := j.hashToken(refreshToken)
		if err := j.tokenStorage.StoreRefreshToken(userID, tokenHash, refreshClaims.ExpiresAt.Time); err != nil {
			return nil, fmt.Errorf("failed to store refresh token: %w", err)
		}
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        int64(j.accessTTL.Seconds()),
		RefreshExpiresIn: int64(j.refreshTTL.Seconds()),
	}, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Additional validation for refresh tokens
	if claims.TokenType == "refresh" && j.tokenStorage != nil {
		tokenHash := j.hashToken(tokenString)
		valid, err := j.tokenStorage.ValidateRefreshToken(claims.UserID, tokenHash)
		if err != nil {
			return nil, fmt.Errorf("failed to validate refresh token: %w", err)
		}
		if !valid {
			return nil, ErrInvalidToken
		}
	}

	return claims, nil
}

// RefreshTokens generates new token pair using a valid refresh token
func (j *JWTManager) RefreshTokens(refreshTokenString string, roles []string, permissions []string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}

	// Revoke old refresh token
	if j.tokenStorage != nil {
		tokenHash := j.hashToken(refreshTokenString)
		if err := j.tokenStorage.RevokeRefreshToken(claims.UserID, tokenHash); err != nil {
			return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
		}
	}

	// Generate new token pair
	return j.GenerateTokenPair(claims.UserID, claims.Username, roles, permissions)
}

// RevokeToken revokes a specific token
func (j *JWTManager) RevokeToken(tokenString string) error {
	if j.tokenStorage == nil {
		return errors.New("token storage not configured")
	}

	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("failed to validate token for revocation: %w", err)
	}

	if claims.TokenType == "refresh" {
		tokenHash := j.hashToken(tokenString)
		return j.tokenStorage.RevokeRefreshToken(claims.UserID, tokenHash)
	}

	// For access tokens, we could implement a blacklist if needed
	return errors.New("access token revocation not implemented")
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (j *JWTManager) RevokeAllUserTokens(userID string) error {
	if j.tokenStorage == nil {
		return errors.New("token storage not configured")
	}

	return j.tokenStorage.RevokeAllUserTokens(userID)
}

// generateToken creates a JWT token from claims
func (j *JWTManager) generateToken(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// hashToken creates a hash of the token for storage (for security)
func (j *JWTManager) hashToken(token string) string {
	// Simple hash for demonstration - in production, use proper hashing
	return fmt.Sprintf("%x", token[len(token)-32:])
}

// TokenInfo provides information about a token without validating it
type TokenInfo struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	Expired   bool      `json:"expired"`
}

// GetTokenInfo extracts information from a token without validating it
func (j *JWTManager) GetTokenInfo(tokenString string) (*TokenInfo, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims structure")
	}

	return &TokenInfo{
		UserID:    claims.UserID,
		Username:  claims.Username,
		TokenType: claims.TokenType,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
		Expired:   time.Now().After(claims.ExpiresAt.Time),
	}, nil
}
