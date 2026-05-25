#!/usr/bin/env bash
set -e

# Fetch test binaries required for conformance tests

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KUBEBUILDER_VERSION=4.11.0

OS=$(go env GOOS)
ARCH=$(go env GOARCH)

mkdir -p "${SCRIPT_DIR}/../kubebuilder/bin"
cd "${SCRIPT_DIR}/../kubebuilder/bin"

# Download kubebuilder tools
echo "Downloading kubebuilder tools..."
curl -sSL "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${KUBEBUILDER_VERSION}-${OS}-${ARCH}.tar.gz" | tar -xvz --strip-components=2

echo "Test binaries downloaded successfully."
