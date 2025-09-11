package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dfedick/gotak/internal/auth"
	"github.com/dfedick/gotak/pkg/logger"
)

// AuthService interface for authentication operations
type AuthService interface {
	ValidateToken(token string) (*auth.Claims, error)
}

// AuthMiddleware handles JWT authentication and authorization
type AuthMiddleware struct {
	authService AuthService
	logger      *logger.Logger
	skipPaths   []string
}

// UserContext represents the authenticated user context
type UserContext struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	SessionID   string   `json:"session_id"`
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService AuthService, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
		skipPaths: []string{
			"/health",
			"/metrics", 
			"/v1/auth/login",
			"/v1/auth/register",
			"/v1/auth/refresh",
			"/docs/",
			"/swagger/",
			"/favicon.ico",
		},
	}
}

// Authenticate middleware validates JWT tokens and sets user context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for certain paths
		if m.shouldSkipAuth(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		token, err := m.extractTokenFromHeader(r)
		if err != nil {
			m.logger.Debug().
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Str("ip", getClientIP(r)).
				Msg("Missing or invalid authorization header")
			m.sendErrorResponse(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Validate JWT token
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			m.logger.Warn().
				Err(err).
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Str("ip", getClientIP(r)).
				Str("user_agent", r.UserAgent()).
				Msg("Invalid or expired token")
			
			if err == auth.ErrTokenExpired {
				m.sendErrorResponse(w, "Token expired", http.StatusUnauthorized)
			} else {
				m.sendErrorResponse(w, "Invalid token", http.StatusUnauthorized)
			}
			return
		}

		// Ensure this is an access token
		if claims.TokenType != "access" {
			m.logger.Warn().
				Str("token_type", claims.TokenType).
				Str("user_id", claims.UserID).
				Str("path", r.URL.Path).
				Msg("Invalid token type for authentication")
			m.sendErrorResponse(w, "Invalid token type", http.StatusUnauthorized)
			return
		}

		// Create user context
		userCtx := &UserContext{
			UserID:      claims.UserID,
			Username:    claims.Username,
			Roles:       claims.Roles,
			Permissions: claims.Permissions,
			SessionID:   claims.SessionID,
		}

		// Add user context to request context
		ctx := context.WithValue(r.Context(), "user", userCtx)
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "roles", claims.Roles)
		ctx = context.WithValue(ctx, "permissions", claims.Permissions)

		m.logger.Debug().
			Str("user_id", claims.UserID).
			Str("username", claims.Username).
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Msg("User authenticated successfully")

		// Continue to next handler with user context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware ensures the user has one of the specified roles
func (m *AuthMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				m.logger.Error().
					Str("path", r.URL.Path).
					Msg("User context not found in RequireRole middleware")
				m.sendErrorResponse(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range user.Roles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				m.logger.Warn().
					Str("user_id", user.UserID).
					Str("username", user.Username).
					Strs("user_roles", user.Roles).
					Strs("required_roles", roles).
					Str("path", r.URL.Path).
					Msg("Access denied - insufficient role")
				m.sendErrorResponse(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			m.logger.Debug().
				Str("user_id", user.UserID).
				Str("username", user.Username).
				Strs("required_roles", roles).
				Str("path", r.URL.Path).
				Msg("Role authorization successful")

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission middleware ensures the user has the specified permission
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				m.logger.Error().
					Str("path", r.URL.Path).
					Msg("User context not found in RequirePermission middleware")
				m.sendErrorResponse(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user has the required permission
			hasPermission := false
			for _, userPerm := range user.Permissions {
				if userPerm == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				m.logger.Warn().
					Str("user_id", user.UserID).
					Str("username", user.Username).
					Strs("user_permissions", user.Permissions).
					Str("required_permission", permission).
					Str("path", r.URL.Path).
					Msg("Access denied - insufficient permission")
				m.sendErrorResponse(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			m.logger.Debug().
				Str("user_id", user.UserID).
				Str("username", user.Username).
				Str("required_permission", permission).
				Str("path", r.URL.Path).
				Msg("Permission authorization successful")

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission middleware ensures the user has at least one of the specified permissions
func (m *AuthMiddleware) RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				m.logger.Error().
					Str("path", r.URL.Path).
					Msg("User context not found in RequireAnyPermission middleware")
				m.sendErrorResponse(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user has any of the required permissions
			hasPermission := false
			for _, requiredPerm := range permissions {
				for _, userPerm := range user.Permissions {
					if userPerm == requiredPerm {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}

			if !hasPermission {
				m.logger.Warn().
					Str("user_id", user.UserID).
					Str("username", user.Username).
					Strs("user_permissions", user.Permissions).
					Strs("required_permissions", permissions).
					Str("path", r.URL.Path).
					Msg("Access denied - insufficient permissions")
				m.sendErrorResponse(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			m.logger.Debug().
				Str("user_id", user.UserID).
				Str("username", user.Username).
				Strs("required_permissions", permissions).
				Str("path", r.URL.Path).
				Msg("Permission authorization successful")

			next.ServeHTTP(w, r)
		})
	}
}

// AdminOnly middleware ensures only system admins can access the resource
func (m *AuthMiddleware) AdminOnly() func(http.Handler) http.Handler {
	return m.RequireRole("system_admin")
}

// OptionalAuth middleware extracts user context if present but doesn't require authentication
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to extract token
		token, err := m.extractTokenFromHeader(r)
		if err != nil {
			// No token present, continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Validate token if present
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			// Invalid token, log but continue without user context
			m.logger.Debug().
				Err(err).
				Str("path", r.URL.Path).
				Msg("Invalid token in optional auth, continuing without user context")
			next.ServeHTTP(w, r)
			return
		}

		// Valid token found, add user context
		if claims.TokenType == "access" {
			userCtx := &UserContext{
				UserID:      claims.UserID,
				Username:    claims.Username,
				Roles:       claims.Roles,
				Permissions: claims.Permissions,
				SessionID:   claims.SessionID,
			}

			ctx := context.WithValue(r.Context(), "user", userCtx)
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "roles", claims.Roles)
			ctx = context.WithValue(ctx, "permissions", claims.Permissions)

			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// shouldSkipAuth checks if the path should skip authentication
func (m *AuthMiddleware) shouldSkipAuth(path string) bool {
	for _, skipPath := range m.skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// extractTokenFromHeader extracts JWT token from Authorization header
func (m *AuthMiddleware) extractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", auth.ErrInvalidToken
	}

	// Check for Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", auth.ErrInvalidToken
	}

	return parts[1], nil
}

// sendErrorResponse sends a JSON error response
func (m *AuthMiddleware) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    statusCode,
	}
	
	json.NewEncoder(w).Encode(response)
}

// GetUserFromContext extracts the user context from the request context
func GetUserFromContext(ctx context.Context) *UserContext {
	user, ok := ctx.Value("user").(*UserContext)
	if !ok {
		return nil
	}
	return user
}

// GetUserIDFromContext extracts the user ID from the request context
func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUsernameFromContext extracts the username from the request context
func GetUsernameFromContext(ctx context.Context) string {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return ""
	}
	return username
}

// GetUserRolesFromContext extracts the user roles from the request context
func GetUserRolesFromContext(ctx context.Context) []string {
	roles, ok := ctx.Value("roles").([]string)
	if !ok {
		return nil
	}
	return roles
}

// GetUserPermissionsFromContext extracts the user permissions from the request context
func GetUserPermissionsFromContext(ctx context.Context) []string {
	permissions, ok := ctx.Value("permissions").([]string)
	if !ok {
		return nil
	}
	return permissions
}

// HasRole checks if the user has a specific role
func HasRole(ctx context.Context, role string) bool {
	roles := GetUserRolesFromContext(ctx)
	for _, userRole := range roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// HasPermission checks if the user has a specific permission
func HasPermission(ctx context.Context, permission string) bool {
	permissions := GetUserPermissionsFromContext(ctx)
	for _, userPerm := range permissions {
		if userPerm == permission {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles
func HasAnyRole(ctx context.Context, roles ...string) bool {
	userRoles := GetUserRolesFromContext(ctx)
	for _, requiredRole := range roles {
		for _, userRole := range userRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

// HasAnyPermission checks if the user has any of the specified permissions
func HasAnyPermission(ctx context.Context, permissions ...string) bool {
	userPerms := GetUserPermissionsFromContext(ctx)
	for _, requiredPerm := range permissions {
		for _, userPerm := range userPerms {
			if userPerm == requiredPerm {
				return true
			}
		}
	}
	return false
}

// getClientIP extracts the real client IP address
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to remote address
	parts := strings.Split(r.RemoteAddr, ":")
	if len(parts) >= 2 {
		return strings.Join(parts[:len(parts)-1], ":")
	}
	
	return r.RemoteAddr
}
