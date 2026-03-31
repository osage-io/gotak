# TLS API Gateway Nomad Job
# Deploys Envoy as a Consul API Gateway with TLS termination
# Uses wildcard certificate for *.demoland.io
# Routes: gotak.demoland.io, lurch.demoland.io, opencode.demoland.io

variable "datacenter" {
  default = "dc1"
}

variable "gateway_address" {
  description = "External IP address for the gateway"
  default     = "192.168.1.185"
}

variable "https_port" {
  description = "HTTPS listener port"
  default     = 8443
}

variable "cot_port" {
  description = "CoT TCP listener port"
  default     = 8089
}

job "demoland-api-gateway" {
  datacenters = [var.datacenter]
  type        = "service"
  priority    = 85

  meta {
    service = "demoland-api-gateway"
    version = "1.0.0"
  }

  group "gateway" {
    count = 1

    network {
      mode = "host"

      # HTTPS listener port (TLS termination)
      port "https" {
        static = var.https_port
      }

      # CoT TLS listener port
      port "cot-tls" {
        static = var.cot_port
      }

      # Envoy admin interface (use 19001 to avoid conflict with other gateways)
      port "admin" {
        static = 19001
      }
    }

    # Restart policy
    restart {
      attempts = 3
      interval = "10m"
      delay    = "15s"
      mode     = "fail"
    }

    # Update strategy
    update {
      max_parallel     = 1
      health_check     = "checks"
      min_healthy_time = "10s"
      healthy_deadline = "5m"
      auto_revert      = true
    }

    task "envoy" {
      driver = "raw_exec"

      # Note: TLS certificates are stored in Consul as an inline-certificate
      # config entry, so no local volume mount is needed. Envoy retrieves
      # the certificates directly from Consul.

      config {
        command = "/usr/local/bin/consul"
        args = [
          "connect", "envoy",
          "-gateway", "api",
          "-register",
          "-service", "demoland-gateway",
          "-address", "${var.gateway_address}:${var.https_port}",
          "-bind-address", "https-listener=0.0.0.0:${var.https_port}",
          "-bind-address", "cot-tls-listener=0.0.0.0:${var.cot_port}",
          "-grpc-addr", "127.0.0.1:8502",
          "-admin-bind", "0.0.0.0:19001",
          "--",
          "-l", "info"
        ]
      }

      env {
        CONSUL_HTTP_ADDR = "127.0.0.1:8500"
      }

      resources {
        cpu    = 300
        memory = 512
      }

      # Health check for Envoy admin interface
      service {
        name = "demoland-gateway-admin"
        port = "admin"

        check {
          type     = "http"
          path     = "/ready"
          interval = "10s"
          timeout  = "2s"
        }
      }
    }
  }
}
