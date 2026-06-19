# Package Map

[Back to Module Decomposition SSOT](../module-decomposition.md)

| Package | Role |
| --- | --- |
| `cmd/riido` | CLI/local daemon adapter. Parses flags/env and composes local-only surfaces. |
| `internal/agentbridge` | Provider-neutral C4 run/request/event/result domain. |
| `internal/agentbridge/session` | Per-run session actor over the process port and protocol driver. |
| `internal/agentbridge/runtimeactor` | One runtime mailbox/slot and capability reconciliation. |
| `internal/agentbridge/supervisor` | Daemon control loop, task claim/dispatch, workdir preparation, event ingest delegation. |
| `internal/agentbridge/controlplane` | Task source/reporter ports and memory/file adapters. |
| `internal/agentbridge/controlplane/taskdbplane` | Local task DB source/reporter adapter. |
| `internal/agentbridge/controlplane/saasplane` | SaaS polling/reporting adapter over assignment contracts. |
| `internal/provider/{claude,codex,openclaw,cursor}` | Concrete provider adapter ACLs. |
| `internal/process` / `internal/process/processexec` | Process port, fake process, and `os/exec` adapter. |
| `internal/workdir` | Isolated run workdir and native config materialization. |
| `internal/scheduling` | Pure runtime eligibility/selection rules. |
| `internal/policy` | C7 security/policy decisions. |
| `internal/hostintegration` | C11 distribution/store/host pure models. |
| `internal/riidoapi` | Local IPC JSON API over Unix socket or Windows named pipe. |
| `internal/taskdb` | Public daemon copy of `riido-task-db.v1` guarded mutation adapter. |
| `internal/mwsdbridge` / `internal/project` | mwsd ACL and workspace/task projection. |
| `internal/ir/ingest` | Daemon-side event append/redaction boundary. |
| `tools/storecontract` | Store distribution contract verifier. |
| `tools/riidogen` | Local generator for executable Riido contracts. |
