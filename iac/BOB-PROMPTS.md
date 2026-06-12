# IBM Bob — generation targets per layer

Bob fills each layer; a human reviews and applies (Trust Ladder posture). Suggested prompts:

**00-network.tf** — "Generate Terraform using terraform-aws-modules/vpc for a 3-AZ VPC in
`{{aws_region}}` with public + private subnets, NAT, and the IAM roles a ROSA cluster needs.
Output `vpc_id` and `private_subnets`."

**10-cluster.tf** — "Generate a ROSA HCP cluster with the rhcs provider: 3 workers, OpenShift
4.16, on the private subnets from module.network. Output api_url, token, ca_cert."

**20-platform.tf** — "Generate helm_release resources for Vault (HA, injector), Consul
(connect inject, OpenShift mode), Boundary, and Keycloak, each in its own namespace."

**30-security.tf** — "Generate the zero-trust wiring: Keycloak realm `gotak` + confidential
`gotak-agent` client; Vault OIDC auth federated to that realm, an agent policy, and a Postgres
dynamic-secrets role; Boundary scope + OIDC auth method + a target; Consul service-intentions
allowing only gotak-server -> postgres."

**40-gotak.tf** — "Deploy the manifests in ../../openshift via kustomize, rendering server.yaml
to use Vault OIDC + dynamic DB creds, register in Consul, and reach its target through Boundary."

**ansible/roles/** — "Convert these configuration steps into reusable Ansible roles
(vault, consul, boundary, keycloak, gotak), one responsibility each."
