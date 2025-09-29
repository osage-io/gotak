#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "🚀 Starting HashiCorp tools for GoTAK development..."

# Start Consul in dev mode
echo "📡 Starting Consul..."
consul agent -dev \
  -client 0.0.0.0 \
  -bind 127.0.0.1 \
  -ui \
  -node gotak-dev \
  -data-dir /tmp/consul-gotak > /tmp/consul.log 2>&1 &

consul_pid=$!
echo "✅ Consul started (PID: $consul_pid) - UI at http://localhost:8500"

# Start Vault in dev mode
echo "🔐 Starting Vault..."
vault server -dev \
  -dev-root-token-id=root \
  -dev-listen-address="0.0.0.0:8200" > /tmp/vault.log 2>&1 &

vault_pid=$!
echo "✅ Vault started (PID: $vault_pid) - UI at http://localhost:8200"

# Wait for services to start
echo "⏳ Waiting for services to be ready..."
sleep 5

# Set environment variables
export CONSUL_HTTP_ADDR=http://localhost:8500
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root

# Verify services are running
if curl -sf http://localhost:8500/v1/status/leader > /dev/null; then
    echo "✅ Consul is ready"
else
    echo "❌ Consul failed to start"
    exit 1
fi

if vault status > /dev/null 2>&1; then
    echo "✅ Vault is ready"
else
    echo "❌ Vault failed to start"
    exit 1
fi

echo ""
echo "🎉 HashiCorp services are running!"
echo ""
echo "Environment variables:"
echo "  export CONSUL_HTTP_ADDR=http://localhost:8500"
echo "  export VAULT_ADDR=http://127.0.0.1:8200"
echo "  export VAULT_TOKEN=root"
echo ""
echo "Web UIs:"
echo "  📡 Consul: http://localhost:8500/ui"
echo "  🔐 Vault:  http://localhost:8200/ui"
echo ""
echo "To register your Docker services with Consul, run:"
echo "  ./scripts/register-consul.sh"
echo ""
echo "To stop services:"
echo "  ./scripts/stop-hashicorp.sh"