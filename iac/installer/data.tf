# VPC / subnets from the gotak-network workspace.
# Requires "Remote state sharing" enabled on gotak-network -> gotak-installer.
data "terraform_remote_state" "network" {
  backend = "remote"

  config = {
    organization = var.tfc_organization
    workspaces = {
      name = var.network_workspace_name
    }
  }
}

# Always-current Amazon Linux 2023 arm64 AMI.
data "aws_ssm_parameter" "al2023_arm64" {
  name = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64"
}
