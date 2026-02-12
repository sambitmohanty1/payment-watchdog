#!/bin/bash

# Lexure Intelligence UI Deployment Script
# This script deploys the UI to Kubernetes

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

echo "🚀 Deploying Lexure Intelligence UI to Kubernetes..."

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

# Build the Docker image
echo "🔨 Building Docker image..."
cd "$PROJECT_ROOT"
docker build -t lexure-intelligence-ui:latest .

# Load image into minikube if using minikube
if kubectl config current-context | grep -q "minikube"; then
    echo "📦 Loading image into minikube..."
    minikube image load lexure-intelligence-ui:latest
fi

# Deploy to Kubernetes
echo "🚀 Deploying to Kubernetes..."
cd "$SCRIPT_DIR"
kubectl apply -k .

# Wait for deployment to be ready
echo "⏳ Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/lexure-intelligence-ui -n lexure-intelligence-ui

# Get service information
echo "📊 Service Status:"
kubectl get svc -n lexure-intelligence-ui

# Get pod status
echo "📊 Pod Status:"
kubectl get pods -n lexure-intelligence-ui

# Get ingress information
echo "📊 Ingress Status:"
kubectl get ingress -n lexure-intelligence-ui

echo ""
echo "🎉 UI deployment completed successfully!"
echo ""
echo "🌐 Access URLs:"
echo "   - Local: http://localhost/ui (if using port-forward)"
echo "   - Cluster: http://ui.lexure-intelligence.local"
echo ""
echo "🔧 Useful commands:"
echo "   - View logs: kubectl logs -f deployment/lexure-intelligence-ui -n lexure-intelligence-ui"
echo "   - Port forward: kubectl port-forward -n lexure-intelligence-ui svc/lexure-intelligence-ui 3001:80"
echo "   - Delete deployment: kubectl delete -k ."
echo ""
echo "📚 For more information, see the README.md file"
