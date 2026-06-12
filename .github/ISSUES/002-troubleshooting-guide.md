# Issue #2: Create Consolidated Troubleshooting Guide

**Epic:** [Deployment Architecture](deployment-architecture-epic.md)
**Phase:** 5 - Documentation & Operations
**Priority:** 🔴 High
**Estimated Effort:** 1 day
**Status:** Ready for Work

## Problem Statement

Troubleshooting information is scattered across multiple README files and the DEPLOYMENT_ARCHITECTURE.md "Operational Pitfalls" section. When issues occur, operators must search through multiple documents to find solutions. We need a single, comprehensive troubleshooting guide.

## Objective

Create a consolidated troubleshooting guide that documents all known issues, their symptoms, root causes, and solutions in a searchable, well-organized format.

## Acceptance Criteria

- [ ] Single markdown document with all troubleshooting content
- [ ] Issues organized by component/phase
- [ ] Each issue includes: symptoms, cause, solution, prevention
- [ ] Quick reference section for common issues
- [ ] Debugging commands and log locations
- [ ] Recovery procedures for critical failures
- [ ] Links to related documentation
- [ ] Searchable format (good headings, keywords)

## Scope

### In Scope
- All known issues from existing documentation
- Component-specific troubleshooting (Terraform, Ansible, OpenShift, Vault, Consul, Boundary)
- Common deployment failures
- Runtime issues
- Performance problems
- Security/credential issues
- Recovery procedures

### Out of Scope
- Application-level debugging (separate from infrastructure)
- Development environment issues (covered elsewhere)
- Feature requests or enhancements

## Known Issues to Document

### From DEPLOYMENT_ARCHITECTURE.md
1. **ARM64 image compatibility** - `exec format error`
2. **Bare image names** - OpenShift resolves to wrong registry
3. **Webhook deadlock** - Consul injector wedges control plane
4. **Gateway API CRD conflicts** - Ownership issues
5. **Wedged Argo sync** - Stuck on old revision
6. **Argo customization** - Direct edits get reverted

### From openshift/platform/README.md
1. **KMS auto-unseal with rotating creds** - Vault can't unseal after cred expiry
2. **Session credential expiration** - AWS STS creds expire
3. **Vault initialization** - First-time setup issues

### From iac/ansible-sno/README.md
1. **CCO manual mode** - Credential refresh failures
2. **SNO installation timeouts** - Long-running operations
3. **Bootstrap failures** - Cluster won't come up

### From gitops/README.md
1. **CRD timing issues** - Apps fail before CRDs exist
2. **Injector not ready** - Pods start before sidecar injection available
3. **Sync wave ordering** - Dependencies not met

## Technical Details

### Document Structure

```markdown
# GoTAK Troubleshooting Guide

## Quick Reference
- Top 10 most common issues
- Emergency recovery procedures
- Where to find logs
- Support contacts

## General Debugging
- Log locations by component
- Useful commands
- Debug mode activation
- Health check procedures

## Infrastructure Issues
### Terraform
- Workspace state issues
- Provider authentication
- Resource conflicts
- State locking

### AWS
- Permission errors
- Quota limits
- Network connectivity
- KMS key issues

### Ansible
- Connection failures
- Task timeouts
- Variable errors
- Idempotency issues

## OpenShift Issues
### Installation
- Bootstrap failures
- Timeout errors
- Certificate issues
- DNS problems

### Runtime
- Node issues
- Pod failures
- Network problems
- Storage issues

## HashiCorp Stack Issues
### Vault
- Seal/unseal problems
- KMS auto-unseal failures
- Token expiration
- Storage backend issues
- CORS configuration

### Consul
- Service mesh connectivity
- Intention denials
- Webhook deadlocks
- CRD conflicts
- Sidecar injection failures

### Boundary
- Controller/worker connectivity
- Target registration
- Session issues

## GitOps Issues
### Argo CD
- Sync failures
- Wedged operations
- CRD timing
- Wave ordering
- Customization reverts

## Application Issues
### Database
- Connection failures
- Migration errors
- Performance problems

### Containers
- Image pull failures
- Architecture mismatches
- Startup failures
- Health check failures

## Recovery Procedures
- Vault recovery
- Consul recovery
- Argo CD recovery
- Database recovery
- Complete cluster recovery

## Appendix
- Log locations reference
- Command cheat sheet
- Port reference
- Credential locations
```

### Issue Template Format

For each issue, use this structure:

```markdown
### Issue: [Short Description]

**Symptoms:**
- What the user sees/experiences
- Error messages
- Failed health checks

**Root Cause:**
- Why this happens
- Common triggers
- Related components

**Solution:**
1. Step-by-step fix
2. Commands to run
3. Validation steps

**Prevention:**
- How to avoid this issue
- Configuration changes
- Best practices

**Related Issues:**
- Links to similar problems
- Dependencies

**References:**
- Documentation links
- GitHub issues
- External resources
```

## Implementation Plan

### Step 1: Gather Known Issues (2 hours)
- Extract from DEPLOYMENT_ARCHITECTURE.md
- Review all README files
- Check openshift/platform scripts
- Review iac/ documentation
- Note any GitHub issues

### Step 2: Organize by Category (1 hour)
- Group by component
- Prioritize by frequency/severity
- Create quick reference list
- Identify recovery procedures

### Step 3: Document Each Issue (4-5 hours)
- Use consistent template
- Add symptoms, cause, solution
- Include commands and examples
- Add prevention tips
- Link to related docs

### Step 4: Add Debugging Section (2 hours)
- Log locations
- Useful commands
- Debug mode instructions
- Health check procedures

### Step 5: Create Recovery Procedures (2 hours)
- Critical component recovery
- Data recovery
- Rollback procedures
- Emergency contacts

### Step 6: Test & Validate (2 hours)
- Verify all commands work
- Test recovery procedures
- Check all links
- Peer review

## Files to Create/Update

### New Files
- `docs/TROUBLESHOOTING.md` - The main guide

### Files to Update
- `README.md` - Add link to troubleshooting guide
- `docs/DEPLOYMENT_RUNBOOK.md` - Reference troubleshooting guide
- `CLAUDE.md` - Update with troubleshooting location

## Testing Plan

1. **Command Verification**: Test all debugging commands
2. **Recovery Testing**: Validate recovery procedures work
3. **Link Checking**: Ensure all references are valid
4. **Peer Review**: Have operators review for completeness
5. **Real-World Test**: Use during actual troubleshooting

## Success Metrics

- [ ] All known issues documented
- [ ] Each issue has complete information
- [ ] Recovery procedures tested
- [ ] Commands verified to work
- [ ] Peer reviewed and approved
- [ ] Used successfully in real troubleshooting
- [ ] Reduced time to resolve common issues

## Dependencies

- Issue #1 (Deployment Runbook) for context
- Access to all environments for testing
- Peer reviewers with operational experience

## Related Issues

- #1: Deployment Runbook (references this guide)
- #3: Disaster Recovery Procedures (uses recovery procedures)
- #4: Monitoring Setup (helps prevent issues)

## Notes

- Keep solutions concise but complete
- Include "why" explanations
- Add warnings for destructive operations
- Note version-specific issues
- Include workarounds when no fix exists
- Add timestamps for time-sensitive issues

## Definition of Done

- [ ] Troubleshooting guide created and committed
- [ ] All known issues documented
- [ ] Recovery procedures included
- [ ] Commands tested and verified
- [ ] Peer reviewed and approved
- [ ] README.md updated with link
- [ ] Successfully used in real troubleshooting
- [ ] No critical gaps identified

---

**Created:** 2026-06-12
**Assignee:** TBD
**Labels:** `documentation`, `high-priority`, `operations`
**Milestone:** Deployment Architecture Epic
