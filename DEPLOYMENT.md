# GoTAK Deployment Guide

## Components Overview

### Required Components
1. **PostgreSQL with PostGIS** - Primary database for all persistent data
2. **Redis** - Session management, caching, and temporary data
3. **GoTAK Server** - Main application server

### Optional Components
4. **NATS** - Message broker for events (only if using event-driven features)
5. **Web UI** - Browser-based interface (can be deployed separately)

### Development-Only Components (NOT for production)
- **postgres-test** - Test database container
- **jaeger** - Distributed tracing (use proper APM in production)
- **adminer** - Database admin UI (use proper tools in production)

## Production Deployment

### 1. Prerequisites
- Docker and Docker Compose installed
- Domain name with SSL certificates
- Minimum 2GB RAM, 2 CPU cores
- 20GB disk space

### 2. Setup Steps

```bash
# Clone repository
git clone https://github.com/yourusername/gotak.git
cd gotak

# Copy environment template
cp .env.example .env

# Edit .env with production values
nano .env

# Generate secure passwords
openssl rand -base64 32  # For POSTGRES_PASSWORD
openssl rand -base64 32  # For REDIS_PASSWORD
openssl rand -base64 64  # For JWT_SECRET

# Create necessary directories
mkdir -p certs logs data

# Add SSL certificates
cp /path/to/your/cert.pem certs/server.crt
cp /path/to/your/key.pem certs/server.key
chmod 600 certs/server.key
```

### 3. Database Initialization

```bash
# Start only the database first
docker-compose up -d postgres

# Wait for it to be ready
docker-compose exec postgres pg_isready -U gotak

# Run migrations
docker-compose run --rm gotak /app/migrate.sh up
```

### 4. Start Services

```bash
# Start all required services
docker-compose up -d postgres redis gotak

# Optional: Start NATS if using events
docker-compose up -d nats

# Optional: Start web UI
docker-compose up -d web

# Check status
docker-compose ps

# View logs
docker-compose logs -f gotak
```

### 5. Building Images for Push

```bash
# Build production images
docker-compose build

# Tag images for registry
docker tag gotak/server:latest your-registry.com/gotak/server:latest
docker tag gotak/web:latest your-registry.com/gotak/web:latest

# Push to registry
docker push your-registry.com/gotak/server:latest
docker push your-registry.com/gotak/web:latest
```

## Minimal Production Setup

For the absolute minimum production deployment, you only need:

```yaml
# docker-compose.minimal.yml
version: '3.8'

services:
  postgres:
    image: postgis/postgis:15-3.4
    environment:
      POSTGRES_DB: gotak
      POSTGRES_USER: gotak
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    restart: unless-stopped

  gotak:
    image: gotak/server:latest
    ports:
      - "8087:8087"  # TAK protocol
      - "8089:8089"  # TAK TLS
      - "8080:8080"  # HTTP API
    environment:
      DATABASE_URL: postgres://gotak:${POSTGRES_PASSWORD}@postgres:5432/gotak
      REDIS_URL: redis://:${REDIS_PASSWORD}@redis:6379
      JWT_SECRET: ${JWT_SECRET}
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

## Security Considerations

1. **Change all default passwords** in production
2. **Use SSL/TLS** for all external connections
3. **Configure firewall** to only expose necessary ports
4. **Use secrets management** (Docker Secrets, Vault, etc.) for sensitive data
5. **Regular backups** of PostgreSQL data
6. **Monitor logs** for security events

## Monitoring

### Health Checks
- GoTAK Server: `http://your-domain:8080/health`
- PostgreSQL: `pg_isready -U gotak`
- Redis: `redis-cli ping`
- NATS: `http://your-domain:8222/healthz`

### Recommended Monitoring Stack
- Prometheus for metrics
- Grafana for visualization
- Loki for log aggregation
- AlertManager for notifications

## Backup & Recovery

### Database Backup
```bash
# Backup
docker-compose exec postgres pg_dump -U gotak gotak > backup_$(date +%Y%m%d).sql

# Restore
docker-compose exec -T postgres psql -U gotak gotak < backup_20240101.sql
```

### Full System Backup
```bash
# Stop services
docker-compose down

# Backup volumes
docker run --rm -v gotak_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_data.tar.gz /data
docker run --rm -v gotak_redis_data:/data -v $(pwd):/backup alpine tar czf /backup/redis_data.tar.gz /data

# Start services
docker-compose up -d
```

## Scaling

### Horizontal Scaling
- Multiple GoTAK server instances behind a load balancer
- Redis Cluster or Redis Sentinel for HA
- PostgreSQL replication for read scaling
- NATS clustering for messaging HA

### Vertical Scaling
Adjust Docker resource limits in compose file:
```yaml
gotak:
  deploy:
    resources:
      limits:
        cpus: '2'
        memory: 2G
      reservations:
        cpus: '1'
        memory: 1G
```

## Troubleshooting

### Common Issues

1. **Database connection errors**
   - Check PostgreSQL is running: `docker-compose ps postgres`
   - Verify credentials in .env
   - Check network connectivity

2. **Port conflicts**
   - Change port mappings in docker-compose.yml
   - Check for other services using same ports

3. **Memory issues**
   - Increase Docker memory allocation
   - Add swap space on host
   - Optimize PostgreSQL settings

### Logs
```bash
# All logs
docker-compose logs

# Specific service
docker-compose logs gotak

# Follow logs
docker-compose logs -f

# Last 100 lines
docker-compose logs --tail=100
```

## Support

For production support and enterprise features, contact: support@gotak.io