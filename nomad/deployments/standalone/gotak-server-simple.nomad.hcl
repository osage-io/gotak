job "gotak-server" {
  datacenters = ["dc1"]
  region      = "global"
  type        = "service"
  priority    = 90

  group "server" {
    count = 1

    update {
      max_parallel     = 1
      min_healthy_time = "30s"
      healthy_deadline = "5m"
      auto_revert      = true
    }

    restart {
      attempts = 3
      interval = "10m"
      delay    = "15s"
      mode     = "fail"
    }

    network {
      port "cot" {
        static = 8087
        to     = 8087
      }

      port "cot_udp" {
        static = 8087
        to     = 8087
      }

      port "cot_tls" {
        static = 8089
        to     = 8089
      }

      port "api" {
        static = 8080
        to     = 8080
      }
    }

    task "server" {
      driver = "docker"

      config {
        image = "gotak/server:latest"
        ports = ["cot", "cot_udp", "cot_tls", "api"]
        
        environment = {
          GIN_MODE               = "release"
          GOTAK_CONFIG_PATH     = "/local/server.yaml"
          GOTAK_LOG_LEVEL       = "debug"
          GOTAK_DATA_DIR        = "/app/data"
          GOTAK_LOG_DIR         = "/app/logs"
          
          # Database (hardcoded for simplicity)
          DB_USER               = "gotak"
          DB_NAME               = "gotak_dev"
          DB_PASSWORD           = "tactical_secure_pass"
          
          # Redis
          REDIS_PASSWORD        = "tactical_cache_pass"
          
          # Feature toggles
          ENABLE_TLS            = "false"
          ENABLE_DEBUG          = "true"
          ENABLE_METRICS        = "true"
          ENABLE_AUTH           = "false"
          LOG_LEVEL             = "debug"
        }
      }

      template {
        data = <<-EOT
# GoTAK Server Configuration for Nomad Deployment
server:
  address: "0.0.0.0"
  port: {{ env "NOMAD_PORT_api" }}
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_header_bytes: 1048576

tak:
  tcp:
    enabled: true
    address: "0.0.0.0"
    port: {{ env "NOMAD_PORT_cot" }}
  udp:
    enabled: true
    address: "0.0.0.0"
    port: {{ env "NOMAD_PORT_cot_udp" }}
  tls:
    enabled: false
    address: "0.0.0.0"
    port: {{ env "NOMAD_PORT_cot_tls" }}
    cert_file: "/app/certs/server.crt"
    key_file: "/app/certs/server.key"
  max_clients: 1000
  client_timeout: 300s
  heartbeat_interval: 30s
  message_buffer_size: 1000
  max_message_size: 8192

database:
  host: "127.0.0.1"
  port: 5432
  user: "{{ env "DB_USER" }}"
  password: "{{ env "DB_PASSWORD" }}"
  name: "{{ env "DB_NAME" }}"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300s
  conn_max_idle_time: 30s

redis:
  host: "127.0.0.1"
  port: 6379
  password: "{{ env "REDIS_PASSWORD" }}"
  db: 0
  pool_size: 10
  min_idle_conns: 5
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
  pool_timeout: 4s
  idle_timeout: 300s

logging:
  level: "{{ env "LOG_LEVEL" }}"
  format: "json"
  output: "/app/logs/gotak.log"
  max_size: 100
  max_age: 30
  max_backups: 10
  compress: true
  local_time: true

metrics:
  enabled: {{ env "ENABLE_METRICS" }}
  path: "/metrics"

security:
  cors:
    enabled: true
    allowed_origins: 
      - "*"
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["*"]
    allow_credentials: true
    max_age: 86400
  rate_limit:
    enabled: true
    requests_per_minute: 1000
    burst: 100

features:
  web_ui: true
  api_docs: true
  metrics: {{ env "ENABLE_METRICS" }}
  auth: {{ env "ENABLE_AUTH" }}
  
development:
  debug: {{ env "ENABLE_DEBUG" }}
EOT
        destination = "local/server.yaml"
        change_mode = "restart"
      }

      resources {
        cpu    = 200
        memory = 256
      }

      # No service registration in standalone mode
      # Services communicate via direct IP addressing
    }
  }
}