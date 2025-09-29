job "gotak-postgres" {
  datacenters = ["dc1"]
  region      = "global"
  type        = "service"
  priority    = 80

  group "postgres" {
    count = 1

    restart {
      attempts = 3
      interval = "10m"
      delay    = "30s"
      mode     = "fail"
    }

    network {
      port "db" {
        static = 5432
        to     = 5432
      }
    }

    task "postgres" {
      driver = "docker"

      config {
        image = "postgis/postgis:15-3.4-alpine"
        ports = ["db"]
        
        # Host volume mount
        volumes = [
          "/tmp/gotak-nomad-volumes/dev/postgres-data:/var/lib/postgresql/data"
        ]
      }

      # Environment variables
      env {
        POSTGRES_DB       = "gotak_dev"
        POSTGRES_USER     = "gotak"
        POSTGRES_PASSWORD = "tactical_secure_pass"
        PGDATA           = "/var/lib/postgresql/data/pgdata"
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