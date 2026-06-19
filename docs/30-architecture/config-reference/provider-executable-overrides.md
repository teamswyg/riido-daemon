# Provider Executable Overrides

[Back to Daemon Config Reference](../config-reference.md)

| Variable | Consumer | Default | Fail-closed rule |
| --- | --- | --- | --- |
| `RIIDO_CLAUDE_PATH` | `internal/provider/claude.Detect` | `exec.LookPath("claude")` | explicit missing/bad path does not fall back to PATH |
| `RIIDO_CODEX_PATH` | `internal/provider/codex.Detect` | `exec.LookPath("codex")` | same |
| `RIIDO_OPENCLAW_PATH` | `internal/provider/openclaw.Detect` | PATH candidate probe for `openclaw` | same |
| `RIIDO_CURSOR_PATH` | `internal/provider/cursor.Detect` | `exec.LookPath("cursor-agent")` | same |

The daemon reports unavailable providers instead of executing a different binary
than the operator selected. Runtime start must reuse the executable selected by
Detect and must not re-resolve a same-name binary from `PATH`.

OpenClaw may probe later same-name PATH candidates only when
`RIIDO_OPENCLAW_PATH` is unset. An explicit OpenClaw override is a pin and never
falls back to PATH.
