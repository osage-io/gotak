#!/usr/bin/env bash
# Deploy GoTAK to OpenShift (demo-grade). Targets OpenShift Local (CRC) or any cluster.
# Usage: ./deploy.sh [project]   (default project: gotak)
set -euo pipefail

PROJECT="${1:-gotak}"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

command -v oc >/dev/null || { echo "oc CLI not found"; exit 1; }
oc whoami >/dev/null || { echo "Not logged in. Run 'oc login' first."; exit 1; }

echo ">> Project: $PROJECT"
oc get project "$PROJECT" >/dev/null 2>&1 || oc new-project "$PROJECT" >/dev/null
oc project "$PROJECT" >/dev/null

# Postgres image (postgis) runs as a fixed UID -> grant anyuid to its SA.
echo ">> Granting anyuid SCC to gotak-postgres SA"
oc apply -f "$DIR/00-postgres.yaml"
oc adm policy add-scc-to-user anyuid -z gotak-postgres -n "$PROJECT" || \
  echo "   (skipped: need cluster-admin for anyuid; see README for alternative)"

echo ">> Applying manifests"
oc apply -f "$DIR/10-gotak-server.yaml"
oc apply -f "$DIR/20-gotak-web.yaml"
# Ingress is the Consul API Gateway (openshift/platform/consul-api-gateway.yaml),
# applied by install-platform.sh. 30-routes.yaml is documentation only now.

# Point the web UI's runtime config at the live in-cluster endpoints:
#  - Vault  -> the `vault` Route (browser calls Vault directly for transit/PKI/KV)
#  - API/WS -> the `gotak-gateway` Route (single hostname, path-routed)
patch_web() { oc patch configmap gotak-web-config -n "$PROJECT" --type merge -p "$1"; }

VAULT_HOST="$(oc get route vault -n "$PROJECT" -o jsonpath='{.spec.host}' 2>/dev/null || true)"
[ -n "$VAULT_HOST" ] && { echo ">> web Vault address -> https://$VAULT_HOST"; \
  patch_web "{\"data\":{\"VAULT_ADDR\":\"https://$VAULT_HOST\"}}"; } || \
  echo ">> (no vault Route yet — Vault address left as configured default)"

GW_HOST="$(oc get route gotak-gateway -n "$PROJECT" -o jsonpath='{.spec.host}' 2>/dev/null || true)"
if [ -n "$GW_HOST" ]; then
  echo ">> web API/WS endpoints -> https://$GW_HOST (via API Gateway)"
  # apiClient appends /api/v1 to apiUrl; websocket uses wsUrl verbatim.
  patch_web "{\"data\":{\"GOTAK_API_URL\":\"https://$GW_HOST\",\"GOTAK_WS_URL\":\"wss://$GW_HOST/ws\"}}"
else
  echo ">> (no gotak-gateway Route yet — run install-platform.sh first for ingress)"
fi
oc rollout restart deploy/gotak-web -n "$PROJECT" >/dev/null 2>&1 || true

echo ">> Waiting for rollouts"
oc rollout status deploy/postgres --timeout=120s || true
oc rollout status deploy/gotak-server --timeout=120s || true
oc rollout status deploy/gotak-web --timeout=120s || true

echo ">> Routes:"
oc get route -o custom-columns=NAME:.metadata.name,HOST:.spec.host
echo ">> Done."
