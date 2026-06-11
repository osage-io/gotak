provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = "gotak"
      ManagedBy = "Terraform"
      Workspace = "gotak-compute"
    }
  }
}

# The rhcs provider authenticates to Red Hat OpenShift Cluster Manager (OCM).
# It reads the token from the RHCS_TOKEN environment variable (set it as a
# sensitive env var in the goTak project's Variable Set).
provider "rhcs" {}
