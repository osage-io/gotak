# Layer 1 — the 3-node OpenShift cluster on AWS.
#
# PATH A (recommended for the demo): ROSA via the rhcs provider — managed control plane,
# 3 worker nodes, least moving parts in a live demo. Red Hat manages the nodes, so the
# "Ansible configures the servers" work shifts UP to platform/app config (see ansible/).
#
# PATH B (heavier, more "we built it"): self-managed UPI install on 3 EC2 nodes. Terraform
# provisions the EC2 + RHCOS ignition; Ansible bootstraps the cluster. Use only if the
# story needs hand-built nodes. Not wired here.
#
# module "cluster" {
#   source       = "terraform-redhat/rosa-hcp/rhcs"   # ROSA Hosted Control Plane
#   cluster_name = var.cluster_name
#   aws_subnet_ids        = module.network.private_subnets
#   replicas              = var.worker_count          # 3
#   openshift_version     = "4.16"
#   ...
# }

# Outputs consumed by providers.tf (kubernetes/helm) and later layers:
# output "api_url" { value = module.cluster.api_url }
# output "token"   { value = module.cluster.token, sensitive = true }
# output "ca_cert" { value = module.cluster.ca_cert }
