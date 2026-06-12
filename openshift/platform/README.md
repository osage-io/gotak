# HashiStack platform on Single-Node OpenShift

Deploys **Vault** then **Consul** (Helm) into the **`gotak`** namespace — the same
namespace as the app — with Vault **auto-unsealed by AWS KMS**. Then `../deploy.sh`
brings up the app and points its UI at in-cluster Vault.

```
gotak namespace
├── vault   (Helm: hashicorp/vault)   raft storage, KMS auto-unseal, Route+CORS
├── consul  (Helm: hashicorp/consul)  single server
└── gotak   postgres + server + web   (web → Vault Route for transit/PKI/KV)
```

Everything runs on the single arm64 (Graviton) node — Vault and Consul images are
multi-arch, so no arch overrides are needed.

## Prerequisites

- `oc` logged in to the SNO cluster as cluster-admin
  (`export KUBECONFIG=../../iac/sno/cluster-auth/kubeconfig`)
- `helm` 3 and the `aws` CLI
- Live AWS creds exported in your shell (the same sandbox STS creds):
  ```bash
  export AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... AWS_SESSION_TOKEN=...
  ```

## Install

```bash
cd openshift/platform
./install-platform.sh        # KMS key → creds Secret → Vault → init/config → Consul
cd ..
./deploy.sh                  # postgres + gotak-server + gotak-web, wired to Vault
```

`install-platform.sh` is idempotent (reuses the KMS key via its alias, skips Vault
re-init). On first run it writes:

| File | Contents |
|------|----------|
| `.vault-init.json` | Vault **root token** + recovery key (chmod 600, gitignored) |
| `.vault-gotak-token` | scoped token for the app (transit/pki/kv) |

## How gotak reaches Vault

The gotak UI calls Vault **from the browser** (transit encryption, PKI device
certs, the Anthropic API key in KV). So:

- Vault is exposed via an OpenShift **Route** (`https://vault-gotak.apps.<cluster>`).
- **CORS** is enabled (`sys/config/cors`) scoped to the gotak web Route origin.
- The web image reads its Vault address from `window.GOTAK_CONFIG.vaultUrl`, fed by
  the `VAULT_ADDR` env var. `deploy.sh` sets that ConfigMap value from the live
  Vault Route, so the UI points at in-cluster Vault out of the box. You can still
  override it in the in-app Vault config modal.

Paste the **gotak token** (`.vault-gotak-token`) into the UI's Vault config when it
asks for a token.

## Service mesh (Consul Connect)

Every app pod in the `gotak` namespace joins the mesh automatically
(`connectInject.default: true`) — an Envoy sidecar is injected unless the pod is
annotated `consul.hashicorp.com/connect-inject: "false"`. On OpenShift the traffic
redirection is done by the **Consul CNI plugin via Multus**, so sidecars run under
the restricted SCC (no privileged init container).

**Who's in the mesh:** `gotak-web`, `gotak-server`, `postgres`. **Out:** Vault
(reached by the browser via its Route + CORS) and Consul's own pods.

**Intentions** ([`consul-intentions.yaml`](consul-intentions.yaml)) are **default-deny** —
nothing talks to anything without an explicit allow:

| Source | Destination | Why |
|--------|-------------|-----|
| `*` | `*` | **deny** (catch-all) |
| `gotak-gateway` | `gotak-web` | gateway serves the SPA |
| `gotak-gateway` | `gotak-server` | gateway routes `/api` + `/ws` |
| `gotak-server` | `postgres` | the DB hop — fully mesh-enforced |

`gotak-server → postgres` is the clean showcase: postgres has no Route, so all its
inbound is meshed. Delete that intention and the server's DB connection breaks;
re-add it and it recovers.

```bash
# watch enforcement
oc apply -f consul-intentions.yaml          # default-deny + allows
oc delete serviceintentions postgres -n gotak   # server -> postgres now blocked
```

### North-south: Consul API Gateway

All ingress goes through a **Consul API Gateway**
([`consul-api-gateway.yaml`](consul-api-gateway.yaml)) — the single front door to
gotak. One external hostname, path-routed, then mTLS into the mesh:

```
browser ─▶ OpenShift edge Route ─▶ gotak-gateway (HTTP listener)
                                      │  mTLS + intentions
                                      ├─ /api ─▶ gotak-server:8080
                                      ├─ /ws  ─▶ gotak-server:8087
                                      └─ /    ─▶ gotak-web:8080
```

Because the gateway is itself a mesh member, `gotak-web` and `gotak-server` keep
**no Routes of their own and no inbound exclusions** — they're reachable only via
`gateway → web` / `gateway → server` intentions, fully mesh-enforced. `deploy.sh`
points the web UI's `GOTAK_API_URL`/`GOTAK_WS_URL` at the gateway Route host.

## ⚠️ KMS auto-unseal and rotating creds

Vault's KMS seal reads AWS creds **at pod start**. The sandbox STS creds expire
(~12 h). If `vault-0` restarts after the creds in the `vault-aws-creds` Secret have
expired, it **cannot auto-unseal**. When you refresh your AWS session:

```bash
export AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... AWS_SESSION_TOKEN=...
./refresh-vault-aws-creds.sh    # updates the Secret + bounces vault-0
```

(For a non-sandbox account you'd use an IAM role via Pod Identity / IRSA instead of
a creds Secret, and this caveat goes away.)

## Teardown

```bash
helm uninstall consul vault -n gotak
oc delete pvc -l app.kubernetes.io/instance=vault -n gotak
aws kms schedule-key-deletion --key-id alias/gotak-vault-unseal --pending-window-in-days 7
```
