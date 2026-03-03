#!/bin/bash
set -e

# Change to script directory
cd "$(dirname "$0")"

if [ ! -f .env ]; then
  echo "Error: .env file not found. Please copy .env.example to .env and configure it."
  exit 1
fi

# Export all variables from .env
set -a
source .env
set +a

echo "==================================================="
echo " Starting Payment Watchdog OCI Deployment..."
echo "==================================================="

# 1. Provision Infrastructure
echo "[1/3] Provisioning OCI Infrastructure via Terraform..."
cd terraform
terraform init
terraform apply -auto-approve

# Get the Kubeconfig from the server
SERVER_IP=$(terraform output -raw k3s_server_public_ip)
echo "Infrastructure deployed. Control plane IP: $SERVER_IP"

cd ..

# Wait for K3s to be ready
echo "Waiting for K3s to be ready on the server..."

# Implement polling for kubeconfig
max_retries=30
retry_count=0
kubeconfig_fetched=false

SSH_KEY_PATH=${TF_VAR_ssh_private_key_path}
mkdir -p ~/.kube

while [ $retry_count -lt $max_retries ]; do
    echo "Attempting to fetch kubeconfig (Attempt $((retry_count + 1))/$max_retries)..."
    if ssh -o StrictHostKeyChecking=no -i "$SSH_KEY_PATH" opc@"$SERVER_IP" "sudo cat /etc/rancher/k3s/k3s.yaml 2>/dev/null" > ~/.kube/oci_k3s.yaml 2>/dev/null; then
        if [ -s ~/.kube/oci_k3s.yaml ]; then
            echo "Successfully fetched kubeconfig!"
            kubeconfig_fetched=true
            break
        fi
    fi
    echo "K3s not ready yet. Retrying in 10 seconds..."
    sleep 10
    retry_count=$((retry_count + 1))
done

if [ "$kubeconfig_fetched" = false ]; then
    echo "Error: Failed to fetch kubeconfig after $max_retries attempts. K3s may have failed to install."
    exit 1
fi

sed -i "s/127.0.0.1/$SERVER_IP/g" ~/.kube/oci_k3s.yaml

export KUBECONFIG=~/.kube/oci_k3s.yaml

# Wait for nodes to be ready
echo "Waiting for Kubernetes nodes to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=300s

# Install cert-manager (Traefik is already installed by K3s)
echo "Installing cert-manager..."
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.1/cert-manager.yaml

# Wait for cert-manager to be ready
echo "Waiting for cert-manager pods to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager --timeout=300s

# Create namespace
kubectl create namespace payment-watchdog --dry-run=client -o yaml | kubectl apply -f -

# 3. Apply Kubernetes Manifests
echo "[3/3] Generating and Applying Kubernetes Manifests..."
cd kubernetes

# Substitute env vars into kustomize files
envsubst < ingress.yaml > ingress-processed.yaml
envsubst < secret.yaml > secret-processed.yaml
envsubst < cluster-issuer.yaml > cluster-issuer-processed.yaml

# Create a temporary kustomization file to use the processed files safely
cp kustomization.yaml kustomization.yaml.bak
sed -i 's/ingress.yaml/ingress-processed.yaml/g' kustomization.yaml
sed -i 's/secret.yaml/secret-processed.yaml/g' kustomization.yaml
sed -i 's/cluster-issuer.yaml/cluster-issuer-processed.yaml/g' kustomization.yaml

# Apply the kustomize
kubectl apply -k .

# Restore original kustomization file and clean up
mv kustomization.yaml.bak kustomization.yaml
rm *-processed.yaml

echo "==================================================="
echo " Deployment Complete!"
echo " Ensure your DNS for $DOMAIN is pointed to: $SERVER_IP"
echo " Access your cluster using: export KUBECONFIG=~/.kube/oci_k3s.yaml"
echo "==================================================="
