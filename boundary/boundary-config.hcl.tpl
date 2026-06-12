# Boundary combined controller+worker for the gotak demo, run on the EC2 node.
# install-boundary.sh renders the __PLACEHOLDERS__ and writes /etc/boundary/boundary.hcl.
#
# aead KMS (self-contained keys) is used instead of AWS KMS so the bastion has no
# dependency on rotating sandbox creds. For production, swap these for `kms "awskms"`.
disable_mlock = true

controller {
  name        = "gotak-controller"
  description = "gotak Boundary controller"
  database {
    url = "postgresql://boundary:__PG_PASS__@127.0.0.1:5432/boundary?sslmode=disable"
  }
}

worker {
  name        = "gotak-worker"
  description = "gotak ingress worker"
  # The address your Boundary client connects to for session proxying — the
  # node's public IP. Must be reachable from your laptop on tcp/9202.
  public_addr       = "__PUBLIC_ADDR__"
  initial_upstreams = ["127.0.0.1:9201"]
}

listener "tcp" {
  address     = "0.0.0.0:9200"
  purpose     = "api"
  tls_disable = true   # demo: API over HTTP. Terminate TLS in front for real use.
}

listener "tcp" {
  address = "0.0.0.0:9201"
  purpose = "cluster"
}

listener "tcp" {
  address     = "0.0.0.0:9202"
  purpose     = "proxy"
  tls_disable = true
}

kms "aead" {
  purpose   = "root"
  aead_type = "aes-gcm"
  key       = "__ROOT_KEY__"
  key_id    = "global_root"
}

kms "aead" {
  purpose   = "worker-auth"
  aead_type = "aes-gcm"
  key       = "__WORKER_AUTH_KEY__"
  key_id    = "global_worker-auth"
}

kms "aead" {
  purpose   = "recovery"
  aead_type = "aes-gcm"
  key       = "__RECOVERY_KEY__"
  key_id    = "global_recovery"
}
