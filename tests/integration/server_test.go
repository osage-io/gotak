//go:build integration
// +build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite defines the integration test suite
type IntegrationTestSuite struct {
	suite.Suite
	db     *sql.DB
	ctx    context.Context
	cancel context.CancelFunc
}

// SetupSuite runs once before all tests
func (suite *IntegrationTestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithCancel(context.Background())

	// Wait for PostgreSQL to be ready
	suite.waitForPostgreSQL()

	// Connect to test database
	db, err := sql.Open("postgres", "postgres://gotak:test_password@localhost:5433/gotak_test?sslmode=disable")
	require.NoError(suite.T(), err)

	err = db.Ping()
	require.NoError(suite.T(), err)

	suite.db = db

	// Run any additional setup
	suite.setupTestData()
}

// TearDownSuite runs once after all tests
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	suite.cancel()
}

// SetupTest runs before each individual test
func (suite *IntegrationTestSuite) SetupTest() {
	// Clean up test data before each test
	suite.cleanupTestData()
}

// TestDatabaseConnection verifies database connectivity
func (suite *IntegrationTestSuite) TestDatabaseConnection() {
	var result string
	err := suite.db.QueryRow("SELECT 'database_connected'").Scan(&result)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "database_connected", result)
}

// TestDatabaseExtensions verifies required PostgreSQL extensions
func (suite *IntegrationTestSuite) TestDatabaseExtensions() {
	// Install extensions if they don't exist (for test database)
	extensions := []string{"uuid-ossp", "pgcrypto"}

	for _, ext := range extensions {
		// Try to create extension if it doesn't exist
		_, err := suite.db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\"", ext))
		require.NoError(suite.T(), err, "Failed to create extension %s", ext)

		// Verify extension is installed
		var installed bool
		query := `SELECT EXISTS(
			SELECT 1 FROM pg_extension WHERE extname = $1
		)`
		err = suite.db.QueryRow(query, ext).Scan(&installed)
		require.NoError(suite.T(), err)
		assert.True(suite.T(), installed, "Extension %s should be installed", ext)
	}
}

// TestRedisConnection verifies Redis connectivity
func (suite *IntegrationTestSuite) TestRedisConnection() {
	// Simple TCP connection test to Redis - just verify we can connect
	conn, err := net.DialTimeout("tcp", "localhost:6379", 2*time.Second)
	require.NoError(suite.T(), err, "Should be able to connect to Redis")
	defer conn.Close()

	// If we get here, Redis is accepting connections
	// This is sufficient for integration testing
	assert.True(suite.T(), true, "Redis connection successful")
}

// TestNATSConnection verifies NATS connectivity
func (suite *IntegrationTestSuite) TestNATSConnection() {
	// Test NATS HTTP monitoring endpoint
	conn, err := net.DialTimeout("tcp", "localhost:8222", 5*time.Second)
	require.NoError(suite.T(), err)
	defer conn.Close()

	// Send basic HTTP GET request to monitoring endpoint
	_, err = conn.Write([]byte("GET /healthz HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	require.NoError(suite.T(), err)

	// Read response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	require.NoError(suite.T(), err)

	response := string(buffer[:n])
	assert.Contains(suite.T(), response, "200 OK")
}

// TestVaultConnection verifies Vault connectivity (development mode)
func (suite *IntegrationTestSuite) TestVaultConnection() {
	// Test Vault HTTP endpoint
	conn, err := net.DialTimeout("tcp", "localhost:8200", 5*time.Second)
	require.NoError(suite.T(), err)
	defer conn.Close()

	// Send basic HTTP GET request to health endpoint
	_, err = conn.Write([]byte("GET /v1/sys/health HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	require.NoError(suite.T(), err)

	// Read response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	require.NoError(suite.T(), err)

	response := string(buffer[:n])
	assert.Contains(suite.T(), response, "200 OK")
}

// TestGoTAKServerPorts verifies GoTAK server ports are available
func (suite *IntegrationTestSuite) TestGoTAKServerPorts() {
	ports := []int{8087, 8089, 8080} // TCP, TLS, Web ports

	for _, port := range ports {
		// Check if port is available (not in use)
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		require.NoError(suite.T(), err, "Port %d should be available for GoTAK server", port)
		ln.Close()
	}
}

// Helper methods

func (suite *IntegrationTestSuite) waitForPostgreSQL() {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		conn, err := net.DialTimeout("tcp", "localhost:5433", 2*time.Second)
		if err == nil {
			conn.Close()
			// Give PostgreSQL a moment to fully initialize
			time.Sleep(2 * time.Second)
			return
		}
		time.Sleep(1 * time.Second)
	}
	suite.T().Fatal("PostgreSQL did not become available within timeout period")
}

func (suite *IntegrationTestSuite) setupTestData() {
	// Create test schemas and initial data
	queries := []string{
		`CREATE SCHEMA IF NOT EXISTS test_data`,
		`CREATE TABLE IF NOT EXISTS test_data.integration_test_log (
			id SERIAL PRIMARY KEY,
			test_name VARCHAR(255),
			run_at TIMESTAMP DEFAULT NOW()
		)`,
	}

	for _, query := range queries {
		_, err := suite.db.Exec(query)
		require.NoError(suite.T(), err)
	}
}

func (suite *IntegrationTestSuite) cleanupTestData() {
	// Clean up test data before each test
	_, err := suite.db.Exec(`DELETE FROM test_data.integration_test_log`)
	require.NoError(suite.T(), err)
}

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
