#!/bin/bash

# Oracle Cloud Cleanup Script
# Usage: ./deployment/cleanup-oracle.sh [--force]

set -e
set -u

# Parse arguments
FORCE_MODE=false
COMPARTMENT_NAME="PaymentWatchdogSovereign"
TENANCY_ID="ocid1.tenancy.oc1..aaaaaaaaitfhbb6ix7yqiavspixzepzm4babf6qjonzom5pe4lvqzxvp2xla"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --force)
            FORCE_MODE=true
            shift
            ;;
        --compartment)
            COMPARTMENT_NAME="$2"
            shift
            ;;
        --region)
            REGION="$2"
            shift
            ;;
        -h|--help)
            echo "Oracle Cloud Cleanup Script"
            echo "Usage: $0 [--force] [--compartment NAME] [--region REGION]"
            echo ""
            echo "Options:"
            echo "  --force         Skip confirmation prompts"
            echo "  --compartment   Specify compartment name (default: $COMPARTMENT_NAME)"
            echo "  --region        Specify region (default: ap-melbourne-1)"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Cleanup default compartment"
            echo "  $0 --force                            # Force cleanup without confirmation"
            echo "  $0 --compartment MyCompartment         # Cleanup specific compartment"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Set defaults
REGION=${REGION:-ap-melbourne-1}

echo "🧹 Oracle Cloud Infrastructure Cleanup"
echo "=================================="
echo "Compartment: $COMPARTMENT_NAME"
echo "Region: $REGION"
echo "Force Mode: $FORCE_MODE"
echo ""

# Get compartment ID
COMPARTMENT_ID=$(oci iam compartment list \
    --compartment-id "$TENANCY_ID" \
    --query "data[?name=='$COMPARTMENT_NAME'].id | [0]" \
    --raw-output)

if [ -z "$COMPARTMENT_ID" ] || [ "$COMPARTMENT_ID" = "null" ]; then
    echo "❌ ERROR: Compartment '$COMPARTMENT_NAME' not found."
    exit 1
fi

echo "📋 Step 1: Deleting Kubernetes Clusters..."
CLUSTERS=$(oci ce cluster list --compartment-id "$COMPARTMENT_ID" --query "data[?\"lifecycle-state\"!='DELETED'].id" --raw-output)
for cluster in $CLUSTERS; do
    if [ -n "$cluster" ] && [ "$cluster" != "null" ]; then
        echo "Deleting cluster: $cluster"
        oci ce cluster delete --cluster-id "$cluster" --force
        echo "✅ Cluster deletion initiated"
    fi
done

echo "⏳ Waiting for cluster deletions to complete..."
sleep 120  # Wait 2 minutes for deletions

echo "📋 Step 2: Deleting Internet Gateways..."
IGW_IDS=$(oci network internet-gateway list --compartment-id "$COMPARTMENT_ID" --query "data[*].id" --raw-output | jq -r '.[]' 2>/dev/null || echo "")
for igw_id in $IGW_IDS; do
    if [ -n "$igw_id" ] && [ "$igw_id" != "null" ]; then
        echo "Deleting Internet Gateway: $igw_id"
        oci network internet-gateway delete --ig-id "$igw_id" --force
        echo "✅ Internet Gateway deleted"
    fi
done

echo "📋 Step 3: Cleaning Route Tables..."
RT_IDS=$(oci network route-table list --compartment-id "$COMPARTMENT_ID" --query "data[*].id" --raw-output | jq -r '.[]' 2>/dev/null || echo "")
for rt_id in $RT_IDS; do
    if [ -n "$rt_id" ] && [ "$rt_id" != "null" ]; then
        echo "Cleaning Route Table: $rt_id"
        oci network route-table update --rt-id "$rt_id" --route-rules '[]' --force
        echo "✅ Route Table cleaned"
    fi
done

echo "📋 Step 4: Deleting Subnets..."
SUBNET_IDS=$(oci network subnet list --compartment-id "$COMPARTMENT_ID" --query "data[*].id" --raw-output | jq -r '.[]' 2>/dev/null || echo "")
for subnet_id in $SUBNET_IDS; do
    if [ -n "$subnet_id" ] && [ "$subnet_id" != "null" ]; then
        echo "Deleting Subnet: $subnet_id"
        oci network subnet delete --subnet-id "$subnet_id" --force
        echo "✅ Subnet deleted"
    fi
done

echo "📋 Step 5: Deleting VCNs..."
VCN_IDS=$(oci network vcn list --compartment-id "$COMPARTMENT_ID" --query "data[*].id" --raw-output | jq -r '.[]' 2>/dev/null || echo "")
for vcn_id in $VCN_IDS; do
    if [ -n "$vcn_id" ] && [ "$vcn_id" != "null" ]; then
        echo "Deleting VCN: $vcn_id"
        oci network vcn delete --vcn-id "$vcn_id" --force
        echo "✅ VCN deleted"
    fi
done

echo "📋 Step 6: Verifying Cleanup..."
echo ""
echo "=== Remaining Resources ==="
echo "Clusters:"
oci ce cluster list --compartment-id "$COMPARTMENT_ID" --query "data[*].{name:name, state:\"lifecycle-state\"}" --output table
echo ""
echo "VCNs:"
oci network vcn list --compartment-id "$COMPARTMENT_ID" --query "data[*].{name:name, cidr:\"cidr-block\"}" --output table
echo ""
echo "Internet Gateways:"
oci network internet-gateway list --compartment-id "$COMPARTMENT_ID" --query "data[*].name" --output table

echo ""
echo "✅ Cleanup Complete!"
echo "=================================="
