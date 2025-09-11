package logger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "console logger",
			config: Config{
				Level:      "debug",
				Format:     "console",
				Output:     "stdout",
				Service:    "test-service",
				Version:    "1.0.0",
				TimeFormat: "rfc3339",
			},
		},
		{
			name: "json logger",
			config: Config{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				Service:    "test-service",
				Version:    "1.0.0",
				TimeFormat: "rfc3339nano",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.config)
			assert.NotNil(t, logger)
		})
	}
}

func TestNewDefault(t *testing.T) {
	logger := NewDefault()
	assert.NotNil(t, logger)
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"warning", "warn"},
		{"error", "error"},
		{"fatal", "fatal"},
		{"panic", "panic"},
		{"unknown", "info"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := parseLevel(tt.input)
			assert.Equal(t, tt.want, level.String())
		})
	}
}

func TestWithContext(t *testing.T) {
	logger := NewDefault()

	// Test with context values
	ctx := context.WithValue(context.Background(), "user_id", "test-user-123")
	ctx = context.WithValue(ctx, "request_id", "req-456")
	ctx = context.WithValue(ctx, "trace_id", "trace-789")

	contextLogger := logger.WithContext(ctx)
	assert.NotNil(t, contextLogger)

	// Test without context values
	emptyCtx := context.Background()
	emptyContextLogger := logger.WithContext(emptyCtx)
	assert.NotNil(t, emptyContextLogger)
}

func TestWithFields(t *testing.T) {
	logger := NewDefault()

	fields := map[string]interface{}{
		"component": "test",
		"operation": "test-operation",
		"count":     42,
	}

	fieldsLogger := logger.WithFields(fields)
	assert.NotNil(t, fieldsLogger)
}

func TestWithField(t *testing.T) {
	logger := NewDefault()

	fieldLogger := logger.WithField("test_key", "test_value")
	assert.NotNil(t, fieldLogger)
}

func TestAudit(t *testing.T) {
	logger := NewDefault()
	logger.Audit("test_event", "user-123", "resource", "action", "success")

	// Test doesn't panic - that's our main assertion
	assert.True(t, true)
}

func TestSecurity(t *testing.T) {
	logger := NewDefault()

	details := map[string]interface{}{
		"ip_address": "192.168.1.1",
		"user_agent": "test-agent",
	}

	// Test doesn't panic
	logger.Security("failed_login", details)
	assert.True(t, true)
}

func TestPerformance(t *testing.T) {
	logger := NewDefault()

	details := map[string]interface{}{
		"query_count": 5,
		"cache_hit":   true,
	}

	duration := 150 * time.Millisecond

	// Test doesn't panic
	logger.Performance("database_query", duration, details)
	assert.True(t, true)
}

func TestHTTP(t *testing.T) {
	logger := NewDefault()

	// Test doesn't panic
	logger.HTTP("GET", "/api/test", 200, 50*time.Millisecond, "192.168.1.1")
	assert.True(t, true)
}

func TestDatabase(t *testing.T) {
	logger := NewDefault()

	// Test successful operation
	logger.Database("SELECT", "users", 10*time.Millisecond, nil)
	assert.True(t, true)

	// Test failed operation
	err := assert.AnError
	logger.Database("INSERT", "users", 5*time.Millisecond, err)
	assert.True(t, true)
}

func TestGlobalLogger(t *testing.T) {
	// Test setting and getting global logger
	testLogger := NewDefault()
	SetGlobalLogger(testLogger)

	retrieved := GetGlobalLogger()
	assert.Equal(t, testLogger, retrieved)

	// Test global convenience functions don't panic
	Debug().Msg("debug message")
	Info().Msg("info message")
	Warn().Msg("warn message")
	Error().Msg("error message")

	// Test context function
	ctx := context.WithValue(context.Background(), "user_id", "test-user")
	ctxLogger := Ctx(ctx)
	assert.NotNil(t, ctxLogger)
}

func TestGetTimeFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"rfc3339", time.RFC3339},
		{"rfc3339nano", time.RFC3339Nano},
		{"unknown", time.RFC3339Nano},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := getTimeFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test that unix formats return something (don't test exact values)
	assert.NotPanics(t, func() {
		getTimeFormat("unix")
		getTimeFormat("unixms")
		getTimeFormat("unixmicro")
	})
}

// Integration test to ensure logging works end-to-end
func TestLoggingIntegration(t *testing.T) {
	// This is more of a smoke test to ensure the logger setup works
	logger := NewDefault()

	// Test various log levels and methods
	logger.Debug().Str("component", "test").Msg("Debug message")
	logger.Info().Int("count", 42).Msg("Info message")
	logger.Warn().Bool("warning", true).Msg("Warning message")
	logger.Error().Str("error", "test error").Msg("Error message")

	// Test structured logging methods
	logger.Audit("user_login", "user-123", "auth", "login", "success")
	logger.Security("brute_force", map[string]interface{}{
		"attempts": 5,
		"blocked":  true,
	})
	logger.Performance("api_request", 100*time.Millisecond, map[string]interface{}{
		"endpoint": "/api/test",
	})
	logger.HTTP("POST", "/api/login", 200, 250*time.Millisecond, "10.0.0.1")
	logger.Database("SELECT", "users", 15*time.Millisecond, nil)

	// If we get here without panicking, the integration test passes
	assert.True(t, true)
}
