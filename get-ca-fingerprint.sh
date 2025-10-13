#!/bin/bash

# Helper script to get the CA root fingerprint
# Usage: ./get-ca-fingerprint.sh <CA_URL>

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <CA_URL>"
    echo "Example: $0 https://ca.home:9000"
    exit 1
fi

CA_URL="$1"
TEMP_CERT=$(mktemp)

echo "Fetching root certificate from ${CA_URL}..."

# Check if python3 or jq is available for JSON parsing
if command -v python3 >/dev/null 2>&1; then
    JSON_PARSER="python3"
elif command -v jq >/dev/null 2>&1; then
    JSON_PARSER="jq"
else
    echo "Error: Neither python3 nor jq found. Please install one of them."
    exit 1
fi

# Download the root certificate from /roots endpoint (returns JSON)
RESPONSE=$(curl -sk "${CA_URL}/roots" 2>/dev/null)

if [ $? -eq 0 ] && [ -n "$RESPONSE" ]; then
    echo "Root certificate downloaded successfully."
    echo ""
    echo "Calculating fingerprint..."
    
    # Extract certificate from JSON and calculate fingerprint
    if [ "$JSON_PARSER" = "python3" ]; then
        FINGERPRINT=$(echo "$RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['crts'][0])" | openssl x509 -noout -fingerprint -sha256 2>/dev/null | cut -d= -f2 | tr -d ':' | tr '[:upper:]' '[:lower:]')
    else
        FINGERPRINT=$(echo "$RESPONSE" | jq -r '.crts[0]' | openssl x509 -noout -fingerprint -sha256 2>/dev/null | cut -d= -f2 | tr -d ':' | tr '[:upper:]' '[:lower:]')
    fi
    
    if [ -n "$FINGERPRINT" ]; then
        echo ""
        echo "========================================="
        echo "CA Root Fingerprint:"
        echo "$FINGERPRINT"
        echo "========================================="
        echo ""
        echo "Add this to your .env file as:"
        echo "CA_ROOT_FINGERPRINT=$FINGERPRINT"
        echo ""
    else
        echo "Error: Could not calculate fingerprint"
        exit 1
    fi
else
    echo "Error: Could not download root certificate from ${CA_URL}/roots"
    echo ""
    echo "Please make sure:"
    echo "1. The CA URL is correct"
    echo "2. The Step-CA server is running"
    echo "3. The /roots endpoint is accessible"
    exit 1
fi

# Clean up
rm -f "${TEMP_CERT}"

