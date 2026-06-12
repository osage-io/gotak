# gotak-installer — OpenShift installer/ops node (Workspace 3)

A small arm64 EC2 node (`t4g.medium`, AL2023) in the `gotak-network` public
subnet. **Ansible drives it to run `openshift-install`** for the single-node
OpenShift (SNO) cluster — see `../ansible/`. It persists afterward as the
cluster ops box.

Why a node instead of a laptop install:
- **Instance-profile credentials** — auto-rotating; the install can't die to
  expired sandbox session tokens, and a cron here keeps the cluster's operator
  credential secrets fresh.
- Runs `openshift-install`/`ccoctl` natively on Linux/arm64.
- Repeatable: Terraform provisions it, Ansible configures it (the demo story).

## Terraform Cloud setup (one-time)

1. **Create the workspace** in org `osage`, **goTak** project:
   - Version control workflow → repo `osage-io/gotak`
   - **Working Directory:** `iac/installer` · Branch: `main`
2. The **goTak project Variable Set** (AWS keys) already applies.
3. **Remote state sharing:** on `gotak-network` → *Settings → Remote state
   sharing* → also share with **`gotak-installer`**.

## Apply

Merge to `main` → TFC plans `gotak-installer` → Confirm & Apply.
Outputs include `ssh_command` (key: `~/.ssh/dfed01`).

## IAM notes

The node's role carries the IPI installer permission set (EC2/ELB/Route53/S3 +
the IAM role/instance-profile subset + PassRole). The sandbox SCP still blocks
account-wide what it blocks (IAM users, OIDC providers) — the manual-credentials
install path doesn't need those. `iam:PassRole` was probe-verified.

## Cost

`t4g.medium` ≈ $0.034/hr (~$0.80/day) + 40 GB gp3. Destroy via TFC when done
(destroy the SNO cluster first — see `../ansible/`).
