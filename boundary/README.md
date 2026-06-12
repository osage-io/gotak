# Boundary — privileged access to the gotak environment

Runs **Boundary (controller + worker) on the EC2 node** via podman, as a bastion in
front of the cluster. You connect with the Boundary desktop client / CLI and the
node worker brokers the session — credentials and target addresses never touch
your laptop directly.

```
your laptop ──auth/proxy──▶  node:9200 (API) / node:9202 (proxy)
   Boundary client                 │  controller + ingress worker (podman)
                                    ├─ openshift-api   api.<domain>:6443
                                    ├─ vault-ui        <vault Route>:443
                                    ├─ consul-ui       <consul Route>:443
                                    ├─ node-ssh        127.0.0.1:22  (this host)
                                    └─ postgres        <SNO node>:30432  (NodePort)
```

Why on the node (not in-cluster): the node already has a public IP, so the worker
proxy (tcp/9202) is directly reachable by your client — no LoadBalancer/NodePort
gymnastics for Boundary itself — and the bastion stays independent of the cluster
it guards. Keys use self-contained `aead` (no AWS-cred dependency); swap to
`kms "awskms"` for production.

## Install (run on the node)

```bash
# from your Mac — copy the two files up
scp -i ~/.ssh/dfed01 boundary/install-boundary.sh boundary/boundary-config.hcl.tpl \
    ec2-user@<node-ip>:~

# on the node
ssh -i ~/.ssh/dfed01 ec2-user@<node-ip>
sudo bash install-boundary.sh        # installs Boundary + podman Postgres, inits DB, starts systemd
```

**Save the `boundary database init` output** it prints (also in
`/etc/boundary/init-output.txt`) — that's the generated global admin.

## Expose the in-cluster targets (run from your Mac)

```bash
export KUBECONFIG=$PWD/iac/sno/cluster-auth/kubeconfig
oc apply -f boundary/cluster-exposure.yaml        # consul-ui Route + postgres NodePort
oc get route -n gotak vault consul-ui              # note the hostnames
```

Open the node security group for your client IP on **9200** and **9202**, and (for
the Postgres target) allow the **SNO node** to receive **30432** from the
Boundary-node's security group.

## Create the targets (run on the node)

```bash
sudo PG_HOST=<sno-node-ip> PG_PORT=30432 \
     VAULT_HOST=<vault Route host> \
     CONSUL_HOST=<consul-ui Route host> \
     bash configure-targets.sh
```

It creates the `gotak-org`/`gotak-project` scopes, a password login, an access
role, and the five targets, then prints the **login-name + password** for the
desktop client. Connect, pick a target, and Boundary opens a local proxy port —
e.g. `psql -h 127.0.0.1 -p <local-port>` for Postgres, or `ssh` to the node.

## ⚠️ The Postgres target and the mesh

Postgres is a **meshed** service — its inbound 5432 is redirected to the Envoy
sidecar, which denies the off-mesh NodePort connection under default-deny. Two ways
to broker it:

1. **NodePort + exclude (simple, ready here).** Add the exclusion annotation (one
   `oc patch`, shown in [`cluster-exposure.yaml`](cluster-exposure.yaml)). Cost: the
   `gotak-server → postgres` intention is no longer enforced on 5432; use the
   `gateway → gotak-server` intention as the mesh-enforcement showcase instead.
2. **In-cluster egress worker (pristine, advanced).** Keep postgres fully meshed and
   run a second Boundary worker as a pod in the mesh (multi-hop: node ingress worker
   → in-cluster egress worker → postgres, with a `worker → postgres` intention). No
   NodePort, nothing excluded. This is the "correct" pattern; wire it up if you want
   the mesh demo and brokered Postgres at the same time.

The other four targets have no such caveat — they're externally routable
(API/Vault/Consul) or local to the node (SSH).

## Teardown

```bash
# on the node
sudo systemctl disable --now boundary
sudo podman rm -f boundary-pg
# on your Mac
oc delete -f boundary/cluster-exposure.yaml
```
