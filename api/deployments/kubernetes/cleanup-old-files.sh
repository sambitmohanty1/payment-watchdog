#!/bin/bash
set -e

echo "Cleaning up old deployment files..."

# Remove old individual deployment files
rm -f app-deployment.yaml
rm -f postgres-deployment.yaml
rm -f redis-deployment.yaml
rm -f configmap.yaml
rm -f secret.yaml
rm -f namespace.yaml
rm -f storage-class.yaml
rm -f postgres-persistent-volume.yaml
rm -f redis-persistent-volume.yaml
rm -f postgres-init-configmap.yaml
rm -f migrations-configmap.yaml
rm -f ingress.yaml
rm -f egress.yaml

# Remove old recovery-orchestration files
rm -rf recovery-orchestration/

# Remove old base directory if empty
rmdir base/recovery-orchestration 2>/dev/null || true

# Remove old base directory if empty
rmdir base 2>/dev/null || true

echo "Cleanup complete. The new structure is now in place."
echo "You can now use kustomize to manage your deployments:"
echo "kubectl apply -k ."
