# Installer-node workspace (gotak-installer) — bastion/ops box that runs
# openshift-install (driven by Ansible) against the gotak-network VPC.
# VCS-driven HCP Terraform workspace: no `cloud {}` block; state lives in TFC.
terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
