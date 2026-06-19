# Provider Matrix

[Back to provider integration matrix](../integration-matrix.md)

| Provider | Executable | Public deterministic CI | Real CLI integration | Worktree routing status |
| --- | --- | --- | --- | --- |
| Claude Code | `claude` | command/parser/translator/golden tests | `AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/claude -run TestIntegration -count=1` | `supports_worktree=true` |
| Codex | `codex --sandbox danger-full-access app-server --listen stdio://` stdio | command/parser/translator/RPC/golden tests | `AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/codex -run TestIntegration -count=1` | `supports_worktree=true` |
| OpenClaw | `openclaw` | command/parser/version/golden tests plus C5 worktree ineligibility gate | `AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/openclaw -run TestIntegration -count=1` | `supports_worktree=false`; `required_surfaces=[worktree]` must fail with `MISSING_REQUIRED_SURFACE:worktree` |
| Cursor Agent | `cursor-agent` | command/parser/profile/golden tests | `AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/cursor -run TestIntegration -count=1` | `supports_worktree=true` |

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
