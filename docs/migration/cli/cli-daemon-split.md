# CLI / Daemon Split

[Back to Riido CLI Migration Plan](../cli.md)

The CLI is a thin adapter. Domain decisions stay in owning packages and SSOT
docs.

| CLI concern | Owner |
| --- | --- |
| Argument parsing and usage text | `cmd/riido` |
| Task FSM legality | `riido-contracts/task` through `internal/taskdb` guarded mutation |
| IR event schema | `riido-contracts/ir` |
| Provider process execution | daemon runtime packages |
| Local IPC transport | `internal/riidoapi` and `internal/hostintegration` |
| SaaS HTTP/SSE server | `riido-control-plane` |
| Deploy/apply behavior | `riido-infra` |
