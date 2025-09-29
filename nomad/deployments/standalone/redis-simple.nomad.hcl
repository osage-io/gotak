job "gotak-redis" {
  datacenters = ["dc1"]
  region      = "global"
  type        = "service"
  priority    = 70

  group "redis" {
    count = 1

    restart {
      attempts = 3
      interval = "10m"
      delay    = "15s"
      mode     = "fail"
    }

    network {
      port "redis" {
        static = 6379
        to     = 6379
      }
    }

    task "redis" {
      driver = "docker"

      config {
        image = "redis:7-alpine"
        ports = ["redis"]
        
        command = "redis-server"
        args = [
          "--appendonly", "yes",
          "--requirepass", "tactical_cache_pass",
          "--bind", "0.0.0.0",
          "--port", "6379",
          "--dir", "/data"
        ]

        volumes = [
          "/tmp/gotak-nomad-volumes/dev/redis-data:/data"
        ]
      }

      resources {
        cpu    = 100
        memory = 128
      }

      # No service registration in standalone mode
      # Services communicate via direct IP addressing
    }
  }
}