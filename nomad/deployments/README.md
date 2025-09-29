# GoTAK Nomad Deployments

This directory contains two different deployment strategies for GoTAK on Nomad:

## 📁 Directory Structure

```
nomad/deployments/
├── README.md                        # This file
├── standalone/                      # Simple deployment without Consul
│   ├── postgres-simple.nomad.hcl  # PostgreSQL without service discovery
│   ├── redis-simple.nomad.hcl     # Redis without service discovery  
│   └── gotak-server-simple.nomad.hcl # GoTAK server with hardcoded IPs
└── consul/                          # Full Consul integration
    ├── postgres.nomad.hcl          # PostgreSQL with Consul service discovery
    ├── redis.nomad.hcl             # Redis with Consul service discovery
    └── gotak-server.nomad.hcl      # GoTAK server with service discovery
```

## 🚀 Quick Start

### Standalone Deployment (Recommended for Development)

Perfect for development, testing, and simple deployments:

```bash
# Deploy standalone version
export NOMAD_ADDR=http://localhost:4646
./nomad/scripts/deploy-universal.sh standalone dev

# Access GoTAK at http://localhost:8080
```

### Consul-Integrated Deployment (Production Ready)

Full service mesh with automatic service discovery:

```bash
# Deploy with Consul integration
export NOMAD_ADDR=http://localhost:4646
./nomad/scripts/deploy-universal.sh consul dev

# Access GoTAK at http://localhost:8080
# Access Consul UI at http://localhost:8500/ui
```

## 🔄 Deployment Types Comparison

| Feature | Standalone | Consul |
|---------|-----------|--------|
| **Service Discovery** | Hardcoded IPs | Automatic via Consul DNS |
| **Health Checking** | Basic Nomad checks | Advanced Consul health checks |
| **Load Balancing** | None | Consul-native load balancing |
| **Service Mesh** | No | Optional Consul Connect |
| **Configuration Complexity** | Simple | Advanced |
| **Production Ready** | Development/Simple prod | Full production |
| **Prerequisites** | Just Nomad + Docker | Nomad + Docker + Consul |

## 🏗️ Standalone Deployment

### Features
- **Direct IP Addressing**: Services connect using `127.0.0.1` and static ports
- **No External Dependencies**: Only requires Nomad and Docker
- **Simple Configuration**: Minimal setup and configuration
- **Fast Startup**: Quick deployment without waiting for service registration

### Use Cases
- Local development
- Testing and CI/CD
- Simple production deployments
- Learning Nomad basics

### Services
- **PostgreSQL**: `localhost:5432`
- **Redis**: `localhost:6379`  
- **GoTAK API**: `localhost:8080`
- **GoTAK CoT TCP**: `localhost:8087`
- **GoTAK CoT TLS**: `localhost:8089`

### Deployment
```bash
# Basic deployment
./nomad/scripts/deploy-universal.sh standalone

# With specific environment
./nomad/scripts/deploy-universal.sh standalone dev

# Dry run
./nomad/scripts/deploy-universal.sh standalone dev true
```

## 🌐 Consul-Integrated Deployment

### Features
- **Service Discovery**: Automatic service registration and discovery
- **DNS Integration**: Services resolve via `postgres.service.consul`
- **Health Monitoring**: Comprehensive health checking and monitoring
- **Configuration Templates**: Dynamic config generation based on service state
- **Service Mesh Ready**: Optional Consul Connect integration

### Use Cases
- Production deployments
- Multi-environment setups
- Advanced networking requirements
- Service mesh architectures
- High availability deployments

### Services
- **Consul**: `localhost:8500` (UI: `/ui`)
- **PostgreSQL**: Auto-discovered via Consul DNS
- **Redis**: Auto-discovered via Consul DNS
- **GoTAK API**: `localhost:8080` (registered in Consul)
- **GoTAK CoT Protocols**: `localhost:8087`, `localhost:8089`

### Deployment
```bash
# Deploy with Consul (includes Consul deployment)
./nomad/scripts/deploy-universal.sh consul

# Production environment
./nomad/scripts/deploy-universal.sh consul prod

# Dry run
./nomad/scripts/deploy-universal.sh consul dev true
```

## 🛠️ Management Commands

### Universal Script Usage
```bash
# Get help
./nomad/scripts/deploy-universal.sh help

# Check deployment status
./nomad/scripts/deploy-universal.sh status

# Stop all services
./nomad/scripts/deploy-universal.sh stop

# Deploy standalone
./nomad/scripts/deploy-universal.sh standalone dev

# Deploy with Consul
./nomad/scripts/deploy-universal.sh consul dev
```

### Individual Service Management
```bash
# Check specific job
nomad job status gotak-server

# View logs
nomad alloc logs -f -job gotak-server

# Scale services
nomad job scale gotak-server 2

# Stop specific service
nomad job stop gotak-postgres
```

## 📊 Service Discovery Examples

### Standalone (Direct IP)
```yaml
database:
  host: "127.0.0.1"
  port: 5432

redis:
  host: "127.0.0.1" 
  port: 6379
```

### Consul (Service Discovery)
```yaml
database:
  host: "{{ range service "postgres" }}{{ .Address }}{{ end }}"
  port: {{ range service "postgres" }}{{ .Port }}{{ end }}

redis:
  host: "{{ range service "redis" }}{{ .Address }}{{ end }}"
  port: {{ range service "redis" }}{{ .Port }}{{ end }}
```

## 🔧 Configuration Customization

### Environment Variables
Both deployment types support environment-specific configuration through variables:

- **Development**: Lower resource allocation, debug enabled
- **Staging**: Moderate resources, production-like config
- **Production**: Full resources, security hardened

### Resource Allocation
Edit the job files to adjust resource allocation:

```hcl
resources {
  cpu    = 500  # MHz
  memory = 512  # MB
}
```

## 🚨 Troubleshooting

### Common Issues

**Standalone Deployment**:
- **Port conflicts**: Ensure ports 5432, 6379, 8080, 8087, 8089 are available
- **Volume issues**: Check that `/tmp/gotak-nomad-volumes/dev/*` directories exist
- **Service startup order**: PostgreSQL → Redis → GoTAK Server

**Consul Deployment**:
- **Consul not starting**: Check if port 8500 is available
- **Service registration fails**: Verify Consul is healthy before deploying services
- **Template rendering errors**: Check Consul service registration status

### Debug Commands
```bash
# Check Nomad cluster status
nomad node status

# Check Consul services (if using Consul deployment)
consul catalog services

# View detailed allocation information
nomad alloc status <allocation-id>

# Check service health in Consul
consul health service postgres
```

## 🔄 Migration Path

### From Standalone to Consul

1. **Stop standalone deployment**:
   ```bash
   ./nomad/scripts/deploy-universal.sh stop
   ```

2. **Deploy Consul version**:
   ```bash
   ./nomad/scripts/deploy-universal.sh consul dev
   ```

3. **Data persistence**: Volume data is preserved during migration

### From Consul to Standalone

1. **Export service configurations** (if customized)
2. **Stop Consul deployment**
3. **Deploy standalone version**
4. **Update any hardcoded service references**

## 📈 Production Considerations

### Standalone Production
- Use external PostgreSQL and Redis instances
- Implement proper backup strategies
- Configure monitoring and alerting
- Use TLS for all connections

### Consul Production  
- Deploy Consul in HA mode (3+ nodes)
- Enable Consul Connect for service mesh
- Implement proper ACL policies
- Configure backup and disaster recovery
- Set up monitoring for Consul cluster

## 🔗 Related Documentation

- [Nomad Job Specification](https://developer.hashicorp.com/nomad/docs/job-specification)
- [Consul Service Discovery](https://developer.hashicorp.com/consul/docs/discovery)
- [Consul Connect Service Mesh](https://developer.hashicorp.com/consul/docs/connect)
- [GoTAK Configuration Guide](../../README.md)

## 💡 Best Practices

1. **Start Simple**: Begin with standalone deployment for development
2. **Upgrade Gradually**: Move to Consul deployment for production
3. **Monitor Everything**: Use both Nomad and Consul monitoring
4. **Plan for Scale**: Design service allocation for expected load
5. **Security First**: Enable TLS and proper authentication in production
6. **Backup Strategy**: Regular backups of PostgreSQL data and Consul state