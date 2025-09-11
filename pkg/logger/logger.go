package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog with additional functionality
type Logger struct {
	zerolog.Logger
}

// Config represents logger configuration
type Config struct {
	Level      string `yaml:"level" mapstructure:"level" json:"level"`
	Format     string `yaml:"format" mapstructure:"format" json:"format"` // "json" or "console"
	Output     string `yaml:"output" mapstructure:"output" json:"output"` // "stdout", "stderr", or file path
	Service    string `yaml:"service" mapstructure:"service" json:"service"`
	Version    string `yaml:"version" mapstructure:"version" json:"version"`
	TimeFormat string `yaml:"time_format" mapstructure:"time_format" json:"time_format"`
}

// New creates a new Logger instance with the given configuration
func New(config Config) *Logger {
	// Configure zerolog global settings
	zerolog.TimeFieldFormat = getTimeFormat(config.TimeFormat)
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"
	zerolog.TimestampFieldName = "timestamp"

	// Set global level
	level := parseLevel(config.Level)
	zerolog.SetGlobalLevel(level)

	// Configure output writer
	var output io.Writer
	switch config.Output {
	case "", "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// For file output, we would open the file here
		// For now, default to stdout
		output = os.Stdout
	}

	// Configure formatting
	if config.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Create base logger
	baseLogger := zerolog.New(output).
		With().
		Timestamp().
		Str("service", config.Service).
		Str("version", config.Version).
		Logger()

	return &Logger{Logger: baseLogger}
}

// NewDefault creates a logger with sensible defaults for development
func NewDefault() *Logger {
	config := Config{
		Level:      "debug",
		Format:     "console",
		Output:     "stdout",
		Service:    "gotak",
		Version:    "dev",
		TimeFormat: "rfc3339",
	}
	return New(config)
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.Logger.With()

	// Add common context values if they exist
	if userID := ctx.Value("user_id"); userID != nil {
		logger = logger.Str("user_id", userID.(string))
	}
	if requestID := ctx.Value("request_id"); requestID != nil {
		logger = logger.Str("request_id", requestID.(string))
	}
	if traceID := ctx.Value("trace_id"); traceID != nil {
		logger = logger.Str("trace_id", traceID.(string))
	}

	return &Logger{Logger: logger.Logger()}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	logger := l.Logger.With()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	return &Logger{Logger: logger.Logger()}
}

// WithField returns a logger with a single additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Logger: l.Logger.With().Interface(key, value).Logger()}
}

// Audit logs security and audit events
func (l *Logger) Audit(event string, userID string, resource string, action string, result string) {
	l.Info().
		Str("event_type", "audit").
		Str("event", event).
		Str("user_id", userID).
		Str("resource", resource).
		Str("action", action).
		Str("result", result).
		Msg("audit event")
}

// Security logs security-related events
func (l *Logger) Security(event string, details map[string]interface{}) {
	evt := l.Warn().
		Str("event_type", "security").
		Str("event", event)

	for k, v := range details {
		evt = evt.Interface(k, v)
	}

	evt.Msg("security event")
}

// Performance logs performance metrics
func (l *Logger) Performance(operation string, duration time.Duration, details map[string]interface{}) {
	evt := l.Info().
		Str("event_type", "performance").
		Str("operation", operation).
		Dur("duration", duration)

	for k, v := range details {
		evt = evt.Interface(k, v)
	}

	evt.Msg("performance metric")
}

// HTTP logs HTTP request/response information
func (l *Logger) HTTP(method, path string, statusCode int, duration time.Duration, clientIP string) {
	l.Info().
		Str("event_type", "http").
		Str("method", method).
		Str("path", path).
		Int("status_code", statusCode).
		Dur("duration", duration).
		Str("client_ip", clientIP).
		Msg("http request")
}

// Database logs database operations
func (l *Logger) Database(operation string, table string, duration time.Duration, err error) {
	logEvent := l.Logger.With().
		Str("event_type", "database").
		Str("operation", operation).
		Str("table", table).
		Dur("duration", duration)

	if err != nil {
		logEvent = logEvent.AnErr("error", err)
		l.Error().
			Str("event_type", "database").
			Str("operation", operation).
			Str("table", table).
			Dur("duration", duration).
			Err(err).
			Msg("database operation failed")
	} else {
		l.Debug().
			Str("event_type", "database").
			Str("operation", operation).
			Str("table", table).
			Dur("duration", duration).
			Msg("database operation")
	}
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch level {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// getTimeFormat returns the time format string
func getTimeFormat(format string) string {
	switch format {
	case "rfc3339":
		return time.RFC3339
	case "rfc3339nano":
		return time.RFC3339Nano
	case "unix":
		return zerolog.TimeFormatUnix
	case "unixms":
		return zerolog.TimeFormatUnixMs
	case "unixmicro":
		return zerolog.TimeFormatUnixMicro
	default:
		return time.RFC3339Nano
	}
}

// Global logger instance for convenience
var global *Logger

// Initialize initializes the global logger with the given configuration
func Initialize(config Config) {
	global = New(config)
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	global = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	if global == nil {
		global = NewDefault()
	}
	return global
}

// Convenience functions that use the global logger
func Debug() *zerolog.Event    { return GetGlobalLogger().Debug() }
func Info() *zerolog.Event     { return GetGlobalLogger().Info() }
func Warn() *zerolog.Event     { return GetGlobalLogger().Warn() }
func Error() *zerolog.Event    { return GetGlobalLogger().Error() }
func Fatal() *zerolog.Event    { return GetGlobalLogger().Fatal() }
func Panic() *zerolog.Event    { return GetGlobalLogger().Panic() }
func With() zerolog.Context    { return GetGlobalLogger().With() }
func Ctx(ctx context.Context) *Logger { return GetGlobalLogger().WithContext(ctx) }
