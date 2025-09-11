package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dfedick/gotak/pkg/logger"
)

// CORS middleware handles Cross-Origin Resource Sharing
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
					break
				}
			}
			
			if !allowed && len(allowedOrigins) > 0 {
				// If no origins match and we have a whitelist, deny
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
				}
			}
			
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "3600")
			
			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger logs HTTP requests with timing and status
func RequestLogger(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap response writer to capture status
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			// Process request
			next.ServeHTTP(wrapped, r)
			
			// Log request details
			duration := time.Since(start)
			userID := GetUserIDFromContext(r.Context())
			
			logEvent := logger.Info()
			if userID != "" {
				logEvent = logEvent.Str("user_id", userID)
			}
			
			logEvent.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("query", r.URL.RawQuery).
				Int("status", wrapped.statusCode).
				Dur("duration", duration).
				Str("ip", getClientIP(r)).
				Str("user_agent", r.UserAgent()).
				Int64("content_length", r.ContentLength).
				Msg("HTTP Request")
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// SecurityHeaders adds common security headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		
		// XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy (basic)
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'")
		
		// Strict Transport Security (HSTS) - only for HTTPS
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		
		next.ServeHTTP(w, r)
	})
}

// RateLimitInfo represents rate limiting configuration
type RateLimitInfo struct {
	Limit     int           // Requests per window
	Window    time.Duration // Time window
	KeyFunc   func(*http.Request) string // Function to generate rate limit key
}

// Simple in-memory rate limiter (for production use Redis or similar)
type rateLimiter struct {
	requests map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
	}
}

func (rl *rateLimiter) allow(key string, limit int, window time.Duration) bool {
	now := time.Now()
	
	// Clean old requests
	if requests, exists := rl.requests[key]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < window {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[key] = validRequests
	}
	
	// Check if limit exceeded
	if len(rl.requests[key]) >= limit {
		return false
	}
	
	// Add current request
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

var globalRateLimiter = newRateLimiter()

// RateLimit implements basic rate limiting
func RateLimit(info RateLimitInfo) func(http.Handler) http.Handler {
	if info.KeyFunc == nil {
		// Default to IP-based rate limiting
		info.KeyFunc = func(r *http.Request) string {
			return getClientIP(r)
		}
	}
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := info.KeyFunc(r)
			
			if !globalRateLimiter.allow(key, info.Limit, info.Window) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				
				response := map[string]interface{}{
					"error":   true,
					"message": "Rate limit exceeded",
					"code":    http.StatusTooManyRequests,
				}
				
				// Don't log error on encode failure for rate limit response
				_ = json.NewEncoder(w).Encode(response)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// JSONContentType ensures responses are JSON
func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set JSON content type for API responses
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v1/") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		
		next.ServeHTTP(w, r)
	})
}

// Recovery middleware handles panics gracefully
func Recovery(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error().
						Interface("panic", err).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Str("ip", getClientIP(r)).
						Msg("Request panic recovered")
					
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					
					response := map[string]interface{}{
						"error":   true,
						"message": "Internal server error",
						"code":    http.StatusInternalServerError,
					}
					
					// Don't log error on encode failure for panic recovery
					_ = json.NewEncoder(w).Encode(response)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// UserAgent extracts and validates user agent
func UserAgent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.UserAgent()
		
		// Block empty or suspicious user agents if desired
		if userAgent == "" {
			// Set a default user agent for monitoring
			r.Header.Set("User-Agent", "unknown")
		}
		
		next.ServeHTTP(w, r)
	})
}

// RequestID adds a unique request ID header
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID (in production, use UUID)
			requestID = generateSimpleID()
		}
		
		// Add to response header
		w.Header().Set("X-Request-ID", requestID)
		
		// Add to request context for logging
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateSimpleID generates a simple ID for request tracking
func generateSimpleID() string {
	// Simple timestamp-based ID (use UUID in production)
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
