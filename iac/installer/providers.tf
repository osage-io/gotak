provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = "gotak"
      ManagedBy = "Terraform"
      Workspace = "gotak-installer"
    }
  }
}
