#!/usr/bin/env bash
# Show a quick status of the local hashistack.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_DIR="${SCRIPT_DIR}/run"

check() {
  # $1 = name, $2 = pidfile, $3 = url
  local name="$1" pidfile="$2" url="$3"
  local state="\033[0;31mDOWN\033[0m"
  local pid="-"

  if [[ -f "${pidfile}" ]]; then
    pid="$(cat "${pidfile}" 2>/dev/null || echo '-')"
    if [[ -n "${pid}" && "${pid}" != "-" ]] && kill -0 "${pid}" 2>/dev/null; then
      if curl -sf -o /dev/null --max-time 2 "${url}"; then
        state="\033[0;32mUP  \033[0m"
      else
        state="\033[1;33mSTART\033[0m"
      fi
    fi
  fi
  printf "  %-7s  %b  pid=%-6s  %s\n" "${name}" "${state}" "${pid}" "${url}"
}

echo "HashiStack status:"
check "Consul" "${RUN_DIR}/consul.pid" "http://127.0.0.1:8500/v1/status/leader"
check "Vault"  "${RUN_DIR}/vault.pid"  "http://127.0.0.1:8200/v1/sys/health?standbyok=true&sealedcode=200&uninitcode=200"
check "Nomad"  "${RUN_DIR}/nomad.pid"  "http://127.0.0.1:4646/v1/status/leader"
