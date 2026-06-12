#!/usr/bin/env bash
# Deploy the HashiStack platform (Vault + Consul) into the gotak namespace on
# Single-Node OpenShift, with Vault auto-unsealed by AWS KMS.
#
# Order:  KMS key -> AWS creds Secret -> Vault (Helm) -> init+config -> Consul (Helm)
#
# Prereqs (on the machine running this):
#   - oc, logged in as cluster-admin (KUBECONFIG from the SNO install)
#   - helm 3
#   - aws CLI with live creds that can use KMS (the same sandbox creds)
#
# Usage: ./install-platform.sh
set -euo pipefail

NS="${NS:-gotak}"
REGION="${AWS_REGION:-us-east-2}"
KMS_ALIAS="${KMS_ALIAS:-alias/gotak-vault-unseal}"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

command -v oc   >/dev/null || { echo "oc not found";   exit 1; }
command -v helm >/dev/null || { echo "helm not found"; exit 1; }
command -v aws  >/dev/null || { echo "aws not found";  exit 1; }
oc whoami >/dev/null || { echo "Not logged in (oc login)"; exit 1; }

echo ">> Namespace: $NS"
oc get project "$NS" >/dev/null 2>&1 || oc new-project "$NS" >/dev/null

# ---------------------------------------------------------------------------
# 1. KMS key for auto-unseal (idempotent via an alias).
# ---------------------------------------------------------------------------
echo ">> Ensuring KMS key ($KMS_ALIAS)"
KMS_KEY_ID="$(aws kms describe-key --key-id "$KMS_ALIAS" --region "$REGION" \
                --query 'KeyMetadata.KeyId' --output text 2>/dev/null || true)"
if [ -z "$KMS_KEY_ID" ] || [ "$KMS_KEY_ID" = "None" ]; then
  KMS_KEY_ID="$(aws kms create-key --region "$REGION" \
                  --description "gotak Vault auto-unseal" \
                  --tags TagKey=app,TagValue=gotak \
                  --query 'KeyMetadata.KeyId' --output text)"
  aws kms create-alias --alias-name "$KMS_ALIAS" \
      --target-key-id "$KMS_KEY_ID" --region "$REGION"
  echo "   created $KMS_KEY_ID"
else
  echo "   reusing $KMS_KEY_ID"
fi

# ---------------------------------------------------------------------------
# 2. AWS creds Secret that Vault's KMS seal reads.
#    NOTE: sandbox STS creds expire (~12h). If the Vault pod restarts after they
#    expire it cannot auto-unseal until refreshed — see refresh-vault-aws-creds.sh.
# ---------------------------------------------------------------------------
echo ">> Writing vault-aws-creds Secret (from current AWS env/credentials)"
: "${AWS_ACCESS_KEY_ID:?export AWS_ACCESS_KEY_ID (and SECRET/SESSION_TOKEN) first}"
: "${AWS_SECRET_ACCESS_KEY:?export AWS_SECRET_ACCESS_KEY first}"
oc create secret generic vault-aws-creds -n "$NS" \
  --from-literal=AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
  --from-literal=AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
  --from-literal=AWS_SESSION_TOKEN="${AWS_SESSION_TOKEN:-}" \
  --dry-run=client -o yaml | oc apply -f -

# ---------------------------------------------------------------------------
# 3. Vault (Helm).
# ---------------------------------------------------------------------------
echo ">> Adding HashiCorp Helm repo"
helm repo add hashicorp https://helm.releases.hashicorp.com >/dev/null 2>&1 || true
helm repo update hashicorp >/dev/null

echo ">> Installing Vault (auto-unseal via KMS $KMS_KEY_ID)"
helm upgrade --install vault hashicorp/vault -n "$NS" \
  -f "$DIR/vault-values.yaml" \
  --set "server.extraEnvironmentVars.VAULT_AWSKMS_SEAL_KEY_ID=$KMS_KEY_ID" \
  --set "server.extraEnvironmentVars.AWS_REGION=$REGION" \
  --wait --timeout 5m || true   # pod stays un-ready until init; that's expected

echo ">> Waiting for the vault-0 pod to be Running"
oc -n "$NS" rollout status statefulset/vault --timeout=180s 2>/dev/null || true
oc -n "$NS" wait --for=jsonpath='{.status.phase}'=Running pod/vault-0 --timeout=180s || true

# ---------------------------------------------------------------------------
# 4. Initialise + configure Vault (transit, pki, kv, CORS, gotak token).
# ---------------------------------------------------------------------------
"$DIR/vault-post-config.sh"

# ---------------------------------------------------------------------------
# 5. Consul (Helm).
# ---------------------------------------------------------------------------
echo ">> Installing Consul (single server)"
helm upgrade --install consul hashicorp/consul -n "$NS" \
  -f "$DIR/consul-values.yaml" \
  --wait --timeout 5m || true
oc -n "$NS" rollout status statefulset/consul-server --timeout=180s 2>/dev/null || true

# Mesh intentions (default-deny + explicit allows). Wait for the CRD that the
# connect-inject controller registers, then apply.
echo ">> Applying service-mesh intentions (default-deny + allows)"
oc wait --for=condition=Established crd/serviceintentions.consul.hashicorp.com --timeout=120s 2>/dev/null || true
oc apply -f "$DIR/consul-intentions.yaml" || \
  echo "   (intentions will apply once the ServiceIntentions CRD is ready — re-run: oc apply -f consul-intentions.yaml)"

echo ""
echo ">> Platform up. Endpoints:"
oc -n "$NS" get route 2>/dev/null | awk 'NR==1 || /vault|consul/'
echo ">> Vault root token + recovery keys: $DIR/.vault-init.json (gitignored)"
