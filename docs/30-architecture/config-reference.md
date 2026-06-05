# Daemon Config Reference

> Riido task: RIID-4711 `[Daemon] Architecture SSOT docs migration`
>
> This document owns the public `riido-daemon` env/flag catalog. SaaS server
> `RIIDO_AI_SERVER_*` variables are owned by `riido-control-plane` and private
> deploy-time evidence paths are owned by `riido-infra`.

## Naming Rules

- User/operator runtime config uses `RIIDO_*`.
- Test-only integration gates use `AGENTBRIDGE_*`.
- Provider executable overrides use `RIIDO_<PROVIDER>_PATH`.
- New env vars must be documented here in the same PR that reads them.

## Provider Executable Overrides

| Variable | Consumer | Default | Fail-closed rule |
| --- | --- | --- | --- |
| `RIIDO_CLAUDE_PATH` | `internal/provider/claude.Detect` | `exec.LookPath("claude")` | explicit missing/bad path does not fall back to PATH |
| `RIIDO_CODEX_PATH` | `internal/provider/codex.Detect` | `exec.LookPath("codex")` | same |
| `RIIDO_OPENCLAW_PATH` | `internal/provider/openclaw.Detect` | PATH candidate probe for `openclaw` | same |
| `RIIDO_CURSOR_PATH` | `internal/provider/cursor.Detect` | `exec.LookPath("cursor-agent")` | same |

The daemon reports unavailable providers instead of executing a different
binary than the operator selected.

The executable that passes Detect is also the executable handed to provider
process start. Runtime start MUST NOT re-resolve a same-name binary from `PATH`
after capability detection has selected a concrete path.

OpenClaw is the only adapter that probes later same-name PATH candidates when
`RIIDO_OPENCLAW_PATH` is unset. The first candidate that passes the
calendar-version gate is selected. An explicit `RIIDO_OPENCLAW_PATH` remains a
pin and never falls back to PATH, even when a supported OpenClaw binary exists
later on PATH.

### Executable Search Path (GUI/launchd-spawned daemons)

When no `RIIDO_<PROVIDER>_PATH` override is set, `internal/agentbridge/detectutil`
resolves the provider executable across an augmented search path, not just the
process `PATH`. A daemon launched by the Riido Desktop app, launchd, or a
service inherits a minimal `PATH` (on macOS launchd typically only
`/usr/bin:/bin:/usr/sbin:/sbin`) that omits the Homebrew and per-user
directories where `claude`, `codex`, `cursor-agent`, and `openclaw` are
installed. Resolving from the process `PATH` alone would report installed
providers as `detection_state=missing`.

The augmented search order is:

1. the process `PATH` (an operator's explicit `PATH` still wins);
2. the user's login-shell `PATH`, read once per process via
   `$SHELL -lc` and cached (skipped on Windows, when `$SHELL` is unset, or on
   timeout — bounded by a short deadline so a slow profile cannot hang Detect);
3. well-known install directories: `/opt/homebrew/{bin,sbin}`,
   `/usr/local/{bin,sbin}`, the standard system dirs, and per-user locations
   (`~/.local/bin`, `~/.npm-global/bin`, `~/.cargo/bin`, `~/.bun/bin`,
   `~/.deno/bin`, `~/go/bin`, `~/.volta/bin`, `~/.asdf/shims`, `~/.cursor/bin`,
   `~/.claude/bin`) plus resolved nvm/fnm/asdf Node version bins.

This augmentation only widens where an unset-override lookup may find a binary.
The fail-closed pin semantics of `RIIDO_<PROVIDER>_PATH` are unchanged: an
explicit override still resolves to exactly that file and never falls back to
the search path.

## Test Integration Gate

| Variable | Consumer | Default |
| --- | --- | --- |
| `AGENTBRIDGE_INTEGRATION=1` | provider `TestIntegration` tests | unset tests skip |

Real provider CLI integration is opt-in. Public CI runs deterministic adapter
tests by default and does not require provider CLIs or vendor credentials.

## Daemon Identity And Runtime Config

| Variable | Consumer | Default | Meaning |
| --- | --- | --- | --- |
| `RIIDO_DAEMON_ID` | `cmd/riido/daemon_config.go` | `agentd-local` | stable local daemon/runtime slot id |
| `RIIDO_DAEMON_VERSION` | daemon config + supervisor event stamps | `riido-agentd v0.0.0` | daemon binary version label |
| `RIIDO_DAEMON_PROFILE` | daemon status | `local` | user-facing daemon profile |
| `RIIDO_SERVER_URL` | daemon status | empty | display-only server URL unless SaaS source is configured |
| `RIIDO_DEVICE_ID` | `saasplane` | empty | device principal id issued by `riido-control-plane`; required with `RIIDO_DEVICE_SECRET` when SaaS auth uses device credentials |
| `RIIDO_DEVICE_SECRET` | `saasplane` | empty | device principal secret issued by Desktop device enrollment; sent only as `X-Riido-Device-Secret` to the SaaS API |
| `RIIDO_DEVICE_NAME` | daemon status | hostname or `localhost` | device display name |
| `RIIDO_RUNTIME_OWNER` | daemon status | `$USER` or `local` | runtime owner display name |
| `RIIDO_RUNTIME_AGENTS` | daemon status | empty list | comma-separated attached agent display names |
| `RIIDO_WORKSPACE_COUNT` | daemon status | `0` | non-negative workspace count display value |
| `RIIDO_WORKDIR_ROOT` | supervisor + workdir FS adapter | C11 dev-local app data workdir root | isolated run workdir root |
| `RIIDO_WORKDIR_RETENTION_SECONDS` | daemon config + workdir cleanup | `0` | opt-in archived workdir TTL cleanup; zero disables cleanup and there is no default size/task-count cleanup |
| `RIIDO_WORKDIR_CLEANUP_INTERVAL_SECONDS` | daemon config + workdir cleanup | `3600` when retention is enabled | cleanup loop interval; invalid without retention |
| `RIIDO_POLICY_BUNDLE_PATH` | daemon config + policy loader | empty | optional `riido-policy-bundle.v1` file |
| `RIIDO_POLICY_BUNDLE_VERSION` | daemon config/runtime/workdir | `policy-bundle.local.v0` | active policy version; must match file version when path is set |

## Task Source Selection

Exactly one production task source may be selected.

| Variable | Consumer | Default | Meaning |
| --- | --- | --- | --- |
| `RIIDO_TASK_QUEUE_DIR` | file queue source | empty | directory of provider-neutral task JSON files |
| `RIIDO_TASK_REPORT_DIR` | file reporter | `RIIDO_TASK_QUEUE_DIR/reports` when queue is set | JSONL report output directory |
| `RIIDO_TASK_DB_SOURCE_PATH` | `taskdbplane` | empty | local `riido-task-db.v1` production source |
| `RIIDO_SAAS_URL` | `saasplane` | empty | SaaS assignment polling endpoint |
| `RIIDO_DAEMON_POLL_INTERVAL_SECONDS` | supervisor | `1` | active/fast claim polling interval |
| `RIIDO_DAEMON_IDLE_POLL_INTERVAL_SECONDS` | supervisor | `5` | idle polling interval; must be >= active interval |
| `RIIDO_DAEMON_HEARTBEAT_INTERVAL_SECONDS` | supervisor | `5` | runtime heartbeat interval |

Queue, task DB, and SaaS source variables are mutually exclusive where their
adapters would otherwise compete for task ownership.

When `RIIDO_SAAS_URL` is selected with `RIIDO_DEVICE_ID` and
`RIIDO_DEVICE_SECRET`, `saasplane` uses DevicePrincipal authentication. The
daemon reports provider runtime snapshots to `/v1/daemon/runtime-snapshot`,
polls `/v1/daemon/agent-bindings`, and only then polls the agent-specific
assignment endpoint. Legacy `RIIDO_SAAS_AGENTS` / `RIIDO_SAAS_TOKEN` values are
not read by the daemon settings model.
The snapshot must preserve the local provider availability verdict: an explicitly
false `provider.<name>.available` capability is projected as `offline` /
`missing`, not normalized to `online` merely because the provider binary was
seen.
This daemon-side flow does not accept `team_id`, `teamId`, OpenAPI task-context
paths, Open API keys, or `X-Workspace-Api-Key` as identity, binding, polling, or
smoke-test inputs. Those values are outside the generated AI Agent assignment
and DevicePrincipal SSOT owned by `riido-contracts`.

When a task is claimed, `saasplane` turns the control-plane assignment into a
provider-neutral `TaskRequest`. The assignment-created `agent_instruction`
snapshot and the Riido telemetry contract are placed by provider capability:
Claude and OpenClaw use the system prompt surface, while Codex and Cursor use a
prompt prefix because their current daemon surface does not rely on a separate
system prompt channel. The chosen placements are recorded in
`TaskRequest.Metadata["riido_agent_instruction"]` and
`TaskRequest.Metadata["riido_telemetry_contract"]` so tests can detect provider
placement drift. The provider-specific placement matrix and the opt-in real
provider effectiveness probe are owned by
[`provider-runtime.md`](../20-domain/provider-runtime.md#761-agent-instruction-placement-and-effectiveness-probe)
and [`integration-matrix.md`](integration-matrix.md#agent-instruction-effectiveness-probe);
this config reference only names the runtime source variable and emitted
metadata keys.

Desktop-launched daemon authentication is device-principal based. Desktop first
uses the logged-in webview session to enroll the device with `riido-control-plane`,
then launches the daemon with `RIIDO_DEVICE_ID` and `RIIDO_DEVICE_SECRET`.
`saasplane` sends those values as `X-Riido-Device-ID` and
`X-Riido-Device-Secret` on poll/heartbeat/event requests. This keeps daemon
polling alive after the webview is closed and avoids treating a browser JWT as a
daemon credential. Team IDs and Open API keys are not a substitute bridge
between the webview and daemon; the bridge is the enrolled device credential
only.

## Local Daemon Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `--socket` | C11 dev-local local IPC endpoint | local API socket or named pipe path |
| `--transport` | `unix-socket` | `unix-socket` or `windows-named-pipe` |
| `--pid-file` | unset | optional background PID file |
| `--log-file` | unset | optional structured log file |
| `--foreground` | false | run daemon in the current process |
| `--lock-file` | `$HOME/.riido/.lock` | local singleton advisory lock |
| `--timeout-seconds` | `5` for stop | graceful stop wait before kill |

`cmd/riido` remains local-only. Health, ready, status, and metrics are exposed
through local IPC subcommands, not public HTTP.

## Change Procedure

When an env var or daemon flag is added, update this document, add a parser
test, and keep failure modes explicit. Provider-specific env reads must go
through testable env helpers rather than direct unscoped `os.Getenv` calls in
adapter internals.
