# Sprint 12: Production Deployment & DevOps

**Duration:** 2 weeks  
**Theme:** Production Readiness & Operational Excellence  
**Sprint Goals:** Complete production deployment pipeline and operational procedures for enterprise delivery

## Objectives

1. **CI/CD Pipeline**: Automated build, test, and deployment pipeline
2. **Container Orchestration**: Kubernetes deployment and management
3. **Infrastructure as Code**: Terraform provisioning and configuration management
4. **Production Monitoring**: Comprehensive observability and alerting in production
5. **Disaster Recovery**: Backup, restore, and business continuity procedures

## User Stories

### Epic: Production Operations Platform

**US-12.1: Automated Deployment Pipeline**
```
As a DevOps engineer
I want a fully automated CI/CD pipeline for safe production deployments
So that we can deliver updates quickly and reliably without manual intervention
```

**Acceptance Criteria:**
- Git-based deployment triggers with branch protection
- Automated testing at multiple stages (unit, integration, security)
- Blue-green deployment capability for zero-downtime updates
- Rollback mechanisms and deployment approval gates
- Artifact management and vulnerability scanning

**US-12.2: Container Orchestration Platform**
```
As a platform engineer
I want the application deployed on Kubernetes with auto-scaling
So that the system can handle variable load and maintain high availability
```

**Acceptance Criteria:**
- Kubernetes manifests for all application components
- Horizontal and vertical pod auto-scaling
- Service mesh for inter-service communication
- Ingress controllers and load balancing
- Resource limits and quality of service policies

**US-12.3: Infrastructure as Code**
```
As an infrastructure engineer  
I want all infrastructure defined as code for consistency
So that environments can be reproduced and managed programmatically
```

**Acceptance Criteria:**
- Terraform modules for all cloud resources
- Environment-specific configuration management
- State management and remote backends
- Drift detection and automated remediation
- Cost optimization and resource tagging

**US-12.4: Production Observability**
```
As a site reliability engineer
I want comprehensive monitoring and alerting for production systems
So that I can detect and resolve issues before they impact users
```

**Acceptance Criteria:**
- Multi-layer monitoring (infrastructure, application, business)
- Distributed tracing for request flows
- Log aggregation and analysis
- Performance profiling and optimization
- Incident response automation

**US-12.5: Disaster Recovery & Business Continuity**
```
As a business continuity manager
I want robust backup and disaster recovery procedures
So that critical operations can continue during system failures
```

**Acceptance Criteria:**
- Automated backup procedures with point-in-time recovery
- Cross-region replication for geographic redundancy
- Disaster recovery playbooks and procedures
- Regular DR testing and validation
- Recovery time objective (RTO) and recovery point objective (RPO) compliance

## Technical Implementation

### CI/CD Pipeline

**GitHub Actions Workflow**
```yaml
# .github/workflows/ci-cd.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]
  release:
    types: [ published ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}
  TERRAFORM_VERSION: 1.6.0
  KUBECTL_VERSION: v1.28.0

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: gotak_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run tests
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/gotak_test?sslmode=disable
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          flags: unittests
          name: codecov-umbrella

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Gosec Security Scanner
        uses: securecodewarrior/github-action-gosec@master
        with:
          args: '-fmt sarif -out gosec.sarif ./...'
      
      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: gosec.sarif

  build:
    needs: [test, lint, security]
    runs-on: ubuntu-latest
    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}
      image-digest: ${{ steps.build.outputs.digest }}
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix=git-
      
      - name: Build and push
        id: build
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  vulnerability-scan:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ needs.build.outputs.image-tag }}
          format: 'sarif'
          output: 'trivy-results.sarif'
      
      - name: Upload Trivy scan results
        uses: github/codeql-action/upload-sarif@v2
        if: always()
        with:
          sarif_file: 'trivy-results.sarif'

  deploy-staging:
    needs: [build, vulnerability-scan]
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'
    environment: staging
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.KUBE_CONFIG_STAGING }}
      
      - name: Deploy to staging
        run: |
          envsubst < k8s/staging/deployment.yaml | kubectl apply -f -
          kubectl rollout status deployment/gotak-server -n staging
        env:
          IMAGE_TAG: ${{ needs.build.outputs.image-tag }}
          DATABASE_URL: ${{ secrets.DATABASE_URL_STAGING }}
      
      - name: Run integration tests
        run: |
          kubectl wait --for=condition=ready pod -l app=gotak-server -n staging --timeout=300s
          go test -tags=integration ./tests/integration/...
        env:
          TEST_BASE_URL: https://staging.gotak.example.com

  deploy-production:
    needs: [build, vulnerability-scan]
    runs-on: ubuntu-latest
    if: github.event_name == 'release'
    environment: production
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.KUBE_CONFIG_PRODUCTION }}
      
      - name: Blue-Green Deployment
        run: |
          # Deploy to green environment
          envsubst < k8s/production/deployment-green.yaml | kubectl apply -f -
          kubectl rollout status deployment/gotak-server-green -n production
          
          # Run health checks
          kubectl wait --for=condition=ready pod -l app=gotak-server,version=green -n production --timeout=600s
          
          # Switch traffic to green
          kubectl patch service gotak-service -n production -p '{"spec":{"selector":{"version":"green"}}}'
          
          # Wait and cleanup blue deployment
          sleep 60
          kubectl delete deployment gotak-server-blue -n production --ignore-not-found=true
          
          # Rename green to blue for next deployment
          kubectl patch deployment gotak-server-green -n production -p '{"metadata":{"name":"gotak-server-blue"},"spec":{"selector":{"matchLabels":{"version":"blue"}},"template":{"metadata":{"labels":{"version":"blue"}}}}}'
        env:
          IMAGE_TAG: ${{ needs.build.outputs.image-tag }}
          DATABASE_URL: ${{ secrets.DATABASE_URL_PRODUCTION }}
```

**Dockerfile**
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o gotak-server \
    ./cmd/gotak-server

# Final stage
FROM scratch

# Import ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/gotak-server /gotak-server

# Copy configuration files
COPY --from=builder /app/config /config

# Create non-root user
USER 65534:65534

EXPOSE 8087 8089 8080

ENTRYPOINT ["/gotak-server"]
```

### Kubernetes Manifests

**Deployment Configuration**
```yaml
# k8s/production/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gotak-server
  namespace: production
  labels:
    app: gotak-server
    version: blue
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: gotak-server
      version: blue
  template:
    metadata:
      labels:
        app: gotak-server
        version: blue
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: gotak-server
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        runAsGroup: 65534
        fsGroup: 65534
      containers:
      - name: gotak-server
        image: ghcr.io/dfedick/gotak:${IMAGE_TAG}
        imagePullPolicy: Always
        ports:
        - name: tak-tcp
          containerPort: 8087
          protocol: TCP
        - name: tak-tls
          containerPort: 8089
          protocol: TCP
        - name: http
          containerPort: 8080
          protocol: TCP
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: gotak-secrets
              key: database-url
        - name: TLS_CERT_PATH
          value: "/certs/tls.crt"
        - name: TLS_KEY_PATH
          value: "/certs/tls.key"
        - name: CONFIG_PATH
          value: "/config/server.yaml"
        - name: LOG_LEVEL
          value: "info"
        - name: METRICS_ENABLED
          value: "true"
        volumeMounts:
        - name: config
          mountPath: /config
          readOnly: true
        - name: tls-certs
          mountPath: /certs
          readOnly: true
        - name: tmp
          mountPath: /tmp
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
      volumes:
      - name: config
        configMap:
          name: gotak-config
      - name: tls-certs
        secret:
          secretName: gotak-tls
      - name: tmp
        emptyDir: {}
      nodeSelector:
        kubernetes.io/arch: amd64
      tolerations:
      - key: node.kubernetes.io/not-ready
        operator: Exists
        effect: NoExecute
        tolerationSeconds: 300
      - key: node.kubernetes.io/unreachable
        operator: Exists
        effect: NoExecute
        tolerationSeconds: 300
      topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: kubernetes.io/hostname
        whenUnsatisfiable: DoNotSchedule
        labelSelector:
          matchLabels:
            app: gotak-server

---
apiVersion: v1
kind: Service
metadata:
  name: gotak-service
  namespace: production
  labels:
    app: gotak-server
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
    service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled: "true"
spec:
  type: LoadBalancer
  ports:
  - name: tak-tcp
    port: 8087
    targetPort: tak-tcp
    protocol: TCP
  - name: tak-tls
    port: 8089
    targetPort: tak-tls
    protocol: TCP
  - name: http
    port: 80
    targetPort: http
    protocol: TCP
  selector:
    app: gotak-server

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: gotak-config
  namespace: production
data:
  server.yaml: |
    server:
      tcp_port: 8087
      tls_port: 8089
      http_port: 8080
      tls_cert_file: /certs/tls.crt
      tls_key_file: /certs/tls.key
      read_timeout: 30s
      write_timeout: 30s
      idle_timeout: 60s
      max_header_bytes: 1048576
    
    database:
      max_open_connections: 25
      max_idle_connections: 10
      connection_max_lifetime: 5m
      connection_max_idle_time: 2m
    
    tak:
      heartbeat_interval: 30s
      max_message_size: 1048576
      allow_anonymous: false
    
    federation:
      enabled: true
      max_connections: 100
      heartbeat_interval: 30s
    
    metrics:
      enabled: true
      port: 8080
      path: /metrics
    
    logging:
      level: info
      format: json

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gotak-server-hpa
  namespace: production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gotak-server
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
      - type: Pods
        value: 2
        periodSeconds: 60
      selectPolicy: Max
```

### Infrastructure as Code

**Terraform Main Configuration**
```hcl
# terraform/main.tf
terraform {
  required_version = ">= 1.6"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
  }
  
  backend "s3" {
    bucket         = "gotak-terraform-state"
    key            = "production/terraform.tfstate"
    region         = "us-west-2"
    encrypt        = true
    dynamodb_table = "terraform-lock"
  }
}

provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = {
      Project     = "gotak"
      Environment = var.environment
      ManagedBy   = "terraform"
      Owner       = var.owner
    }
  }
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_caller_identity" "current" {}

# Local values
locals {
  cluster_name = "${var.project_name}-${var.environment}"
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
    Owner       = var.owner
  }
}

# VPC Module
module "vpc" {
  source = "./modules/vpc"
  
  project_name        = var.project_name
  environment         = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = data.aws_availability_zones.available.names
  
  tags = local.common_tags
}

# EKS Module
module "eks" {
  source = "./modules/eks"
  
  cluster_name           = local.cluster_name
  cluster_version        = var.eks_cluster_version
  vpc_id                = module.vpc.vpc_id
  subnet_ids            = module.vpc.private_subnet_ids
  control_plane_subnet_ids = module.vpc.public_subnet_ids
  
  node_groups = {
    main = {
      desired_size    = 3
      max_size       = 10
      min_size       = 2
      instance_types = ["t3.large"]
      capacity_type  = "ON_DEMAND"
      
      k8s_labels = {
        Environment = var.environment
        NodeGroup   = "main"
      }
    }
    
    spot = {
      desired_size    = 2
      max_size       = 8
      min_size       = 0
      instance_types = ["t3.large", "t3a.large", "m5.large", "m5a.large"]
      capacity_type  = "SPOT"
      
      k8s_labels = {
        Environment = var.environment
        NodeGroup   = "spot"
      }
      
      taints = [{
        key    = "spot"
        value  = "true"
        effect = "NO_SCHEDULE"
      }]
    }
  }
  
  tags = local.common_tags
}

# RDS Module
module "rds" {
  source = "./modules/rds"
  
  identifier              = "${local.cluster_name}-postgres"
  engine_version         = var.postgres_version
  instance_class         = var.rds_instance_class
  allocated_storage      = var.rds_allocated_storage
  max_allocated_storage  = var.rds_max_allocated_storage
  storage_encrypted      = true
  
  db_name  = "gotak"
  username = "gotak"
  
  vpc_id                = module.vpc.vpc_id
  subnet_ids           = module.vpc.private_subnet_ids
  allowed_cidr_blocks  = [var.vpc_cidr]
  
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"
  
  monitoring_interval = 60
  performance_insights_enabled = true
  
  tags = local.common_tags
}

# Redis Module
module "redis" {
  source = "./modules/redis"
  
  cluster_id              = "${local.cluster_name}-redis"
  node_type              = var.redis_node_type
  num_cache_nodes        = var.redis_num_nodes
  parameter_group_name   = "default.redis7"
  port                   = 6379
  
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnet_ids
  
  tags = local.common_tags
}

# S3 Module for artifacts
module "s3" {
  source = "./modules/s3"
  
  bucket_name = "${local.cluster_name}-artifacts"
  
  versioning_enabled = true
  lifecycle_rules = [
    {
      id     = "delete_old_versions"
      status = "Enabled"
      noncurrent_version_expiration = {
        days = 30
      }
    }
  ]
  
  tags = local.common_tags
}

# Outputs
output "cluster_name" {
  description = "Name of the EKS cluster"
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = module.eks.cluster_endpoint
}

output "cluster_security_group_id" {
  description = "Security group ID attached to the EKS cluster"
  value       = module.eks.cluster_security_group_id
}

output "database_endpoint" {
  description = "RDS instance endpoint"
  value       = module.rds.endpoint
}

output "redis_endpoint" {
  description = "Redis cluster endpoint"
  value       = module.redis.endpoint
}
```

**EKS Module**
```hcl
# terraform/modules/eks/main.tf
resource "aws_eks_cluster" "main" {
  name     = var.cluster_name
  version  = var.cluster_version
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids              = var.subnet_ids
    endpoint_private_access = true
    endpoint_public_access  = true
    public_access_cidrs    = ["0.0.0.0/0"]
    security_group_ids     = [aws_security_group.cluster.id]
  }

  encryption_config {
    provider {
      key_arn = aws_kms_key.eks.arn
    }
    resources = ["secrets"]
  }

  enabled_cluster_log_types = ["api", "audit", "authenticator", "controllerManager", "scheduler"]

  depends_on = [
    aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy,
    aws_cloudwatch_log_group.eks,
  ]

  tags = var.tags
}

resource "aws_eks_node_group" "main" {
  for_each = var.node_groups

  cluster_name    = aws_eks_cluster.main.name
  node_group_name = each.key
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = var.subnet_ids

  capacity_type  = each.value.capacity_type
  instance_types = each.value.instance_types

  scaling_config {
    desired_size = each.value.desired_size
    max_size     = each.value.max_size
    min_size     = each.value.min_size
  }

  update_config {
    max_unavailable_percentage = 25
  }

  labels = each.value.k8s_labels

  dynamic "taint" {
    for_each = lookup(each.value, "taints", [])
    content {
      key    = taint.value.key
      value  = taint.value.value
      effect = taint.value.effect
    }
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Node Group handling.
  depends_on = [
    aws_iam_role_policy_attachment.node_AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.node_AmazonEKS_CNI_Policy,
    aws_iam_role_policy_attachment.node_AmazonEC2ContainerRegistryReadOnly,
  ]

  tags = var.tags

  lifecycle {
    ignore_changes = [scaling_config[0].desired_size]
  }
}

# EKS Addons
resource "aws_eks_addon" "addons" {
  for_each = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
    }
    aws-ebs-csi-driver = {
      most_recent = true
    }
  }

  cluster_name             = aws_eks_cluster.main.name
  addon_name               = each.key
  addon_version            = each.value.most_recent ? data.aws_eks_addon_version.latest[each.key].version : each.value.version
  resolve_conflicts        = "OVERWRITE"
  service_account_role_arn = each.key == "aws-ebs-csi-driver" ? aws_iam_role.ebs_csi.arn : null

  depends_on = [aws_eks_node_group.main]
}

data "aws_eks_addon_version" "latest" {
  for_each = {
    coredns            = {}
    kube-proxy         = {}
    vpc-cni            = {}
    aws-ebs-csi-driver = {}
  }

  addon_name         = each.key
  kubernetes_version = aws_eks_cluster.main.version
  most_recent        = true
}
```

### Production Monitoring

**Prometheus Configuration**
```yaml
# k8s/monitoring/prometheus.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: monitoring
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
      external_labels:
        cluster: 'gotak-production'
        region: 'us-west-2'
    
    rule_files:
      - '/etc/prometheus/rules/*.yml'
    
    alerting:
      alertmanagers:
        - static_configs:
            - targets:
              - alertmanager:9093
    
    scrape_configs:
      - job_name: 'kubernetes-apiservers'
        kubernetes_sd_configs:
        - role: endpoints
        scheme: https
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
        relabel_configs:
        - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
          action: keep
          regex: default;kubernetes;https
      
      - job_name: 'kubernetes-nodes'
        scheme: https
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
        kubernetes_sd_configs:
        - role: node
        relabel_configs:
        - action: labelmap
          regex: __meta_kubernetes_node_label_(.+)
        - target_label: __address__
          replacement: kubernetes.default.svc:443
        - source_labels: [__meta_kubernetes_node_name]
          regex: (.+)
          target_label: __metrics_path__
          replacement: /api/v1/nodes/${1}/proxy/metrics
      
      - job_name: 'kubernetes-cadvisor'
        scheme: https
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
        kubernetes_sd_configs:
        - role: node
        relabel_configs:
        - action: labelmap
          regex: __meta_kubernetes_node_label_(.+)
        - target_label: __address__
          replacement: kubernetes.default.svc:443
        - source_labels: [__meta_kubernetes_node_name]
          regex: (.+)
          target_label: __metrics_path__
          replacement: /api/v1/nodes/${1}/proxy/metrics/cadvisor
      
      - job_name: 'kubernetes-service-endpoints'
        kubernetes_sd_configs:
        - role: endpoints
        relabel_configs:
        - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
          action: keep
          regex: true
        - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
          action: replace
          target_label: __scheme__
          regex: (https?)
        - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
          action: replace
          target_label: __metrics_path__
          regex: (.+)
        - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
          action: replace
          target_label: __address__
          regex: ([^:]+)(?::\d+)?;(\d+)
          replacement: $1:$2
        - action: labelmap
          regex: __meta_kubernetes_service_label_(.+)
        - source_labels: [__meta_kubernetes_namespace]
          action: replace
          target_label: kubernetes_namespace
        - source_labels: [__meta_kubernetes_service_name]
          action: replace
          target_label: kubernetes_name
      
      - job_name: 'gotak-server'
        kubernetes_sd_configs:
        - role: endpoints
          namespaces:
            names:
            - production
        relabel_configs:
        - source_labels: [__meta_kubernetes_service_name]
          action: keep
          regex: gotak-service
        - source_labels: [__meta_kubernetes_endpoint_port_name]
          action: keep
          regex: http
        - source_labels: [__address__]
          action: replace
          target_label: __address__
          regex: ([^:]+):(.+)
          replacement: $1:8080
        - action: replace
          target_label: __metrics_path__
          replacement: /metrics
        - action: labelmap
          regex: __meta_kubernetes_service_label_(.+)
        - source_labels: [__meta_kubernetes_namespace]
          action: replace
          target_label: kubernetes_namespace
        - source_labels: [__meta_kubernetes_service_name]
          action: replace
          target_label: kubernetes_name

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-rules
  namespace: monitoring
data:
  gotak-rules.yml: |
    groups:
    - name: gotak.rules
      rules:
      - alert: GoTAKHighCPU
        expr: rate(container_cpu_usage_seconds_total{pod=~"gotak-server-.*"}[5m]) > 0.8
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "GoTAK server high CPU usage"
          description: "GoTAK server {{ $labels.pod }} has been using more than 80% CPU for more than 2 minutes"
      
      - alert: GoTAKHighMemory
        expr: container_memory_usage_bytes{pod=~"gotak-server-.*"} / container_spec_memory_limit_bytes > 0.9
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "GoTAK server high memory usage"
          description: "GoTAK server {{ $labels.pod }} is using more than 90% of its memory limit"
      
      - alert: GoTAKServerDown
        expr: up{job="gotak-server"} == 0
        for: 30s
        labels:
          severity: critical
        annotations:
          summary: "GoTAK server is down"
          description: "GoTAK server {{ $labels.instance }} has been down for more than 30 seconds"
      
      - alert: GoTAKHighErrorRate
        expr: rate(gotak_messages_processed_total{status="error"}[5m]) / rate(gotak_messages_processed_total[5m]) > 0.05
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "GoTAK high error rate"
          description: "GoTAK server has error rate of {{ $value | humanizePercentage }} for more than 2 minutes"
      
      - alert: GoTAKDatabaseConnectionsHigh
        expr: gotak_database_connections{state="active"} / gotak_database_connections{state="max"} > 0.8
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "GoTAK database connections high"
          description: "GoTAK is using more than 80% of database connections"
```

### Disaster Recovery

**Backup Procedure**
```bash
#!/bin/bash
# scripts/backup.sh

set -euo pipefail

# Configuration
NAMESPACE="production"
BACKUP_BUCKET="s3://gotak-backups"
RETENTION_DAYS=30
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" >&2
}

# Database backup
backup_database() {
    local backup_name="gotak-db-backup-${TIMESTAMP}.sql"
    
    log "Starting database backup: ${backup_name}"
    
    kubectl exec -n ${NAMESPACE} deployment/gotak-server -- \
        pg_dump ${DATABASE_URL} | \
        gzip > "/tmp/${backup_name}.gz"
    
    aws s3 cp "/tmp/${backup_name}.gz" "${BACKUP_BUCKET}/database/"
    
    # Verify backup
    aws s3api head-object \
        --bucket "${BACKUP_BUCKET#s3://}" \
        --key "database/${backup_name}.gz" >/dev/null
    
    log "Database backup completed: ${backup_name}.gz"
    rm "/tmp/${backup_name}.gz"
}

# Kubernetes resources backup
backup_k8s_resources() {
    local backup_name="gotak-k8s-backup-${TIMESTAMP}.tar.gz"
    local backup_dir="/tmp/k8s-backup-${TIMESTAMP}"
    
    log "Starting Kubernetes resources backup: ${backup_name}"
    
    mkdir -p "${backup_dir}"
    
    # Backup configurations
    kubectl get configmaps -n ${NAMESPACE} -o yaml > "${backup_dir}/configmaps.yaml"
    kubectl get secrets -n ${NAMESPACE} -o yaml > "${backup_dir}/secrets.yaml"
    kubectl get deployments -n ${NAMESPACE} -o yaml > "${backup_dir}/deployments.yaml"
    kubectl get services -n ${NAMESPACE} -o yaml > "${backup_dir}/services.yaml"
    kubectl get ingress -n ${NAMESPACE} -o yaml > "${backup_dir}/ingress.yaml"
    kubectl get hpa -n ${NAMESPACE} -o yaml > "${backup_dir}/hpa.yaml"
    
    # Create archive
    tar -czf "/tmp/${backup_name}" -C "/tmp" "k8s-backup-${TIMESTAMP}"
    
    aws s3 cp "/tmp/${backup_name}" "${BACKUP_BUCKET}/k8s/"
    
    # Verify backup
    aws s3api head-object \
        --bucket "${BACKUP_BUCKET#s3://}" \
        --key "k8s/${backup_name}" >/dev/null
    
    log "Kubernetes backup completed: ${backup_name}"
    rm -rf "${backup_dir}" "/tmp/${backup_name}"
}

# Certificate backup
backup_certificates() {
    local backup_name="gotak-certs-backup-${TIMESTAMP}.tar.gz"
    local backup_dir="/tmp/certs-backup-${TIMESTAMP}"
    
    log "Starting certificates backup: ${backup_name}"
    
    mkdir -p "${backup_dir}"
    
    kubectl get secret gotak-tls -n ${NAMESPACE} -o yaml > "${backup_dir}/tls-secret.yaml"
    kubectl get secret gotak-ca -n ${NAMESPACE} -o yaml > "${backup_dir}/ca-secret.yaml" 2>/dev/null || true
    
    tar -czf "/tmp/${backup_name}" -C "/tmp" "certs-backup-${TIMESTAMP}"
    
    aws s3 cp "/tmp/${backup_name}" "${BACKUP_BUCKET}/certificates/"
    
    log "Certificates backup completed: ${backup_name}"
    rm -rf "${backup_dir}" "/tmp/${backup_name}"
}

# Cleanup old backups
cleanup_old_backups() {
    log "Cleaning up backups older than ${RETENTION_DAYS} days"
    
    local cutoff_date=$(date -d "${RETENTION_DAYS} days ago" +%Y-%m-%d)
    
    # Database backups
    aws s3 ls "${BACKUP_BUCKET}/database/" | \
        awk '$1 < "'${cutoff_date}'" {print $4}' | \
        xargs -I {} aws s3 rm "${BACKUP_BUCKET}/database/{}"
    
    # K8s backups  
    aws s3 ls "${BACKUP_BUCKET}/k8s/" | \
        awk '$1 < "'${cutoff_date}'" {print $4}' | \
        xargs -I {} aws s3 rm "${BACKUP_BUCKET}/k8s/{}"
    
    # Certificate backups
    aws s3 ls "${BACKUP_BUCKET}/certificates/" | \
        awk '$1 < "'${cutoff_date}'" {print $4}' | \
        xargs -I {} aws s3 rm "${BACKUP_BUCKET}/certificates/{}"
}

# Main execution
main() {
    log "Starting backup procedure"
    
    backup_database
    backup_k8s_resources
    backup_certificates
    cleanup_old_backups
    
    log "Backup procedure completed successfully"
}

# Error handling
trap 'log "Backup failed with exit code $?"' ERR

main "$@"
```

**Disaster Recovery Runbook**
```markdown
# GoTAK Disaster Recovery Runbook

## Emergency Contacts
- On-call Engineer: [REDACTED]
- DevOps Lead: [REDACTED]  
- Platform Manager: [REDACTED]

## Recovery Procedures

### Complete Infrastructure Loss

1. **Assess Situation**
   - Determine scope of outage
   - Identify affected regions/zones
   - Estimate data loss window

2. **Initialize Recovery**
   ```bash
   # Clone infrastructure repo
   git clone https://github.com/dfedick/gotak-infrastructure.git
   cd gotak-infrastructure
   
   # Initialize Terraform in DR region
   cd terraform/disaster-recovery
   terraform init
   terraform plan -var="region=us-east-1"
   terraform apply -auto-approve
   ```

3. **Restore Database**
   ```bash
   # Find latest backup
   aws s3 ls s3://gotak-backups/database/ --recursive | tail -1
   
   # Download and restore
   aws s3 cp s3://gotak-backups/database/latest-backup.sql.gz .
   gunzip latest-backup.sql.gz
   
   # Connect to new RDS instance
   psql $DR_DATABASE_URL < latest-backup.sql
   ```

4. **Deploy Application**
   ```bash
   # Update kubeconfig for DR cluster
   aws eks update-kubeconfig --region us-east-1 --name gotak-dr
   
   # Restore K8s resources
   aws s3 cp s3://gotak-backups/k8s/latest-k8s-backup.tar.gz .
   tar -xzf latest-k8s-backup.tar.gz
   
   # Apply configurations
   kubectl apply -f k8s-backup/
   
   # Deploy latest application image
   kubectl set image deployment/gotak-server \
     gotak-server=ghcr.io/dfedick/gotak:latest -n production
   ```

5. **Update DNS**
   ```bash
   # Update Route53 records to point to DR environment
   aws route53 change-resource-record-sets \
     --hosted-zone-id Z123456789 \
     --change-batch file://dns-failover.json
   ```

6. **Verify Recovery**
   - Check application health endpoints
   - Verify database connectivity
   - Test user authentication
   - Validate federation connections

### RTO/RPO Targets
- **RTO (Recovery Time Objective)**: 4 hours
- **RPO (Recovery Point Objective)**: 15 minutes

### Testing Schedule
- Monthly: Backup restoration test
- Quarterly: Full DR drill
- Annually: Complete infrastructure rebuild
```

## Database Schema

```sql
-- Deployment tracking
CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version VARCHAR(50) NOT NULL,
    environment VARCHAR(20) NOT NULL,
    deployed_by VARCHAR(255),
    deployed_at TIMESTAMP DEFAULT NOW(),
    rollback_version VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    notes TEXT
);

-- Infrastructure monitoring
CREATE TABLE infrastructure_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    component VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    metadata JSONB,
    resolved BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    resolved_at TIMESTAMP
);

-- Performance optimization indexes
CREATE INDEX idx_deployments_environment ON deployments(environment, deployed_at DESC);
CREATE INDEX idx_infrastructure_events_type_time ON infrastructure_events(event_type, created_at DESC);
CREATE INDEX idx_infrastructure_events_component ON infrastructure_events(component, resolved);
```

## API Specifications

### Deployment API
```
GET    /api/v1/deployment/status           # Current deployment status
GET    /api/v1/deployment/history          # Deployment history
POST   /api/v1/deployment/rollback         # Trigger rollback
GET    /api/v1/infrastructure/health       # Infrastructure health check
GET    /api/v1/infrastructure/events       # Infrastructure events
```

## Testing Strategy

### Infrastructure Tests
```go
func TestKubernetesDeployment(t *testing.T) {
    config, err := rest.InClusterConfig()
    require.NoError(t, err)
    
    clientset, err := kubernetes.NewForConfig(config)
    require.NoError(t, err)
    
    // Test deployment exists
    deployment, err := clientset.AppsV1().Deployments("production").Get(
        context.TODO(), "gotak-server", metav1.GetOptions{})
    require.NoError(t, err)
    
    // Verify replicas
    assert.Equal(t, int32(3), *deployment.Spec.Replicas)
    
    // Check readiness
    assert.Equal(t, deployment.Status.Replicas, deployment.Status.ReadyReplicas)
}

func TestDatabaseConnectivity(t *testing.T) {
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    require.NoError(t, err)
    defer db.Close()
    
    // Test connection
    err = db.Ping()
    assert.NoError(t, err)
    
    // Test query
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
    assert.NoError(t, err)
}
```

## Acceptance Criteria

### CI/CD Pipeline
- [ ] Automated tests run on every commit
- [ ] Security scanning integrated in pipeline
- [ ] Blue-green deployment working
- [ ] Rollback procedure tested and documented
- [ ] Artifact versioning and retention policies implemented

### Container Orchestration  
- [ ] Kubernetes cluster operational with 99.9% uptime
- [ ] Auto-scaling working under load
- [ ] Pod disruption budgets preventing outages
- [ ] Resource limits preventing resource exhaustion
- [ ] Service mesh providing traffic management

### Infrastructure as Code
- [ ] All infrastructure provisioned via Terraform
- [ ] State management and locking working
- [ ] Environment parity maintained
- [ ] Cost optimization policies implemented
- [ ] Drift detection and remediation automated

### Production Monitoring
- [ ] Comprehensive metrics collected from all components
- [ ] Alerting rules tuned to minimize false positives
- [ ] Distributed tracing operational
- [ ] Log aggregation and searching functional
- [ ] Performance profiling available

### Disaster Recovery
- [ ] Automated backups running successfully
- [ ] DR procedures tested monthly
- [ ] RTO and RPO targets consistently met
- [ ] Cross-region replication working
- [ ] Recovery automation reduces manual intervention

## Dependencies

### Infrastructure
- AWS EKS cluster with managed node groups
- RDS PostgreSQL with Multi-AZ deployment
- ElastiCache Redis for session storage
- Application Load Balancer with SSL termination
- Route 53 for DNS management

### Tools and Services
- GitHub Actions for CI/CD
- Terraform for infrastructure provisioning
- Prometheus and Grafana for monitoring
- Fluentd and Elasticsearch for logging
- AWS Backup for automated backups

## Definition of Done

### Code Quality
- [ ] All code reviewed and approved
- [ ] Infrastructure tests passing
- [ ] Security scans completed without high/critical issues
- [ ] Performance benchmarks meet requirements
- [ ] Disaster recovery procedures tested

### Functionality
- [ ] All user stories completed and accepted
- [ ] CI/CD pipeline operational
- [ ] Production deployment successful
- [ ] Monitoring and alerting active
- [ ] Backup and recovery procedures verified

### Operations
- [ ] Production runbooks completed
- [ ] On-call procedures established
- [ ] Performance baselines documented
- [ ] Capacity planning completed
- [ ] Cost optimization implemented

---

**Sprint Review Date:** [TBD]  
**Sprint Retrospective Date:** [TBD]  
**Project Completion Celebration:** [TBD]
