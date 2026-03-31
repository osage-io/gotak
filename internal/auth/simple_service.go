package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dfedick/gotak/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// SimpleAuthService handles authentication with the existing database schema
type SimpleAuthService struct {
	db         *sqlx.DB
	jwtManager *JWTManager
	logger     *logger.Logger
}

// NewSimpleAuthService creates a new simplified auth service
func NewSimpleAuthService(db *sqlx.DB, jwtConfig JWTConfig, logger *logger.Logger) *SimpleAuthService {
	tokenStorage := NewInMemoryTokenStorage()
	jwtManager := NewJWTManager(jwtConfig, tokenStorage)

	return &SimpleAuthService{
		db:         db,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// Register creates a new user
func (s *SimpleAuthService) Register(ctx context.Context, req *SimpleRegisterRequest) (*SimpleUser, error) {
	// Check if user exists
	var count int
	err := s.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2",
		req.Username, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if count > 0 {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &SimpleUser{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "user", // Default role
		Active:       true,
		IsVerified:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Insert user
	query := `
		INSERT INTO users (id, username, email, password_hash, role, is_active, is_verified, created_at, updated_at)
		VALUES (:id, :username, :email, :password_hash, :role, :is_active, :is_verified, :created_at, :updated_at)
	`

	_, err = s.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("username", user.Username).
		Msg("User registered successfully")

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *SimpleAuthService) Login(ctx context.Context, req *SimpleLoginRequest) (*TokenPair, error) {
	// Get user
	var user SimpleUser
	err := s.db.GetContext(ctx, &user,
		"SELECT * FROM users WHERE username = $1",
		req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn().
				Str("username", req.Username).
				Msg("Login attempt for non-existent user")
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.Active {
		s.logger.Warn().
			Str("username", req.Username).
			Msg("Login attempt for inactive user")
		return nil, ErrUserNotActive
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		s.logger.Warn().
			Str("username", req.Username).
			Msg("Invalid password attempt")
		return nil, ErrInvalidCredentials
	}

	// Update last login
	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		"UPDATE users SET last_login_at = $1 WHERE id = $2",
		now, user.ID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to update last login")
		// Don't fail the login for this
	}

	// Generate tokens
	tokenPair, err := s.jwtManager.GenerateTokenPair(
		user.ID.String(),
		user.Username,
		[]string{user.Role}, // Use single role as array
		[]string{},          // No permissions in simple schema
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("username", user.Username).
		Msg("User logged in successfully")

	return tokenPair, nil
}

// GetUserByID retrieves a user by ID
func (s *SimpleAuthService) GetUserByID(ctx context.Context, userID string) (*SimpleUser, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user SimpleUser
	err = s.db.GetContext(ctx, &user,
		"SELECT * FROM users WHERE id = $1",
		id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (s *SimpleAuthService) GetUserByUsername(ctx context.Context, username string) (*SimpleUser, error) {
	var user SimpleUser
	err := s.db.GetContext(ctx, &user,
		"SELECT * FROM users WHERE username = $1",
		username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// ValidateToken validates a JWT token
func (s *SimpleAuthService) ValidateToken(tokenString string) (*Claims, error) {
	return s.jwtManager.ValidateToken(tokenString)
}

// RefreshToken refreshes an access token using a refresh token
func (s *SimpleAuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}

	// Get user to get updated role
	user, err := s.GetUserByID(context.Background(), claims.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return s.jwtManager.RefreshTokens(
		refreshToken,
		[]string{user.Role},
		[]string{},
	)
}

// ChangePassword changes a user's password
func (s *SimpleAuthService) ChangePassword(ctx context.Context, userID string, currentPassword, newPassword string) error {
	// Get user
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	_, err = s.db.ExecContext(ctx,
		"UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3",
		string(hashedPassword), time.Now(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID).
		Msg("Password changed successfully")

	return nil
}
