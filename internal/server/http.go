package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/dfedick/gotak/internal/auth"
	"github.com/dfedick/gotak/internal/chat"
	"github.com/dfedick/gotak/internal/events"
	"github.com/dfedick/gotak/internal/handlers"
	"github.com/dfedick/gotak/internal/mission"
	"github.com/dfedick/gotak/internal/position"
	"github.com/dfedick/gotak/pkg/config"
	"github.com/dfedick/gotak/pkg/database"
	"github.com/dfedick/gotak/pkg/logger"
)

// PositionService interface for position tracking operations
type PositionService interface {
	GetAllPositions() []*position.EntityPosition
	GetPosition(entityID string) (*position.EntityPosition, bool)
	GetTrail(entityID string) []position.PositionHistory
	RemoveEntity(entityID string)
	GetStatistics() map[string]interface{}
}

// HTTPServer represents the HTTP server for web API endpoints
type HTTPServer struct {
	server           *http.Server
	router           *mux.Router
	config           *config.ServerConfig
	logger           *logger.Logger
	wsHub            *handlers.TacticalWSHub
	entityService    handlers.EntityService
	chatHandlers     *handlers.ChatHandlers
	positionHandlers *handlers.PositionHandlers
	simpleAuthHandlers *handlers.SimpleAuthHandlers
	missionHandlers  *mission.Handlers
	authMiddleware   *auth.Middleware
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(cfg *config.ServerConfig, log *logger.Logger, chatService *chat.Service, positionService PositionService, db *sqlx.DB) *HTTPServer {
	router := mux.NewRouter()
	
	// Create WebSocket hub
	wsHub := handlers.NewTacticalWSHub(log, chatService)
	
	// Create entity service (using mock for now)
	entityService := handlers.NewMockEntityService(log)
	
	// Create chat handlers
	chatHandlers := handlers.NewChatHandlers(chatService, log)
	
	// Create position handlers
	positionHandlers := handlers.NewPositionHandlers(positionService, log)
	
	// Create auth system components
	var simpleAuthHandlers *handlers.SimpleAuthHandlers
	var authMiddleware *auth.Middleware
	var missionHandlers *mission.Handlers
	
	if db != nil {
		// Create simplified auth service
		jwtConfig := auth.JWTConfig{
			SecretKey:  "your-secret-key-change-in-production", // TODO: Load from config
			AccessTTL:  24 * time.Hour,
			RefreshTTL: 7 * 24 * time.Hour,
			Issuer:     "gotak-server",
		}
		
		simpleAuthService := auth.NewSimpleAuthService(db, jwtConfig, log)
		
		// Create simple auth handlers
		simpleAuthHandlers = handlers.NewSimpleAuthHandlers(simpleAuthService, log)
		
		// Create auth middleware
		tokenStorage := auth.NewInMemoryTokenStorage()
		jwtManager := auth.NewJWTManager(jwtConfig, tokenStorage)
		authMiddleware = auth.NewMiddleware(jwtManager, log)
		
		// Create mission service and handlers
		eventPublisher := events.NewSimplePublisher(log)
		dbAdapter := database.NewSQLXAdapter(db)
		missionService := mission.NewService(dbAdapter, log, eventPublisher)
		missionHandlers = mission.NewHandlers(missionService, log)
	}
	
	httpServer := &HTTPServer{
		router:           router,
		config:           cfg,
		logger:           log,
		wsHub:            wsHub,
		entityService:    entityService,
		chatHandlers:     chatHandlers,
		positionHandlers: positionHandlers,
		simpleAuthHandlers: simpleAuthHandlers,
		missionHandlers:  missionHandlers,
		authMiddleware:   authMiddleware,
	}
	
	// Setup routes
	httpServer.setupRoutes()
	
	// Create HTTP server
	httpServer.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	return httpServer
}

// setupRoutes configures all HTTP routes
func (h *HTTPServer) setupRoutes() {
	// Enable CORS middleware
	h.router.Use(corsMiddleware)
	h.router.Use(loggingMiddleware(h.logger))
	
	// API v1 routes
	api := h.router.PathPrefix("/api/v1").Subrouter()
	
	// Auth routes (public - no auth required)
	if h.simpleAuthHandlers != nil {
		api.HandleFunc("/auth/register", h.simpleAuthHandlers.Register).Methods("POST")
		api.HandleFunc("/auth/login", h.simpleAuthHandlers.Login).Methods("POST")
		api.HandleFunc("/auth/refresh", h.simpleAuthHandlers.RefreshToken).Methods("POST")
		api.HandleFunc("/auth/forgot-password", h.simpleAuthHandlers.ForgotPassword).Methods("POST")
		api.HandleFunc("/auth/reset-password", h.simpleAuthHandlers.ResetPassword).Methods("POST")
		
		// Protected auth routes
		api.HandleFunc("/auth/logout", h.simpleAuthHandlers.Logout).Methods("POST")
		api.HandleFunc("/auth/me", h.simpleAuthHandlers.GetCurrentUser).Methods("GET")
		api.HandleFunc("/auth/change-password", h.simpleAuthHandlers.ChangePassword).Methods("PUT")
		
		// Add OPTIONS handlers for auth routes
		api.HandleFunc("/auth/register", handlePreflight).Methods("OPTIONS")
		api.HandleFunc("/auth/login", handlePreflight).Methods("OPTIONS")
		api.HandleFunc("/auth/refresh", handlePreflight).Methods("OPTIONS")
		api.HandleFunc("/auth/forgot-password", handlePreflight).Methods("OPTIONS")
		api.HandleFunc("/auth/reset-password", handlePreflight).Methods("OPTIONS")
		api.HandleFunc("/auth/logout", handlePreflight).Methods("OPTIONS")
		api.HandleFunc("/auth/me", handlePreflight).Methods("OPTIONS")
		api.HandleFunc("/auth/change-password", handlePreflight).Methods("OPTIONS")
	}
	
	// Entity endpoints
	api.HandleFunc("/entities", handlers.HandleGetEntities(h.entityService, h.logger)).Methods("GET")
	api.HandleFunc("/entities/{id}", handlers.HandleGetEntity(h.entityService, h.logger)).Methods("GET")
	api.HandleFunc("/entities/{id}/history", handlers.HandleGetEntityHistory(h.entityService, h.logger)).Methods("GET")
	
	// Chat endpoints
	api.HandleFunc("/chat/rooms", h.chatHandlers.CreateRoom).Methods("POST")
	api.HandleFunc("/chat/rooms", h.chatHandlers.GetRooms).Methods("GET")
	api.HandleFunc("/chat/rooms/{roomId}", h.chatHandlers.GetRoom).Methods("GET")
	api.HandleFunc("/chat/rooms/{roomId}/messages", h.chatHandlers.SendMessage).Methods("POST")
	api.HandleFunc("/chat/rooms/{roomId}/messages", h.chatHandlers.GetMessages).Methods("GET")
	api.HandleFunc("/chat/rooms/{roomId}/participants", h.chatHandlers.GetRoomParticipants).Methods("GET")
	api.HandleFunc("/chat/rooms/{roomId}/participants", h.chatHandlers.AddParticipant).Methods("POST")
	api.HandleFunc("/chat/rooms/{roomId}/participants/{userId}", h.chatHandlers.RemoveParticipant).Methods("DELETE")
	api.HandleFunc("/chat/messages/{messageId}/acknowledge", h.chatHandlers.AcknowledgeMessage).Methods("POST")
	api.HandleFunc("/chat/messages/{messageId}/reactions", h.chatHandlers.AddReaction).Methods("POST")
	api.HandleFunc("/chat/statistics", h.chatHandlers.GetStatistics).Methods("GET")
	
	// Position endpoints
	api.HandleFunc("/positions", h.positionHandlers.GetAllPositions).Methods("GET")
	api.HandleFunc("/positions/active", h.positionHandlers.GetActivePositions).Methods("GET")
	api.HandleFunc("/positions/friendly", h.positionHandlers.GetFriendlyPositions).Methods("GET")
	api.HandleFunc("/positions/hostile", h.positionHandlers.GetHostilePositions).Methods("GET")
	api.HandleFunc("/positions/bounds", h.positionHandlers.GetPositionsInBounds).Methods("GET")
	api.HandleFunc("/positions/statistics", h.positionHandlers.GetPositionStatistics).Methods("GET")
	api.HandleFunc("/positions/{entityId}", h.positionHandlers.GetPosition).Methods("GET")
	api.HandleFunc("/positions/{entityId}", h.positionHandlers.DeletePosition).Methods("DELETE")
	api.HandleFunc("/positions/{entityId}/trail", h.positionHandlers.GetPositionTrail).Methods("GET")
	
	// Mission endpoints - register routes if mission handlers are available
	if h.missionHandlers != nil {
		h.missionHandlers.RegisterRoutes(api)
	}
	
	// Handle preflight requests
	api.HandleFunc("/entities", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/entities/{id}", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/entities/{id}/history", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/chat/rooms", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/chat/rooms/{roomId}", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/chat/rooms/{roomId}/messages", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/chat/rooms/{roomId}/participants", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/chat/rooms/{roomId}/participants/{userId}", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/chat/messages/{messageId}/acknowledge", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/chat/messages/{messageId}/reactions", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/chat/statistics", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/positions", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/positions/active", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/positions/friendly", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/positions/hostile", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/positions/bounds", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/positions/statistics", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/positions/{entityId}", handlePreflight).Methods("OPTIONS")
	api.HandleFunc("/positions/{entityId}/trail", handlePreflight).Methods("OPTIONS")
	
	// WebSocket endpoint for tactical data
	h.router.HandleFunc("/ws/tactical", handlers.HandleTacticalWebSocket(h.wsHub, h.logger))
	
	// Health check endpoint
	h.router.HandleFunc("/health", h.handleHealth).Methods("GET")
	
	// Static file serving for web UI
	if h.config.Server.ServeStatic {
		// Check if running in container (web directory at /app/web)
		webDir := "/app/web"
		if _, err := os.Stat(webDir); os.IsNotExist(err) {
			// Fallback to local development path
			webDir = "./web/dist"
		}
		h.router.PathPrefix("/").Handler(http.FileServer(http.Dir(webDir)))
	}
}

// Start starts the HTTP server
func (h *HTTPServer) Start(ctx context.Context) error {
	h.logger.Info().Str("address", h.server.Addr).Msg("Starting HTTP server")
	
	// Start WebSocket hub
	go h.wsHub.Run()
	
	// Start HTTP server
	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Error().Err(err).Msg("HTTP server error")
		}
	}()
	
	h.logger.Info().Str("address", h.server.Addr).Msg("HTTP server started")
	
	return nil
}

// Shutdown gracefully shuts down the HTTP server
func (h *HTTPServer) Shutdown(ctx context.Context) error {
	h.logger.Info().Msg("Shutting down HTTP server")
	return h.server.Shutdown(ctx)
}

// GetWSHub returns the WebSocket hub for integration with TAK server
func (h *HTTPServer) GetWSHub() *handlers.TacticalWSHub {
	return h.wsHub
}

// Health check handler
func (h *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// Simple JSON response
	fmt.Fprintf(w, `{"status":"ok","service":"gotak-server","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func loggingMiddleware(logger *logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap the response writer to capture status code
			wrapped := &responseWrapper{ResponseWriter: w, statusCode: 200}
			
			next.ServeHTTP(wrapped, r)
			
			duration := time.Since(start)
			
			logger.Info().Str("method", r.Method).Str("path", r.URL.Path).Str("remote_addr", r.RemoteAddr).Str("user_agent", r.UserAgent()).Int("status", wrapped.statusCode).Int64("duration_ms", duration.Milliseconds()).Msg("HTTP request")
		})
	}
}

// responseWrapper wraps http.ResponseWriter to capture status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Handle preflight requests
func handlePreflight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusNoContent)
}
