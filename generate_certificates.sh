#!/bin/bash

# Create a directory for certificates
mkdir -p certs

# Generate a private key and self-signed certificate
echo "Generating self-signed certificate for localhost development..."
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout certs/server.key \
  -out certs/server.crt \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

# Set proper permissions
chmod 600 certs/server.key
chmod 644 certs/server.crt

echo "Self-signed certificate generated successfully!"
echo "  - Certificate: certs/server.crt"
echo "  - Private key: certs/server.key"
echo ""
echo "Note: These certificates are for development only and will cause security warnings in browsers." 