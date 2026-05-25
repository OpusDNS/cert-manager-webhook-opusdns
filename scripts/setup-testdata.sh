#!/usr/bin/env bash
set -euo pipefail

# Prepare testdata for integration tests.
# This script generates config.json and secret from environment variables.
#
# Required environment variables:
#   OPUSDNS_API_KEY       - OpusDNS API key for testing
#   OPUSDNS_API_ENDPOINT  - OpusDNS API endpoint (e.g. https://api.preview1.opusdns.dev)
#
# Optional environment variables:
#   TEST_ZONE_NAME        - DNS zone to use for tests (default: cert-manager-test.opusdns.dev.)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTDATA_DIR="${SCRIPT_DIR}/../testdata/opusdns"

if [ -z "${OPUSDNS_API_KEY:-}" ]; then
    echo "ERROR: OPUSDNS_API_KEY environment variable is required"
    exit 1
fi

if [ -z "${OPUSDNS_API_ENDPOINT:-}" ]; then
    echo "ERROR: OPUSDNS_API_ENDPOINT environment variable is required"
    exit 1
fi

echo "Generating testdata for integration tests..."

# Generate config.json with API endpoint
cat > "${TESTDATA_DIR}/config.json" <<EOF
{
  "apiKeySecretRef": {
    "name": "opusdns-secret",
    "key": "api-key"
  },
  "apiEndpoint": "${OPUSDNS_API_ENDPOINT}",
  "ttl": 120
}
EOF

# Generate secret YAML with API key
cat > "${TESTDATA_DIR}/opusdns-secret.yaml" <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: opusdns-secret
type: Opaque
stringData:
  api-key: "${OPUSDNS_API_KEY}"
EOF

echo "Testdata generated successfully."
echo "  Config: ${TESTDATA_DIR}/config.json"
echo "  Secret: ${TESTDATA_DIR}/opusdns-secret.yaml"
