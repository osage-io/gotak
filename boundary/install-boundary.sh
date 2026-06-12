#!/usr/bin/env bash
# Install + start Boundary (controller + worker) on the EC2 node, with a Postgres
# in podman. Run THIS ON THE NODE:
#
#   scp -i ~/.ssh/dfed01 boundary/{install-boundary.sh,boundary-config.hcl.tpl} ec2-user@<node-ip>:
#   ssh -i ~/.ssh/dfed01 ec2-user@<node-ip> 'sudo bash install-boundary.sh'
#
# Idempotent-ish: re-running regenerates config + restarts. It will NOT wipe the
# Boundary DB once initialised (the generated admin creds in the init output stay
# valid), so capture that output the first time.
set -euo pipefail

BOUNDARY_VERSION="${BOUNDARY_VERSION:-0.18.0}"
ARCH="arm64"   # Graviton node
PG_PASS="${PG_PASS:-$(openssl rand -hex 16)}"
CFG_DIR=/etc/boundary
TPL="$(dirname "$0")/boundary-config.hcl.tpl"

echo ">> Installing podman + unzip"
dnf install -y podman unzip >/dev/null

echo ">> Installing Boundary $BOUNDARY_VERSION ($ARCH)"
cd /tmp
curl -fsSLo boundary.zip "https://releases.hashicorp.com/boundary/${BOUNDARY_VERSION}/boundary_${BOUNDARY_VERSION}_linux_${ARCH}.zip"
unzip -o boundary.zip boundary >/dev/null
install -m 0755 boundary /usr/local/bin/boundary
/usr/local/bin/boundary version

echo ">> Starting Postgres (podman) for Boundary state"
if ! podman container exists boundary-pg; then
  podman run -d --name boundary-pg --restart=always \
    -e POSTGRES_USER=boundary -e POSTGRES_PASSWORD="$PG_PASS" -e POSTGRES_DB=boundary \
    -p 127.0.0.1:5432:5432 docker.io/library/postgres:15 >/dev/null
  sleep 6
else
  echo "   boundary-pg already running (reusing; keeping its password)"
  # Reuse the stored password so the DSN matches the existing DB.
  PG_PASS="$(grep -oP '(?<=boundary:)[^@]+' "$CFG_DIR/boundary.hcl" 2>/dev/null || echo "$PG_PASS")"
fi

echo ">> Rendering $CFG_DIR/boundary.hcl"
mkdir -p "$CFG_DIR"
PUBLIC_ADDR="$(curl -fsS -H 'X-aws-ec2-metadata-token: '"$(curl -fsS -X PUT 'http://169.254.169.254/latest/api/token' -H 'X-aws-ec2-metadata-token-ttl-seconds: 60')" http://169.254.169.254/latest/meta-data/public-ipv4)"
gen_key() { openssl rand -base64 32; }
sed -e "s|__PG_PASS__|$PG_PASS|g" \
    -e "s|__PUBLIC_ADDR__|$PUBLIC_ADDR|g" \
    -e "s|__ROOT_KEY__|$(gen_key)|g" \
    -e "s|__WORKER_AUTH_KEY__|$(gen_key)|g" \
    -e "s|__RECOVERY_KEY__|$(gen_key)|g" \
    "$TPL" > "$CFG_DIR/boundary.hcl"
chmod 600 "$CFG_DIR/boundary.hcl"

echo ">> Initialising the Boundary database (first run only)"
if /usr/local/bin/boundary database init -config "$CFG_DIR/boundary.hcl" \
      -format table 2>&1 | tee "$CFG_DIR/init-output.txt"; then
  echo "   >>> SAVE the generated admin login above (also in $CFG_DIR/init-output.txt) <<<"
else
  echo "   (database already initialised — skipping)"
fi

echo ">> Installing systemd unit"
cat > /etc/systemd/system/boundary.service <<'UNIT'
[Unit]
Description=Boundary controller+worker
After=network-online.target
Wants=network-online.target
[Service]
ExecStart=/usr/local/bin/boundary server -config /etc/boundary/boundary.hcl
Restart=on-failure
LimitMEMLOCK=infinity
[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload
systemctl enable --now boundary

sleep 4
echo ">> Boundary status:"
systemctl --no-pager --lines=0 status boundary || true
echo ">> API:   http://$PUBLIC_ADDR:9200"
echo ">> Proxy: $PUBLIC_ADDR:9202 (open this in the node security group for your client IP)"
echo ">> Next:  run configure-targets.sh (uses the recovery KMS) to create targets."
