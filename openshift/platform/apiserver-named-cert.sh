#!/usr/bin/env bash
# Serve your *.demoland.io wildcard cert on the OpenShift API for the hostname
# sno.demoland.io, so `oc login https://sno.demoland.io:6443` validates cleanly.
#
#   export KUBECONFIG=../../iac/sno/cluster-auth/kubeconfig
#   TLS_CERT=~/.certs/demoland.crt TLS_KEY=~/.certs/demoland.key \
#   [API_HOST=sno.demoland.io] ./apiserver-named-cert.sh
#
# How it works: the cert goes in a Secret in openshift-config, then the cluster's
# APIServer config gets a namedCertificate entry binding that Secret to API_HOST.
# The kube-apiserver keeps its built-in cert for the canonical api.<base-domain>
# name and serves yours for the sno.demoland.io SNI.
#
# ⚠️ This triggers a kube-apiserver rollout. On single-node that's the only control
#    plane — expect a few-minute API blip. Do it outside a live demo.
set -euo pipefail

API_HOST="${API_HOST:-sno.demoland.io}"
SECRET="${SECRET:-demoland-api-cert}"
CERT="${TLS_CERT:?set TLS_CERT to your PEM cert (e.g. *.demoland.io)}"
KEY="${TLS_KEY:?set TLS_KEY to your PEM private key}"
[ -f "$CERT" ] || { echo "cert not found: $CERT"; exit 1; }
[ -f "$KEY" ]  || { echo "key not found: $KEY"; exit 1; }

echo ">> Creating TLS Secret $SECRET in openshift-config"
oc create secret tls "$SECRET" --cert="$CERT" --key="$KEY" -n openshift-config \
  --dry-run=client -o yaml | oc apply -f -

echo ">> Binding $API_HOST -> $SECRET on the APIServer config"
oc patch apiserver cluster --type merge -p "{\"spec\":{\"servingCerts\":{\"namedCertificates\":[
  {\"names\":[\"$API_HOST\"],\"servingCertificate\":{\"name\":\"$SECRET\"}}]}}}"

cat <<EOF

>> Applied. The kube-apiserver will roll out now (single-node => brief API blip).
   Watch it settle:
     oc get clusteroperator kube-apiserver -w     # PROGRESSING true -> false

   Add the DNS record (Namecheap):
     A    sno    16.58.42.236        # sno.demoland.io -> kube-apiserver

   Then log in cleanly:
     oc login https://$API_HOST:6443 -u kubeadmin
EOF
