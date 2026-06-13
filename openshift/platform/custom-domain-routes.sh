#!/usr/bin/env bash
# Put demoland.io hostnames in front of the gotak HTTPS endpoints, using a
# bring-your-own TLS cert (a *.demoland.io wildcard covers all three).
#
#   export KUBECONFIG=../../iac/sno/cluster-auth/kubeconfig
#   TLS_CERT=~/certs/demoland.crt TLS_KEY=~/certs/demoland.key \
#   [TLS_CACERT=~/certs/chain.crt] [DOMAIN=demoland.io] ./custom-domain-routes.sh
#
# Creates edge Routes (cert inline) for:
#   gotak.<domain>  -> gotak-gateway   (the app, via the Consul API Gateway)
#   vault.<domain>  -> vault           (Vault UI/API)
#   consul.<domain> -> consul-ui       (Consul UI)
# then repoints the web UI + Vault CORS at the new app origin, and prints the DNS
# records to add in Namecheap.
set -euo pipefail

NS="${NS:-gotak}"
DOMAIN="${DOMAIN:-demoland.io}"
CERT="${TLS_CERT:?set TLS_CERT to your PEM cert (e.g. *.demoland.io fullchain)}"
KEY="${TLS_KEY:?set TLS_KEY to your PEM private key}"
CA="${TLS_CACERT:-}"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
[ -f "$CERT" ] || { echo "cert not found: $CERT"; exit 1; }
[ -f "$KEY" ]  || { echo "key not found: $KEY"; exit 1; }

mkroute() {  # name service port host
  local name="$1" svc="$2" port="$3" host="$4"
  local args=(--service="$svc" --port="$port" --hostname="$host" --cert="$CERT" --key="$KEY")
  [ -n "$CA" ] && args+=(--ca-cert="$CA")
  oc -n "$NS" delete route "$name" --ignore-not-found >/dev/null 2>&1 || true
  oc -n "$NS" create route edge "$name" "${args[@]}" >/dev/null
  oc -n "$NS" patch route "$name" --type merge \
    -p '{"spec":{"tls":{"insecureEdgeTerminationPolicy":"Redirect"}}}' >/dev/null
  echo "   $host -> $svc:$port"
}

echo ">> Creating custom-host Routes (BYO cert) in $NS"
mkroute gotak-demoland  gotak-gateway 8080 "gotak.$DOMAIN"
mkroute vault-demoland  vault         8200 "vault.$DOMAIN"
mkroute consul-demoland consul-ui     http "consul.$DOMAIN"

# Vault/Consul are admin surfaces: restrict to ALLOWLIST_IP (default dan's home
# IP) until Boundary brokers access. The gotak app route stays public.
ALLOWLIST_IP="${ALLOWLIST_IP:-143.105.191.161/32 3.148.232.34/32}"
echo ">> Restricting vault/consul routes to $ALLOWLIST_IP"
for r in vault-demoland consul-demoland; do
  oc -n "$NS" annotate route "$r" "haproxy.router.openshift.io/ip_whitelist=$ALLOWLIST_IP" --overwrite >/dev/null
done

echo ">> Repointing the web UI at the demoland.io endpoints"
oc -n "$NS" patch configmap gotak-web-config --type merge -p "{\"data\":{
  \"GOTAK_API_URL\":\"https://gotak.$DOMAIN\",
  \"GOTAK_WS_URL\":\"wss://gotak.$DOMAIN/ws\",
  \"VAULT_ADDR\":\"https://vault.$DOMAIN\"}}" >/dev/null
oc -n "$NS" rollout restart deploy/gotak-web >/dev/null 2>&1 || true

# Vault is called from the browser, so its CORS allow-list must include the new
# app origin (https://gotak.<domain>). Needs the Vault root token.
INIT_FILE="$DIR/.vault-init.json"
if [ -f "$INIT_FILE" ]; then
  echo ">> Updating Vault CORS allow-origin -> https://gotak.$DOMAIN"
  ROOT_TOKEN="$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["root_token"])' "$INIT_FILE")"
  oc -n "$NS" exec vault-0 -- sh -c \
    "VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=$ROOT_TOKEN vault write sys/config/cors \
       enabled=true allowed_origins='https://gotak.$DOMAIN' \
       allowed_headers='X-Vault-Token,Content-Type,X-Vault-Namespace'" >/dev/null
else
  echo ">> (no $INIT_FILE — set Vault CORS manually to https://gotak.$DOMAIN)"
fi

# DNS records for Namecheap.
CANON="$(oc -n "$NS" get route gotak-demoland -o jsonpath='{.status.ingress[0].routerCanonicalHostname}' 2>/dev/null || true)"
NODE_IP="${BOUNDARY_NODE_IP:-<boundary-node-public-ip>}"
cat <<EOF

>> DNS records to add in Namecheap (Advanced DNS) for $DOMAIN:

   Type    Host       Value
   ----    ----       -----
   CNAME   gotak      ${CANON:-<router canonical host: oc get route gotak-demoland -o jsonpath='{.status.ingress[0].routerCanonicalHostname}'>}
   CNAME   vault      ${CANON:-<same router canonical host>}
   CNAME   consul     ${CANON:-<same router canonical host>}
   A       boundary   $NODE_IP

   (Namecheap can't CNAME the apex, so these are all subdomains. If CNAME to the
   router host misbehaves, use A records to the router IPs:
     $(oc -n "$NS" get route gotak-demoland -o jsonpath='{.status.ingress[0].host}' >/dev/null 2>&1 && dig +short "gotak.$DOMAIN" 2>/dev/null | head -2 | sed 's/^/       /' || true)
   resolve the router with: host console-openshift-console.apps.gotak.<base-domain>)
EOF
