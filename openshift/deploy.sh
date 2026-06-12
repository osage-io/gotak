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
oc apply -f "$DIR/30-routes.yaml"

# If the HashiStack platform is installed, point the web UI's Vault address at
# the live `vault` Route so the browser talks to in-cluster Vault out of the box.
VAULT_HOST="$(oc get route vault -n "$PROJECT" -o jsonpath='{.spec.host}' 2>/dev/null || true)"
if [ -n "$VAULT_HOST" ]; then
  echo ">> Wiring web Vault address -> https://$VAULT_HOST"
  oc patch configmap gotak-web-config -n "$PROJECT" --type merge \
    -p "{\"data\":{\"VAULT_ADDR\":\"https://$VAULT_HOST\"}}"
  oc rollout restart deploy/gotak-web -n "$PROJECT" >/dev/null 2>&1 || true
else
  echo ">> (no vault Route yet — web Vault address left as configured default)"
fi

echo ">> Waiting for rollouts"
oc rollout status deploy/postgres --timeout=120s || true
oc rollout status deploy/gotak-server --timeout=120s || true
oc rollout status deploy/gotak-web --timeout=120s || true

echo ">> Routes:"
oc get route -o custom-columns=NAME:.metadata.name,HOST:.spec.host
echo ">> Done."
