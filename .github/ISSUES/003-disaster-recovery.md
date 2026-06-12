# Issue #3: Document Disaster Recovery Procedures

**Epic:** [Deployment Architecture](deployment-architecture-epic.md)
**Phase:** 5 - Documentation & Operations
**Priority:** 🟡 Medium
**Estimated Effort:** 1 day
**Status:** Ready for Work

## Problem Statement

There are currently no documented disaster recovery (DR) procedures for the GoTAK deployment. In the event of data loss, component failure, or complete environment loss, operators have no clear guidance on how to recover. This creates significant operational risk.

## Objective

Create comprehensive disaster recovery documentation covering backup procedures, restore procedures, RTO/RPO targets, and testing schedules for all critical components.

## Acceptance Criteria

- [ ] Backup procedures for all stateful components
- [ ] Restore procedures with step-by-step instructions
- [ ] RTO (Recovery Time Objective) and RPO (Recovery Point Objective) defined
- [ ] Automated backup scripts where applicable
- [ ] DR testing schedule and procedures
- [ ] Runbook for common failure scenarios
- [ ] Data retention policies documented
- [ ] Off-site backup strategy defined

## Scope

### Components Requiring DR

#### Critical (Must Backup)
1. **Vault Data**
   - Raft storage backend
   - Encryption keys
   - Secrets and policies
   - PKI certificates

2. **PostgreSQL Database**
   - Application data
   - User accounts
   - Mission data
   - Chat history

3. **Consul Data**
   - Service mesh configuration
   - KV store
   - Intentions and policies

4. **Configuration**
   - Terraform state
   - Ansible inventories
   - GitOps manifests
   - Custom certificates

#### Important (Should Backup)
5. **Boundary Data**
   - Target configurations
   - User sessions
   - Access policies

6. **Argo CD Configuration**
   - Application definitions
   - Sync state

### Out of Scope
- Application code (in git)
- Container images (in registry)
- Ephemeral data (logs, metrics)
- Development environments

## Technical Details

### Document Structure

```markdown
# GoTAK Disaster Recovery Guide

## Overview
- DR strategy and objectives
- RTO/RPO targets
- Backup schedule
- Retention policies

## Backup Procedures

### Vault Backup
- Raft snapshot procedure
- Encryption key backup
- Recovery key storage
- Automated backup script
- Verification steps

### PostgreSQL Backup
- pg_dump procedure
- Point-in-time recovery setup
- Automated backup script
- Backup verification
- Retention policy

### Consul Backup
- Snapshot procedure
- KV store backup
- Configuration backup
- Automated backup script

### Configuration Backup
- Terraform state backup
- GitOps repository backup
- Certificate backup
- Secrets backup (encrypted)

## Restore Procedures

### Complete Environment Recovery
1. Infrastructure rebuild
2. Platform restoration
3. Data restoration
4. Validation

### Component-Specific Recovery
- Vault restore
- Database restore
- Consul restore
- Boundary restore

### Partial Recovery
- Single service recovery
- Data corruption recovery
- Configuration rollback

## Failure Scenarios

### Scenario 1: Database Corruption
- Detection
- Impact assessment
- Recovery steps
- Validation

### Scenario 2: Vault Seal/Data Loss
- Detection
- Recovery from backup
- Re-initialization if needed
- Secret restoration

### Scenario 3: Complete Cluster Loss
- Infrastructure rebuild
- Platform redeployment
- Data restoration
- Service validation

### Scenario 4: Accidental Deletion
- Identify what was deleted
- Restore from backup
- Verify integrity

## RTO/RPO Targets

| Component | RPO | RTO | Backup Frequency |
|-----------|-----|-----|------------------|
| Vault | 1 hour | 2 hours | Hourly |
| PostgreSQL | 15 min | 1 hour | Every 15 min |
| Consul | 1 hour | 1 hour | Hourly |
| Configuration | 1 day | 30 min | Daily |

## Backup Storage

### Primary Backup Location
- S3 bucket configuration
- Encryption at rest
- Access controls
- Lifecycle policies

### Off-Site Backup
- Secondary region/account
- Replication strategy
- Access procedures

## Testing Procedures

### Monthly DR Test
- Restore to test environment
- Validate data integrity
- Document results
- Update procedures

### Quarterly Full DR Test
- Complete environment rebuild
- Full data restoration
- Application validation
- Performance testing

## Automation

### Backup Scripts
- Vault snapshot script
- Database backup script
- Consul snapshot script
- Configuration backup script

### Monitoring
- Backup success/failure alerts
- Storage capacity monitoring
- Backup age monitoring
- Restore test reminders

## Appendix
- Backup script reference
- S3 bucket configuration
- Encryption key management
- Contact information
```

### Backup Script Examples

**Vault Backup Script:**
```bash
#!/bin/bash
# vault-backup.sh
BACKUP_DIR="/backups/vault"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
VAULT_ADDR="https://vault-gotak.apps.cluster"

# Take Raft snapshot
vault operator raft snapshot save "${BACKUP_DIR}/vault-${TIMESTAMP}.snap"

# Upload to S3
aws s3 cp "${BACKUP_DIR}/vault-${TIMESTAMP}.snap" \
  s3://gotak-backups/vault/ \
  --sse AES256

# Cleanup old local backups (keep 7 days)
find "${BACKUP_DIR}" -name "vault-*.snap" -mtime +7 -delete
```

**PostgreSQL Backup Script:**
```bash
#!/bin/bash
# postgres-backup.sh
BACKUP_DIR="/backups/postgres"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
DB_NAME="gotak"

# Dump database
pg_dump -h postgres.gotak.svc -U gotak -Fc \
  "${DB_NAME}" > "${BACKUP_DIR}/gotak-${TIMESTAMP}.dump"

# Upload to S3
aws s3 cp "${BACKUP_DIR}/gotak-${TIMESTAMP}.dump" \
  s3://gotak-backups/postgres/ \
  --sse AES256

# Cleanup old local backups (keep 7 days)
find "${BACKUP_DIR}" -name "gotak-*.dump" -mtime +7 -delete
```

## Implementation Plan

### Step 1: Define Strategy (2 hours)
- Determine RTO/RPO targets
- Choose backup storage
- Define retention policies
- Plan automation approach

### Step 2: Document Backup Procedures (3 hours)
- Vault backup procedure
- PostgreSQL backup procedure
- Consul backup procedure
- Configuration backup procedure
- Create backup scripts

### Step 3: Document Restore Procedures (3 hours)
- Component-specific restores
- Complete environment recovery
- Partial recovery scenarios
- Validation procedures

### Step 4: Create Failure Scenarios (2 hours)
- Common failure modes
- Detection methods
- Recovery steps
- Prevention measures

### Step 5: Implement Automation (4 hours)
- Write backup scripts
- Set up cron jobs
- Configure S3 buckets
- Set up monitoring/alerts

### Step 6: Test Procedures (4 hours)
- Test each backup script
- Perform test restores
- Validate data integrity
- Document any issues

### Step 7: Create Testing Schedule (1 hour)
- Monthly test plan
- Quarterly test plan
- Assign responsibilities
- Set up reminders

## Files to Create/Update

### New Files
- `docs/DISASTER_RECOVERY.md` - Main DR guide
- `scripts/backup-vault.sh` - Vault backup automation
- `scripts/backup-postgres.sh` - Database backup automation
- `scripts/backup-consul.sh` - Consul backup automation
- `scripts/restore-vault.sh` - Vault restore script
- `scripts/restore-postgres.sh` - Database restore script

### Files to Update
- `README.md` - Add link to DR documentation
- `docs/DEPLOYMENT_RUNBOOK.md` - Reference DR procedures
- `CLAUDE.md` - Update with DR script locations

## Testing Plan

1. **Backup Testing**
   - Run each backup script
   - Verify files created
   - Check S3 upload
   - Validate encryption

2. **Restore Testing**
   - Restore to test environment
   - Verify data integrity
   - Check application functionality
   - Measure restore time

3. **Failure Scenario Testing**
   - Simulate database corruption
   - Test Vault recovery
   - Test complete cluster loss
   - Document actual RTO/RPO

4. **Automation Testing**
   - Verify cron jobs run
   - Check monitoring alerts
   - Test failure notifications
   - Validate retention policies

## Success Metrics

- [ ] All backup procedures documented and tested
- [ ] All restore procedures documented and tested
- [ ] Backup scripts created and automated
- [ ] RTO/RPO targets met in testing
- [ ] Monthly DR test completed successfully
- [ ] Monitoring and alerting configured
- [ ] Peer reviewed and approved

## Dependencies

- S3 bucket for backup storage
- AWS credentials for backup scripts
- Test environment for restore validation
- Issue #1 (Deployment Runbook) for context

## Related Issues

- #1: Deployment Runbook (references DR procedures)
- #2: Troubleshooting Guide (includes recovery procedures)
- #4: Monitoring Setup (alerts on backup failures)

## Notes

- Encrypt all backups at rest and in transit
- Store recovery keys securely (separate from backups)
- Test restores regularly (monthly minimum)
- Document actual RTO/RPO from tests
- Update procedures after each test
- Consider multi-region backup strategy
- Plan for credential rotation in backups

## Definition of Done

- [ ] DR documentation created and committed
- [ ] All backup procedures documented
- [ ] All restore procedures documented
- [ ] Backup scripts created and tested
- [ ] Restore scripts created and tested
- [ ] RTO/RPO targets defined and validated
- [ ] Automation configured and running
- [ ] First DR test completed successfully
- [ ] Monitoring and alerting configured
- [ ] Peer reviewed and approved

---

**Created:** 2026-06-12
**Assignee:** TBD
**Labels:** `documentation`, `operations`, `disaster-recovery`
**Milestone:** Deployment Architecture Epic
