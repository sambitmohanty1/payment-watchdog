#!/bin/bash

# Oracle Cloud Free Tier Setup Script for Payment Watchdog
# This script automates the setup of Oracle Cloud infrastructure

set -e
set -u

# Ensure USER variable is always defined
USER=${USER:-$(whoami)}

REGION=${REGION:-ap-melbourne-1}

if [[ "$REGION" != "ap-sydney-1" && "$REGION" != "ap-melbourne-1" ]]; then
    echo "❌ ERROR: For sovereign compliance, REGION must be 'ap-sydney-1' or 'ap-melbourne-1'."
    exit 1
fi

echo "🚀 Oracle Cloud Free Tier Setup for Payment Watchdog in region: $REGION"
echo "================================================"

# Check for force flag to bypass duplicate warnings
FORCE_MODE=${FORCE:-false}
if [ "$FORCE_MODE" = "true" ]; then
    echo "🔧 FORCE MODE: Will create resources even if they exist"
    echo ""
fi

# Configuration
COMPARTMENT_NAME="PaymentWatchdogSovereign"
CLUSTER_NAME="payment-watchdog-cluster"
NODE_POOL_NAME="payment-watchdog-nodes"
VCN_NAME="payment-watchdog-vcn"
SUBNET_NAME="payment-watchdog-subnet"
IGW_NAME="payment-watchdog-igw"
DOMAIN=${DOMAIN:-""}  # Optional - leave empty for IP-based access
TENANCY_ID="ocid1.tenancy.oc1..aaaaaaaaitfhbb6ix7yqiavspixzepzm4babf6qjonzom5pe4lvqzxvp2xla"
KUBERNETES_VERSION=${KUBERNETES_VERSION:-"v1.34.2"}

# Region-specific network configuration
if [[ "$REGION" == "ap-sydney-1" ]]; then
    VCN_CIDR="10.0.0.0/16"
    SUBNET_CIDR="10.0.1.0/24"
    POD_CIDR="10.244.0.0/16"
    SERVICE_CIDR="10.96.0.0/12"
elif [[ "$REGION" == "ap-melbourne-1" ]]; then
    VCN_CIDR="10.1.0.0/16"
    SUBNET_CIDR="10.1.1.0/24"
    POD_CIDR="10.245.0.0/16"
    SERVICE_CIDR="10.97.0.0/12"
else
    echo "❌ ERROR: Unsupported region: $REGION"
    exit 1
fi

# -------------------------------------------------------
# Preflight checks
# -------------------------------------------------------
echo "🔍 Running preflight checks..."

# Check for existing resources to prevent duplicates
echo "🔍 Checking for existing resources in compartment..."

EXISTING_VCN=$(oci network vcn list --compartment-id "$COMPARTMENT_ID" --query "data[?name=='$VCN_NAME'].id | [0]" --raw-output)
EXISTING_CLUSTER=$(oci ce cluster list --compartment-id "$COMPARTMENT_ID" --query "data[?name=='$CLUSTER_NAME'].id | [0]" --raw-output)

if [ -n "$EXISTING_VCN" ] && [ "$EXISTING_VCN" != "null" ]; then
    echo "⚠️  WARNING: VCN '$VCN_NAME' already exists in compartment"
    echo "   Existing VCN ID: $EXISTING_VCN"
fi

if [ -n "$EXISTING_CLUSTER" ] && [ "$EXISTING_CLUSTER" != "null" ]; then
    echo "⚠️  WARNING: Cluster '$CLUSTER_NAME' already exists in compartment"
    echo "   Existing Cluster ID: $EXISTING_CLUSTER"
fi

# Ask for confirmation if resources exist (unless force mode)
if [ "$FORCE_MODE" != "true" ]; then
    if [ -n "$EXISTING_VCN" ] || [ -n "$EXISTING_CLUSTER" ]; then
        echo ""
        read -p "⚠️  Existing resources found. Do you want to continue and potentially create duplicates? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "❌ Setup cancelled by user."
            exit 0
        fi
    fi
else
    echo "🔧 Skipping duplicate checks due to FORCE_MODE=true"
fi

for cmd in oci kubectl helm; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "❌ ERROR: Required CLI '$cmd' is not installed or not in PATH."
        exit 1
    fi
done

if ! oci iam region list &>/dev/null; then
    echo "❌ ERROR: OCI CLI is not authenticated. Run 'oci setup config' first."
    exit 1
fi

echo "✅ Preflight checks passed"

# -------------------------------------------------------
# Cleanup trap - logs created resources on failure
# -------------------------------------------------------
CREATED_RESOURCES=()

cleanup() {
    if [ ${#CREATED_RESOURCES[@]} -gt 0 ]; then
        echo ""
        echo "⚠️  Script exited with an error. The following resources were created and may need manual cleanup:"
        for resource in "${CREATED_RESOURCES[@]}"; do
            echo "   - $resource"
        done
    fi
}

trap cleanup ERR

# -------------------------------------------------------
# Step 1: Creating Compartment
# -------------------------------------------------------
echo " Step 1: Creating# Get compartment ID"
COMPARTMENT_ID=$(oci iam compartment list \
    --compartment-id "$TENANCY_ID" \
    --query "data[?name=='$COMPARTMENT_NAME'].id | [0]" \
    --raw-output)

# Export for global availability
export COMPARTMENT_ID

if [ -z "$COMPARTMENT_ID" ] || [ "$COMPARTMENT_ID" = "null" ]; then
    COMPARTMENT_ID=$(oci iam compartment create \
        --name "$COMPARTMENT_NAME" \
        --description "Isolated Sovereign Compartment" \
        --compartment-id "$TENANCY_ID" \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Compartment: $COMPARTMENT_ID")
    echo "✅ Created compartment: $COMPART_ID"

    echo "⏳ Waiting for IAM compartment to propagate..."
    sleep 15
else
    echo "✅ Using existing compartment: $COMPARTMENT_ID"
fi

# -------------------------------------------------------
# Step 2: Creating VCN
# -------------------------------------------------------
echo "📋 Step 2: Creating VCN..."
VCN_ID=$(oci network vcn list \
    --compartment-id "$COMPARTMENT_ID" \
    --query "data[?name=='$VCN_NAME'].id | [0]" \
    --raw-output)

if [ -z "$VCN_ID" ] || [ "$VCN_ID" = "null" ]; then
    VCN_ID=$(oci network vcn create \
        --compartment-id "$COMPARTMENT_ID" \
        --cidr-block "$VCN_CIDR" \
        --display-name "$VCN_NAME" \
        --freeform-tags '{"project":"payment-watchdog"}' \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("VCN: $VCN_ID")
    echo "✅ Created VCN: $VCN_ID"
else
    echo "✅ Using existing VCN: $VCN_ID"
fi

# -------------------------------------------------------
# Step 3: Creating Subnet
# -------------------------------------------------------
echo "📋 Step 3: Creating Subnet..."
SUBNET_ID=$(oci network subnet list \
    --compartment-id "$COMPARTMENT_ID" \
    --vcn-id "$VCN_ID" \
    --query "data[?name=='$SUBNET_NAME'].id | [0]" \
    --raw-output)

if [ -z "$SUBNET_ID" ] || [ "$SUBNET_ID" = "null" ]; then
    SUBNET_ID=$(oci network subnet create \
        --compartment-id "$COMPARTMENT_ID" \
        --vcn-id "$VCN_ID" \
        --cidr-block "$SUBNET_CIDR" \
        --display-name "$SUBNET_NAME" \
        --freeform-tags '{"project":"payment-watchdog"}' \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Subnet: $SUBNET_ID")
    echo "✅ Created Subnet: $SUBNET_ID"
else
    echo "✅ Using existing Subnet: $SUBNET_ID"
fi

# -------------------------------------------------------
# Step 4: Creating Internet Gateway
# -------------------------------------------------------
echo "📋 Step 4: Creating Internet Gateway..."
IGW_ID=$(oci network internet-gateway list \
    --compartment-id "$COMPARTMENT_ID" \
    --vcn-id "$VCN_ID" \
    --query "data[?name=='$IGW_NAME'].id | [0]" \
    --raw-output)

if [ -z "$IGW_ID" ] || [ "$IGW_ID" = "null" ]; then
    IGW_ID=$(oci network internet-gateway create \
        --compartment-id "$COMPARTMENT_ID" \
        --vcn-id "$VCN_ID" \
        --display-name "$IGW_NAME" \
        --is-enabled true \
        --freeform-tags '{"project":"payment-watchdog"}' \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Internet Gateway: $IGW_ID")
    echo "✅ Created Internet Gateway: $IGW_ID"

    # Fetch the VCN ID from the gateway to ensure correctness
    CORRECT_VCN_ID=$(oci network internet-gateway get \
        --ig-id "$IGW_ID" \
        --query 'data."vcn-id"' --raw-output)

    RT_ID=$(oci network route-table list \
        --compartment-id "$COMPARTMENT_ID" \
        --vcn-id "$CORRECT_VCN_ID" \
        --query "data[0].id" --raw-output)

    # Fetch existing route rules and merge the internet gateway route
    EXISTING_RULES=$(oci network route-table get \
        --rt-id "$RT_ID" \
        --query 'data."route-rules"' --raw-output)

    # Build merged route rules - append IGW default route to existing rules
    MERGED_RULES=$(echo "$EXISTING_RULES" | python3 -c "
import json, sys
rules = json.load(sys.stdin) or []
igw_route = {'cidrBlock': '0.0.0.0/0', 'networkEntityId': '$IGW_ID'}
if not any(r.get('cidrBlock') == '0.0.0.0/0' for r in rules):
    rules.append(igw_route)
print(json.dumps(rules))
")

    echo "$MERGED_RULES" > /tmp/route_rules.json

    oci network route-table update \
        --rt-id "$RT_ID" \
        --route-rules file:///tmp/route_rules.json \
        --force

    echo "✅ Updated route table with internet gateway route"
    rm -f /tmp/route_rules.json
else
    echo "✅ Using existing Internet Gateway: $IGW_ID"
fi

# -------------------------------------------------------
# Step 5: Creating Kubernetes Cluster
# -------------------------------------------------------
echo "📋 Step 5: Creating Kubernetes Cluster..."
CLUSTER_ID=$(oci ce cluster list \
    --compartment-id "$COMPARTMENT_ID" \
    --query "data[?name=='$CLUSTER_NAME'].id | [0]" \
    --raw-output)

if [ -z "$CLUSTER_ID" ] || [ "$CLUSTER_ID" = "null" ]; then
    CLUSTER_ID=$(oci ce cluster create \
        --compartment-id "$COMPARTMENT_ID" \
        --name "$CLUSTER_NAME" \
        --kubernetes-version "$KUBERNETES_VERSION" \
        --type BASIC_CLUSTER \
        --vcn-id "$VCN_ID" \
        --endpoint-subnet-id "$SUBNET_ID" \
        --service-lb-subnet-ids '["'"$SUBNET_ID"'"]' \
        --pods-cidr "$POD_CIDR" \
        --services-cidr "$SERVICE_CIDR" \
        --is-public-ip-enabled true \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Kubernetes Cluster: $CLUSTER_ID")
    echo "✅ Submitted cluster creation: $CLUSTER_ID"

    echo "⏳ Waiting for cluster to reach ACTIVE state (this takes 5-10 minutes)..."
    for i in {1..40}; do
        CLUSTER_STATE=$(oci ce cluster get \
            --cluster-id "$CLUSTER_ID" \
            --query "data.\"lifecycle-state\"" --raw-output)

        if [ "$CLUSTER_STATE" = "ACTIVE" ]; then
            echo "✅ Cluster is ACTIVE"
            break
        elif [ "$CLUSTER_STATE" = "FAILED" ]; then
            echo "❌ ERROR: Cluster entered FAILED state. Check OCI console for details."
            exit 1
        fi

        echo "⏳ Cluster state: $CLUSTER_STATE ($i/40) — waiting 30s..."
        sleep 30
    done

    if [ "$CLUSTER_STATE" != "ACTIVE" ]; then
        echo "❌ ERROR: Cluster did not become ACTIVE within 20 minutes."
        exit 1
    fi
else
    echo "✅ Using existing Cluster: $CLUSTER_ID"
fi

# -------------------------------------------------------
# Step 6: Creating Node Pool
# -------------------------------------------------------
echo "📋 Step 6: Creating Node Pool..."
NODE_POOL_ID=$(oci ce node-pool list \
    --cluster-id "$CLUSTER_ID" \
    --compartment-id "$COMPARTMENT_ID" \
    --query "data[?name=='$NODE_POOL_NAME'].id | [0]" \
    --raw-output)

if [ -z "$NODE_POOL_ID" ] || [ "$NODE_POOL_ID" = "null" ]; then
    NODE_POOL_ID=$(oci ce node-pool create \
        --cluster-id "$CLUSTER_ID" \
        --compartment-id "$COMPARTMENT_ID" \
        --name "$NODE_POOL_NAME" \
        --kubernetes-version "$KUBERNETES_VERSION" \
        --node-shape VM.Standard.A1.Flex \
        --size 2 \
        --node-shape-config '{"memoryInGBs": 6, "ocpus": 2}' \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Node Pool: $NODE_POOL_ID")
    echo "✅ Created Node Pool: $NODE_POOL_ID"
else
    echo "✅ Using existing Node Pool: $NODE_POOL_ID"
fi

# -------------------------------------------------------
# Step 7: Configuring Kubernetes Access
# -------------------------------------------------------
echo "📋 Step 7: Configuring Kubernetes Access..."
oci ce cluster create-kubeconfig \
    --cluster-id "$CLUSTER_ID" \
    --file "$HOME/.kube/payment-watchdog-config" \
    --region "$REGION" \
    --token-version 2.0.0

export KUBECONFIG="$HOME/.kube/payment-watchdog-config"
echo "✅ Kubeconfig created: $KUBECONFIG"

# -------------------------------------------------------
# Step 8: Installing NGINX Ingress Controller
# -------------------------------------------------------
echo "📋 Step 8: Installing NGINX Ingress Controller..."
kubectl get namespace ingress-nginx 2>/dev/null || kubectl create namespace ingress-nginx

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx 2>/dev/null || true
#helm repo update ingress-nginx
helm repo update

helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
    --namespace ingress-nginx \
    --create-namespace \
    --set controller.service.type=LoadBalancer \
    --set controller.publishService.enabled=true

echo "⏳ Waiting for external IP (up to 5 minutes)..."
EXTERNAL_IP=""
for i in {1..30}; do
    EXTERNAL_IP=$(kubectl get svc ingress-nginx-controller \
        -n ingress-nginx \
        -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")

    if [ -n "$EXTERNAL_IP" ]; then
        echo "✅ External IP assigned: $EXTERNAL_IP"
        break
    fi

    echo "⏳ Waiting for external IP... ($i/30)"
    sleep 10
done

if [ -z "$EXTERNAL_IP" ]; then
    echo "❌ Failed to get external IP within timeout. Check manually:"
    kubectl get svc ingress-nginx-controller -n ingress-nginx
    exit 1
fi

# -------------------------------------------------------
# Step 9: Updating Domain Configuration
# -------------------------------------------------------
echo "📋 Step 9: Updating Configuration..."
if [ -n "$DOMAIN" ]; then
    echo "🌐 Updating configuration for domain: $DOMAIN"

    INGRESS_FILE="api/deployments/kubernetes/base/components/ingress/ingress.yaml"
    CI_FILE=".github/workflows/payment-watchdog-ci.yml"

    if [ -f "$INGRESS_FILE" ]; then
        sed -i "s/payment-watchdog\.local/$DOMAIN/g" "$INGRESS_FILE"
        echo "✅ Updated ingress host: $INGRESS_FILE"
    else
        echo "⚠️  Ingress file not found, skipping: $INGRESS_FILE"
    fi

    if [ -f "$CI_FILE" ]; then
        sed -i "s/staging\.payment-watchdog\.example\.com/https:\/\/$DOMAIN/g" "$CI_FILE"
        echo "✅ Updated CI environment URL: $CI_FILE"
    else
        echo "⚠️  CI workflow file not found, skipping: $CI_FILE"
    fi
else
    echo "🌐 No domain provided - using IP-based access"
    echo "   Services accessible via: http://$EXTERNAL_IP"
fi

# -------------------------------------------------------
# Step 10: Creating Staging Namespace
# -------------------------------------------------------
echo "📋 Step 10: Creating Staging Namespace..."
kubectl create namespace staging --dry-run=client -o yaml | kubectl apply -f -

# -------------------------------------------------------
# Done
# -------------------------------------------------------
echo ""
echo "🎉 Oracle Cloud Setup Complete!"
echo "=================================="
echo "Region:      $REGION"
echo "Compartment: $COMPARTMENT_ID"
echo "Cluster ID:  $CLUSTER_ID"
echo "External IP: $EXTERNAL_IP"

if [ -n "$DOMAIN" ]; then
    echo "Domain:      $DOMAIN"
    echo ""
    echo "Next Steps:"
    echo "1. Point your domain DNS A record to: $EXTERNAL_IP"
    echo "2. Run: kubectl apply -f api/deployments/kubernetes"
    echo "3. Update GitHub KUBE_CONFIG secret with:"
    echo "   cat $KUBECONFIG | base64 -w 0"
else
    echo ""
    echo "Next Steps:"
    echo "1. Run: kubectl apply -f api/deployments/kubernetes"
    echo "2. Access services via: http://$EXTERNAL_IP"
    echo "3. Add GitHub KUBE_CONFIG secret:"
    echo "   cat $KUBECONFIG | base64 -w 0"
fi

echo "=================================="

# Show created resources for tracking
if [ ${#CREATED_RESOURCES[@]} -gt 0 ]; then
    echo ""
    echo "📋 Created Resources Summary:"
    for resource in "${CREATED_RESOURCES[@]}"; do
        echo "   - $resource"
    done
    echo ""
    echo "💡 To cleanup these resources, run:"
    echo "   ./deployment/cleanup-oracle.sh"
fi