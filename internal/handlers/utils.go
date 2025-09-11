package handlers

import (
	"encoding/json"
	"net/http"
)

// APIResponse represents a standard API response format
type APIResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// WriteJSONResponse writes a JSON response with the given data and status code
func WriteJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// WriteJSONError writes a JSON error response
func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	response := APIResponse{
		Status:  "error",
		Message: message,
		Error:   message,
	}
	WriteJSONResponse(w, response, statusCode)
}

// WriteJSONSuccess writes a JSON success response
func WriteJSONSuccess(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Status: "success",
		Data:   data,
	}
	WriteJSONResponse(w, response, http.StatusOK)
}
