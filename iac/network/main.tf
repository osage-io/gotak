# Single-AZ VPC for the GoTAK ROSA Classic demo.
#   public subnet  -> IGW (ingress, NAT host)
#   private subnet -> NAT (ROSA worker nodes, egress-only)
# Subnets carry the ROSA-required ELB role tags so the cluster can place load
# balancers correctly. Outputs feed the gotak-compute workspace.

locals {
  name = var.cluster_name
}

resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true # required by ROSA

  tags = merge(var.tags, {
    Name = "${local.name}-${var.vpc_name_suffix}"
  })
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${local.name}-igw"
  })
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.public_subnet_cidr
  availability_zone       = var.availability_zone
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name                     = "${local.name}-public-${var.availability_zone}"
    "kubernetes.io/role/elb" = "1" # ROSA: public/internet-facing load balancers
  })
}

resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnet_cidr
  availability_zone = var.availability_zone

  tags = merge(var.tags, {
    Name                              = "${local.name}-private-${var.availability_zone}"
    "kubernetes.io/role/internal-elb" = "1" # ROSA: internal load balancers
  })
}

resource "aws_eip" "nat" {
  domain = "vpc"

  tags = merge(var.tags, {
    Name = "${local.name}-nat-eip"
  })
}

resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public.id

  tags = merge(var.tags, {
    Name = "${local.name}-nat"
  })

  depends_on = [aws_internet_gateway.main]
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = merge(var.tags, {
    Name = "${local.name}-public-rt"
  })
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main.id
  }

  tags = merge(var.tags, {
    Name = "${local.name}-private-rt"
  })
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private" {
  subnet_id      = aws_subnet.private.id
  route_table_id = aws_route_table.private.id
}
