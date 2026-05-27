# Module Decomposition SSOT

> Riido task: RIID-4711 `[Daemon] Architecture SSOT docs migration`
>
> This document owns the public `riido-daemon` package layout, hexagonal import
> rules, and 12-factor daemon adapter boundary.

## Decisions

1. The Go module is `github.com/teamswyg/riido-daemon`.
2. The only deployable binary in this repository is `cmd/riido`.
3. SaaS server code stays in `riido-control-plane`; deploy/apply/evidence code
   stays in `riido-infra`; shared DTO/schema facts stay in `riido-contracts`.
4. Claude, Codex, OpenClaw, and Cursor CLIs are external attached resources.
   This repository never bundles, installs, or silently downloads them.
5. Domain packages depend inward on contracts and ports. Provider, local IPC,
   filesystem, process, SaaS HTTP, and host/store behavior enter through
   adapters.

## Package Map

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

## Import Rules

| Package group | May import | Must not import |
| --- | --- | --- |
| `internal/agentbridge` root | stdlib | provider packages, process implementations, local API, task DB, mwsd/project, SaaS HTTP adapters |
| `internal/agentbridge/session` | `internal/agentbridge`, `internal/process` | concrete provider packages, task DB, mwsd/project, local API |
| `internal/provider/<name>` | `internal/agentbridge`, provider-local helpers, allowed policy/workdir helpers | another provider package, local task DB/project/server internals |
| `internal/scheduling` | stdlib and contract capability types | provider/process/local API/project/server implementations |
| `internal/hostintegration` | stdlib and contract vocabulary | provider execution, task DB/project, local API, workdir, server/deploy code |
| `internal/policy` | stdlib and C11 value types | provider execution, OS adapters, task DB/project, server/deploy code |
| `internal/riidoapi` | local task/validation adapters and local transports | public TCP listener, SaaS server packages |
| `cmd/riido` | composition packages in this repository | server binary code, Terraform/AWS/deploy evidence, provider CLI binaries |

Production code must keep these rules. Tests may use package-local fakes, but
must not normalize cross-context imports as production design.

## Hexagonal Ports

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

## 12-Factor Boundary

Configuration is injected through `RIIDO_*` env vars or explicit local CLI
flags. Test-only gates use `AGENTBRIDGE_*`. Local daemon state is disposable;
durable facts live in task DB JSON, sidecar lease/registry files, mwsd
projection files, workdir metadata, or SaaS assignment events. `cmd/riido`
opens local IPC only and must not add a public HTTP listener.

## Change Procedure

When adding a package, env var, CLI flag, or adapter:

1. Update this document if the package/dependency map changes.
2. Update `docs/20-domain/context-map.md` if bounded-context ownership changes.
3. Update `docs/30-architecture/config-reference.md` for env/config changes.
4. Add or update a focused public workflow when the boundary can be checked in
   GitHub Actions without secrets or provider binaries.
