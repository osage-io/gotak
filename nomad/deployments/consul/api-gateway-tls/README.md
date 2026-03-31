# Demoland API Gateway with TLS

This directory contains the Consul API Gateway configuration for routing HTTPS traffic to services using the `*.demoland.io` wildcard certificate.

## Overview

The API gateway provides:
- **TLS termination** on port 8443 using the wildcard certificate
- **Host-based routing** to multiple backend services
- **CoT TLS** on port 8089 for encrypted Cursor-on-Target protocol

### Routed Services

| URL | Backend Service |
|-----|-----------------|
| https://gotak.demoland.io:8443 | `gotak-api` |
| https://lurch.demoland.io:8443 | `lurch` |
| https://opencode.demoland.io:8443 | `opencode` |
| Port 8089 (TCP/TLS) | `gotak-cot` |

## Prerequisites

1. **Consul** and **Nomad** must be running and accessible
2. **Backend services** must be registered in Consul:
   - `gotak-api`
   - `lurch`
   - `opencode`
   - `gotak-cot` (for CoT traffic)
3. **Wildcard certificate** files in `~/sw/demoland-wildcard-certs/`:
   - `wildcard_demoland_io_fullchain.crt`
   - `wildcard_demoland_io.key`

## Quick Start

```bash
# Deploy everything (configs + Nomad job)
./setup.sh deploy

# Apply Consul configs only (no Nomad job)
./setup.sh config-only

# Tear down everything
./setup.sh cleanup
```

## Files

| File | Description |
|------|-------------|
| `setup.sh` | Main deployment script |
| `gateway.nomad.hcl` | Nomad job for Envoy API gateway |
| `api-gateway.hcl` | Consul API gateway config entry with TLS |
| `inline-certificate.hcl.template` | Template for certificate (auto-generated) |
| `http-route-gotak.hcl` | Route for gotak.demoland.io |
| `http-route-lurch.hcl` | Route for lurch.demoland.io |
| `http-route-opencode.hcl` | Route for opencode.demoland.io |
| `tcp-route-cot.hcl` | TCP route for CoT TLS traffic |
| `intentions-*.hcl` | Service intentions for gateway access |

## Configuration

### Changing the Gateway IP

Edit `gateway.nomad.hcl` and update the `gateway_address` variable:

```hcl
variable "gateway_address" {
  default = "192.168.1.185"  # Change to your host IP
}
```

### Changing Ports

Edit `gateway.nomad.hcl`:

```hcl
variable "https_port" {
  default = 8443  # HTTPS listener port
}

variable "cot_port" {
  default = 8089  # CoT TLS listener port
}
```

Also update the corresponding ports in `api-gateway.hcl`.

### Adding a New Service

1. Create a new HTTP route file (e.g., `http-route-myservice.hcl`):

```hcl
Kind = "http-route"
Name = "myservice-demoland-route"

Parents = [
  {
    Kind        = "api-gateway"
    Name        = "demoland-gateway"
    SectionName = "https-listener"
  }
]

Hostnames = ["myservice.demoland.io"]

Rules = [
  {
    Matches = [
      {
        Path = {
          Match = "prefix"
          Value = "/"
        }
      }
    ]
    Services = [
      {
        Name   = "myservice"
        Weight = 100
      }
    ]
  }
]
```

2. Create an intention file (e.g., `intentions-myservice.hcl`):

```hcl
Kind = "service-intentions"
Name = "myservice"
Sources = [
  {
    Name   = "demoland-gateway"
    Action = "allow"
  }
]
```

3. Apply the new configs:

```bash
consul config write http-route-myservice.hcl
consul config write intentions-myservice.hcl
```

## Useful Commands

```bash
# Check Nomad job status
nomad job status demoland-api-gateway

# View Consul services
consul catalog services

# Read gateway config
consul config read -kind api-gateway -name demoland-gateway

# List all routes
consul config list -kind http-route

# Test locally (bypassing DNS)
curl -k https://localhost:8443 -H 'Host: gotak.demoland.io'

# Check Envoy admin interface
curl http://localhost:19001/clusters
curl http://localhost:19001/config_dump
```

## Troubleshooting

### Gateway not starting

1. Check Nomad job status:
   ```bash
   nomad job status demoland-api-gateway
   nomad alloc logs <alloc-id>
   ```

2. Verify Consul is accessible:
   ```bash
   consul members
   consul catalog services
   ```

### Routes not working

1. Verify the inline certificate was created:
   ```bash
   consul config read -kind inline-certificate -name demoland-wildcard-cert
   ```

2. Check route bindings:
   ```bash
   consul config read -kind http-route -name gotak-demoland-route
   ```

3. Verify intentions are in place:
   ```bash
   consul config read -kind service-intentions -name gotak-api
   ```

### Certificate issues

1. Verify certificate files exist:
   ```bash
   ls -la ~/sw/demoland-wildcard-certs/
   ```

2. Check certificate validity:
   ```bash
   openssl x509 -in ~/sw/demoland-wildcard-certs/wildcard_demoland_io_fullchain.crt -text -noout
   ```

3. Regenerate the inline certificate:
   ```bash
   ./setup.sh config-only
   ```

## Architecture

```
                    Internet
                        │
                        ▼
┌───────────────────────────────────────────────────┐
│              demoland-gateway (Envoy)             │
│  ┌─────────────────────┐  ┌────────────────────┐  │
│  │   https-listener    │  │  cot-tls-listener  │  │
│  │       :8443         │  │       :8089        │  │
│  │   (TLS termination) │  │  (TLS termination) │  │
│  └──────────┬──────────┘  └─────────┬──────────┘  │
└─────────────┼───────────────────────┼─────────────┘
              │                       │
    ┌─────────┼─────────┐             │
    │         │         │             │
    ▼         ▼         ▼             ▼
┌───────┐ ┌───────┐ ┌────────┐  ┌───────────┐
│gotak- │ │ lurch │ │opencode│  │ gotak-cot │
│  api  │ │       │ │        │  │           │
└───────┘ └───────┘ └────────┘  └───────────┘
```

## DNS Configuration

Ensure your DNS records point to the gateway IP:

```
gotak.demoland.io     A    192.168.1.185
lurch.demoland.io     A    192.168.1.185
opencode.demoland.io  A    192.168.1.185
```

Or use a wildcard record:

```
*.demoland.io         A    192.168.1.185
```
