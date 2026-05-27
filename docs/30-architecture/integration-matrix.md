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

## Assertions

| Provider | Integration assertion |
| --- | --- |
| Claude | simple prompt reaches `ResultCompleted` through stream JSON parsing |
| Codex | app-server JSON-RPC initialize/thread/turn flow reaches `ResultCompleted` |
| OpenClaw | JSON or NDJSON result flow reaches `ResultCompleted` with deterministic session id |
| Cursor | selected launch profile and stream JSON flow reach `ResultCompleted`; missing login probe skips |

## Change Procedure

When a provider adapter changes real CLI behavior, update the provider test and
this matrix in the same PR. New providers must add deterministic public tests
before adding optional real-CLI integration.
