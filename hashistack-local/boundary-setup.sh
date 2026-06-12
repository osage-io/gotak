#!/usr/bin/env bash
# Configure the local `boundary dev` instance for the GoTAK demo:
#   - org scope     "demoland"
#   - project scope "demoland-infra"
#   - password auth method with account/user "dfedick" (password: password)
#   - a role granting dfedick connect/list/read on targets
#   - TCP targets for Consul (8500), Nomad (4646), Vault (8200) at 127.0.0.1
#
# `boundary dev` keeps everything in memory, so this must be re-run after every
# Boundary restart. Idempotent-ish: re-running creates duplicate named resources
# (Boundary allows duplicate names), so prefer running against a fresh dev server.
#
# Usage:  ./hashistack/boundary-setup.sh
set -euo pipefail

export BOUNDARY_ADDR="${BOUNDARY_ADDR:-http://127.0.0.1:9200}"

# dev-mode fixed admin auth method + creds (see hashistack/logs/boundary.log)
ADMIN_AM="${ADMIN_AM:-ampw_1234567890}"
ADMIN_LOGIN="${ADMIN_LOGIN:-admin}"
export ADMIN_PW="${ADMIN_PW:-password}"

# the login we provision for accessing the targets
ORG_NAME="demoland"
PROJ_NAME="demoland"
export USER_PW="${USER_PW:-password}"
USER_LOGIN="${USER_LOGIN:-dfedick}"

need() { command -v "$1" >/dev/null || { echo "missing dependency: $1" >&2; exit 1; }; }
need boundary; need jq

echo "[boundary-setup] authenticating as ${ADMIN_LOGIN}…"
BOUNDARY_TOKEN="$(boundary authenticate password \
  -auth-method-id "${ADMIN_AM}" -login-name "${ADMIN_LOGIN}" -password env://ADMIN_PW \
  -format json 2>/dev/null | jq -r '.item.attributes.token // .item.token')"
[ -n "${BOUNDARY_TOKEN}" ] || { echo "auth failed" >&2; exit 1; }
export BOUNDARY_TOKEN

# stderr is dropped on capture commands to avoid the BOUNDARY_TOKEN deprecation
# notice polluting JSON parsed by jq.
ORG_ID="$(boundary scopes create -scope-id global -name "${ORG_NAME}" \
  -description 'Demoland org' -format json 2>/dev/null | jq -r '.item.id')"
echo "[boundary-setup] org    ${ORG_NAME} = ${ORG_ID}"

PROJ_ID="$(boundary scopes create -scope-id "${ORG_ID}" -name "${PROJ_NAME}" \
  -description 'HashiStack targets' -format json 2>/dev/null | jq -r '.item.id')"
echo "[boundary-setup] proj   ${PROJ_NAME} = ${PROJ_ID}"

AM_ID="$(boundary auth-methods create password -scope-id "${ORG_ID}" \
  -name "${ORG_NAME}-password" -format json 2>/dev/null | jq -r '.item.id')"
boundary scopes update -id "${ORG_ID}" -primary-auth-method-id "${AM_ID}" \
  -format json 2>/dev/null >/dev/null
echo "[boundary-setup] auth   ${AM_ID} (primary)"

ACCT_ID="$(boundary accounts create password -auth-method-id "${AM_ID}" \
  -login-name "${USER_LOGIN}" -password env://USER_PW -name "${USER_LOGIN}" \
  -format json 2>/dev/null | jq -r '.item.id')"
USER_ID="$(boundary users create -scope-id "${ORG_ID}" -name "${USER_LOGIN}" \
  -format json 2>/dev/null | jq -r '.item.id')"
boundary users add-accounts -id "${USER_ID}" -account "${ACCT_ID}" \
  -format json 2>/dev/null >/dev/null
echo "[boundary-setup] user   ${USER_LOGIN} = ${USER_ID} (account ${ACCT_ID})"

ROLE_ID="$(boundary roles create -scope-id "${ORG_ID}" -name "${ORG_NAME}-connect" \
  -description "${USER_LOGIN} connect to HashiStack targets" \
  -format json 2>/dev/null | jq -r '.item.id')"
boundary roles add-grant-scopes -id "${ROLE_ID}" \
  -grant-scope-id this -grant-scope-id children -format json 2>/dev/null >/dev/null
boundary roles add-grants -id "${ROLE_ID}" \
  -grant "ids=*;type=target;actions=authorize-session,read,list" \
  -grant "ids=*;type=scope;actions=list,read" \
  -grant "ids=*;type=session;actions=read,list,cancel" \
  -format json 2>/dev/null >/dev/null
boundary roles add-principals -id "${ROLE_ID}" -principal "${USER_ID}" \
  -format json 2>/dev/null >/dev/null
echo "[boundary-setup] role   ${ROLE_ID}"

# TCP targets with a direct network address (no host catalog needed).
# Returns the new target id on stdout.
create_target() {
  local name="$1" port="$2" id
  id="$(boundary targets create tcp -scope-id "${PROJ_ID}" -name "${name}" \
    -default-port "${port}" -address 127.0.0.1 -format json 2>/dev/null | jq -r '.item.id')"
  echo "[boundary-setup] target ${name} = ${id} (127.0.0.1:${port})" >&2
  echo "${id}"
}
create_target consul 8500 >/dev/null
create_target nomad  4646 >/dev/null
VAULT_TGT="$(create_target vault 8200)"
GOTAK_TGT="$(create_target gotak 8080)"

# Broker the Vault token as a credential on the vault target, so connecting to it
# surfaces username/password (= the Vault root token) to copy into the
# "Configure HashiCorp Vault" modal.
export VAULT_DEMO_PW="${VAULT_DEMO_PW:-root}"
CS_ID="$(boundary credential-stores create static -scope-id "${PROJ_ID}" \
  -name demoland-static -description 'Static creds for demoland targets' \
  -format json 2>/dev/null | jq -r '.item.id')"
CRED_ID="$(boundary credentials create username-password -credential-store-id "${CS_ID}" \
  -name vault-root-token -username root -password env://VAULT_DEMO_PW \
  -format json 2>/dev/null | jq -r '.item.id')"
boundary targets add-credential-sources -id "${VAULT_TGT}" \
  -brokered-credential-source "${CRED_ID}" -format json 2>/dev/null >/dev/null
echo "[boundary-setup] brokered Vault credential ${CRED_ID} on target ${VAULT_TGT}"

# Broker the GoTAK login (dfedick/password) on the gotak target so connecting to
# it surfaces the username/password to use on the GoTAK web login.
export GOTAK_DEMO_PW="${GOTAK_DEMO_PW:-password}"
GOTAK_CRED_ID="$(boundary credentials create username-password -credential-store-id "${CS_ID}" \
  -name gotak-login -username "${USER_LOGIN}" -password env://GOTAK_DEMO_PW \
  -format json 2>/dev/null | jq -r '.item.id')"
boundary targets add-credential-sources -id "${GOTAK_TGT}" \
  -brokered-credential-source "${GOTAK_CRED_ID}" -format json 2>/dev/null >/dev/null
echo "[boundary-setup] brokered GoTAK credential ${GOTAK_CRED_ID} on target ${GOTAK_TGT}"

cat <<EOF

[boundary-setup] Done.

  Org / Project : ${ORG_NAME} (${ORG_ID}) / ${PROJ_NAME} (${PROJ_ID})
  Login         : ${USER_LOGIN} / ${USER_PW}   (auth-method ${AM_ID})

  Authenticate, then connect to a target (org and project share the name
  "${ORG_NAME}", so connect by target ID or by scope ID to avoid ambiguity):
    export BOUNDARY_ADDR=${BOUNDARY_ADDR}
    boundary authenticate password -auth-method-id ${AM_ID} -login-name ${USER_LOGIN}
    boundary targets list -scope-id ${PROJ_ID}                 # grab the target IDs
    boundary connect -target-id <ttcp_…>                       # localhost proxy -> 127.0.0.1:<port>
EOF
