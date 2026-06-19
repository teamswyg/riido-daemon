# RIID-4683: Taskdbplane Local Task DB Adapter

[Back to codex-auth-full-access](../codex-auth-full-access.md)

This slice moves the local task DB control-plane adapter.

Moved surfaces:

- `internal/taskdb`
- `internal/agentbridge/controlplane/taskdbplane`
- docs updates in provider-runtime, runtime-scheduling, locking, and daemon
  migration SSOT files
- focused public CI for guarded task DB mutation, command-id idempotent replay,
  runtime registry sidecars, lease sidecars, fencing tokens, expired lease
  handoff, and taskdbplane claim/report black-box scenarios

`internal/taskdb` owns the public daemon copy of `riido-task-db.v1` JSON schema
and guarded mutation rules. It persists transitions, deterministic validation
evidence, and command receipts without importing private `riido_daemon` packages
or workspace projection code.

`internal/agentbridge/controlplane/taskdbplane` adapts the local JSON DB into
`TaskSourcePort` and `TaskReporterPort`. It claims eligible `Queued` rows,
records guarded C1 transitions, stores runtime registry / lease sidecars next to
the task DB, and rejects progress/result reports without matching active lease
metadata.

This slice imports C1/C2/C3 domain types from `github.com/teamswyg/riido-contracts`
and does not reintroduce private `riido_daemon` internal packages. It does not
move project/mwsd sync, local API/socket, CLI commands, `controlplane/saasplane`,
server HTTP transport, packaging artifacts, private infra, secrets, or local
machine state.
