# Template-manifest environments must not wait for a nonexistent HelmRelease

## Observed

Flux Kustomization `pr-20260826-test-app-full-114.generic` renders raw scanned Kubernetes manifests, but generated a health check for `HelmRelease/app`. No HelmRelease is included in that render set, so Flux remains `Ready=Unknown` for the full timeout even after applying the resources.

## Expected

Raw manifest/template renderers must either omit HelmRelease health checks or use health checks that match resources actually rendered. Helm chart renderers may retain the HelmRelease check.

## Acceptance criteria

1. Default the health check to disabled for raw template-manifest environments when no explicit check is configured.
2. Preserve an explicitly configured health-check override and HelmRelease behavior for Helm-based renderers.
3. Cover both cases in renderer and control-plane integration tests.
