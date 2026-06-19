# Core Packages

[Back to Public Migration Status](../public-migration-status.md)

| RIID | Public package | Migration status |
| --- | --- | --- |
| RIID-4652 | `internal/agentbridge/toolargs`, `toolpolicy` | bounded/redacted `ToolRef.Args`, C7 risk surface, `AutoApprover`, `ToolStartGate` |
| RIID-4653 | `internal/agentbridge/session` | one-run session actor, parser/driver/reducer stream, timeout, cancellation, telemetry, fail-closed gate |
| RIID-4654 | `internal/agentbridge/bridge`, `detectutil` | provider registry, detect/run entrypoint, PATH lookup, env override pin, version probe |
| RIID-4656 | `internal/agentbridge/runtimeactor` | mailbox actor, detect, capability reconciliation, bounded slots, submit/cancel/status/heartbeat |
| RIID-4657 | `internal/agentbridge/controlplane` | provider-neutral source/reporter port for registration, heartbeat, claim, cancel-watch, events |

Provider-native hook/RPC pre-start interrupt and SaaS/web approval handoff are
not owned by the early tool policy slice alone; later session/runtimeactor and
provider adapter slices execute those paths.

Still outside these core package moves: supervisor polling loop,
`controlplane/saasplane`, `controlplane/taskdbplane`, concrete provider adapters,
server HTTP/SSE transport, task DB, project, and MWSD adapters.
