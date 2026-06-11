# Read the VPC / subnets from the gotak-network workspace.
# Requires "Remote state sharing" enabled on gotak-network -> gotak-compute.
data "terraform_remote_state" "network" {
  backend = "remote"

  config = {
    organization = var.tfc_organization
    workspaces = {
      name = var.network_workspace_name
    }
  }
}
