#!/usr/bin/env bash
set -euo pipefail

echo "🛑 Stopping HashiCorp services..."

# Stop Consul
consul_pids=$(pgrep -f "consul agent" || true)
if [[ -n "$consul_pids" ]]; then
    echo "📡 Stopping Consul..."
    for pid in $consul_pids; do
        kill "$pid" 2>/dev/null || true
    done
    echo "✅ Consul stopped"
else
    echo "ℹ️ Consul not running"
fi

# Stop Vault
vault_pids=$(pgrep -f "vault server" || true)
if [[ -n "$vault_pids" ]]; then
    echo "🔐 Stopping Vault..."
    for pid in $vault_pids; do
        kill "$pid" 2>/dev/null || true
    done
    echo "✅ Vault stopped"
else
    echo "ℹ️ Vault not running"
fi

# Cleanup temporary files
if [[ -f /tmp/consul.log ]]; then
    rm /tmp/consul.log
    echo "🧹 Cleaned up Consul log"
fi

if [[ -f /tmp/vault.log ]]; then
    rm /tmp/vault.log
    echo "🧹 Cleaned up Vault log"
fi

if [[ -d /tmp/consul-gotak ]]; then
    rm -rf /tmp/consul-gotak
    echo "🧹 Cleaned up Consul data directory"
fi

echo ""
echo "✅ HashiCorp services stopped and cleaned up"
echo ""
echo "To stop your Docker containers:"
echo "  make dev-down"
echo "  # or: docker compose -f docker-compose.dev.yml down -v"