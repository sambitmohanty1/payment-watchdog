#!/bin/bash

# Lexure Intelligence MVP - Kubernetes Prerequisites Installation

set -e

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

print_status "🚀 Setting up Kubernetes prerequisites for Lexure Intelligence MVP..."

# Check if kubectl is available
if ! command_exists kubectl; then
    error_exit "kubectl is not installed. Please install kubectl and try again."
fi

# Check if helm is installed
if ! command_exists helm; then
    print_warning "Helm is not installed. Installing Helm..."
    if command_exists brew; then
        brew install helm
    else
        error_exit "Please install Helm: https://helm.sh/docs/intro/install/"
    fi
fi

# Add required Helm repositories
print_status "Adding required Helm repositories..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# Install NGINX Ingress Controller
print_status "Installing NGINX Ingress Controller..."
if ! helm list -n ingress-nginx | grep -q "ingress-nginx"; then
    helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
        --create-namespace \
        --namespace ingress-nginx \
        --set controller.service.type=LoadBalancer \
        --set controller.service.annotations."service\.beta\.kubernetes\.io/aws-load-balancer-type"=nlb
else
    print_status "NGINX Ingress Controller is already installed."
fi

# Install kube-prometheus-stack (includes Prometheus Operator which provides ServiceMonitor CRD)
print_status "Installing kube-prometheus-stack..."
if ! helm list -n monitoring | grep -q "kube-prometheus-stack"; then
    helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
        --create-namespace \
        --namespace monitoring \
        --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
        --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false \
        --set prometheus.prometheusSpec.ruleSelectorNilUsesHelmValues=false \
        --set prometheus.prometheusSpec.probeSelectorNilUsesHelmValues=false
else
    print_status "kube-prometheus-stack is already installed."
fi

print_status "✅ Prerequisites installation completed successfully!"

# Show instructions for next steps
echo -e "\n${GREEN}Next steps:${NC}"
echo "1. Wait for all pods to be ready:"
echo "   watch kubectl get pods -A"
echo "2. Once all pods are ready, run the deployment script:"
echo "   ./deploy.sh"

# Print NGINX Ingress Controller service details
echo -e "\n${GREEN}NGINX Ingress Controller details:${NC}"
kubectl get svc -n ingress-nginx ingress-nginx-controller

# Print Prometheus service details
echo -e "\n${GREEN}Prometheus Stack details:${NC}"
echo "Prometheus UI: kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090"
echo "Grafana UI: kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 8080:80"
echo "Grafana admin password: kubectl get secret -n monitoring kube-prometheus-stack-grafana -o jsonpath='{.data.admin-password}' | base64 --decode"
