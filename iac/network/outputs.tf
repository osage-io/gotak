# Subnet IDs are emitted as lists so the compute (ROSA) workspace can pass them
# straight to the cluster module, and a future multi-AZ change is a drop-in.

output "vpc_id" {
  description = "ID of the VPC."
  value       = aws_vpc.main.id
}

output "vpc_cidr" {
  description = "CIDR block of the VPC."
  value       = aws_vpc.main.cidr_block
}

output "public_subnet_ids" {
  description = "Public subnet IDs (for the ROSA compute workspace)."
  value       = [aws_subnet.public.id]
}

output "private_subnet_ids" {
  description = "Private subnet IDs (ROSA worker nodes)."
  value       = [aws_subnet.private.id]
}

output "availability_zone" {
  description = "AZ the subnets are placed in."
  value       = var.availability_zone
}

output "region" {
  description = "AWS region."
  value       = var.aws_region
}
