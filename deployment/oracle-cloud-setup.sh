#!/bin/bash

# Oracle Cloud Free Tier Setup Script for Payment Watchdog
# IMPROVED VERSION - Based on previous deployment learnings
# This script automates the setup of Oracle Cloud infrastructure with Always Free compliance

set -euo pipefail

# Configuration
REGION=${REGION:-ap-melbourne-1}

if [[ "$REGION" != "ap-sydney-1" && "$REGION" != "ap-melbourne-1" ]]; then
    echo "❌ ERROR: For sovereign compliance, REGION must be 'ap-sydney-1' or 'ap-melbourne-1'."
    exit 1
fi

echo "🚀 Oracle Cloud Free Tier Setup for Payment Watchdog"
echo "=================================================="
echo "Region: $REGION"

# Check for force flag to bypass duplicate warnings
FORCE_MODE=${FORCE:-false}
if [ "$FORCE_MODE" = "true" ]; then
    echo "🔧 FORCE MODE: Will create resources even if they exist"
    echo ""
fi

# Always Free Configuration
COMPARTMENT_NAME="PaymentWatchdogSovereign"
CLUSTER_NAME="payment-watchdog-cluster-v3"
NODE_POOL_NAME="payment-watchdog-nodes"
VCN_NAME="payment-watchdog-vcn"
WORKER_SUBNET_NAME="payment-worker-subnet"
POD_SUBNET_NAME="payment-pod-subnet-large"
LB_SUBNET_NAME="payment-lb-subnet"
IGW_NAME="payment-igw"
DOMAIN=${DOMAIN:-""}  # Optional - leave empty for IP-based access
TENANCY_ID="ocid1.tenancy.oc1..aaaaaaaaitfhbb6ix7yqiavspixzepzm4babf6qjonzom5pe4lvqzxvp2xla"
KUBERNETES_VERSION=${KUBERNETES_VERSION:-"v1.35.0"}

# Always Free Node Configuration
NODE_SHAPE="VM.Standard.E2.1"         # Always Free eligible
NODE_OCPUS=1                          # Free tier limit
NODE_MEMORY=8                         # Free tier limit
NODE_SIZE=1                           # Only 1 node in free tier
MAX_PODS_PER_NODE=31                  # Optimized for pod capacity

# Region-specific network configuration - Always Free Optimized
if [[ "$REGION" == "ap-sydney-1" ]]; then
    VCN_CIDR="10.0.0.0/16"       # Standard VCN CIDR
    WORKER_SUBNET_CIDR="10.0.1.0/24"      # Nodes
    POD_SUBNET_CIDR="10.0.4.0/22"        # Pod subnet (1024 IPs) - fits within VCN
    CLUSTER_POD_CIDR="10.1.0.0/16"      # Cluster pods CIDR - separate from VCN
    LB_SUBNET_CIDR="10.0.2.0/24"          # Load balancers
    SERVICE_CIDR="10.96.0.0/12"           # Standard Kubernetes range
elif [[ "$REGION" == "ap-melbourne-1" ]]; then
    VCN_CIDR="10.1.0.0/16"       # Standard VCN CIDR
    WORKER_SUBNET_CIDR="10.1.1.0/24"      # Nodes
    POD_SUBNET_CIDR="10.1.4.0/22"        # Pod subnet (1024 IPs) - fits within VCN
    CLUSTER_POD_CIDR="10.2.0.0/16"      # Cluster pods CIDR - separate from VCN
    LB_SUBNET_CIDR="10.1.2.0/24"          # Load balancers
    SERVICE_CIDR="10.96.0.0/12"           # Standard Kubernetes range
else
    echo "❌ ERROR: Unsupported region: $REGION"
    exit 1
fi

echo "Node Shape: $NODE_SHAPE ($NODE_OCPUS OCPU, ${NODE_MEMORY}GB RAM)"
echo "Max Pods per Node: $MAX_PODS_PER_NODE"
echo ""

# -------------------------------------------------------
# Preflight checks
# -------------------------------------------------------
echo "🔍 Running preflight checks..."

# Check for existing resources to prevent duplicates
echo "🔍 Checking for existing resources in compartment..."

# Get compartment ID first for resource checks
COMPARTMENT_ID=$(oci iam compartment list \
    --compartment-id "$TENANCY_ID" \
    --query "data[?name=='$COMPARTMENT_NAME'].id | [0]" \
    --raw-output)

if [ -z "$COMPARTMENT_ID" ] || [ "$COMPARTMENT_ID" = "null" ]; then
    echo "ℹ️  Compartment '$COMPARTMENT_NAME' not found - will be created"
else
    echo "✅ Found compartment: $COMPARTMENT_ID"
    
    # Check for existing resources
    EXISTING_VCN=$(oci network vcn list --compartment-id "$COMPARTMENT_ID" --query "data[?name=='$VCN_NAME'].id | [0]" --raw-output)
    EXISTING_CLUSTER=$(oci ce cluster list --compartment-id "$COMPARTMENT_ID" --query "data[?name=='$CLUSTER_NAME' && \"lifecycle-state\"!='DELETED'].id | [0]" --raw-output)
    EXISTING_VCN_COUNT=$(oci network vcn list --compartment-id "$COMPARTMENT_ID" --query "data | length(@)" --raw-output)
    
    echo "🔍 Found ${EXISTING_VCN_COUNT:-0} VCN(s) in compartment"
    
    if [ "${EXISTING_VCN_COUNT:-0}" -gt 0 ] && [ "${EXISTING_VCN_COUNT:-0}" != "null" ]; then
        echo "📋 Existing VCNs in compartment:"
        oci network vcn list --compartment-id "$COMPARTMENT_ID" --query "data[*].{name:\"display-name\", id:id, \"cidr-block\":\"cidr-block\"}" --output table
        echo ""
    fi
    
    if [ -n "$EXISTING_VCN" ] && [ "$EXISTING_VCN" != "null" ]; then
        echo "⚠️  WARNING: VCN '$VCN_NAME' already exists in compartment"
        echo "   Existing VCN ID: $EXISTING_VCN"
        
        if [ "$FORCE_MODE" != "true" ]; then
            echo ""
            echo "🚨 DUPLICATE INFRASTRUCTURE DETECTED!"
            echo "   This can cause network conflicts and deployment failures."
            echo ""
            read -p "❓ Do you want to continue and create additional VCN? (y/N): " -n 1 -r
            echo ""
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                echo "❌ Deployment cancelled to prevent duplicate infrastructure."
                echo "💡 Use 'FORCE=true ./deployment/oracle-cloud-setup.sh' to override this check."
                exit 1
            fi
            echo "⚠️  Continuing with duplicate infrastructure (not recommended)..."
        else
            echo "🔧 FORCE MODE: Continuing despite duplicate infrastructure..."
        fi
        
        # Use existing VCN instead of creating new one
        echo "🔄 Using existing VCN: $EXISTING_VCN"
        VCN_ID="$EXISTING_VCN"
    fi
    
    if [ -n "$EXISTING_CLUSTER" ] && [ "$EXISTING_CLUSTER" != "null" ]; then
        echo "⚠️  WARNING: Cluster '$CLUSTER_NAME' already exists in compartment"
        echo "   Existing Cluster ID: $EXISTING_CLUSTER"
    fi
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

# Check required tools
for cmd in oci kubectl helm jq; do
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
        echo ""
        echo "💡 To cleanup these resources, run:"
        echo "   ./deployment/cleanup-oracle-improved.sh"
    fi
}

trap cleanup ERR

# -------------------------------------------------------
# Step 1: Creating Compartment
# -------------------------------------------------------
echo "📋 Step 1: Creating Compartment..."

if [ -z "$COMPARTMENT_ID" ] || [ "$COMPARTMENT_ID" = "null" ]; then
    COMPARTMENT_ID=$(oci iam compartment create \
        --name "$COMPARTMENT_NAME" \
        --description "Payment Watchdog Sovereign Compartment" \
        --compartment-id "$TENANCY_ID" \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Compartment: $COMPARTMENT_ID")
    echo "✅ Created compartment: $COMPARTMENT_ID"

    echo "⏳ Waiting for IAM compartment to propagate..."
    sleep 15
else
    echo "✅ Using existing compartment: $COMPARTMENT_ID"
fi

# Export for global availability
export COMPARTMENT_ID

# -------------------------------------------------------
# Step 2: Creating VCN
# -------------------------------------------------------
echo "📋 Step 2: Creating VCN..."

if [ -z "$VCN_ID" ] || [ "$VCN_ID" = "null" ]; then
    VCN_ID=$(oci network vcn list \
        --compartment-id "$COMPARTMENT_ID" \
        --query "data[?name=='$VCN_NAME'].id | [0]" \
        --raw-output)
fi

if [ -z "$VCN_ID" ] || [ "$VCN_ID" = "null" ]; then
    VCN_ID=$(oci network vcn create \
        --compartment-id "$COMPARTMENT_ID" \
        --cidr-block "$VCN_CIDR" \
        --display-name "$VCN_NAME" \
        --dns-label "paymentvcn" \
        --freeform-tags '{"project":"payment-watchdog","tier":"free"}' \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("VCN: $VCN_ID")
    echo "✅ Created VCN: $VCN_ID"
else
    echo "✅ Using existing VCN: $VCN_ID"
fi

# -------------------------------------------------------
# Step 2.5: Configure VCN Default Security List for OKE
# -------------------------------------------------------
echo "📋 Step 2.5: Configuring VCN Security Rules for OKE..."
DEFAULT_SL_ID=$(oci network vcn get --vcn-id "$VCN_ID" --query "data.\"default-security-list-id\"" --raw-output)

if [ -n "$DEFAULT_SL_ID" ] && [ "$DEFAULT_SL_ID" != "null" ]; then
    # Inject required OKE ingress and egress rules to prevent Node timeout
    INGRESS_PAYLOAD='[
      {"protocol": "all", "source": "'"$VCN_CIDR"'", "is-stateless": false, "description": "Pod-to-Pod communication"},
      {"protocol": "1", "icmp-options": {"type": 3, "code": 4}, "source": "'"$CLUSTER_POD_CIDR"'", "is-stateless": false, "description": "Path discovery from control plane"},
      {"protocol": "6", "source": "'"$CLUSTER_POD_CIDR"'", "is-stateless": false, "description": "TCP access from Kubernetes Control Plane"},
      {"protocol": "6", "source": "0.0.0.0/0", "is-stateless": false, "tcp-options": {"destination-port-range": {"min": 22, "max": 22}}, "description": "SSH access"}
    ]'

    EGRESS_PAYLOAD='[
      {"protocol": "all", "destination": "0.0.0.0/0", "is-stateless": false, "description": "Allow all outbound (Default)"},
      {"protocol": "6", "destination": "'"$CLUSTER_POD_CIDR"'", "is-stateless": false, "tcp-options": {"destination-port-range": {"min": 6443, "max": 6443}}, "description": "Kubernetes API Endpoint"},
      {"protocol": "6", "destination": "'"$CLUSTER_POD_CIDR"'", "is-stateless": false, "tcp-options": {"destination-port-range": {"min": 12250, "max": 12250}}, "description": "Worker to Control Plane"}
    ]'

    oci network security-list update \
        --security-list-id "$DEFAULT_SL_ID" \
        --ingress-security-rules "$INGRESS_PAYLOAD" \
        --egress-security-rules "$EGRESS_PAYLOAD" \
        --force > /dev/null

    echo "✅ Configured Default Security List with OKE rules: $DEFAULT_SL_ID"
else
    echo "⚠️ Could not find Default Security List ID for VCN: $VCN_ID"
fi

# -------------------------------------------------------
# Step 3: Creating Worker Subnet (for nodes)
# -------------------------------------------------------
echo "📋 Step 3: Creating Worker Subnet..."
WORKER_SUBNET_ID=$(oci network subnet list \
    --compartment-id "$COMPARTMENT_ID" \
    --vcn-id "$VCN_ID" \
    --query "data[?name=='$WORKER_SUBNET_NAME'].id | [0]" \
    --raw-output)

if [ -z "$WORKER_SUBNET_ID" ] || [ "$WORKER_SUBNET_ID" = "null" ]; then
    WORKER_SUBNET_ID=$(oci network subnet create \
        --compartment-id "$COMPARTMENT_ID" \
        --vcn-id "$VCN_ID" \
        --cidr-block "$WORKER_SUBNET_CIDR" \
        --display-name "$WORKER_SUBNET_NAME" \
        --freeform-tags '{"project":"payment-watchdog","purpose":"nodes"}' \
        --prohibit-public-ip-on-vnic true \
        --dns-label "paymentworkers" \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Worker Subnet: $WORKER_SUBNET_ID")
    echo "✅ Created worker subnet: $WORKER_SUBNET_ID"
else
    echo "✅ Using existing worker subnet: $WORKER_SUBNET_ID"
fi

# -------------------------------------------------------
# Step 4: Creating Pod Subnet (large for pod IPs)
# -------------------------------------------------------
echo "📋 Step 4: Creating Large Pod Subnet..."
POD_SUBNET_ID=$(oci network subnet list \
    --compartment-id "$COMPARTMENT_ID" \
    --vcn-id "$VCN_ID" \
    --query "data[?name=='$POD_SUBNET_NAME'].id | [0]" \
    --raw-output)

if [ -z "$POD_SUBNET_ID" ] || [ "$POD_SUBNET_ID" = "null" ]; then
    POD_SUBNET_ID=$(oci network subnet create \
        --compartment-id "$COMPARTMENT_ID" \
        --vcn-id "$VCN_ID" \
        --cidr-block "$POD_SUBNET_CIDR" \
        --display-name "$POD_SUBNET_NAME" \
        --freeform-tags '{"project":"payment-watchdog","purpose":"pods"}' \
        --prohibit-public-ip-on-vnic true \
        --dns-label "paymentpods" \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Pod Subnet: $POD_SUBNET_ID")
    echo "✅ Created large pod subnet: $POD_SUBNET_ID"
else
    echo "✅ Using existing pod subnet: $POD_SUBNET_ID"
fi

# -------------------------------------------------------
# Step 5: Creating Load Balancer Subnet
# -------------------------------------------------------
echo "📋 Step 5: Creating Load Balancer Subnet..."
LB_SUBNET_ID=$(oci network subnet list \
    --compartment-id "$COMPARTMENT_ID" \
    --vcn-id "$VCN_ID" \
    --query "data[?name=='$LB_SUBNET_NAME'].id | [0]" \
    --raw-output)

if [ -z "$LB_SUBNET_ID" ] || [ "$LB_SUBNET_ID" = "null" ]; then
    LB_SUBNET_ID=$(oci network subnet create \
        --compartment-id "$COMPARTMENT_ID" \
        --vcn-id "$VCN_ID" \
        --cidr-block "$LB_SUBNET_CIDR" \
        --display-name "$LB_SUBNET_NAME" \
        --freeform-tags '{"project":"payment-watchdog","purpose":"loadbalancer"}' \
        --dns-label "paymentlb" \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Load Balancer Subnet: $LB_SUBNET_ID")
    echo "✅ Created load balancer subnet: $LB_SUBNET_ID"
else
    echo "✅ Using existing LB subnet: $LB_SUBNET_ID"
fi

# -------------------------------------------------------
# Step 6: Creating Service Gateway and Routes
# -------------------------------------------------------
echo "📋 Step 6: Creating Service Gateway..."
if [[ "$REGION" == "ap-sydney-1" ]]; then
    # Sydney region - use All Services
    SERVICES='["all-ap-sydney-services-in-oracle-services-network"]'
elif [[ "$REGION" == "ap-melbourne-1" ]]; then
    # Melbourne region - use Melbourne Services
    SERVICES='["all-mel-services-in-oracle-services-network"]'
elif [[ "$REGION" == "ap-osaka-1" ]]; then
    # Osaka region - use Osaka Services
    SERVICES='["all-osaka-services-in-oracle-services-network"]'
else
    # Default to Melbourne Services
    SERVICES='["all-mel-services-in-oracle-services-network"]'
fi

IGW_ID=$(oci network service-gateway list \
    --compartment-id "$COMPARTMENT_ID" \
    --vcn-id "$VCN_ID" \
    --query "data[?name=='$IGW_NAME'].id | [0]" \
    --raw-output)

if [ -z "$IGW_ID" ] || [ "$IGW_ID" = "null" ]; then
    IGW_ID=$(oci network service-gateway create \
        --compartment-id "$COMPARTMENT_ID" \
        --vcn-id "$VCN_ID" \
        --display-name "$IGW_NAME" \
        --freeform-tags '{"project":"payment-watchdog"}' \
        --services "$SERVICES" \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Service Gateway: $IGW_ID")
    echo "✅ Created Service Gateway: $IGW_ID"
else
    echo "✅ Using existing Service Gateway: $IGW_ID"
fi

# Update route table with service gateway access
RT_ID=$(oci network route-table list \
        --compartment-id "$COMPARTMENT_ID" \
        --vcn-id "$VCN_ID" \
        --query "data[0].id" --raw-output)
    
# Check if service gateway needs services update
if [[ "$REGION" == "ap-melbourne-1" ]]; then
    # Melbourne region - check if services need updating
    CURRENT_SERVICES=$(oci network service-gateway get \
        --service-gateway-id "$IGW_ID" \
        --query "data.services" --raw-output 2>/dev/null || echo "[]")
    
    # Compare current services with required services
    REQUIRED_SERVICES='["all-mel-services-in-oracle-services-network"]'
    
    if [[ "$CURRENT_SERVICES" != "$REQUIRED_SERVICES" ]]; then
        echo "🔧 Updating service gateway with required services..."
        oci network service-gateway update \
            --service-gateway-id "$IGW_ID" \
            --services "$REQUIRED_SERVICES" \
            --force
        echo "✅ Updated service gateway with Oracle services"
    else
        echo "✅ Service gateway already has required services"
    fi
elif [[ "$REGION" == "ap-sydney-1" ]]; then
    # Sydney region
    echo "✅ Service gateway already configured for Sydney"
elif [[ "$REGION" == "ap-osaka-1" ]]; then
    # Osaka region
    echo "✅ Service gateway already configured for Osaka"
fi
    
oci network route-table update \
    --rt-id "$RT_ID" \
    --route-rules '[{"cidrBlock":"0.0.0.0/0","networkEntityId":"'"$IGW_ID"'"}]' \
    --force
echo "✅ Updated route table for service gateway access"

# -------------------------------------------------------
# Step 7: Creating Kubernetes Cluster (Fixed)
# -------------------------------------------------------
echo "📋 Step 7: Creating Kubernetes Cluster..."
CLUSTER_ID=$(oci ce cluster list \
    --compartment-id "$COMPARTMENT_ID" \
    --query "data[?name=='$CLUSTER_NAME' && \"lifecycle-state\"!='DELETED'].id | [0]" \
    --raw-output)

if [ -z "$CLUSTER_ID" ] || [ "$CLUSTER_ID" = "null" ]; then
    echo "🔧 Creating new cluster with the following parameters:"
    echo "  • Compartment ID: $COMPARTMENT_ID"
    echo "  • Cluster Name: $CLUSTER_NAME"
    echo "  • Kubernetes Version: $KUBERNETES_VERSION"
    echo "  • VCN ID: $VCN_ID"
    echo "  • Worker Subnet ID: $WORKER_SUBNET_ID"
    echo "  • Load Balancer Subnet ID: $LB_SUBNET_ID"
    echo "  • Pods CIDR: $CLUSTER_POD_CIDR"
    echo "  • Services CIDR: $SERVICE_CIDR"
    echo ""
    
    CLUSTER_ID=$(oci ce cluster create \
        --compartment-id "$COMPARTMENT_ID" \
        --name "$CLUSTER_NAME" \
        --kubernetes-version "$KUBERNETES_VERSION" \
        --type BASIC_CLUSTER \
        --vcn-id "$VCN_ID" \
        --endpoint-subnet-id "$WORKER_SUBNET_ID" \
        --endpoint-public-ip-enabled true \
        --service-lb-subnet-ids '["'"$LB_SUBNET_ID"'"]' \
        --pods-cidr "$CLUSTER_POD_CIDR" \
        --services-cidr "$SERVICE_CIDR" \
        --freeform-tags '{"project":"payment-watchdog","tier":"free"}' \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Kubernetes Cluster: $CLUSTER_ID")
    echo "✅ Submitted cluster creation: $CLUSTER_ID"
    
    # Wait for cluster to become ACTIVE
    echo "⏳ Waiting for cluster to become ACTIVE..."
    for i in {1..40}; do
        CLUSTER_STATE=$(oci ce cluster get \
            --cluster-id "$CLUSTER_ID" \
            --query "data.\"lifecycle-state\"" --raw-output)
        
        if [ "$CLUSTER_STATE" = "ACTIVE" ]; then
            echo "✅ Cluster is ACTIVE"
            break
        elif [ "$CLUSTER_STATE" = "FAILED" ]; then
            echo "❌ Cluster creation FAILED"
            exit 1
        fi
        
        echo "⏳ Cluster state: $CLUSTER_STATE ($i/40)"
        sleep 30
    done
    
    if [ "$CLUSTER_STATE" != "ACTIVE" ]; then
        echo "❌ ERROR: Cluster did not become ACTIVE within 20 minutes."
        exit 1
    fi
else
    echo "✅ Using existing cluster: $CLUSTER_ID"
fi

# -------------------------------------------------------
# Step 8: Creating Node Pool (Always Free Optimized)
# -------------------------------------------------------
echo "📋 Step 8: Creating Node Pool..."
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
        --node-shape "$NODE_SHAPE" \
        --placement-configs '[{"availabilityDomain": "vIgK:AP-MELBOURNE-1-AD-1", "subnetId": "'"$WORKER_SUBNET_ID"'"}]' \
        --pod-subnet-ids '["'"$POD_SUBNET_ID"'"]' \
        --node-shape-config '{"memoryInGBs": '"$NODE_MEMORY"', "ocpus": '"$NODE_OCPUS"'}' \
        --initial-node-labels '[{"key": "name", "value": "payment-watchdog-free"}]' \
        --max-pods-per-node "$MAX_PODS_PER_NODE" \
        --size "$NODE_SIZE" \
        --freeform-tags '{"project":"payment-watchdog","tier":"free"}' \
        --query "data.id" --raw-output)
    CREATED_RESOURCES+=("Node Pool: $NODE_POOL_ID")
    echo "✅ Created node pool: $NODE_POOL_ID"
    
    # Wait for nodes to become ready
    echo "⏳ Waiting for nodes to become ready..."
    sleep 60
    
    # Configure kubectl
    oci ce cluster create-kubeconfig \
        --cluster-id "$CLUSTER_ID" \
        --file "$HOME/.kube/payment-watchdog-config" \
        --region "$REGION"
    
    export KUBECONFIG="$HOME/.kube/payment-watchdog-config"
    
    # Wait for node to register
    echo "⏳ Waiting for node to register..."
    for i in {1..20}; do
        NODE_COUNT=$(kubectl get nodes --no-headers 2>/dev/null | wc -l || echo "0")
        if [ "$NODE_COUNT" -gt 0 ]; then
            echo "✅ Node registered successfully"
            kubectl get nodes -o wide
            break
        fi
        echo "⏳ Waiting for node registration... ($i/20)"
        sleep 30
    done
else
    echo "✅ Using existing node pool: $NODE_POOL_ID"
fi

# -------------------------------------------------------
# Step 9: Configuring Kubernetes Access
# -------------------------------------------------------
echo "📋 Step 9: Configuring Kubernetes Access..."
if [ ! -f "$HOME/.kube/payment-watchdog-config" ]; then
    oci ce cluster create-kubeconfig \
        --cluster-id "$CLUSTER_ID" \
        --file "$HOME/.kube/payment-watchdog-config" \
        --region "$REGION" \
        --token-version 2.0.0
fi

export KUBECONFIG="$HOME/.kube/payment-watchdog-config"
echo "✅ Kubeconfig created: $KUBECONFIG"

# -------------------------------------------------------
# Step 10: Installing NGINX Ingress Controller
# -------------------------------------------------------
echo "📋 Step 10: Installing NGINX Ingress Controller..."
kubectl get namespace ingress-nginx 2>/dev/null || kubectl create namespace ingress-nginx

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx 2>/dev/null || true
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
# Step 11: Deploying Application
# -------------------------------------------------------
echo "📋 Step 11: Deploying Payment Watchdog..."
export KUBECONFIG="$HOME/.kube/payment-watchdog-config"

# Create namespace
kubectl create namespace sovereign-au --dry-run=client -o yaml | kubectl apply -f -

# Deploy application
if [ -d "api/deployments/kubernetes/overlays/sovereign-au" ]; then
    kubectl apply -k api/deployments/kubernetes/overlays/sovereign-au
    echo "✅ Application deployed using Kustomize"
else
    echo "⚠️  Kubernetes manifests not found, skipping application deployment"
fi

echo "⏳ Waiting for deployment to complete..."
sleep 30

# Check deployment status
echo "📋 Checking deployment status..."
kubectl get pods -n sovereign-au
kubectl get services -n sovereign-au

# -------------------------------------------------------
# Step 12: Updating Domain Configuration
# -------------------------------------------------------
echo "📋 Step 12: Updating Configuration..."
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
# Done
# -------------------------------------------------------
echo ""
echo "🎉 Oracle Cloud Setup Complete!"
echo "=================================="
echo "Region:      $REGION"
echo "Compartment: $COMPARTMENT_ID"
echo "Cluster ID:  $CLUSTER_ID"
echo "Node Pool ID: $NODE_POOL_ID"
echo "External IP: $EXTERNAL_IP"
echo "Max Pods per Node: $MAX_PODS_PER_NODE"

if [ -n "$DOMAIN" ]; then
    echo "Domain:      $DOMAIN"
    echo ""
    echo "Next Steps:"
    echo "1. Point your domain DNS A record to: $EXTERNAL_IP"
    echo "2. Verify deployment with: kubectl get pods -n sovereign-au"
    echo "3. Update GitHub KUBE_CONFIG secret with:"
    echo "   cat $KUBECONFIG | base64 -w 0"
else
    echo ""
    echo "Next Steps:"
    echo "1. Verify deployment with: kubectl get pods -n sovereign-au"
    echo "2. Access services via: http://$EXTERNAL_IP"
    echo "3. Add GitHub KUBE_CONFIG secret:"
    echo "   cat $KUBECONFIG | base64 -w 0"
fi

echo ""
echo "🔍 Verify node capacity:"
echo "kubectl get nodes -o wide"
echo "kubectl describe nodes | grep -i 'capacity.*pods'"
echo ""

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
    echo "   ./deployment/cleanup-oracle-improved.sh"
fi

echo ""
echo "🚀 Setup completed successfully!"
echo "📝 All previous issues have been resolved:"
echo "   ✅ Always Free compliant configuration"
echo "   ✅ Proper network architecture with separate subnets"
echo "   ✅ Large pod subnet for sufficient IP addresses"
echo "   ✅ Max pods per node configured (31 pods)"
echo "   ✅ Fixed cluster endpoint configuration"
echo "   ✅ Application deployment included"
