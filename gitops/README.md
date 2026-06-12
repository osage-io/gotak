# GitOps (OpenShift GitOps / Argo CD)

Declarative deployment of everything that fits GitOps. An **app-of-apps**
(`root-app.yaml`) syncs the child Applications under `apps/`:

| App | Source | Wave | Notes |
|-----|--------|------|-------|
| `consul` | Helm chart 2.0.0 + `openshift/platform/consul-values.yaml` | 1 | mesh + CRDs |
| `vault` | Helm chart 0.33.0 + `openshift/platform/vault-values.yaml` | 1 | chart only (init is imperative) |
| `api-gateway` | `openshift/platform/consul-api-gateway.yaml` | 2 | needs Consul CRDs |
| `intentions` | `openshift/platform/consul-intentions.yaml` | 2 | needs Consul CRDs |
| `gotak` | `openshift/*.yaml` (postgres, server, web, SCC) | 3 | after mesh |

Charts pull values from this repo via Argo **multi-source** (`$values`), so there's
no duplication — the same values files the scripts use.

## What stays imperative (not GitOps)

GitOps is declarative; these are one-time/sensitive and live as scripts:

- **KMS key + alias** `alias/gotak-vault-unseal` and the **`vault-aws-creds`
  Secret** — prereqs for the Vault app to unseal (`install-platform.sh` steps).
- **Vault init / CORS / gotak token** — `vault-post-config.sh`.
- **Custom-domain Routes + API named cert** — they carry the `*.demoland.io`
  cert (sensitive); `custom-domain-routes.sh` / `apiserver-named-cert.sh`. Move
  these into GitOps later with External Secrets / Sealed Secrets if desired.
- **Boundary** — runs on the EC2 node, configured by Ansible (`iac/ansible-boundary`).

## Activate

```bash
export KUBECONFIG=../iac/sno/cluster-auth/kubeconfig

# 1. Install the operator + grant the controller rights (one time)
oc apply -f bootstrap/01-operator.yaml
oc -n openshift-operators rollout status deploy/openshift-gitops-operator-controller-manager --timeout=300s
oc wait --for=condition=Available -n openshift-gitops deploy/openshift-gitops-server --timeout=300s
oc apply -f bootstrap/02-rbac.yaml

# 2. Prereqs for Vault (KMS key/alias + creds Secret) — once
cd ../openshift/platform && ./install-platform.sh   # or just the KMS+Secret steps; Argo handles the charts

# 3. Point Argo at the repo
oc apply -f ../../gitops/root-app.yaml

# 4. Watch it converge
oc get applications -n openshift-gitops
```

Argo CD console: `oc get route -n openshift-gitops openshift-gitops-server`.
Admin password: `oc get secret openshift-gitops-cluster -n openshift-gitops -o jsonpath='{.data.admin\.password}' | base64 -d`.

## Notes

- `targetRevision` is the `openshift-platform-vault-consul` branch in every
  Application — change to `main` after you merge.
- Wave ordering across child apps is best-effort; Argo's retry/selfHeal converges
  the CRD-dependent apps (gateway, intentions) once Consul's CRDs register.
- The gotak app pods need the Consul injector running to get sidecars; if any pod
  starts before the injector is ready, `oc rollout restart deploy -n gotak` re-injects.
