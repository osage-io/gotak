#!/usr/bin/env bash
# Deploy the standalone GoTAK stack (postgres + redis + gotak-server) to the
# local single-node Nomad started by hashistack/up.sh.
#
# The canonical job spec at nomad/deployments/standalone/gotak-complete.nomad.hcl
# is hardcoded to a remote node named "hashinuc01" (constraints + DB hosts).
# We render a temp copy with the local node name substituted in so it actually
# places on a developer laptop.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
JOB_SRC="${PROJECT_ROOT}/nomad/deployments/standalone/gotak-complete.nomad.hcl"
RENDER_DIR="${SCRIPT_DIR}/.rendered"

export NOMAD_ADDR="${NOMAD_ADDR:-http://127.0.0.1:4646}"

log()  { echo -e "\033[0;34m[nomad-deploy]\033[0m $*"; }
ok()   { echo -e "\033[0;32m[ ok ]\033[0m $*"; }
err()  { echo -e "\033[0;31m[err ]\033[0m $*" >&2; }

# ---- prerequisites ----------------------------------------------------------

if ! command -v nomad >/dev/null 2>&1; then
  err "nomad CLI not found. Run 'make hashi-up' first."
  exit 1
fi

if ! nomad node status >/dev/null 2>&1; then
  err "Cannot reach Nomad at ${NOMAD_ADDR}. Run 'make hashi-up' first."
  exit 1
fi

if [[ ! -f "${JOB_SRC}" ]]; then
  err "Job spec not found: ${JOB_SRC}"
  exit 1
fi

# ---- detect local node ------------------------------------------------------

# The dev-mode Nomad agent is started with -node gotak-dev (see hashistack/up.sh),
# but fall back to whatever the first ready client node is.
# Columns: ID  NodePool  DC  Name  Class  Drain  Eligibility  Status
# The Name column is the 4th field (older Nomad lacked the Node Pool column,
# which is why the previous $2 broke and selected the pool name "default").
LOCAL_NODE="$(nomad node status -short 2>/dev/null \
  | awk 'NR>1 && NF {print $4; exit}')"
LOCAL_NODE="${LOCAL_NODE:-gotak-dev}"
log "Targeting Nomad node: ${LOCAL_NODE}"

# ---- render job -------------------------------------------------------------

mkdir -p "${RENDER_DIR}"
JOB_OUT="${RENDER_DIR}/gotak-complete.nomad.hcl"

# Replace both:
#   - the constraint regex value "hashinuc01"  → local node name
#   - env var hostnames "hashinuc01"           → host.docker.internal
#
# NOTE: each Nomad task runs in its own Docker network namespace, so "127.0.0.1"
# inside the server container points at the server itself, NOT postgres/redis.
# Their static ports are published to the host, which Docker Desktop exposes to
# containers via the special DNS name host.docker.internal. Use that so the
# gotak-server can actually reach the databases.
sed \
  -e "s/value     = \"hashinuc01\"/value     = \"${LOCAL_NODE}\"/g" \
  -e "s/= \"hashinuc01\"/= \"host.docker.internal\"/g" \
  "${JOB_SRC}" >"${JOB_OUT}"

log "Rendered job → ${JOB_OUT}"

# ---- validate + deploy ------------------------------------------------------

log "Validating job spec…"
nomad job validate "${JOB_OUT}"

log "Submitting job…"
nomad job run "${JOB_OUT}"

ok "Submitted. Watch with:"
echo "    nomad job status gotak-complete"
echo "    nomad alloc logs -f -job gotak-complete"
echo
echo "Service endpoints (once healthy):"
echo "    GoTAK API/UI : http://127.0.0.1:8080"
echo "    CoT TCP      : 127.0.0.1:8087"
echo "    CoT TLS      : 127.0.0.1:8089"
echo "    PostgreSQL   : 127.0.0.1:5432"
echo "    Redis        : 127.0.0.1:6379"
