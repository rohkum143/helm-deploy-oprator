#!/bin/bash

# Comprehensive test script for nginx-helm-operator deployment functionality
# This script validates that the operator correctly creates Kubernetes deployments and pods

set -e

echo "=== nginx-helm-operator Deployment Functionality Test Suite ==="
echo "Start time: $(date)"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Helper functions
log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((FAILED_TESTS++))
}

wait_for_deployment() {
    local deployment_name=$1
    local namespace=$2
    local timeout=${3:-300}
    
    echo "Waiting for deployment $deployment_name in namespace $namespace to be ready..."
    kubectl wait --for=condition=available --timeout=${timeout}s deployment/$deployment_name -n $namespace
}

wait_for_nginxdeployment() {
    local nginx_deployment_name=$1
    local timeout=${2:-300}
    local max_attempts=$((timeout / 10))
    local attempt=0
    
    echo "Waiting for NginxDeployment $nginx_deployment_name to reach Deployed phase..."
    while [ $attempt -lt $max_attempts ]; do
        phase=$(kubectl get nginxdeployment $nginx_deployment_name -o jsonpath='{.status.phase}' 2>/dev/null || echo "")
        if [ "$phase" = "Deployed" ]; then
            return 0
        elif [ "$phase" = "Failed" ]; then
            echo "NginxDeployment $nginx_deployment_name failed"
            return 1
        fi
        sleep 10
        ((attempt++))
    done
    echo "Timeout waiting for NginxDeployment $nginx_deployment_name"
    return 1
}

cleanup_test() {
    local test_name=$1
    echo "Cleaning up test: $test_name"
    kubectl delete nginxdeployment $test_name --ignore-not-found=true
    # Wait a bit for cleanup
    sleep 5
}

run_test() {
    local test_name=$1
    local test_description="$2"
    ((TOTAL_TESTS++))
    
    log_info "Running Test $TOTAL_TESTS: $test_description"
    
    # Apply the test
    if ! kubectl apply -f tests/test-deployment-functionality.yaml | grep "$test_name"; then
        log_error "Failed to apply test $test_name"
        return 1
    fi
    
    # Wait for NginxDeployment to be processed
    if wait_for_nginxdeployment $test_name; then
        # Check if actual Kubernetes deployment was created
        deployment_name=$(kubectl get nginxdeployment $test_name -o jsonpath='{.spec.deploymentName}')
        target_namespace=$(kubectl get nginxdeployment $test_name -o jsonpath='{.spec.namespace}')
        
        if kubectl get deployment $deployment_name -n $target_namespace >/dev/null 2>&1; then
            if wait_for_deployment $deployment_name $target_namespace; then
                # Verify pods are running
                pod_count=$(kubectl get pods -n $target_namespace -l app.kubernetes.io/name=$deployment_name --field-selector=status.phase=Running --no-headers | wc -l)
                expected_replicas=$(kubectl get nginxdeployment $test_name -o jsonpath='{.spec.replicas}')
                
                if [ "$pod_count" -eq "$expected_replicas" ]; then
                    log_success "Test $test_name: $pod_count/$expected_replicas pods running successfully"
                else
                    log_error "Test $test_name: Expected $expected_replicas pods, but only $pod_count are running"
                fi
            else
                log_error "Test $test_name: Deployment $deployment_name not ready in time"
            fi
        else
            log_error "Test $test_name: Kubernetes deployment $deployment_name not found in namespace $target_namespace"
        fi
    else
        log_error "Test $test_name: NginxDeployment did not reach Deployed phase"
    fi
    
    # Cleanup
    cleanup_test $test_name
    echo
}

# Setup test environment
log_info "Setting up test environment..."
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: operator-test
EOF

# Run individual tests
echo "Starting deployment functionality tests..."
echo

# Test 1: Basic deployment functionality
run_test "test-basic-deployment" "Basic nginx deployment with minimal configuration"

# Test 2: Deployment with custom resources
run_test "test-custom-resources" "Deployment with custom resource limits and requests"

# Test 3: Deployment with affinity and tolerations
run_test "test-affinity-tolerations" "Deployment with node affinity and tolerations"

# Test 4: Deployment with custom nginx configuration
run_test "test-custom-config" "Deployment with custom nginx configuration"

# Test 5: Deployment with multiple replicas and load balancing
run_test "test-load-balancing" "Deployment with multiple replicas and load balancer service"

# Additional validation tests
log_info "Running additional validation tests..."

# Test operator health
((TOTAL_TESTS++))
log_info "Test $TOTAL_TESTS: Operator health check"
if kubectl get deployment nginx-helm-operator-controller-manager -n nginx-helm-operator-system >/dev/null 2>&1; then
    if kubectl rollout status deployment/nginx-helm-operator-controller-manager -n nginx-helm-operator-system --timeout=30s >/dev/null 2>&1; then
        log_success "Operator is healthy and running"
    else
        log_error "Operator deployment is not ready"
    fi
else
    log_error "Operator deployment not found"
fi

# Test CRD availability
((TOTAL_TESTS++))
log_info "Test $TOTAL_TESTS: CRD availability check"
if kubectl get crd nginxdeployments.deploy.example.com >/dev/null 2>&1; then
    log_success "NginxDeployment CRD is available"
else
    log_error "NginxDeployment CRD not found"
fi

# Cleanup test namespace
log_info "Cleaning up test environment..."
kubectl delete namespace operator-test --ignore-not-found=true

# Test summary
echo
echo "=== Test Summary ==="
echo "Total tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo "Success rate: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
echo "End time: $(date)"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi