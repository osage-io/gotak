#!/usr/bin/env bash
# Bring up Consul, Vault, and Nomad as single-node dev services for GoTAK.
#
# Each service runs in `-dev` mode in the background. PIDs and logs are kept
# in this directory so `down.sh` can stop them cleanly.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="${SCRIPT_DIR}/logs"
RUN_DIR="${SCRIPT_DIR}/run"
DATA_DIR="${SCRIPT_DIR}/data"

mkdir -p "${LOG_DIR}" "${RUN_DIR}" "${DATA_DIR}"

# ---- helpers ----------------------------------------------------------------

log()  { echo -e "\033[0;34m[hashi-up]\033[0m $*"; }
ok()   { echo -e "\033[0;32m[ ok ]\033[0m $*"; }
warn() { echo -e "\033[1;33m[warn]\033[0m $*"; }
err()  { echo -e "\033[0;31m[err ]\033[0m $*" >&2; }

require_brew() {
  if ! command -v brew >/dev/null 2>&1; then
    err "Homebrew is required to auto-install missing HashiCorp tools."
    err "Install brew first: https://brew.sh"
    exit 1
  fi
}

ensure_tool() {
  local tool="$1"
  if command -v "${tool}" >/dev/null 2>&1; then
    return 0
  fi
  log "${tool} not found — installing via Homebrew…"
  require_brew
  if ! brew list hashicorp/tap/"${tool}" >/dev/null 2>&1; then
    brew tap hashicorp/tap >/dev/null 2>&1 || true
    brew install "hashicorp/tap/${tool}"
  fi
}

is_running() {
  # $1 = pidfile
  local pidfile="$1"
  [[ -f "${pidfile}" ]] || return 1
  local pid
  pid="$(cat "${pidfile}")"
  [[ -n "${pid}" ]] && kill -0 "${pid}" 2>/dev/null
}

wait_http() {
  # $1 = label, $2 = url, $3 = max attempts (default 30)
  local label="$1" url="$2" max="${3:-30}"
  local i=0
  until curl -sf -o /dev/null "${url}"; do
    i=$((i + 1))
    if [[ $i -ge $max ]]; then
      err "${label} did not become ready at ${url} after ${max}s"
      return 1
    fi
    sleep 1
  done
  ok "${label} is ready (${url})"
}

# ---- prerequisites ----------------------------------------------------------

log "Checking HashiCorp tooling…"
ensure_tool consul
ensure_tool vault
ensure_tool nomad

if ! command -v docker >/dev/null 2>&1; then
  warn "Docker not found on PATH. Nomad's docker driver will fail until Docker Desktop is running."
fi

# ---- consul -----------------------------------------------------------------

CONSUL_PID="${RUN_DIR}/consul.pid"
if is_running "${CONSUL_PID}"; then
  ok "Consul already running (pid $(cat "${CONSUL_PID}"))"
else
  log "Starting Consul (dev mode)…"
  rm -rf "${DATA_DIR}/consul"
  mkdir -p "${DATA_DIR}/consul"
  nohup consul agent -dev \
    -client 0.0.0.0 \
    -bind 127.0.0.1 \
    -ui \
    -node gotak-dev \
    -data-dir "${DATA_DIR}/consul" \
    >"${LOG_DIR}/consul.log" 2>&1 &
  echo $! >"${CONSUL_PID}"
  wait_http "Consul" "http://127.0.0.1:8500/v1/status/leader" 30
fi

# ---- vault ------------------------------------------------------------------

VAULT_PID="${RUN_DIR}/vault.pid"
if is_running "${VAULT_PID}"; then
  ok "Vault already running (pid $(cat "${VAULT_PID}"))"
else
  log "Starting Vault (dev mode, root token = 'root')…"
  nohup vault server -dev \
    -dev-root-token-id=root \
    -dev-listen-address="0.0.0.0:8200" \
    >"${LOG_DIR}/vault.log" 2>&1 &
  echo $! >"${VAULT_PID}"
  wait_http "Vault" "http://127.0.0.1:8200/v1/sys/health?standbyok=true&sealedcode=200&uninitcode=200" 30
fi

# ---- nomad ------------------------------------------------------------------

NOMAD_PID="${RUN_DIR}/nomad.pid"
if is_running "${NOMAD_PID}"; then
  ok "Nomad already running (pid $(cat "${NOMAD_PID}"))"
else
  log "Starting Nomad (dev mode + docker driver, integrating with Consul)…"
  mkdir -p "${DATA_DIR}/nomad"
  # Nomad dev mode runs server+client in one process. We point it at the
  # Consul agent we just started so jobs can register services.
  # Note: dev mode requires sudo for the docker driver on macOS in some setups,
  # but with Docker Desktop's socket it usually works rootless.
  nohup nomad agent -dev \
    -bind 127.0.0.1 \
    -node gotak-dev \
    -dc dc1 \
    -consul-address=127.0.0.1:8500 \
    -data-dir "${DATA_DIR}/nomad" \
    -config "${SCRIPT_DIR}/nomad-client.hcl" \
    >"${LOG_DIR}/nomad.log" 2>&1 &
  echo $! >"${NOMAD_PID}"
  wait_http "Nomad" "http://127.0.0.1:4646/v1/status/leader" 60
fi

# ---- summary ----------------------------------------------------------------

cat <<EOF

🎉 HashiStack is up.

  Consul UI : http://127.0.0.1:8500/ui
  Vault UI  : http://127.0.0.1:8200/ui   (token: root)
  Nomad UI  : http://127.0.0.1:4646/ui

Environment (eval to load into current shell):

  export CONSUL_HTTP_ADDR=http://127.0.0.1:8500
  export VAULT_ADDR=http://127.0.0.1:8200
  export VAULT_TOKEN=root
  export NOMAD_ADDR=http://127.0.0.1:4646

Logs : ${LOG_DIR}
Stop : make hashi-down   (or ./hashistack/down.sh)
EOF
