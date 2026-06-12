# ansible-sno — install single-node OpenShift from the installer node

Drives `openshift-install` on the **gotak-installer** EC2 node (see
`../installer/`) to build a **single-node OpenShift** cluster
(`gotak.daniel-fedick.aws.sbx.hashicorpdemo.com`, 1 × `m6g.2xlarge` arm64) on
the `gotak-network` VPC.

Separate from `../ansible/` (the Bob Phase-2 skeleton), which stays as-is.

## How credentials work (the trick that fits the sandbox)

The sandbox blocks IAM users and OIDC providers, so neither of OpenShift's
default credential modes works. Instead: **CCO Manual mode** — the per-operator
AWS credential Secrets are generated from the installer node's
**instance-profile credentials** (IMDSv2) at install time, and a **cron
(every 30 min)** re-stamps them so the cluster never holds stale creds. No
laptop session tokens involved anywhere.

## Run it

```bash
cd iac/ansible-sno
cp inventory.example.ini inventory.ini   # set the node IP from TFC gotak-installer outputs
ansible-playbook site.yml                # full install (~50 min total)
```

Prereqs on the control machine:
- `ansible` (`brew install ansible`)
- `~/.ssh/dfed01` (private key for the node)
- the pull secret at `../sno/pull-secret.json` (gitignored; already in place)

Phases (tags): `tools`, `config`, `install`, `postinstall` — e.g. re-fetch auth
with `-t postinstall`.

## What you get

- Console: `https://console-openshift-console.apps.gotak.daniel-fedick.aws.sbx.hashicorpdemo.com`
- `kubeconfig` + `kubeadmin-password` fetched to `../sno/cluster-auth/` (gitignored)
- Cred-refresh cron + logs on the node (`~/sno/refresh-creds.log`)

## Destroy

```bash
ssh -i ~/.ssh/dfed01 ec2-user@<node-ip> \
  'openshift-install destroy cluster --dir ~/sno/cluster'
```
Then destroy the `gotak-installer` workspace, then `gotak-network` (in that order).

## Safety

The play **refuses to run the install phase if installer state already exists**
on the node (protects a running cluster). Destroy first or remove
`~/sno/cluster` deliberately.
