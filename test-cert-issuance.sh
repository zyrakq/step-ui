#!/bin/bash

# Test script to verify step CLI certificate issuance works correctly

set -e

echo "========================================="
echo "Testing Step CLI Certificate Issuance"
echo "========================================="
echo ""

# Check environment variables
echo "1. Checking environment variables..."
echo "   CA_URL: ${CA_URL}"
echo "   CA_ROOT_FINGERPRINT: ${CA_ROOT_FINGERPRINT}"
echo "   PROVISIONER_NAME: ${PROVISIONER_NAME}"
echo ""

if [ -z "$CA_URL" ] || [ -z "$CA_ROOT_FINGERPRINT" ] || [ -z "$PROVISIONER_NAME" ] || [ -z "$PROVISIONER_PASSWORD" ]; then
    echo "ERROR: Missing required environment variables"
    echo "Please ensure CA_URL, CA_ROOT_FINGERPRINT, PROVISIONER_NAME, and PROVISIONER_PASSWORD are set"
    exit 1
fi

# Test inside backend container
echo "2. Testing inside backend container..."
docker exec step-ui-backend-1 sh -c '
  echo "Creating temp directory..."
  TEMP_DIR=$(mktemp -d)
  cd "$TEMP_DIR"
  
  echo "Downloading root certificate..."
  if ! step ca root root.crt --ca-url "$CA_URL" --fingerprint "$CA_ROOT_FINGERPRINT" 2>&1; then
    echo "FAILED: Could not download root certificate"
    exit 1
  fi
  echo "✓ Root certificate downloaded"
  
  echo "Writing provisioner password to file..."
  echo -n "$PROVISIONER_PASSWORD" > password.txt
  
  echo "Generating token..."
  TOKEN_OUTPUT=$(step ca token test.example.com \
    --ca-url "$CA_URL" \
    --root root.crt \
    --provisioner "$PROVISIONER_NAME" \
    --provisioner-password-file password.txt 2>&1)
  
  if [ $? -ne 0 ]; then
    echo "FAILED: Could not generate token"
    echo "Error: $TOKEN_OUTPUT"
    exit 1
  fi
  
  # Extract only the JWT token (last line, base64 format)
  TOKEN=$(echo "$TOKEN_OUTPUT" | grep -E "^ey" | tail -1)
  
  if [ -z "$TOKEN" ]; then
    echo "FAILED: Could not extract token from output"
    echo "Output: $TOKEN_OUTPUT"
    exit 1
  fi
  
  echo "✓ Token generated successfully"
  
  echo "Issuing certificate..."
  if ! step ca certificate test.example.com cert.crt cert.key \
    --token "$TOKEN" \
    --ca-url "$CA_URL" \
    --root root.crt \
    --not-after 24h 2>&1; then
    echo "FAILED: Could not issue certificate"
    exit 1
  fi
  echo "✓ Certificate issued successfully"
  
  echo "Verifying certificate..."
  if ! step certificate inspect cert.crt --short 2>&1; then
    echo "FAILED: Could not inspect certificate"
    exit 1
  fi
  
  echo "✓ Certificate verified"
  
  # Cleanup
  cd /
  rm -rf "$TEMP_DIR"
  
  echo ""
  echo "========================================="
  echo "ALL TESTS PASSED"
  echo "========================================="
'

echo ""
echo "If this test passes, the issue is in the Go code implementation."
echo "If this test fails, there's an environment or step CLI configuration issue."

