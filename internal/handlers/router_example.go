package handlers

import (
	"net/http"
	"time"

	"github.com/dfedick/gotak/internal/auth"
	"github.com/dfedick/gotak/internal/middleware"
	"github.com/dfedick/gotak/pkg/logger"
)

// SetupExampleRouter demonstrates how to set up routes with authentication middleware
func SetupExampleRouter(authService *auth.AuthService, logger *logger.Logger) http.Handler {
	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	
	// Create example handlers
	handlers := NewExampleHandlers(logger)
	
	// Create a new ServeMux
	mux := http.NewServeMux()
	
	// === PUBLIC ROUTES (no authentication required) ===
	mux.Handle("/health", 
		middleware.RequestLogger(logger)(
			middleware.SecurityHeaders(
				http.HandlerFunc(handlers.HealthCheckHandler),
			),
		),
	)
	
	mux.Handle("/public", 
		middleware.RequestLogger(logger)(
			middleware.SecurityHeaders(
				http.HandlerFunc(handlers.PublicHandler),
			),
		),
	)
	
	// === OPTIONAL AUTH ROUTES (work with or without authentication) ===
	mux.Handle("/optional", 
		middleware.RequestLogger(logger)(
			middleware.SecurityHeaders(
				authMiddleware.OptionalAuth(
					http.HandlerFunc(handlers.OptionalAuthHandler),
				),
			),
		),
	)
	
	// === AUTHENTICATED ROUTES (authentication required) ===
	mux.Handle("/protected", 
		middleware.RequestLogger(logger)(
			middleware.SecurityHeaders(
				authMiddleware.Authenticate(
					http.HandlerFunc(handlers.AuthenticatedHandler),
				),
			),
		),
	)
	
	mux.Handle("/profile", 
		middleware.RequestLogger(logger)(
			middleware.SecurityHeaders(
				authMiddleware.Authenticate(
					http.HandlerFunc(handlers.UserProfileHandler),
				),
			),
		),
	)
	
	// === ROLE-BASED ROUTES ===
	
	// Admin-only route
	mux.Handle("/admin", 
		middleware.RequestLogger(logger)(
			middleware.SecurityHeaders(
				authMiddleware.Authenticate(
					authMiddleware.AdminOnly()(
						http.HandlerFunc(handlers.AdminHandler),
					),
				),
			),
		),
	)
	
	// Multiple role access
	mux.Handle("/command", 
		middleware.RequestLogger(logger)(
			middleware.SecurityHeaders(
				authMiddleware.Authenticate(
					authMiddleware.RequireRole("system_admin", "mission_commander")(
						http.HandlerFunc(handlers.AuthenticatedHandler),
					),
				),
			),
		),
	)
	
	// === PERMISSION-BASED ROUTES ===
	
	// Specific permission required
	mux.Handle("/missions", 
		middleware.RequestLogger(logger)(
			middleware.SecurityHeaders(
				authMiddleware.Authenticate(
					authMiddleware.RequirePermission("missions.read")(
						http.HandlerFunc(handlers.MissionHandler),
					),
				),
			),
		),
	)
	
	// Any of multiple permissions required
	mux.Handle("/reports", 
		middleware.RequestLogger(logger)(
			middleware.SecurityHeaders(
				authMiddleware.Authenticate(
					authMiddleware.RequireAnyPermission("reports.read", "reports.create")(
						http.HandlerFunc(handlers.AuthenticatedHandler),
					),
				),
			),
		),
	)
	
	// === API ROUTES WITH RATE LIMITING ===
	
	// API endpoint with rate limiting
	mux.Handle("/api/data", 
		middleware.RequestLogger(logger)(
			middleware.SecurityHeaders(
				middleware.RateLimit(middleware.RateLimitInfo{
					Limit:  100,                // 100 requests
					Window: 1 * time.Hour,      // per hour
				})(
					middleware.JSONContentType(
						authMiddleware.Authenticate(
							http.HandlerFunc(handlers.AuthenticatedHandler),
						),
					),
				),
			),
		),
	)
	
	// === COMPLEX ROUTE EXAMPLE ===
	
	// Complex route with multiple middleware layers
	mux.Handle("/api/admin/system", 
		middleware.Recovery(logger)(          // Panic recovery (outermost)
			middleware.RequestLogger(logger)(  // Request logging
				middleware.SecurityHeaders(     // Security headers
					middleware.CORS([]string{   // CORS handling
						"https://admin.gotak.local",
						"https://dashboard.gotak.local",
					})(
						middleware.RateLimit(middleware.RateLimitInfo{ // Rate limiting
							Limit:  10,
							Window: 1 * time.Minute,
						})(
							middleware.JSONContentType(     // JSON content type
								authMiddleware.Authenticate( // Authentication
									authMiddleware.RequireRole("system_admin")( // Authorization
										http.HandlerFunc(handlers.AdminHandler), // Final handler
									),
								),
							),
						),
					),
				),
			),
		),
	)
	
	// Return the configured mux wrapped with global middleware
	return middleware.Recovery(logger)(
		middleware.RequestID(
			middleware.UserAgent(mux),
		),
	)
}

// SetupSimpleRouter creates a minimal router for basic testing
func SetupSimpleRouter(authService *auth.AuthService, logger *logger.Logger) http.Handler {
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	handlers := NewExampleHandlers(logger)
	
	mux := http.NewServeMux()
	
	// Simple public route
	mux.HandleFunc("/health", handlers.HealthCheckHandler)
	
	// Simple protected route
	mux.Handle("/protected", 
		authMiddleware.Authenticate(
			http.HandlerFunc(handlers.AuthenticatedHandler),
		),
	)
	
	// Simple admin route
	mux.Handle("/admin", 
		authMiddleware.Authenticate(
			authMiddleware.AdminOnly()(
				http.HandlerFunc(handlers.AdminHandler),
			),
		),
	)
	
	return mux
}

// SetupAPIRouter creates a router specifically for API endpoints
func SetupAPIRouter(authService *auth.AuthService, logger *logger.Logger) http.Handler {
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	handlers := NewExampleHandlers(logger)
	
	mux := http.NewServeMux()
	
	// All API routes get JSON content type and security headers
	apiMiddleware := func(next http.Handler) http.Handler {
		return middleware.SecurityHeaders(
			middleware.JSONContentType(next),
		)
	}
	
	// Public API endpoints
	mux.Handle("/v1/health", 
		apiMiddleware(
			http.HandlerFunc(handlers.HealthCheckHandler),
		),
	)
	
	// Authenticated API endpoints
	mux.Handle("/v1/profile", 
		apiMiddleware(
			authMiddleware.Authenticate(
				http.HandlerFunc(handlers.UserProfileHandler),
			),
		),
	)
	
	// Permission-based API endpoints
	mux.Handle("/v1/missions", 
		apiMiddleware(
			authMiddleware.Authenticate(
				authMiddleware.RequirePermission("missions.read")(
					http.HandlerFunc(handlers.MissionHandler),
				),
			),
		),
	)
	
	// Admin API endpoints
	mux.Handle("/v1/admin/", 
		apiMiddleware(
			authMiddleware.Authenticate(
				authMiddleware.AdminOnly()(
					http.HandlerFunc(handlers.AdminHandler),
				),
			),
		),
	)
	
	// Wrap with request logging and recovery
	return middleware.RequestLogger(logger)(
		middleware.Recovery(logger)(mux),
	)
}
