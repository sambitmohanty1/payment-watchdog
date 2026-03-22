#!/bin/bash

# Oracle Cloud Infrastructure Cleanup Script - Improved Version
# Usage: ./deployment/cleanup-oracle-improved.sh [--force] [--compartment NAME] [--region REGION]

set -euo pipefail

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
            shift 2
            ;;
        --region)
            REGION="$2"
            shift 2
            ;;
        --all)
            COMPARTMENT_NAME="PaymentWatchdogSovereign"
            echo "🔧 Will clean up all Payment Watchdog compartments"
            shift
            ;;
        -h|--help)
            echo "Oracle Cloud Infrastructure Cleanup Script - Improved"
            echo "Usage: $0 [--force] [--compartment NAME] [--region REGION] [--all]"
            echo ""
            echo "Options:"
            echo "  --force         Skip confirmation prompts"
            echo "  --compartment   Specify compartment name (default: $COMPARTMENT_NAME)"
            echo "  --region        Specify region (default: ap-melbourne-1)"
            echo "  --all           Clean up all Payment Watchdog related compartments"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Cleanup default compartment"
            echo "  $0 --force                            # Force cleanup without confirmation"
            echo "  $0 --compartment MyCompartment         # Cleanup specific compartment"
            echo "  $0 --all                              # Cleanup all Payment Watchdog compartments"
            exit 0
            ;;
        *)
            echo "❌ Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Set defaults
REGION=${REGION:-ap-melbourne-1}

echo "🧹 Oracle Cloud Infrastructure Cleanup - Improved"
echo "==============================================="
echo "Compartment: $COMPARTMENT_NAME"
echo "Region: $REGION"
echo "Force Mode: $FORCE_MODE"
echo ""

# Function to cleanup a single compartment
cleanup_compartment() {
    local compartment_name="$1"
    local compartment_id="$2"
    
    echo "🔍 Cleaning up compartment: $compartment_name ($compartment_id)"
    echo ""
    
    # Step 1: Delete Kubernetes Clusters
    echo "📋 Step 1: Deleting Kubernetes Clusters..."
    clusters=$(oci ce cluster list --compartment-id "$compartment_id" --query "data[?\"lifecycle-state\"!='DELETED'].{id:id, name:name, state:\"lifecycle-state\"}" --output table 2>/dev/null || echo "No clusters found")
    
    if [[ "$clusters" != "No clusters found" ]]; then
        cluster_ids=$(oci ce cluster list --compartment-id "$compartment_id" --query "data[?\"lifecycle-state\"!='DELETED'].id" --raw-output)
        echo "  Raw cluster IDs output: $cluster_ids"
        # Parse JSON array properly using jq
        cluster_ids=$(echo "$cluster_ids" | jq -r '.[]' 2>/dev/null || echo "")
        echo "  Parsed cluster IDs: $cluster_ids"
        for cluster_id in $cluster_ids; do
            if [ -n "$cluster_id" ] && [ "$cluster_id" != "null" ] && [ "$cluster_id" != "[" ] && [ "$cluster_id" != "]" ]; then
                echo "  Deleting cluster: $cluster_id"
                oci ce cluster delete --cluster-id "$cluster_id" --force || echo "  ⚠️  Cluster deletion may have already been initiated"
                echo "  ✅ Cluster deletion initiated"
            else
                echo "  ⚠️  Invalid cluster ID format: $cluster_id"
            fi
        done
        
        echo "⏳ Waiting for cluster deletions to complete..."
        sleep 120  # Wait 2 minutes for deletions
    else
        echo "  ✅ No clusters to delete"
    fi
    
    # Step 2: Delete Node Pools (in case clusters didn't delete properly)
    echo "📋 Step 2: Deleting Node Pools..."
    node_pools=$(oci ce node-pool list --compartment-id "$compartment_id" --query "data[?\"lifecycle-state\"!='DELETED'].{id:id, name:name, state:\"lifecycle-state\"}" --output table 2>/dev/null || echo "No node pools found")
    
    if [[ "$node_pools" != "No node pools found" ]]; then
        node_pool_ids=$(oci ce node-pool list --compartment-id "$compartment_id" --query "data[?\"lifecycle-state\"!='DELETED'].id" --raw-output)
        echo "  Raw node pool IDs output: $node_pool_ids"
        # Parse JSON array properly using jq
        node_pool_ids=$(echo "$node_pool_ids" | jq -r '.[]' 2>/dev/null || echo "")
        echo "  Parsed node pool IDs: $node_pool_ids"
        for node_pool_id in $node_pool_ids; do
            if [ -n "$node_pool_id" ] && [ "$node_pool_id" != "null" ] && [ "$node_pool_id" != "[" ] && [ "$node_pool_id" != "]" ]; then
                echo "  Deleting node pool: $node_pool_id"
                oci ce node-pool delete --node-pool-id "$node_pool_id" --force || echo "  ⚠️  Node pool deletion may have already been initiated"
                echo "  ✅ Node pool deletion initiated"
            else
                echo "  ⚠️  Invalid node pool ID format: $node_pool_id"
            fi
        done
        
        echo "⏳ Waiting for node pool deletions to complete..."
        sleep 60
    else
        echo "  ✅ No node pools to delete"
    fi
    
    # Step 3: Delete Load Balancers
    echo "📋 Step 3: Deleting Load Balancers..."
    lb_ids=$(oci lb load-balancer list --compartment-id "$compartment_id" --query "data[?\"lifecycle-state\"!='DELETED'].id" --raw-output 2>/dev/null || echo "")
    echo "  Raw load balancer IDs output: $lb_ids"
    # Parse JSON array properly using jq
    lb_ids=$(echo "$lb_ids" | jq -r '.[]' 2>/dev/null || echo "")
    echo "  Parsed load balancer IDs: $lb_ids"
    for lb_id in $lb_ids; do
        if [ -n "$lb_id" ] && [ "$lb_id" != "null" ] && [ "$lb_id" != "[" ] && [ "$lb_id" != "]" ]; then
            echo "  Deleting Load Balancer: $lb_id"
            oci lb load-balancer delete --load-balancer-id "$lb_id" --force || echo "  ⚠️  Load balancer deletion may have already been initiated"
            echo "  ✅ Load balancer deletion initiated"
        else
            echo "  ⚠️  Invalid load balancer ID format: $lb_id"
        fi
    done
    
    # Step 4: Delete Internet Gateways
    echo "📋 Step 4: Deleting Internet Gateways..."
    igw_ids=$(oci network internet-gateway list --compartment-id "$compartment_id" --query "data[*].id" --raw-output 2>/dev/null || echo "")
    echo "  Raw internet gateway IDs output: $igw_ids"
    # Parse JSON array properly using jq
    igw_ids=$(echo "$igw_ids" | jq -r '.[]' 2>/dev/null || echo "")
    echo "  Parsed internet gateway IDs: $igw_ids"
    for igw_id in $igw_ids; do
        if [ -n "$igw_id" ] && [ "$igw_id" != "null" ] && [ "$igw_id" != "[" ] && [ "$igw_id" != "]" ]; then
            echo "  Deleting Internet Gateway: $igw_id"
            oci network internet-gateway delete --ig-id "$igw_id" --force || echo "  ⚠️  Internet gateway deletion may have already been initiated"
            echo "  ✅ Internet gateway deleted"
        else
            echo "  ⚠️  Invalid internet gateway ID format: $igw_id"
        fi
    done
    
    # Step 5: Clean Route Tables
    echo "📋 Step 5: Cleaning Route Tables..."
    rt_ids=$(oci network route-table list --compartment-id "$compartment_id" --query "data[*].id" --raw-output 2>/dev/null || echo "")
    echo "  Raw route table IDs output: $rt_ids"
    # Parse JSON array properly using jq
    rt_ids=$(echo "$rt_ids" | jq -r '.[]' 2>/dev/null || echo "")
    echo "  Parsed route table IDs: $rt_ids"
    for rt_id in $rt_ids; do
        if [ -n "$rt_id" ] && [ "$rt_id" != "null" ] && [ "$rt_id" != "[" ] && [ "$rt_id" != "]" ]; then
            echo "  Cleaning Route Table: $rt_id"
            oci network route-table update --rt-id "$rt_id" --route-rules '[]' --force || echo "  ⚠️  Route table cleanup may have already been done"
            echo "  ✅ Route table cleaned"
        else
            echo "  ⚠️  Invalid route table ID format: $rt_id"
        fi
    done
    
    # Step 6: Delete Subnets
    echo "📋 Step 6: Deleting Subnets..."
    subnet_ids=$(oci network subnet list --compartment-id "$compartment_id" --query "data[*].id" --raw-output 2>/dev/null || echo "")
    echo "  Raw subnet IDs output: $subnet_ids"
    # Parse JSON array properly using jq
    subnet_ids=$(echo "$subnet_ids" | jq -r '.[]' 2>/dev/null || echo "")
    echo "  Parsed subnet IDs: $subnet_ids"
    for subnet_id in $subnet_ids; do
        if [ -n "$subnet_id" ] && [ "$subnet_id" != "null" ] && [ "$subnet_id" != "[" ] && [ "$subnet_id" != "]" ]; then
            echo "  Deleting Subnet: $subnet_id"
            oci network subnet delete --subnet-id "$subnet_id" --force || echo "  ⚠️  Subnet deletion may have already been initiated"
            echo "  ✅ Subnet deleted"
        else
            echo "  ⚠️  Invalid subnet ID format: $subnet_id"
        fi
    done
    
    # Step 7: Delete VCNs
    echo "📋 Step 7: Deleting VCNs..."
    vcn_ids=$(oci network vcn list --compartment-id "$compartment_id" --query "data[*].id" --raw-output 2>/dev/null || echo "")
    echo "  Raw VCN IDs output: $vcn_ids"
    # Parse JSON array properly using jq
    vcn_ids=$(echo "$vcn_ids" | jq -r '.[]' 2>/dev/null || echo "")
    echo "  Parsed VCN IDs: $vcn_ids"
    for vcn_id in $vcn_ids; do
        if [ -n "$vcn_id" ] && [ "$vcn_id" != "null" ] && [ "$vcn_id" != "[" ] && [ "$vcn_id" != "]" ]; then
            echo "  Deleting VCN: $vcn_id"
            oci network vcn delete --vcn-id "$vcn_id" --force || echo "  ⚠️  VCN deletion may have already been initiated"
            echo "  ✅ VCN deleted"
        else
            echo "  ⚠️  Invalid VCN ID format: $vcn_id"
        fi
    done
    
    echo "✅ Cleanup completed for compartment: $compartment_name"
    echo ""
}

# Function to get all Payment Watchdog related compartments
get_payment_watchdog_compartments() {
    local compartments=$(oci iam compartment list --compartment-id "$TENANCY_ID" --query "data[?contains(name, 'PaymentWatchdog') || contains(name, 'payment-watchdog')].{id:id, name:name}" --output table)
    echo "$compartments"
}

# Main cleanup logic
if [ "$COMPARTMENT_NAME" = "PaymentWatchdogSovereign" ] && [ "${1:-}" != "--all" ]; then
    # Single compartment cleanup (original behavior)
    echo "🔍 Looking for compartment: $COMPARTMENT_NAME"
    
    COMPARTMENT_ID=$(oci iam compartment list \
        --compartment-id "$TENANCY_ID" \
        --query "data[?name=='$COMPARTMENT_NAME'].id | [0]" \
        --raw-output)
    
    if [ -z "$COMPARTMENT_ID" ] || [ "$COMPARTMENT_ID" = "null" ]; then
        echo "❌ ERROR: Compartment '$COMPARTMENT_NAME' not found."
        echo ""
        echo "Available compartments:"
        oci iam compartment list --compartment-id "$TENANCY_ID" --query "data[*].{name:name, id:id}" --output table
        exit 1
    fi
    
    # Show what will be deleted
    echo "📋 Resources to be deleted in compartment '$COMPARTMENT_NAME':"
    echo ""
    
    echo "Kubernetes Clusters:"
    oci ce cluster list --compartment-id "$COMPARTMENT_ID" --query "data[?\"lifecycle-state\"!='DELETED'].{name:name, state:\"lifecycle-state\"}" --output table 2>/dev/null || echo "  No clusters found"
    echo ""
    
    echo "VCNs:"
    oci network vcn list --compartment-id "$COMPARTMENT_ID" --query "data[*].{name:name, cidr:\"cidr-block\"}" --output table 2>/dev/null || echo "  No VCNs found"
    echo ""
    
    echo "Load Balancers:"
    oci lb load-balancer list --compartment-id "$COMPARTMENT_ID" --query "data[?\"lifecycle-state\"!='DELETED'].{name:name, state:\"lifecycle-state\"}" --output table 2>/dev/null || echo "  No load balancers found"
    echo ""
    
    # Ask for confirmation unless force mode
    if [ "$FORCE_MODE" != "true" ]; then
        read -p "⚠️  This will permanently delete ALL resources in compartment '$COMPARTMENT_NAME'. Continue? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "❌ Cleanup cancelled by user."
            exit 0
        fi
    else
        echo "🔧 FORCE MODE: Skipping confirmation"
    fi
    
    cleanup_compartment "$COMPARTMENT_NAME" "$COMPARTMENT_ID"
    
elif [ "${1:-}" = "--all" ]; then
    # Multiple compartment cleanup
    echo "🔍 Looking for all Payment Watchdog related compartments..."
    echo ""
    
    compartments=$(get_payment_watchdog_compartments)
    echo "$compartments"
    echo ""
    
    if [ "$FORCE_MODE" != "true" ]; then
        read -p "⚠️  This will permanently delete ALL resources in ALL Payment Watchdog compartments. Continue? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "❌ Cleanup cancelled by user."
            exit 0
        fi
    else
        echo "🔧 FORCE MODE: Skipping confirmation"
    fi
    
    # Extract compartment IDs and clean up each
    compartment_ids=$(oci iam compartment list --compartment-id "$TENANCY_ID" --query "data[?contains(name, 'PaymentWatchdog') || contains(name, 'payment-watchdog')].id" --raw-output)
    compartment_names=$(oci iam compartment list --compartment-id "$TENANCY_ID" --query "data[?contains(name, 'PaymentWatchdog') || contains(name, 'payment-watchdog')].name" --raw-output)
    
    i=1
    for comp_id in $compartment_ids; do
        if [ -n "$comp_id" ] && [ "$comp_id" != "null" ]; then
            comp_name=$(echo "$compartment_names" | sed -n "${i}p")
            cleanup_compartment "$comp_name" "$comp_id"
        fi
        i=$((i+1))
    done
fi

# Step 8: Final Verification
echo "📋 Step 8: Verifying Cleanup..."
echo ""

if [ "${1:-}" = "--all" ]; then
    echo "=== Remaining Payment Watchdog Resources ==="
    echo "Clusters:"
    get_payment_watchdog_compartments | while read -r line; do
        comp_name=$(echo "$line" | awk '{print $2}')
        comp_id=$(echo "$line" | awk '{print $1}')
        if [ -n "$comp_id" ] && [ "$comp_id" != "null" ]; then
            oci ce cluster list --compartment-id "$comp_id" --query "data[?\"lifecycle-state\"!='DELETED'].{name:name, state:\"lifecycle-state\"}" --output table 2>/dev/null || true
        fi
    done
    echo ""
    
    echo "VCNs:"
    get_payment_watchdog_compartments | while read -r line; do
        comp_name=$(echo "$line" | awk '{print $2}')
        comp_id=$(echo "$line" | awk '{print $1}')
        if [ -n "$comp_id" ] && [ "$comp_id" != "null" ]; then
            oci network vcn list --compartment-id "$comp_id" --query "data[*].{name:name, cidr:\"cidr-block\"}" --output table 2>/dev/null || true
        fi
    done
else
    echo "=== Remaining Resources in $COMPARTMENT_NAME ==="
    echo "Clusters:"
    oci ce cluster list --compartment-id "$COMPARTMENT_ID" --query "data[?\"lifecycle-state\"!='DELETED'].{name:name, state:\"lifecycle-state\"}" --output table 2>/dev/null || echo "  No clusters found"
    echo ""
    
    echo "VCNs:"
    oci network vcn list --compartment-id "$COMPARTMENT_ID" --query "data[*].{name:name, cidr:\"cidr-block\"}" --output table 2>/dev/null || echo "  No VCNs found"
    echo ""
    
    echo "Load Balancers:"
    oci lb load-balancer list --compartment-id "$COMPARTMENT_ID" --query "data[?\"lifecycle-state\"!='DELETED'].{name:name, state:\"lifecycle-state\"}" --output table 2>/dev/null || echo "  No load balancers found"
    echo ""
    
    echo "Internet Gateways:"
    oci network internet-gateway list --compartment-id "$COMPARTMENT_ID" --query "data[*].name" --output table 2>/dev/null || echo "  No internet gateways found"
fi

# Clean up local kubeconfig files
echo "📋 Step 9: Cleaning up local kubeconfig files..."
kubeconfig_files=(
    "$HOME/.kube/payment-watchdog-config"
    "$HOME/.kube/payment-watchdog-free-config"
    "$HOME/.kube/config"
)

for kubeconfig_file in "${kubeconfig_files[@]}"; do
    if [ -f "$kubeconfig_file" ]; then
        echo "  Removing kubeconfig: $kubeconfig_file"
        rm -f "$kubeconfig_file"
        echo "  ✅ Removed"
    fi
done

echo ""
echo "✅ Cleanup Complete!"
echo "=================================="
echo ""
echo "📝 Notes:"
echo "- Some resources may take a few minutes to be fully deleted"
echo "- Check the OCI Console for final verification"
echo "- Kubeconfig files have been cleaned up"
echo ""
echo "🔄 To redeploy, run:"
echo "  ./deployment/oracle-cloud-setup-improved.sh"
