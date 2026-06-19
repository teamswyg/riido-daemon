# RIID-4651: Agentbridge Root Provider Runtime Domain

[Back to distribution-host-integration](../distribution-host-integration.md)

This slice moves the provider-neutral C4 Provider Runtime / Adapter root domain.

Moved surfaces:

- `internal/agentbridge`
- `docs/20-domain/provider-runtime.md`
- focused public CI for reducer, telemetry, blocked-arg, and semantic-activity
  gates

The package is stdlib-only and intentionally does not import concrete provider
packages, task/project persistence, process execution implementations, local API
packages, or filesystem/network adapters.

This slice does not move `internal/agentbridge/session`, `runtimeactor`,
`supervisor`, `bridge`, `controlplane`, `detectutil`, concrete provider adapters,
`ToolRef.Args` flattening/toolpolicy execution, task DB/project/mwsd local API
packages, packaging artifacts, private infra, secrets, or local machine state.
