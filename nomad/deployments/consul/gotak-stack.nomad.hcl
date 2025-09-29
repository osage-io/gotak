job "gotak-stack" {
  datacenters = ["dc1"]
  type        = "service"
  namespace   = "default"

  # Spread across all 3 nodes
  spread {
    attribute = "${node.unique.name}"
    weight    = 100
  }

  # Update strategy
  update {
    max_parallel      = 1
    health_check      = "checks"
    min_healthy_time  = "30s"
    healthy_deadline  = "5m"
    progress_deadline = "10m"
    auto_revert       = true
    auto_promote      = true
    canary            = 1
    stagger           = "30s"
  }

  # PostgreSQL Database Group
  group "database" {
    count = 1

    restart {
      attempts = 3
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    # Constraint to specific node for data persistence
    constraint {
      attribute = "${node.unique.name}"
      operator  = "regexp"
      value     = "hashinuc01"
    }

    network {
      port "postgres" {
        static = 5432
      }
    }

    volume "postgres-data" {
      type      = "host"
      read_only = false
      source    = "postgres-data"
    }

    task "postgres" {
      driver = "docker"

      config {
        image = "postgis/postgis:15-3.4"
        ports = ["postgres"]
        
        volumes = [
          "local/init.sql:/docker-entrypoint-initdb.d/init.sql",
        ]

        logging {
          type = "json-file"
          config {
            max-size = "10m"
            max-file = "3"
          }
        }
      }

      volume_mount {
        volume      = "postgres-data"
        destination = "/var/lib/postgresql/data"
        read_only   = false
      }

      env {
        POSTGRES_DB       = "gotak"
        POSTGRES_USER     = "gotak"
        POSTGRES_PASSWORD = "gotak"
        PGDATA           = "/var/lib/postgresql/data/pgdata"
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

CREATE INDEX idx_entities_location ON entities USING GIST(location);
CREATE INDEX idx_entities_callsign ON entities(callsign);
CREATE INDEX idx_entities_updated_at ON entities(updated_at);
EOF
        destination = "local/init.sql"
      }

      resources {
        cpu    = 1000
        memory = 2048
      }

      service {
        name = "gotak-postgres"
        port = "postgres"
        
        check {
          type     = "script"
          command  = "pg_isready"
          args     = ["-U", "gotak", "-d", "gotak"]
          interval = "10s"
          timeout  = "5s"
        }

        tags = [
          "database",
          "postgres",
          "gotak"
        ]
      }
    }
  }

  # Redis Cache Group
  group "cache" {
    count = 1

    restart {
      attempts = 3
      interval = "5m"
      delay    = "15s"
      mode     = "delay"
    }

    network {
      port "redis" {
        static = 6379
      }
    }

    volume "redis-data" {
      type      = "host"
      read_only = false
      source    = "redis-data"
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

      volume_mount {
        volume      = "redis-data"
        destination = "/data"
        read_only   = false
      }

      resources {
        cpu    = 500
        memory = 512
      }

      service {
        name = "gotak-redis"
        port = "redis"
        
        check {
          type     = "script"
          command  = "redis-cli"
          args     = ["ping"]
          interval = "10s"
          timeout  = "5s"
        }

        tags = [
          "cache",
          "redis",
          "gotak"
        ]
      }
    }
  }

  # NATS Messaging Group
  group "messaging" {
    count = 1

    restart {
      attempts = 3
      interval = "5m"
      delay    = "15s"
      mode     = "delay"
    }

    network {
      port "nats" {
        static = 4222
      }
      port "nats-cluster" {
        static = 6222
      }
      port "nats-http" {
        static = 8222
      }
    }

    task "nats" {
      driver = "docker"

      config {
        image = "nats:2.9-alpine"
        ports = ["nats", "nats-cluster", "nats-http"]
        
        args = [
          "-js",
          "-sd", "/data",
          "-m", "8222"
        ]

        volumes = [
          "local/nats.conf:/nats.conf"
        ]

        logging {
          type = "json-file"
          config {
            max-size = "10m"
            max-file = "3"
          }
        }
      }

      template {
        data = <<EOF
# NATS Server Configuration
port: 4222
http_port: 8222

# JetStream configuration
jetstream {
  store_dir: /data
  max_memory_store: 1GB
  max_file_store: 10GB
}

# Cluster configuration
cluster {
  port: 6222
  routes: []
}

# Logging
debug: false
trace: false
logtime: true
EOF
        destination = "local/nats.conf"
      }

      resources {
        cpu    = 500
        memory = 512
      }

      service {
        name = "gotak-nats"
        port = "nats"
        
        check {
          type     = "http"
          path     = "/healthz"
          port     = "nats-http"
          interval = "10s"
          timeout  = "5s"
        }

        tags = [
          "messaging",
          "nats",
          "gotak"
        ]
      }
    }
  }

  # GoTAK Server Group
  group "gotak-server" {
    count = 3  # Run 3 instances for HA

    restart {
      attempts = 3
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    spread {
      attribute = "${node.unique.name}"
      weight    = 100
    }

    network {
      port "tcp" {
        static = 8087
      }
      port "tls" {
        static = 8089
      }
      port "http" {
        static = 8082
      }
    }

    task "gotak-server" {
      driver = "docker"

      config {
        image = "gotak-server:latest"
        ports = ["tcp", "tls", "http"]
        
        volumes = [
          "local/server.yaml:/config/server.yaml",
          "secrets/certs:/certs"
        ]

        logging {
          type = "json-file"
          config {
            max-size = "10m"
            max-file = "5"
          }
        }
      }

      template {
        data = <<EOF
server:
  host: 0.0.0.0
  tcp_port: 8087
  tls_port: 8089
  http_port: 8082
  max_connections: 10000
  connection_timeout: 300s
  heartbeat_interval: 30s
  stale_threshold: 5m

database:
  host: {{ range service "gotak-postgres" }}{{ .Address }}{{ end }}
  port: {{ range service "gotak-postgres" }}{{ .Port }}{{ end }}
  database: gotak
  user: gotak
  password: gotak
  max_connections: 50
  max_idle_connections: 10

redis:
  host: {{ range service "gotak-redis" }}{{ .Address }}{{ end }}
  port: {{ range service "gotak-redis" }}{{ .Port }}{{ end }}
  db: 0

nats:
  url: nats://{{ range service "gotak-nats" }}{{ .Address }}:{{ .Port }}{{ end }}
  
security:
  tls_enabled: true
  cert_file: /certs/server.crt
  key_file: /certs/server.key
  ca_file: /certs/ca.crt
  client_auth: true

logging:
  level: info
  format: json
EOF
        destination = "local/server.yaml"
        change_mode = "restart"
      }

      # Vault integration for certificates
      template {
        data = <<EOF
{{ with secret "pki_int/issue/gotak-server" "common_name=gotak.service.consul" }}
{{ .Data.certificate }}
{{ end }}
EOF
        destination = "secrets/certs/server.crt"
        change_mode = "restart"
      }

      template {
        data = <<EOF
{{ with secret "pki_int/issue/gotak-server" "common_name=gotak.service.consul" }}
{{ .Data.private_key }}
{{ end }}
EOF
        destination = "secrets/certs/server.key"
        change_mode = "restart"
      }

      template {
        data = <<EOF
{{ with secret "pki_int/issue/gotak-server" "common_name=gotak.service.consul" }}
{{ .Data.issuing_ca }}
{{ end }}
EOF
        destination = "secrets/certs/ca.crt"
        change_mode = "restart"
      }

      resources {
        cpu    = 2000
        memory = 2048
      }

      service {
        name = "gotak-server"
        port = "http"
        
        check {
          type     = "http"
          path     = "/health"
          interval = "10s"
          timeout  = "5s"
        }

        tags = [
          "gotak",
          "server",
          "api"
        ]
      }

      service {
        name = "gotak-tcp"
        port = "tcp"
        
        tags = [
          "gotak",
          "tcp",
          "cot"
        ]
      }

      service {
        name = "gotak-tls"
        port = "tls"
        
        tags = [
          "gotak",
          "tls",
          "secure"
        ]
      }
    }
  }

  # GoTAK Web UI Group
  group "gotak-web" {
    count = 2  # Run 2 instances for HA

    restart {
      attempts = 3
      interval = "5m"
      delay    = "15s"
      mode     = "delay"
    }

    spread {
      attribute = "${node.unique.name}"
      weight    = 100
    }

    network {
      port "http" {
        to = 80
      }
    }

    task "gotak-web" {
      driver = "docker"

      config {
        image = "gotak-web:latest"
        ports = ["http"]
        
        volumes = [
          "local/config.json:/usr/share/nginx/html/config/config.json"
        ]

        logging {
          type = "json-file"
          config {
            max-size = "10m"
            max-file = "3"
          }
        }
      }

      template {
        data = <<EOF
{
  "apiUrl": "https://{{ env "NOMAD_GROUP_NAME" }}.service.consul:8082",
  "wsUrl": "wss://{{ env "NOMAD_GROUP_NAME" }}.service.consul:8089",
  "mapboxToken": "",
  "features": {
    "vault": {
      "enabled": {{ env "VAULT_ENABLED" | default "false" }},
      "url": "{{ env "VAULT_ADDR" | default "https://vault.service.consul:8200" }}"
    }
  }
}
EOF
        destination = "local/config.json"
      }

      resources {
        cpu    = 500
        memory = 256
      }

      service {
        name = "gotak-web"
        port = "http"
        
        check {
          type     = "http"
          path     = "/"
          interval = "10s"
          timeout  = "5s"
        }

        tags = [
          "gotak",
          "web",
          "ui",
          "traefik.enable=true",
          "traefik.http.routers.gotak.rule=Host(`gotak.demoland.io`)",
          "traefik.http.routers.gotak.tls=true",
          "traefik.http.routers.gotak.tls.certresolver=vault"
        ]
      }
    }
  }

  # Jaeger Tracing Group (Optional)
  group "tracing" {
    count = 1

    restart {
      attempts = 3
      interval = "5m"
      delay    = "15s"
      mode     = "delay"
    }

    network {
      port "jaeger-ui" {
        static = 16686
      }
      port "jaeger-collector" {
        static = 14268
      }
      port "jaeger-agent" {
        static = 14250
      }
    }

    task "jaeger" {
      driver = "docker"

      config {
        image = "jaegertracing/all-in-one:1.51"
        ports = ["jaeger-ui", "jaeger-collector", "jaeger-agent"]
        
        env {
          COLLECTOR_OTLP_ENABLED = "true"
          SPAN_STORAGE_TYPE      = "memory"
        }

        logging {
          type = "json-file"
          config {
            max-size = "10m"
            max-file = "3"
          }
        }
      }

      resources {
        cpu    = 500
        memory = 512
      }

      service {
        name = "gotak-jaeger"
        port = "jaeger-ui"
        
        check {
          type     = "http"
          path     = "/"
          interval = "10s"
          timeout  = "5s"
        }

        tags = [
          "tracing",
          "jaeger",
          "gotak",
          "traefik.enable=true",
          "traefik.http.routers.jaeger.rule=Host(`jaeger.gotak.demoland.io`)",
          "traefik.http.routers.jaeger.tls=true"
        ]
      }
    }
  }
}
