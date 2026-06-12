#!/usr/bin/env bash
# Create a scoped ServiceAccount + long-lived token that the TFC kubernetes
# provider uses to manage ONLY secrets in the gotak namespace. Run once with
# cluster-admin (your kubeconfig), then paste the printed values into TFC.
#
#   export KUBECONFIG=../sno/cluster-auth/kubeconfig   # adjust path
#   ./bootstrap-sa.sh
set -euo pipefail
NS="${NS:-gotak}"
SA="${SA:-tf-registry}"

oc get ns "$NS" >/dev/null 2>&1 || oc create ns "$NS"

oc apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata: { name: $SA, namespace: $NS }
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata: { name: $SA-secrets, namespace: $NS }
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get","list","create","update","patch","delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata: { name: $SA-secrets, namespace: $NS }
roleRef: { apiGroup: rbac.authorization.k8s.io, kind: Role, name: $SA-secrets }
subjects:
  - kind: ServiceAccount
    name: $SA
    namespace: $NS
---
apiVersion: v1
kind: Secret
metadata:
  name: $SA-token
  namespace: $NS
  annotations:
    kubernetes.io/service-account.name: $SA
type: kubernetes.io/service-account-token
EOF

# Wait for the controller to populate the token.
for i in $(seq 1 10); do
  TOKEN=$(oc get secret "$SA-token" -n "$NS" -o jsonpath='{.data.token}' 2>/dev/null | base64 -d || true)
  [ -n "$TOKEN" ] && break; sleep 2
done

cat <<EOF

>> Set these in the TFC workspace (k8s_token = SENSITIVE):

   k8s_host  = https://sno.demoland.io:6443
   k8s_token = $TOKEN

   (host uses the named-cert name so the public Sectigo cert validates — no CA needed.)
EOF
