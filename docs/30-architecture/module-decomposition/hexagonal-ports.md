# Hexagonal Ports

[Back to Module Decomposition SSOT](../module-decomposition.md)

| Port | Package | Adapters |
| --- | --- | --- |
| Provider run | `agentbridge.Adapter` | Claude/Codex/OpenClaw/Cursor adapters |
| Process | `process.Process` | `processexec`, `FakeProcess` |
| Task source | `controlplane.TaskSourcePort` | memory, file queue, task DB, SaaS polling |
| Task reporter | `controlplane.TaskReporterPort` | memory, file, task DB, SaaS events |
| Workdir FS | `workdir.FSAdapter` | local filesystem |
| Validation runner | `validation.RunCommand` | daemon-measured `/bin/sh -lc` |
| Local API transport | `riidoapi` transport | Unix socket, Windows named pipe |
| Host integration | `hostintegration` pure models | future GUI/helper OS adapters |

Adapters translate into provider-neutral types before crossing inward. Raw
provider payloads may be retained only behind event/adapter audit contracts.
