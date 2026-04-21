# Nginx Helm Operator Helm Chart

This Helm chart deploys the Nginx Helm Operator, a Kubernetes operator that manages nginx deployments through Custom Resource Definitions (CRDs) and Helm charts.

## Overview

The Nginx Helm Operator allows you to deploy and manage nginx instances declaratively using Kubernetes custom resources. The operator watches for `NginxDeployment` resources and automatically deploys nginx using Helm charts with the specified configuration.

## Features

- **Declarative Management**: Deploy nginx instances using Kubernetes custom resources
- **Helm Integration**: Uses Helm charts for nginx deployment management
- **Auto-scaling**: Configure replica counts for nginx deployments
- **Custom Configuration**: Override Helm chart values through custom resource specs
- **Namespace Management**: Deploy nginx instances to specific namespaces
- **RBAC Security**: Comprehensive role-based access control configuration
- **Health Monitoring**: Built-in health checks and metrics endpoints
- **Leader Election**: High availability with leader election support

## Prerequisites

- Kubernetes 1.19+
- Helm 3.8+
- kubectl configured to communicate with your cluster

## Installation

### Quick Start

```bash
# Add the helm repository (if available)
# helm repo add nginx-helm-operator https://your-repo-url
# helm repo update

# Install the operator
helm install nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system \
  --create-namespace
```

### Custom Installation

```bash
# Install with custom values
helm install nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system \
  --create-namespace \
  --set image.tag=v1.0.0 \
  --set operator.replicaCount=2
```

### Installation with Custom Values File

```bash
# Create custom values file
cat > custom-values.yaml <<EOF
image:
  repository: your-registry/nginx-helm-operator
  tag: "v1.0.0"

operator:
  replicaCount: 2
  resources:
    limits:
      cpu: 1000m
      memory: 256Mi

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
EOF

# Install with custom values
helm install nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system \
  --create-namespace \
  --values custom-values.yaml
```

## Configuration

### Values

The following table lists the configurable parameters and their default values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Operator image repository | `rohtash672/nginx-helm-operator` |
| `image.tag` | Operator image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `operator.replicaCount` | Number of operator replicas | `1` |
| `operator.resources.limits.cpu` | CPU limit | `500m` |
| `operator.resources.limits.memory` | Memory limit | `128Mi` |
| `operator.resources.requests.cpu` | CPU request | `10m` |
| `operator.resources.requests.memory` | Memory request | `64Mi` |
| `serviceAccount.create` | Create service account | `true` |
| `rbac.create` | Create RBAC resources | `true` |
| `crd.create` | Create CRDs | `true` |
| `namespace.create` | Create namespace | `true` |
| `leaderElection.enabled` | Enable leader election | `true` |
| `healthCheck.enabled` | Enable health checks | `true` |
| `metrics.enabled` | Enable metrics | `true` |
| `logging.level` | Log level | `info` |
| `logging.format` | Log format | `json` |

### Advanced Configuration

#### Security Context

```yaml
operator:
  securityContext:
    runAsNonRoot: true
    runAsUser: 65532
    fsGroup: 65532
  containerSecurityContext:
    allowPrivilegeEscalation: false
    readOnlyRootFilesystem: true
    runAsNonRoot: true
    capabilities:
      drop:
        - ALL
```

#### Resource Limits

```yaml
operator:
  resources:
    limits:
      cpu: 1000m
      memory: 256Mi
    requests:
      cpu: 50m
      memory: 128Mi
```

#### Monitoring

```yaml
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
    scrapeTimeout: 10s
```

## Usage

### Creating an Nginx Deployment

After installing the operator, you can create nginx deployments using the `NginxDeployment` custom resource:

```yaml
apiVersion: deploy.example.com/v1alpha1
kind: NginxDeployment
metadata:
  name: my-nginx
  namespace: default
spec:
  deploymentName: my-nginx-app
  image: nginx:1.24
  replicas: 3
  namespace: production
  helmValues:
    service:
      type: LoadBalancer
    ingress:
      enabled: true
      hosts:
        - host: my-nginx.example.com
          paths:
            - path: /
              pathType: Prefix
```

### Applying the Resource

```bash
kubectl apply -f nginx-deployment.yaml
```

### Checking Status

```bash
# Check the NginxDeployment status
kubectl get nginxdeployments

# Get detailed information
kubectl describe nginxdeployment my-nginx

# Check operator logs
kubectl logs -n nginx-helm-operator-system deployment/nginx-helm-operator-controller-manager
```

## Troubleshooting

### Common Issues

#### 1. Operator Pod Not Starting

```bash
# Check pod status
kubectl get pods -n nginx-helm-operator-system

# Check pod logs
kubectl logs -n nginx-helm-operator-system deployment/nginx-helm-operator-controller-manager

# Check events
kubectl get events -n nginx-helm-operator-system
```

#### 2. RBAC Permission Issues

```bash
# Check if RBAC resources are created
kubectl get clusterrole | grep nginx-helm-operator
kubectl get clusterrolebinding | grep nginx-helm-operator

# Check service account
kubectl get serviceaccount -n nginx-helm-operator-system
```

#### 3. CRD Not Available

```bash
# Check if CRD is installed
kubectl get crd nginxdeployments.deploy.example.com

# Check CRD details
kubectl describe crd nginxdeployments.deploy.example.com
```

#### 4. NginxDeployment Stuck

```bash
# Check NginxDeployment status
kubectl get nginxdeployments -A

# Check operator logs for errors
kubectl logs -n nginx-helm-operator-system deployment/nginx-helm-operator-controller-manager --tail=100

# Check if target namespace exists
kubectl get namespace <target-namespace>
```

### Debug Mode

Enable debug logging:

```yaml
logging:
  level: debug
  development: true
```

## Upgrading

### Upgrade the Operator

```bash
# Upgrade to a new version
helm upgrade nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system
```

### Upgrade with New Values

```bash
# Upgrade with new configuration
helm upgrade nginx-helm-operator ./helm-chart/nginx-helm-operator \
  --namespace nginx-helm-operator-system \
  --set image.tag=v2.0.0 \
  --reuse-values
```

## Uninstalling

### Remove the Operator

```bash
# Uninstall the Helm release
helm uninstall nginx-helm-operator --namespace nginx-helm-operator-system

# Clean up CRDs (if needed)
kubectl delete crd nginxdeployments.deploy.example.com

# Clean up namespace (if needed)
kubectl delete namespace nginx-helm-operator-system
```

### Clean Up NginxDeployments

```bash
# Remove all NginxDeployment resources
kubectl delete nginxdeployments --all --all-namespaces
```

## Development

### Building the Operator Image

```bash
# Build and push the operator image
docker build -t your-registry/nginx-helm-operator:latest .
docker push your-registry/nginx-helm-operator:latest
```

### Testing

```bash
# Run Helm chart tests
helm test nginx-helm-operator --namespace nginx-helm-operator-system

# Validate templates
helm template nginx-helm-operator ./helm-chart/nginx-helm-operator --debug

# Lint the chart
helm lint ./helm-chart/nginx-helm-operator
```

## Security Considerations

- The operator runs with minimal privileges using a non-root user
- RBAC is configured with least-privilege access
- Security contexts are enforced for all containers
- Network policies can be enabled for additional security

## Support

For issues and questions:

1. Check the [troubleshooting section](#troubleshooting)
2. Review operator logs
3. Check Kubernetes events
4. Open an issue in the project repository

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please read the contributing guidelines before submitting pull requests.

## Changelog

### v0.1.0
- Initial release
- Basic nginx deployment management
- Helm chart integration
- RBAC configuration
- Health monitoring
- Leader election support