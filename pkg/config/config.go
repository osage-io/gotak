package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ServerConfig holds the complete server configuration
type ServerConfig struct {
	Server   ServerSettings   `yaml:"server"`
	Database DatabaseSettings `yaml:"database"`
	Security SecuritySettings `yaml:"security"`
	Logging  LoggingSettings  `yaml:"logging"`
	TAK      TAKSettings      `yaml:"tak"`
}

// ServerSettings contains basic server configuration
type ServerSettings struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	
	// TAK-specific ports
	TCPPort int `yaml:"tcp_port"`
	UDPPort int `yaml:"udp_port"`
	TLSPort int `yaml:"tls_port"`
	
	// Web interface
	HTTPPort   int  `yaml:"http_port"`
	WebPort    int  `yaml:"web_port"`
	WebEnabled bool `yaml:"web_enabled"`
	ServeStatic bool `yaml:"serve_static"`
	
	// Performance settings
	MaxConnections     int           `yaml:"max_connections"`
	ReadTimeout        time.Duration `yaml:"read_timeout"`
	WriteTimeout       time.Duration `yaml:"write_timeout"`
	IdleTimeout        time.Duration `yaml:"idle_timeout"`
	KeepAliveInterval  time.Duration `yaml:"keepalive_interval"`
}

// DatabaseSettings contains database configuration
type DatabaseSettings struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
	
	// Connection pool settings
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// SecuritySettings contains security and authentication configuration
type SecuritySettings struct {
	// TLS Configuration
	TLSEnabled   bool   `yaml:"tls_enabled"`
	CertFile     string `yaml:"cert_file"`
	KeyFile      string `yaml:"key_file"`
	CAFile       string `yaml:"ca_file"`
	
	// Client certificate authentication
	ClientAuthRequired bool `yaml:"client_auth_required"`
	
	// JWT settings for web interface
	JWTSecret     string        `yaml:"jwt_secret"`
	JWTExpiration time.Duration `yaml:"jwt_expiration"`
	
	// Password policy
	MinPasswordLength int  `yaml:"min_password_length"`
	RequireUppercase  bool `yaml:"require_uppercase"`
	RequireNumbers    bool `yaml:"require_numbers"`
	RequireSymbols    bool `yaml:"require_symbols"`
}

// LoggingSettings contains logging configuration
type LoggingSettings struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

// TAKSettings contains TAK-specific configuration
type TAKSettings struct {
	// Server identification
	ServerName string `yaml:"server_name"`
	ServerUID  string `yaml:"server_uid"`
	
	// Message settings
	MaxMessageSize    int           `yaml:"max_message_size"`
	MessageTimeout    time.Duration `yaml:"message_timeout"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
	
	// CoT message settings
	DefaultStale      time.Duration `yaml:"default_stale"`
	MaxStale          time.Duration `yaml:"max_stale"`
	EnablePersistence bool          `yaml:"enable_persistence"`
	
	// Group settings
	DefaultGroup      string   `yaml:"default_group"`
	AllowedGroups     []string `yaml:"allowed_groups"`
	
	// Federation settings
	Federation FederationSettings `yaml:"federation"`
}

// FederationSettings contains federation configuration
type FederationSettings struct {
	Enabled bool                  `yaml:"enabled"`
	Peers   []FederationPeer      `yaml:"peers"`
}

// FederationPeer represents a federated TAK server
type FederationPeer struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Groups   []string `yaml:"groups"`
}

// LoadServerConfig loads server configuration from a YAML file
func LoadServerConfig(path string) (*ServerConfig, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file %s does not exist", path)
	}
	
	// Read file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}
	
	// Parse YAML
	var config ServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}
	
	// Set defaults
	setDefaults(&config)
	
	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// setDefaults sets default values for configuration fields
func setDefaults(config *ServerConfig) {
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8089
	}
	if config.Server.TCPPort == 0 {
		config.Server.TCPPort = 8087
	}
	if config.Server.UDPPort == 0 {
		config.Server.UDPPort = 8087
	}
	if config.Server.TLSPort == 0 {
		config.Server.TLSPort = 8089
	}
	if config.Server.HTTPPort == 0 {
		config.Server.HTTPPort = 8080
	}
	if config.Server.WebPort == 0 {
		config.Server.WebPort = 8080
	}
	if config.Server.MaxConnections == 0 {
		config.Server.MaxConnections = 1000
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30 * time.Second
	}
	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = 120 * time.Second
	}
	if config.Server.KeepAliveInterval == 0 {
		config.Server.KeepAliveInterval = 30 * time.Second
	}
	
	// Database defaults
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.Database == "" {
		config.Database.Database = "gotak"
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 25
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 5
	}
	if config.Database.ConnMaxLifetime == 0 {
		config.Database.ConnMaxLifetime = 5 * time.Minute
	}
	
	// Security defaults
	if config.Security.JWTExpiration == 0 {
		config.Security.JWTExpiration = 24 * time.Hour
	}
	if config.Security.MinPasswordLength == 0 {
		config.Security.MinPasswordLength = 8
	}
	
	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}
	if config.Logging.MaxSize == 0 {
		config.Logging.MaxSize = 100
	}
	if config.Logging.MaxBackups == 0 {
		config.Logging.MaxBackups = 3
	}
	if config.Logging.MaxAge == 0 {
		config.Logging.MaxAge = 28
	}
	
	// TAK defaults
	if config.TAK.ServerName == "" {
		config.TAK.ServerName = "GoTAK-Server"
	}
	if config.TAK.MaxMessageSize == 0 {
		config.TAK.MaxMessageSize = 8192 // 8KB
	}
	if config.TAK.MessageTimeout == 0 {
		config.TAK.MessageTimeout = 30 * time.Second
	}
	if config.TAK.HeartbeatInterval == 0 {
		config.TAK.HeartbeatInterval = 60 * time.Second
	}
	if config.TAK.DefaultStale == 0 {
		config.TAK.DefaultStale = 5 * time.Minute
	}
	if config.TAK.MaxStale == 0 {
		config.TAK.MaxStale = 24 * time.Hour
	}
	if config.TAK.DefaultGroup == "" {
		config.TAK.DefaultGroup = "__ANON__"
	}
}

// validateConfig validates the configuration
func validateConfig(config *ServerConfig) error {
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}
	
	if config.Server.TCPPort < 1 || config.Server.TCPPort > 65535 {
		return fmt.Errorf("invalid TCP port: %d", config.Server.TCPPort)
	}
	
	if config.Server.UDPPort < 1 || config.Server.UDPPort > 65535 {
		return fmt.Errorf("invalid UDP port: %d", config.Server.UDPPort)
	}
	
	if config.Security.TLSEnabled {
		if config.Security.CertFile == "" {
			return fmt.Errorf("TLS enabled but cert_file not specified")
		}
		if config.Security.KeyFile == "" {
			return fmt.Errorf("TLS enabled but key_file not specified")
		}
	}
	
	return nil
}
