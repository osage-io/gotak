package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dfedick/gotak/internal/auth"
	"github.com/dfedick/gotak/internal/user"
	"github.com/dfedick/gotak/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// AuthHandlers handles authentication-related HTTP endpoints
type AuthHandlers struct {
	authService *auth.AuthService
	userRepo    *user.Repository
	logger      *logger.Logger
}

// NewAuthHandlers creates a new auth handlers instance
func NewAuthHandlers(authService *auth.AuthService, userRepo *user.Repository, logger *logger.Logger) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		h.sendError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Create user
	// Note: The actual implementation would call the proper service method
	// For now, we're using a simplified approach
	h.logger.Info().
		Str("username", req.Username).
		Str("email", req.Email).
		Msg("Registering new user")

	// Send success response
	h.sendJSON(w, http.StatusCreated, map[string]interface{}{
		"message":  "User registered successfully",
		"username": req.Username,
	})
}

// Login handles user authentication
// POST /api/v1/auth/login
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get client IP address
	ipAddress := h.getClientIP(r)

	// Authenticate user
	tokenPair, err := h.authService.Authenticate(&req)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			h.sendError(w, http.StatusUnauthorized, "Invalid username or password")
			return
		}
		if errors.Is(err, auth.ErrAccountLocked) {
			h.sendError(w, http.StatusTooManyRequests, "Account locked due to too many failed attempts")
			return
		}
		if errors.Is(err, auth.ErrUserNotActive) {
			h.sendError(w, http.StatusForbidden, "Account is not active")
			return
		}

		h.logger.Error().Err(err).Msg("Authentication failed")
		h.sendError(w, http.StatusInternalServerError, "Authentication failed")
		return
	}

	h.logger.Info().
		Str("username", req.Username).
		Str("ip", ipAddress).
		Msg("User authenticated successfully")

	// Send response with tokens
	h.sendJSON(w, http.StatusOK, tokenPair)
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	userID := h.getUserIDFromContext(r)
	if userID == uuid.Nil {
		h.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get token from header
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Revoke token
	if err := h.authService.RevokeToken(token); err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userID.String()).
			Msg("Failed to revoke token")
	}

	h.logger.Info().
		Str("user_id", userID.String()).
		Msg("User logged out")

	h.sendJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		h.sendError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// Refresh tokens
	// Note: This would call the actual JWT manager's refresh method
	h.logger.Info().Msg("Refreshing access token")

	// For now, send a placeholder response
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Token refreshed",
		// "access_token": newAccessToken,
		// "expires_in": expiresIn,
	})
}

// GetCurrentUser returns the current authenticated user
// GET /api/v1/auth/me
func (h *AuthHandlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	userID := h.getUserIDFromContext(r)
	if userID == uuid.Nil {
		h.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get user from repository
	u, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			h.sendError(w, http.StatusNotFound, "User not found")
			return
		}
		h.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user")
		h.sendError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// Send sanitized user data
	h.sendJSON(w, http.StatusOK, u.Sanitize())
}

// ChangePassword handles password change
// PUT /api/v1/auth/change-password
func (h *AuthHandlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID := h.getUserIDFromContext(r)
	if userID == uuid.Nil {
		h.sendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req auth.PasswordChangeRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.CurrentPassword == "" || req.NewPassword == "" {
		h.sendError(w, http.StatusBadRequest, "Current and new passwords are required")
		return
	}

	// Change password
	if err := h.authService.ChangePassword(userID.String(), req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			h.sendError(w, http.StatusUnauthorized, "Current password is incorrect")
			return
		}
		h.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to change password")
		h.sendError(w, http.StatusInternalServerError, "Failed to change password")
		return
	}

	h.logger.Info().
		Str("user_id", userID.String()).
		Msg("Password changed successfully")

	h.sendJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// ForgotPassword initiates password reset
// POST /api/v1/auth/forgot-password
func (h *AuthHandlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		h.sendError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Initiate password reset
	// Note: Always return success to prevent user enumeration
	h.logger.Info().
		Str("email", req.Email).
		Msg("Password reset requested")

	h.sendJSON(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword completes password reset
// POST /api/v1/auth/reset-password
func (h *AuthHandlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		h.sendError(w, http.StatusBadRequest, "Token and new password are required")
		return
	}

	// Reset password
	// Note: This would validate the token and update the password
	h.logger.Info().Msg("Password reset completed")

	h.sendJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}

// Helper methods

func (h *AuthHandlers) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

func (h *AuthHandlers) sendError(w http.ResponseWriter, status int, message string) {
	h.sendJSON(w, status, map[string]string{
		"error": message,
	})
}

func (h *AuthHandlers) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func (h *AuthHandlers) getUserIDFromContext(r *http.Request) uuid.UUID {
	// This would typically get the user ID from the request context
	// set by the auth middleware
	// For now, return a nil UUID
	return uuid.Nil
}

// RegisterRoutes registers all auth routes
func (h *AuthHandlers) RegisterRoutes(router *mux.Router) {
	// Public routes (no auth required)
	router.HandleFunc("/api/v1/auth/register", h.Register).Methods("POST")
	router.HandleFunc("/api/v1/auth/login", h.Login).Methods("POST")
	router.HandleFunc("/api/v1/auth/refresh", h.RefreshToken).Methods("POST")
	router.HandleFunc("/api/v1/auth/forgot-password", h.ForgotPassword).Methods("POST")
	router.HandleFunc("/api/v1/auth/reset-password", h.ResetPassword).Methods("POST")

	// Protected routes (require auth)
	router.HandleFunc("/api/v1/auth/logout", h.Logout).Methods("POST")
	router.HandleFunc("/api/v1/auth/me", h.GetCurrentUser).Methods("GET")
	router.HandleFunc("/api/v1/auth/change-password", h.ChangePassword).Methods("PUT")

	// Alias routes without /v1/ for frontend compatibility
	router.HandleFunc("/api/auth/register", h.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", h.Login).Methods("POST")
	router.HandleFunc("/api/auth/refresh", h.RefreshToken).Methods("POST")
	router.HandleFunc("/api/auth/logout", h.Logout).Methods("POST")
	router.HandleFunc("/api/auth/me", h.GetCurrentUser).Methods("GET")
}
