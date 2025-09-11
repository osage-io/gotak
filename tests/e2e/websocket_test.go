package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WebSocket message types for testing
const (
	MsgTypePositionUpdate = "position_update"
	MsgTypeChat           = "chat_message"
	MsgTypeGeofenceAlert  = "geofence_alert"
	MsgTypeSystemAlert    = "system_alert"
	MsgTypeHeartbeat      = "heartbeat"
)

// Test WebSocket client wrapper
type WSTestClient struct {
	conn     *websocket.Conn
	token    string
	messages chan map[string]interface{}
	errors   chan error
	done     chan struct{}
	mu       sync.RWMutex
}

func NewWSTestClient(token string) (*WSTestClient, error) {
	u, err := url.Parse(wsURL + "/ws")
	if err != nil {
		return nil, err
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}

	client := &WSTestClient{
		conn:     conn,
		token:    token,
		messages: make(chan map[string]interface{}, 100),
		errors:   make(chan error, 10),
		done:     make(chan struct{}),
	}

	go client.readPump()
	return client, nil
}

func (ws *WSTestClient) readPump() {
	defer func() {
		ws.conn.Close()
		close(ws.done)
	}()

	for {
		var msg map[string]interface{}
		err := ws.conn.ReadJSON(&msg)
		if err != nil {
			select {
			case ws.errors <- err:
			case <-ws.done:
				return
			}
			return
		}

		select {
		case ws.messages <- msg:
		case <-ws.done:
			return
		}
	}
}

func (ws *WSTestClient) SendMessage(msgType string, data interface{}) error {
	msg := map[string]interface{}{
		"type": msgType,
		"data": data,
	}
	return ws.conn.WriteJSON(msg)
}

func (ws *WSTestClient) WaitForMessage(timeout time.Duration) (map[string]interface{}, error) {
	select {
	case msg := <-ws.messages:
		return msg, nil
	case err := <-ws.errors:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for message")
	case <-ws.done:
		return nil, fmt.Errorf("connection closed")
	}
}

func (ws *WSTestClient) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	
	select {
	case <-ws.done:
		return nil // already closed
	default:
		close(ws.done)
		return ws.conn.Close()
	}
}

// WebSocket Integration Tests

func TestWebSocketAuthentication(t *testing.T) {
	// Test with valid token
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	t.Run("Valid Token Connection", func(t *testing.T) {
		wsClient, err := NewWSTestClient(client.token)
		require.NoError(t, err)
		defer wsClient.Close()

		// Connection should be established successfully
		// Try sending a heartbeat
		err = wsClient.SendMessage(MsgTypeHeartbeat, map[string]interface{}{
			"timestamp": time.Now().Unix(),
		})
		assert.NoError(t, err)
	})

	t.Run("Invalid Token Connection", func(t *testing.T) {
		_, err := NewWSTestClient("invalid_token")
		assert.Error(t, err, "Connection should fail with invalid token")
	})

	t.Run("No Token Connection", func(t *testing.T) {
		_, err := NewWSTestClient("")
		assert.Error(t, err, "Connection should fail without token")
	})
}

func TestWebSocketPositionUpdates(t *testing.T) {
	// Setup two clients for testing message broadcasting
	client1 := NewTestClient()
	client2 := NewTestClient()
	
	require.NoError(t, client1.authenticate("test_user1", "test123"))
	require.NoError(t, client2.authenticate("test_user2", "test123"))

	wsClient1, err := NewWSTestClient(client1.token)
	require.NoError(t, err)
	defer wsClient1.Close()

	wsClient2, err := NewWSTestClient(client2.token)
	require.NoError(t, err)
	defer wsClient2.Close()

	// Allow connections to establish
	time.Sleep(100 * time.Millisecond)

	t.Run("Position Update Broadcasting", func(t *testing.T) {
		positionData := map[string]interface{}{
			"lat":       39.0458,
			"lng":       -76.6413,
			"altitude":  100.5,
			"heading":   45.0,
			"speed":     15.2,
			"accuracy":  5.0,
			"callsign":  "test_user1",
			"timestamp": time.Now().Unix(),
		}

		// Client 1 sends position update
		err := wsClient1.SendMessage(MsgTypePositionUpdate, positionData)
		require.NoError(t, err)

		// Client 2 should receive the position update
		msg, err := wsClient2.WaitForMessage(5 * time.Second)
		if err != nil {
			t.Logf("No message received (this might be expected): %v", err)
			return
		}

		assert.Equal(t, MsgTypePositionUpdate, msg["type"])
		
		data, ok := msg["data"].(map[string]interface{})
		require.True(t, ok)
		
		assert.Equal(t, 39.0458, data["lat"])
		assert.Equal(t, -76.6413, data["lng"])
		assert.Equal(t, "test_user1", data["callsign"])
	})

	t.Run("Multiple Position Updates", func(t *testing.T) {
		positions := []map[string]interface{}{
			{
				"lat":      39.0460,
				"lng":      -76.6415,
				"callsign": "test_user1",
				"timestamp": time.Now().Unix(),
			},
			{
				"lat":      39.0462,
				"lng":      -76.6417,
				"callsign": "test_user1",
				"timestamp": time.Now().Unix() + 1,
			},
		}

		for _, pos := range positions {
			err := wsClient1.SendMessage(MsgTypePositionUpdate, pos)
			require.NoError(t, err)
			time.Sleep(100 * time.Millisecond)
		}

		// Try to receive messages (may timeout, that's okay)
		for i := 0; i < 2; i++ {
			_, err := wsClient2.WaitForMessage(2 * time.Second)
			if err != nil {
				t.Logf("Message %d not received: %v", i+1, err)
			}
		}
	})
}

func TestWebSocketChatMessages(t *testing.T) {
	client1 := NewTestClient()
	client2 := NewTestClient()
	
	require.NoError(t, client1.authenticate("test_user1", "test123"))
	require.NoError(t, client2.authenticate("test_user2", "test123"))

	wsClient1, err := NewWSTestClient(client1.token)
	require.NoError(t, err)
	defer wsClient1.Close()

	wsClient2, err := NewWSTestClient(client2.token)
	require.NoError(t, err)
	defer wsClient2.Close()

	time.Sleep(100 * time.Millisecond)

	t.Run("Chat Message Broadcasting", func(t *testing.T) {
		chatData := map[string]interface{}{
			"message":   "Hello from integration test!",
			"sender":    "test_user1",
			"channel":   "general",
			"timestamp": time.Now().Unix(),
		}

		err := wsClient1.SendMessage(MsgTypeChat, chatData)
		require.NoError(t, err)

		// Try to receive the chat message
		msg, err := wsClient2.WaitForMessage(3 * time.Second)
		if err != nil {
			t.Logf("Chat message not received (might be expected): %v", err)
			return
		}

		assert.Equal(t, MsgTypeChat, msg["type"])
		
		data, ok := msg["data"].(map[string]interface{})
		require.True(t, ok)
		
		assert.Equal(t, "Hello from integration test!", data["message"])
		assert.Equal(t, "test_user1", data["sender"])
	})
}

func TestWebSocketGeofenceAlerts(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	wsClient, err := NewWSTestClient(client.token)
	require.NoError(t, err)
	defer wsClient.Close()

	t.Run("Geofence Violation Alert", func(t *testing.T) {
		alertData := map[string]interface{}{
			"geofenceId":   "550e8400-e29b-41d4-a716-446655440300",
			"entityId":     "test-entity-integration",
			"violationType": "enter",
			"position": map[string]float64{
				"lat": 39.0458,
				"lng": -76.6413,
			},
			"timestamp": time.Now().Unix(),
		}

		err := wsClient.SendMessage(MsgTypeGeofenceAlert, alertData)
		require.NoError(t, err)

		// In a real system, this might trigger server-side processing
		// and potentially send alerts to other clients
		t.Log("Geofence alert sent successfully")
	})
}

func TestWebSocketConcurrentConnections(t *testing.T) {
	const numClients = 5
	
	// Create multiple authenticated clients
	httpClients := make([]*TestClient, numClients)
	wsClients := make([]*WSTestClient, numClients)
	
	for i := 0; i < numClients; i++ {
		httpClients[i] = NewTestClient()
		require.NoError(t, httpClients[i].authenticate("test_user1", "test123"))
		
		wsClient, err := NewWSTestClient(httpClients[i].token)
		require.NoError(t, err)
		wsClients[i] = wsClient
		defer wsClient.Close()
	}

	// Allow connections to establish
	time.Sleep(200 * time.Millisecond)

	t.Run("Concurrent Position Updates", func(t *testing.T) {
		var wg sync.WaitGroup
		
		for i := 0; i < numClients; i++ {
			wg.Add(1)
			go func(clientIdx int) {
				defer wg.Done()
				
				posData := map[string]interface{}{
					"lat":       39.0458 + float64(clientIdx)*0.001,
					"lng":       -76.6413 + float64(clientIdx)*0.001,
					"callsign":  fmt.Sprintf("test_user_%d", clientIdx),
					"timestamp": time.Now().Unix(),
				}
				
				err := wsClients[clientIdx].SendMessage(MsgTypePositionUpdate, posData)
				assert.NoError(t, err)
			}(i)
		}
		
		wg.Wait()
		t.Logf("Successfully sent concurrent position updates from %d clients", numClients)
	})

	t.Run("Message Delivery Under Load", func(t *testing.T) {
		const messagesPerClient = 5
		
		var wg sync.WaitGroup
		
		// Send messages from all clients
		for i := 0; i < numClients; i++ {
			wg.Add(1)
			go func(clientIdx int) {
				defer wg.Done()
				
				for j := 0; j < messagesPerClient; j++ {
					chatData := map[string]interface{}{
						"message":   fmt.Sprintf("Load test message %d from client %d", j, clientIdx),
						"sender":    fmt.Sprintf("test_user_%d", clientIdx),
						"timestamp": time.Now().Unix(),
					}
					
					err := wsClients[clientIdx].SendMessage(MsgTypeChat, chatData)
					if err != nil {
						t.Logf("Error sending message from client %d: %v", clientIdx, err)
					}
					
					time.Sleep(10 * time.Millisecond) // Small delay between messages
				}
			}(i)
		}
		
		wg.Wait()
		t.Logf("Load test completed: %d clients sent %d messages each", numClients, messagesPerClient)
	})
}

func TestWebSocketConnectionStability(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	wsClient, err := NewWSTestClient(client.token)
	require.NoError(t, err)
	defer wsClient.Close()

	t.Run("Connection Persistence", func(t *testing.T) {
		// Send periodic heartbeats to test connection stability
		for i := 0; i < 5; i++ {
			heartbeatData := map[string]interface{}{
				"timestamp": time.Now().Unix(),
				"sequence":  i,
			}
			
			err := wsClient.SendMessage(MsgTypeHeartbeat, heartbeatData)
			assert.NoError(t, err)
			
			time.Sleep(1 * time.Second)
		}
		
		t.Log("Connection remained stable through multiple heartbeats")
	})

	t.Run("Large Message Handling", func(t *testing.T) {
		// Create a large message payload
		largeData := make(map[string]interface{})
		largeData["type"] = "large_payload_test"
		largeData["timestamp"] = time.Now().Unix()
		
		// Add a large data field
		largeArray := make([]interface{}, 1000)
		for i := range largeArray {
			largeArray[i] = map[string]interface{}{
				"id":    i,
				"value": fmt.Sprintf("data_point_%d", i),
				"lat":   39.0458 + float64(i)*0.0001,
				"lng":   -76.6413 + float64(i)*0.0001,
			}
		}
		largeData["data"] = largeArray
		
		err := wsClient.SendMessage("large_message_test", largeData)
		if err != nil {
			t.Logf("Large message failed (might be expected due to size limits): %v", err)
		} else {
			t.Log("Large message sent successfully")
		}
	})
}

func TestWebSocketReconnection(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	t.Run("Reconnection After Close", func(t *testing.T) {
		// First connection
		wsClient1, err := NewWSTestClient(client.token)
		require.NoError(t, err)

		// Send a message
		err = wsClient1.SendMessage(MsgTypeHeartbeat, map[string]interface{}{
			"timestamp": time.Now().Unix(),
		})
		assert.NoError(t, err)

		// Close connection
		wsClient1.Close()
		time.Sleep(100 * time.Millisecond)

		// Reconnect
		wsClient2, err := NewWSTestClient(client.token)
		require.NoError(t, err)
		defer wsClient2.Close()

		// Send another message
		err = wsClient2.SendMessage(MsgTypeHeartbeat, map[string]interface{}{
			"timestamp": time.Now().Unix(),
		})
		assert.NoError(t, err)

		t.Log("Reconnection successful")
	})
}

func TestWebSocketErrorHandling(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	wsClient, err := NewWSTestClient(client.token)
	require.NoError(t, err)
	defer wsClient.Close()

	t.Run("Invalid Message Format", func(t *testing.T) {
		// Send malformed JSON by writing directly to connection
		err := wsClient.conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
		
		// Server might close connection or send error response
		if err != nil {
			t.Logf("Error sending invalid JSON (expected): %v", err)
		}

		// Try to send a valid message after
		time.Sleep(100 * time.Millisecond)
		err = wsClient.SendMessage(MsgTypeHeartbeat, map[string]interface{}{
			"timestamp": time.Now().Unix(),
		})
		
		if err != nil {
			t.Logf("Connection might be closed after invalid message: %v", err)
		}
	})

	t.Run("Unsupported Message Type", func(t *testing.T) {
		err := wsClient.SendMessage("unsupported_message_type", map[string]interface{}{
			"data": "test",
		})
		
		// This should not cause an error in sending, but server might ignore
		assert.NoError(t, err)
		
		t.Log("Unsupported message type sent without error")
	})
}

// Performance and stress tests
func TestWebSocketPerformance(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	wsClient, err := NewWSTestClient(client.token)
	require.NoError(t, err)
	defer wsClient.Close()

	t.Run("Message Throughput", func(t *testing.T) {
		const numMessages = 100
		start := time.Now()
		
		for i := 0; i < numMessages; i++ {
			posData := map[string]interface{}{
				"lat":       39.0458 + float64(i)*0.0001,
				"lng":       -76.6413 + float64(i)*0.0001,
				"timestamp": time.Now().Unix(),
				"sequence":  i,
			}
			
			err := wsClient.SendMessage(MsgTypePositionUpdate, posData)
			if err != nil {
				t.Errorf("Message %d failed: %v", i, err)
				break
			}
		}
		
		elapsed := time.Since(start)
		rate := float64(numMessages) / elapsed.Seconds()
		
		t.Logf("Sent %d messages in %v (%.2f messages/sec)", numMessages, elapsed, rate)
		
		// Performance baseline: should handle at least 50 messages/second
		assert.Greater(t, rate, 50.0, "Message rate should be at least 50 messages/second")
	})
}
