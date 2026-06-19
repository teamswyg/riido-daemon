# Command Groups

[Back to CLI Surface SSOT](../cli-surface.md)

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
