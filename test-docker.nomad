job "test-docker" {
  datacenters = ["dc1"]
  type        = "batch"

  group "test" {
    count = 1

    task "hello" {
      driver = "docker"

      config {
        image = "alpine:latest"
        command = "echo"
        args = ["Hello from Docker on Nomad!"]
      }

      resources {
        cpu    = 100
        memory = 64
      }
    }
  }
}