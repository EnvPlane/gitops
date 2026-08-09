# EnvPlane GitOps

GitOps rendering and repository operations for [EnvPlane](https://envplane.dev).
The package turns environment intent into reviewable manifests and provides the
boundary for repository-backed delivery.

## Responsibilities

- Render Flux CD and Kubernetes manifest sets.
- Generate environment-specific configuration.
- Support local and repository-backed GitOps writers.
- Prepare commit and pull-request proposals.

## Development

```bash
go test ./...
go vet ./...
```

## Related components

- [Control Plane](https://github.com/EnvPlane/control-plane)
- [Contracts](https://github.com/EnvPlane/contracts)
- [Deploy](https://github.com/EnvPlane/deploy)

## Security

Git credentials and signing keys are runtime configuration. Do not commit
tokens, private keys, or repository URLs containing credentials.

## Status

Private EnvPlane platform component under active development.
