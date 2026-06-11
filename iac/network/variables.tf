variable "aws_region" {
  description = "AWS region for the VPC."
  type        = string
  default     = "us-east-2"
}

variable "availability_zone" {
  description = "Single AZ to place the subnets in (must be within aws_region)."
  type        = string
  default     = "us-east-2a"
}

variable "cluster_name" {
  description = "Name used for resource naming and the ROSA subnet role tags. Keep in sync with the compute workspace."
  type        = string
  default     = "gotak-demo"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidr" {
  description = "CIDR for the public subnet (ingress / NAT)."
  type        = string
  default     = "10.0.0.0/24"
}

variable "private_subnet_cidr" {
  description = "CIDR for the private subnet (ROSA worker nodes)."
  type        = string
  default     = "10.0.1.0/24"
}

variable "tags" {
  description = "Extra tags merged onto all resources."
  type        = map(string)
  default     = {}
}
