#!/usr/bin/env bash
# Refresh the vault-aws-creds Secret from the current AWS env creds, then bounce
# Vault so the KMS seal picks them up.
#
# Why: sandbox STS creds expire (~12h). Vault's AWS KMS seal reads creds at
# pod-start. If vault-0 restarts after the creds in the Secret have expired, it
# cannot auto-unseal. Run this whenever you refresh your AWS session creds (and
# certainly before any planned Vault restart).
#
# Usage:  export AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... AWS_SESSION_TOKEN=...
#         ./refresh-vault-aws-creds.sh
set -euo pipefail
NS="${NS:-gotak}"

: "${AWS_ACCESS_KEY_ID:?export AWS_ACCESS_KEY_ID first}"
: "${AWS_SECRET_ACCESS_KEY:?export AWS_SECRET_ACCESS_KEY first}"

echo ">> Updating vault-aws-creds Secret in $NS"
oc create secret generic vault-aws-creds -n "$NS" \
  --from-literal=AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
  --from-literal=AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
  --from-literal=AWS_SESSION_TOKEN="${AWS_SESSION_TOKEN:-}" \
  --dry-run=client -o yaml | oc apply -f -

echo ">> Restarting vault-0 (re-reads creds, auto-unseals via KMS)"
oc -n "$NS" delete pod vault-0 --wait=false
oc -n "$NS" wait --for=jsonpath='{.status.phase}'=Running pod/vault-0 --timeout=120s || true
sleep 5
oc -n "$NS" exec vault-0 -- sh -c 'vault status -format=json' 2>/dev/null \
  | grep -E '"sealed"|"initialized"' || true
echo ">> Done."
