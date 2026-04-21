# nginx-helm-operator Build and Deployment Instructions

## Overview
This document provides complete step-by-step instructions for building, pushing, and deploying the nginx-helm-operator with the corrected configuration.

## Problem Solved
The original deployment was failing with the error:
```
flag provided but not defined: -leader-elect-lease-duration
```

This was caused by the Helm chart deployment template trying to pass unsupported leader election flags to the manager binary. The main.go file only supports basic `--leader-elect` flag, not the additional duration/deadline/retry flags.

## Prerequisites

### Required Tools
- Docker (for building and pushing images)
- Kubernetes cluster (for deployment)
- Helm 3.x (for chart deployment)
- kubectl (for cluster interaction)
- Go 1.21+ (for local development)

### Registry Access
- Ensure you have push access to `docker.io/rohtash672/nginx-helm-operator`
- Login to Docker registry: `docker login docker.io`

## Build Instructions

### Step 1: Clean Build Environment
```bash
# Ensure you're in the project root directory
cd /path/to/nginx-helm-operator

# Clean any previous builds
docker system prune -f
```

### Step 2: Build Docker Image
```bash
# Build the Docker image with the corrected configuration
docker build -t docker.io/rohtash672/nginx-helm-operator:latest .

# Alternative with specific tag
docker build -t docker.io/rohtash672/nginx-helm-operator:v1.0.1 .
```

### Step 3: Verify Build
```bash
# List the built image
docker images | grep nginx-helm-operator

# Inspect the image
docker inspect docker.io/rohtash672/nginx-helm-operator:latest
```

## Push Instructions

### Step 1: Login to Registry
```bash
# Login to Docker Hub (if not already logged in)
docker login docker.io
```

### Step 2: Push Image
```bash
# Push the latest tag
docker push docker.io/rohtash672/nginx-helm-operator:latest

# If you built with a specific version tag
docker push docker.io/rohtash672/nginx-helm-operator:v1.0.1
```

### Step 3: Verify Push
```bash
# Verify the image is available in the registry
docker pull docker.io/rohtash672/nginx-helm-operator:latest
```

## Deployment Instructions

### Option 1: Deploy Using Helm Chart (Recommended)

#### Step 1: Deploy the Operator
```bash
# Install the nginx-helm-operator using Helm
helm install nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system \
  --create-namespace

# Alternative with custom values
helm install nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system \
  --create-namespace \
  --set image.tag=v1.0.1
```

#### Step 2: Verify Deployment
```bash
# Check if the operator pod is running
kubectl get pods -n nginx-helm-operator-system

# Check operator logs
kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager

# Verify CRD installation
kubectl get crd nginxdeployments.deploy.example.com
```

### Option 2: Deploy Using Kustomize (Alternative)

```bash
# Build and deploy using kustomize
kustomize build config/default | kubectl apply -f -

# Or using kubectl with kustomize
kubectl apply -k config/default
```

## Testing the Deployment

### Step 1: Apply Sample CRD
```bash
# Apply the sample nginx deployment
kubectl apply -f helm-crd.yml

# Or apply the sample from config
kubectl apply -f config/samples/nginx-deployment-sample.yaml
```

### Step 2: Verify CRD Processing
```bash
# Check the NginxDeployment resource
kubectl get nginxdeployments -n default

# Check if the operator created the nginx deployment
kubectl get deployments -n staging

# Check operator logs for processing
kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager -f
```

## Troubleshooting

### Common Issues

1. **Image Pull Errors**
   ```bash
   # Check if image exists
   docker pull docker.io/rohtash672/nginx-helm-operator:latest
   
   # Verify image pull secrets
   kubectl get secrets -n nginx-helm-operator-system
   ```

2. **RBAC Permission Issues**
   ```bash
   # Check service account
   kubectl get serviceaccount -n nginx-helm-operator-system
   
   # Check cluster role bindings
   kubectl get clusterrolebinding | grep nginx-helm-operator
   ```

3. **CRD Issues**
   ```bash
   # Reinstall CRDs
   kubectl delete crd nginxdeployments.deploy.example.com
   helm upgrade nginx-helm-operator ./helm-chart/nginx-helm-operator
   ```

### Debugging Commands

```bash
# Get detailed pod information
kubectl describe pod -n nginx-helm-operator-system -l control-plane=controller-manager

# Check events
kubectl get events -n nginx-helm-operator-system --sort-by=.metadata.creationTimestamp

# Check operator configuration
kubectl get deployment -n nginx-helm-operator-system -o yaml
```

## Cleanup

### Remove Operator
```bash
# Uninstall using Helm
helm uninstall nginx-helm-operator -n nginx-helm-operator-system

# Remove namespace
kubectl delete namespace nginx-helm-operator-system

# Remove CRDs (if needed)
kubectl delete crd nginxdeployments.deploy.example.com
```

### Remove Sample Deployments
```bash
# Remove sample CRD instances
kubectl delete -f helm-crd.yml
kubectl delete -f config/samples/nginx-deployment-sample.yaml
```

## Configuration Changes Made

The following changes were made to fix the deployment issue:

1. **helm-chart/nginx-helm-operator/templates/deployment.yaml**
   - Removed unsupported leader election flags:
     - `--leader-elect-lease-duration`
     - `--leader-elect-renew-deadline`
     - `--leader-elect-retry-period`
   - Kept only the supported `--leader-elect` flag

2. **helm-chart/nginx-helm-operator/values.yaml**
   - Simplified leader election configuration
   - Removed unsupported duration parameters
   - Kept only the `enabled` flag

## Next Steps

1. Test the operator with various nginx deployment configurations
2. Monitor operator performance and resource usage
3. Implement additional features like scaling and updates
4. Add comprehensive monitoring and alerting
5. Create additional Helm chart values for different environments

## Support

For issues or questions:
1. Check operator logs: `kubectl logs -n nginx-helm-operator-system -l control-plane=controller-manager`
2. Review Kubernetes events: `kubectl get events -n nginx-helm-operator-system`
3. Verify RBAC permissions and CRD installation
4. Ensure the Docker image is accessible from your cluster