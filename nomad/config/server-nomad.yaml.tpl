# GoTAK Server Configuration for Nomad Deployment
# This template is rendered by Nomad template stanza

server:
  # Server address and port
  address: "0.0.0.0"
  port: {{ env "NOMAD_PORT_api" }}
  
  # Read timeout for HTTP requests
  read_timeout: 30s
  
  # Write timeout for HTTP responses  
  write_timeout: 30s
  
  # Idle timeout for keep-alive connections
  idle_timeout: 120s
  
  # Maximum header size
  max_header_bytes: 1048576

# TAK server configuration
tak:
  # TCP listener for TAK clients
  tcp:
    enabled: true
    address: "0.0.0.0"
    port: {{ env "NOMAD_PORT_cot" }}
    
  # UDP listener for TAK clients
  udp:
    enabled: true
    address: "0.0.0.0"
    port: {{ env "NOMAD_PORT_cot_udp" }}
    
  # TLS listener for secure TAK clients
  tls:
    enabled: {{ env "ENABLE_TLS" | toBool }}
    address: "0.0.0.0"
    port: {{ env "NOMAD_PORT_cot_tls" }}
    cert_file: "/app/certs/server.crt"
    key_file: "/app/certs/server.key"
    
  # Client settings
  max_clients: 1000
  client_timeout: 300s
  heartbeat_interval: 30s
  message_buffer_size: 1000
  max_message_size: 8192

# Database configuration
database:
  # PostgreSQL connection using Consul DNS
  host: "{{ range service "postgres" }}{{ .Address }}{{ end }}"
  port: {{ range service "postgres" }}{{ .Port }}{{ end }}
  user: "{{ env "DB_USER" }}"
  password: "{{ env "DB_PASSWORD" }}"
  name: "{{ env "DB_NAME" }}"
  sslmode: "disable"
  
  # Connection pool settings
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300s
  conn_max_idle_time: 30s

# Redis configuration
redis:
  # Redis connection using Consul DNS
  host: "{{ range service "redis" }}{{ .Address }}{{ end }}"
  port: {{ range service "redis" }}{{ .Port }}{{ end }}
  password: "{{ env "REDIS_PASSWORD" }}"
  db: 0
  
  # Connection pool settings
  pool_size: 10
  min_idle_conns: 5
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
  pool_timeout: 4s
  idle_timeout: 300s

# Logging configuration
logging:
  level: "{{ env "LOG_LEVEL" }}"
  format: "json"
  output: "/app/logs/gotak.log"
  max_size: 100
  max_age: 30
  max_backups: 10
  compress: true
  local_time: true

# Metrics and monitoring
metrics:
  enabled: {{ env "ENABLE_METRICS" | toBool }}
  path: "/metrics"

# Security settings
security:
  # CORS settings
  cors:
    enabled: true
    allowed_origins: 
      - "*"  # In production, restrict this
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["*"]
    allow_credentials: true
    max_age: 86400
    
  # Rate limiting
  rate_limit:
    enabled: true
    requests_per_minute: 1000
    burst: 100

# Feature flags
features:
  web_ui: true
  api_docs: true
  metrics: {{ env "ENABLE_METRICS" | toBool }}
  auth: {{ env "ENABLE_AUTH" | toBool }}
  
# Development specific settings
development:
  debug: {{ env "ENABLE_DEBUG" | toBool }}