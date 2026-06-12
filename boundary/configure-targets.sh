#!/usr/bin/env bash
# Create the gotak org/project, a login, and the privileged-access targets in
# Boundary. Run ON THE NODE after install-boundary.sh (it uses the recovery KMS
# from /etc/boundary/boundary.hcl, so no prior login is needed).
#
#   sudo PG_HOST=<sno-node-ip> PG_PORT=30432 \
#        VAULT_HOST=vault-gotak.apps.gotak.<domain> \
#        CONSUL_HOST=consul-ui-gotak.apps.gotak.<domain> \
#        bash configure-targets.sh
#
# Targets created (all TCP, brokered through the node worker):
#   openshift-api  <API_HOST>:6443      kube-apiserver
#   vault-ui       <VAULT_HOST>:443     Vault UI/API (via its Route)
#   consul-ui      <CONSUL_HOST>:443    Consul UI/API (via its Route)
#   node-ssh       127.0.0.1:22         SSH to this EC2 host
#   postgres       <PG_HOST>:<PG_PORT>  in-cluster Postgres (via NodePort) *see README
set -euo pipefail

export BOUNDARY_ADDR="${BOUNDARY_ADDR:-http://127.0.0.1:9200}"
CFG=/etc/boundary/boundary.hcl
REC=(-recovery-config "$CFG")
j() { python3 -c 'import json,sys;print(json.load(sys.stdin)["item"]["id"])'; }

API_HOST="${API_HOST:-api.gotak.daniel-fedick.aws.sbx.hashicorpdemo.com}"
VAULT_HOST="${VAULT_HOST:?set VAULT_HOST to the vault Route host}"
CONSUL_HOST="${CONSUL_HOST:?set CONSUL_HOST to the consul-ui Route host}"
PG_HOST="${PG_HOST:?set PG_HOST to the SNO node IP exposing the postgres NodePort}"
PG_PORT="${PG_PORT:-30432}"
LOGIN_NAME="${LOGIN_NAME:-gotak}"
LOGIN_PASS="${LOGIN_PASS:-$(openssl rand -hex 12)}"

echo ">> Creating org + project scopes"
ORG=$(boundary scopes create "${REC[@]}" -scope-id global -name gotak-org \
        -description "gotak" -format json | j)
PROJ=$(boundary scopes create "${REC[@]}" -scope-id "$ORG" -name gotak-project \
        -description "gotak targets" -format json | j)
echo "   org=$ORG project=$PROJ"

echo ">> Creating password login ($LOGIN_NAME)"
AM=$(boundary auth-methods create password "${REC[@]}" -scope-id "$ORG" \
       -name gotak-pw -format json | j)
ACCT=$(boundary accounts create password "${REC[@]}" -auth-method-id "$AM" \
       -login-name "$LOGIN_NAME" -password "env://LOGIN_PASS" -format json | j)
USR=$(boundary users create "${REC[@]}" -scope-id "$ORG" -name "$LOGIN_NAME" -format json | j)
boundary users add-accounts "${REC[@]}" -id "$USR" -account "$ACCT" >/dev/null

echo ">> Granting the user access to targets in the project"
ROLE=$(boundary roles create "${REC[@]}" -scope-id "$PROJ" -name gotak-access \
        -grant-scope-id "$PROJ" -format json | j)
boundary roles add-principals "${REC[@]}" -id "$ROLE" -principal "$USR" >/dev/null
boundary roles add-grants "${REC[@]}" -id "$ROLE" \
  -grant 'ids=*;type=target;actions=list,read,authorize-session' \
  -grant 'ids=*;type=session;actions=list,read,read:self,cancel:self' >/dev/null

mk() {  # name host port
  echo "   target $1 -> $2:$3"
  boundary targets create tcp "${REC[@]}" -scope-id "$PROJ" \
    -name "$1" -default-port "$3" -address "$2" \
    -session-connection-limit -1 -format json >/dev/null
}
echo ">> Creating targets"
mk openshift-api "$API_HOST"   6443
mk vault-ui      "$VAULT_HOST" 443
mk consul-ui     "$CONSUL_HOST" 443
mk node-ssh      127.0.0.1     22
mk postgres      "$PG_HOST"    "$PG_PORT"

cat <<EOF

>> Done. Log in with the Boundary desktop client / CLI:
     Address:  http://$(curl -fsS http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null || echo '<node-public-ip>'):9200
     Auth:     password    login-name=$LOGIN_NAME    password=$LOGIN_PASS
   Then 'Connect' to a target (e.g. postgres -> psql on 127.0.0.1:<local-port>).
   (Save that password — it is not stored anywhere else.)
EOF
