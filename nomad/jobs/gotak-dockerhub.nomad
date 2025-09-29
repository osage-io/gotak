job "gotak-dockerhub" {
  datacenters = ["dc1"]
  type = "service"
  priority = 95

  update {
    max_parallel = 1
    min_healthy_time = "30s"
    healthy_deadline = "5m"
    progress_deadline = "10m"
    auto_revert = true
    auto_promote = false
    stagger = "30s"
  }

  group "postgres" {
    count = 1

    constraint {
      attribute = "${node.unique.name}"
      operator = "regexp"
      value = "hashinuc01"
    }

    network {
      port "postgres" {
        static = 5433
      }
    }

    migrate {
      max_parallel = 1
      health_check = "checks"
      min_healthy_time = "10s"
      healthy_deadline = "5m"
    }

    restart {
      attempts = 3
      interval = "5m"
      delay = "25s"
      mode = "delay"
    }

    task "postgres" {
      driver = "docker"

      config {
        image = "postgis/postgis:15-3.4"
        ports = ["postgres"]
        volumes = [
          "local/init.sql:/docker-entrypoint-initdb.d/init.sql"
        ]
        logging {
          type = "json-file"
          config {
            max-size = "10m"
            max-file = "3"
          }
        }
      }

      env {
        POSTGRES_USER = "gotak"
        POSTGRES_PASSWORD = "tactical_secure_pass"
        POSTGRES_DB = "gotak"
        PGDATA = "/var/lib/postgresql/data/pgdata"
      }

      template {
        data = <<EOF
-- Create PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

-- Create tables for GoTAK
CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    callsign VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    location GEOGRAPHY(POINT, 4326),
    altitude DOUBLE PRECISION,
    speed DOUBLE PRECISION,
    course DOUBLE PRECISION,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_entities_location ON entities USING GIST(location);
CREATE INDEX IF NOT EXISTS idx_entities_callsign ON entities(callsign);
CREATE INDEX IF NOT EXISTS idx_entities_updated_at ON entities(updated_at);
EOF
        destination = "local/init.sql"
        change_mode = "restart"
      }

      resources {
        cpu = 200
        memory = 512
      }
    }
  }

  group "redis" {
    count = 1

    constraint {
      attribute = "${node.unique.name}"
      operator = "regexp"
      value = "hashinuc01"
    }

    network {
      port "redis" {
        static = 6380
      }
    }

    migrate {
      max_parallel = 1
      health_check = "checks"
      min_healthy_time = "10s"
      healthy_deadline = "5m"
    }

    restart {
      attempts = 3
      interval = "5m"
      delay = "15s"
      mode = "delay"
    }

    task "redis" {
      driver = "docker"

      config {
        image = "redis:7-alpine"
        ports = ["redis"]
        command = "redis-server"
        args = [
          "--appendonly", "yes",
          "--maxmemory", "512mb",
          "--maxmemory-policy", "allkeys-lru"
        ]
        logging {
          type = "json-file"
          config {
            max-size = "10m"
            max-file = "3"
          }
        }
      }

      resources {
        cpu = 100
        memory = 256
      }
    }
  }

  group "server" {
    count = 1

    constraint {
      attribute = "${node.unique.name}"
      operator = "regexp"
      value = "hashinuc01"
    }

    network {
      port "cot_tcp" {
        static = 8090
      }
      port "cot_udp" {
        static = 8091
      }
      port "cot_tls" {
        static = 8092
      }
      port "api" {
        static = 8093
      }
    }

    migrate {
      max_parallel = 1
      health_check = "checks"
      min_healthy_time = "10s"
      healthy_deadline = "5m"
    }

    restart {
      attempts = 3
      interval = "10m"
      delay = "15s"
      mode = "fail"
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
            max-file = "5"
          }
        }
      }

      constraint {
        attribute = "${attr.consul.version}"
        operator = "semver"
        value = ">= 1.8.0"
      }

      env {
        # Database Configuration
        POSTGRES_HOST = "hashinuc01"
        POSTGRES_PORT = "5433"
        POSTGRES_DB = "gotak"
        POSTGRES_USER = "gotak"
        POSTGRES_PASSWORD = "tactical_secure_pass"

        # Redis Configuration
        REDIS_HOST = "hashinuc01"
        REDIS_PORT = "6380"
        REDIS_PASSWORD = ""

        # Server Configuration
        GOTAK_CONFIG_PATH = "/local/server.yaml"
        GOTAK_LOG_LEVEL = "info"
        GOTAK_LOG_DIR = "/app/logs"
        GOTAK_DATA_DIR = "/app/data"
        GIN_MODE = "release"

        # Security Configuration
        JWT_SECRET = "tactical-jwt-secret-change-in-production-12345"
        ENABLE_TLS = "false"
        ENABLE_AUTH = "false"

        # Feature Configuration
        ENABLE_DEBUG = "false"
        ENABLE_METRICS = "true"
      }

      template {
        data = <<EOF
# GoTAK Server Configuration for Nomad Deployment
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

# Database Configuration
database:
  host: "{{ env "POSTGRES_HOST" }}"
  port: {{ env "POSTGRES_PORT" }}
  name: "{{ env "POSTGRES_DB" }}"
  user: "{{ env "POSTGRES_USER" }}"
  password: "{{ env "POSTGRES_PASSWORD" }}"
  sslmode: "disable"
  timezone: "UTC"
  max_open_connections: 25
  max_idle_connections: 5
  connection_max_lifetime: "300s"
  connection_max_idle_time: "60s"

# Redis Configuration
redis:
  host: "{{ env "REDIS_HOST" }}"
  port: {{ env "REDIS_PORT" }}
  password: "{{ env "REDIS_PASSWORD" }}"
  database: 0
  pool_size: 10
  min_idle_connections: 1
  dial_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
  pool_timeout: "4s"
  idle_timeout: "300s"

# Security Configuration
security:
  jwt:
    secret: "{{ env "JWT_SECRET" }}"
    access_token_duration: "24h"
    refresh_token_duration: "168h"
    issuer: "gotak-server"
  
  tls:
    enabled: {{ env "ENABLE_TLS" }}
    cert_file: "/app/certs/server.crt"
    key_file: "/app/certs/server.key"
    client_auth: "none"
    min_version: "1.2"
  
  cors:
    enabled: true
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"]
    allow_headers: ["Authorization", "Content-Type", "Accept", "Origin"]
    allow_credentials: true
    max_age: 3600

  rate_limiting:
    enabled: true
    requests_per_minute: 1000
    burst: 50
    cleanup_interval: "60s"

# Logging Configuration
logging:
  level: "{{ env "GOTAK_LOG_LEVEL" }}"
  format: "json"
  output: 
    - type: "file"
      path: "{{ env "GOTAK_LOG_DIR" }}/gotak.log"
      max_size: 100
      max_backups: 5
      max_age: 30
      compress: true
    - type: "stdout"

# TAK Protocol Configuration
tak:
  protocol_version: "2.0"
  max_message_size: 1048576
  heartbeat_interval: "30s"
  client_timeout: "300s"
  max_concurrent_connections: 10000
  
  message_filtering:
    enabled: true
    max_position_rate: 10
    max_chat_rate: 5
  
  client_auth:
    enabled: false
    require_certificate: false
    certificate_validation: "optional"

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
    buffer_size: 1000
    workers: 4
    batch_size: 50
    flush_interval: "100ms"

# Performance Configuration
performance:
  connection_pooling:
    enabled: true
    max_connections: 1000
    max_idle_connections: 100
    connection_timeout: "30s"
  
  caching:
    enabled: true
    default_ttl: "3600s"
    cleanup_interval: "300s"
    max_memory: "1GB"
  
  background_jobs:
    enabled: true
    workers: 4
    queue_size: 1000
    retry_attempts: 3
    retry_delay: "30s"

# Monitoring and Health Checks
monitoring:
  metrics:
    enabled: {{ env "ENABLE_METRICS" }}
    endpoint: "/metrics"
    namespace: "gotak"
  
  health:
    enabled: true
    endpoint: "/health"
    checks:
      - name: "database"
        enabled: true
        timeout: "5s"
        interval: "30s"
      - name: "redis"
        enabled: true
        timeout: "3s"
        interval: "30s"

# Feature Flags
features:
  user_registration: true
  password_reset: true
  admin_api: true
  bulk_operations: true
  export_data: true
  import_data: true
  advanced_search: true
  audit_logging: true
  backup_restore: true

# Environment-specific settings
environment: "production"
debug: {{ env "ENABLE_DEBUG" }}
profiling: false
EOF
        destination = "local/server.yaml"
        change_mode = "restart"
      }

      service {
        name = "gotak-api-dockerhub"
        port = "api"
        tags = ["gotak", "api", "http", "tactical", "dockerhub"]

        check {
          type = "http"
          name = "api-health"
          path = "/health"
          interval = "30s"
          timeout = "5s"
          task = "server"
        }

        check {
          type = "http"
          name = "web-ui"
          path = "/"
          interval = "60s"
          timeout = "5s"
          task = "server"
        }
      }

      service {
        name = "gotak-cot-tcp-dockerhub"
        port = "cot_tcp"
        tags = ["gotak", "cot", "tcp", "tactical", "dockerhub"]

        check {
          type = "tcp"
          name = "cot-tcp"
          interval = "10s"
          timeout = "3s"
          task = "server"
        }
      }

      service {
        name = "gotak-cot-tls-dockerhub"
        port = "cot_tls"
        tags = ["gotak", "cot", "tls", "tactical", "secure", "dockerhub"]

        check {
          type = "tcp"
          name = "cot-tls"
          interval = "10s"
          timeout = "3s"
          task = "server"
        }
      }

      resources {
        cpu = 300
        memory = 512
      }
    }
  }

  group "web" {
    count = 1

    constraint {
      attribute = "${node.unique.name}"
      operator = "regexp"
      value = "hashinuc01"
    }

    network {
      port "web" {
        static = 8094
      }
    }

    migrate {
      max_parallel = 1
      health_check = "checks"
      min_healthy_time = "10s"
      healthy_deadline = "5m"
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
        GOTAK_API_URL = "http://hashinuc01:8093"
        GOTAK_WS_URL = "ws://hashinuc01:8093/ws"
        GOTAK_SERVER_URL = "ws://hashinuc01:8093"
        NGINX_PORT = "80"
      }


      service {
        name = "gotak-web-dockerhub"
        port = "web"
        tags = ["gotak", "web", "ui", "tactical", "dockerhub"]

        check {
          type = "http"
          name = "web-health"
          path = "/health"
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
}