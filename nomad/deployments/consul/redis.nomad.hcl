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

variable "redis_replicas" {
  description = "Number of Redis replicas"
  type        = number
  default     = 1
}

variable "redis_cpu" {
  description = "CPU allocation for Redis"
  type        = number
  default     = 100
}

variable "redis_memory" {
  description = "Memory allocation for Redis"
  type        = number
  default     = 128
}

variable "redis_image_tag" {
  description = "Redis image tag"
  type        = string
  default     = "7-alpine"
}

variable "volume_type" {
  description = "Volume type"
  type        = string
  default     = "host"
}

job "gotak-redis" {
  datacenters = [var.datacenter]
  region      = var.region
  namespace   = var.namespace
  type        = "service"
  priority    = 70

  meta {
    service = "redis"
    version = var.redis_image_tag
  }

  group "redis" {
    count = var.redis_replicas

    # Restart policy
    restart {
      attempts = 3
      interval = "10m"
      delay    = "15s"
      mode     = "fail"
    }

    # Volume for data persistence (optional)
    volume "redis-data" {
      type            = var.volume_type
      source          = "gotak-redis-data"
      attachment_mode = "file-system"
      access_mode     = "single-node-writer"
    }

    network {
      port "redis" {
        static = 6379
        to     = 6379
      }
    }

    task "redis" {
      driver = "docker"

      # Mount the persistent volume
      volume_mount {
        volume      = "redis-data"
        destination = "/data"
      }

      config {
        image = "redis:${var.redis_image_tag}"
        ports = ["redis"]
        
        # Redis configuration with persistence and password
        command = "redis-server"
        args = [
          "--appendonly", "yes",
          "--requirepass", "tactical_cache_pass",  # In production, use Vault template
          "--bind", "0.0.0.0",
          "--port", "6379",
          "--dir", "/data"
        ]

      }

      # Environment template (for Vault integration in production)
      # template {
      #   data        = <<-EOT
      #   REDIS_PASSWORD="{{ with secret "secret/gotak/redis" }}{{ .Data.data.password }}{{ end }}"
      #   EOT
      #   destination = "secrets/redis.env"
      #   env         = true
      # }

      resources {
        cpu    = var.redis_cpu
        memory = var.redis_memory
      }

      service {
        name = "redis"
        port = "redis"
        tags = [
          "cache",
          "pubsub",
          "session-store",
          "version-${var.redis_image_tag}"
        ]

        check {
          name     = "redis-tcp"
          type     = "tcp"
          interval = "10s"
          timeout  = "3s"
        }

        check {
          name     = "redis-ping"
          type     = "script"
          command  = "redis-cli"
          args     = ["-a", "tactical_cache_pass", "ping"]
          interval = "30s"
          timeout  = "5s"
        }
      }
    }
  }
}