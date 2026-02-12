#!/bin/bash

# Payment Watchdog - Kubernetes Deployment Script

set -e

# Configuration
NAMESPACE="lexure"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print status messages
print_status() {
    echo -e "${GREEN}[STATUS]${NC} $1"
}

# Function to print warnings
print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to print errors and exit
error_exit() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
    exit 1
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

echo -e "${GREEN}🚀 Deploying Payment Watchdog to Kubernetes...${NC}"

# Check if kubectl is available
if ! command_exists kubectl; then
    error_exit "kubectl is not installed. Please install kubectl and try again."
fi

# Check if kustomize is installed
if ! command_exists kustomize; then
    print_warning "kustomize is not installed. Installing..."
    if command_exists brew; then
        brew install kustomize
    else
        error_exit "Please install kustomize: https://kubectl.docs.kubernetes.io/installation/kustomize/"
    fi
fi

# Check if connected to a Kubernetes cluster
if ! kubectl cluster-info &> /dev/null; then
    error_exit "Not connected to a Kubernetes cluster. Please configure your kubeconfig."
fi

# Check if minikube is running (for local development)
if command_exists minikube; then
    if ! minikube status --format='{{.Host}}' | grep -q "Running"; then
        echo "🔄 Starting minikube..."
        minikube start
    fi
    echo "✅ Minikube is running"
fi

echo "🔨 Building Docker images..."

# Build MVP backend image
echo "Building MVP backend image..."
cd "${SCRIPT_DIR}/../.."
docker build -t payment-watchdog-mvp:latest .

# Build recovery-orchestration image
echo "Building recovery-orchestration image..."
docker build -t payment-watchdog/recovery-orchestration:latest -f Dockerfile.recovery-orchestration . 2>/dev/null || \
docker build -t payment-watchdog/recovery-orchestration:latest -f ../../recovery-orchestration/Dockerfile . 2>/dev/null || \
print_warning "Recovery orchestration Dockerfile not found, using existing image"

# Build UI image
echo "Building UI image..."
cd "${SCRIPT_DIR}/../../../ui"
docker build -t payment-watchdog-ui:latest .

# Return to deployment directory
cd "${SCRIPT_DIR}"

# Apply kustomize configurations
apply_kustomize() {
    local dir="$1"
    print_status "Applying Kustomize configuration from $dir..."
    
    # First, validate the kustomization
    if ! kustomize build "$dir" > /dev/null; then
        error_exit "Kustomize validation failed for $dir"
    fi
    
    # Then apply
    kubectl apply -k "$dir" --validate=false || error_exit "Failed to apply Kustomize configuration from $dir"
}

# Create namespace if it doesn't exist
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    print_status "Creating namespace $NAMESPACE..."
    kubectl create namespace "$NAMESPACE" || error_exit "Failed to create namespace $NAMESPACE"
else
    print_status "Using existing namespace $NAMESPACE"
fi

# Apply base components
print_status "Deploying base components..."
apply_kustomize "${SCRIPT_DIR}/base"

# Apply applications
print_status "Deploying applications..."
for app in "${SCRIPT_DIR}/apps"/*/; do
    if [ -f "${app}kustomization.yaml" ] || [ -f "${app}kustomization.yml" ] || [ -f "${app}Kustomization" ]; then
        apply_kustomize "$app"
    fi
done

# Wait for deployments to be ready
print_status "Waiting for deployments to be ready..."
for deployment in $(kubectl get deployments -n "$NAMESPACE" -o name); do
    print_status "Waiting for $deployment..."
    kubectl rollout status -n "$NAMESPACE" "$deployment" --timeout=300s || \
        print_warning "$deployment did not become ready in time"
done

# Show deployment status
print_status "\nDeployment completed successfully!"
print_status "Current status:"
echo ""

# Show namespaces
print_status "Namespaces:"
kubectl get namespaces | grep -E "NAME|${NAMESPACE}"
echo ""

# Show pods
print_status "Pods:"
kubectl get pods -n "$NAMESPACE"
echo ""

# Show services
print_status "Services:"
kubectl get svc -n "$NAMESPACE"
echo ""

# Show deployments
print_status "Deployments:"
kubectl get deployments -n "$NAMESPACE"
echo ""

# Show ingress
print_status "Ingress:"
kubectl get ingress -n "$NAMESPACE" 2>/dev/null || print_warning "No ingress found"
echo ""

# Get ingress URL if available
INGRESS_HOST=$(kubectl get ingress -n "$NAMESPACE" -o jsonpath='{.items[0].spec.rules[0].host}' 2>/dev/null || true)
if [ -n "$INGRESS_HOST" ]; then
    print_status "Applications are accessible at:"
    kubectl get ingress -n "$NAMESPACE" -o jsonpath='{range .items[*]}{range .spec.rules[*]}{.host}{end}{end}' | tr ' ' '\n' | sort -u | sed 's/^/  https:\/\//'
else
    print_warning "No ingress found. Services are only accessible within the cluster."
    print_status "To access services locally, use port-forwarding:"
    echo "  # For the main application:"
    echo "  kubectl port-forward svc/payment-watchdog 8085:80 -n $NAMESPACE"
    echo "  # For the recovery orchestration service:"
    echo "  kubectl port-forward svc/recovery-orchestration 8086:80 -n $NAMESPACE"
fi

echo -e "\n${GREEN}All components deployed successfully!"

# Show service URLs
echo ""
echo "🌐 Service URLs:"
echo "  - Main Application (ClusterIP): payment-watchdog.lexure.svc.cluster.local:80"
echo "  - Recovery Orchestration (ClusterIP): recovery-orchestration.lexure.svc.cluster.local:80"
echo "  - PostgreSQL: lexure-postgres.lexure.svc.cluster.local:5432"
echo "  - Redis: lexure-redis.lexure.svc.cluster.local:6379"

# If using minikube, show the external URL
if command -v minikube &> /dev/null; then
    echo ""
    echo "🚀 To access the service externally:"
    echo "  - Run: minikube service payment-watchdog -n lexure"
    echo "  - Or use port forwarding: kubectl port-forward -n lexure svc/payment-watchdog 8080:8080"
fi

echo ""
echo "📝 Useful Commands:"
echo "  - View main app logs: kubectl logs -f deployment/payment-watchdog -n lexure"
echo "  - View recovery orchestration logs: kubectl logs -f deployment/recovery-orchestration -n lexure"
echo "  - Check status: kubectl get all -n lexure"
echo "  - Delete deployment: kubectl delete -k deployments/kubernetes/"
echo "  - Port forward main app: kubectl port-forward -n lexure svc/payment-watchdog 8085:8085"
echo "  - Port forward recovery orchestration: kubectl port-forward -n lexure svc/recovery-orchestration 8086:80"

echo ""
echo "🎉 Deployment completed successfully!"
echo "The Payment Watchdog service is now running in Kubernetes."
