# GoTAK - Golang TAK Server

A high-performance, modern implementation of a TAK (Team Awareness Kit) server written in Go.

[![CI/CD Pipeline](https://github.com/dfedick/gotak/actions/workflows/ci.yml/badge.svg)](https://github.com/dfedick/gotak/actions/workflows/ci.yml)
[![Security Scan](https://github.com/dfedick/gotak/actions/workflows/security.yml/badge.svg)](https://github.com/dfedick/gotak/actions/workflows/security.yml)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Docker](https://img.shields.io/badge/docker-supported-blue)](Dockerfile)

## Overview

GoTAK is a compatible TAK server implementation that provides situational awareness and real-time coordination capabilities for military, first responders, and emergency management teams. It supports the Cursor on Target (CoT) protocol and is designed to be lightweight, fast, and highly scalable.

> 📋 **Active Development**: See [GitHub Issues](.github/ISSUES/) for current work and the [Deployment Architecture Epic](.github/ISSUES/deployment-architecture-epic.md) (75% complete, 1 week remaining).

## Features

- **CoT Protocol Support**: Full support for Cursor on Target XML messaging
- **Multiple Protocols**: TCP, UDP, and TLS connections
- **Real-time Communication**: Live position updates, chat messages, and system alerts
- **Scalable Architecture**: Designed to handle thousands of concurrent clients
- **Security First**: TLS encryption, client certificate authentication, and automated security scanning
- **Structured Logging**: Comprehensive logging with zerolog for production monitoring
- **Database Integration**: PostgreSQL support with embedded migrations
- **HashiCorp Stack Deployment**: Run on Nomad with Consul service mesh (Connect) and Vault-backed secrets
- **Vault Secrets Encryption**: Comms messages encrypted via Vault's transit engine
- **CI/CD Pipeline**: Automated testing, security scanning, and Docker image building
- **Developer Experience**: Hot reload, pre-commit hooks, integration testing, and a one-command local HashiStack
- **Cross-platform**: Runs on Linux, macOS, and Windows
- **Docker Support**: Multi-stage optimized builds for production deployment
- **Federation Support**: Connect multiple TAK servers (coming soon)
- **Web Interface**: Advanced tactical interface with intelligent search and keyboard shortcuts

## Quick Start

### Development Setup

```bash
# Clone the repository
git clone https://github.com/osage-io/gotak
cd gotak

# Start development environment (PostgreSQL, Redis, NATS, etc.)
make dev-up

# Build and run with hot reload
make dev
```

### Production Deployment

```bash
# Build and run with Docker
make docker-run

# Or build from source
make build
make run
```

### Using Pre-built Binaries

Download the latest release from [releases](https://github.com/dfedick/gotak/releases) and run:

```bash
./gotak-server -config config/server.yaml
```

## Project Structure

```
gotak/
├── cmd/                    # Application entry points
│   ├── gotak-server/       # Main server application
│   └── gotak-client/       # Test client application
├── internal/               # Private application code
│   ├── server/             # Server implementation
│   ├── client/             # Client connection handling
│   ├── auth/               # Authentication (planned)
│   └── database/           # Database layer (planned)
├── pkg/                    # Public library code
│   ├── config/             # Configuration management
│   ├── cot/                # CoT message handling
│   └── tak/                # TAK protocol utilities
├── config/                 # Configuration files
├── deployments/            # Deployment configurations
│   ├── docker/             # Docker configurations
│   └── k8s/                # Kubernetes manifests
├── hashistack-local/       # Local single-node Consul + Vault + Nomad dev runtime
├── nomad/                  # Nomad job specs (standalone + Consul Connect)
│   └── deployments/
│       ├── standalone/     # Self-contained jobs (no service mesh)
│       └── consul/         # Consul Connect (service mesh) jobs + API gateway
├── migrations/             # Embedded SQL migrations
├── docs/                   # Documentation
├── web/                    # React tactical web interface
├── test/                   # Test files
└── scripts/                # Build and utility scripts
```

## Configuration

The server uses YAML configuration files. See [config/server.yaml](config/server.yaml) for a complete example.

Key configuration sections:

- **Server**: Network settings, ports, and performance tuning
- **Security**: TLS certificates, client authentication
- **TAK**: Protocol-specific settings, message handling
- **Database**: Storage configuration (PostgreSQL support planned)
- **Logging**: Log levels, output formats, and rotation

## Usage

### Starting the Server

```bash
# Default configuration
./gotak-server

# Custom configuration file
./gotak-server -config /path/to/config.yaml

# Enable debug logging
./gotak-server -debug

# Show version information
./gotak-server -version
```

### Default Ports

- **8087**: TCP/UDP for TAK client connections
- **8089**: TLS for secure TAK client connections
- **8080**: Web interface (when enabled)

### Testing with the Client

```bash
# Build test client
go build -o bin/gotak-client ./cmd/gotak-client

# Connect to server
./bin/gotak-client -server localhost:8087 -callsign "TestUser"

# UDP mode
./bin/gotak-client -server localhost:8087 -protocol udp -callsign "TestUser"
```

### Client Commands

When connected with the test client:

- `pos <lat> <lon>` - Send position update
- `chat <message>` - Send chat message to all users
- `ping` - Send ping to server
- `quit` - Disconnect and exit

### Web Interface

GoTAK includes a modern tactical web interface with advanced search capabilities:

**Global Search & Navigation:**
- `Ctrl/Cmd + K` - Open command palette
- `/` - Quick search focus
- `Ctrl + 1-9` - Navigate to pages (Dashboard, Map, Comms, etc.)
- `Ctrl + E` - Emergency alert
- `Ctrl + I` - AI Intel Officer
- `Ctrl + L` - Alerts
- `Ctrl + G` - Settings

**Search Features:**
- Intelligent search across pages, entities, and actions
- Keyboard navigation (Arrow keys, Enter, Escape)
- Categorized results (Pages, Commands, AI Actions)
- Real-time filtering and highlighting

Access the web interface at `http://localhost:8080` when the server is running.

## Development

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (for build automation)
- Pre-commit (for code quality hooks)

### Development Infrastructure

The project includes a comprehensive development environment:

```bash
# Start all development services
make dev-up
```

**Available Services:**
- **PostgreSQL**: `localhost:5432` (user: gotak, db: gotak_dev)
- **PostgreSQL Test**: `localhost:5433` (user: gotak, db: gotak_test)
- **Redis**: `localhost:6379` (password: dev_redis_pass)
- **NATS**: `localhost:4222` (monitoring: localhost:8222)
- **Vault**: `localhost:8200` (token: dev-token)
- **Jaeger UI**: `localhost:16686` (distributed tracing)
- **Adminer**: `localhost:8081` (database admin interface)

### Quality Assurance

```bash
# Install development tools and pre-commit hooks
make install-tools
make precommit-install

# Run all tests
make test
make test-integration

# Code quality checks
make lint
make security
```

### Build Commands

```bash
# Show all available commands
make help

# Build the server
make build

# Run with development settings
make dev

# Run tests
make test

# Build for all platforms
make build-all

# Generate TLS certificates for testing
make certs
```

### HashiStack Commands

Spin up a local single-node Consul + Vault + Nomad runtime and deploy GoTAK to it:

```bash
# Install (via brew if needed) and start Consul, Vault, and Nomad in dev mode
make hashi-up

# Show health of the local HashiStack
make hashi-status

# Tail Consul/Vault/Nomad logs
make hashi-logs

# Deploy the GoTAK standalone stack to the local Nomad
make nomad-deploy

# Stop the GoTAK Nomad job
make nomad-stop

# Stop everything and clean ephemeral data
make hashi-down
```

See [hashistack-local/README.md](hashistack-local/README.md) for endpoints, tokens, and CLI environment variables.

### Adding Features

1. **CoT Message Types**: Extend `pkg/cot/cot.go` with new message types
2. **Server Handlers**: Add message processors in `internal/server/server.go`
3. **Client Features**: Enhance client handling in `internal/server/client.go`
4. **Configuration**: Update `pkg/config/config.go` for new settings

## Security

### TLS Configuration

For production deployments, enable TLS:

```yaml
security:
  tls_enabled: true
  cert_file: "certs/server.crt"
  key_file: "certs/server.key"
  client_auth_required: true  # Require client certificates
```

### Client Certificates

When `client_auth_required` is enabled, clients must present valid certificates signed by the configured CA.

### Best Practices

- Use TLS for all production deployments
- Implement proper firewall rules
- Regular certificate rotation
- Monitor for unauthorized access attempts
- Keep the server updated

## Deployment

### Docker Deployment

```bash
# Build Docker image
docker build -t gotak-server .

# Run container
docker run -p 8087:8087 -p 8089:8089 -p 8080:8080 gotak-server
```

### Kubernetes Deployment

Kubernetes manifests are available in `deployments/k8s/`:

```bash
kubectl apply -f deployments/k8s/
```

### HashiCorp Nomad Deployment

GoTAK ships Nomad job specs for deploying the full stack (server, PostgreSQL,
Redis) on a HashiCorp scheduler. Two flavors are provided under `nomad/deployments/`:

- **`standalone/`** — self-contained jobs that bind directly to host ports, ideal
  for a single-node Nomad or local development.
- **`consul/`** — jobs wired into the **Consul Connect** service mesh, with an API
  gateway for ingress and mutual-TLS between services.

To try it locally, bring up the bundled single-node HashiStack and deploy:

```bash
make hashi-up        # start Consul, Vault, and Nomad (dev mode)
make nomad-deploy    # render + submit the GoTAK standalone job to local Nomad
make hashi-status    # check health
```

The local Nomad UI is at `http://127.0.0.1:4646`, Consul at
`http://127.0.0.1:8500`, and Vault at `http://127.0.0.1:8200`. See
[hashistack-local/README.md](hashistack-local/README.md) for full details.

### Secrets with Vault

GoTAK uses Vault's **transit** secrets engine to encrypt communications messages.
For the local stack, configure the transit key (re-run after any dev-Vault
restart, since dev mode is in-memory):

```bash
./hashistack-local/vault-setup.sh
```

### Secure Access with Boundary (optional)

For brokered access to the infrastructure endpoints (Consul, Nomad, Vault) via
HashiCorp Boundary, a demo setup script is provided:

```bash
./hashistack-local/boundary-setup.sh
```

### Systemd Service

Example systemd service file:

```ini
[Unit]
Description=GoTAK Server
After=network.target

[Service]
Type=simple
User=gotak
ExecStart=/usr/local/bin/gotak-server -config /etc/gotak/server.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## Protocol Details

### Cursor on Target (CoT)

GoTAK implements the standard CoT protocol with XML message format:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<event version="2.0" uid="unique-id" type="a-f-G" time="2024-01-01T12:00:00.000Z"
       start="2024-01-01T12:00:00.000Z" stale="2024-01-01T12:05:00.000Z" how="h-g">
    <point lat="37.7749" lon="-122.4194" hae="0" ce="1" le="1"/>
    <detail>
        <contact callsign="TestUser" endpoint="*:-1:tcp"/>
        <__group name="Blue" role="Team Member"/>
    </detail>
</event>
```

### Supported Message Types

- Position reports (`a-f-*`, `a-h-*`)
- Chat messages (`b-t-f`)
- System messages (`t-x-*`)
- Emergency alerts (`b-t-f-e`)
- Custom message types

## Performance

### Benchmarks

- **Concurrent Connections**: 10,000+ clients
- **Message Throughput**: 50,000+ messages/second
- **Memory Usage**: ~50MB base, ~1KB per client
- **CPU Usage**: Low latency, optimized for high throughput

### Tuning

Adjust these settings for your deployment:

```yaml
server:
  max_connections: 10000
  read_timeout: 30s
  write_timeout: 30s
  keepalive_interval: 30s

tak:
  max_message_size: 8192
  heartbeat_interval: 60s
```

## Troubleshooting

### Common Issues

1. **Connection Refused**: Check if server is running and ports are open
2. **TLS Errors**: Verify certificate files and permissions
3. **High Memory Usage**: Check for message loops or stale connections
4. **Message Parsing Errors**: Validate CoT XML format

### Debug Logging

Enable debug logging for troubleshooting:

```bash
./gotak-server -debug
```

Or set in configuration:

```yaml
logging:
  level: "debug"
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go conventions and best practices
- Add tests for new functionality
- Update documentation
- Use meaningful commit messages
- Ensure code passes linting (`make lint`)

## Roadmap

- [x] Web administration interface (React tactical UI)
- [x] Database persistence layer (PostgreSQL + embedded migrations)
- [x] User authentication and authorization
- [x] Nomad deployment with Consul Connect service mesh
- [x] Vault-backed secrets and transit encryption
- [ ] Federation with other TAK servers
- [ ] Plugin system for custom message types
- [ ] REST API for external integrations
- [ ] Metrics and monitoring endpoints
- [ ] Message filtering and routing rules

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- TAK Protocol specification
- Go community for excellent libraries
- Contributors and testers

## Support

For questions, issues, or contributions:

- Create an [Issue](https://github.com/dfedick/gotak/issues)
- Start a [Discussion](https://github.com/dfedick/gotak/discussions)
- Submit a [Pull Request](https://github.com/dfedick/gotak/pulls)

---

**Note**: This is an independent implementation and is not affiliated with or endorsed by any official TAK development teams.
