# cert-manager-webhook-opusdns

[![CI](https://github.com/OpusDNS/cert-manager-webhook-opusdns/actions/workflows/ci.yaml/badge.svg)](https://github.com/OpusDNS/cert-manager-webhook-opusdns/actions/workflows/ci.yaml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/OpusDNS/cert-manager-webhook-opusdns)](go.mod)

A [cert-manager](https://cert-manager.io) webhook solver for [OpusDNS](https://opusdns.com) DNS API.

## Requirements

- [Kubernetes](https://kubernetes.io/) >= 1.22
- [Helm](https://helm.sh/) >= 3.0
- [cert-manager](https://cert-manager.io/) >= 1.16

## Installation

### Using Helm

```bash
helm repo add opusdns https://opusdns.github.io/cert-manager-webhook-opusdns
helm install cert-manager-webhook-opusdns opusdns/cert-manager-webhook-opusdns \
  --namespace cert-manager
```

### From Source

```bash
helm install cert-manager-webhook-opusdns deploy/cert-manager-webhook-opusdns \
  --namespace cert-manager
```

## Configuration

### 1. Create Secret

Create a Kubernetes secret with your OpusDNS API key:

```bash
kubectl create secret generic opusdns-secret \
  --namespace cert-manager \
  --from-literal=api-key=opk_YOUR_API_KEY
```

### 2. Create Issuer

Create a `ClusterIssuer` or `Issuer`:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - dns01:
          webhook:
            groupName: acme.opusdns.com
            solverName: opusdns
            config:
              apiKeySecretRef:
                name: opusdns-secret
                key: api-key
              # Optional: specify zone name explicitly
              # zoneName: example.com
              # Optional: custom TTL (default: 60)
              # ttl: 120
```

### 3. Create Certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-cert
  namespace: default
spec:
  secretName: example-cert-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
    - example.com
    - "*.example.com"
```

## Helm Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `groupName` | API group name for the webhook | `acme.opusdns.com` |
| `image.registry` | Image registry | `ghcr.io` |
| `image.repository` | Image repository | `opusdns/cert-manager-webhook-opusdns` |
| `image.tag` | Image tag | `appVersion` |
| `replicaCount` | Number of replicas | `1` |
| `secretName` | Allowed secret names | `[opusdns-secret]` |
| `resources` | Resource requests/limits | `{}` |
| `topologySpreadConstraints` | Topology spread constraints | `[]` |
| `podDisruptionBudget.enabled` | Enable PodDisruptionBudget | `false` |

## Webhook Configuration

| Field | Description | Required |
|-------|-------------|----------|
| `apiKeySecretRef.name` | Secret name containing API key | Yes |
| `apiKeySecretRef.key` | Key within the secret | No (default: `api-key`) |
| `zoneName` | DNS zone name (auto-detected if not set) | No |
| `ttl` | TTL for TXT records | No (default: `60`) |
| `apiEndpoint` | Custom API endpoint | No |

## Contributing

See [DEVELOPMENT.md](DEVELOPMENT.md) for development setup, testing, and release instructions.

## Support

- Documentation: https://developers.opusdns.com
- Issues: [GitHub Issues](https://github.com/OpusDNS/cert-manager-webhook-opusdns/issues)
- Email: support@opusdns.com
