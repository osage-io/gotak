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

variable "postgres_replicas" {
  description = "Number of PostgreSQL replicas"
  type        = number
  default     = 1
}

variable "postgres_cpu" {
  description = "CPU allocation for PostgreSQL"
  type        = number
  default     = 200
}

variable "postgres_memory" {
  description = "Memory allocation for PostgreSQL"
  type        = number
  default     = 256
}

variable "postgres_image_tag" {
  description = "PostgreSQL image tag"
  type        = string
  default     = "15-3.4-alpine"
}

variable "volume_type" {
  description = "Volume type"
  type        = string
  default     = "host"
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

job "gotak-postgres" {
  datacenters = [var.datacenter]
  region      = var.region
  namespace   = var.namespace
  type        = "service"
  priority    = 80

  meta {
    service = "postgres"
    version = var.postgres_image_tag
  }

  group "postgres" {
    count = var.postgres_replicas

    # Restart policy
    restart {
      attempts = 3
      interval = "10m"
      delay    = "30s"
      mode     = "fail"
    }

    # Volume for data persistence
    volume "postgres-data" {
      type            = var.volume_type
      source          = "gotak-postgres-data"
      attachment_mode = "file-system"
      access_mode     = "single-node-writer"
    }

    network {
      port "db" {
        static = 5432
        to     = 5432
      }
    }

    task "postgres" {
      driver = "docker"

      # Mount the persistent volume
      volume_mount {
        volume      = "postgres-data"
        destination = "/var/lib/postgresql/data"
      }

      config {
        image = "postgis/postgis:${var.postgres_image_tag}"
        ports = ["db"]
      }

      env {
        POSTGRES_DB       = var.db_name
        POSTGRES_USER     = var.db_user
        POSTGRES_PASSWORD = "tactical_secure_pass"  # In production, use Vault template
        PGDATA            = "/var/lib/postgresql/data/pgdata"
      }

      # Environment template (for Vault integration in production)
      # template {
      #   data        = <<-EOT
      #   POSTGRES_PASSWORD="{{ with secret "secret/gotak/postgres" }}{{ .Data.data.password }}{{ end }}"
      #   EOT
      #   destination = "secrets/postgres.env"
      #   env         = true
      # }

      resources {
        cpu    = var.postgres_cpu
        memory = var.postgres_memory
      }

      service {
        name = "postgres"
        port = "db"
        tags = [
          "database",
          "postgis", 
          "primary",
          "version-${var.postgres_image_tag}"
        ]

        check {
          name     = "postgres-tcp"
          type     = "tcp"
          interval = "10s"
          timeout  = "3s"
        }

        check {
          name     = "postgres-ready"
          type     = "script"
          command  = "/usr/local/bin/pg_isready"
          args     = ["-h", "localhost", "-p", "5432", "-U", var.db_user, "-d", var.db_name]
          interval = "30s"
          timeout  = "5s"
        }
      }
    }
  }
}