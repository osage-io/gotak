# GoTAK Secure-Runtime IaC (Bob-generated, 3-node OpenShift on AWS)

The "automation builds the secure runtime" stack. **IBM Bob** generates the Terraform and
Ansible into the layered skeleton here; a human reviews and applies it. The result: a 3-node
OpenShift cluster on AWS, the HashiCorp + Keycloak zero-trust control plane, and GoTAK
deployed already enrolled in that control plane.

This is the Phase-2 layer above `../openshift/` (which holds the app manifests for Phase 1).

## Layers

```
iac/
├── terraform/
│   ├── providers.tf     aws, rhcs(ROSA), kubernetes, helm, vault, consul, boundary, keycloak
│   ├── variables.tf     region, cluster_name, worker_count=3, base_domain
│   ├── 00-network.tf    AWS VPC / subnets / IAM / Route53           (Bob)
│   ├── 10-cluster.tf    3-node ROSA OpenShift cluster                (Bob)
│   ├── 20-platform.tf   Vault, Consul, Boundary, Keycloak via Helm   (Bob)
│   ├── 30-security.tf   the zero-trust wiring — "secures it auto"    (Bob)
│   └── 40-gotak.tf      deploy ../../openshift onto the cluster       (Bob)
├── ansible/
│   ├── site.yml         platform + security config, reusable roles   (Bob)
│   ├── inventory.example.ini
│   └── roles/           vault / consul / boundary / keycloak / gotak (Bob)
└── BOB-PROMPTS.md       per-layer generation prompts
```

## Two install paths (pick before building)

**Path A — ROSA (recommended for the demo).** Terraform's `rhcs` provider stands up a managed
3-node OpenShift cluster. Red Hat runs the nodes, so **Ansible's job shifts up to platform/app
configuration** (Vault/Consul/Boundary/Keycloak/GoTAK), which is the cleaner, lower-risk live
demo and still showcases Ansible. Fits the Red Hat block of the briefing.

**Path B — self-managed UPI on 3 EC2 nodes.** Terraform provisions EC2 + RHCOS; **Ansible
configures the actual servers** and bootstraps the cluster. More "we hand-built it," much more
to go wrong live. Use only if the story specifically needs node-level config. Hooks are noted
in `10-cluster.tf` and `ansible/site.yml` but not wired.

> Recommendation: **Path A for the live demo**, mention Path B as the on-prem/edge variant
> (which is where USCG mission systems actually live — a good forward-looking aside).

## The division of labor

- **Terraform** = *provision* (cluster, platform services, security objects). Declarative, stateful.
- **Ansible** = *configure* (init/unseal Vault, bootstrap ACLs, load realms/policies, render app config). Procedural, ordered.
- **Bob** = *generate* both, from intent, into reusable modules/roles.

## Apply order

`terraform apply` 00→10 (cluster up) → wire kube/helm providers → 20 (platform) → set
`vault_addr`/`boundary_addr`/`keycloak_url` → 30 (security) + `ansible-playbook site.yml` → 40 (GoTAK).

See `Ideas/GoTAK Agentic Runtime Security Demo` in the vault for the briefing narrative.
