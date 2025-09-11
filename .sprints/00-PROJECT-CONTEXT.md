# GoTAK Project Context & Requirements

**Last Updated:** 2025-09-08  
**Status:** Sprint Planning Complete  
**Next Session:** Ready for Sprint 1 execution

## Project Vision

GoTAK is a modern, zero-trust tactical situational awareness platform designed for DoD deployment. It combines the standard TAK/CoT protocol with enterprise-grade security, compliance, and deployment flexibility.

## Key Architecture Decisions

### Deployment Strategy: Embedded-First, Scale-Out
- **Demo Mode**: Single container with embedded SQLite, in-memory cache, static files
- **Production Mode**: Distributed microservices with PostgreSQL/TimescaleDB, Redis, Nomad/Consul/Vault
- **Hybrid Mode**: Mix embedded and external services based on configuration

### Technology Stack

**Backend:**
- Go 1.21+ server with CoT protocol support
- Embedded SQLite → PostgreSQL/TimescaleDB
- In-memory cache → Redis
- Configuration-driven service detection

**Frontend:**
- React + TypeScript + Vite
- Leaflet/Mapbox GL JS for mapping
- WebSocket real-time updates
- Mobile-responsive design (web-first MVP)

**Infrastructure:**
- Nomad orchestration with Consul service discovery
- Vault for secrets management
- Consul Connect service mesh
- Docker containerization

## Security & Compliance

### Authentication (Zero-Trust)
- **Primary**: Vault OIDC integration
- **Secondary**: TLS client certificate authentication
- **Fallback**: Built-in user/password when Vault unavailable

### Authorization (RBAC)
- **System Admin**: Full server management
- **Mission Commander**: Mission creation, personnel assignment
- **Operator**: Position updates, messaging, mission participation
- **Observer**: Read-only operational picture

### DoD Compliance Requirements
- Comprehensive audit logging
- Data classification labels
- Secure data transmission (TLS 1.3)
- Access control enforcement
- Incident response capabilities

## Core Features

### Tactical Capabilities
- Real-time position tracking with history
- Secure chat and messaging
- Mission planning and management
- Group-based permissions and visibility
- Emergency alerts and notifications

### Mapping Features
- Basic position tracking and markers
- Advanced tactical overlays (zones, routes, boundaries)
- Multiple map tile service integration
- Offline mapping capabilities
- Real-time updates via WebSocket

### Integration Points
- TAK server federation capabilities
- External system integration framework
- Mock integrations for testing (weather, intel feeds, etc.)
- REST API for external applications

## Deployment Targets

### Demo Environment
```bash
# Single command demo
docker run -p 8080:8080 gotak/server:latest
```

### Development Environment
```bash
# Local development with hot reload
make dev
```

### Production Environment
```hcl
# Nomad cluster with Consul/Vault integration
job "gotak" {
  # Full distributed deployment
}
```

## Success Criteria

### Sprint 1-3: Foundation
- Embedded SQLite database layer
- Authentication system (OIDC + fallback)
- Basic REST API with audit logging
- Configuration system for embedded vs external services

### Sprint 4-6: Frontend & Maps
- React frontend with real-time WebSocket connection
- Interactive maps with position tracking
- Basic mission management interface
- Mobile-responsive design

### Sprint 7-9: Advanced Features
- Federation with other TAK servers
- Advanced mapping overlays and tactical features
- Comprehensive role-based access control
- External integration framework

### Sprint 10-12: Production Deployment
- Nomad job definitions with Consul/Vault integration
- PostgreSQL/TimescaleDB migration
- Load testing and performance optimization
- DoD compliance documentation and auditing

## Technical Implementation Notes

### Database Strategy
- **Embedded**: SQLite with time-series simulation using JSON columns
- **Production**: PostgreSQL with TimescaleDB extension
- **Migration**: Built-in migration system supporting both backends

### Cache Strategy  
- **Embedded**: In-memory Go maps with TTL
- **Production**: Redis with consistent hashing
- **Interface**: Common cache interface for seamless switching

### Frontend Asset Strategy
- **Embedded**: Static files compiled into Go binary using embed
- **Production**: Served by CDN or dedicated static file service
- **Development**: Vite dev server with proxy

### Configuration Hierarchy
1. Environment variables (highest priority)
2. YAML configuration file
3. Consul KV (if available)
4. Vault secrets (if available)
5. Embedded defaults (lowest priority)

## Current Status

**Completed:**
- Basic TAK server with CoT protocol support
- TCP/UDP/TLS listeners and client management
- Configuration system and project structure
- Docker containerization

**Next Steps:**
- Begin Sprint 1: Database abstraction layer
- Implement embedded SQLite with external PostgreSQL fallback
- Add authentication system with Vault OIDC integration

---

## Session Handoff Context

**For next development session:**
1. We have a working basic TAK server foundation
2. Sprint plan is defined and ready for execution
3. Start with Sprint 1: Database & Auth Foundation
4. All clarifying questions have been answered
5. Architecture decisions are documented and agreed upon

**To resume development:**
```bash
cd /Users/dfedick/projects/gotak
ls .sprints/
# Review sprint files and begin Sprint 1 execution
```
