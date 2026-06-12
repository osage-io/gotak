# KMS key for Vault auto-unseal
#
# The Vault pods read AWS credentials from a Secret (vault-aws-creds) and use this
# KMS key to auto-unseal on startup. The installer node's instance profile already
# has kms:* permissions via the installer policy, so it can create/use this key.

resource "aws_kms_key" "vault_unseal" {
  description             = "GoTAK Vault auto-unseal key"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = {
    Name        = "${var.cluster_name}-vault-unseal"
    Application = "gotak"
    Component   = "vault"
    ManagedBy   = "terraform"
  }
}

resource "aws_kms_alias" "vault_unseal" {
  name          = "alias/${var.cluster_name}-vault-unseal"
  target_key_id = aws_kms_key.vault_unseal.key_id
}

# Grant the installer node's role permission to use the key
resource "aws_kms_grant" "installer_vault" {
  name              = "${var.cluster_name}-installer-vault-unseal"
  key_id            = aws_kms_key.vault_unseal.key_id
  grantee_principal = aws_iam_role.installer.arn

  operations = [
    "Encrypt",
    "Decrypt",
    "DescribeKey",
  ]
}
