# Module ownership

This workspace treats `gitops` as the canonical owner of GitOps rendering and
writing. New consumers must import `github.com/envpilot/gitops/render` or
`github.com/envpilot/gitops/writer`; they must not copy `internal/gitops`.

`control-plane` owns persistence, API orchestration, and lifecycle services.
`webhook` owns SCM provider parsing. `contracts` owns transport and domain
contracts. `bootstrap` is a public library of manifest and cleanup primitives;
it does not ship a standalone service binary.
