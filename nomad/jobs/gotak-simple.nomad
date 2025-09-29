job "gotak-simple" {
  datacenters = ["dc1"]
  type = "service"
  priority = 80

  update {
    max_parallel = 1
    min_healthy_time = "10s"
    healthy_deadline = "3m"
    progress_deadline = "5m"
    auto_revert = false
    stagger = "10s"
  }

  group "web" {
    count = 1

    network {
      port "web" {
        static = 8095
      }
    }

    migrate {
      max_parallel = 1
      health_check = "checks"
      min_healthy_time = "10s"
      healthy_deadline = "3m"
    }

    restart {
      attempts = 3
      interval = "5m"
      delay = "15s"
      mode = "delay"
    }

    task "web" {
      driver = "docker"

      config {
        image = "thefed/gotak-web:20250922-4208282"
        ports = ["web"]
        force_pull = true
        logging {
          type = "json-file"
          config {
            max-size = "10m"
            max-file = "3"
          }
        }
      }

      env {
        # Configure web UI to point to the server API port
        GOTAK_API_URL = "http://hashinuc01:8099"
        GOTAK_WS_URL = "ws://hashinuc01:8099/ws"
        GOTAK_SERVER_URL = "ws://hashinuc01:8099"
        NGINX_PORT = "80"
      }

      service {
        name = "gotak-web-simple"
        port = "web"
        tags = ["gotak", "web", "ui", "simple"]

        check {
          type = "http"
          name = "web-health"
          path = "/"
          interval = "30s"
          timeout = "3s"
        }
      }

      resources {
        cpu = 200
        memory = 256
      }
    }
  }

  group "server" {
    count = 1

    network {
      port "cot_tcp" {
        static = 8096
      }
      port "cot_udp" {
        static = 8097
      }
      port "cot_tls" {
        static = 8098
      }
      port "api" {
        static = 8099
      }
    }

    migrate {
      max_parallel = 1
      health_check = "checks"
      min_healthy_time = "10s"
      healthy_deadline = "3m"
    }

    restart {
      attempts = 2
      interval = "5m"
      delay = "15s"
      mode = "delay"
    }

    task "server" {
      driver = "docker"

      config {
        image = "thefed/gotak-server:20250922-4208282"
        ports = ["cot_tcp", "cot_udp", "cot_tls", "api"]
        force_pull = true
        mount {
          type = "tmpfs"
          target = "/app/logs"
          readonly = false
        }
        mount {
          type = "tmpfs"
          target = "/app/data"
          readonly = false
        }
        logging {
          type = "json-file"
          config {
            max-size = "10m"
            max-file = "3"
          }
        }
      }

      env {
        # Server Configuration
        GOTAK_CONFIG_PATH = "/local/server.yaml"
        GOTAK_LOG_LEVEL = "info"
        GOTAK_LOG_DIR = "/app/logs"
        GOTAK_DATA_DIR = "/app/data"
        GIN_MODE = "release"

        # Minimal database config to pass validation (not used)
        POSTGRES_HOST = "localhost"
        POSTGRES_PORT = "5432"
        POSTGRES_DB = "gotak"
        POSTGRES_USER = "gotak"
        POSTGRES_PASSWORD = "placeholder"

        # Minimal Redis config to pass validation (not used)
        REDIS_HOST = "localhost"
        REDIS_PORT = "6379"
        REDIS_PASSWORD = ""

        # Feature Configuration
        ENABLE_DEBUG = "true"
        ENABLE_METRICS = "true"
        ENABLE_DATABASE = "false"
      }

      template {
        data = <<EOF
# GoTAK Server Configuration - Simplified Deployment
server:
  host: "0.0.0.0"
  http_port: {{ env "NOMAD_PORT_api" }}
  tcp_port: {{ env "NOMAD_PORT_cot_tcp" }}
  udp_port: {{ env "NOMAD_PORT_cot_udp" }}
  tls_port: {{ env "NOMAD_PORT_cot_tls" }}
  serve_static: true
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  max_header_bytes: 1048576

# Disable database for initial deployment
database:
  enabled: false

# Disable Redis for initial deployment  
redis:
  enabled: false

# Security Configuration
security:
  jwt:
    secret: "tactical-jwt-secret-simple-deployment"
    access_token_duration: "24h"
    refresh_token_duration: "168h"
    issuer: "gotak-server"
  
  cors:
    enabled: true
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"]
    allow_headers: ["Authorization", "Content-Type", "Accept", "Origin"]
    allow_credentials: true
    max_age: 3600

  rate_limiting:
    enabled: false

# Logging Configuration
logging:
  level: "{{ env "GOTAK_LOG_LEVEL" }}"
  format: "json"
  output: 
    - type: "stdout"

# TAK Protocol Configuration
tak:
  protocol_version: "2.0"
  max_message_size: 1048576
  heartbeat_interval: "30s"
  client_timeout: "300s"
  max_concurrent_connections: 1000
  
  message_filtering:
    enabled: false
  
  client_auth:
    enabled: false
    require_certificate: false
    certificate_validation: "none"

# Real-time Features
realtime:
  websocket:
    enabled: true
    path: "/ws"
    read_buffer_size: 4096
    write_buffer_size: 4096
    max_message_size: 1048576
    ping_period: "54s"
    pong_wait: "60s"
    write_wait: "10s"
    max_connections_per_ip: 10
  
  broadcasting:
    enabled: true
    buffer_size: 100
    workers: 2
    batch_size: 25
    flush_interval: "100ms"

# Performance Configuration - Minimal for simple deployment
performance:
  connection_pooling:
    enabled: false
  
  caching:
    enabled: false
  
  background_jobs:
    enabled: false

# Monitoring and Health Checks
monitoring:
  metrics:
    enabled: {{ env "ENABLE_METRICS" }}
    endpoint: "/metrics"
    namespace: "gotak"
  
  health:
    enabled: true
    endpoint: "/health"
    checks: []

# Feature Flags - Minimal for initial deployment
features:
  user_registration: false
  password_reset: false
  admin_api: true
  bulk_operations: false
  export_data: false
  import_data: false
  advanced_search: false
  audit_logging: false
  backup_restore: false

# Environment-specific settings
environment: "development"
debug: {{ env "ENABLE_DEBUG" }}
profiling: false
EOF
        destination = "local/server.yaml"
        change_mode = "restart"
      }

      service {
        name = "gotak-api-simple"
        port = "api"
        tags = ["gotak", "api", "http", "simple"]

        check {
          type = "http"
          name = "api-health"
          path = "/health"
          interval = "30s"
          timeout = "5s"
        }

        check {
          type = "http"
          name = "api-metrics"
          path = "/metrics"
          interval = "60s"
          timeout = "3s"
        }
      }

      service {
        name = "gotak-cot-tcp-simple"
        port = "cot_tcp"
        tags = ["gotak", "cot", "tcp", "simple"]

        check {
          type = "tcp"
          name = "cot-tcp"
          interval = "30s"
          timeout = "3s"
        }
      }

      service {
        name = "gotak-cot-tls-simple"
        port = "cot_tls"
        tags = ["gotak", "cot", "tls", "simple"]

        check {
          type = "tcp"
          name = "cot-tls"
          interval = "30s"
          timeout = "3s"
        }
      }

      resources {
        cpu = 400
        memory = 512
      }
    }
  }
}