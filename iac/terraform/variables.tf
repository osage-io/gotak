variable "aws_region" {
  type    = string
  default = "us-gov-west-1" # GovCloud-friendly default for the federal story; switch to us-east-1 for a commercial demo
}

variable "cluster_name" {
  type    = string
  default = "gotak-demo"
}

variable "worker_count" {
  type    = int
  default = 3 # 3-node cluster
}

variable "base_domain" {
  type        = string
  description = "Route53 base domain for the cluster (e.g. demo.fedick.net)"
  default     = ""
}

# Security-layer endpoints (filled in after 20-platform deploys the services).
variable "vault_addr" {
  type    = string
  default = ""
}

variable "boundary_addr" {
  type    = string
  default = ""
}

variable "keycloak_url" {
  type    = string
  default = ""
}
