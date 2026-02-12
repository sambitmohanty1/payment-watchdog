#!/bin/bash

# 🔒 Generate Kubernetes TLS Secret from SSL Certificates
# Converts local SSL certificates to Kubernetes TLS secret

set -e

CERT_DIR="certs"
SECRET_FILE="deployments/kubernetes/tls-secret.yaml"

echo "🔒 Generating Kubernetes TLS secret from SSL certificates..."

# Check if certificates exist
if [ ! -f "$CERT_DIR/server.crt" ] || [ ! -f "$CERT_DIR/server.key" ]; then
    echo "❌ SSL certificates not found. Run ./scripts/generate-ssl-certs.sh first."
    exit 1
fi

# Generate base64 encoded certificate and key
CERT_B64=$(base64 -i "$CERT_DIR/server.crt" | tr -d '\n')
KEY_B64=$(base64 -i "$CERT_DIR/server.key" | tr -d '\n')

# Create the TLS secret YAML
cat > "$SECRET_FILE" << EOF
apiVersion: v1
kind: Secret
metadata:
  name: lexure-intelligence-ui-tls
  namespace: lexure-intelligence-ui
  labels:
    app: lexure-intelligence
    component: ui
type: kubernetes.io/tls
data:
  tls.crt: $CERT_B64
  tls.key: $KEY_B64
EOF

echo "✅ Kubernetes TLS secret generated successfully!"
echo "📁 Secret file: $SECRET_FILE"
echo ""
echo "🚀 To apply the secret to Kubernetes:"
echo "   kubectl apply -f $SECRET_FILE"
echo ""
echo "🔧 To verify the secret:"
echo "   kubectl get secret lexure-intelligence-ui-tls -n lexure-intelligence-ui"
