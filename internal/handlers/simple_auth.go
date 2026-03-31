package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dfedick/gotak/internal/auth"
	"github.com/dfedick/gotak/pkg/logger"
)

// SimpleAuthHandlers handles simplified authentication requests
type SimpleAuthHandlers struct {
	authService *auth.SimpleAuthService
	logger      *logger.Logger
}

// NewSimpleAuthHandlers creates new simplified auth handlers
func NewSimpleAuthHandlers(authService *auth.SimpleAuthService, logger *logger.Logger) *SimpleAuthHandlers {
	return &SimpleAuthHandlers{
		authService: authService,
		logger:      logger,
	}
}

// Register handles user registration
func (h *SimpleAuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req auth.SimpleRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode registration request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Register user
	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if err == auth.ErrUserAlreadyExists {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to register user")
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id":       user.ID.String(),
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Login handles user login
func (h *SimpleAuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.SimpleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode login request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Authenticate user
	tokenPair, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if err == auth.ErrInvalidCredentials || err == auth.ErrUserNotActive {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to authenticate user")
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Return tokens
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenPair)
}

// RefreshToken handles token refresh
func (h *SimpleAuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode refresh request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Refresh tokens
	tokenPair, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to refresh token")
		http.Error(w, "Token refresh failed", http.StatusUnauthorized)
		return
	}

	// Return new tokens
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenPair)
}

// GetCurrentUser returns the current authenticated user
func (h *SimpleAuthHandlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get token from header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	// Extract token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
		return
	}
	token := parts[1]

	// Validate token
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get user
	user, err := h.authService.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get user")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Return user info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"active":   user.Active,
	})
}

// ChangePassword handles password change requests
func (h *SimpleAuthHandlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user ID from token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
		return
	}

	claims, err := h.authService.ValidateToken(parts[1])
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Change password
	err = h.authService.ChangePassword(r.Context(), claims.UserID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			http.Error(w, "Invalid current password", http.StatusUnauthorized)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to change password")
		http.Error(w, "Password change failed", http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password changed successfully",
	})
}

// Logout handles user logout (placeholder for token revocation)
func (h *SimpleAuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, you would revoke the token here
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// ForgotPassword handles forgot password requests (placeholder)
func (h *SimpleAuthHandlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	// This would need email service integration
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// ResetPassword handles password reset (placeholder)
func (h *SimpleAuthHandlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// This would need email service integration
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
