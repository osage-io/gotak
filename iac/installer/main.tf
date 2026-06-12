# Installer / ops node for the single-node OpenShift (SNO) cluster.
#
# Why it exists:
#   - Runs openshift-install + ccoctl natively on Linux (no laptop quirks).
#   - Its IAM instance profile supplies auto-rotating credentials, so the
#     install never dies to expired sandbox session tokens, and a cron here can
#     re-stamp the cluster's operator credential secrets.
#   - Persists as the cluster ops box that Ansible drives (install + platform).

locals {
  name             = "${var.cluster_name}-installer"
  public_subnet_id = data.terraform_remote_state.network.outputs.public_subnet_ids[0]
  vpc_id           = data.terraform_remote_state.network.outputs.vpc_id
}

resource "aws_key_pair" "installer" {
  key_name   = local.name
  public_key = var.ssh_public_key
}

resource "aws_security_group" "installer" {
  name        = local.name
  description = "SSH to the OpenShift installer node"
  vpc_id      = local.vpc_id

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.ssh_ingress_cidr]
  }

  egress {
    description = "all egress"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = local.name }
}

# Permissions openshift-install (IPI) + ccoctl-style cleanup need. The sandbox
# SCP still applies on top of this (e.g. iam:CreateUser / OIDC providers stay
# blocked account-wide), which is fine: the manual-credentials install path
# only needs the role/instance-profile subset, which the sandbox allows.
data "aws_iam_policy_document" "installer" {
  statement {
    sid = "Infra"
    actions = [
      "ec2:*",
      "elasticloadbalancing:*",
      "route53:*",
      "s3:*",
      "tag:GetResources",
      "tag:TagResources",
      "tag:UntagResources",
      "ssm:GetParameter",
      "ssm:GetParameters",
      "servicequotas:GetServiceQuota",
      "sts:GetCallerIdentity",
    ]
    resources = ["*"]
  }

  statement {
    sid = "KmsForVault"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:DescribeKey",
      "kms:CreateKey",
      "kms:CreateAlias",
      "kms:ListKeys",
      "kms:ListAliases",
    ]
    resources = ["*"]
  }

  statement {
    sid = "IamForCluster"
    actions = [
      "iam:AddRoleToInstanceProfile",
      "iam:CreateInstanceProfile",
      "iam:CreateRole",
      "iam:DeleteInstanceProfile",
      "iam:DeleteRole",
      "iam:DeleteRolePolicy",
      "iam:GetInstanceProfile",
      "iam:GetRole",
      "iam:GetRolePolicy",
      "iam:ListAttachedRolePolicies",
      "iam:ListInstanceProfiles",
      "iam:ListInstanceProfilesForRole",
      "iam:ListRolePolicies",
      "iam:ListRoles",
      "iam:PassRole",
      "iam:PutRolePolicy",
      "iam:RemoveRoleFromInstanceProfile",
      "iam:TagInstanceProfile",
      "iam:TagRole",
      "iam:UntagRole",
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role" "installer" {
  name = local.name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "installer" {
  name   = "openshift-installer"
  role   = aws_iam_role.installer.id
  policy = data.aws_iam_policy_document.installer.json
}

resource "aws_iam_instance_profile" "installer" {
  name = local.name
  role = aws_iam_role.installer.name
}

resource "aws_instance" "installer" {
  ami                         = nonsensitive(data.aws_ssm_parameter.al2023_arm64.value)
  instance_type               = var.instance_type
  subnet_id                   = local.public_subnet_id
  associate_public_ip_address = true
  key_name                    = aws_key_pair.installer.key_name
  vpc_security_group_ids      = [aws_security_group.installer.id]
  iam_instance_profile        = aws_iam_instance_profile.installer.name

  root_block_device {
    volume_size = var.root_volume_gb
    volume_type = "gp3"
  }

  # Ansible needs python3 (preinstalled on AL2023); add the basics it shells to.
  user_data = <<-EOF
    #!/bin/bash
    dnf install -y git tar gzip jq
  EOF

  tags = { Name = local.name }
}
