# EnvPilot GitOps

GitOps rendering and repository operations for EnvPilot.

## Scope

- Flux manifest rendering.
- Environment manifest set generation.
- Local and repository-backed GitOps writers.
- Git commit and PR proposal helpers.
- Environment orchestration boundary around rendered GitOps manifests.

## Source Origin

This repository was split from:

- `internal/gitops`
- `internal/orchestrator`
- `apps/gitops-committer`

## Follow-up

Turn the current library boundary into a dedicated commit/render worker if Git push and PR creation need to run outside the control plane.
