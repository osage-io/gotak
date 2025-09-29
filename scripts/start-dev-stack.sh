#!/bin/bash

# Start the GoTAK development stack including the web UI
# This script builds and starts all required containers

set -e

# Enable color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting GoTAK development stack...${NC}"
echo -e "${YELLOW}This will build and start all required services, including:${NC}"
echo "  - PostgreSQL database"
echo "  - Redis cache"
echo "  - NATS messaging"
echo "  - Jaeger tracing"
echo "  - GoTAK server"
echo "  - GoTAK web UI"
echo

# Check if the config directory exists
if [ ! -d "./config" ]; then
  echo -e "${YELLOW}Config directory not found. Creating sample configurations...${NC}"
  mkdir -p ./config
  cp ./config.example/* ./config/ 2>/dev/null || echo "No example configs found. Please create configs manually."
fi

# Check if certificates are available
if [ ! -d "./certs" ] || [ ! -f "./certs/server.crt" ]; then
  echo -e "${YELLOW}Certificates not found. Generating self-signed certificates...${NC}"
  make certs
fi

# Make sure directories exist
mkdir -p ./logs ./data

# Start all services
echo -e "${GREEN}Building and starting containers...${NC}"
docker-compose -f docker-compose.dev.yml up -d --build

# Show running services
echo
echo -e "${GREEN}Services started successfully!${NC}"
echo -e "${YELLOW}Services available at:${NC}"
echo "  - GoTAK Server:"
echo "    * API: http://localhost:8080"
echo "    * TAK TCP/UDP: localhost:8087"
echo "    * TAK TLS: localhost:8089"
echo "  - GoTAK Web UI: http://localhost:3000"
echo "  - PostgreSQL: localhost:5432 (user: gotak, password: dev_password, db: gotak_dev)"
echo "  - Redis: localhost:6379 (password: dev_redis_pass)"
echo "  - NATS: localhost:4222 (monitoring: http://localhost:8222)"
echo "  - Jaeger UI: http://localhost:16686"
echo "  - Adminer: http://localhost:8081 (database admin)"
echo
echo -e "${YELLOW}To view logs:${NC}"
echo "  docker-compose -f docker-compose.dev.yml logs -f"
echo
echo -e "${YELLOW}To stop the stack:${NC}"
echo "  docker-compose -f docker-compose.dev.yml down"
echo
echo -e "${GREEN}Done!${NC}"
