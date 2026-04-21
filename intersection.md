# nginx-helm-operator Troubleshooting Guide

## Overview

This document provides comprehensive instructions for troubleshooting nginx-helm-operator reconciliation conflicts, deployment procedures, and best practices.

## Table of Contents

1. [Common Issues and Solutions](#common-issues-and-solutions)
2. [Reconciliation Conflict Resolution](#reconciliation-conflict-resolution)
3. [Deployment Procedures](#deployment-procedures)
4. [Best Practices](#best-practices)
5. [Debugging Commands](#debugging-commands)
6. [Performance Optimization](#performance-optimization)
7. [Security Considerations](#security-considerations)

## Common Issues and Solutions

### 1. "Operation cannot be fulfilled" Error

**Error Message:**
```
Operation cannot be fulfilled on nginxdeployments.deploy.example.com "staging-nginx": the object has been modified; please apply your changes to the latest version and try again
```

**Root Cause:**
- Resource version conflicts due to concurrent modifications
- Multiple controllers or manual changes interfering with operator
- Lack of proper conflict resolution mechanisms

**Solution:**
1. **Immediate Fix:**
   ```bash
   # Delete the stuck resource
   kubectl delete nginxdeployment staging-nginx --force --grace-period=0
   
   # Wait for cleanup
   kubectl get nginxdeployment -w
   
   # Reapply the resource
   kubectl apply -f helm-crd.yml
   ```

2. **Long-term Fix:**
   - Ensure the operator has been updated with conflict resolution mechanisms
   - Rebuild and redeploy the operator image with the latest fixes

### 2. Controller Panic/Crash

**Error Message:**
```
panic: odd number of arguments passed as key-value pairs for logging
```

**Solution:**
1. Check controller logs:
   ```bash
   kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager
   ```

2. Rebuild operator with logging fixes:
   ```bash
   docker build -t docker.io/rohtash672/nginx-helm-operator:latest .
   docker push docker.io/rohtash672/nginx-helm-operator:latest
   ```

### 3. Helm Chart Deployment Failures

**Common Causes:**
- Missing RBAC permissions
- Incorrect Helm chart path
- Invalid Helm values
- Target namespace issues

**Solution:**
1. Verify RBAC permissions:
   ```bash
   kubectl auth can-i create serviceaccounts --as=system:serviceaccount:nginx-helm-operator-system:nginx-helm-operator-controller-manager -n staging
   ```

2. Check Helm chart availability:
   ```bash
   kubectl exec -it deployment/nginx-helm-operator-controller-manager -n nginx-helm-operator-system -- ls -la /charts/nginx
   ```

## Reconciliation Conflict Resolution

### Understanding the Conflict Resolution Mechanism

The updated nginx-helm-operator implements several conflict resolution strategies:

1. **Exponential Backoff Retry:**
   - Initial delay: 1 second
   - Maximum delay: 32 seconds
   - Backoff multiplier: 2.0
   - Maximum retry attempts: 5

2. **Resource Version Management:**
   - Always fetch latest resource version before updates
   - Separate retry logic for spec and status updates
   - Conflict-aware finalizer management

3. **Optimistic Locking:**
   - Uses Kubernetes resource versions for optimistic concurrency control
   - Automatic retry on conflict errors
   - Graceful degradation on persistent conflicts

### Configuration Parameters

```go
// Conflict resolution constants
maxRetryAttempts = 5
initialRetryDelay = 1 * time.Second
maxRetryDelay = 32 * time.Second
backoffMultiplier = 2.0
```

### Monitoring Conflict Resolution

1. **Enable Debug Logging:**
   ```yaml
   # In deployment.yaml
   args:
     - --v=2  # Increase verbosity for conflict resolution logs
   ```

2. **Watch for Conflict Resolution Events:**
   ```bash
   kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager | grep -i "conflict\|retry"
   ```

## Deployment Procedures

### 1. Building and Pushing the Operator Image

```bash
# Build the operator image
docker build -t docker.io/rohtash672/nginx-helm-operator:latest .

# Push to registry
docker push docker.io/rohtash672/nginx-helm-operator:latest

# Verify image
docker images | grep nginx-helm-operator
```

### 2. Deploying via Helm Chart

```bash
# Install the operator
helm install nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system \
  --create-namespace

# Verify deployment
kubectl get pods -n nginx-helm-operator-system
kubectl get crd | grep nginxdeployments
```

### 3. Deploying NginxDeployment Resources

```bash
# Apply the custom resource
kubectl apply -f helm-crd.yml

# Monitor deployment
kubectl get nginxdeployment staging-nginx -o yaml
kubectl describe nginxdeployment staging-nginx
```

### 4. Upgrading the Operator

```bash
# Update the Helm chart
helm upgrade nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system

# Verify upgrade
kubectl rollout status deployment/nginx-helm-operator-controller-manager -n nginx-helm-operator-system
```

## Best Practices

### 1. Resource Management

- **Always use resource limits:**
  ```yaml
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 250m
      memory: 256Mi
  ```

- **Implement proper health checks:**
  ```yaml
  livenessProbe:
    httpGet:
      path: /healthz
      port: 8081
    initialDelaySeconds: 15
    periodSeconds: 20
  ```

### 2. Configuration Management

- **Use annotations for tracking:**
  ```yaml
  metadata:
    annotations:
      deploy.example.com/last-applied-configuration: "..."
      deploy.example.com/conflict-resolution: "enabled"
  ```

- **Leverage Helm values for flexibility:**
  ```yaml
  spec:
    helmValues:
      service:
        type: ClusterIP
        port: 80
      resources:
        limits:
          cpu: 500m
          memory: 512Mi
  ```

### 3. Monitoring and Observability

- **Enable metrics collection:**
  ```bash
  kubectl port-forward -n nginx-helm-operator-system svc/nginx-helm-operator-controller-manager-metrics-service 8080:8443
  curl -k https://localhost:8080/metrics
  ```

- **Set up log aggregation:**
  ```bash
  kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager -f
  ```

### 4. Security Best Practices

- **Use least privilege RBAC:**
  ```yaml
  rules:
  - apiGroups: [""]
    resources: ["namespaces", "services", "configmaps"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  ```

- **Enable network policies:**
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

## Debugging Commands

### 1. Controller Status

```bash
# Check controller pod status
kubectl get pods -n nginx-helm-operator-system

# View controller logs
kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager --tail=100

# Check controller metrics
kubectl top pods -n nginx-helm-operator-system
```

### 2. Custom Resource Status

```bash
# List all NginxDeployments
kubectl get nginxdeployments --all-namespaces

# Describe specific resource
kubectl describe nginxdeployment staging-nginx

# Get resource YAML
kubectl get nginxdeployment staging-nginx -o yaml
```

### 3. Helm Release Status

```bash
# List Helm releases in target namespace
helm list -n staging

# Get release status
helm status staging-app -n staging

# View release history
helm history staging-app -n staging
```

### 4. RBAC Verification

```bash
# Check service account permissions
kubectl auth can-i '*' '*' --as=system:serviceaccount:nginx-helm-operator-system:nginx-helm-operator-controller-manager

# Verify specific permissions
kubectl auth can-i create deployments --as=system:serviceaccount:nginx-helm-operator-system:nginx-helm-operator-controller-manager -n staging
```

## Performance Optimization

### 1. Controller Tuning

```yaml
# Adjust reconciliation frequency
args:
  - --leader-elect
  - --metrics-bind-address=127.0.0.1:8080
  - --health-probe-bind-address=:8081
  - --reconcile-period=5m  # Reduce frequency if needed
```

### 2. Resource Optimization

```yaml
# Optimize resource requests/limits
resources:
  limits:
    cpu: 200m    # Reduce if CPU usage is low
    memory: 256Mi # Adjust based on actual usage
  requests:
    cpu: 100m
    memory: 128Mi
```

### 3. Caching Configuration

```yaml
# Configure client-side caching
env:
- name: KUBE_CACHE_MUTATION_DETECTOR
  value: "true"
- name: GOMAXPROCS
  value: "2"
```

## Security Considerations

### 1. Image Security

```bash
# Scan container image for vulnerabilities
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image docker.io/rohtash672/nginx-helm-operator:latest
```

### 2. Network Security

```yaml
# Implement pod security standards
apiVersion: v1
kind: Pod
metadata:
  annotations:
    seccomp.security.alpha.kubernetes.io/pod: runtime/default
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 65532
    fsGroup: 65532
  containers:
  - name: manager
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      readOnlyRootFilesystem: true
```

### 3. Secret Management

```bash
# Use Kubernetes secrets for sensitive data
kubectl create secret generic nginx-operator-config \
  --from-literal=helm-repo-username=user \
  --from-literal=helm-repo-password=pass \
  -n nginx-helm-operator-system
```

## Troubleshooting Checklist

- [ ] Verify operator pod is running and ready
- [ ] Check controller logs for errors
- [ ] Confirm RBAC permissions are correct
- [ ] Validate CRD is properly installed
- [ ] Ensure target namespace exists
- [ ] Verify Helm chart is accessible
- [ ] Check resource quotas and limits
- [ ] Confirm network connectivity
- [ ] Validate custom resource syntax
- [ ] Monitor resource versions for conflicts

## Support and Maintenance

### Regular Maintenance Tasks

1. **Weekly:**
   - Review controller logs for warnings
   - Check resource usage metrics
   - Verify backup procedures

2. **Monthly:**
   - Update operator image to latest version
   - Review and update RBAC permissions
   - Audit deployed resources

3. **Quarterly:**
   - Performance review and optimization
   - Security audit and updates
   - Documentation updates

### Getting Help

1. **Check logs first:**
   ```bash
   kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager --previous
   ```

2. **Gather debugging information:**
   ```bash
   kubectl describe nginxdeployment <name>
   kubectl get events --sort-by=.metadata.creationTimestamp
   ```

3. **Create support bundle:**
   ```bash
   kubectl cluster-info dump --namespaces nginx-helm-operator-system,staging > support-bundle.yaml
   ```

---

**Note:** This troubleshooting guide should be updated regularly as new issues are discovered and resolved. Always test changes in a non-production environment first.
