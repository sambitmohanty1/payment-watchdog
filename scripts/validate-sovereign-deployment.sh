#!/bin/bash

# Payment Watchdog - Sovereign AU Deployment Validation
# Prevents base deployments from being applied to sovereign-au namespace

set -euo pipefail

NAMESPACE="sovereign-au"
SCRIPT_NAME="validate-sovereign-deployment.sh"

echo "🔍 Validating sovereign-au deployment configuration..."

# Check if we're in the right directory
if [[ ! -f "api/deployments/kubernetes/overlays/sovereign-au/kustomization.yaml" ]]; then
    echo "❌ Error: Must be run from project root directory"
    echo "Expected: api/deployments/kubernetes/overlays/sovereign-au/kustomization.yaml"
    exit 1
fi

# Check for conflicting deployments in sovereign-au namespace
echo "📊 Checking for conflicting deployments in namespace: $NAMESPACE"
CONFLICTING_DEPS=$(kubectl get deployments -n "$NAMESPACE" -o name 2>/dev/null | grep -v "sovereign-au" | wc -l || echo "0")

if [[ "$CONFLICTING_DEPS" -gt 0 ]]; then
    echo "❌ Found $CONFLICTING_DEPS base deployments in $NAMESPACE namespace:"
    kubectl get deployments -n "$NAMESPACE" -o name | grep -v "sovereign-au" || true
    echo ""
    echo "🔧 To fix: Delete base deployments before applying sovereign overlay"
    echo "kubectl delete deployment payment-watchdog-api -n $NAMESPACE --ignore-not-found=true"
    echo "kubectl delete deployment payment-watchdog-ui -n $NAMESPACE --ignore-not-found=true"  
    echo "kubectl delete deployment recovery-orchestration -n $NAMESPACE --ignore-not-found=true"
    exit 1
fi

# Check if sovereign-au overlay exists
echo "📁 Checking sovereign-au overlay configuration..."
OVERLAY_DIR="api/deployments/kubernetes/overlays/sovereign-au"

if [[ ! -d "$OVERLAY_DIR" ]]; then
    echo "❌ Error: Sovereign AU overlay directory not found: $OVERLAY_DIR"
    exit 1
fi

# Validate kustomization.yaml
KUSTOMIZATION_FILE="$OVERLAY_DIR/kustomization.yaml"
if [[ ! -f "$KUSTOMIZATION_FILE" ]]; then
    echo "❌ Error: Kustomization file not found: $KUSTOMIZATION_FILE"
    exit 1
fi

# Check namespace is correctly set
NAMESPACE_CHECK=$(grep -E "^\s*namespace:" "$KUSTOMIZATION_FILE" | grep "sovereign-au" | wc -l)
if [[ "$NAMESPACE_CHECK" -eq 0 ]]; then
    echo "❌ Error: Namespace not set to sovereign-au in kustomization.yaml"
    echo "Expected: namespace: sovereign-au"
    exit 1
fi

# Check nameSuffix is set
SUFFIX_CHECK=$(grep -E "^\s*nameSuffix:" "$KUSTOMIZATION_FILE" | grep "sovereign-au" | wc -l)
if [[ "$SUFFIX_CHECK" -eq 0 ]]; then
    echo "❌ Error: nameSuffix not set to -sovereign-au in kustomization.yaml"
    echo "Expected: nameSuffix: \"-sovereign-au\""
    exit 1
fi

# Check secretGenerator exists
SECRET_CHECK=$(grep -A 10 "secretGenerator:" "$KUSTOMIZATION_FILE" | grep "recovery-orchestration-secrets" | wc -l)
if [[ "$SECRET_CHECK" -eq 0 ]]; then
    echo "❌ Error: recovery-orchestration-secrets not found in secretGenerator"
    exit 1
fi

# Check behavior is set to merge
BEHAVIOR_CHECK=$(grep -A 5 "secretGenerator:" "$KUSTOMIZATION_FILE" | grep "behavior: merge" | wc -l)
if [[ "$BEHAVIOR_CHECK" -eq 0 ]]; then
    echo "❌ Error: secretGenerator behavior not set to merge"
    echo "Expected: behavior: merge"
    exit 1
fi

# Validate patch files exist
PATCH_FILES=("deployment-patch.yaml" "configmap-patch.yaml")
for patch_file in "${PATCH_FILES[@]}"; do
    if [[ ! -f "$OVERLAY_DIR/$patch_file" ]]; then
        echo "❌ Error: Patch file not found: $OVERLAY_DIR/$patch_file"
        exit 1
    fi
done

# Check deployment patch references correct secret names
DEPLOYMENT_PATCH="$OVERLAY_DIR/deployment-patch.yaml"
SECRET_REF_CHECK=$(grep -c "name: recovery-orchestration-secrets" "$DEPLOYMENT_PATCH" || echo "0")
if [[ "$SECRET_REF_CHECK" -lt 2 ]]; then
    echo "❌ Error: deployment-patch.yaml should reference recovery-orchestration-secrets (without suffix)"
    exit 1
fi

# Check configmap patch references correct names
CONFIGMAP_PATCH="$OVERLAY_DIR/configmap-patch.yaml"
CONFIG_REF_CHECK=$(grep -c "name: recovery-orchestration-config" "$CONFIGMAP_PATCH" || echo "0")
if [[ "$CONFIG_REF_CHECK" -lt 1 ]]; then
    echo "❌ Error: configmap-patch.yaml should reference recovery-orchestration-config (without suffix)"
    exit 1
fi

echo "✅ All validation checks passed!"
echo "🚀 Ready to apply sovereign-au configuration:"
echo "kubectl apply -k api/deployments/kubernetes/overlays/sovereign-au"
