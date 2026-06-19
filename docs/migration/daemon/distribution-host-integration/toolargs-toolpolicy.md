# RIID-4652: Toolargs / Toolpolicy Migration

[Back to distribution-host-integration](../distribution-host-integration.md)

This slice moves the provider-neutral C4/C7 tool-use bridge.

Moved surfaces:

- `internal/agentbridge/toolargs`
- `internal/agentbridge/toolpolicy`
- docs updates in provider-runtime, security, and security-redaction SSOT files
- focused public CI for ToolRef.Args redaction, risk-surface classification,
  AutoApprover, and ToolStartGate gates

`toolargs` turns provider raw tool input into a bounded string map and redacts
sensitive keys or values with the `ToolRef.Args` marker. `toolpolicy` maps
`agentbridge.ToolRef` into C7 `ToolUseSurface` decisions and only auto-approves
when the active policy bundle explicitly allows the classified surface.

This slice does not move concrete provider adapter parser/wiring,
session/runtimeactor/supervisor execution, provider-native approval RPC/hook
implementation, ToolCallStarted fail-close wiring, task DB/project/mwsd local API
packages, packaging artifacts, private infra, secrets, or local machine state.
