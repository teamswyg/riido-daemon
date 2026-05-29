# DDD Context Map SSOT

> Riido task: RIID-4711 `[Daemon] Architecture SSOT docs migration`
>
> This document owns the public `riido-daemon` bounded-context map after the
> repository split. Shared C1/C2/C3 contracts are consumed from
> `riido-contracts`; SaaS C10 server behavior is owned by
> `riido-control-plane`; private deploy/evidence wiring is owned by
> `riido-infra`.

## Bounded Contexts

| ID | Context | Public daemon owner |
| --- | --- | --- |
| C1 | Task Lifecycle | `github.com/teamswyg/riido-contracts/task`; daemon mutates local rows through `internal/taskdb` only |
| C2 | IR Event Log | `github.com/teamswyg/riido-contracts/ir`; daemon-side append/redaction boundary is `internal/ir/ingest` |
| C3 | Provider Capability | `github.com/teamswyg/riido-contracts/provider/capability`; daemon provider adapters produce detected snapshots |
| C4 | Provider Runtime / Adapter | `internal/agentbridge`, `internal/agentbridge/session`, `internal/provider/{claude,codex,openclaw,cursor}` |
| C5 | Runtime Scheduling | `internal/scheduling`, `internal/agentbridge/runtimeactor`, `internal/agentbridge/supervisor`, `internal/agentbridge/controlplane/taskdbplane` |
| C6 | Workspace / Native Config | `internal/workdir` |
| C7 | Security / Policy | `internal/policy` |
| C8 | Validation | `internal/validation` |
| C9 | Locking / Lease Primitive | `internal/lock` plus task DB sidecar leases |
| C10 | SaaS Control Plane Adapter | daemon-side polling/reporting adapter in `internal/agentbridge/controlplane/saasplane`; server behavior remains in `riido-control-plane` |
| C11 | Distribution / Host Integration | `internal/hostintegration`, `internal/riidoapi`, `packaging/store`, `tools/storecontract` |

## Dependency Direction

```text
riido-contracts/task,ir,provider/capability
        |
        v
 C4 provider adapters -> C4 session/runtime actors -> C5 supervisor
        |                         |                       |
        v                         v                       v
 C2 ingest/redaction        C6 workdir              C10 adapter ports
        |                         |                       |
        v                         v                       v
 C1 local task DB       C7 policy decisions       local taskdb / SaaS polling

C11 host integration supplies local IPC, app data roots, consent, provider
provenance, review-demo surfaces, and store-channel policy inputs. C11 may
consult C7 but C7 must not call OS/store adapters.
```

The daemon imports contracts inward and adapts host/provider/server reality at
the edges. Runtime/domain packages must not import `cmd/riido`,
`riido-control-plane`, `riido-infra`, Terraform, AWS account data, or provider
CLI binaries.

## ACL Locations

| ACL | Input | Output |
| --- | --- | --- |
| Provider adapter ACL | Claude/Codex/OpenClaw/Cursor raw stdout or RPC events | `agentbridge.RawEvent`, `CanonicalEvent` drafts, provider-neutral result state |
| Event ingestor ACL | provider-neutral event drafts and unknown fields | validated/redacted `riido-contracts/ir` envelopes |
| SaaS polling ACL | assignment polling DTOs from `riido-contracts/assignment` | `agentbridge.TaskRequest` and reporter events |
| Task DB ACL | local `riido-task-db.v1` records | `TaskSourcePort` claims and guarded task mutations |
| mwsd bridge ACL | mwsd socket JSON contracts | project/task DB projection records |
| Host integration ACL | OS/store facts, external tool paths, consent/grants | provider-neutral routing status, local IPC endpoints, app/workdir roots |

## Split-Repo Ownership

`riido-daemon` must not redefine shared task/IR/provider capability facts.
When both daemon and control-plane need the same DTO/schema, promote it to
`riido-contracts` first. When a fact is deployment-only, keep it in
`riido-infra`. When a fact is server runtime behavior, keep it in
`riido-control-plane`.

Agent settings follow the same direction. `riido-contracts` owns the shared
meaning of agent profile fields and instruction limits. `riido-control-plane`
owns save/update API behavior. `riido-daemon` owns only the customer-PC runtime
consumption of an assigned instruction value and must not redefine thumbnail
presentation, one-line description presentation, RBAC/editability, API shape, or
server storage policy.

Figma `node-id=156-19307` menu placement is a client route affordance. The
daemon may power runtime status after a route is opened, but it does not own
menu labels, ordering, selected state, or route availability.

Figma `node-id=162-23090` runtime settings is a client composition over
control-plane device/runtime liveness and the local daemon lifecycle surface.
The daemon owns the current-device local facts exposed by `riido daemon status`,
`health`, `ready`, `metrics`, `stop`, `logs`, and `start`. It does not own the
agent hover popover, daemon stop modal copy, restart animation, remote-device
presentation, or SaaS `GET /v1/client/ai-agent/devices` projection.

## Change Procedure

Changing context ownership or dependency direction is a policy-breaking change.
The same PR must update this document, `docs/30-architecture/module-decomposition.md`,
and any package/workflow gate that enforces the boundary.
