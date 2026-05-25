# -*- mode: Python -*-

# Tiltfile for local development of cert-manager-webhook-opusdns
# Prerequisites:
#   - Local Kubernetes cluster (kind, k3d, minikube, etc.)
#   - cert-manager installed in the cluster
#   - Tilt installed (https://tilt.dev)

# Configuration
config.define_string('namespace', args=False, usage='Kubernetes namespace for deployment')
config.define_string('registry', args=False, usage='Container registry to push images')
config.define_string('api_key', args=False, usage='OpusDNS API key for testing')
cfg = config.parse()

namespace = cfg.get('namespace', 'cert-manager')
registry = cfg.get('registry', '')  # Empty for local-only builds
api_key = cfg.get('api_key', '')

# Build the webhook image
docker_build(
    'ghcr.io/opusdns/cert-manager-webhook-opusdns',
    '.',
    dockerfile='Dockerfile',
    live_update=[
        sync('.', '/app'),
        run('cd /app && go build -o webhook .', trigger=['*.go', 'go.mod', 'go.sum']),
    ],
    # For faster iteration, use a simpler build
    build_args={
        'TARGETOS': 'linux',
        'TARGETARCH': 'amd64',
    },
)

# Deploy using Helm chart
k8s_yaml(helm(
    'deploy/cert-manager-webhook-opusdns',
    name='cert-manager-webhook-opusdns',
    namespace=namespace,
    values=[
        'deploy/cert-manager-webhook-opusdns/values.yaml',
    ],
    set=[
        'image.pullPolicy=Never',  # Use local image
        'image.tag=latest',
        'replicaCount=1',
    ],
))

# Create test secret if API key is provided
if api_key:
    k8s_yaml(blob("""
apiVersion: v1
kind: Secret
metadata:
  name: opusdns-secret
  namespace: {namespace}
type: Opaque
stringData:
  api-key: "{api_key}"
""".format(namespace=namespace, api_key=api_key)))

# Resource configuration
k8s_resource(
    'cert-manager-webhook-opusdns',
    port_forwards=[],  # Webhook doesn't need port forwarding
    labels=['webhook'],
    resource_deps=['cert-manager'],  # Depends on cert-manager being ready
)

# Optional: Deploy a test ClusterIssuer for quick testing
local_resource(
    'deploy-test-issuer',
    cmd='kubectl apply -f testdata/clusterissuer.yaml || true',
    deps=['testdata/clusterissuer.yaml'],
    labels=['test'],
    auto_init=False,  # Don't run automatically
    trigger_mode=TRIGGER_MODE_MANUAL,
)

# Optional: Run conformance tests
local_resource(
    'run-tests',
    cmd='go test -v ./...',
    labels=['test'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
)

# Print helpful info
print("""
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  cert-manager-webhook-opusdns - Local Development
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Prerequisites:
  1. cert-manager must be installed in your cluster
  2. Create API key secret:
     kubectl create secret generic opusdns-secret \\
       --from-literal=api-key=opk_your_api_key \\
       -n cert-manager

Quick Start:
  tilt up
  tilt up -- --api_key=opk_your_key  # Auto-create secret

Manual Testing:
  - Click 'deploy-test-issuer' in Tilt UI to create test issuer
  - Click 'run-tests' to run conformance tests

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
""")
