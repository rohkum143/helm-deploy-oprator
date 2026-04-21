# Nginx Helm Operator

A production-ready Kubernetes operator that manages Nginx deployments using Helm charts with advanced conflict resolution, comprehensive monitoring, and enterprise-grade features. This operator enables seamless deployment and management of Nginx instances across multiple namespaces using Custom Resource Definitions (CRDs).

## 🚀 Features

- **🔧 Custom Resource Definitions (CRDs)**: Define Nginx deployments using Kubernetes CRDs with comprehensive validation
- **⚡ Helm Integration**: Uses Helm v3 for robust deployment management with automatic rollback capabilities
- **🌐 Multi-namespace Support**: Deploy to different namespaces with isolated configurations
- **⚙️ Dynamic Configuration**: Override Helm values through CRD specifications with real-time updates
- **🔄 Lifecycle Management**: Automatic installation, upgrade, rollback, and cleanup of deployments
- **📊 Status Tracking**: Real-time status updates, deployment tracking, and comprehensive health monitoring
- **🛡️ Conflict Resolution**: Advanced reconciliation with exponential backoff and optimistic locking
- **🔒 Security**: RBAC-enabled with security contexts and network policies
- **📈 Monitoring**: Prometheus metrics integration and comprehensive logging
- **🏗️ Production Ready**: Includes health checks, graceful shutdown, and resource management

## 🏗️ Architecture

The operator consists of:

1. **NginxDeployment CRD**: Custom resource for defining Nginx deployments with validation
2. **Controller Manager**: Reconciles desired state with actual state using advanced conflict resolution
3. **Helm Charts**: Pre-packaged Nginx Helm chart with configurable templates
4. **Operator Image**: Container image with operator binary, charts, and dependencies
5. **RBAC Configuration**: Role-based access control for secure operations
6. **Monitoring Stack**: Metrics, health checks, and observability components

## 🚀 Quick Start

### Prerequisites

- **Kubernetes cluster** (v1.20+)
- **kubectl** configured and connected to your cluster
- **Helm v3** (for development and chart management)
- **Docker** (for building custom images)
- **Go 1.21+** (for development)

### 📦 Installation Methods

#### Method 1: Helm Chart Installation (Recommended)

```bash
# Install the operator using Helm chart
helm install nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system \
  --create-namespace

# Verify installation
kubectl get pods -n nginx-helm-operator-system
kubectl get crd nginxdeployments.deploy.example.com
```

#### Method 2: Kustomize Installation

```bash
# Install CRDs
make install

# Deploy the operator
make deploy IMG=docker.io/rohtash672/nginx-helm-operator:latest

# Verify deployment
kubectl get pods -n nginx-helm-operator-system
```

#### Method 3: Manual Installation

```bash
# Apply CRDs manually
kubectl apply -f config/crd/bases/

# Apply RBAC and operator deployment
kubectl apply -f config/rbac/
kubectl apply -f config/manager/
```

### 🎯 Create Your First Nginx Deployment

```yaml
apiVersion: deploy.example.com/v1alpha1
kind: NginxDeployment
metadata:
  name: my-nginx
  namespace: default
spec:
  deploymentName: my-nginx-app
  image: nginx:1.25
  replicas: 2
  namespace: production
  helmValues:
    service:
      type: LoadBalancer
      port: 80
    resources:
      limits:
        cpu: 200m
        memory: 256Mi
      requests:
        cpu: 100m
        memory: 128Mi
```

```bash
# Apply the deployment
kubectl apply -f my-nginx-deployment.yaml

# Check status
kubectl get nginxdeployments
kubectl describe nginxdeployment my-nginx
```

## 📋 CRD Specification

### NginxDeployment Resource

| Field | Type | Description | Required | Default |
|-------|------|-------------|----------|----------|
| `deploymentName` | string | Name of the Helm deployment | Yes | - |
| `image` | string | Nginx container image | No | `nginx:latest` |
| `replicas` | int32 | Number of pod replicas | No | `1` |
| `namespace` | string | Target deployment namespace | No | CR namespace |
| `helmValues` | object | Custom Helm chart values | No | `{}` |
| `chartVersion` | string | Specific Helm chart version | No | `latest` |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | Current deployment phase (`Deploying`, `Deployed`, `Failed`, `Deleting`) |
| `message` | string | Human-readable status message |
| `lastUpdated` | time | Timestamp of last status update |
| `helmReleaseStatus` | string | Helm release status (`deployed`, `failed`, `pending-install`, etc.) |
| `deployedRevision` | int | Current deployed Helm revision number |
| `conditions` | array | Detailed status conditions with types and reasons |
| `observedGeneration` | int64 | Last observed generation of the resource |
| `conflictRetries` | int | Number of conflict resolution retries performed |

## 📚 Examples

### 🔰 Basic Deployment

```yaml
apiVersion: deploy.example.com/v1alpha1
kind: NginxDeployment
metadata:
  name: simple-nginx
  namespace: default
spec:
  deploymentName: simple-nginx
  image: nginx:1.25
  replicas: 1
```

### ⚡ Advanced Deployment with Custom Values

```yaml
apiVersion: deploy.example.com/v1alpha1
kind: NginxDeployment
metadata:
  name: advanced-nginx
  namespace: default
  annotations:
    deploy.example.com/description: "Production Nginx with autoscaling"
spec:
  deploymentName: advanced-nginx
  image: nginx:1.25-alpine
  replicas: 3
  namespace: production
  helmValues:
    service:
      type: LoadBalancer
      port: 8080
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: nlb
    ingress:
      enabled: true
      className: nginx
      annotations:
        cert-manager.io/cluster-issuer: letsencrypt-prod
      hosts:
        - host: nginx.example.com
          paths:
            - path: /
              pathType: Prefix
      tls:
        - secretName: nginx-tls
          hosts:
            - nginx.example.com
    resources:
      limits:
        cpu: 500m
        memory: 512Mi
      requests:
        cpu: 250m
        memory: 256Mi
    autoscaling:
      enabled: true
      minReplicas: 3
      maxReplicas: 20
      targetCPUUtilizationPercentage: 70
      targetMemoryUtilizationPercentage: 80
    nodeSelector:
      node-type: compute
    tolerations:
      - key: "compute-node"
        operator: "Equal"
        value: "true"
        effect: "NoSchedule"
```

### 🌐 Multi-Environment Deployment

```yaml
# Staging Environment
apiVersion: deploy.example.com/v1alpha1
kind: NginxDeployment
metadata:
  name: staging-nginx
  namespace: default
  labels:
    environment: staging
    team: platform
spec:
  deploymentName: staging-app
  namespace: staging
  image: nginx:1.24
  replicas: 2
  helmValues:
    service:
      type: ClusterIP
    resources:
      limits:
        cpu: 200m
        memory: 256Mi
      requests:
        cpu: 100m
        memory: 128Mi
    configMap:
      data:
        nginx.conf: |
          server {
            listen 80;
            location / {
              return 200 'Staging Environment';
            }
          }
---
# Production Environment
apiVersion: deploy.example.com/v1alpha1
kind: NginxDeployment
metadata:
  name: prod-nginx
  namespace: default
  labels:
    environment: production
    team: platform
spec:
  deploymentName: prod-app
  namespace: production
  image: nginx:1.25
  replicas: 5
  helmValues:
    service:
      type: LoadBalancer
    resources:
      limits:
        cpu: 1000m
        memory: 1Gi
      requests:
        cpu: 500m
        memory: 512Mi
    podDisruptionBudget:
      enabled: true
      minAvailable: 3
    affinity:
      podAntiAffinity:
        preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - nginx
            topologyKey: kubernetes.io/hostname
```

### 🔒 Secure Deployment with Network Policies

```yaml
apiVersion: deploy.example.com/v1alpha1
kind: NginxDeployment
metadata:
  name: secure-nginx
  namespace: default
spec:
  deploymentName: secure-app
  namespace: secure
  image: nginx:1.25-alpine
  replicas: 3
  helmValues:
    securityContext:
      runAsNonRoot: true
      runAsUser: 101
      fsGroup: 101
    containerSecurityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
    networkPolicy:
      enabled: true
      ingress:
        - from:
          - namespaceSelector:
              matchLabels:
                name: ingress-nginx
          ports:
          - protocol: TCP
            port: 80
    serviceAccount:
      create: true
      annotations:
        eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/nginx-service-role
```

## 🛠️ Development

### 🏗️ Building and Testing

#### Code Generation and Validation

```bash
# Generate DeepCopy methods and CRD manifests
make generate

# Generate RBAC, CRD, and webhook configurations
make manifests

# Format code
make fmt

# Lint code
make vet

# Run comprehensive tests
make test

# Build operator binary
make build
```

#### Local Development

```bash
# Run operator locally (requires kubeconfig)
make run

# Run with custom flags
go run ./main.go \
  --metrics-bind-address=:8080 \
  --health-probe-bind-address=:8081 \
  --leader-elect=false \
  --helm-charts-path=./charts
```

### 🐳 Docker Operations

#### Building Images

```bash
# Build operator image
make docker-build IMG=docker.io/rohtash672/nginx-helm-operator:latest

# Build with specific tag
make docker-build IMG=docker.io/rohtash672/nginx-helm-operator:v1.2.0

# Build and test image
docker build -t nginx-helm-operator:test .
docker run --rm nginx-helm-operator:test --help
```

#### Pushing Images

```bash
# Push to registry
make docker-push IMG=docker.io/rohtash672/nginx-helm-operator:latest

# Push with authentication
docker login docker.io
make docker-push IMG=docker.io/rohtash672/nginx-helm-operator:v1.2.0
```

### 📦 Helm Chart Development

#### Updating Base Nginx Charts

```bash
# Modify charts in charts/nginx/
vim charts/nginx/values.yaml
vim charts/nginx/templates/deployment.yaml

# Validate chart
helm lint charts/nginx/
helm template test charts/nginx/ --debug

# Package chart (optional)
helm package charts/nginx/
```

#### Updating Operator Helm Chart

```bash
# Modify operator chart
vim helm-chart/nginx-helm-operator/values.yaml
vim helm-chart/nginx-helm-operator/templates/deployment.yaml

# Validate operator chart
helm lint helm-chart/nginx-helm-operator/
helm template nginx-helm-operator helm-chart/nginx-helm-operator/ --debug

# Test installation
helm install test-operator helm-chart/nginx-helm-operator/ --dry-run
```

#### Complete Update Workflow

```bash
# 1. Update charts
vim charts/nginx/templates/deployment.yaml

# 2. Build new operator image
make docker-build docker-push IMG=docker.io/rohtash672/nginx-helm-operator:v1.2.0

# 3. Update operator deployment
helm upgrade nginx-helm-operator helm-chart/nginx-helm-operator/ \
  --set image.tag=v1.2.0 \
  --namespace nginx-helm-operator-system

# 4. Verify update
kubectl rollout status deployment/nginx-helm-operator-controller-manager -n nginx-helm-operator-system
```

### 🧪 Testing

#### Unit Tests

```bash
# Run unit tests
go test ./... -v

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test
go test ./controllers -run TestNginxDeploymentReconciler
```

#### Integration Tests

```bash
# Run integration tests (requires cluster)
TEST_USE_EXISTING_CLUSTER=true make test

# Run with custom kubeconfig
KUBECONFIG=/path/to/kubeconfig TEST_USE_EXISTING_CLUSTER=true make test
```

#### End-to-End Tests

```bash
# Deploy operator for testing
make deploy IMG=nginx-helm-operator:test

# Run E2E tests
kubectl apply -f config/samples/
kubectl wait --for=condition=Ready nginxdeployment --all --timeout=300s

# Cleanup
make undeploy
```

### 🔧 CLI Flags and Configuration

#### Operator Manager Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--metrics-bind-address` | string | `:8080` | Address for metrics endpoint |
| `--health-probe-bind-address` | string | `:8081` | Address for health probe endpoint |
| `--leader-elect` | bool | `false` | Enable leader election |
| `--helm-charts-path` | string | `/charts` | Path to Helm charts directory |

#### Environment Variables

```bash
# Set log level
export LOG_LEVEL=debug

# Set custom kubeconfig
export KUBECONFIG=/path/to/kubeconfig

# Set namespace for operator
export OPERATOR_NAMESPACE=nginx-helm-operator-system

# Set resource limits
export GOMAXPROCS=2
export GOMEMLIMIT=512MiB
```

#### Development Configuration

```yaml
# config/dev/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../default

patchesStrategicMerge:
- manager_dev_patch.yaml

namespace: nginx-helm-operator-dev
```

### 🐛 Debugging

#### Enable Debug Logging

```bash
# Run with debug logging
go run ./main.go --v=2

# Enable development mode
go run ./main.go --zap-devel=true --zap-log-level=debug
```

#### Debug in Cluster

```bash
# Update deployment with debug flags
kubectl patch deployment nginx-helm-operator-controller-manager \
  -n nginx-helm-operator-system \
  --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--v=2"}]'

# View debug logs
kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager -f
```

## 📊 Monitoring and Observability

### 🔍 Checking Deployment Status

#### Basic Status Commands

```bash
# List all nginx deployments across namespaces
kubectl get nginxdeployments --all-namespaces

# Get detailed status with events
kubectl describe nginxdeployment my-nginx

# Watch deployment status in real-time
kubectl get nginxdeployments -w

# Check deployment status with custom columns
kubectl get nginxdeployments -o custom-columns=NAME:.metadata.name,PHASE:.status.phase,MESSAGE:.status.message,AGE:.metadata.creationTimestamp
```

#### Operator Monitoring

```bash
# Check operator pod status
kubectl get pods -n nginx-helm-operator-system

# View operator logs with timestamps
kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager --timestamps=true

# Follow operator logs in real-time
kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager -f

# Check operator metrics
kubectl port-forward -n nginx-helm-operator-system svc/nginx-helm-operator-controller-manager-metrics-service 8080:8443
curl -k https://localhost:8080/metrics
```

#### Helm Release Monitoring

```bash
# List all Helm releases managed by operator
helm list --all-namespaces | grep -E "(staging-app|prod-app)"

# Get specific release status
helm status staging-app -n staging

# View release history
helm history staging-app -n staging

# Check release values
helm get values staging-app -n staging
```

### 📈 Status Phases and Conditions

#### Deployment Phases

| Phase | Description | Next Actions |
|-------|-------------|-------------|
| `Deploying` | Deployment in progress | Wait for completion or check logs |
| `Deployed` | Successfully deployed and healthy | Monitor for changes |
| `Failed` | Deployment failed | Check events and logs for errors |
| `Deleting` | Cleanup in progress | Wait for finalizer completion |
| `Unknown` | Status cannot be determined | Check operator connectivity |

#### Condition Types

```bash
# Check detailed conditions
kubectl get nginxdeployment my-nginx -o jsonpath='{.status.conditions}' | jq .
```

| Condition Type | Status | Reason | Description |
|---------------|--------|--------|-----------|
| `Ready` | `True`/`False` | `DeploymentReady`/`DeploymentFailed` | Overall readiness status |
| `HelmReleaseReady` | `True`/`False` | `InstallSucceeded`/`InstallFailed` | Helm release status |
| `Progressing` | `True`/`False` | `NewReplicaSetAvailable` | Deployment progress |
| `Available` | `True`/`False` | `MinimumReplicasAvailable` | Pod availability |

### 📊 Metrics and Alerts

#### Prometheus Metrics

The operator exposes the following metrics:

```bash
# Controller metrics
controller_runtime_reconcile_total{controller="nginxdeployment"}
controller_runtime_reconcile_errors_total{controller="nginxdeployment"}
controller_runtime_reconcile_time_seconds{controller="nginxdeployment"}

# Custom operator metrics
nginx_operator_deployments_total
nginx_operator_deployment_duration_seconds
nginx_operator_helm_operations_total
nginx_operator_conflict_retries_total
```

#### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Nginx Helm Operator",
    "panels": [
      {
        "title": "Active Deployments",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(nginx_operator_deployments_total{phase=\"Deployed\"})"
          }
        ]
      },
      {
        "title": "Reconciliation Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(controller_runtime_reconcile_total{controller=\"nginxdeployment\"}[5m])"
          }
        ]
      }
    ]
  }
}
```

#### AlertManager Rules

```yaml
groups:
- name: nginx-helm-operator
  rules:
  - alert: NginxOperatorDown
    expr: up{job="nginx-helm-operator-metrics"} == 0
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "Nginx Helm Operator is down"
      
  - alert: NginxDeploymentFailed
    expr: nginx_operator_deployments_total{phase="Failed"} > 0
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "Nginx deployment failed"
      
  - alert: HighReconciliationErrors
    expr: rate(controller_runtime_reconcile_errors_total{controller="nginxdeployment"}[5m]) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High reconciliation error rate"
```

### 🔍 Health Checks

#### Operator Health Endpoints

```bash
# Check operator health
kubectl port-forward -n nginx-helm-operator-system svc/nginx-helm-operator-controller-manager 8081:8081
curl http://localhost:8081/healthz
curl http://localhost:8081/readyz

# Check health in cluster
kubectl get pods -n nginx-helm-operator-system -o jsonpath='{.items[0].status.conditions}' | jq .
```

#### Deployment Health Validation

```bash
# Validate nginx deployment health
kubectl get deployment -n staging staging-app
kubectl get pods -n staging -l app=staging-app

# Check service endpoints
kubectl get endpoints -n staging staging-app

# Test service connectivity
kubectl run test-pod --image=curlimages/curl --rm -it --restart=Never -- curl http://staging-app.staging.svc.cluster.local
```

### 📋 Monitoring Checklist

- [ ] Operator pods are running and ready
- [ ] Controller manager metrics are accessible
- [ ] CRD resources are being reconciled
- [ ] Helm releases are in desired state
- [ ] Target deployments are healthy
- [ ] No persistent error conditions
- [ ] Resource usage within limits
- [ ] Logs show normal operation
- [ ] Health endpoints responding
- [ ] Prometheus metrics available

## 🔧 Troubleshooting

### 🚨 Common Issues and Solutions

#### 1. Reconciliation Conflicts

**Error**: `Operation cannot be fulfilled on nginxdeployments.deploy.example.com "staging-nginx": the object has been modified`

**Solutions**:
```bash
# Force delete stuck resource
kubectl delete nginxdeployment staging-nginx --force --grace-period=0

# Check for multiple operators
kubectl get pods -n nginx-helm-operator-system

# Verify operator has conflict resolution enabled
kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager | grep -i "conflict\|retry"
```

#### 2. Helm Chart Issues

**Error**: `Helm chart not found` or `failed to install Helm chart`

**Solutions**:
```bash
# Verify chart exists in operator image
kubectl exec -it deployment/nginx-helm-operator-controller-manager -n nginx-helm-operator-system -- ls -la /charts/nginx

# Check helm-charts-path flag
kubectl get deployment nginx-helm-operator-controller-manager -n nginx-helm-operator-system -o yaml | grep helm-charts-path

# Validate chart syntax
helm lint charts/nginx/
helm template test charts/nginx/ --debug
```

#### 3. RBAC Permission Errors

**Error**: `serviceaccounts "staging-app" is forbidden`

**Solutions**:
```bash
# Check operator service account permissions
kubectl auth can-i '*' '*' --as=system:serviceaccount:nginx-helm-operator-system:nginx-helm-operator-controller-manager

# Verify cluster role binding
kubectl get clusterrolebinding | grep nginx-helm-operator

# Check specific permission
kubectl auth can-i create serviceaccounts --as=system:serviceaccount:nginx-helm-operator-system:nginx-helm-operator-controller-manager -n staging

# Fix RBAC (if needed)
kubectl apply -f config/rbac/
```

#### 4. Namespace Issues

**Error**: `namespace "staging" not found`

**Solutions**:
```bash
# Create missing namespace
kubectl create namespace staging

# Check operator permissions to create namespaces
kubectl auth can-i create namespaces --as=system:serviceaccount:nginx-helm-operator-system:nginx-helm-operator-controller-manager

# Verify namespace in CRD spec
kubectl get nginxdeployment staging-nginx -o jsonpath='{.spec.namespace}'
```

#### 5. Controller Crashes

**Error**: `panic: odd number of arguments passed as key-value pairs for logging`

**Solutions**:
```bash
# Check for latest operator image
kubectl get deployment nginx-helm-operator-controller-manager -n nginx-helm-operator-system -o jsonpath='{.spec.template.spec.containers[0].image}'

# Update to fixed image
helm upgrade nginx-helm-operator helm-chart/nginx-helm-operator/ --set image.tag=latest

# Verify fix
kubectl rollout status deployment/nginx-helm-operator-controller-manager -n nginx-helm-operator-system
```

#### 6. Resource Stuck in Deleting State

**Error**: NginxDeployment stuck with finalizers

**Solutions**:
```bash
# Check finalizers
kubectl get nginxdeployment staging-nginx -o jsonpath='{.metadata.finalizers}'

# Remove finalizers (emergency only)
kubectl patch nginxdeployment staging-nginx -p '{"metadata":{"finalizers":null}}' --type=merge

# Check for orphaned Helm releases
helm list -A | grep staging-app
helm delete staging-app -n staging
```

### 🔍 Debug Commands

#### Operator Diagnostics

```bash
# Check operator status
kubectl get pods -n nginx-helm-operator-system -o wide
kubectl describe pod -n nginx-helm-operator-system -l control-plane=controller-manager

# View recent logs with context
kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager --tail=100 --timestamps

# Check operator metrics
kubectl port-forward -n nginx-helm-operator-system svc/nginx-helm-operator-controller-manager-metrics-service 8080:8443
curl -k https://localhost:8080/metrics | grep nginx_operator
```

#### CRD and Resource Diagnostics

```bash
# Verify CRD installation
kubectl get crd nginxdeployments.deploy.example.com -o yaml

# Check CRD validation
kubectl explain nginxdeployment.spec

# List all nginx deployments with status
kubectl get nginxdeployments --all-namespaces -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,PHASE:.status.phase,MESSAGE:.status.message

# Get detailed resource information
kubectl describe nginxdeployment <name> -n <namespace>
```

#### Helm Release Diagnostics

```bash
# List all Helm releases
helm list --all-namespaces

# Check specific release status
helm status <release-name> -n <namespace>

# View release history
helm history <release-name> -n <namespace>

# Get release manifest
helm get manifest <release-name> -n <namespace>

# Check release values
helm get values <release-name> -n <namespace>
```

#### Kubernetes Events

```bash
# View operator events
kubectl get events -n nginx-helm-operator-system --sort-by=.metadata.creationTimestamp

# View target namespace events
kubectl get events -n staging --sort-by=.metadata.creationTimestamp

# Filter events by reason
kubectl get events --all-namespaces --field-selector reason=FailedCreate
```

### 🩺 Health Check Script

```bash
#!/bin/bash
# nginx-operator-health-check.sh

echo "=== Nginx Helm Operator Health Check ==="

# Check operator namespace
echo "Checking operator namespace..."
kubectl get namespace nginx-helm-operator-system || echo "❌ Operator namespace not found"

# Check operator pods
echo "Checking operator pods..."
OPERATOR_PODS=$(kubectl get pods -n nginx-helm-operator-system -l control-plane=controller-manager --no-headers)
if [[ -z "$OPERATOR_PODS" ]]; then
    echo "❌ No operator pods found"
else
    echo "✅ Operator pods:"
    echo "$OPERATOR_PODS"
fi

# Check CRD
echo "Checking CRD installation..."
kubectl get crd nginxdeployments.deploy.example.com >/dev/null 2>&1 && echo "✅ CRD installed" || echo "❌ CRD not found"

# Check RBAC
echo "Checking RBAC permissions..."
kubectl auth can-i create deployments --as=system:serviceaccount:nginx-helm-operator-system:nginx-helm-operator-controller-manager >/dev/null 2>&1 && echo "✅ RBAC configured" || echo "❌ RBAC issues detected"

# Check operator logs for errors
echo "Checking for recent errors..."
ERRORS=$(kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager --tail=50 | grep -i error | wc -l)
if [[ $ERRORS -gt 0 ]]; then
    echo "⚠️  Found $ERRORS recent errors in logs"
else
    echo "✅ No recent errors in logs"
fi

# Check nginx deployments
echo "Checking nginx deployments..."
DEPLOYMENTS=$(kubectl get nginxdeployments --all-namespaces --no-headers 2>/dev/null | wc -l)
echo "📊 Total nginx deployments: $DEPLOYMENTS"

echo "=== Health Check Complete ==="
```

### 📞 Getting Support

#### Information to Gather

1. **Operator Version and Configuration**:
   ```bash
   kubectl get deployment nginx-helm-operator-controller-manager -n nginx-helm-operator-system -o yaml > operator-deployment.yaml
   ```

2. **CRD and Resource Status**:
   ```bash
   kubectl get nginxdeployments --all-namespaces -o yaml > nginx-deployments.yaml
   kubectl describe nginxdeployment <problematic-deployment> > deployment-details.txt
   ```

3. **Logs and Events**:
   ```bash
   kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager --tail=500 > operator-logs.txt
   kubectl get events --all-namespaces --sort-by=.metadata.creationTimestamp > cluster-events.txt
   ```

4. **Cluster Information**:
   ```bash
   kubectl version > cluster-version.txt
   kubectl get nodes > cluster-nodes.txt
   helm version > helm-version.txt
   ```

#### Creating Support Bundle

```bash
#!/bin/bash
# create-support-bundle.sh

BUNDLE_DIR="nginx-operator-support-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BUNDLE_DIR"

echo "Creating support bundle in $BUNDLE_DIR..."

# Operator information
kubectl get deployment nginx-helm-operator-controller-manager -n nginx-helm-operator-system -o yaml > "$BUNDLE_DIR/operator-deployment.yaml"
kubectl get pods -n nginx-helm-operator-system -o yaml > "$BUNDLE_DIR/operator-pods.yaml"
kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager --tail=1000 > "$BUNDLE_DIR/operator-logs.txt"

# CRD and resources
kubectl get crd nginxdeployments.deploy.example.com -o yaml > "$BUNDLE_DIR/crd.yaml"
kubectl get nginxdeployments --all-namespaces -o yaml > "$BUNDLE_DIR/nginx-deployments.yaml"

# Events and cluster info
kubectl get events --all-namespaces --sort-by=.metadata.creationTimestamp > "$BUNDLE_DIR/events.txt"
kubectl version > "$BUNDLE_DIR/cluster-version.txt"
helm version > "$BUNDLE_DIR/helm-version.txt"

# Create archive
tar -czf "$BUNDLE_DIR.tar.gz" "$BUNDLE_DIR"
echo "Support bundle created: $BUNDLE_DIR.tar.gz"
```

## 🤝 Contributing

We welcome contributions to the Nginx Helm Operator! Please follow these guidelines to ensure a smooth contribution process.

### 🚀 Getting Started

1. **Fork the repository**:
   ```bash
   git clone https://github.com/your-username/nginx-helm-operator.git
   cd nginx-helm-operator
   ```

2. **Set up development environment**:
   ```bash
   # Install dependencies
   go mod download
   
   # Install development tools
   make install-tools
   
   # Verify setup
   make verify
   ```

3. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

### 🛠️ Development Workflow

1. **Make your changes**:
   - Follow Go best practices and existing code style
   - Update documentation for any new features
   - Add or update tests for your changes

2. **Run quality checks**:
   ```bash
   # Format code
   make fmt
   
   # Run linting
   make vet
   
   # Run tests
   make test
   
   # Generate manifests
   make manifests
   
   # Verify all checks pass
   make verify-all
   ```

3. **Test your changes**:
   ```bash
   # Build and test locally
   make build
   make run
   
   # Test with Docker
   make docker-build IMG=nginx-helm-operator:test
   make deploy IMG=nginx-helm-operator:test
   ```

### 📝 Contribution Guidelines

#### Code Style
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions small and focused

#### Testing Requirements
- Add unit tests for new functionality
- Ensure all tests pass: `make test`
- Maintain or improve code coverage
- Add integration tests for significant features

#### Documentation
- Update README.md for new features
- Add inline code comments
- Update CRD documentation
- Include examples for new functionality

#### Commit Messages
Use conventional commit format:
```
type(scope): description

[optional body]

[optional footer]
```

Examples:
```
feat(controller): add conflict resolution mechanism
fix(helm): resolve chart path issue
docs(readme): update installation instructions
test(controller): add reconciliation tests
```

### 🔍 Pull Request Process

1. **Ensure your PR**:
   - [ ] Has a clear title and description
   - [ ] References any related issues
   - [ ] Includes tests for new functionality
   - [ ] Updates documentation as needed
   - [ ] Passes all CI checks

2. **Submit your pull request**:
   ```bash
   git push origin feature/your-feature-name
   ```
   Then create a PR through GitHub interface.

3. **Address review feedback**:
   - Respond to reviewer comments
   - Make requested changes
   - Update tests if needed

### 🧪 Testing

#### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests
TEST_USE_EXISTING_CLUSTER=true make test

# E2E tests
make test-e2e

# Performance tests
make test-performance
```

#### Adding Tests

```go
// Example test structure
func TestNginxDeploymentReconciler_Reconcile(t *testing.T) {
    tests := []struct {
        name    string
        setup   func(*testing.T) *NginxDeploymentReconciler
        want    ctrl.Result
        wantErr bool
    }{
        {
            name: "successful reconciliation",
            setup: func(t *testing.T) *NginxDeploymentReconciler {
                // Setup test environment
                return &NginxDeploymentReconciler{}
            },
            want:    ctrl.Result{},
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r := tt.setup(t)
            got, err := r.Reconcile(context.TODO(), ctrl.Request{})
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Reconcile() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 📋 Release Process

1. **Version Tagging**:
   ```bash
   git tag -a v1.2.0 -m "Release v1.2.0"
   git push origin v1.2.0
   ```

2. **Release Notes**:
   - Document new features
   - List bug fixes
   - Note breaking changes
   - Include upgrade instructions

### 🏆 Recognition

Contributors will be recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project documentation

### 📞 Getting Help

- **Discussions**: Use GitHub Discussions for questions
- **Issues**: Report bugs and request features via GitHub Issues
- **Chat**: Join our community Slack/Discord (if available)

## 📄 License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.

```
Copyright 2024 Nginx Helm Operator Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

---

## �️ CLI Reference

### Make Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `make help` | Display all available make targets | `make help` |
| `make build` | Build the operator binary | `make build` |
| `make run` | Run the operator locally | `make run` |
| `make test` | Run unit tests with coverage | `make test` |
| `make fmt` | Format Go code | `make fmt` |
| `make vet` | Run Go vet for code analysis | `make vet` |
| `make generate` | Generate DeepCopy methods | `make generate` |
| `make manifests` | Generate CRD and RBAC manifests | `make manifests` |
| `make docker-build` | Build Docker image | `make docker-build IMG=<image>` |
| `make docker-push` | Push Docker image to registry | `make docker-push IMG=<image>` |
| `make install` | Install CRDs to cluster | `make install` |
| `make uninstall` | Remove CRDs from cluster | `make uninstall` |
| `make deploy` | Deploy operator to cluster | `make deploy IMG=<image>` |
| `make undeploy` | Remove operator from cluster | `make undeploy` |

### Operator Binary Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--metrics-bind-address` | string | `:8080` | Address for metrics endpoint |
| `--health-probe-bind-address` | string | `:8081` | Address for health probe endpoint |
| `--leader-elect` | bool | `false` | Enable leader election for HA |
| `--helm-charts-path` | string | `/charts` | Path to Helm charts directory |
| `--v` | int | `0` | Log verbosity level (0-10) |
| `--zap-devel` | bool | `false` | Enable development mode logging |
| `--zap-log-level` | string | `info` | Log level (debug, info, warn, error) |

### Helm Chart Configuration

#### Installation Commands

```bash
# Basic installation
helm install nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system \
  --create-namespace

# Installation with custom values
helm install nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system \
  --create-namespace \
  --set image.tag=v1.2.0 \
  --set operator.replicaCount=2 \
  --set monitoring.enabled=true

# Installation with values file
helm install nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system \
  --create-namespace \
  --values custom-values.yaml
```

#### Management Commands

```bash
# Upgrade operator
helm upgrade nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system

# Rollback operator
helm rollback nginx-helm-operator 1 --namespace nginx-helm-operator-system

# Uninstall operator
helm uninstall nginx-helm-operator --namespace nginx-helm-operator-system

# Get status
helm status nginx-helm-operator --namespace nginx-helm-operator-system

# Get values
helm get values nginx-helm-operator --namespace nginx-helm-operator-system
```

### kubectl Commands for NginxDeployment

#### Resource Management

```bash
# Create nginx deployment
kubectl apply -f nginx-deployment.yaml

# List all nginx deployments
kubectl get nginxdeployments --all-namespaces

# Get specific deployment
kubectl get nginxdeployment my-nginx -o yaml

# Describe deployment with events
kubectl describe nginxdeployment my-nginx

# Delete deployment
kubectl delete nginxdeployment my-nginx

# Edit deployment
kubectl edit nginxdeployment my-nginx

# Patch deployment
kubectl patch nginxdeployment my-nginx --type='merge' -p='{"spec":{"replicas":5}}'
```

#### Status and Monitoring

```bash
# Watch deployment status
kubectl get nginxdeployments -w

# Get deployment status with custom columns
kubectl get nginxdeployments -o custom-columns=NAME:.metadata.name,PHASE:.status.phase,MESSAGE:.status.message

# Check deployment conditions
kubectl get nginxdeployment my-nginx -o jsonpath='{.status.conditions}' | jq .

# Monitor deployment events
kubectl get events --field-selector involvedObject.kind=NginxDeployment
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|----------|
| `KUBECONFIG` | Path to kubeconfig file | `~/.kube/config` |
| `OPERATOR_NAMESPACE` | Operator deployment namespace | `nginx-helm-operator-system` |
| `LOG_LEVEL` | Logging level | `info` |
| `GOMAXPROCS` | Maximum Go processes | `auto` |
| `GOMEMLIMIT` | Go memory limit | `unlimited` |
| `HELM_CACHE_HOME` | Helm cache directory | `/tmp/.helmcache` |
| `HELM_CONFIG_HOME` | Helm config directory | `/tmp/.helmconfig` |

## �📚 Additional Resources

### 🔗 Useful Links

- **Kubernetes Operator Pattern**: [https://kubernetes.io/docs/concepts/extend-kubernetes/operator/](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- **Helm Documentation**: [https://helm.sh/docs/](https://helm.sh/docs/)
- **Controller Runtime**: [https://pkg.go.dev/sigs.k8s.io/controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- **Kubebuilder**: [https://book.kubebuilder.io/](https://book.kubebuilder.io/)
- **Kubernetes API Reference**: [https://kubernetes.io/docs/reference/kubernetes-api/](https://kubernetes.io/docs/reference/kubernetes-api/)

### 📖 Related Projects

- **Helm Operator**: [https://github.com/fluxcd/helm-operator](https://github.com/fluxcd/helm-operator)
- **Operator SDK**: [https://github.com/operator-framework/operator-sdk](https://github.com/operator-framework/operator-sdk)
- **Kubernetes SIGs**: [https://github.com/kubernetes-sigs](https://github.com/kubernetes-sigs)
- **CNCF Landscape**: [https://landscape.cncf.io/](https://landscape.cncf.io/)

### 🎓 Learning Resources

- **Kubernetes Operators Book**: Programming Kubernetes Operators
- **Cloud Native Computing Foundation**: [https://www.cncf.io/](https://www.cncf.io/)
- **Kubernetes Documentation**: [https://kubernetes.io/docs/](https://kubernetes.io/docs/)
- **Go Programming**: [https://golang.org/doc/](https://golang.org/doc/)
- **Container Best Practices**: [https://cloud.google.com/architecture/best-practices-for-building-containers](https://cloud.google.com/architecture/best-practices-for-building-containers)

### 🏗️ Architecture Diagrams

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   kubectl user  │    │  Nginx Operator │    │  Target Cluster │
│                 │    │   Controller    │    │   Namespaces    │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          │ 1. Apply CRD         │                      │
          ├─────────────────────►│                      │
          │                      │                      │
          │                      │ 2. Reconcile        │
          │                      ├─────────────────────►│
          │                      │                      │
          │                      │ 3. Deploy Helm      │
          │                      │    Chart             │
          │                      ├─────────────────────►│
          │                      │                      │
          │ 4. Status Update     │                      │
          │◄─────────────────────┤                      │
          │                      │                      │
```

---

**Made with ❤️ by the Nginx Helm Operator community**

**Version**: 1.0.0 | **Last Updated**: April 2026 | **License**: Apache 2.0
