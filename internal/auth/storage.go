package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/dfedick/gotak/pkg/logger"
)

// PostgreSQLTokenStorage implements TokenStorage interface using PostgreSQL
type PostgreSQLTokenStorage struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewPostgreSQLTokenStorage creates a new PostgreSQL token storage
func NewPostgreSQLTokenStorage(db *sql.DB, logger *logger.Logger) *PostgreSQLTokenStorage {
	return &PostgreSQLTokenStorage{
		db:     db,
		logger: logger,
	}
}

// StoreRefreshToken stores a refresh token in the database
func (s *PostgreSQLTokenStorage) StoreRefreshToken(userID string, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO gotak.refresh_tokens (user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	
	// Hash the token for secure storage
	hashedToken := s.hashToken(tokenHash)
	
	_, err := s.db.Exec(query, userID, hashedToken, expiresAt)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("user_id", userID).
			Msg("Failed to store refresh token")
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	s.logger.Debug().
		Str("user_id", userID).
		Time("expires_at", expiresAt).
		Msg("Stored refresh token")

	return nil
}

// ValidateRefreshToken checks if a refresh token is valid and not revoked
func (s *PostgreSQLTokenStorage) ValidateRefreshToken(userID string, tokenHash string) (bool, error) {
	query := `
		SELECT expires_at, revoked_at
		FROM gotak.refresh_tokens
		WHERE user_id = $1 AND token_hash = $2
	`
	
	hashedToken := s.hashToken(tokenHash)
	
	var expiresAt time.Time
	var revokedAt sql.NullTime
	
	err := s.db.QueryRow(query, userID, hashedToken).Scan(&expiresAt, &revokedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn().
				Str("user_id", userID).
				Msg("Refresh token not found")
			return false, nil
		}
		s.logger.Error().
			Err(err).
			Str("user_id", userID).
			Msg("Failed to validate refresh token")
		return false, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Check if token is revoked
	if revokedAt.Valid {
		s.logger.Debug().
			Str("user_id", userID).
			Time("revoked_at", revokedAt.Time).
			Msg("Refresh token is revoked")
		return false, nil
	}

	// Check if token is expired
	if time.Now().After(expiresAt) {
		s.logger.Debug().
			Str("user_id", userID).
			Time("expires_at", expiresAt).
			Msg("Refresh token is expired")
		return false, nil
	}

	// Update last used timestamp
	updateQuery := `
		UPDATE gotak.refresh_tokens
		SET last_used_at = NOW()
		WHERE user_id = $1 AND token_hash = $2
	`
	
	if _, err := s.db.Exec(updateQuery, userID, hashedToken); err != nil {
		s.logger.Warn().
			Err(err).
			Str("user_id", userID).
			Msg("Failed to update token last used timestamp")
		// Don't fail validation for this error
	}

	s.logger.Debug().
		Str("user_id", userID).
		Msg("Refresh token is valid")

	return true, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (s *PostgreSQLTokenStorage) RevokeRefreshToken(userID string, tokenHash string) error {
	query := `
		UPDATE gotak.refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND token_hash = $2 AND revoked_at IS NULL
	`
	
	hashedToken := s.hashToken(tokenHash)
	
	result, err := s.db.Exec(query, userID, hashedToken)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("user_id", userID).
			Msg("Failed to revoke refresh token")
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("user_id", userID).
			Msg("Could not check rows affected when revoking token")
	} else if rowsAffected == 0 {
		s.logger.Warn().
			Str("user_id", userID).
			Msg("No refresh token found to revoke")
	}

	s.logger.Debug().
		Str("user_id", userID).
		Int64("rows_affected", rowsAffected).
		Msg("Revoked refresh token")

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *PostgreSQLTokenStorage) RevokeAllUserTokens(userID string) error {
	query := `
		UPDATE gotak.refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	
	result, err := s.db.Exec(query, userID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("user_id", userID).
			Msg("Failed to revoke all user refresh tokens")
		return fmt.Errorf("failed to revoke all user refresh tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("user_id", userID).
			Msg("Could not check rows affected when revoking all tokens")
		rowsAffected = 0
	}

	s.logger.Info().
		Str("user_id", userID).
		Int64("tokens_revoked", rowsAffected).
		Msg("Revoked all user refresh tokens")

	return nil
}

// CleanExpiredTokens removes expired refresh tokens from the database
func (s *PostgreSQLTokenStorage) CleanExpiredTokens() error {
	query := `
		DELETE FROM gotak.refresh_tokens
		WHERE expires_at < NOW()
	`
	
	result, err := s.db.Exec(query)
	if err != nil {
		s.logger.Error().
			Err(err).
			Msg("Failed to clean expired refresh tokens")
		return fmt.Errorf("failed to clean expired refresh tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Warn().
			Err(err).
			Msg("Could not check rows affected when cleaning expired tokens")
		rowsAffected = 0
	}

	if rowsAffected > 0 {
		s.logger.Info().
			Int64("tokens_cleaned", rowsAffected).
			Msg("Cleaned expired refresh tokens")
	}

	return nil
}

// GetUserTokensInfo returns information about a user's tokens
func (s *PostgreSQLTokenStorage) GetUserTokensInfo(userID string) ([]TokenStorageInfo, error) {
	query := `
		SELECT id, token_hash, expires_at, created_at, last_used_at, revoked_at, device_info, ip_address
		FROM gotak.refresh_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := s.db.Query(query, userID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("user_id", userID).
			Msg("Failed to get user tokens info")
		return nil, fmt.Errorf("failed to get user tokens info: %w", err)
	}
	defer rows.Close()

	var tokens []TokenStorageInfo
	for rows.Next() {
		var token TokenStorageInfo
		var deviceInfo sql.NullString
		var ipAddress sql.NullString
		
		err := rows.Scan(
			&token.ID,
			&token.TokenHash,
			&token.ExpiresAt,
			&token.CreatedAt,
			&token.LastUsedAt,
			&token.RevokedAt,
			&deviceInfo,
			&ipAddress,
		)
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("user_id", userID).
				Msg("Failed to scan token info")
			continue
		}
		
		if deviceInfo.Valid {
			token.DeviceInfo = &deviceInfo.String
		}
		if ipAddress.Valid {
			token.IPAddress = &ipAddress.String
		}
		
		token.IsExpired = time.Now().After(token.ExpiresAt)
		token.IsRevoked = token.RevokedAt != nil
		
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// hashToken creates a SHA-256 hash of the token for secure storage
func (s *PostgreSQLTokenStorage) hashToken(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return hex.EncodeToString(hasher.Sum(nil))
}

// TokenStorageInfo represents stored token information
type TokenStorageInfo struct {
	ID          string     `json:"id"`
	TokenHash   string     `json:"-"` // Never expose hash
	ExpiresAt   time.Time  `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	RevokedAt   *time.Time `json:"revoked_at"`
	DeviceInfo  *string    `json:"device_info"`
	IPAddress   *string    `json:"ip_address"`
	IsExpired   bool       `json:"is_expired"`
	IsRevoked   bool       `json:"is_revoked"`
}
