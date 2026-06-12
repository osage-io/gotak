# Issue #5: Update GitOps targetRevision to main Branch

**Epic:** [Deployment Architecture](deployment-architecture-epic.md)
**Phase:** 3 - GitOps Configuration (Cleanup)
**Priority:** 🟢 Low
**Estimated Effort:** 30 minutes
**Status:** Ready for Work

## Problem Statement

The GitOps applications currently have `targetRevision` pointing to a feature branch (`openshift-platform-vault-consul` or similar) instead of `main`. This was appropriate during development but should be updated to `main` now that the work is complete and merged.

## Objective

Update all Argo CD Application manifests to use `targetRevision: main` so that GitOps tracks the main branch for production deployments.

## Acceptance Criteria

- [ ] All Application manifests updated to `targetRevision: main`
- [ ] Changes committed to git
- [ ] Argo CD applications synced successfully
- [ ] No sync errors after update
- [ ] Documentation updated if needed

## Scope

### Files to Update

Based on the GitOps structure, these files likely need updates:

1. `gitops/root-app.yaml` - Root application
2. `gitops/apps/consul.yaml` - Consul application
3. `gitops/apps/vault.yaml` - Vault application
4. `gitops/apps/api-gateway.yaml` - API Gateway application
5. `gitops/apps/intentions.yaml` - Intentions application
6. `gitops/apps/gotak.yaml` - GoTAK application

### Current State
```yaml
spec:
  source:
    repoURL: https://github.com/osage-io/gotak
    targetRevision: openshift-platform-vault-consul  # ← Feature branch
    path: openshift/platform
```

### Desired State
```yaml
spec:
  source:
    repoURL: https://github.com/osage-io/gotak
    targetRevision: main  # ← Main branch
    path: openshift/platform
```

## Technical Details

### Implementation Steps

1. **Verify Feature Branch is Merged**
   ```bash
   git branch -r --merged main | grep openshift-platform-vault-consul
   ```
   - If not merged, merge the feature branch first
   - If already merged, proceed with updates

2. **Update Application Manifests**
   ```bash
   # Find all files with targetRevision
   grep -r "targetRevision:" gitops/

   # Update each file
   sed -i 's/targetRevision: openshift-platform-vault-consul/targetRevision: main/g' gitops/**/*.yaml
   ```

3. **Verify Changes**
   ```bash
   # Check the diff
   git diff gitops/

   # Ensure only targetRevision changed
   ```

4. **Commit and Push**
   ```bash
   git add gitops/
   git commit -m "chore: update GitOps targetRevision to main branch"
   git push origin main
   ```

5. **Sync Argo CD Applications**
   ```bash
   # Hard refresh all applications
   oc patch application consul -n openshift-gitops \
     --type json -p='[{"op":"remove","path":"/operation"}]'

   # Repeat for each application
   # Then sync
   argocd app sync --force consul vault api-gateway intentions gotak
   ```

6. **Verify Sync Success**
   ```bash
   # Check application status
   oc get applications -n openshift-gitops

   # Verify all are synced and healthy
   argocd app list
   ```

## Testing Plan

1. **Pre-Update Verification**
   - [ ] Note current application sync status
   - [ ] Verify all applications are healthy
   - [ ] Document current targetRevision values

2. **Update Testing**
   - [ ] Update one application first (test)
   - [ ] Verify it syncs successfully
   - [ ] Update remaining applications
   - [ ] Verify all sync successfully

3. **Post-Update Verification**
   - [ ] All applications show "Synced" status
   - [ ] No degraded services
   - [ ] Application functionality unchanged
   - [ ] Argo CD UI shows correct branch

4. **Rollback Plan**
   - If issues occur, revert the commit
   - Force sync applications back to previous state
   - Investigate and fix issues before retrying

## Files to Update

### GitOps Applications
- `gitops/root-app.yaml`
- `gitops/apps/consul.yaml`
- `gitops/apps/vault.yaml`
- `gitops/apps/api-gateway.yaml`
- `gitops/apps/intentions.yaml`
- `gitops/apps/gotak.yaml`

### Documentation (if needed)
- `gitops/README.md` - Update if it mentions the feature branch
- `docs/DEPLOYMENT_ARCHITECTURE.md` - Update if it references the branch

## Success Metrics

- [ ] All applications using `targetRevision: main`
- [ ] All applications synced successfully
- [ ] No service disruption
- [ ] Argo CD UI shows correct branch
- [ ] Future commits to main trigger automatic sync

## Dependencies

- Feature branch must be merged to main
- Access to the cluster and Argo CD
- No other pending changes in GitOps manifests

## Related Issues

- Part of the Deployment Architecture Epic
- Completes Phase 3 (GitOps Configuration)
- No blocking dependencies

## Notes

- This is a low-risk change (just updating a reference)
- Can be done during normal business hours
- No service downtime expected
- Argo CD will automatically sync after the change
- Consider doing this after other documentation issues are complete
- This enables continuous deployment from main branch

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Sync failure | Medium | Test with one app first, have rollback ready |
| Wrong branch name | Low | Verify branch name before updating |
| Uncommitted changes | Low | Check git status before updating |
| Argo CD cache | Low | Use hard refresh if needed |

## Definition of Done

- [ ] All Application manifests updated
- [ ] Changes committed and pushed to main
- [ ] All applications synced successfully
- [ ] No sync errors in Argo CD
- [ ] All services healthy and functional
- [ ] Documentation updated if needed
- [ ] Verified in Argo CD UI

---

**Created:** 2026-06-12
**Assignee:** TBD
**Labels:** `gitops`, `cleanup`, `low-priority`
**Milestone:** Deployment Architecture Epic
