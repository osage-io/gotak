package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dfedick/gotak/internal/middleware"
	"github.com/dfedick/gotak/pkg/logger"
)

// ExampleHandlers demonstrates how to use authentication middleware
type ExampleHandlers struct {
	logger *logger.Logger
}

// NewExampleHandlers creates new example handlers
func NewExampleHandlers(logger *logger.Logger) *ExampleHandlers {
	return &ExampleHandlers{
		logger: logger,
	}
}

// PublicHandler demonstrates a public endpoint (no authentication required)
func (h *ExampleHandlers) PublicHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "This is a public endpoint - no authentication required",
		"public":  true,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AuthenticatedHandler demonstrates a protected endpoint (authentication required)
func (h *ExampleHandlers) AuthenticatedHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		// This shouldn't happen if middleware is properly configured
		http.Error(w, "Authentication context not found", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"message":     "This is a protected endpoint",
		"user_id":     user.UserID,
		"username":    user.Username,
		"roles":       user.Roles,
		"permissions": user.Permissions,
		"authenticated": true,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AdminHandler demonstrates an admin-only endpoint
func (h *ExampleHandlers) AdminHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	
	response := map[string]interface{}{
		"message":  "This is an admin-only endpoint",
		"user_id":  user.UserID,
		"username": user.Username,
		"admin":    true,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MissionHandler demonstrates a permission-based endpoint
func (h *ExampleHandlers) MissionHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	
	// Additional permission checks can be done in the handler if needed
	if !middleware.HasPermission(r.Context(), "missions.read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}
	
	response := map[string]interface{}{
		"message":     "Mission data access granted",
		"user_id":     user.UserID,
		"username":    user.Username,
		"permissions": user.Permissions,
		"missions":    []string{"Mission Alpha", "Mission Beta", "Mission Gamma"},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// OptionalAuthHandler demonstrates optional authentication
func (h *ExampleHandlers) OptionalAuthHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	
	response := map[string]interface{}{
		"message": "This endpoint works with or without authentication",
	}
	
	if user != nil {
		response["authenticated"] = true
		response["user_id"] = user.UserID
		response["username"] = user.Username
		response["greeting"] = "Hello, " + user.Username + "!"
	} else {
		response["authenticated"] = false
		response["greeting"] = "Hello, anonymous user!"
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UserProfileHandler demonstrates accessing user information
func (h *ExampleHandlers) UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	username := middleware.GetUsernameFromContext(r.Context())
	roles := middleware.GetUserRolesFromContext(r.Context())
	permissions := middleware.GetUserPermissionsFromContext(r.Context())
	
	// In a real implementation, you might fetch additional user data from the database
	profile := map[string]interface{}{
		"user_id":     userID,
		"username":    username,
		"roles":       roles,
		"permissions": permissions,
		"profile": map[string]interface{}{
			"full_name":    "User Full Name", // Would come from database
			"email":        "user@example.com", // Would come from database
			"last_login":   "2025-01-08T20:30:00Z", // Would come from database
			"account_type": "standard",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// HealthCheckHandler is always public and doesn't need authentication
func (h *ExampleHandlers) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": "2025-01-08T20:30:00Z",
		"version":   "1.0.0",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
