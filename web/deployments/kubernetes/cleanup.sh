#!/bin/bash

# Lexure Intelligence UI Cleanup Script
# This script removes the UI deployment from Kubernetes

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🧹 Cleaning up Lexure Intelligence UI from Kubernetes..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Check if we're connected to a cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Not connected to a Kubernetes cluster"
    exit 1
fi

echo "✅ Connected to Kubernetes cluster: $(kubectl config current-context)"

# Remove the deployment
echo "🗑️  Removing UI deployment..."
cd "$SCRIPT_DIR"
kubectl delete -k .

# Wait for resources to be removed
echo "⏳ Waiting for resources to be removed..."
kubectl wait --for=delete --timeout=60s namespace/lexure-intelligence-ui 2>/dev/null || true

echo ""
echo "🎉 UI cleanup completed successfully!"
echo ""
echo "📊 Remaining resources in cluster:"
kubectl get all --all-namespaces | grep -E "(lexure-intelligence|lexure-mvp)" || echo "No lexure resources found"
