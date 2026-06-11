variable "aws_region" {
  description = "AWS region (must match the network workspace)."
  type        = string
  default     = "us-east-2"
}

variable "cluster_name" {
  description = "ROSA cluster name. MUST match cluster_name in the gotak-network workspace (subnet tags)."
  type        = string
  default     = "gotak-demo"
}

variable "openshift_version" {
  description = "ROSA OpenShift version. Must be a currently-supported install version — run `rosa list versions` and set this accordingly."
  type        = string
  default     = "4.17.0"
}

variable "replicas" {
  description = "Worker node count. Single-AZ minimum is 2; we want 3."
  type        = number
  default     = 3
}

variable "compute_machine_type" {
  description = "EC2 instance type for the worker machine pool (ROSA minimum is 4 vCPU / 16 GiB, e.g. m5.xlarge)."
  type        = string
  default     = "m5.xlarge"
}

# Cross-workspace state wiring.
variable "tfc_organization" {
  description = "Terraform Cloud organization that holds the gotak-network workspace."
  type        = string
  default     = "org-LoxuyV1DiwAxdXPf"
}

variable "network_workspace_name" {
  description = "Name of the network workspace whose outputs (VPC, subnets) this cluster consumes."
  type        = string
  default     = "gotak-network"
}
