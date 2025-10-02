package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/dfedick/gotak/pkg/logger"
)

// contextKey is a type for context keys
type contextKey string

const (
	// UserContextKey is the key for storing user in context
	UserContextKey contextKey = "user"
	// UserIDContextKey is the key for storing user ID in context
	UserIDContextKey contextKey = "user_id"
	// ClaimsContextKey is the key for storing JWT claims in context
	ClaimsContextKey contextKey = "claims"
)

// Middleware provides authentication middleware
type Middleware struct {
	jwtManager *JWTManager
	logger     *logger.Logger
	// List of paths that don't require authentication
	publicPaths []string
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware(jwtManager *JWTManager, logger *logger.Logger) *Middleware {
	return &Middleware{
		jwtManager: jwtManager,
		logger:     logger,
		publicPaths: []string{
			"/api/v1/auth/register",
			"/api/v1/auth/login",
			"/api/v1/auth/refresh",
			"/api/v1/auth/forgot-password",
			"/api/v1/auth/reset-password",
			"/health",
			"/ws/tactical", // WebSocket endpoint might have its own auth
		},
	}
}

// Authenticate is the main authentication middleware
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path is public
		if m.isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		token, err := m.extractToken(r)
		if err != nil {
			m.logger.Debug().
				Err(err).
				Str("path", r.URL.Path).
				Msg("Failed to extract token")
			m.sendUnauthorized(w, "Missing or invalid authorization header")
			return
		}

		// Validate token
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			if err == ErrTokenExpired {
				m.sendUnauthorized(w, "Token has expired")
				return
			}
			m.logger.Debug().
				Err(err).
				Str("path", r.URL.Path).
				Msg("Token validation failed")
			m.sendUnauthorized(w, "Invalid token")
			return
		}

		// Check token type - only access tokens are allowed for API calls
		if claims.TokenType != "access" {
			m.sendUnauthorized(w, "Invalid token type")
			return
		}

		// Parse user ID from claims
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			m.logger.Error().
				Err(err).
				Str("user_id", claims.UserID).
				Msg("Failed to parse user ID from token")
			m.sendUnauthorized(w, "Invalid user ID in token")
			return
		}

		// Add user information to context
		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
		ctx = context.WithValue(ctx, ClaimsContextKey, claims)

		// Log the authenticated request
		m.logger.Debug().
			Str("user_id", userID.String()).
			Str("username", claims.Username).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Authenticated request")

		// Continue with the request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole creates a middleware that requires specific roles
func (m *Middleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context
			claims, ok := r.Context().Value(ClaimsContextKey).(*Claims)
			if !ok {
				m.sendForbidden(w, "No authentication claims found")
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range claims.Roles {
					if userRole == requiredRole || userRole == "*" {
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
					Str("user_id", claims.UserID).
					Strs("user_roles", claims.Roles).
					Strs("required_roles", roles).
					Msg("Access denied - insufficient role")
				m.sendForbidden(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission creates a middleware that requires specific permissions
func (m *Middleware) RequirePermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context
			claims, ok := r.Context().Value(ClaimsContextKey).(*Claims)
			if !ok {
				m.sendForbidden(w, "No authentication claims found")
				return
			}

			// Check if user has any of the required permissions
			hasPermission := false
			for _, requiredPerm := range permissions {
				for _, userPerm := range claims.Permissions {
					if userPerm == requiredPerm || userPerm == "*" {
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
					Str("user_id", claims.UserID).
					Strs("user_permissions", claims.Permissions).
					Strs("required_permissions", permissions).
					Msg("Access denied - insufficient permission")
				m.sendForbidden(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper methods

func (m *Middleware) extractToken(r *http.Request) (string, error) {
	// Get Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("unauthorized")
	}

	// Check for Bearer token
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", ErrInvalidToken
	}

	// Extract token
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", ErrInvalidToken
	}

	return token, nil
}

func (m *Middleware) isPublicPath(path string) bool {
	for _, publicPath := range m.publicPaths {
		if path == publicPath || strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	// Also allow static files and root
	if path == "/" || strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/assets/") {
		return true
	}
	return false
}

func (m *Middleware) sendUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

func (m *Middleware) sendForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uuid.UUID)
	return userID, ok
}

// GetClaimsFromContext retrieves the JWT claims from the request context
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}