# GoTAK Terminating Gateway Nomad Job
# Bridges Consul mesh traffic to non-mesh services (gotak-api, gotak-cot)

job "gotak-terminating-gateway" {
  datacenters = ["dc1"]
  type        = "service"
  priority    = 75

  group "gateway" {
    count = 1

    network {
      port "mesh" {
        static = 8444
      }
    }

    restart {
      attempts = 3
      interval = "10m"
      delay    = "15s"
      mode     = "fail"
    }

    task "envoy" {
      driver = "raw_exec"

      config {
        command = "/usr/local/bin/consul"
        args = [
          "connect", "envoy",
          "-gateway", "terminating",
          "-register",
          "-service", "gotak-terminating-gateway",
          "-address", "192.168.1.185:8444",
          "-grpc-addr", "127.0.0.1:8502",
          "-admin-bind", "127.0.0.1:19002",
          "--",
          "-l", "info"
        ]
      }

      env {
        CONSUL_HTTP_ADDR = "127.0.0.1:8500"
      }

      resources {
        cpu    = 200
        memory = 256
      }
    }
  }
}
