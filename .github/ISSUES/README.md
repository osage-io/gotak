# GoTAK GitHub Issues

This directory contains issue specifications for the GoTAK project. Issues are organized by epic and priority.

## Active Epics

### [Deployment Architecture Epic](deployment-architecture-epic.md)
**Status:** 75% Complete | **Priority:** High | **Timeline:** 1 week remaining

Complete AWS/OpenShift/HashiCorp deployment architecture with infrastructure as code, GitOps, and operational documentation.

**Current Status:** [View Status Assessment](deployment-architecture-status.md)

#### Remaining Issues

| # | Issue | Priority | Effort | Status |
|---|-------|----------|--------|--------|
| [#1](001-deployment-runbook.md) | Create Consolidated Deployment Runbook | 🔴 High | 1-2 days | Ready |
| [#2](002-troubleshooting-guide.md) | Create Consolidated Troubleshooting Guide | 🔴 High | 1 day | Ready |
| [#3](003-disaster-recovery.md) | Document Disaster Recovery Procedures | 🟡 Medium | 1 day | Ready |
| [#4](004-monitoring-observability.md) | Implement Monitoring & Observability Stack | 🟡 Medium | 2-3 days | Ready |
| [#5](005-gitops-target-revision.md) | Update GitOps targetRevision to main | 🟢 Low | 30 min | Ready |

#### Completed Work (75%)

✅ **Phase 1: Infrastructure Foundation (100%)**
- VPC/Network Terraform
- Installer Node Terraform
- SNO Installation Ansible
- CCO Manual Mode
- Registry Secret Terraform

✅ **Phase 2: HashiCorp Stack Integration (100%)**
- Vault Helm Deployment
- Vault Post-Configuration
- Consul Helm Deployment
- API Gateway Configuration
- Service Mesh Intentions
- Boundary Deployment

✅ **Phase 3: GitOps Configuration (95%)**
- Argo CD Root App
- All Child Applications
- Custom Domain Routes
- API Named Cert
- Argo Custom Domain

✅ **Phase 4: CI/CD & Container Registry (100%)**
- Multi-arch Container Builds
- GitHub Actions Workflows
- GHCR Integration
- Image Tagging

⚠️ **Phase 5: Documentation & Operations (40%)**
- Architecture documentation ✅
- Platform READMEs ✅
- Deployment runbook ❌ (Issue #1)
- Troubleshooting guide ❌ (Issue #2)
- Disaster recovery ❌ (Issue #3)
- Monitoring setup ❌ (Issue #4)
- GitOps cleanup ❌ (Issue #5)

## Issue Workflow

### Creating Issues
1. Use the appropriate template from `.github/ISSUE_TEMPLATE/`
2. Fill in all required sections
3. Add appropriate labels and milestone
4. Link to related issues and epic

### Working on Issues
1. Assign yourself to the issue
2. Create a feature branch: `git checkout -b issue-N-description`
3. Make changes following the implementation plan
4. Test according to the testing plan
5. Update documentation as specified
6. Create PR referencing the issue: `closes #N`

### Closing Issues
Issues are automatically closed when:
- PR is merged with `closes #N` in commit message
- All acceptance criteria are met
- Definition of done is satisfied

## Priority Levels

- 🔴 **High**: Critical for epic completion, blocks other work
- 🟡 **Medium**: Important but not blocking
- 🟢 **Low**: Nice to have, can be deferred

## Effort Estimates

- **Small**: < 1 day
- **Medium**: 1-2 days
- **Large**: 3-5 days
- **Epic**: Multiple weeks, needs breakdown

## Labels

- `epic` - Epic-level work
- `infrastructure` - IaC and platform work
- `gitops` - GitOps configuration
- `security` - Security-related changes
- `documentation` - Documentation updates
- `ci-cd` - CI/CD pipeline work
- `monitoring` - Monitoring and observability
- `operations` - Operational procedures
- `disaster-recovery` - DR and backup procedures
- `high-priority` - Must be completed soon
- `cleanup` - Technical debt or cleanup work

## Quick Start

### For New Contributors
1. Read the [Deployment Architecture Epic](deployment-architecture-epic.md)
2. Review the [Status Assessment](deployment-architecture-status.md)
3. Pick an issue marked "Ready for Work"
4. Follow the implementation plan in the issue
5. Submit PR when complete

### For Reviewers
1. Verify all acceptance criteria are met
2. Check that testing plan was followed
3. Ensure documentation is updated
4. Validate definition of done
5. Approve and merge

## Related Documentation

- [CLAUDE.md](../../CLAUDE.md) - AI assistant context
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Contribution guidelines
- [docs/DEPLOYMENT_ARCHITECTURE.md](../../docs/DEPLOYMENT_ARCHITECTURE.md) - Architecture overview
- [docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md) - Application architecture

## Contact

For questions about issues or the epic:
- Review the issue documentation
- Check related documentation
- Ask in project discussions
- Contact the epic owner

---

**Last Updated:** 2026-06-12
**Total Issues:** 5 remaining
**Epic Completion:** 75%
**Estimated Time to Complete:** 1 week
