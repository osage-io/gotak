# Nomad configuration for local development with Docker driver

datacenter = "dc1"
data_dir = "/Users/dfedick/projects/gotak/nomad/data"
log_level = "INFO"
log_json = false

bind_addr = "127.0.0.1"

# Server configuration
server {
  enabled = true
  bootstrap_expect = 1
}

# Client configuration  
client {
  enabled = true
  
  # Enable Docker driver
  options {
    "driver.allowlist" = "docker,exec,raw_exec"
  }
  
  # Host volumes for persistent storage
  host_volume "postgres-data" {
    path      = "/tmp/gotak-nomad-volumes/dev/postgres-data"
    read_only = false
  }
}

# Docker driver configuration
plugin "docker" {
  config {
    allow_privileged = true
    allow_caps = ["ALL"]
    volumes {
      enabled = true
    }
    
    # Docker daemon settings
    endpoint = "unix:///var/run/docker.sock"
    
    # Enable Docker image pulling
    auth {
      config = "/Users/dfedick/.docker/config.json"
    }
    
    # Garbage collection
    gc {
      image = true
      image_delay = "3m"
      container = true
      dangling_containers {
        enabled = true
        dry_run = false
        period = "5m"
        creation_grace = "5m"
      }
    }
  }
}

# Consul integration (optional - uncomment if you have Consul running)
# consul {
#   address = "127.0.0.1:8500"
# }

# Web UI is enabled by default in server mode

# Ports configuration
ports {
  http = 4646
  rpc  = 4647
  serf = 4648
}
