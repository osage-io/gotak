#!/usr/bin/env bash
set -euo pipefail

CONSUL_ADDR=${CONSUL_HTTP_ADDR:-http://localhost:8500}

services=(
  "gotak-postgres-dev:postgres:5432:db,postgres,gotak"
  "gotak-redis-dev:redis:6379:cache,redis,gotak"
  "gotak-nats-dev:nats:4222:messaging,nats,gotak"
  "gotak-jaeger-dev:jaeger-ui:16686:tracing,jaeger,monitoring,gotak"
  "gotak-adminer-dev:adminer:8080:db,admin,ui,gotak"
  "gotak-web-dev:gotak-web:80:web,ui,gotak,frontend"
  "gotak-server-dev:gotak-api:8080:api,tak,gotak,server"
  "gotak-server-dev:gotak-tak:8087:tak,protocol"
)

get_ip() {
  local cname=$1
  docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$cname" 2>/dev/null || true
}

register_service() {
  local id=$1 name=$2 address=$3 port=$4 tags=$5
  # Convert comma-separated tags to JSON array
  local json_tags=$(echo "$tags" | sed 's/,/","/g' | sed 's/^/"/' | sed 's/$/"/')
  
  cat <<JSON | curl -sfS -X PUT --data-binary @- "$CONSUL_ADDR/v1/agent/service/register"
{
  "ID": "${id}",
  "Name": "${name}",
  "Address": "${address}",
  "Port": ${port},
  "Tags": [${json_tags}],
  "Check": {
    "TCP": "${address}:${port}",
    "Interval": "15s",
    "Timeout": "3s"
  }
}
JSON
  echo "Registered ${name} (${id}) at ${address}:${port}"
}

for entry in "${services[@]}"; do
  IFS=":" read -r cname sname sport stags <<<"$entry"
  ip=$(get_ip "$cname")
  if [[ -z "$ip" ]]; then
    echo "Skipping $sname: container $cname not found or no IP yet" >&2
    continue
  fi
  sid="${sname}-${ip}-${sport}"
  register_service "$sid" "$sname" "$ip" "$sport" "$stags" || echo "Failed to register $sname" >&2
done

echo "Done. Visit Consul UI at $CONSUL_ADDR/ui"
