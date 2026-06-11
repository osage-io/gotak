# Network workspace (gotak-network) — Terraform + provider pins.
# No `cloud {}` block: this is a VCS-driven HCP Terraform workspace, so the
# workspace (and its state) is defined in Terraform Cloud, not in code.
terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
