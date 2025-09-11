// +build integration

package auth

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"

	"github.com/dfedick/gotak/pkg/logger"
)

// This is an integration test that requires a running PostgreSQL database
// Run with: go test -tags=integration ./internal/auth

func setupTestDB(t *testing.T) *sql.DB {
	// Use test database
	dbURL := "postgres://gotak:dev_password@localhost:5432/gotak_dev?sslmode=disable"
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	
	err = db.Ping()
	require.NoError(t, err, "Failed to connect to test database")
	
	return db
}

func setupTestAuthService(t *testing.T) (*AuthService, *sql.DB) {
	db := setupTestDB(t)
	
	// Initialize logger
	loggerConfig := logger.Config{
		Level:      "debug",
		Format:     "console",
		Output:     "stdout",
		Service:    "auth-test",
		Version:    "1.0.0",
		TimeFormat: "rfc3339",
	}
	logger.Initialize(loggerConfig)
	log := logger.GetGlobalLogger()

	config := AuthConfig{
		MinPasswordLength: 8,
		MaxPasswordLength: 128,
		MaxFailedAttempts: 3,
		LockoutDuration:   5 * time.Minute,
		JWT: JWTConfig{
			SecretKey:  "test-integration-secret-key",
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 24 * time.Hour,
			Issuer:     "gotak-test",
		},
	}

	authService, err := NewAuthService(db, config, log)
	require.NoError(t, err)
	
	return authService, db
}

func TestAuthService_Integration_CompleteFlow(t *testing.T) {
	authService, db := setupTestAuthService(t)
	defer db.Close()

	// Test user registration
	registerReq := &RegisterRequest{
		Username: "testuser_integration",
		Email:    "test_integration@gotak.dev",
		Password: "TestPassword123!",
		FirstName: stringPtr("Test"),
		LastName:  stringPtr("User"),
	}

	user, err := authService.RegisterUser(registerReq)
	require.NoError(t, err)
	require.NotNil(t, user)
	
	assert.Equal(t, registerReq.Username, user.Username)
	assert.Equal(t, registerReq.Email, user.Email)
	assert.True(t, user.IsActive)
	assert.NotEmpty(t, user.ID)

	// Test authentication
	loginReq := &LoginRequest{
		Username:   registerReq.Username,
		Password:   registerReq.Password,
		AuthMethod: "local",
	}

	tokenPair, err := authService.Authenticate(loginReq)
	require.NoError(t, err)
	require.NotNil(t, tokenPair)
	
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, "Bearer", tokenPair.TokenType)

	// Test token validation
	claims, err := authService.ValidateToken(tokenPair.AccessToken)
	require.NoError(t, err)
	require.NotNil(t, claims)
	
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, "access", claims.TokenType)

	// Test token refresh
	newTokenPair, err := authService.RefreshTokens(tokenPair.RefreshToken)
	require.NoError(t, err)
	require.NotNil(t, newTokenPair)
	
	assert.NotEmpty(t, newTokenPair.AccessToken)
	assert.NotEmpty(t, newTokenPair.RefreshToken)
	assert.NotEqual(t, tokenPair.AccessToken, newTokenPair.AccessToken)
	assert.NotEqual(t, tokenPair.RefreshToken, newTokenPair.RefreshToken)

	// Validate new access token
	newClaims, err := authService.ValidateToken(newTokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, newClaims.UserID)
	assert.Equal(t, user.Username, newClaims.Username)

	// Test token revocation
	err = authService.RevokeAllUserTokens(user.ID)
	require.NoError(t, err)

	// Old refresh token should no longer work
	_, err = authService.RefreshTokens(tokenPair.RefreshToken)
	assert.Error(t, err)

	// Clean up - delete test user
	_, err = db.Exec("DELETE FROM gotak.users WHERE id = $1", user.ID)
	require.NoError(t, err)
}

func TestAuthService_Integration_UserRolesAndPermissions(t *testing.T) {
	authService, db := setupTestAuthService(t)
	defer db.Close()

	// Get existing admin user (created during migration)
	user, err := authService.GetUserByUsername("admin")
	require.NoError(t, err)
	require.NotNil(t, user)

	// Update user with password for testing
	passwordHash, err := authService.hashPassword("AdminPassword123!")
	require.NoError(t, err)

	_, err = db.Exec("UPDATE gotak.users SET password_hash = $1 WHERE id = $2", passwordHash, user.ID)
	require.NoError(t, err)

	// Test authentication
	loginReq := &LoginRequest{
		Username:   "admin",
		Password:   "AdminPassword123!",
		AuthMethod: "local",
	}

	tokenPair, err := authService.Authenticate(loginReq)
	require.NoError(t, err)
	require.NotNil(t, tokenPair)

	// Validate token contains roles and permissions
	claims, err := authService.ValidateToken(tokenPair.AccessToken)
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, "admin", claims.Username)
	assert.Contains(t, claims.Roles, "system_admin")
	assert.NotEmpty(t, claims.Permissions)

	// Test getting user roles directly
	roles, err := authService.GetUserRoles(user.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, roles)

	// Find system_admin role
	var hasSystemAdmin bool
	for _, role := range roles {
		if role.Name == "system_admin" {
			hasSystemAdmin = true
			break
		}
	}
	assert.True(t, hasSystemAdmin, "Admin user should have system_admin role")

	// Test getting user permissions
	permissions, err := authService.GetUserPermissions(user.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, permissions)

	// System admin should have all permissions
	var hasSystemAdminPerm bool
	for _, perm := range permissions {
		if perm.Name == "system.admin" {
			hasSystemAdminPerm = true
			break
		}
	}
	assert.True(t, hasSystemAdminPerm, "Admin user should have system.admin permission")
}

func TestAuthService_Integration_FailedLoginAttempts(t *testing.T) {
	authService, db := setupTestAuthService(t)
	defer db.Close()

	// Register test user
	registerReq := &RegisterRequest{
		Username: "testuser_lockout",
		Email:    "lockout@gotak.dev",
		Password: "TestPassword123!",
	}

	user, err := authService.RegisterUser(registerReq)
	require.NoError(t, err)
	defer func() {
		db.Exec("DELETE FROM gotak.users WHERE id = $1", user.ID)
	}()

	// Attempt login with wrong password multiple times
	wrongLoginReq := &LoginRequest{
		Username:   registerReq.Username,
		Password:   "WrongPassword",
		AuthMethod: "local",
	}

	// First few attempts should fail but not lock
	for i := 0; i < 2; i++ {
		_, err := authService.Authenticate(wrongLoginReq)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err)
	}

	// Third attempt should cause lockout (config.MaxFailedAttempts = 3)
	_, err = authService.Authenticate(wrongLoginReq)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)

	// Fourth attempt should return account locked error
	_, err = authService.Authenticate(wrongLoginReq)
	assert.Error(t, err)
	assert.Equal(t, ErrAccountLocked, err)

	// Even correct password should fail when locked
	correctLoginReq := &LoginRequest{
		Username:   registerReq.Username,
		Password:   registerReq.Password,
		AuthMethod: "local",
	}

	_, err = authService.Authenticate(correctLoginReq)
	assert.Error(t, err)
	assert.Equal(t, ErrAccountLocked, err)
}

func TestAuthService_Integration_TokenStorage(t *testing.T) {
	authService, db := setupTestAuthService(t)
	defer db.Close()

	// Register and authenticate user
	registerReq := &RegisterRequest{
		Username: "testuser_storage",
		Email:    "storage@gotak.dev", 
		Password: "TestPassword123!",
	}

	user, err := authService.RegisterUser(registerReq)
	require.NoError(t, err)
	defer func() {
		db.Exec("DELETE FROM gotak.users WHERE id = $1", user.ID)
	}()

	loginReq := &LoginRequest{
		Username:   registerReq.Username,
		Password:   registerReq.Password,
		AuthMethod: "local",
	}

	_, err = authService.Authenticate(loginReq)
	require.NoError(t, err)

	// Verify refresh token is stored in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM gotak.refresh_tokens WHERE user_id = $1", user.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test token cleanup
	err = authService.tokenStorage.CleanExpiredTokens()
	require.NoError(t, err)

	// Should still have the token (not expired)
	err = db.QueryRow("SELECT COUNT(*) FROM gotak.refresh_tokens WHERE user_id = $1", user.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Get user tokens info
	tokensInfo, err := authService.tokenStorage.GetUserTokensInfo(user.ID)
	require.NoError(t, err)
	assert.Len(t, tokensInfo, 1)
	assert.False(t, tokensInfo[0].IsExpired)
	assert.False(t, tokensInfo[0].IsRevoked)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
