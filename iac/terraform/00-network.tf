# Layer 0 — AWS substrate. Bob generates: VPC, public/private subnets across 3 AZs,
# IGW/NAT, security groups, IAM roles for the cluster, Route53 zone.
#
# Recommended: terraform-aws-modules/vpc/aws. Stub below marks the contract the
# cluster layer (10) expects.

# module "network" {
#   source  = "terraform-aws-modules/vpc/aws"
#   version = "~> 5.0"
#   name    = var.cluster_name
#   cidr    = "10.0.0.0/16"
#   azs             = ["${var.aws_region}a", "${var.aws_region}b", "${var.aws_region}c"]
#   private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
#   public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
#   enable_nat_gateway = true
# }

# Outputs consumed by 10-cluster.tf:
# output "vpc_id"          { value = module.network.vpc_id }
# output "private_subnets" { value = module.network.private_subnets }
