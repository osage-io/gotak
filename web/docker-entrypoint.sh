#!/bin/sh

# GoTAK Web UI Docker Entrypoint Script
# This script allows for runtime configuration of the web application

set -e

# Default environment variables
GOTAK_SERVER_URL=${GOTAK_SERVER_URL:-"ws://localhost:8087"}
GOTAK_API_URL=${GOTAK_API_URL:-"http://localhost:8080"}
GOTAK_WS_URL=${GOTAK_WS_URL:-"ws://localhost:8087/ws"}
# Vault address the browser talks to directly (transit/PKI/KV). In-cluster this is
# the Vault Route; locally it defaults to the dev Vault on 127.0.0.1:8200.
VAULT_ADDR=${VAULT_ADDR:-"http://127.0.0.1:8200"}

# Create runtime configuration file
cat > /usr/share/nginx/html/config/runtime-config.js << EOF
window.GOTAK_CONFIG = {
  serverUrl: '${GOTAK_SERVER_URL}',
  apiUrl: '${GOTAK_API_URL}',
  wsUrl: '${GOTAK_WS_URL}',
  vaultUrl: '${VAULT_ADDR}',
  mapConfig: {
    defaultCenter: [${MAP_DEFAULT_LAT:-38.9072}, ${MAP_DEFAULT_LNG:--77.0369}],
    defaultZoom: ${MAP_DEFAULT_ZOOM:-12},
    maxZoom: ${MAP_MAX_ZOOM:-18},
    minZoom: ${MAP_MIN_ZOOM:-3}
  },
  features: {
    chat: ${ENABLE_CHAT:-true},
    drawing: ${ENABLE_DRAWING:-true},
    measurements: ${ENABLE_MEASUREMENTS:-true},
    geofencing: ${ENABLE_GEOFENCING:-true}
  }
};
EOF

echo "Runtime configuration created:"
cat /usr/share/nginx/html/config/runtime-config.js

# Add the configuration script to index.html if it doesn't exist
if ! grep -q "runtime-config.js" /usr/share/nginx/html/index.html; then
  sed -i 's|</head>|  <script src="/config/runtime-config.js"></script>\n  </head>|' /usr/share/nginx/html/index.html
fi

echo "Starting nginx with GoTAK Web UI..."
echo "Configuration:"
echo "  Server URL: ${GOTAK_SERVER_URL}"
echo "  API URL: ${GOTAK_API_URL}"
echo "  WebSocket URL: ${GOTAK_WS_URL}"

# Start nginx
exec "$@"
