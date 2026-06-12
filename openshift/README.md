# GoTAK on OpenShift

A simple OpenShift deployment for GoTAK, intended for demos on **OpenShift Local (CRC)**
or any OpenShift 4.x cluster. This is a **new, self-contained deployment path** — it does
not touch the existing `nomad/`, `hashistack/`, or `web/k8s/` setups.

## What's here (Phase 1 — base app)

| File | What it deploys |
|------|-----------------|
| `00-postgres.yaml` | PostgreSQL + PostGIS (Secret, PVC, Deployment, Service, SA) |
| `10-gotak-server.yaml` | GoTAK server (ConfigMap-mounted `server.yaml`, Deployment, Service) |
| `20-gotak-web.yaml` | GoTAK web UI (ConfigMap, Deployment, Service) |
| `30-routes.yaml` | OpenShift Routes for web, API, and WebSocket/TAK |
| `deploy.sh` | One-command apply against the current `oc` context |

## Quick start

```bash
# 1. Log in to your cluster (CRC: `crc start` then copy the oc login line)
oc login ...

# 2. Build + push the images, then update the `image:` fields in the manifests
#    (default tags are gotak-server:latest / gotak-web:latest)
make docker            # or: docker build -f Dockerfile -t <registry>/gotak-server:latest .

# 3. Deploy
./openshift/deploy.sh gotak
```

The script prints the Route hosts when it finishes.

## OpenShift-specific notes (read before deploying)

- **Restricted SCC / non-root.** OpenShift runs containers as a random non-root UID by
  default. The manifests don't pin `runAsUser`, so OpenShift assigns one. The **web image's
  nginx must listen on `:8080`, not `:80`** — if it still binds `:80`, update the web
  `Dockerfile`/`nginx.conf`. The Go server already binds 8080/8087/8089 and is fine.
- **Postgres needs `anyuid`.** The `postgis/postgis` image runs as a fixed UID, so
  `deploy.sh` grants the `anyuid` SCC to the `gotak-postgres` service account. This needs
  cluster-admin (you have it on CRC). Alternative: swap in `quay.io/sclorg/postgresql-15-c9s`
  (arbitrary-UID friendly) and drop the SCC grant — but it has no PostGIS, so check whether
  the migrations require it first.
- **Demo secrets.** `POSTGRES_PASSWORD` and `jwt_secret` are demo values and are duplicated
  between `00-postgres.yaml` and the server ConfigMap. Rotate both together; do not ship as-is.
- **WebSocket.** The `gotak-ws` Route relies on OpenShift's native WebSocket upgrade on
  edge-terminated routes.

## Phase 2 — the agentic runtime-security demo layer (planned)

These add the "show all of it together" story on top of the base app. Not yet in this dir.

1. **Keycloak** — IdP for OAuth 2.0 / OIDC. Agents register and receive scoped tokens;
   GoTAK validates the OIDC token on the dynamic-query path.
2. **Vault** — OIDC auth method federated to Keycloak; issues scoped dynamic credentials.
   (GoTAK already uses Vault transit for comms encryption via the `hashistack/` setup.)
3. **Consul** — service discovery + Connect mesh between GoTAK components.
4. **Boundary** — identity-based access brokering to the target the agent queries; OIDC
   auth to the same Keycloak realm.
5. **IBM Bob → Terraform** — Bob generates the Terraform that provisions this whole stack
   on OpenShift, closing the "automation builds the secure runtime" loop.

Demo narrative: agent registers (Keycloak) → gets scoped creds (Vault) → is discovered/meshed
(Consul) → reaches its target via identity brokering (Boundary) → runs a dynamic query
authorized by its OAuth2 token. Every hop identity-first = zero trust on a running system.

See the vault note `Ideas/GoTAK Agentic Runtime Security Demo` for the full plan.
