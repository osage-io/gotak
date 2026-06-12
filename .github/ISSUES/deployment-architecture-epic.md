# [EPIC] Complete AWS/OpenShift/HashiCorp Deployment Architecture

## Epic Overview

Implement the complete cloud deployment architecture for GoTAK on AWS using Single-Node OpenShift (SNO) integrated with the HashiCorp stack (Vault, Consul, Boundary). This epic encompasses infrastructure as code, GitOps configuration, automation scripts, and comprehensive documentation to enable reproducible deployments of the full platform.

**Business Value:** Provides a production-ready, secure, and scalable deployment model for GoTAK that can be replicated across environments, demonstrating enterprise-grade infrastructure practices.

## Goals

1. **Infrastructure as Code**: Complete Terraform workspaces for network, compute, and supporting resources
2. **Automated Provisioning**: Ansible playbooks for SNO installation and HashiCorp stack deployment
3. **GitOps Delivery**: Argo CD-based continuous deployment of all platform and application components
4. **Security Integration**: Vault for secrets management, Consul service mesh with default-deny intentions, Boundary for privileged access
5. **Multi-Architecture Support**: ARM64-native deployment with multi-arch container builds
6. **Operational Excellence**: Comprehensive documentation, runbooks, and troubleshooting guides

## Architecture Summary

```
demoland.io (Namecheap DNS, Sectigo *.demoland.io cert)
  gotak. vault. consul. argo. ──▶ OpenShift router
  boundary. ──────────────────▶ EC2 node (Boundary PAM)

┌─ AWS us-east-2 ───────────────────────────────────────────────────────────┐
│  VPC 10.0.0.0/16 (gotak-network workspace)                                │
│  ├── EC2 t4g.medium  installer/bastion node (gotak-installer workspace)   │
│  │     └── Boundary controller+worker (podman, Ansible-managed)           │
│  └── EC2 m6g.2xlarge SNO cluster node (created by openshift-install)      │
│        OpenShift 4.22 (arm64) — namespace `gotak`:                        │
│          Consul mesh (default-deny intentions)                            │
│          Consul API Gateway ──▶ /api,/ws → gotak-server; / → gotak-web    │
│          Vault (raft, AWS KMS auto-unseal) ◀── browser via Route+CORS     │
│          postgres (PostGIS, arm64 image)                                  │
│        openshift-gitops: Argo CD app-of-apps deploys all of the above     │
└───────────────────────────────────────────────────────────────────────────┘
```

## Scope

### In Scope
- ✅ Terraform workspaces for AWS infrastructure
- ✅ Ansible automation for SNO and Boundary
- ✅ GitOps configuration (Argo CD app-of-apps)
- ✅ HashiCorp stack integration (Vault, Consul, Boundary)
- ✅ Multi-arch container builds (arm64 + amd64)
- ✅ DNS and TLS configuration
- ✅ Security hardening and mesh configuration
- ✅ Operational documentation and runbooks

### Out of Scope
- ROSA deployment (blocked by sandbox constraints)
- Multi-node OpenShift clusters
- Production-grade HA configuration (future enhancement)
- Automated disaster recovery (future enhancement)

## Components

### Phase 1: Infrastructure Foundation (Week 1-2)
**Issues to Create:**
- [ ] #TBD: Terraform workspace for VPC and networking (`iac/network`)
- [ ] #TBD: Terraform workspace for installer/bastion node (`iac/installer`)
- [ ] #TBD: Ansible playbook for SNO installation (`iac/ansible-sno`)
- [ ] #TBD: CCO manual mode credential management
- [ ] #TBD: Terraform workspace for registry pull secret (`iac/gotak-registry-secret`)

**Deliverables:**
- VPC with proper subnets, security groups, and routing
- Installer node with required tools (openshift-install, oc, ansible)
- SNO cluster running OpenShift 4.22 on arm64
- Automated credential refresh for cloud-credential-operator

### Phase 2: HashiCorp Stack Integration (Week 2-3)
**Issues to Create:**
- [ ] #TBD: Vault Helm chart deployment with KMS auto-unseal
- [ ] #TBD: Vault post-configuration (transit, PKI, CORS)
- [ ] #TBD: Consul Helm chart with service mesh
- [ ] #TBD: Consul API Gateway configuration
- [ ] #TBD: Default-deny intentions and service mesh policies
- [ ] #TBD: Boundary controller/worker Ansible deployment

**Deliverables:**
- Vault running with AWS KMS auto-unseal
- Transit encryption engine configured
- PKI engine for device certificates
- Consul service mesh with transparent proxy
- API Gateway as single ingress point
- Boundary for privileged access management

### Phase 3: GitOps Configuration (Week 3-4)
**Issues to Create:**
- [ ] #TBD: Argo CD root application (`gitops/root-app.yaml`)
- [ ] #TBD: Consul application with sync waves
- [ ] #TBD: Vault application with sync waves
- [ ] #TBD: API Gateway and intentions applications
- [ ] #TBD: GoTAK application (postgres, server, web)
- [ ] #TBD: Custom domain routes automation

**Deliverables:**
- App-of-apps pattern fully implemented
- Sync waves for proper dependency ordering
- All platform components deployed via GitOps
- Custom domain routes with wildcard TLS

### Phase 4: CI/CD & Container Registry (Week 4)
**Issues to Create:**
- [ ] #TBD: Multi-arch container builds (arm64 + amd64)
- [ ] #TBD: GitHub Actions workflow enhancements
- [ ] #TBD: GHCR integration and image publishing
- [ ] #TBD: Automated image promotion pipeline

**Deliverables:**
- Multi-arch images for gotak-server and gotak-web
- Automated builds on push to main
- Images published to ghcr.io/osage-io/gotak-*
- Version tagging and latest updates

### Phase 5: Documentation & Operations (Week 5)
**Issues to Create:**
- [ ] #TBD: Deployment runbook
- [ ] #TBD: Troubleshooting guide
- [ ] #TBD: DNS/TLS setup documentation
- [ ] #TBD: Disaster recovery procedures
- [ ] #TBD: Monitoring and observability setup

**Deliverables:**
- Step-by-step deployment guide
- Common issues and solutions
- DNS record management guide
- Backup and restore procedures
- Metrics and logging configuration

## Success Criteria

### Technical
- [ ] Complete infrastructure deployed via Terraform Cloud
- [ ] SNO cluster running OpenShift 4.22 on arm64
- [ ] All HashiCorp components operational (Vault, Consul, Boundary)
- [ ] GitOps managing all platform and application deployments
- [ ] Multi-arch containers building and deploying successfully
- [ ] Service mesh with default-deny working correctly
- [ ] All services accessible via custom domains with TLS

### Operational
- [ ] Zero-touch deployment from code
- [ ] Automated credential refresh working
- [ ] Vault auto-unseal functioning
- [ ] Argo CD syncing all applications
- [ ] Monitoring and alerting in place

### Documentation
- [ ] Complete deployment runbook
- [ ] Architecture diagrams updated
- [ ] Troubleshooting guide with known issues
- [ ] DNS/TLS configuration documented

## Dependencies

### External
- AWS account with appropriate permissions
- Terraform Cloud organization and workspaces
- GitHub repository with Actions enabled
- Domain name (demoland.io) with DNS access
- Sectigo wildcard certificate (*.demoland.io)

### Internal
- Container images for gotak-server and gotak-web
- Database migrations up to date
- Application configuration files

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| AWS sandbox constraints (no IAM users/OIDC) | High | Certain | Use CCO manual mode with instance profile + cron refresh |
| ARM64 image compatibility | High | Medium | Explicitly build multi-arch, use qualified image names |
| Webhook deadlock on Consul | Critical | Low | Set failurePolicy: Ignore, document recovery |
| Vault seal with session credentials | Medium | Medium | Refresh script + pod restart automation |
| Gateway API CRD conflicts | Medium | Low | Disable Consul CRD installation |
| Argo sync wedges | Medium | Medium | Document patch command for operation removal |

## Timeline

**Total Duration:** 5 weeks

- **Week 1-2:** Infrastructure Foundation
- **Week 2-3:** HashiCorp Stack Integration (overlaps with Week 2)
- **Week 3-4:** GitOps Configuration
- **Week 4:** CI/CD & Container Registry
- **Week 5:** Documentation & Operations

**Key Milestones:**
- End of Week 2: SNO cluster operational
- End of Week 3: HashiCorp stack fully integrated
- End of Week 4: GitOps deploying all components
- End of Week 5: Complete documentation and handoff

## Operational Pitfalls (Lessons Learned)

1. **ARM64 everywhere** - Node is Graviton; amd64-only images fail with `exec format error`
2. **Fully qualify docker.io** - OpenShift resolves bare names to Red Hat registry
3. **Webhook deadlock** - Mutating webhook with `failurePolicy: Fail` can wedge control plane
4. **Gateway API CRDs** - OpenShift owns them; Consul must not try to install
5. **Wedged Argo sync** - Remove operation field and hard-refresh
6. **Argo customization** - Must go in ArgoCD CR, not direct Route edits

## Related Documentation

- [DEPLOYMENT_ARCHITECTURE.md](../docs/DEPLOYMENT_ARCHITECTURE.md) - This epic's source
- [ARCHITECTURE.md](../docs/ARCHITECTURE.md) - Application architecture
- [DEPLOYMENT.md](../DEPLOYMENT.md) - Docker Compose deployment
- [BUILD.md](../BUILD.md) - Build instructions
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines

## Child Issues

This epic will be broken down into approximately 25-30 individual issues across the five phases. Each issue will be created with:
- Clear acceptance criteria
- Technical implementation details
- Testing requirements
- Documentation updates needed

## Labels

- `epic` - Epic-level work
- `infrastructure` - IaC and platform work
- `gitops` - GitOps configuration
- `security` - Security-related changes
- `documentation` - Documentation updates
- `ci-cd` - CI/CD pipeline work

## Assignees

TBD based on team capacity and expertise

---

**Created:** 2026-06-12
**Target Completion:** 2026-07-17 (5 weeks)
**Status:** Planning
