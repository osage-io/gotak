# Compute workspace (gotak-compute) — ROSA Classic cluster on the gotak-network VPC.
# VCS-driven HCP Terraform workspace, so no `cloud {}` block (state lives in TFC).
terraform {
  required_version = ">= 1.5"

  required_providers {
    # ROSA Classic module v1.7.x requires the AWS provider >= 6.42. (The network
    # workspace stays on v5 — separate state/lock, so the mismatch is fine.)
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    rhcs = {
      source  = "terraform-redhat/rhcs"
      version = "~> 1.6"
    }
  }
}
