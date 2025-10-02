package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/dfedick/gotak/internal/chat"
	"github.com/dfedick/gotak/internal/position"
	"github.com/dfedick/gotak/pkg/config"
	"github.com/dfedick/gotak/pkg/cot"
	"github.com/dfedick/gotak/pkg/logger"
)

// Server represents the main TAK server instance
type Server struct {
	config *config.ServerConfig
	logger *logger.Logger
	db     *sqlx.DB
	
	// Connection management
	clients    map[string]*Client
	clientsMux sync.RWMutex
	
	// Message channels
	broadcast chan *cot.Event
	register  chan *Client
	unregister chan *Client
	
	// Server listeners
	tcpListener net.Listener
	udpConn     *net.UDPConn
	tlsListener net.Listener
	
	// HTTP server for web API
	httpServer *HTTPServer
	
	// Position tracking service
	positionService *position.Service
	
	// Shutdown handling
	shutdownCh chan struct{}
	wg         sync.WaitGroup
}

// Client represents a connected TAK client
type Client struct {
	ID       string
	Conn     net.Conn
	Callsign string
	Group    string
	
	// Client metadata
	Endpoint   string
	ConnectedAt time.Time
	LastSeen    time.Time
	
	// Message handling
	Send   chan *cot.Event
	server *Server
	
	// Connection type
	Protocol string // "tcp", "udp", "tls"
}

// New creates a new TAK server instance
func New(cfg *config.ServerConfig) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}
	
	log := logger.GetGlobalLogger()
	
	// Initialize database connection (optional for now)
	var db *sqlx.DB
	if cfg.Database.Host != "" {
		// Debug: Log the configuration values
		log.Debug().
			Str("host", cfg.Database.Host).
			Int("port", cfg.Database.Port).
			Str("username", cfg.Database.Username).
			Str("database", cfg.Database.Database).
			Str("sslmode", cfg.Database.SSLMode).
			Msg("Attempting database connection")
		
		// Use URL format for better password handling
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Database,
			cfg.Database.SSLMode,
		)
		
		var err error
		db, err = sqlx.Connect("postgres", dsn)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to database, running without persistence")
			// Continue without database - auth will be disabled
		} else {
			log.Info().Msg("Database connection established")
			// Set connection pool settings
			db.SetMaxOpenConns(25)
			db.SetMaxIdleConns(5)
			db.SetConnMaxLifetime(5 * time.Minute)
		}
	}
	
	// Create chat service
	chatService := chat.NewService(nil, log) // TODO: Pass db when chat persistence is ready
	
	// Create position service
	positionService := position.NewService()
	
	// Create HTTP server
	httpServer := NewHTTPServer(cfg, log, chatService, positionService, db)
	
	server := &Server{
		config:          cfg,
		logger:          log,
		db:              db,
		clients:         make(map[string]*Client),
		broadcast:       make(chan *cot.Event, 1000),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		httpServer:      httpServer,
		positionService: positionService,
		shutdownCh:      make(chan struct{}),
	}
	
	return server, nil
}

// BroadcastPositionUpdate broadcasts a position update to web clients
func (s *Server) BroadcastPositionUpdate(callsign string, lat, lng float64, altitude, speed, course *float64) {
	if s.httpServer != nil && s.httpServer.wsHub != nil {
		s.httpServer.wsHub.BroadcastPositionUpdate(callsign, lat, lng, altitude, speed, course)
	}
}

// Start starts the TAK server
func (s *Server) Start(ctx context.Context) error {
	log := logger.GetGlobalLogger()
	log.Info().Str("host", s.config.Server.Host).Msg("Starting GoTAK Server")
	
	// Start the message hub
	s.wg.Add(1)
	go s.messageHub(ctx)
	
	// Start TCP listener
	if err := s.startTCPListener(ctx); err != nil {
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}
	
	// Start UDP listener
	if err := s.startUDPListener(ctx); err != nil {
		return fmt.Errorf("failed to start UDP listener: %w", err)
	}
	
	// Start TLS listener if enabled
	if s.config.Security.TLSEnabled {
		if err := s.startTLSListener(ctx); err != nil {
			return fmt.Errorf("failed to start TLS listener: %w", err)
		}
	}
	
	// Start HTTP server
	if err := s.httpServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	
	log.Info().Msg("GoTAK Server started successfully")
	log.Info().Int("port", s.config.Server.TCPPort).Msg("TCP listening")
	log.Info().Int("port", s.config.Server.UDPPort).Msg("UDP listening")
	log.Info().Int("port", s.config.Server.HTTPPort).Msg("HTTP listening")
	if s.config.Security.TLSEnabled {
		log.Info().Int("port", s.config.Server.TLSPort).Msg("TLS listening")
	}
	
	// Wait for shutdown
	<-s.shutdownCh
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log := logger.GetGlobalLogger()
	log.Info().Msg("Shutting down GoTAK Server")
	
	// Signal shutdown
	close(s.shutdownCh)
	
	// Close all client connections
	s.clientsMux.Lock()
	for _, client := range s.clients {
		client.Conn.Close()
	}
	s.clientsMux.Unlock()
	
	// Shutdown position service
	if s.positionService != nil {
		s.positionService.Close()
	}
	
	// Shutdown HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Warn().Err(err).Msg("Error shutting down HTTP server")
		}
	}
	
	// Close database connection
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			log.Warn().Err(err).Msg("Error closing database connection")
		}
	}
	
	// Close listeners
	if s.tcpListener != nil {
		s.tcpListener.Close()
	}
	if s.udpConn != nil {
		s.udpConn.Close()
	}
	if s.tlsListener != nil {
		s.tlsListener.Close()
	}
	
	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		log.Info().Msg("GoTAK Server shutdown complete")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// startTCPListener starts the TCP listener for TAK clients
func (s *Server) startTCPListener(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.TCPPort)
	
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on TCP %s: %w", addr, err)
	}
	
	s.tcpListener = listener
	
	s.wg.Add(1)
	go func() {
		log := logger.GetGlobalLogger()
		defer s.wg.Done()
		defer listener.Close()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.shutdownCh:
				return
			default:
			}
			
			// Set accept timeout to allow checking for shutdown
			if tcpListener, ok := listener.(*net.TCPListener); ok {
				tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
			}
			
			conn, err := listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Timeout, check for shutdown
				}
				select {
				case <-s.shutdownCh:
					return // Server is shutting down
				default:
					log.Error().Err(err).Msg("Error accepting TCP connection")
					continue
				}
			}
			
			// Handle the connection
			go s.handleTCPConnection(conn)
		}
	}()
	
	return nil
}

// startUDPListener starts the UDP listener for TAK clients
func (s *Server) startUDPListener(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.UDPPort)
	
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address %s: %w", addr, err)
	}
	
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP %s: %w", addr, err)
	}
	
	s.udpConn = conn
	
	s.wg.Add(1)
	go func() {
		log := logger.GetGlobalLogger()
		defer s.wg.Done()
		defer conn.Close()
		
		buffer := make([]byte, s.config.TAK.MaxMessageSize)
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.shutdownCh:
				return
			default:
			}
			
			// Set read timeout
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			n, clientAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Timeout, check for shutdown
				}
				select {
				case <-s.shutdownCh:
					return
				default:
					log.Error().Err(err).Msg("Error reading UDP message")
					continue
				}
			}
			
			// Process UDP message
			go s.handleUDPMessage(buffer[:n], clientAddr, conn)
		}
	}()
	
	return nil
}

// startTLSListener starts the TLS listener for secure TAK clients
func (s *Server) startTLSListener(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.TLSPort)
	
	// Load TLS certificate
	cert, err := tls.LoadX509KeyPair(s.config.Security.CertFile, s.config.Security.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}
	
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	
	// Configure client certificate authentication if required
	if s.config.Security.ClientAuthRequired {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	
	listener, err := tls.Listen("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to listen on TLS %s: %w", addr, err)
	}
	
	s.tlsListener = listener
	
	s.wg.Add(1)
	go func() {
		log := logger.GetGlobalLogger()
		defer s.wg.Done()
		defer listener.Close()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.shutdownCh:
				return
			default:
			}
			
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-s.shutdownCh:
					return
				default:
					log.Error().Err(err).Msg("Error accepting TLS connection")
					continue
				}
			}
			
			// Handle the TLS connection
			go s.handleTLSConnection(conn)
		}
	}()
	
	return nil
}

// handleTCPConnection handles a new TCP connection from a TAK client
func (s *Server) handleTCPConnection(conn net.Conn) {
	log := logger.GetGlobalLogger()
	client := &Client{
		ID:          generateClientID(),
		Conn:        conn,
		Endpoint:    conn.RemoteAddr().String(),
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
		Send:        make(chan *cot.Event, 256),
		server:      s,
		Protocol:    "tcp",
	}
	
	log.Info().Str("endpoint", client.Endpoint).Str("protocol", "tcp").Msg("New client connected")
	
	// Register client
	s.register <- client
	
	// Start client handlers
	go client.readPump()
	go client.writePump()
}

// handleTLSConnection handles a new TLS connection from a TAK client
func (s *Server) handleTLSConnection(conn net.Conn) {
	log := logger.GetGlobalLogger()
	client := &Client{
		ID:          generateClientID(),
		Conn:        conn,
		Endpoint:    conn.RemoteAddr().String(),
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
		Send:        make(chan *cot.Event, 256),
		server:      s,
		Protocol:    "tls",
	}
	
	log.Info().Str("endpoint", client.Endpoint).Str("protocol", "tls").Msg("New client connected")
	
	// Register client
	s.register <- client
	
	// Start client handlers
	go client.readPump()
	go client.writePump()
}

// handleUDPMessage handles a UDP message from a TAK client
func (s *Server) handleUDPMessage(data []byte, clientAddr *net.UDPAddr, conn *net.UDPConn) {
	log := logger.GetGlobalLogger()
	// Parse CoT message
	event, err := cot.ParseCoT(data)
	if err != nil {
		log.Error().Err(err).Str("client_addr", clientAddr.String()).Msg("Error parsing UDP CoT message")
		return
	}
	
	// Create temporary client for UDP (stateless)
	client := &Client{
		ID:       generateClientID(),
		Endpoint: clientAddr.String(),
		LastSeen: time.Now(),
		Protocol: "udp",
		server:   s,
	}
	
	// Extract client info from CoT message
	client.Callsign = event.GetCallsign()
	client.Group = event.GetGroup()
	
	// Process the message
	s.processMessage(client, event)
}

// messageHub handles message broadcasting and client management
func (s *Server) messageHub(ctx context.Context) {
	defer s.wg.Done()
	log := logger.GetGlobalLogger()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.shutdownCh:
			return
		case client := <-s.register:
			s.clientsMux.Lock()
			s.clients[client.ID] = client
			s.clientsMux.Unlock()
			log.Info().Str("client_id", client.ID).Str("endpoint", client.Endpoint).Msg("Client registered")
			
		case client := <-s.unregister:
			s.clientsMux.Lock()
			if _, ok := s.clients[client.ID]; ok {
				delete(s.clients, client.ID)
				close(client.Send)
			}
			s.clientsMux.Unlock()
			log.Info().Str("client_id", client.ID).Str("endpoint", client.Endpoint).Msg("Client unregistered")
			
		case event := <-s.broadcast:
			// Broadcast message to all clients
			s.clientsMux.RLock()
			for _, client := range s.clients {
				select {
				case client.Send <- event:
				default:
					// Client send buffer is full, close it
					close(client.Send)
					delete(s.clients, client.ID)
				}
			}
			s.clientsMux.RUnlock()
			
		case <-ticker.C:
			// Clean up stale clients
			s.cleanupStaleClients()
		}
	}
}

// processMessage processes a received CoT message
func (s *Server) processMessage(client *Client, event *cot.Event) {
	log := logger.GetGlobalLogger()
	// Update client info from message
	if callsign := event.GetCallsign(); callsign != "" {
		client.Callsign = callsign
	}
	if group := event.GetGroup(); group != "" {
		client.Group = group
	}
	
	client.LastSeen = time.Now()
	
	log.Debug().Str("type", event.Type).Str("callsign", client.Callsign).Str("endpoint", client.Endpoint).Msg("Received CoT message")
	
	// Validate message
	if err := event.IsValid(); err != nil {
		log.Warn().Err(err).Str("endpoint", client.Endpoint).Msg("Invalid CoT message")
		return
	}
	
	// Process based on message type
	switch {
	case cot.IsTypeSystem(event.Type):
		s.handleSystemMessage(client, event)
	case cot.IsTypeChat(event.Type):
		s.handleChatMessage(client, event)
	case cot.IsTypeAtom(event.Type) || cot.IsTypeBit(event.Type):
		s.handlePositionMessage(client, event)
	default:
		// Generic message handling
		s.broadcast <- event
	}
}

// handleSystemMessage handles system messages (heartbeats, pings, etc.)
func (s *Server) handleSystemMessage(client *Client, event *cot.Event) {
	switch event.Type {
	case cot.TypeSystemHeartbeat:
		// Update client last seen time
		client.LastSeen = time.Now()
	case cot.TypeSystemPing:
		// Respond to ping (echo back)
		s.broadcast <- event
	}
}

// handleChatMessage handles chat messages
func (s *Server) handleChatMessage(client *Client, event *cot.Event) {
	log := logger.GetGlobalLogger()
	messageText := ""
	if event.Detail != nil && event.Detail.Remarks != nil {
		messageText = event.Detail.Remarks.Text
	}
	log.Info().Str("callsign", client.Callsign).Str("message", messageText).Msg("Chat message received")
	
	// Broadcast chat message to all clients
	s.broadcast <- event
}

// handlePositionMessage handles position update messages
func (s *Server) handlePositionMessage(client *Client, event *cot.Event) {
	log := logger.GetGlobalLogger()
	if lat, lon, err := event.GetPosition(); err == nil {
		log.Debug().Str("callsign", client.Callsign).Float64("lat", lat).Float64("lon", lon).Msg("Position update received")
		
		// Extract additional position data
		var altitude, speed, course *float64
		
		// Parse altitude from HAE (Height Above Ellipsoid)
		if event.Point != nil && event.Point.Hae != "" {
			if hae, err := strconv.ParseFloat(event.Point.Hae, 64); err == nil {
				altitude = &hae
			}
		}
		
		// Parse speed and course from track data
		if event.Detail != nil && event.Detail.Track != nil {
			if event.Detail.Track.Speed != "" {
				if s, err := strconv.ParseFloat(event.Detail.Track.Speed, 64); err == nil {
					speed = &s
				}
			}
			if event.Detail.Track.Course != "" {
				if c, err := strconv.ParseFloat(event.Detail.Track.Course, 64); err == nil {
					course = &c
				}
			}
		}
		
		// Use callsign from event or client as entity ID
		entityID := client.Callsign
		if entityID == "" {
			entityID = event.UID
		}
		
		// Update position in tracking service
		if err := s.positionService.UpdatePosition(event, client.Callsign); err != nil {
			log.Warn().Err(err).Str("entity_id", entityID).Msg("Failed to update position in service")
		}
		
		// Broadcast to WebSocket clients for real-time map updates
		s.BroadcastPositionUpdate(entityID, lat, lon, altitude, speed, course)
	}
	
	// Broadcast CoT message to TAK clients
	s.broadcast <- event
}

// cleanupStaleClients removes clients that haven't been seen recently
func (s *Server) cleanupStaleClients() {
	log := logger.GetGlobalLogger()
	staleTimeout := 5 * time.Minute
	now := time.Now()
	
	s.clientsMux.Lock()
	for id, client := range s.clients {
		if now.Sub(client.LastSeen) > staleTimeout {
			log.Info().Str("client_id", id).Str("endpoint", client.Endpoint).Msg("Removing stale client")
			client.Conn.Close()
			delete(s.clients, id)
		}
	}
	s.clientsMux.Unlock()
}

// GetConnectedClients returns a list of currently connected clients
func (s *Server) GetConnectedClients() []*Client {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()
	
	clients := make([]*Client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	
	return clients
}

// BroadcastMessage broadcasts a CoT message to all connected clients
func (s *Server) BroadcastMessage(event *cot.Event) {
	log := logger.GetGlobalLogger()
	select {
	case s.broadcast <- event:
	default:
		log.Warn().Str("event_type", event.Type).Msg("Broadcast channel full, dropping message")
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}
