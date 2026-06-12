# gotak-registry-secret (Terraform)

Creates the **ghcr.io image pull secret** in the `gotak` namespace from a sensitive
variable, so private GitHub Container Registry images can be pulled. Runs in TF
Cloud; the PAT lives in TFC's Vault-encrypted, access-scoped state.

## What it makes

A `kubernetes.io/dockerconfigjson` Secret named `ghcr` (override with
`secret_name`). The Argo-managed deployments reference it via `imagePullSecrets`.

## Auth to the cluster

The kubernetes provider connects to **`https://sno.demoland.io:6443`** — the
API's named-cert host, whose Sectigo wildcard cert is publicly trusted, so TFC's
remote runner reaches it with no cluster CA. Auth is a **scoped SA token** that can
manage only secrets in `gotak`:

```bash
export KUBECONFIG=../sno/cluster-auth/kubeconfig
./bootstrap-sa.sh           # prints k8s_host + k8s_token to paste into TFC
```

## TFC variables

| Variable | Kind | Notes |
|----------|------|-------|
| `k8s_token` | env/terraform, **sensitive** | from `bootstrap-sa.sh` |
| `ghcr_pat` | terraform, **sensitive** | GitHub PAT, `read:packages` |
| `ghcr_username` | terraform | GitHub user/org owning the packages |
| `k8s_host` | terraform | defaults to `https://sno.demoland.io:6443` |

Workspace working directory: `iac/gotak-registry-secret`.

## Tool boundary

Terraform owns **only this secret**. Argo owns the app manifests and references the
secret by name (`imagePullSecrets: [{ name: ghcr }]` — already added to the
`gotak-server` / `gotak-web` deployments). They don't manage the same objects.

## Note

`sensitive = true` keeps the PAT out of plan/apply logs and the TFC UI, but the
value is still stored in state (plaintext within the encrypted state blob). That's
acceptable here because TFC state is Vault-encrypted at rest and scoped to you. To
remove it from state entirely later, have Terraform read the PAT from Vault/AWS
Secrets Manager via a `data` source instead of a variable.
