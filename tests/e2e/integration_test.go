package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test configuration
var (
	baseURL    = getEnv("GOTAK_BASE_URL", "http://localhost:8080")
	wsURL      = getEnv("GOTAK_WS_URL", "ws://localhost:8080")
	testUser   = "test_user1"
	testPass   = "test123"
	testGroup  = "test_group"
	adminUser  = "test_admin"
	adminPass  = "admin123"
)

// Test client for HTTP requests
type TestClient struct {
	httpClient *http.Client
	token      string
	baseURL    string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func NewTestClient() *TestClient {
	return &TestClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}
}

func (tc *TestClient) authenticate(username, password string) error {
	loginData := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(loginData)

	resp, err := tc.httpClient.Post(
		tc.baseURL+"/api/auth/login",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	token, ok := result["token"].(string)
	if !ok {
		return fmt.Errorf("no token received")
	}

	tc.token = token
	return nil
}

func (tc *TestClient) request(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := tc.baseURL + endpoint
	var req *http.Request
	var err error

	if reqBody != nil {
		req, err = http.NewRequest(method, url, reqBody)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	if tc.token != "" {
		req.Header.Set("Authorization", "Bearer "+tc.token)
	}

	return tc.httpClient.Do(req)
}

// Test Suite Setup
func TestMain(m *testing.M) {
	// Wait for services to be ready
	if err := waitForServices(); err != nil {
		fmt.Printf("Services not ready: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func waitForServices() error {
	client := &http.Client{Timeout: 5 * time.Second}
	
	// Wait for backend health check
	for i := 0; i < 60; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("backend service not ready")
}

// Integration Tests

func TestAuthentication(t *testing.T) {
	client := NewTestClient()

	t.Run("Valid Login", func(t *testing.T) {
		err := client.authenticate(testUser, testPass)
		assert.NoError(t, err)
		assert.NotEmpty(t, client.token)
	})

	t.Run("Invalid Login", func(t *testing.T) {
		client := NewTestClient()
		err := client.authenticate("invalid_user", "invalid_pass")
		assert.Error(t, err)
	})

	t.Run("Token Validation", func(t *testing.T) {
		resp, err := client.request("GET", "/api/user/profile", nil)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestRoutesAPI(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	t.Run("List Routes", func(t *testing.T) {
		resp, err := client.request("GET", "/api/routes", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var routes []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&routes)
		require.NoError(t, err)
		
		// Should have test routes from seed data
		assert.GreaterOrEqual(t, len(routes), 2)
	})

	t.Run("Create Route", func(t *testing.T) {
		newRoute := map[string]interface{}{
			"name":        "Integration Test Route",
			"description": "Created during integration testing",
			"waypoints": []map[string]float64{
				{"lat": 39.0500, "lng": -76.6500},
				{"lat": 39.0510, "lng": -76.6510},
			},
			"vehicle":   "car",
			"routeType": "fastest",
		}

		resp, err := client.request("POST", "/api/routes", newRoute)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["id"])
		assert.Equal(t, "Integration Test Route", result["name"])
	})

	t.Run("Get Route Details", func(t *testing.T) {
		// First get a route ID from the list
		resp, err := client.request("GET", "/api/routes", nil)
		require.NoError(t, err)

		var routes []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&routes)
		require.NoError(t, err)
		resp.Body.Close()

		if len(routes) > 0 {
			routeID := routes[0]["id"].(string)
			
			resp, err = client.request("GET", "/api/routes/"+routeID, nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var route map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&route)
			require.NoError(t, err)

			assert.Equal(t, routeID, route["id"])
			assert.NotEmpty(t, route["name"])
		}
	})
}

func TestGeofencesAPI(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	t.Run("List Geofences", func(t *testing.T) {
		resp, err := client.request("GET", "/api/geofences", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var geofences []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&geofences)
		require.NoError(t, err)

		// Should have test geofences from seed data
		assert.GreaterOrEqual(t, len(geofences), 2)
	})

	t.Run("Create Geofence", func(t *testing.T) {
		newGeofence := map[string]interface{}{
			"name":        "Test Integration Geofence",
			"description": "Created during integration testing",
			"type":        "circle",
			"geometry": map[string]interface{}{
				"center": map[string]float64{
					"lat": 39.0550,
					"lng": -76.6550,
				},
				"radius": 500,
			},
			"enabled":      true,
			"alertOnEnter": true,
			"alertOnExit":  false,
		}

		resp, err := client.request("POST", "/api/geofences", newGeofence)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["id"])
		assert.Equal(t, "Test Integration Geofence", result["name"])
	})
}

func TestWebSocketConnection(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	t.Run("WebSocket Connection", func(t *testing.T) {
		// Create WebSocket connection URL with auth token
		u, err := url.Parse(wsURL + "/ws")
		require.NoError(t, err)

		header := http.Header{}
		header.Set("Authorization", "Bearer "+client.token)

		dialer := websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		}

		conn, _, err := dialer.Dial(u.String(), header)
		require.NoError(t, err)
		defer conn.Close()

		// Send a test message
		testMsg := map[string]interface{}{
			"type": "position_update",
			"data": map[string]interface{}{
				"lat":      39.0458,
				"lng":      -76.6413,
				"callsign": testUser,
				"timestamp": time.Now().Unix(),
			},
		}

		err = conn.WriteJSON(testMsg)
		assert.NoError(t, err)

		// Set read deadline and try to read response
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		
		var response map[string]interface{}
		err = conn.ReadJSON(&response)
		if err != nil {
			// WebSocket might not echo back immediately, that's okay
			t.Logf("WebSocket read timeout or error (expected): %v", err)
		} else {
			t.Logf("Received WebSocket response: %+v", response)
		}
	})
}

func TestOfflineAreasAPI(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	t.Run("List Offline Areas", func(t *testing.T) {
		resp, err := client.request("GET", "/api/offline-areas", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var areas []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&areas)
		require.NoError(t, err)

		// Should have test areas from seed data
		assert.GreaterOrEqual(t, len(areas), 2)
	})

	t.Run("Create Offline Area", func(t *testing.T) {
		newArea := map[string]interface{}{
			"name": "Integration Test Area",
			"bounds": map[string]float64{
				"north": 39.0600,
				"south": 39.0500,
				"east":  -76.6400,
				"west":  -76.6500,
			},
			"minZoom": 10,
			"maxZoom": 14,
			"layers":  []string{"streets", "satellite"},
		}

		resp, err := client.request("POST", "/api/offline-areas", newArea)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["id"])
		assert.Equal(t, "Integration Test Area", result["name"])
	})
}

func TestTacticalOverlaysAPI(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	t.Run("List Tactical Overlays", func(t *testing.T) {
		resp, err := client.request("GET", "/api/tactical-overlays", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var overlays []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&overlays)
		require.NoError(t, err)

		// Should have test overlays from seed data
		assert.GreaterOrEqual(t, len(overlays), 2)
	})

	t.Run("Create Tactical Overlay", func(t *testing.T) {
		newOverlay := map[string]interface{}{
			"name": "Integration Test Marker",
			"type": "marker",
			"geometry": map[string]interface{}{
				"type":        "Point",
				"coordinates": []float64{-76.6600, 39.0600},
			},
			"style": map[string]interface{}{
				"color": "green",
				"size":  "medium",
				"icon":  "checkpoint",
			},
			"metadata": map[string]interface{}{
				"priority":       "normal",
				"classification": "unclassified",
				"notes":          "Created during integration testing",
			},
		}

		resp, err := client.request("POST", "/api/tactical-overlays", newOverlay)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["id"])
		assert.Equal(t, "Integration Test Marker", result["name"])
	})
}

func TestSystemHealth(t *testing.T) {
	client := NewTestClient()

	t.Run("Health Check", func(t *testing.T) {
		resp, err := client.request("GET", "/health", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)

		assert.Equal(t, "ok", health["status"])
		assert.NotEmpty(t, health["timestamp"])
	})

	t.Run("Database Health", func(t *testing.T) {
		resp, err := client.request("GET", "/health/database", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var dbHealth map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&dbHealth)
		require.NoError(t, err)

		assert.Equal(t, "healthy", dbHealth["status"])
	})
}

func TestErrorHandling(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	t.Run("404 Not Found", func(t *testing.T) {
		resp, err := client.request("GET", "/api/nonexistent", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		unauthorizedClient := NewTestClient()
		resp, err := unauthorizedClient.request("GET", "/api/routes", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Invalid JSON Request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/api/routes", 
			strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+client.token)

		resp, err := client.httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestConcurrentAccess(t *testing.T) {
	const numClients = 5
	const numRequests = 10

	clients := make([]*TestClient, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = NewTestClient()
		require.NoError(t, clients[i].authenticate(testUser, testPass))
	}

	t.Run("Concurrent Route Requests", func(t *testing.T) {
		done := make(chan bool, numClients*numRequests)

		for i := 0; i < numClients; i++ {
			client := clients[i]
			go func() {
				for j := 0; j < numRequests; j++ {
					resp, err := client.request("GET", "/api/routes", nil)
					if err == nil {
						resp.Body.Close()
						assert.Equal(t, http.StatusOK, resp.StatusCode)
					}
					done <- true
				}
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < numClients*numRequests; i++ {
			<-done
		}
	})
}

// Performance baseline tests
func TestPerformanceBaseline(t *testing.T) {
	client := NewTestClient()
	require.NoError(t, client.authenticate(testUser, testPass))

	t.Run("Route List Response Time", func(t *testing.T) {
		start := time.Now()
		resp, err := client.request("GET", "/api/routes", nil)
		elapsed := time.Since(start)

		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Less(t, elapsed, 2*time.Second, "Route list should respond within 2 seconds")

		t.Logf("Route list response time: %v", elapsed)
	})

	t.Run("Authentication Response Time", func(t *testing.T) {
		testClient := NewTestClient()
		
		start := time.Now()
		err := testClient.authenticate(testUser, testPass)
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, elapsed, 1*time.Second, "Authentication should complete within 1 second")

		t.Logf("Authentication response time: %v", elapsed)
	})
}
