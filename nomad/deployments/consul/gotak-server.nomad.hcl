# Variable declarations
variable "datacenter" {
  description = "The datacenter to run the job in"
  type        = string
  default     = "dc1"
}

variable "region" {
  description = "The region to run the job in"
  type        = string
  default     = "global"
}

variable "namespace" {
  description = "The namespace to run the job in"
  type        = string
  default     = "default"
}

variable "gotak_server_replicas" {
  description = "Number of GoTAK server replicas"
  type        = number
  default     = 1
}

variable "gotak_server_cpu" {
  description = "CPU allocation for GoTAK server"
  type        = number
  default     = 200
}

variable "gotak_server_memory" {
  description = "Memory allocation for GoTAK server"
  type        = number
  default     = 256
}

variable "image_tag" {
  description = "GoTAK server image tag"
  type        = string
  default     = "latest"
}

variable "registry_url" {
  description = "Docker registry URL"
  type        = string
  default     = "docker.io"
}

variable "registry_namespace" {
  description = "Docker registry namespace"
  type        = string
  default     = "gotak"
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "gotak_dev"
}

variable "db_user" {
  description = "Database user"
  type        = string
  default     = "gotak"
}

variable "log_level" {
  description = "Log level"
  type        = string
  default     = "debug"
}

variable "enable_tls" {
  description = "Enable TLS"
  type        = bool
  default     = false
}

variable "enable_debug" {
  description = "Enable debug mode"
  type        = bool
  default     = true
}

variable "enable_metrics" {
  description = "Enable metrics"
  type        = bool
  default     = true
}

variable "enable_auth" {
  description = "Enable authentication"
  type        = bool
  default     = false
}

job "gotak-server" {
  datacenters = [var.datacenter]
  region      = var.region
  namespace   = var.namespace
  type        = "service"
  priority    = 90

  meta {
    service = "gotak-server"
    version = var.image_tag
    rebuild = "2026-02-03-v4-frontend-api-fix"
  }

  group "server" {
    count = var.gotak_server_replicas

    # Update strategy
    update {
      max_parallel     = 1
      min_healthy_time = "30s"
      healthy_deadline = "5m"
      auto_revert      = true
    }

    # Restart policy
    restart {
      attempts = 3
      interval = "10m"
      delay    = "15s"
      mode     = "fail"
    }

    # Network configuration
    network {
      mode = "bridge"

      # CoT TCP port
      port "cot" {
        static = 8087
        to     = 8087
      }

      # CoT TLS port (reserved for future use when TLS is enabled)
      # port "cot_tls" {
      #   static = 8089
      #   to     = 8089
      # }

      # API/Web port
      port "api" {
        static = 8080
        to     = 8080
      }
    }

    # API Service with Connect sidecar
    service {
      name = "gotak-api"
      port = "8080"
      tags = [
        "api",
        "http",
        "tactical",
        "version-${var.image_tag}"
      ]

      check {
        name     = "api-health"
        type     = "http"
        path     = "/health"
        interval = "30s"
        timeout  = "5s"
        expose   = true
      }

      connect {
        sidecar_service {
          proxy {
            local_service_port = 8080
          }
        }
      }
    }

    # CoT TCP Service with Connect sidecar
    service {
      name = "gotak-cot"
      port = "8087"
      tags = [
        "tak",
        "cot-tcp",
        "tactical",
        "version-${var.image_tag}"
      ]

      connect {
        sidecar_service {
          proxy {
            local_service_port = 8087
          }
        }
      }
    }

    task "server" {
      driver = "docker"

      config {
        image              = "thefed/gotak-server:v5-202602031429"
        force_pull         = false
        ports              = ["cot", "api"]

        # Volume mounts for logs
        mount {
          type     = "tmpfs"
          target   = "/app/logs"
          readonly = false
        }

        # Volume mounts for data
        mount {
          type     = "tmpfs"
          target   = "/app/data"
          readonly = false
        }
      }

      env {
        GIN_MODE           = "release"
        GOTAK_CONFIG_PATH  = "/local/server.yaml"
        GOTAK_LOG_LEVEL    = var.log_level
        GOTAK_DATA_DIR     = "/app/data"
        GOTAK_LOG_DIR      = "/app/logs"
        GOTAK_STANDALONE   = "true"  # Skip migrations - they're already applied
        MIGRATIONS_DIR     = "/app/migrations"
        
        # Database - using container expected variables
        POSTGRES_HOST      = "192.168.1.185"
        POSTGRES_PORT      = "5432"
        POSTGRES_USER      = var.db_user
        POSTGRES_DB        = var.db_name
        POSTGRES_PASSWORD  = "tactical_secure_pass"  # Use Vault in production
        
        # Legacy DB vars for config template
        DB_USER            = var.db_user
        DB_NAME            = var.db_name
        DB_PASSWORD        = "tactical_secure_pass"
        
        # JWT Secret (required by container)
        JWT_SECRET         = "dev-jwt-secret-change-in-production"
        
        # Redis
        REDIS_HOST         = "192.168.1.185"
        REDIS_PORT         = "6379"
        REDIS_PASSWORD     = "tactical_cache_pass"   # Use Vault in production
        
        # Feature toggles
        ENABLE_TLS         = var.enable_tls
        ENABLE_DEBUG       = var.enable_debug
        ENABLE_METRICS     = var.enable_metrics
        ENABLE_AUTH        = var.enable_auth
        LOG_LEVEL          = var.log_level
      }

      # Configuration template
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
  serve_static: true

tak:
  tcp:
    enabled: true
    address: "0.0.0.0"
    port: {{ env "NOMAD_PORT_cot" }}
  udp:
    enabled: false
    address: "0.0.0.0"
    port: 8088
  tls:
    enabled: {{ env "ENABLE_TLS" }}
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
  host: "{{ range service "postgres" }}{{ .Address }}{{ end }}"
  port: {{ range service "postgres" }}{{ .Port }}{{ end }}
  username: "{{ env "DB_USER" }}"
  password: "{{ env "DB_PASSWORD" }}"
  database: "{{ env "DB_NAME" }}"
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300s

redis:
  host: "{{ range service "redis" }}{{ .Address }}{{ end }}"
  port: {{ range service "redis" }}{{ .Port }}{{ end }}
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

      # Environment template (for Vault secrets in production)
      # template {
      #   data        = <<-EOT
      #   DB_PASSWORD="{{ with secret "secret/gotak/postgres" }}{{ .Data.data.password }}{{ end }}"
      #   REDIS_PASSWORD="{{ with secret "secret/gotak/redis" }}{{ .Data.data.password }}{{ end }}"
      #   EOT
      #   destination = "secrets/app.env"
      #   env         = true
      # }

      resources {
        cpu    = var.gotak_server_cpu
        memory = var.gotak_server_memory
      }
    }
  }
}