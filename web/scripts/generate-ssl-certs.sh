#!/bin/bash

# 🔒 SSL Certificate Generation Script for Local Development
# Generates self-signed certificates for HTTPS development

set -e

CERT_DIR="certs"
DOMAIN="localhost"

echo "🔒 Generating SSL certificates for local development..."

# Create certs directory if it doesn't exist
mkdir -p "$CERT_DIR"

# Generate private key
echo "📝 Generating private key..."
openssl genrsa -out "$CERT_DIR/server.key" 2048

# Generate certificate signing request
echo "📝 Generating certificate signing request..."
openssl req -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" -subj "/C=AU/ST=NSW/L=Sydney/O=Lexure Intelligence/OU=Development/CN=$DOMAIN"

# Generate self-signed certificate with proper key usage
echo "📝 Generating self-signed certificate..."
openssl x509 -req -days 365 -in "$CERT_DIR/server.csr" -signkey "$CERT_DIR/server.key" -out "$CERT_DIR/server.crt" -extensions v3_req -extfile <(
cat <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = AU
ST = NSW
L = Sydney
O = Lexure Intelligence
OU = Development
CN = $DOMAIN

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = $DOMAIN
DNS.2 = *.localhost
DNS.3 = localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF
)

# Set proper permissions
chmod 600 "$CERT_DIR/server.key"
chmod 644 "$CERT_DIR/server.crt"

# Clean up CSR file
rm "$CERT_DIR/server.csr"

echo "✅ SSL certificates generated successfully!"
echo ""
echo "📁 Certificate files:"
echo "   Private Key: $CERT_DIR/server.key"
echo "   Certificate: $CERT_DIR/server.crt"
echo ""
echo "🔧 To trust the certificate in your browser:"
echo "   1. Open Chrome/Safari"
echo "   2. Go to https://localhost:3050"
echo "   3. Click 'Advanced' → 'Proceed to localhost (unsafe)'"
echo ""
echo "🚀 You can now run the UI with HTTPS:"
echo "   npm run dev:https"
