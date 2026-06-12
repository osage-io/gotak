# GoTAK Installer Node + KMS Key

Terraform workspace for the installer/bastion node and Vault KMS auto-unseal key.

## What This Creates

1. **Installer Node** (`t4g.medium` arm64)
   - Runs `openshift-install` and Ansible
   - IAM instance profile with auto-rotating credentials
   - SSH access for operations

2. **KMS Key for Vault** (NEW)
   - Auto-unseal key for Vault
   - Managed by Terraform (not bash script)
   - Key rotation enabled
   - 7-day deletion window

## Prerequisites

- Terraform Cloud workspace: `gotak-installer`
- Network workspace applied: `gotak-network`
- SSH public key (default: dfed01)

## Usage

### Initial Apply

```bash
cd iac/installer
terraform init
terraform plan
terraform apply
```

### Outputs

```bash
# Get installer node IP
terraform output public_ip

# Get SSH command
terraform output ssh_command

# Get KMS key ID for Vault
terraform output kms_key_id

# Get KMS alias
terraform output kms_alias
```

## KMS Key Management

**Before this change:** The `install-platform.sh` script created the KMS key using AWS CLI.

**After this change:** The KMS key is managed by Terraform in `kms.tf`.

### Why Terraform?

- **Infrastructure as Code**: Key lifecycle tracked in git
- **State Management**: Terraform knows if key exists
- **Idempotent**: Safe to run multiple times
- **Proper Tagging**: Consistent resource tagging
- **Key Rotation**: Enabled by default
- **Deletion Protection**: 7-day window prevents accidents

### Migration Path

If you already have a KMS key created by the bash script:

```bash
# Import existing key into Terraform state
terraform import aws_kms_key.vault_unseal <key-id>
terraform import aws_kms_alias.vault_unseal alias/gotak-vault-unseal

# Verify state
terraform plan  # Should show no changes
```

## Platform Deployment

After applying this workspace, deploy the platform:

```bash
cd ../../openshift/platform
./install-platform.sh
```

The script now expects the KMS key to exist (created by Terraform).

## IAM Permissions

The installer role has these KMS permissions:

- `kms:Encrypt`
- `kms:Decrypt`
- `kms:DescribeKey`
- `kms:CreateKey`
- `kms:CreateAlias`
- `kms:ListKeys`
- `kms:ListAliases`

## Troubleshooting

### KMS Key Not Found

```
ERROR: KMS key not found. Run 'terraform apply' in iac/installer first.
```

**Solution:** Apply this Terraform workspace first:
```bash
cd iac/installer
terraform apply
```

### Permission Denied

```
Error: creating KMS Key: AccessDeniedException
```

**Solution:** Check your AWS credentials have KMS permissions. Sandbox accounts may restrict `kms:CreateKey`.

### Key Already Exists

If you get a conflict, import the existing key:
```bash
terraform import aws_kms_key.vault_unseal <existing-key-id>
```

## Files

- `main.tf` - Installer node and IAM role
- `kms.tf` - KMS key for Vault auto-unseal (NEW)
- `outputs.tf` - Outputs including KMS key details
- `variables.tf` - Input variables
- `data.tf` - Data sources (AMI lookup, network state)
- `providers.tf` - AWS provider configuration
- `versions.tf` - Terraform version constraints

## Related Documentation

- [Platform Deployment](../../openshift/platform/README.md)
- [Network Infrastructure](../network/README.md)
- [Deployment Architecture](../../docs/DEPLOYMENT_ARCHITECTURE.md)
