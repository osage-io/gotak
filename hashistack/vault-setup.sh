#!/usr/bin/env bash
# Configure the local dev Vault for the GoTAK comms-encryption demo:
#   - enable the transit secrets engine (if not already)
#   - create the transit key used to encrypt comms messages
#   - enable CORS so the browser UI can call Vault's transit endpoint directly
#
# The dev Vault (`vault server -dev`) keeps everything in memory, so re-run this
# after every Vault restart. Safe to run repeatedly (idempotent).
#
# Usage:  ./hashistack/vault-setup.sh
set -euo pipefail

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
export VAULT_TOKEN="${VAULT_TOKEN:-root}"
TRANSIT_KEY="${TRANSIT_KEY:-gotak-comms}"
# Browser origin allowed to call Vault. "*" keeps the demo simple; tighten to the
# UI origin (e.g. http://localhost:8080) for anything less throwaway.
CORS_ORIGINS="${CORS_ORIGINS:-*}"

command -v vault >/dev/null || { echo "missing dependency: vault" >&2; exit 1; }

echo "[vault-setup] enabling transit engine (if needed)…"
if ! vault secrets list -format=json 2>/dev/null | grep -q '"transit/"'; then
  vault secrets enable transit
else
  echo "[vault-setup]   transit already enabled"
fi

echo "[vault-setup] creating transit key '${TRANSIT_KEY}' (if needed)…"
vault write -f "transit/keys/${TRANSIT_KEY}" >/dev/null
vault read "transit/keys/${TRANSIT_KEY}" | grep -E 'name|type|latest_version' | sed 's/^/[vault-setup]   /'

echo "[vault-setup] enabling PKI engine for device certificates…"
if ! vault secrets list -format=json 2>/dev/null | grep -q '"pki/"'; then
  vault secrets enable pki
  vault secrets tune -max-lease-ttl=87600h pki
  vault write -field=certificate pki/root/generate/internal \
    common_name="GoTAK Demo Root CA" issuer_name="gotak-root" ttl=87600h >/dev/null
  vault write pki/config/urls \
    issuing_certificates="${VAULT_ADDR}/v1/pki/ca" \
    crl_distribution_points="${VAULT_ADDR}/v1/pki/crl" >/dev/null
else
  echo "[vault-setup]   pki already enabled"
fi
# Device role: client certs, any common name, 7-day default / 30-day max.
vault write pki/roles/gotak-device \
  allow_any_name=true allow_subdomains=true enforce_hostnames=false \
  client_flag=true server_flag=false \
  key_type=rsa key_bits=2048 ttl=168h max_ttl=720h >/dev/null
echo "[vault-setup]   pki role 'gotak-device' ready"

echo "[vault-setup] enabling CORS for browser access…"
vault write sys/config/cors \
  enabled=true \
  allowed_origins="${CORS_ORIGINS}" \
  allowed_headers="X-Vault-Token,Content-Type" >/dev/null

echo "[vault-setup] round-trip check…"
CT="$(vault write -field=ciphertext "transit/encrypt/${TRANSIT_KEY}" plaintext="$(printf 'demo' | base64)")"
PT="$(vault write -field=plaintext "transit/decrypt/${TRANSIT_KEY}" ciphertext="${CT}" | base64 -d)"
[ "${PT}" = "demo" ] && echo "[vault-setup]   encrypt/decrypt OK (${CT})" || { echo "round-trip FAILED" >&2; exit 1; }

cat <<EOF

[vault-setup] Done.
  VAULT_ADDR  : ${VAULT_ADDR}
  Transit key : ${TRANSIT_KEY}
  CORS origins: ${CORS_ORIGINS}

The Communications page "Encrypt" button calls:
  POST ${VAULT_ADDR}/v1/transit/encrypt/${TRANSIT_KEY}
EOF
