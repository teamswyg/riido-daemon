# CLI Surface SSOT

> Riido task: RIID-4714 `[Cli] Architecture SSOT docs migration`

This file owns the public `cmd/riido` command boundary. `cmd/riido` is a
local-only adapter shell for the customer-PC daemon and local task tooling.

## Role

The CLI may:

- parse args and print usage
- call local daemon packages in this repository
- read/write local JSON state through guarded adapters
- open local IPC transports only
- emit JSON for shell/operator automation

The CLI must not:

- start a public network listener
- bundle or install provider CLIs
- run infrastructure deploy/apply workflows
- own SaaS server behavior
- bypass task mutation guards or policy gates
- redefine contract facts owned by `riido-contracts`

## Command Groups

| Command group | Backing owner | Boundary |
| --- | --- | --- |
| `riido mwsd ...` | `internal/mwsdbridge`, `internal/project`, `internal/taskdb` | reads mwsd snapshots and promotes public workspace/task projections |
| `riido task ...` | `internal/taskdb`, `internal/validation` | local guarded task mutation/evidence/validation over JSON task DB |
| `riido serve` | `internal/riidoapi` | local IPC server over Unix socket or Windows named pipe |
| `riido api ...` | `internal/riidoapi` | client for the local IPC API |
| `riido bridge ...` | provider adapter packages | public provider list/detect smoke surface without executing provider tasks |
| `riido daemon ...` | `internal/agentbridge/runtimeactor`, `supervisor`, and control-plane adapters | local runtime lifecycle, health, ready, metrics, stop, and logs |

`printUsage()` in `cmd/riido/main.go` is the executable usage matrix. This
document describes command ownership; it does not replace usage text.

## Runtime Settings Mapping

Figma `node-id=162-23090` shows a runtime settings page with current-device
daemon status, daemon detail fields (`실행 시간`, PID, daemon ID, profile, and
device name), daemon stop, restart-in-progress UI, attached agents, and other
devices. This CLI owns only the current-device local lifecycle facts that a
desktop helper can read or invoke through local IPC/CLI:

- `riido daemon status|health|ready|metrics` expose current daemon status,
  readiness, PID, uptime, profile, device name, and runtime snapshots
- `riido daemon stop` performs cooperative local shutdown with PID fallback
- `riido daemon start` starts the local daemon process

There is no separate `riido daemon restart` command today. A desktop helper may
compose restart from local stop/start behavior and own the spinner/animation.
Remote device rows and the SaaS `GET /v1/client/ai-agent/devices` projection are
owned by `riido-control-plane` and `riido-contracts`; this CLI must not add a
public network listener or SaaS endpoint for the runtime settings screen.
Attached-agent avatar/profile rendering and hover details remain client
presentation over control-plane records.

The broader Figma AI Agent daemon boundary is projected in
[`figma-ai-agent-daemon-boundary.md`](figma-ai-agent-daemon-boundary.md).
That projection does not add CLI commands by itself; this CLI changes only when
the daemon boundary becomes a local lifecycle, IPC, provider detection, or
provider execution surface.

## Local IPC Rule

`riido serve`, `riido api`, and `riido daemon` may use:

- Unix socket
- Windows named pipe

They must not add TCP/HTTP listeners to the local CLI binary. SaaS HTTP routes
belong to the public control-plane repository.

## Guarded Mutation Rule

`riido task transition`, `riido task evidence`, `riido task validate`, and their
`riido api ...` equivalents must go through the same guarded mutation path used
by the local API. Approval IDs, command IDs, idempotent receipts, replay
mismatch checks, and deterministic validation evidence remain adapter-invariant.

## Validation

Required black-box checks for CLI docs/adapter changes:

```bash
go test ./...
go build -o /tmp/riido ./cmd/riido
go run ./cmd/riido --help
go run ./cmd/riido bridge providers
```

Commands that require a running daemon, mwsd socket, provider CLI, or local app
state should have explicit skip conditions in tests.
