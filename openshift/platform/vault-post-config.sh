#!/usr/bin/env bash
# Initialise (if needed) and configure Vault for gotak:
#   - operator init (auto-unseal via KMS, so no manual unseal)
#   - enable transit, pki, kv-v2
#   - enable CORS for the gotak web Route origin (browser calls Vault directly)
#   - create a gotak policy + token
#
# Idempotent: safe to re-run. Root token + recovery keys are saved locally to
# .vault-init.json (gitignored) on first init.
set -euo pipefail

NS="${NS:-gotak}"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INIT_FILE="$DIR/.vault-init.json"
POD="vault-0"

vex() { oc -n "$NS" exec "$POD" -- sh -c "$1"; }

echo ">> Checking Vault init status"
if vex 'vault status -format=json' 2>/dev/null | grep -q '"initialized": true'; then
  echo "   already initialised"
else
  echo ">> Initialising Vault (KMS auto-unseal, 1 recovery key)"
  vex 'VAULT_ADDR=http://127.0.0.1:8200 vault operator init \
        -recovery-shares=1 -recovery-threshold=1 -format=json' > "$INIT_FILE"
  chmod 600 "$INIT_FILE"
  echo "   saved recovery key + root token -> $INIT_FILE"
  sleep 5  # give KMS auto-unseal a moment
fi

ROOT_TOKEN="$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["root_token"])' "$INIT_FILE")"
RUN="VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=$ROOT_TOKEN vault"

echo ">> Enabling secrets engines (transit, pki, kv-v2)"
vex "$RUN secrets enable -path=transit transit"            2>/dev/null || echo "   transit already enabled"
vex "$RUN secrets enable -path=pki pki"                    2>/dev/null || echo "   pki already enabled"
vex "$RUN secrets enable -path=secret -version=2 kv"       2>/dev/null || echo "   kv already enabled"
vex "$RUN secrets tune -max-lease-ttl=8760h pki"           2>/dev/null || true

# Discover the Vault Route host so we can scope CORS to the gotak origin.
WEB_HOST="$(oc -n "$NS" get route gotak-web -o jsonpath='{.spec.host}' 2>/dev/null || true)"
VAULT_HOST="$(oc -n "$NS" get route vault -o jsonpath='{.spec.host}' 2>/dev/null || true)"
ORIGINS="*"
if [ -n "$WEB_HOST" ]; then ORIGINS="https://$WEB_HOST"; fi

echo ">> Enabling CORS (allowed origin: $ORIGINS)"
# The browser sends X-Vault-Token; allow it explicitly.
vex "$RUN write sys/config/cors enabled=true \
      allowed_origins='$ORIGINS' \
      allowed_headers='X-Vault-Token,Content-Type,X-Vault-Namespace'" >/dev/null

echo ">> Creating gotak policy + token"
vex "cat > /tmp/gotak-policy.hcl <<'EOF'
path \"transit/*\"     { capabilities = [\"create\",\"read\",\"update\",\"delete\",\"list\"] }
path \"pki/*\"         { capabilities = [\"create\",\"read\",\"update\",\"delete\",\"list\"] }
path \"secret/data/*\" { capabilities = [\"create\",\"read\",\"update\",\"delete\",\"list\"] }
path \"secret/metadata/*\" { capabilities = [\"read\",\"list\"] }
EOF"
vex "$RUN policy write gotak /tmp/gotak-policy.hcl" >/dev/null
GOTAK_TOKEN="$(vex "$RUN token create -policy=gotak -ttl=720h -format=json" | python3 -c 'import json,sys;print(json.load(sys.stdin)["auth"]["client_token"])')"
echo "$GOTAK_TOKEN" > "$DIR/.vault-gotak-token"
chmod 600 "$DIR/.vault-gotak-token"

echo ""
echo ">> Vault configured."
[ -n "$VAULT_HOST" ] && echo "   Vault URL (use in gotak config):  https://$VAULT_HOST"
echo "   gotak token saved -> $DIR/.vault-gotak-token"
echo "   (root token in $INIT_FILE)"
