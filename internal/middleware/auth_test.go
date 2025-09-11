package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/dfedick/gotak/internal/auth"
	"github.com/dfedick/gotak/pkg/logger"
)

// MockAuthService implements auth.AuthService for testing
type MockAuthService struct {
	validateTokenFunc func(token string) (*auth.Claims, error)
}

func (m *MockAuthService) ValidateToken(token string) (*auth.Claims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(token)
	}
	return nil, auth.ErrInvalidToken
}

func setupTestMiddleware() (*AuthMiddleware, *MockAuthService) {
	loggerConfig := logger.Config{
		Level:   "error", // Reduce noise in tests
		Format:  "console",
		Output:  "stdout",
		Service: "middleware-test",
	}
	logger.Initialize(loggerConfig)
	log := logger.GetGlobalLogger()
	
	mockAuth := &MockAuthService{}
	middleware := NewAuthMiddleware(mockAuth, log)
	
	return middleware, mockAuth
}

func createTestToken() *auth.Claims {
	return &auth.Claims{
		UserID:      "user-123",
		Username:    "testuser",
		Roles:       []string{"operator", "mission_commander"},
		Permissions: []string{"missions.read", "cot.send", "reports.read"},
		TokenType:   "access",
		SessionID:   "session-123",
	}
}

func TestAuthMiddleware_Authenticate_Success(t *testing.T) {
	middleware, mockAuth := setupTestMiddleware()
	
	// Mock successful token validation
	testClaims := createTestToken()
	mockAuth.validateTokenFunc = func(token string) (*auth.Claims, error) {
		if token == "valid-token" {
			return testClaims, nil
		}
		return nil, auth.ErrInvalidToken
	}
	
	// Create test handler
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		require.NotNil(t, user)
		
		assert.Equal(t, testClaims.UserID, user.UserID)
		assert.Equal(t, testClaims.Username, user.Username)
		assert.Equal(t, testClaims.Roles, user.Roles)
		assert.Equal(t, testClaims.Permissions, user.Permissions)
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authenticated"))
	}))
	
	// Create request with valid token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "authenticated", rr.Body.String())
}

func TestAuthMiddleware_Authenticate_MissingToken(t *testing.T) {
	middleware, _ := setupTestMiddleware()
	
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for missing token")
	}))
	
	req := httptest.NewRequest("GET", "/protected", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Authentication required")
}

func TestAuthMiddleware_Authenticate_InvalidToken(t *testing.T) {
	middleware, mockAuth := setupTestMiddleware()
	
	// Mock token validation failure
	mockAuth.validateTokenFunc = func(token string) (*auth.Claims, error) {
		return nil, auth.ErrInvalidToken
	}
	
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for invalid token")
	}))
	
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid token")
}

func TestAuthMiddleware_Authenticate_ExpiredToken(t *testing.T) {
	middleware, mockAuth := setupTestMiddleware()
	
	// Mock expired token
	mockAuth.validateTokenFunc = func(token string) (*auth.Claims, error) {
		return nil, auth.ErrTokenExpired
	}
	
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for expired token")
	}))
	
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Token expired")
}

func TestAuthMiddleware_Authenticate_RefreshToken(t *testing.T) {
	middleware, mockAuth := setupTestMiddleware()
	
	// Mock refresh token (should not be accepted for authentication)
	mockAuth.validateTokenFunc = func(token string) (*auth.Claims, error) {
		return &auth.Claims{
			UserID:    "user-123",
			Username:  "testuser",
			TokenType: "refresh", // This should be rejected
		}, nil
	}
	
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for refresh token")
	}))
	
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer refresh-token")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid token type")
}

func TestAuthMiddleware_SkipPaths(t *testing.T) {
	middleware, _ := setupTestMiddleware()
	
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("public"))
	}))
	
	// Test skip paths
	skipPaths := []string{
		"/health",
		"/metrics",
		"/v1/auth/login",
		"/v1/auth/register",
		"/docs/api",
	}
	
	for _, path := range skipPaths {
		req := httptest.NewRequest("GET", path, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusOK, rr.Code, "Path %s should be skipped", path)
		assert.Equal(t, "public", rr.Body.String())
	}
}

func TestAuthMiddleware_RequireRole_Success(t *testing.T) {
	middleware, mockAuth := setupTestMiddleware()
	
	testClaims := createTestToken()
	mockAuth.validateTokenFunc = func(token string) (*auth.Claims, error) {
		return testClaims, nil
	}
	
	// Create handler chain: auth -> role check -> final handler
	handler := middleware.Authenticate(
		middleware.RequireRole("operator")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("authorized"))
			}),
		),
	)
	
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "authorized", rr.Body.String())
}

func TestAuthMiddleware_RequireRole_Failure(t *testing.T) {
	middleware, mockAuth := setupTestMiddleware()
	
	testClaims := createTestToken()
	mockAuth.validateTokenFunc = func(token string) (*auth.Claims, error) {
		return testClaims, nil
	}
	
	// Create handler chain requiring system_admin role (user doesn't have it)
	handler := middleware.Authenticate(
		middleware.RequireRole("system_admin")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Handler should not be called - insufficient role")
			}),
		),
	)
	
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Insufficient permissions")
}

func TestAuthMiddleware_RequirePermission_Success(t *testing.T) {
	middleware, mockAuth := setupTestMiddleware()
	
	testClaims := createTestToken()
	mockAuth.validateTokenFunc = func(token string) (*auth.Claims, error) {
		return testClaims, nil
	}
	
	// Test permission that user has
	handler := middleware.Authenticate(
		middleware.RequirePermission("missions.read")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("authorized"))
			}),
		),
	)
	
	req := httptest.NewRequest("GET", "/missions", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "authorized", rr.Body.String())
}

func TestAuthMiddleware_RequirePermission_Failure(t *testing.T) {
	middleware, mockAuth := setupTestMiddleware()
	
	testClaims := createTestToken()
	mockAuth.validateTokenFunc = func(token string) (*auth.Claims, error) {
		return testClaims, nil
	}
	
	// Test permission that user doesn't have
	handler := middleware.Authenticate(
		middleware.RequirePermission("system.admin")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Handler should not be called - insufficient permission")
			}),
		),
	)
	
	req := httptest.NewRequest("GET", "/admin/system", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Insufficient permissions")
}

func TestAuthMiddleware_RequireAnyPermission_Success(t *testing.T) {
	middleware, mockAuth := setupTestMiddleware()
	
	testClaims := createTestToken()
	mockAuth.validateTokenFunc = func(token string) (*auth.Claims, error) {
		return testClaims, nil
	}
	
	// Test multiple permissions, user has at least one
	handler := middleware.Authenticate(
		middleware.RequireAnyPermission("system.admin", "missions.read")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("authorized"))
			}),
		),
	)
	
	req := httptest.NewRequest("GET", "/missions", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "authorized", rr.Body.String())
}

func TestAuthMiddleware_OptionalAuth_WithToken(t *testing.T) {
	middleware, mockAuth := setupTestMiddleware()
	
	testClaims := createTestToken()
	mockAuth.validateTokenFunc = func(token string) (*auth.Claims, error) {
		return testClaims, nil
	}
	
	handler := middleware.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("anonymous"))
		}
	}))
	
	req := httptest.NewRequest("GET", "/public", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "authenticated", rr.Body.String())
}

func TestAuthMiddleware_OptionalAuth_WithoutToken(t *testing.T) {
	middleware, _ := setupTestMiddleware()
	
	handler := middleware.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("anonymous"))
		}
	}))
	
	req := httptest.NewRequest("GET", "/public", nil)
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "anonymous", rr.Body.String())
}

func TestContextHelpers(t *testing.T) {
	testClaims := createTestToken()
	
	// Create context with user data
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user", &UserContext{
		UserID:      testClaims.UserID,
		Username:    testClaims.Username,
		Roles:       testClaims.Roles,
		Permissions: testClaims.Permissions,
	})
	ctx = context.WithValue(ctx, "user_id", testClaims.UserID)
	ctx = context.WithValue(ctx, "username", testClaims.Username)
	ctx = context.WithValue(ctx, "roles", testClaims.Roles)
	ctx = context.WithValue(ctx, "permissions", testClaims.Permissions)
	
	// Test helper functions
	assert.Equal(t, testClaims.UserID, GetUserIDFromContext(ctx))
	assert.Equal(t, testClaims.Username, GetUsernameFromContext(ctx))
	assert.Equal(t, testClaims.Roles, GetUserRolesFromContext(ctx))
	assert.Equal(t, testClaims.Permissions, GetUserPermissionsFromContext(ctx))
	
	// Test role and permission checks
	assert.True(t, HasRole(ctx, "operator"))
	assert.False(t, HasRole(ctx, "system_admin"))
	assert.True(t, HasAnyRole(ctx, "system_admin", "operator"))
	
	assert.True(t, HasPermission(ctx, "missions.read"))
	assert.False(t, HasPermission(ctx, "system.admin"))
	assert.True(t, HasAnyPermission(ctx, "system.admin", "missions.read"))
	
	user := GetUserFromContext(ctx)
	require.NotNil(t, user)
	assert.Equal(t, testClaims.UserID, user.UserID)
}

func TestAuthMiddleware_InvalidAuthHeader(t *testing.T) {
	middleware, _ := setupTestMiddleware()
	
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for invalid header")
	}))
	
	// Test invalid header formats
	invalidHeaders := []string{
		"invalid-format",
		"Basic dGVzdDp0ZXN0", // Basic auth instead of Bearer
		"Bearer", // Missing token
		"", // Empty header
	}
	
	for _, header := range invalidHeaders {
		req := httptest.NewRequest("GET", "/protected", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusUnauthorized, rr.Code, "Invalid header should be rejected: %s", header)
	}
}
