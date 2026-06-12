# GoTAK Deployment Architecture — AWS / OpenShift / HashiCorp Stack

The cloud demo environment: GoTAK on **Single-Node OpenShift (SNO)** in AWS,
integrated with **Vault** (transit encryption, PKI device certs, KV secrets),
**Consul** (service mesh + API Gateway ingress), and **Boundary** (privileged
access), with every layer captured as code.

> App-level system design lives in [ARCHITECTURE.md](ARCHITECTURE.md). The
> docker-compose path is [../DEPLOYMENT.md](../DEPLOYMENT.md). This document is
> the cloud/platform layer.

```
                          demoland.io (Namecheap DNS, Sectigo *.demoland.io cert)
   gotak. vault. consul. argo. ──▶ OpenShift router        sno. ──▶ kube-apiserver:6443
   boundary. ──────────────────▶ EC2 node (Boundary PAM)

  ┌─ AWS us-east-2 ───────────────────────────────────────────────────────────┐
  │  VPC 10.0.0.0/16 (gotak-network workspace)                                │
  │  ├── EC2 t4g.medium  installer/bastion node (gotak-installer workspace)   │
  │  │     └── Boundary controller+worker (podman, Ansible-managed)           │
  │  └── EC2 m6g.2xlarge SNO cluster node (created by openshift-install)      │
  │        OpenShift 4.22 (arm64) — namespace `gotak`:                        │
  │          Consul mesh (default-deny intentions)                            │
  │          Consul API Gateway ──▶ /api,/ws → gotak-server; / → gotak-web    │
  │          Vault (raft, AWS KMS auto-unseal) ◀── browser via Route+CORS     │
  │          postgres (PostGIS, arm64 image)                                  │
  │        openshift-gitops: Argo CD app-of-apps deploys all of the above     │
  └───────────────────────────────────────────────────────────────────────────┘
```

## Layers and where they live

| Layer | Tool | Source | Notes |
|---|---|---|---|
| VPC/network | Terraform (TFC `gotak-network`) | `iac/network` | VCS-driven, us-east-2 |
| Installer node | Terraform (TFC `gotak-installer`) | `iac/installer` | arm64 jump box; runs the install; hosts Boundary |
| Cluster install | Ansible → `openshift-install` | `iac/ansible-sno` | SNO 4.22 arm64; CCO **Manual** mode (see below) |
| ghcr pull secret | Terraform (TFC `gotak-registry-secret`) | `iac/gotak-registry-secret` | sensitive vars; connects via `sno.demoland.io` |
| Platform + app | Argo CD (OpenShift GitOps) | `gitops/` + `openshift/` | app-of-apps, targetRevision `main` |
| Boundary | Ansible | `iac/ansible-boundary` | controller+worker on the node, podman |
| Images | GitHub Actions | `.github/workflows/container-release.yml` | `ghcr.io/osage-io/gotak-{server,web}`, **multi-arch (arm64!)** |
| (parked) ROSA | Terraform | `iac/compute` | unmerged by design — sandbox SCP denies ROSA prerequisites |
| (parallel) Bob 3-node stack | Terraform/Ansible | `iac/terraform`, `iac/ansible` | IBM Bob-generated track |

## GitOps model

`gitops/root-app.yaml` (apply once by hand) → child Applications in sync waves:

1. **consul** (Helm 2.0.0) + **vault** (Helm 0.33.0) — charts from
   `helm.releases.hashicorp.com`, values from this repo (`openshift/platform/*-values.yaml`)
2. **api-gateway**, **intentions** — need Consul's CRDs
3. **gotak** — postgres + server + web (`openshift/*.yaml`)

**Deliberately imperative** (scripted, not GitOps): KMS key + `vault-aws-creds`
Secret (`openshift/platform/install-platform.sh`), Vault init/transit/PKI/CORS
(`vault-post-config.sh`), BYO-cert routes (`custom-domain-routes.sh`), API named
cert (`apiserver-named-cert.sh`), Argo console domain (`gitops/argocd-custom-domain.sh`),
Boundary (Ansible).

## Mesh + ingress

Every pod in `gotak` is injected (Consul connect, transparent proxy via the CNI
plugin). Intentions are **default-deny**; allows: `gateway→web`, `gateway→server`,
`server→postgres`. The **Consul API Gateway is the only ingress** — one edge
Route, path-routed. Vault opts out of the mesh (the browser calls it directly
through its Route, CORS-scoped to the gotak origin).

## Sandbox constraints that shaped the design

- **ROSA blocked** (marketplace subscription + IAM user/OIDC creation denied) →
  pivot to SNO via `openshift-install`.
- **No IAM users/OIDC** → cloud-credential-operator in **Manual mode**: six
  operator secrets stamped from the node's instance-profile creds (IMDSv2) with a
  30-min cron refresh.
- **STS session creds (~12 h)** → Vault's KMS seal reads creds at pod start;
  `refresh-vault-aws-creds.sh` updates the Secret and bounces `vault-0`.
- **`kms:CreateKey` may be explicitly denied** by the session policy — the
  auto-unseal key (`alias/gotak-vault-unseal`) needs a less-restricted session.

## Operational pitfalls (learned the hard way)

1. **arm64 everywhere.** The node is Graviton. amd64-only images fail with
   `exec format error` (postgis → use `docker.io/imresamu/postgis`). CI builds
   must set `platforms: linux/arm64`.
2. **Fully qualify `docker.io/...`** — OpenShift resolves bare `hashicorp/*` to
   `registry.connect.redhat.com`, where those images don't exist.
3. **Webhook deadlock.** A mutating webhook with `failurePolicy: Fail` whose pod
   can't start blocks ALL pod creation — including openshift-apiserver — wedging
   the control plane. Consul's injector is pinned to `failurePolicy: Ignore`.
   Recovery: delete the consul webhookconfigurations.
4. **OpenShift owns the Gateway API CRDs** (ingress-operator) — Consul must set
   `global.installK8sNetworkingCRDs: false` or the Argo sync fails on ownership.
5. **Wedged Argo sync** retries the old revision forever:
   `oc patch application X -n openshift-gitops --type json -p='[{"op":"remove","path":"/operation"}]'`
   then hard-refresh.
6. **Argo CD customization goes in the ArgoCD CR** — the operator reverts direct
   Route edits.

## DNS / TLS

All names under `demoland.io`, covered by one Sectigo `*.demoland.io` wildcard:
`gotak`/`vault`/`consul`/`argo` → router IPs (A records), `sno` → API IP (named
serving cert bound via the APIServer config), `boundary` → node IP (no cert;
worker uses `tls_disable` for the demo). Records table:
`openshift/platform/DNS-demoland.md`.
