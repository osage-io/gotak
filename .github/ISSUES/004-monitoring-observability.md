# Issue #4: Implement Monitoring & Observability Stack

**Epic:** [Deployment Architecture](deployment-architecture-epic.md)
**Phase:** 5 - Documentation & Operations
**Priority:** 🟡 Medium
**Estimated Effort:** 2-3 days
**Status:** Ready for Work

## Problem Statement

The current deployment has no monitoring or observability stack. Operators have no visibility into system health, performance metrics, or early warning signs of issues. This makes it difficult to maintain system reliability and troubleshoot problems proactively.

## Objective

Implement a comprehensive monitoring and observability solution using Prometheus, Grafana, and Loki to provide visibility into infrastructure health, application performance, and system logs.

## Acceptance Criteria

- [ ] Prometheus deployed and scraping metrics
- [ ] Grafana deployed with dashboards
- [ ] Loki deployed for log aggregation
- [ ] Key metrics dashboards created
- [ ] Alerting rules configured
- [ ] Alert routing to appropriate channels
- [ ] Documentation for dashboard usage
- [ ] Runbook for responding to alerts

## Scope

### In Scope

#### Metrics Collection (Prometheus)
- OpenShift cluster metrics
- Node metrics (CPU, memory, disk, network)
- Pod metrics (resource usage, restarts)
- Vault metrics (seal status, requests, errors)
- Consul metrics (service health, intentions, mesh)
- PostgreSQL metrics (connections, queries, replication)
- Application metrics (GoTAK server/web)
- Custom business metrics

#### Visualization (Grafana)
- Cluster overview dashboard
- Node health dashboard
- Application performance dashboard
- Vault operations dashboard
- Consul service mesh dashboard
- Database performance dashboard
- Custom business metrics dashboard

#### Log Aggregation (Loki)
- Container logs
- Application logs
- Audit logs
- System logs
- Centralized search

#### Alerting
- Critical system alerts
- Performance degradation alerts
- Security alerts
- Capacity alerts
- Custom application alerts

### Out of Scope
- Distributed tracing (future enhancement)
- APM (Application Performance Monitoring) - future
- Cost monitoring - separate concern
- External synthetic monitoring - future

## Technical Details

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│ OpenShift Cluster (gotak namespace)                     │
│                                                          │
│  ┌──────────────┐    ┌──────────────┐                  │
│  │ Prometheus   │◄───│ ServiceMonitor│                  │
│  │ Operator     │    │ Resources     │                  │
│  └──────┬───────┘    └──────────────┘                  │
│         │                                                │
│         │ scrapes                                        │
│         ▼                                                │
│  ┌──────────────────────────────────────┐              │
│  │ Targets:                              │              │
│  │ - OpenShift metrics                   │              │
│  │ - Node exporters                      │              │
│  │ - Vault metrics                       │              │
│  │ - Consul metrics                      │              │
│  │ - PostgreSQL exporter                 │              │
│  │ - GoTAK application metrics           │              │
│  └──────────────────────────────────────┘              │
│         │                                                │
│         │ queries                                        │
│         ▼                                                │
│  ┌──────────────┐    ┌──────────────┐                  │
│  │   Grafana    │◄───│    Loki      │                  │
│  │  Dashboards  │    │ Log Storage  │                  │
│  └──────────────┘    └──────┬───────┘                  │
│         │                    │                           │
│         │                    │ collects                  │
│         │                    ▼                           │
│         │             ┌──────────────┐                  │
│         │             │  Promtail    │                  │
│         │             │  DaemonSet   │                  │
│         │             └──────────────┘                  │
│         │                                                │
│         │ alerts                                         │
│         ▼                                                │
│  ┌──────────────┐                                       │
│  │ Alertmanager │──► Slack/Email/PagerDuty             │
│  └──────────────┘                                       │
└─────────────────────────────────────────────────────────┘
```

### Components to Deploy

1. **Prometheus Stack** (via kube-prometheus-stack Helm chart)
   - Prometheus Operator
   - Prometheus server
   - Alertmanager
   - Node exporters
   - kube-state-metrics

2. **Grafana** (included in kube-prometheus-stack)
   - Pre-configured data sources
   - Custom dashboards
   - User management

3. **Loki Stack** (via loki-stack Helm chart)
   - Loki server
   - Promtail (log collector)
   - Grafana integration

4. **Exporters**
   - PostgreSQL exporter
   - Vault metrics (built-in)
   - Consul metrics (built-in)

### Key Dashboards

1. **Cluster Overview**
   - Node status and health
   - Pod count and status
   - Resource utilization
   - Network traffic
   - Storage usage

2. **Application Performance**
   - Request rate
   - Response time
   - Error rate
   - Active connections
   - WebSocket connections

3. **Vault Operations**
   - Seal status
   - Request rate
   - Token usage
   - Secret access patterns
   - Performance metrics

4. **Consul Service Mesh**
   - Service health
   - Intention denials
   - Sidecar status
   - Mesh traffic
   - API Gateway metrics

5. **Database Performance**
   - Connection pool usage
   - Query performance
   - Replication lag
   - Cache hit ratio
   - Slow queries

6. **Business Metrics**
   - Active users
   - Mission count
   - Message throughput
   - Position updates
   - API usage

### Alert Rules

#### Critical Alerts
- Node down
- Pod crash looping
- Vault sealed
- Database down
- Disk space critical (>90%)
- Memory pressure

#### Warning Alerts
- High CPU usage (>80%)
- High memory usage (>85%)
- Disk space warning (>75%)
- High error rate
- Slow response times
- Certificate expiring soon

#### Info Alerts
- Deployment events
- Scaling events
- Backup completion
- Configuration changes

## Implementation Plan

### Step 1: Deploy Prometheus Stack (4 hours)
```bash
# Add Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Create values file
cat > monitoring-values.yaml <<EOF
prometheus:
  prometheusSpec:
    retention: 15d
    storageSpec:
      volumeClaimTemplate:
        spec:
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 50Gi
grafana:
  enabled: true
  adminPassword: <secure-password>
  ingress:
    enabled: true
    hosts:
      - grafana-gotak.apps.cluster
alertmanager:
  enabled: true
  config:
    route:
      receiver: 'slack'
    receivers:
      - name: 'slack'
        slack_configs:
          - api_url: '<webhook-url>'
            channel: '#gotak-alerts'
EOF

# Install
helm install prometheus prometheus-community/kube-prometheus-stack \
  -n gotak -f monitoring-values.yaml
```

### Step 2: Deploy Loki Stack (2 hours)
```bash
# Add Helm repo
helm repo add grafana https://grafana.github.io/helm-charts

# Create values file
cat > loki-values.yaml <<EOF
loki:
  persistence:
    enabled: true
    size: 50Gi
promtail:
  enabled: true
grafana:
  enabled: false  # Using existing Grafana
EOF

# Install
helm install loki grafana/loki-stack \
  -n gotak -f loki-values.yaml
```

### Step 3: Configure ServiceMonitors (2 hours)
- Create ServiceMonitor for Vault
- Create ServiceMonitor for Consul
- Create ServiceMonitor for PostgreSQL
- Create ServiceMonitor for GoTAK

### Step 4: Create Dashboards (4 hours)
- Import community dashboards
- Customize for GoTAK
- Create custom business metrics dashboard
- Test all visualizations

### Step 5: Configure Alerts (3 hours)
- Define alert rules
- Configure Alertmanager routing
- Set up notification channels
- Test alert firing

### Step 6: Documentation (2 hours)
- Dashboard usage guide
- Alert response runbook
- Metrics reference
- Troubleshooting guide

### Step 7: Testing (2 hours)
- Verify metrics collection
- Test dashboard functionality
- Trigger test alerts
- Validate log aggregation

## Files to Create/Update

### New Files
- `openshift/monitoring/prometheus-values.yaml` - Prometheus configuration
- `openshift/monitoring/loki-values.yaml` - Loki configuration
- `openshift/monitoring/servicemonitors/` - ServiceMonitor definitions
- `openshift/monitoring/dashboards/` - Grafana dashboard JSON
- `openshift/monitoring/alerts/` - PrometheusRule definitions
- `docs/MONITORING.md` - Monitoring documentation
- `docs/ALERT_RUNBOOK.md` - Alert response procedures

### Files to Update
- `gitops/apps/monitoring.yaml` - Add monitoring app to GitOps
- `README.md` - Add monitoring section
- `CLAUDE.md` - Update with monitoring info

## Testing Plan

1. **Metrics Collection**
   - Verify Prometheus targets are up
   - Check metric cardinality
   - Validate scrape intervals
   - Test metric queries

2. **Dashboard Functionality**
   - Load each dashboard
   - Verify data displays correctly
   - Test time range selection
   - Check variable functionality

3. **Alerting**
   - Trigger test alerts
   - Verify routing works
   - Check notification delivery
   - Test alert silencing

4. **Log Aggregation**
   - Search for specific logs
   - Test log filtering
   - Verify retention policy
   - Check query performance

5. **Performance**
   - Monitor Prometheus resource usage
   - Check Grafana response times
   - Verify Loki query performance
   - Test under load

## Success Metrics

- [ ] All components deployed and healthy
- [ ] Metrics being collected from all targets
- [ ] Dashboards displaying data correctly
- [ ] Alerts firing and routing properly
- [ ] Logs aggregated and searchable
- [ ] Documentation complete
- [ ] Team trained on usage
- [ ] Performance acceptable

## Dependencies

- Sufficient cluster resources (CPU, memory, storage)
- Slack/email/PagerDuty for alerting
- Issue #1 (Deployment Runbook) for integration
- Issue #2 (Troubleshooting Guide) for alert responses

## Related Issues

- #1: Deployment Runbook (includes monitoring setup)
- #2: Troubleshooting Guide (uses metrics for debugging)
- #3: Disaster Recovery (monitors backup success)

## Notes

- Start with community dashboards, customize as needed
- Keep alert noise low - only alert on actionable items
- Document what each alert means and how to respond
- Consider retention policies for metrics and logs
- Plan for storage growth
- Use Grafana folders to organize dashboards
- Set up RBAC for Grafana users
- Consider multi-tenancy if needed

## Definition of Done

- [ ] Prometheus stack deployed and collecting metrics
- [ ] Grafana deployed with dashboards
- [ ] Loki deployed and aggregating logs
- [ ] All ServiceMonitors created and working
- [ ] Alert rules configured and tested
- [ ] Alertmanager routing configured
- [ ] Documentation complete
- [ ] Team trained on usage
- [ ] Performance validated
- [ ] Peer reviewed and approved

---

**Created:** 2026-06-12
**Assignee:** TBD
**Labels:** `monitoring`, `observability`, `operations`
**Milestone:** Deployment Architecture Epic
