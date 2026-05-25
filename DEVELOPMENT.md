# Development

## Prerequisites

- [Go](https://go.dev/) >= 1.26
- [Docker](https://www.docker.com/)
- [Helm](https://helm.sh/) >= 3.0
- [Tilt](https://tilt.dev) (optional, for local K8s development)

## Building

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Build multi-arch Docker image
make docker-buildx

# Lint (requires golangci-lint)
make lint

# Render Helm template
make helm-template
```

## Local Development with Tilt

For rapid iteration with a local Kubernetes cluster:

```bash
# Start Tilt (requires local K8s cluster with cert-manager installed)
tilt up

# Or with API key (auto-creates secret)
tilt up -- --api_key=opk_your_api_key
```

Copy `tilt-settings.yaml.example` to `tilt-settings.yaml` for custom configuration.

## Running Integration Tests

Integration tests run the cert-manager conformance suite against a real OpusDNS environment.

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `OPUSDNS_API_KEY` | OpusDNS API key | Yes |
| `OPUSDNS_API_ENDPOINT` | OpusDNS API endpoint URL | Yes |
| `TEST_ZONE_NAME` | DNS zone for tests (with trailing dot) | Yes |

### Running Locally

```bash
# Set environment
export OPUSDNS_API_KEY=opk_your_api_key
export OPUSDNS_API_ENDPOINT=https://api.preview1.opusdns.dev
export TEST_ZONE_NAME=cert-manager-test.opusdns.dev.

# Generate testdata from environment
./scripts/setup-testdata.sh

# Run tests (downloads kubebuilder binaries automatically)
make test
```

The test zone must exist in the target environment before running tests.

### CI

Integration tests run automatically on push to `main` via GitHub Actions using the `preview1` environment. The workflow uses the same secrets (`OPUSDNS_API_KEY`, `OPUSDNS_API_ENDPOINT`) configured in the repository settings.

## Release Process

1. Ensure all tests pass on `main`
2. Create and push a version tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. The release workflow will automatically:
   - Build multi-arch Docker images (linux/amd64, linux/arm64)
   - Push to `ghcr.io/opusdns/cert-manager-webhook-opusdns`
   - Create a GitHub Release with auto-generated notes

### Versioning

- Docker image tags: `v1.0.0`, `latest`
- Helm chart version: updated in `deploy/cert-manager-webhook-opusdns/Chart.yaml`
- User-Agent version: updated in `main.go` (`Version` constant)

## Project Structure

```
├── main.go                     # Webhook solver implementation
├── main_test.go                # Conformance tests
├── Dockerfile                  # Multi-stage Docker build
├── Makefile                    # Build/test/lint targets
├── deploy/
│   └── cert-manager-webhook-opusdns/   # Helm chart
├── scripts/
│   ├── fetch-test-binaries.sh  # Download kubebuilder tools
│   └── setup-testdata.sh       # Generate test config from env
├── testdata/
│   └── opusdns/                # Test fixtures
└── .github/workflows/          # CI/CD pipelines
```
