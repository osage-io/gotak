package mission

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// ContextKey type for context keys
type ContextKey string

const (
	UserIDKey  ContextKey = "user_id"
	GroupIDKey ContextKey = "group_id"
)

// WithAuthContext is a middleware that extracts user info from JWT and adds to context
func WithAuthContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// For now, set default values for testing
			ctx := context.WithValue(r.Context(), UserIDKey, "129a0629-31ca-49a5-afde-a56db4f20487")
			ctx = context.WithValue(ctx, GroupIDKey, "default-group")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// For now, set default values
			ctx := context.WithValue(r.Context(), UserIDKey, "129a0629-31ca-49a5-afde-a56db4f20487")
			ctx = context.WithValue(ctx, GroupIDKey, "default-group")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		tokenString := parts[1]

		// Parse JWT token (simplified - in production should verify signature)
		token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// For testing, we're not verifying the signature
			return []byte("your-secret-key-change-in-production"), nil
		})

		if token != nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				// Extract user_id from claims
				if userID, ok := claims["user_id"].(string); ok {
					ctx := context.WithValue(r.Context(), UserIDKey, userID)
					// For now, use a default group
					ctx = context.WithValue(ctx, GroupIDKey, "default-group")
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// Fallback to defaults
		ctx := context.WithValue(r.Context(), UserIDKey, "129a0629-31ca-49a5-afde-a56db4f20487")
		ctx = context.WithValue(ctx, GroupIDKey, "default-group")
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// WrapHandler wraps a handler with auth context middleware
func WrapHandler(h http.HandlerFunc) http.HandlerFunc {
	return WithAuthContext(h)
}
