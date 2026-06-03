# Provider Integration Matrix

> Riido task: RIID-4711 `[Daemon] Architecture SSOT docs migration`
>
> This document owns how public `riido-daemon` verifies real provider CLIs.
> Provider CLIs are external attached resources and are never bundled.

## Gate Policy

Each provider `TestIntegration` is optional until all gates pass:

1. `AGENTBRIDGE_INTEGRATION=1` must be set, otherwise the test skips.
2. The provider executable must be discoverable or explicitly configured with
   `RIIDO_<PROVIDER>_PATH`, otherwise the test skips.
3. The adapter `Detect` result must be available, otherwise the test skips with
   the detect reason.

After all gates pass, a failed prompt roundtrip is a real integration failure.
Provider authentication probes may classify missing login/API-key state as
operator environment skip only when the provider exposes a deterministic probe.

## Provider Matrix

| Provider | Executable | Public deterministic CI | Real CLI integration |
| --- | --- | --- | --- |
| Claude Code | `claude` | command/parser/translator/golden tests | `AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/claude -run TestIntegration -count=1` |
| Codex | `codex` app-server stdio | command/parser/translator/RPC/golden tests | `AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/codex -run TestIntegration -count=1` |
| OpenClaw | `openclaw` | command/parser/version/golden tests | `AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/openclaw -run TestIntegration -count=1` |
| Cursor Agent | `cursor-agent` | command/parser/profile/golden tests | `AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/cursor -run TestIntegration -count=1` |

Public pull-request CI runs deterministic tests and keeps real provider
integration opt-in. A future scheduled/manual workflow may install provider
CLIs with `continue-on-error` for install steps only; the integration test step
itself must fail when the gates pass but behavior regresses.

Operators can run the current local provider matrix with:

```bash
./scripts/integration-smoke.sh
```

The script probes `claude`, `codex`, `openclaw`, and `cursor-agent`, honors
`RIIDO_CLAUDE_PATH`, `RIIDO_CODEX_PATH`, `RIIDO_OPENCLAW_PATH`, and
`RIIDO_CURSOR_PATH`, then runs `TestIntegration` only for providers that are
present. Missing executables remain an operator-environment skip, not a PASS.
Once a provider is detected and available, a failing roundtrip is a real
integration failure.

## Assertions

| Provider | Integration assertion |
| --- | --- |
| Claude | stream JSON flow reaches `ResultCompleted`, and the run writes the expected file artifact inside the daemon-selected workdir |
| Codex | app-server JSON-RPC initialize/thread/turn flow reaches `ResultCompleted` |
| OpenClaw | JSON or NDJSON result flow reaches `ResultCompleted` with deterministic session id, using the executable path that passed Detect, and the run writes the expected file artifact inside the daemon-selected workdir |
| Cursor | selected launch profile adds daemon-workdir `--trust` without `--yolo`, stream JSON flow reaches `ResultCompleted`, and the run writes the expected file artifact inside the daemon-selected workdir; missing login probe skips |

## Agent Instruction Effectiveness Probe

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

## Change Procedure

When a provider adapter changes real CLI behavior, update the provider test and
this matrix in the same PR. New providers must add deterministic public tests,
an instruction placement strategy, and an effectiveness probe marker before
adding optional real-CLI integration.
