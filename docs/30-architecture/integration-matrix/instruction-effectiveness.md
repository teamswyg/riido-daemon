# Agent Instruction Effectiveness Probe

[Back to provider integration matrix](../integration-matrix.md)

Provider instruction effectiveness is separate from deterministic prompt
placement. Public CI verifies the placement matrix and probe shape with
`go test ./internal/agentbridge`, without launching provider CLIs.

Real provider evidence is opt-in and must use the same harness:

1. Build the provider-specific probe with `BuildInstructionEffectivenessProbe`.
2. Send the generated prompt/system prompt through the provider adapter's normal
   integration path.
3. Validate the provider output with `ValidateInstructionEffectivenessOutput`.
4. Record missing executable or missing authentication as a skip only before the
   provider roundtrip starts. Once the provider accepts the run, a missing marker
   is an instruction-effectiveness failure.

| Provider | Probe marker | Expected instruction surface |
| --- | --- | --- |
| Claude | `RIIDO_INSTRUCTION_ACK:claude` | `system-prompt` |
| Codex | `RIIDO_INSTRUCTION_ACK:codex` | `prompt` prefix |
| OpenClaw | `RIIDO_INSTRUCTION_ACK:openclaw` | `system-prompt-inline` |
| Cursor | `RIIDO_INSTRUCTION_ACK:cursor` | `prompt` prefix |
