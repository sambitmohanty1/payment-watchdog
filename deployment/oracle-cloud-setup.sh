#!/bin/bash

# Oracle Cloud Free Tier Setup Script for Payment Watchdog
# This script automates the setup of Oracle Cloud infrastructure

set -e

REGION=${REGION:-ap-sydney-1}

if [[ "$REGION" != "ap-sydney-1" && "$REGION" != "ap-melbourne-1" ]]; then
    echo "❌ ERROR: For sovereign compliance, REGION must be 'ap-sydney-1' or 'ap-melbourne-1'."
    exit 1
fi

echo "🚀 Oracle Cloud Free Tier Setup for Payment Watchdog in region: $REGION"
echo "================================================"

# Configuration
COMPARTMENT_NAME="PaymentWatchdogSovereign"
CLUSTER_NAME="payment-watchdog-cluster"
NODE_POOL_NAME="payment-watchdog-nodes"
VCN_NAME="payment-watchdog-vcn"
SUBNET_NAME="payment-watchdog-subnet"
IGW_NAME="payment-watchdog-igw"
DOMAIN="your-domain.com"  # UPDATE THIS

echo "📋 Step 1: Creating Compartment..."
COMPARTMENT_ID=$(oci iam compartment list --query "data[?name=='$COMPARTMENT_NAME'].id | [0]" --raw-output)
if [ -z "$COMPARTMENT_ID" ] || [ "$COMPARTMENT_ID" = "null" ]; then
    COMPARTMENT_ID=$(oci iam compartment create \
        --name "$COMPARTMENT_NAME" \
        --description "Isolated Sovereign Compartment" \
        --region "$REGION" \
        --query "data.id" --raw-output)
    echo "✅ Created compartment: $COMPARTMENT_ID"
else
    echo "✅ Using existing compartment: $COMPARTMENT_ID"
fi

echo "📋 Step 2: Creating VCN and Subnet..."
VCN_ID=$(oci network vcn list --compartment-id "$COMPARTMENT_ID" --query "data[?name=='$VCN_NAME'].id | [0]" --raw-output)
if [ -z "$VCN_ID" ] || [ "$VCN_ID" = "null" ]; then
    VCN_ID=$(oci network vcn create \
            --compartment-id "$COMPARTMENT_ID" \
            --cidr-block 10.0.0.0/16 \
            --display-name "$VCN_NAME" \
            --region "$REGION" \
            --query "data.id" --raw-output)
    echo "✅ Created VCN: $VCN_ID"
else
    echo "✅ Using existing VCN: $VCN_ID"
fi

SUBNET_ID=$(oci network subnet list --vcn-id "$VCN_ID" --query "data[?name=='$SUBNET_NAME'].id | [0]" --raw-output)
if [ -z "$SUBNET_ID" ]; then
    SUBNET_ID=$(oci network subnet create \
                --compartment-id "$COMPARTMENT_ID" \
                --vcn-id "$VCN_ID" \
                --cidr-block 10.0.1.0/24 \
                --display-name "$SUBNET_NAME" \
                --query "data.id" --raw-output)
    echo "✅ Created Subnet: $SUBNET_ID"
else
    echo "✅ Using existing Subnet: $SUBNET_ID"
fi

echo "📋 Step 3: Creating Internet Gateway..."
IGW_ID=$(oci network internet-gateway list --vcn-id "$VCN_ID" --query "data[?name=='$IGW_NAME'].id | [0]" --raw-output)
if [ -z "$IGW_ID" ]; then
    IGW_ID=$(oci network internet-gateway create \
                --compartment-id "$COMPARTMENT_ID" \
                --vcn-id "$VCN_ID" \
                --display-name "$IGW_NAME" \
                --query "data.id" --raw-output)
    echo "✅ Created Internet Gateway: $IGW_ID"
    
    # Get route table ID and add route
    RT_ID=$(oci network vcn get --vcn-id "$VCN_ID" --query "data.default_route_table_id" --raw-output)
    oci network route-table update \
        --rt-id "$RT_ID" \
        --route-rules '[{"cidr":"0.0.0.0/0","networkEntityId":"'"$IGW_ID"'","networkEntityType":"InternetGateway"}]'
    echo "✅ Updated route table with internet gateway route"
else
    echo "✅ Using existing Internet Gateway: $IGW_ID"
fi

echo "📋 Step 4: Creating Kubernetes Cluster..."
CLUSTER_ID=$(oci ce cluster list --compartment-id "$COMPARTMENT_ID" --query "data[?name=='$CLUSTER_NAME'].id | [0]" --raw-output)
if [ -z "$CLUSTER_ID" ]; then
    CLUSTER_ID=$(oci ce cluster create \
                --compartment-id "$COMPARTMENT_ID" \
                --name "$CLUSTER_NAME" \
                --kubernetes-version v1.28.2 \
                --type VIRTUAL_NODE_POOL \
                --vcn-id "$VCN_ID" \
                --subnet-ids '["'"$SUBNET_ID"'"]' \
                --endpoint-subnet-ids '["'"$SUBNET_ID"'"]' \
                --service-lb-subnet-ids '["'"$SUBNET_ID"'"]' \
                --pod-cidr 10.244.0.0/16 \
                --service-cidr 10.96.0.0/12 \
                --query "data.id" --raw-output)
    echo "✅ Created Cluster: $CLUSTER_ID"
else
    echo "✅ Using existing Cluster: $CLUSTER_ID"
fi

echo "📋 Step 5: Creating Node Pool..."
NODE_POOL_ID=$(oci ce node-pool list --cluster-id "$CLUSTER_ID" --query "data[?name=='$NODE_POOL_NAME'].id | [0]" --raw-output)
if [ -z "$NODE_POOL_ID" ]; then
    NODE_POOL_ID=$(oci ce node-pool create \
                    --cluster-id "$CLUSTER_ID" \
                    --name "$NODE_POOL_NAME" \
                    --kubernetes-version v1.28.2 \
                    --node-shape VM.Standard.A1.Flex \
                    --node-count 2 \
                    --node-config '{"memoryInGBs": 6, "ocpus": 2}' \
                    --query "data.id" --raw-output)
    echo "✅ Created Node Pool: $NODE_POOL_ID"
else
    echo "✅ Using existing Node Pool: $NODE_POOL_ID"
fi

echo "📋 Step 6: Configuring Kubernetes Access..."
oci ce cluster create-kubeconfig \
    --cluster-id "$CLUSTER_ID" \
    --file "$HOME/.kube/payment-watchdog-config" \
    --region "$REGION"

echo "📋 Step 6.1: Configuring Block Volume Replication (intra-AU)"
REPLICA_REGION="ap-melbourne-1"
if [ "$REGION" = "ap-melbourne-1" ]; then
    REPLICA_REGION="ap-sydney-1"
fi
echo "Configuring Block Volume replication to $REPLICA_REGION for High Availability..."

VOLUME_ID=$(oci bv volume list --compartment-id "$COMPARTMENT_ID" --query "data[?\"display-name\"=='Sovereign-DB-Vol'].id | [0]" --raw-output 2>/dev/null || echo "")
if [ -z "$VOLUME_ID" ] || [ "$VOLUME_ID" = "null" ]; then
    VOLUME_ID=$(oci bv volume create --availability-domain AD-1 --compartment-id "$COMPARTMENT_ID" --size-in-gbs 50 --display-name "Sovereign-DB-Vol" --region "$REGION" --query "data.id" --raw-output)
    echo "Created Block Volume: $VOLUME_ID"
else
    echo "Using existing Block Volume: $VOLUME_ID"
fi

if [ -n "$SOVEREIGN_POLICY_ID" ] && [ -n "$VOLUME_ID" ] && [ "$VOLUME_ID" != "null" ]; then
    oci bv volume-backup-policy-assignment create --asset-id "$VOLUME_ID" --policy-id "$SOVEREIGN_POLICY_ID"
    echo "Block Volume Backup Policy Assigned."
else
    echo "⚠️ SOVEREIGN_POLICY_ID not set or Volume ID missing. Skipping backup policy assignment."
fi

export KUBECONFIG="$HOME/.kube/payment-watchdog-config"
echo "✅ Kubeconfig created: $KUBECONFIG"

echo "📋 Step 7: Installing NGINX Ingress Controller..."
kubectl get namespace ingress-nginx 2>/dev/null || kubectl create namespace ingress-nginx

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
    --namespace ingress-nginx \
    --create-namespace \
    --set controller.service.type=LoadBalancer \
    --set controller.publishService.enabled=true

echo "⏳ Waiting for external IP..."
EXTERNAL_IP=""
for i in {1..30}; do
    EXTERNAL_IP=$(kubectl get svc ingress-nginx-controller -n ingress-nginx -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    if [ -n "$EXTERNAL_IP" ]; then
        echo "✅ External IP assigned: $EXTERNAL_IP"
        break
    fi
    echo "⏳ Waiting for external IP... ($i/30)"
    sleep 10
done

if [ -z "$EXTERNAL_IP" ]; then
    echo "❌ Failed to get external IP. Please check manually:"
    kubectl get svc ingress-nginx-controller -n ingress-nginx
    exit 1
fi

echo "📋 Step 8: Updating Configuration..."
# Update domain in ingress
sed -i "s/lexure-mvp.local/$DOMAIN/g" api/deployments/kubernetes/base/components/ingress/ingress.yaml

# Update GitHub environment URL
sed -i "s/staging.payment-watchdog.example.com/https:\/\/$DOMAIN/g" .github/workflows/payment-watchdog-ci.yml

echo "📋 Step 9: Creating Staging Namespace..."
kubectl create namespace staging --dry-run=client -o yaml | kubectl apply -f -

echo "🎉 Oracle Cloud Setup Complete!"
echo "=================================="
echo "External IP: $EXTERNAL_IP"
echo "Domain: $DOMAIN"
echo "Next Steps:"
echo "1. Point your domain DNS to: $EXTERNAL_IP"
echo "2. Run: kubectl apply -f api/deployments/kubernetes"
echo "3. Update GitHub KUBE_CONFIG secret with:"
echo "   cat $KUBECONFIG | base64 -w 0"
echo "=================================="
