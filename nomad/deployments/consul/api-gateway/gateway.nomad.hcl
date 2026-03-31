# API Gateway Nomad Job
# Deploys Envoy as a Consul API Gateway using consul connect envoy
# Listeners: gotak HTTP (8443), gotak CoT TCP (9087), lurch HTTP (5724)

job "gotak-api-gateway" {
  datacenters = ["dc1"]
  type        = "service"
  priority    = 80

  group "gateway" {
    count = 1

    network {
      # GoTAK HTTP listener port
      port "http" {
        static = 8443
      }

      # GoTAK CoT TCP listener port
      port "cot" {
        static = 9087
      }

      # Lurch HTTP listener port
      port "lurch" {
        static = 5724
      }
    }

    # Restart policy
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
          "-gateway", "api",
          "-register",
          "-service", "gotak-gateway",
          "-address", "192.168.1.185:8443",
          "-bind-address", "http-listener=0.0.0.0:8443",
          "-bind-address", "cot-listener=0.0.0.0:9087",
          "-bind-address", "lurch-listener=0.0.0.0:5724",
          "-grpc-addr", "127.0.0.1:8502",
          "-admin-bind", "127.0.0.1:19000",
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
