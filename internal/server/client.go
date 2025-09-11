package server

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"strings"
	"time"

	"github.com/dfedick/gotak/pkg/cot"
	"github.com/dfedick/gotak/pkg/logger"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second
	
	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second
	
	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10
	
	// Maximum message size allowed from peer
	maxMessageSize = 8192
)

// readPump pumps messages from the TCP/TLS connection to the server
func (c *Client) readPump() {
	log := logger.GetGlobalLogger()
	defer func() {
		c.server.unregister <- c
		c.Conn.Close()
	}()

	// Set connection timeouts
	if tcpConn, ok := c.Conn.(*net.TCPConn); ok {
		tcpConn.SetReadBuffer(maxMessageSize)
	}

	reader := bufio.NewReader(c.Conn)
	buffer := make([]byte, 0, maxMessageSize)

	for {
		// Set read deadline
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))

		// Read data from connection
		data := make([]byte, 1024)
		n, err := reader.Read(data)
		if err != nil {
			if err == io.EOF {
				log.Info().Str("endpoint", c.Endpoint).Msg("Client disconnected")
			} else {
				log.Error().Err(err).Str("endpoint", c.Endpoint).Msg("Error reading from client")
			}
			break
		}

		// Append to buffer
		buffer = append(buffer, data[:n]...)

		// Process complete messages
		for {
			message, remaining := c.extractMessage(buffer)
			if message == nil {
				break
			}

			// Process the CoT message
			c.processCoTMessage(message)

			buffer = remaining
		}

		// Check buffer size to prevent memory exhaustion
		if len(buffer) > maxMessageSize {
			log.Warn().Str("endpoint", c.Endpoint).Int("buffer_size", len(buffer)).Msg("Buffer overflow, disconnecting client")
			break
		}
	}
}

// writePump pumps messages from the server to the TCP/TLS connection
func (c *Client) writePump() {
	log := logger.GetGlobalLogger()
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case event, ok := <-c.Send:
			// Set write deadline
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			
			if !ok {
				// The server closed the channel
				return
			}

			// Convert CoT event to XML and send
			xmlData, err := event.ToXML()
			if err != nil {
				log.Error().Err(err).Msg("Error converting CoT event to XML")
				continue
			}

			if _, err := c.Conn.Write(xmlData); err != nil {
				log.Error().Err(err).Str("endpoint", c.Endpoint).Msg("Error writing to client")
				return
			}

		case <-ticker.C:
			// Send periodic ping/heartbeat
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			
			// Create a heartbeat CoT message
			heartbeat := cot.NewEvent("server-heartbeat", cot.TypeSystemHeartbeat)
			xmlData, err := heartbeat.ToXML()
			if err != nil {
				log.Error().Err(err).Msg("Error creating heartbeat message")
				continue
			}

			if _, err := c.Conn.Write(xmlData); err != nil {
				log.Error().Err(err).Str("endpoint", c.Endpoint).Msg("Error sending heartbeat to client")
				return
			}
		}
	}
}

// extractMessage extracts a complete CoT XML message from the buffer
func (c *Client) extractMessage(buffer []byte) ([]byte, []byte) {
	// Look for XML message boundaries
	startTag := []byte("<?xml")
	endTag := []byte("</event>")
	
	// Find start of XML message
	startIdx := bytes.Index(buffer, startTag)
	if startIdx == -1 {
		// Try alternative start without XML declaration
		altStartTag := []byte("<event")
		startIdx = bytes.Index(buffer, altStartTag)
		if startIdx == -1 {
			return nil, buffer
		}
	}
	
	// Find end of XML message after the start
	searchBuffer := buffer[startIdx:]
	endIdx := bytes.Index(searchBuffer, endTag)
	if endIdx == -1 {
		// Complete message not found
		return nil, buffer
	}
	
	// Calculate actual end position
	messageEnd := startIdx + endIdx + len(endTag)
	
	// Extract complete message
	message := buffer[startIdx:messageEnd]
	remaining := buffer[messageEnd:]
	
	return message, remaining
}

// processCoTMessage processes a received CoT message
func (c *Client) processCoTMessage(xmlData []byte) {
	log := logger.GetGlobalLogger()
	// Parse the CoT message
	event, err := cot.ParseCoT(xmlData)
	if err != nil {
		log.Error().Err(err).Str("endpoint", c.Endpoint).Msg("Error parsing CoT message from client")
		return
	}

	// Update last seen timestamp
	c.LastSeen = time.Now()

	// Process the message through the server
	c.server.processMessage(c, event)
}

// SendMessage sends a CoT message to this client
func (c *Client) SendMessage(event *cot.Event) {
	log := logger.GetGlobalLogger()
	select {
	case c.Send <- event:
	default:
		// Channel is full, client is slow
		log.Warn().Str("endpoint", c.Endpoint).Msg("Client send channel full, closing connection")
		close(c.Send)
	}
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	return c.Conn != nil
}

// GetInfo returns basic client information
func (c *Client) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":           c.ID,
		"callsign":     c.Callsign,
		"group":        c.Group,
		"endpoint":     c.Endpoint,
		"protocol":     c.Protocol,
		"connected_at": c.ConnectedAt,
		"last_seen":    c.LastSeen,
		"uptime":       time.Since(c.ConnectedAt),
	}
}

// Disconnect forcibly disconnects the client
func (c *Client) Disconnect() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}

// parseCoTFromStream handles streaming XML parsing for clients that send
// multiple CoT messages in a single stream
func (c *Client) parseCoTFromStream(data []byte) []*cot.Event {
	var events []*cot.Event
	
	// Split by XML declarations or event tags
	parts := bytes.Split(data, []byte("<?xml"))
	
	for i, part := range parts {
		if i == 0 && !bytes.HasPrefix(part, []byte("version=")) {
			// First part might not have XML declaration
			if bytes.HasPrefix(part, []byte("<event")) {
				if event, err := cot.ParseCoT(part); err == nil {
					events = append(events, event)
				}
			}
			continue
		}
		
		// Reconstruct XML with declaration
		if len(part) > 0 {
			xmlData := append([]byte("<?xml"), part...)
			if event, err := cot.ParseCoT(xmlData); err == nil {
				events = append(events, event)
			}
		}
	}
	
	return events
}

// handleTAKProtocolNegotiation handles TAK protocol negotiation messages
func (c *Client) handleTAKProtocolNegotiation(data []byte) {
	log := logger.GetGlobalLogger()
	// TAK clients often send protocol negotiation messages
	// This is a simplified version - real TAK protocol negotiation is more complex
	
	dataStr := string(data)
	
	// Look for TAK protocol version information
	if strings.Contains(dataStr, "version") {
		log.Info().Str("endpoint", c.Endpoint).Str("data", strings.TrimSpace(dataStr)).Msg("Client negotiating TAK protocol")
		
		// Send back a simple acknowledgment
		response := `<?xml version="1.0" encoding="UTF-8"?><event version="2.0" uid="server-ack" type="t-x-c-t" time="` + 
			time.Now().Format(time.RFC3339Nano) + `" start="` + 
			time.Now().Format(time.RFC3339Nano) + `" stale="` + 
			time.Now().Add(time.Minute).Format(time.RFC3339Nano) + `" how="h-g"><point lat="0" lon="0" hae="0" ce="1" le="1"/></event>`
		
		c.Conn.Write([]byte(response))
	}
}

// GetStatistics returns connection statistics for this client
func (c *Client) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"messages_sent":     0, // TODO: implement message counters
		"messages_received": 0,
		"bytes_sent":        0,
		"bytes_received":    0,
		"connection_errors": 0,
		"last_error":        nil,
	}
}

// SetGroup sets the client's group membership
func (c *Client) SetGroup(group string) {
	log := logger.GetGlobalLogger()
	c.Group = group
	log.Info().Str("callsign", c.Callsign).Str("group", group).Msg("Client assigned to group")
}

// SetCallsign sets the client's callsign
func (c *Client) SetCallsign(callsign string) {
	log := logger.GetGlobalLogger()
	c.Callsign = callsign
	log.Info().Str("endpoint", c.Endpoint).Str("callsign", callsign).Msg("Client callsign set")
}
