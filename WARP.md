# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

GoTAK is a high-performance, modern implementation of a TAK (Team Awareness Kit) server written in Go. It provides situational awareness and real-time coordination capabilities for military, first responders, and emergency management teams using the Cursor on Target (CoT) protocol.

## Development Commands

### Building
```bash
# Build server and client
make build

# Build for all platforms
make build-all

# Build and run with debug
make dev
```

### Testing
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Build test client
go build -o bin/gotak-client ./cmd/gotak-client
```

### Running
```bash
# Start server with default config
make run

# Start server with custom config
./bin/gotak-server -config config/server.yaml

# Start server in debug mode
./bin/gotak-server -debug

# Test client connection
./bin/gotak-client -server localhost:8087 -callsign "TestUser"
```

### Docker
```bash
# Build and run with Docker
make docker-build
make docker-run

# Stop Docker container
make docker-stop
```

### Development Tools
```bash
# Install development tools
make install-tools

# Format code
make fmt

# Run linter
make lint

# Generate TLS certificates for testing
make certs
```

## Architecture Overview

### Core Components

1. **CoT Protocol Engine** (`pkg/cot/`): Handles Cursor on Target XML message parsing, validation, and generation
2. **Server Core** (`internal/server/`): Main TAK server with TCP/UDP/TLS listeners and message broadcasting
3. **Client Management** (`internal/server/client.go`): Connection handling, message routing, and client state
4. **Configuration** (`pkg/config/`): YAML-based configuration with validation and defaults
5. **Network Protocols**: Support for TCP, UDP, and TLS connections with proper message framing

### Message Flow

1. Clients connect via TCP, UDP, or TLS on ports 8087/8089
2. CoT XML messages are parsed and validated
3. Messages are processed based on type (position, chat, system)
4. Valid messages are broadcast to appropriate clients
5. Connection health is maintained with heartbeats

### Key Data Structures

- `cot.Event`: Represents a complete CoT message with XML marshaling
- `server.Client`: Manages individual client connections and state
- `server.Server`: Main server instance with connection management
- `config.ServerConfig`: Complete server configuration structure

## TAK Protocol Implementation

### Supported CoT Message Types
- Position reports (`a-f-*` friendly, `a-h-*` hostile)
- Chat messages (`b-t-f`)
- System messages (`t-x-*` heartbeat, ping)
- Emergency alerts (`b-t-f-e`)

### Message Processing Pipeline
1. XML parsing and validation in `cot.ParseCoT()`
2. Message type detection using helper functions
3. Client state updates (callsign, group, last seen)
4. Message routing and broadcasting
5. Protocol-specific handling (TCP stream vs UDP datagram)

## Configuration Structure

The server uses YAML configuration with these main sections:

- `server`: Network settings, ports, timeouts
- `security`: TLS certificates, client authentication
- `tak`: Protocol settings, message sizes, heartbeat intervals
- `database`: Storage configuration (planned)
- `logging`: Log levels and output formatting

Default ports:
- 8087: TCP/UDP for TAK clients
- 8089: TLS for secure clients  
- 8080: Web interface (planned)

## Development Guidelines

### Adding New CoT Message Types
1. Define constants in `pkg/cot/cot.go`
2. Add XML struct fields if needed
3. Create helper functions (`IsTypeXxx`)
4. Add processing logic in `internal/server/server.go`

### Extending Server Functionality
1. Add configuration options in `pkg/config/config.go`
2. Implement handlers in `internal/server/server.go`
3. Update client handling in `internal/server/client.go`
4. Add tests in appropriate `_test.go` files

### Security Considerations
- All production deployments should use TLS
- Client certificate authentication is recommended
- Message size limits prevent DoS attacks
- Connection timeouts prevent resource exhaustion
- Audit logging captures all authentication events

### Performance Notes
- Server is designed for 10,000+ concurrent clients
- Message broadcasting uses efficient channel operations
- Connection pooling minimizes resource usage
- Stale client cleanup prevents memory leaks

## Testing Strategy

### Unit Tests
- CoT message parsing and generation
- Configuration validation
- Message type detection functions

### Integration Tests  
- Client connection workflows
- Message broadcasting scenarios
- Protocol compatibility testing

### Manual Testing
Use the included test client for functional verification:
```bash
./bin/gotak-client -server localhost:8087 -callsign "TestUser"
# Commands: pos <lat> <lon>, chat <message>, ping, quit
```

## Common Development Tasks

### Adding a New Message Handler
1. Identify CoT type in `pkg/cot/cot.go`
2. Add `handleXxxMessage()` function in `internal/server/server.go`
3. Update `processMessage()` switch statement
4. Test with client and verify message flow

### Implementing New Configuration Options
1. Add fields to appropriate struct in `pkg/config/config.go`
2. Set defaults in `setDefaults()` function
3. Add validation in `validateConfig()`
4. Use configuration in relevant server components

### Debugging Connection Issues
1. Enable debug logging with `-debug` flag
2. Check server startup messages for port binding
3. Verify client connection attempts in logs
4. Monitor message parsing errors
5. Use test client for protocol verification

## External Dependencies

- `gopkg.in/yaml.v3`: YAML configuration parsing
- Go standard library for networking, XML, crypto

The project minimizes external dependencies to maintain security and reduce attack surface.

## Future Development Areas

1. **Database Layer**: PostgreSQL integration for persistence
2. **Web Interface**: Administration and monitoring UI
3. **Federation**: Multi-server connectivity 
4. **Authentication**: User management and RBAC
5. **REST API**: External integration endpoints
6. **Metrics**: Prometheus monitoring integration

This architecture provides a solid foundation for a production-ready TAK server while maintaining simplicity and performance.
