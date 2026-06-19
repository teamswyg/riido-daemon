# Agent Instruction Placement And Effectiveness Probe

[Back to adapter-acl.md](../adapter-acl.md)

C4 Provider Runtime owns only how an assignment-created `Assignment.agent_instruction` snapshot is delivered to each provider surface. It does not own the client-authored instruction text, the 1000-character limit, or the assignment-time snapshot decision.

Current placement matrix:

| Provider | Instruction placement | Telemetry placement | Deterministic gate |
| --- | --- | --- | --- |
| Claude Code | `system-prompt` | `system-prompt` | `go test ./internal/agentbridge` |
| OpenClaw | `system-prompt-inline` | `system-prompt-inline` | `go test ./internal/agentbridge` |
| Codex | `prompt` prefix | `prompt` prefix | `go test ./internal/agentbridge` |
| Cursor Agent | `prompt` prefix | `prompt` prefix | `go test ./internal/agentbridge` |

The matrix is implemented by `RuntimeInstructionStrategies()` and consumed by `ApplyRuntimeInstructionContract`. Public CI verifies deterministic placement, metadata, idempotent section composition, and provider-neutral effectiveness probe shape without executing external provider CLIs.

Provider-specific effectiveness means the real provider obeys the delivered instruction after it is placed on that provider's chosen surface. It is verified by `BuildInstructionEffectivenessProbe` and `ValidateInstructionEffectivenessOutput`.

The probe asks the provider to echo a provider-specific marker such as `RIIDO_INSTRUCTION_ACK:codex`, and the validator accepts only outputs containing that marker. Real provider execution remains an opt-in integration/evidence gate because provider CLIs, credentials, model selection, latency, and vendor behavior are external attached resources.
