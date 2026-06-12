# Boundary (Ansible) — PAM bastion on the EC2 node

Installs and configures **Boundary (controller + worker)** on the node via podman,
the Ansible way. Replaces the hand-run `boundary/*.sh` scripts; same outcome.

```
your laptop ──auth/proxy──▶ boundary.demoland.io:9200 / :9202
                                 controller + worker (podman) + Postgres (podman)
                                 targets: openshift-api, vault-ui, consul-ui,
                                          node-ssh, postgres
```

## Layout

```
site.yml                     play: hosts=boundary, become
group_vars/all.yml           version, public_addr, login name, target list
roles/boundary/
  tasks/install.yml          podman+Boundary, aead keys (once), Postgres, DB init, systemd
  tasks/targets.yml          org/project, login, role, targets (recovery KMS, run-once)
  templates/boundary.hcl.j2  controller+worker config
  templates/boundary.service.j2
  handlers/main.yml          restart boundary
```

## Run

```bash
cp inventory.example.ini inventory.ini    # set the node IP
# fill in group_vars/all.yml: the postgres target's SNO node IP, route hosts
ansible-playbook -i inventory.ini site.yml

# phases:
ansible-playbook -i inventory.ini site.yml --tags install   # just install/config
ansible-playbook -i inventory.ini site.yml --tags targets   # just create targets
```

Idempotent: aead keys + the Postgres password are generated once into
`/etc/boundary/keys.env` and reused; DB init and target creation are guarded by
marker files, so re-runs are safe.

## After it runs

- Login is printed at the end (and saved on the node at
  `/etc/boundary/login-pass.txt`): user `gotak`, generated password.
- Connect with the Boundary desktop client / CLI to `http://boundary.demoland.io:9200`.
- Open the node security group for your client IP on **9200** and **9202**.
- For the **postgres** target: apply `boundary/cluster-exposure.yaml` (consul-ui
  Route + postgres NodePort) and set the target's address to the SNO node IP. See
  that file for the mesh trade-off note on the Postgres NodePort.

## Prereqs

- The node reachable over SSH (key in `inventory.ini`).
- DNS: `A boundary -> node public IP` (no cert needed — the demo config uses
  `tls_disable`).
