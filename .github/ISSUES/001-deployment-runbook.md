# Issue #1: Create Consolidated Deployment Runbook

**Epic:** [Deployment Architecture](deployment-architecture-epic.md)
**Phase:** 5 - Documentation & Operations
**Priority:** 🔴 High
**Estimated Effort:** 1-2 days
**Status:** Ready for Work

## Problem Statement

The deployment process is currently documented across multiple README files (`iac/README.md`, `gitops/README.md`, `openshift/platform/README.md`), making it difficult for new operators to understand the complete end-to-end deployment flow. We need a single, authoritative deployment guide.

## Objective

Create a comprehensive, step-by-step deployment runbook that consolidates all deployment knowledge into a single document, enabling any operator to deploy the complete GoTAK stack from scratch.

## Acceptance Criteria

- [ ] Single markdown document covering complete deployment
- [ ] Prerequisites checklist (tools, credentials, access)
- [ ] Step-by-step instructions with commands
- [ ] Validation steps after each major phase
- [ ] Troubleshooting section for common issues
- [ ] Estimated time for each phase
- [ ] Links to detailed component documentation
- [ ] Tested by at least one other person

## Scope

### In Scope
- Complete deployment from AWS account to running application
- All five phases: Network → Installer → SNO → Platform → Application
- Terraform Cloud workspace setup
- Ansible playbook execution
- GitOps activation
- DNS and TLS configuration
- Validation procedures

### Out of Scope
- Detailed troubleshooting (separate issue #2)
- Disaster recovery (separate issue #3)
- Monitoring setup (separate issue #4)
- Development environment setup (covered in existing docs)

## Technical Details

### Document Structure

```markdown
# GoTAK Deployment Runbook

## Overview
- Architecture diagram
- Deployment phases
- Total time estimate (~4-6 hours)

## Prerequisites
- [ ] AWS account with appropriate permissions
- [ ] Terraform Cloud organization
- [ ] GitHub access
- [ ] Domain name and DNS access
- [ ] TLS certificate
- [ ] Required tools installed

## Phase 1: Network Infrastructure (30 min)
- Terraform Cloud workspace setup
- VPC deployment
- Validation steps

## Phase 2: Installer Node (20 min)
- Bastion node deployment
- Tool installation verification
- SSH access validation

## Phase 3: OpenShift SNO (2-3 hours)
- Ansible inventory configuration
- SNO installation
- CCO manual mode setup
- Cluster validation

## Phase 4: HashiCorp Platform (1 hour)
- KMS key creation
- Vault deployment and initialization
- Consul deployment
- Boundary setup
- Service mesh validation

## Phase 5: GitOps & Application (30 min)
- Argo CD installation
- Root app deployment
- Application sync
- Custom domain configuration
- End-to-end validation

## Post-Deployment
- Health checks
- Smoke tests
- Credential management
- Next steps

## Quick Reference
- Common commands
- Important URLs
- Credential locations
```

### Key Sections to Include

1. **Prerequisites Checklist**
   - AWS permissions required
   - Terraform Cloud setup
   - Tool versions (terraform, ansible, oc, helm, aws cli)
   - Environment variables needed
   - Credentials and secrets

2. **Phase-by-Phase Instructions**
   - Clear command sequences
   - Expected output examples
   - Validation commands
   - Time estimates
   - "What's happening" explanations

3. **Validation Procedures**
   - How to verify each phase succeeded
   - Health check commands
   - What to look for in logs
   - Common success indicators

4. **Quick Troubleshooting**
   - Top 5 issues and quick fixes
   - Link to detailed troubleshooting guide
   - Where to find logs
   - How to rollback

## Implementation Plan

### Step 1: Gather Existing Documentation (1 hour)
- Review all README files
- Extract deployment commands
- Note validation steps
- Identify gaps

### Step 2: Create Document Outline (30 min)
- Structure the runbook
- Define sections
- Plan command flow

### Step 3: Write Phase Instructions (4-6 hours)
- Phase 1: Network (from `iac/network/README.md`)
- Phase 2: Installer (from `iac/installer/README.md`)
- Phase 3: SNO (from `iac/ansible-sno/README.md`)
- Phase 4: Platform (from `openshift/platform/README.md`)
- Phase 5: GitOps (from `gitops/README.md`)

### Step 4: Add Validation & Troubleshooting (2 hours)
- Validation commands for each phase
- Common issues and quick fixes
- Links to detailed docs

### Step 5: Test & Refine (2-4 hours)
- Follow the runbook on a fresh deployment
- Note any missing steps
- Refine commands and explanations
- Add screenshots if helpful

### Step 6: Peer Review (1 hour)
- Have another team member review
- Test with someone unfamiliar with the process
- Incorporate feedback

## Files to Create/Update

### New Files
- `docs/DEPLOYMENT_RUNBOOK.md` - The main runbook

### Files to Update
- `README.md` - Add link to deployment runbook
- `docs/DEPLOYMENT_ARCHITECTURE.md` - Reference the runbook
- `CLAUDE.md` - Update with runbook location

## Testing Plan

1. **Dry Run**: Walk through the runbook without executing
2. **Fresh Deployment**: Follow the runbook on a new AWS account
3. **Peer Test**: Have another operator follow it independently
4. **Time Validation**: Verify time estimates are accurate
5. **Validation Steps**: Ensure all validation commands work

## Success Metrics

- [ ] Complete deployment possible following only the runbook
- [ ] No missing steps or commands
- [ ] All validation procedures work
- [ ] Time estimates within 20% of actual
- [ ] Peer reviewer successfully deploys using runbook
- [ ] Zero critical gaps identified in testing

## Dependencies

- Access to all existing README files
- Working deployment environment for testing
- Peer reviewer availability

## Related Issues

- #2: Troubleshooting Guide (references this runbook)
- #3: Disaster Recovery Procedures (builds on this runbook)
- #4: Monitoring & Observability Setup (follows this runbook)

## Notes

- Keep commands copy-pasteable
- Use consistent formatting
- Include "why" explanations, not just "how"
- Add warnings for destructive operations
- Note sandbox-specific constraints
- Include rollback procedures for each phase

## Definition of Done

- [ ] Runbook document created and committed
- [ ] All phases documented with commands
- [ ] Validation steps included
- [ ] Tested on fresh deployment
- [ ] Peer reviewed and approved
- [ ] README.md updated with link
- [ ] Time estimates validated
- [ ] No critical gaps or errors

---

**Created:** 2026-06-12
**Assignee:** TBD
**Labels:** `documentation`, `high-priority`, `deployment`
**Milestone:** Deployment Architecture Epic
