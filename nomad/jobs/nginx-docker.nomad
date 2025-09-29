job "nginx-docker" {
  datacenters = ["dc1"]
  type = "service"

  group "nginx" {
    count = 1

    network {
      port "http" {
        static = 8080
      }
    }

    service {
      name = "nginx"
      port = "http"

      check {
        type = "http"
        path = "/"
        interval = "10s"
        timeout = "3s"
      }
    }

    task "nginx" {
      driver = "docker"

      config {
        image = "nginx:latest"
        ports = ["http"]
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
