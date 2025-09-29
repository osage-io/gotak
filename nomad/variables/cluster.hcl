# GoTAK Cluster Variables
# Contains static cluster-level configuration

datacenter = "dc1"
region = "global"

# Consul configuration
consul_address = "localhost:8500"
consul_datacenter = "dc1"

# Registry configuration
registry_url = "docker.io"
registry_namespace = "gotak"

# Network configuration
network_cidr = "172.20.0.0/16"

# TLS and security
tls_skip_verify = true

# Volume driver (adjust based on your CSI setup)
volume_driver = "host"
volume_type = "host"