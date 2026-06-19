# RIID-4684 — riidoapi Local API Adapter Migration

[Back to riidoapi local API](../riidoapi-local-api.md)

This slice moves the daemon local-only API adapter:

- `internal/riidoapi`
- docs updates in distribution-host-integration, CLI migration, and daemon
  migration SSOT files
- focused public CI for local API status/tasks/transition/evidence/validate,
  review-demo mode, Unix socket transport, Windows named pipe path behavior, and
  no-public-TCP listener boundary checks

`internal/riidoapi` owns the local JSON envelope and the local transport
adapters used by GUI/Zed/CLI surfaces. The handler exposes `status`, `tasks`,
`transition`, `evidence`, `validate`, and `review-demo` over local IPC only.
Task mutations use public `internal/taskdb` guarded transition/evidence
receipts. Validation uses public `internal/validation` and rejects missing
`approval_id` before running the command.

This slice imports C1/C2 domain types from `github.com/teamswyg/riido-contracts`
and does not reintroduce private `riido_daemon` internal packages. It does not
move `cmd/riido` CLI commands, mwsdbridge/project projection sync,
`controlplane/saasplane`, server HTTP transport, packaging artifacts, private
infra, secrets, or local machine state.
