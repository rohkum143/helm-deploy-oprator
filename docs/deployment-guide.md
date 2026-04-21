# Deployment Guide

## Prerequisites

Before deploying the Nginx Helm Operator, ensure you have:

1. **Kubernetes Cluster**: Version 1.20 or higher
2. **kubectl**: Configured to access your cluster
3. **Helm**: Version 3.x (optional, for manual chart operations)
4. **Docker**: For building custom images (optional)

## Installation Methods

### Method 1: Using Pre-built Manifests

1. **Install CRDs**:
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/your-org/nginx-helm-operator/main/config/crd/bases/deploy.example.com_nginxdeployments.yaml
   ```

2. **Create Namespace**:
   ```bash
   kubectl create namespace nginx-helm-operator-system
   ```

3. **Deploy Operator**:
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/your-org/nginx-helm-operator/main/config/manager/manager.yaml
   ```

### Method 2: Using Kustomize

1. **Clone Repository**:
   ```bash
   git clone https://github.com/your-org/nginx-helm-operator.git
   cd nginx-helm-operator
   ```

2. **Install CRDs**:
   ```bash
   make install
   ```

3. **Deploy Operator**:
   ```bash
   make deploy IMG=nginx-helm-operator:latest
   ```

### Method 3: Development Setup

1. **Build and Deploy**:
   ```bash
   git clone https://github.com/your-org/nginx-helm-operator.git
   cd nginx-helm-operator
   make docker-build docker-push IMG=your-registry/nginx-helm-operator:v1.0.0
   make deploy IMG=your-registry/nginx-helm-operator:v1.0.0
   ```

## Configuration

### Operator Configuration

The operator supports the following configuration options:

| Flag | Description | Default |
|------|-------------|----------|
| `--metrics-bind-address` | Metrics server address | `:8080` |
| `--health-probe-bind-address` | Health probe address | `:8081` |
| `--leader-elect` | Enable leader election | `false` |
| `--helm-charts-path` | Path to Helm charts | `/charts` |

### Environment Variables

```yaml
env:
  - name: HELM_CHARTS_PATH
    value: "/charts"
  - name: LOG_LEVEL
    value: "info"
```

## RBAC Configuration

The operator requires the following permissions:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: nginx-helm-operator-manager-role
rules:
- apiGroups:
  - deploy.example.com
  resources:
  - nginxdeployments
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - deploy.example.com
  resources:
  - nginxdeployments/finalizers
  verbs:
  - update
- apiGroups:
  - deploy.example.com
  resources:
  - nginxdeployments/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - namespaces
  - secrets
  - configmaps
  - services
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
```

## Verification

### Check Operator Status

```bash
# Check if operator is running
kubectl get pods -n nginx-helm-operator-system

# Check operator logs
kubectl logs -n nginx-helm-operator-system deployment/nginx-helm-operator-controller-manager

# Verify CRD installation
kubectl get crd nginxdeployments.deploy.example.com
```

### Deploy Test Application

1. **Create test deployment**:
   ```bash
   kubectl apply -f config/samples/nginx-deployment-sample.yaml
   ```

2. **Check deployment status**:
   ```bash
   kubectl get nginxdeployments
   kubectl describe nginxdeployment nginx-sample
   ```

3. **Verify Nginx pods**:
   ```bash
   kubectl get pods -l app.kubernetes.io/name=sample-nginx
   ```

## Production Considerations

### High Availability

For production deployments, enable leader election:

```yaml
spec:
  template:
    spec:
      containers:
      - name: manager
        args:
        - --leader-elect
        - --metrics-bind-address=:8080
        - --health-probe-bind-address=:8081
```

### Resource Limits

Set appropriate resource limits:

```yaml
resources:
  limits:
    cpu: 500m
    memory: 128Mi
  requests:
    cpu: 10m
    memory: 64Mi
```

### Security

1. **Use specific service account**:
   ```yaml
   serviceAccountName: nginx-helm-operator-controller-manager
   ```

2. **Enable security context**:
   ```yaml
   securityContext:
     allowPrivilegeEscalation: false
     capabilities:
       drop:
       - ALL
     runAsNonRoot: true
     runAsUser: 65532
   ```

3. **Network policies** (if required):
   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: nginx-helm-operator-netpol
   spec:
     podSelector:
       matchLabels:
         control-plane: controller-manager
     policyTypes:
     - Ingress
     - Egress
   ```

### Monitoring

1. **Metrics endpoint**: Available at `:8080/metrics`
2. **Health checks**: Available at `:8081/healthz` and `:8081/readyz`
3. **Prometheus integration**:
   ```yaml
   apiVersion: monitoring.coreos.com/v1
   kind: ServiceMonitor
   metadata:
     name: nginx-helm-operator-metrics
   spec:
     endpoints:
     - path: /metrics
       port: metrics
     selector:
       matchLabels:
         control-plane: controller-manager
   ```

## Troubleshooting

### Common Issues

1. **CRD not found**:
   ```bash
   kubectl apply -f config/crd/bases/
   ```

2. **Permission denied**:
   ```bash
   kubectl apply -f config/rbac/
   ```

3. **Helm chart not found**:
   - Verify charts are included in operator image
   - Check `--helm-charts-path` flag

4. **Operator not starting**:
   ```bash
   kubectl describe pod -n nginx-helm-operator-system
   kubectl logs -n nginx-helm-operator-system deployment/nginx-helm-operator-controller-manager
   ```

### Debug Commands

```bash
# Get all operator resources
kubectl get all -n nginx-helm-operator-system

# Check events
kubectl get events -n nginx-helm-operator-system --sort-by=.metadata.creationTimestamp

# Describe operator deployment
kubectl describe deployment -n nginx-helm-operator-system nginx-helm-operator-controller-manager

# Check RBAC
kubectl auth can-i create nginxdeployments --as=system:serviceaccount:nginx-helm-operator-system:nginx-helm-operator-controller-manager
```

## Upgrading

### Operator Upgrade

1. **Update CRDs** (if changed):
   ```bash
   make install
   ```

2. **Update operator image**:
   ```bash
   make deploy IMG=nginx-helm-operator:v1.1.0
   ```

3. **Verify upgrade**:
   ```bash
   kubectl get deployment -n nginx-helm-operator-system
   kubectl rollout status deployment/nginx-helm-operator-controller-manager -n nginx-helm-operator-system
   ```

### Chart Updates

When updating the base Helm chart:

1. Update chart files in `charts/nginx/`
2. Build new operator image with updated charts
3. Deploy updated operator
4. Existing deployments will be upgraded on next reconciliation

## Uninstallation

### Complete Removal

1. **Delete all NginxDeployment resources**:
   ```bash
   kubectl delete nginxdeployments --all --all-namespaces
   ```

2. **Remove operator**:
   ```bash
   make undeploy
   ```

3. **Remove CRDs**:
   ```bash
   make uninstall
   ```

4. **Clean up namespaces** (if needed):
   ```bash
   kubectl delete namespace nginx-helm-operator-system
   ```

### Partial Removal

To remove only the operator but keep CRDs and existing deployments:

```bash
kubectl delete deployment -n nginx-helm-operator-system nginx-helm-operator-controller-manager
```

Existing Nginx deployments will continue running but won't be managed until the operator is redeployed.
