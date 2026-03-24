#!/bin/bash

# Script to update GitHub Actions KUBE_CONFIG secret
# This updates the secret with the correct cluster configuration

echo "🔧 Updating GitHub Actions KUBE_CONFIG secret..."

# Get the current kubeconfig
KUBECONFIG_BASE64=$(cat /tmp/current-kubeconfig | base64 -w 0)

echo "📋 New kubeconfig details:"
echo "Context: context-cumd24tyoaq"
echo "Cluster: https://207.211.153.246:6443"
echo "User: user-cumd24tyoaq"

echo ""
echo "🚀 To update the GitHub secret, run:"
echo "gh secret set KUBE_CONFIG --body='$KUBECONFIG_BASE64'"
echo ""
echo "🔍 Or manually update in GitHub:"
echo "1. Go to: https://github.com/sambitmohanty1/payment-watchdog/settings/secrets/actions"
echo "2. Update 'KUBE_CONFIG' secret with the base64 value above"
echo ""
echo "📝 Base64 kubeconfig:"
echo "$KUBECONFIG_BASE64"
