variable "aws_region" {
  description = "AWS region (must match the network workspace)."
  type        = string
  default     = "us-east-2"
}

variable "cluster_name" {
  description = "Cluster name this installer node serves; used for resource naming."
  type        = string
  default     = "gotak"
}

variable "instance_type" {
  description = "Installer node type (arm64 — runs the aarch64 openshift-install/ccoctl natively)."
  type        = string
  default     = "t4g.medium"
}

variable "root_volume_gb" {
  description = "Root EBS volume size. The installer binary + release extracts need a few GB."
  type        = number
  default     = 40
}

variable "ssh_ingress_cidr" {
  description = "CIDR allowed to SSH to the installer node. Tighten to your IP/32 for anything beyond a short-lived demo."
  type        = string
  default     = "0.0.0.0/0"
}

variable "ssh_public_key" {
  description = "SSH public key installed on the node (default: dfed01)."
  type        = string
  default     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCi6HlM9Sj68uK5fiHFwepIq6uZL2xE/TFdtik4xm9Nmm4cviXSdnxaivGMKEPEZl0okZk69CPa9rTktY28OAq5+L1NHRm/dHbITJxFKLevDIYjVQc1zjvCnkqI23T31LO20ZQClvz86xRu8t12RPiu1Q9TDjxNwa9aIQjQesGuxXrZkrBjxJSyvBSpEUiTAYcaR04C7cAlEK+SitXmHbZyTLOAtCnv0DRrJh4bpsBcGGgKKnHfybLBPhKZCLLPaC2vSphlGLYHMdRtAvolijZUbXIqIknLcSKgvtPLyzL+0XCixZAlBPMfmgVH/F1OWspeR/sxwm5Cw0EsuTo8onPjL dfed01"
}

# Cross-workspace state wiring.
variable "tfc_organization" {
  description = "Terraform Cloud organization holding the gotak-network workspace."
  type        = string
  default     = "org-LoxuyV1DiwAxdXPf"
}

variable "network_workspace_name" {
  description = "Network workspace whose VPC/subnets this node lives in."
  type        = string
  default     = "gotak-network"
}
