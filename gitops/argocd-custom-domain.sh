#!/usr/bin/env bash
# Put the Argo CD console behind a demoland.io hostname with the BYO wildcard cert.
#
#   export KUBECONFIG=../iac/sno/cluster-auth/kubeconfig
#   TLS_CERT=~/.certs/demoland-fullchain.crt TLS_KEY=~/.certs/demoland.key \
#   [ARGOCD_HOST=argocd.demoland.io] ./argocd-custom-domain.sh
#
# Argo CD is operator-managed, so we set the host + route TLS on the ArgoCD CR
# (not the Route directly — the operator would revert that). `insecure: true`
# makes argocd-server serve HTTP internally; the edge Route terminates TLS with
# your cert at the router. Triggers an argocd-server restart.
set -euo pipefail

NS="${NS:-openshift-gitops}"
CR="${CR:-openshift-gitops}"
HOST="${ARGOCD_HOST:-argocd.demoland.io}"
CERT="${TLS_CERT:?set TLS_CERT (e.g. ~/.certs/demoland-fullchain.crt)}"
KEY="${TLS_KEY:?set TLS_KEY (e.g. ~/.certs/demoland.key)}"
[ -f "$CERT" ] || { echo "cert not found: $CERT"; exit 1; }
[ -f "$KEY" ]  || { echo "key not found: $KEY"; exit 1; }

PATCH="$(mktemp)"; trap 'rm -f "$PATCH"' EXIT
ind() { sed 's/^/          /'; }   # 10-space indent for block scalars
{
  echo "spec:"
  echo "  server:"
  echo "    host: $HOST"
  echo "    insecure: true"
  echo "    route:"
  echo "      enabled: true"
  echo "      tls:"
  echo "        termination: edge"
  echo "        insecureEdgeTerminationPolicy: Redirect"
  echo "        certificate: |"
  ind < "$CERT"
  echo "        key: |"
  ind < "$KEY"
} > "$PATCH"

echo ">> Patching ArgoCD/$CR -> host $HOST (edge TLS, BYO cert)"
oc patch argocd "$CR" -n "$NS" --type merge --patch-file "$PATCH"

echo ">> Waiting for the operator to reconcile the Route + restart argocd-server"
oc rollout status deploy/openshift-gitops-server -n "$NS" --timeout=240s || true
sleep 5
echo ">> Route now:"
oc get route openshift-gitops-server -n "$NS" \
  -o custom-columns=HOST:.spec.host,TLS:.spec.tls.termination 2>/dev/null

cat <<EOF

>> Console: https://$HOST
   DNS (Namecheap):  A  argocd  ->  18.217.224.91   (router IP; add 18.116.145.144 too if you like)
   admin password unchanged:
     oc get secret openshift-gitops-cluster -n $NS -o jsonpath='{.data.admin\.password}' | base64 -d
EOF
