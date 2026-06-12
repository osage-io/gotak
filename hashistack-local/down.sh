#!/usr/bin/env bash
# Stop the single-node Consul/Vault/Nomad dev services started by up.sh.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_DIR="${SCRIPT_DIR}/run"
LOG_DIR="${SCRIPT_DIR}/logs"
DATA_DIR="${SCRIPT_DIR}/data"

log()  { echo -e "\033[0;34m[hashi-down]\033[0m $*"; }
ok()   { echo -e "\033[0;32m[ ok ]\033[0m $*"; }

stop_one() {
  # $1 = name, $2 = pidfile, $3 = pgrep pattern (fallback)
  local name="$1" pidfile="$2" pattern="$3"

  if [[ -f "${pidfile}" ]]; then
    local pid
    pid="$(cat "${pidfile}")"
    if [[ -n "${pid}" ]] && kill -0 "${pid}" 2>/dev/null; then
      log "Stopping ${name} (pid ${pid})…"
      kill "${pid}" 2>/dev/null || true
      # Give it a moment, then SIGKILL if still alive.
      for _ in 1 2 3 4 5; do
        kill -0 "${pid}" 2>/dev/null || break
        sleep 1
      done
      kill -9 "${pid}" 2>/dev/null || true
      ok "${name} stopped"
    else
      log "${name} pidfile present but process not running"
    fi
    rm -f "${pidfile}"
  fi

  # Fallback: kill any leftover instance not tracked by our pidfile.
  local stragglers
  stragglers="$(pgrep -f "${pattern}" || true)"
  if [[ -n "${stragglers}" ]]; then
    log "Killing untracked ${name} processes: ${stragglers}"
    # shellcheck disable=SC2086
    kill ${stragglers} 2>/dev/null || true
  fi
}

stop_one "Nomad"  "${RUN_DIR}/nomad.pid"  "nomad agent -dev"
stop_one "Vault"  "${RUN_DIR}/vault.pid"  "vault server -dev"
stop_one "Consul" "${RUN_DIR}/consul.pid" "consul agent -dev"

# Cleanup
if [[ "${HASHI_KEEP_DATA:-0}" != "1" ]]; then
  rm -rf "${DATA_DIR}"
  ok "Removed ephemeral data dir (${DATA_DIR}). Set HASHI_KEEP_DATA=1 to preserve."
fi

log "Logs preserved at ${LOG_DIR}"
ok "HashiStack is down."
