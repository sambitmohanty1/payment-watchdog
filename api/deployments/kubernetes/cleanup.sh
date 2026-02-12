#!/bin/bash

# Lexure Intelligence MVP - Kubernetes Cleanup Script

set -e

# Configuration
NAMESPACE="lexure-mvp"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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

print_status "🧹 Cleaning up Lexure Intelligence MVP Kubernetes resources..."

# Check if kubectl is available
if ! command_exists kubectl; then
    error_exit "kubectl is not installed. Please install kubectl and try again."
fi

# Check if connected to a Kubernetes cluster
if ! kubectl cluster-info &> /dev/null; then
    error_exit "Not connected to a Kubernetes cluster. Please configure your kubeconfig."
fi

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    print_warning "Namespace $NAMESPACE does not exist. Nothing to clean up."
    exit 0
fi

print_status "Deleting all resources in namespace $NAMESPACE..."

# Delete applications first
print_status "Deleting applications..."
for app in "${SCRIPT_DIR}/apps"/*/; do
    if [ -f "${app}kustomization.yaml" ] || [ -f "${app}kustomization.yml" ] || [ -f "${app}Kustomization" ]; then
        app_name=$(basename "$app")
        print_status "Deleting $app_name..."
        kubectl delete -k "$app" --ignore-not-found=true || true
    fi
done

# Delete base components
print_status "Deleting base components..."
kubectl delete -k "${SCRIPT_DIR}/base" --ignore-not-found=true || true

# Delete any remaining resources in the namespace
print_status "Deleting any remaining resources in namespace $NAMESPACE..."
kubectl delete all --all -n "$NAMESPACE" --ignore-not-found=true
kubectl delete pvc --all -n "$NAMESPACE" --ignore-not-found=true
kubectl delete configmap,secret,ingress,serviceaccount,role,rolebinding \
    -l "app.kubernetes.io/part-of=payment-watchdog" -n "$NAMESPACE" --ignore-not-found=true

# Delete the namespace if it's empty
if [ "$(kubectl get all -n "$NAMESPACE" 2>/dev/null | wc -l)" -le 1 ]; then
    print_status "Deleting namespace $NAMESPACE..."
    kubectl delete namespace "$NAMESPACE" --ignore-not-found=true
else
    print_warning "Namespace $NAMESPACE is not empty. Some resources may still exist."
    print_status "To force delete the namespace, run:"
    echo "  kubectl delete namespace $NAMESPACE --force --grace-period=0"
fi

print_status "✅ Cleanup completed successfully!"

# Wait for namespace to be deleted
kubectl wait --for=delete namespace/lexure-mvp --timeout=300s 2>/dev/null || true

echo "🧹 Cleaning up local storage directories..."

# Clean up local storage directories (if using hostPath volumes)
if [ -d "/tmp/lexure-mvp-postgres" ]; then
    echo "Removing PostgreSQL data directory..."
    sudo rm -rf /tmp/lexure-mvp-postgres
fi

if [ -d "/tmp/lexure-mvp-redis" ]; then
    echo "Removing Redis data directory..."
    sudo rm -rf /tmp/lexure-mvp-redis
fi

echo "✅ Cleanup completed successfully!"
echo "All Kubernetes resources and local storage have been removed."
