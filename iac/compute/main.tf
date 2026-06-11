# ROSA Classic cluster on the bring-your-own VPC from gotak-network.
# Single-AZ, public. The module also creates the required IAM account roles,
# operator roles, and OIDC config (create_* = true) in this same apply.
#
# Prereqs (see README): ROSA enabled in the AWS account, a Red Hat OCM token in
# RHCS_TOKEN, sufficient EC2 vCPU quota, and a supported openshift_version.

locals {
  # ROSA wants all cluster subnets (public + private) in one list.
  subnet_ids = concat(
    data.terraform_remote_state.network.outputs.public_subnet_ids,
    data.terraform_remote_state.network.outputs.private_subnet_ids,
  )
}

module "rosa" {
  source  = "terraform-redhat/rosa-classic/rhcs"
  version = "~> 1.7"

  cluster_name      = var.cluster_name
  openshift_version = var.openshift_version

  # Provision the IAM prerequisites alongside the cluster.
  create_account_roles  = true
  create_operator_roles = true
  create_oidc           = true
  account_role_prefix   = var.cluster_name
  operator_role_prefix  = var.cluster_name

  # Bring-your-own VPC from gotak-network. Single-AZ, public endpoint.
  aws_availability_zones = [data.terraform_remote_state.network.outputs.availability_zone]
  aws_subnet_ids         = local.subnet_ids
  machine_cidr           = data.terraform_remote_state.network.outputs.vpc_cidr
  multi_az               = false
  private                = false

  replicas             = var.replicas
  compute_machine_type = var.compute_machine_type
}
