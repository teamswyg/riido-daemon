# Riido Daemon Migration Plan

> Riido task: RIID-4636 `[Daemon] 기존 riido_daemon daemon 마이그레이션 계획/문서화`

This document defines how the daemon/runtime part of the former private
`riido_daemon` repository moves into the public `riido-daemon` repository.

## Goal

`riido-daemon` owns the customer-PC daemon runtime, local host integration, and
provider execution boundary. It must stay public, store-reviewable, and
stdlib-only unless a later ADR explicitly changes that rule.

## Source In The Private Repository

The first migration source is the current private `riido_daemon` main branch.
The daemon slice is the code and documentation that implement these bounded
contexts:

| Context | Source paths |
| --- | --- |
| C3 Provider Capability | `internal/provider/capability` |
| C4 Provider Runtime / Adapter | `internal/agentbridge`, `internal/provider/{claude,codex,openclaw,cursor}`, `pkg/agentbridge`, `internal/process` |
| C5 Runtime Scheduling | `internal/scheduling`, `internal/agentbridge/runtimeactor`, `internal/agentbridge/supervisor` |
| C6 Workspace / Native Config | `internal/workdir` |
| C7 Security / Policy | `internal/policy` |
| C8 Validation | `internal/validation` |
| C9 Local Locking | `internal/lock` |
| C11 Distribution / Host Integration | `internal/hostintegration`, local transport pieces in `internal/riidoapi` |

The documentation source is:

- `docs/20-domain/provider-runtime.md`
- `docs/20-domain/provider-capability.md`
- `docs/20-domain/security.md`
- `docs/20-domain/security-redaction.md`
- `docs/20-domain/distribution-host-integration.md`
- `docs/30-architecture/module-decomposition.md`
- `docs/30-architecture/integration-matrix.md`
- `docs/30-architecture/config-reference.md`
- daemon-related roadmap/audit files under `docs/50-roadmap/`

## Target Boundary

Move into `riido-daemon`:

- provider-neutral runtime actors and session actors
- provider adapter ACLs for Claude, Codex, OpenClaw, and Cursor
- process spawning ports and fakes
- local-only daemon control surfaces
- host integration models for store-safe local execution
- daemon-side validation and black-box tests
- daemon SSOT docs and daemon-specific ADRs

Do not move into `riido-daemon`:

- `cmd/riido_ai_server` or `internal/riidoaiserver`
- Terraform, AWS, ECS, ECR, WAF, ACM, Route53, or release evidence workflows
- `.riido-local`, state files, credentials, account IDs, or deploy artifacts
- shared contract code that must be consumed by both daemon and control-plane
- bundled Claude/Codex/OpenClaw/Cursor CLI binaries

## Migration Order

1. Port SSOT docs first.
   Keep domain decisions in docs before moving code that executes them.

2. Move provider-neutral primitives.
   Start with `internal/agentbridge` root types, reducer, command/result/event
   contracts, and tests that do not import concrete providers.

3. Move process/workdir/policy/validation support packages.
   Keep adapters behind ports and preserve stdlib-only verification.

4. Move provider adapters one at a time.
   Each provider migration must include parser/golden tests and detect command
   tests. Real CLI integration tests remain opt-in and skipped unless the CLI is
   installed.

5. Move daemon runtime actors and local host integration.
   Supervisor/runtime/session actors should remain mailbox-owned. Do not add
   shared mutable state as a migration shortcut.

6. Rebuild daemon workflows in the public repo.
   Public CI should run unit, domain, generated-drift, dependency, and
   black-box daemon checks. Private CI should not duplicate those expensive
   checks.

## Current Migration Slices

### RIID-4643 — contracts import gate

`riido-daemon` consumes `github.com/teamswyg/riido-contracts v0.1.0` and keeps
CI limited to Riido-owned Go module dependencies. This is a compatibility gate
only; it does not move runtime packages.

### RIID-4645 — local process / validation / lock core

This slice moves provider-neutral local daemon primitives that have no external
dependencies:

- `internal/process` and `internal/process/processexec`
- `internal/validation`
- `internal/lock`
- `internal/logging`
- `internal/jsontest`
- C8 validation and C9 locking SSOT docs under `docs/20-domain/`

This slice does not move provider adapters, runtime/session/supervisor actors,
task DB/project/mwsd/local API packages, CLI commands, private infra, secrets,
or local machine state.

### RIID-4646 — runtime scheduling domain

This slice moves the pure C5 scheduling domain:

- `internal/scheduling`
- `docs/20-domain/runtime-scheduling.md`

The package imports provider capability types from
`github.com/teamswyg/riido-contracts/provider/capability v0.1.0`.

This slice does not move supervisor/runtimeactor/session/provider adapters,
task DB/project/mwsd/local API packages, provider process execution, private
infra, secrets, or local machine state.

### RIID-4647 — workspace native config domain

This slice moves the pure C6 workspace / native config domain:

- `internal/workdir`
- `docs/20-domain/workspace.md`
- native config plan generator support needed by `go generate ./internal/workdir`

The package imports IR types from `github.com/teamswyg/riido-contracts/ir
v0.1.0`.

This slice does not move provider adapters, runtime/session/supervisor actors,
C7 policy/security implementation, C11 host integration implementation,
task DB/project/mwsd/local API packages, private infra, secrets, or local
machine state.

### RIID-4648 — distribution host integration domain

This slice moves the pure C11 distribution / host integration domain:

- `internal/hostintegration`
- `docs/20-domain/distribution-host-integration.md`
- `privacy_metadata_allowlist.riido.json` as C10/C11 privacy-boundary evidence

The package imports provider capability types from
`github.com/teamswyg/riido-contracts/provider/capability v0.1.0`.

This slice does not move provider adapters, runtime/session/supervisor actors,
C7 policy/security implementation, concrete OS adapters, task DB/project/mwsd
local API packages, packaging artifacts, private infra, secrets, or local
machine state.

### RIID-4649 — security policy domain

This slice moves the pure C7 security / policy decision domain:

- `internal/policy`
- `docs/20-domain/security.md`
- `docs/20-domain/security-redaction.md`

The package imports C11 host integration types from the public
`internal/hostintegration` package.

This slice does not move provider adapters, runtime/session/supervisor actors,
ToolRef.Args / EventIngestor wiring, concrete sandbox/network/OS adapters, task
DB/project/mwsd local API packages, packaging artifacts, private infra, secrets,
or local machine state.

### RIID-4650 — EventIngestor boundary

This slice moves the daemon-side C2 EventIngestor implementation:

- `internal/ir/ingest`
- `docs/20-domain/security-redaction.md` references to the now-public
  EventIngestor verification point

The package imports CanonicalEvent and envelope validation types from
`github.com/teamswyg/riido-contracts/ir v0.1.0`, and imports the local public
C7 policy redaction catalog from `internal/policy`.

This slice does not move provider adapters, runtime/session/supervisor actors,
ToolRef.Args flattening, concrete event sink wiring beyond the existing C6
workdir sink port, task DB/project/mwsd local API packages, packaging
artifacts, private infra, secrets, or local machine state.

### RIID-4651 — agentbridge root provider runtime domain

This slice moves the provider-neutral C4 Provider Runtime / Adapter root
domain:

- `internal/agentbridge`
- `docs/20-domain/provider-runtime.md`
- focused public CI for reducer / telemetry / blocked-arg / semantic-activity
  gates

The package is stdlib-only and intentionally does not import concrete provider
packages, task/project persistence, process execution implementations, local
API packages, or filesystem/network adapters.

This slice does not move `internal/agentbridge/session`,
`runtimeactor`, `supervisor`, `bridge`, `controlplane`, `detectutil`, concrete
provider adapters, `ToolRef.Args` flattening/toolpolicy execution, task
DB/project/mwsd local API packages, packaging artifacts, private infra,
secrets, or local machine state.

### RIID-4652 — toolargs / toolpolicy migration

This slice moves the provider-neutral C4/C7 tool-use bridge:

- `internal/agentbridge/toolargs`
- `internal/agentbridge/toolpolicy`
- docs updates in provider-runtime, security, and security-redaction SSOT files
- focused public CI for ToolRef.Args redaction, risk-surface classification,
  AutoApprover, and ToolStartGate gates

`toolargs` turns provider raw tool input into a bounded string map and redacts
sensitive keys or values with the `ToolRef.Args` marker. `toolpolicy` maps
`agentbridge.ToolRef` into C7 `ToolUseSurface` decisions and only auto-approves
when the active policy bundle explicitly allows the classified surface.

This slice does not move concrete provider adapter parser/wiring,
session/runtimeactor/supervisor execution, provider-native approval RPC/hook
implementation, ToolCallStarted fail-close wiring, task DB/project/mwsd local
API packages, packaging artifacts, private infra, secrets, or local machine
state.

### RIID-4653 — session actor migration

This slice moves the provider-neutral C4 run-scope session actor:

- `internal/agentbridge/session`
- docs updates in provider-runtime and daemon migration SSOT files
- focused public CI for session backpressure, timeout/cancellation, process
  exit ordering, protocol-driver lifecycle, telemetry extraction, tool-start
  blocking, and adapter temp-file cleanup gates

The session actor connects Process -> Parser/ProtocolDriver -> reducer ->
bounded Events/Result streams for a single provider run. It is still
provider-neutral and uses only the public `internal/process` port plus the
public `internal/agentbridge` domain.

This slice does not move runtimeactor, supervisor, bridge/controlplane,
concrete provider adapters, task DB/project/mwsd local API packages,
provider-native approval RPC/hook implementations, packaging artifacts,
private infra, secrets, or local machine state.

### RIID-4654 — bridge/detectutil migration

This slice moves the provider-neutral C4 bridge entrypoint and provider adapter
detect helpers:

- `internal/agentbridge/bridge`
- `internal/agentbridge/detectutil`
- docs updates in provider-runtime and daemon migration SSOT files
- focused public CI for bridge run/detect/session handoff and detectutil
  fail-closed probe gates

The bridge package wires adapter `BuildStart` output into the public
`internal/process` port and the public `internal/agentbridge/session` actor. It
also preserves `ProtocolDriverProvider`, dropped args, and adapter temp-file
handoff behavior. The detectutil package owns env override pinning, PATH
fallback, version probe, and strict exit-code probe helpers that concrete
provider adapters can use later.

This slice does not move runtimeactor, supervisor, controlplane, concrete
provider adapters, provider-native approval RPC/hook implementations, task
DB/project/mwsd local API packages, packaging artifacts, private infra,
secrets, or local machine state.

### RIID-4656 — runtimeactor migration

This slice moves the provider-neutral C4/C5 runtime actor:

- `internal/agentbridge/runtimeactor`
- docs updates in provider-runtime and daemon migration SSOT files
- focused public CI for runtime actor mailbox, slot, cancellation, heartbeat,
  capability reconciliation, detected fingerprint, and protocol-driver handoff
  gates

The runtimeactor package owns one mailbox actor per RuntimeID/capability
boundary. It reconciles adapter `Detect` output into the public C3
`github.com/teamswyg/riido-contracts/provider/capability` model, enforces
runtime slot guards, starts one-run sessions, handles cancellation cascade, and
publishes status/heartbeat snapshots.

This slice does not move supervisor, controlplane, concrete provider adapters,
task DB/project/mwsd local API packages, provider-native approval RPC/hook
implementations, packaging artifacts, private infra, secrets, or local machine
state.

### RIID-4657 — controlplane ports migration

This slice moves the provider-neutral control-plane port contract:

- `internal/agentbridge/controlplane`
- docs updates in provider-runtime and daemon migration SSOT files
- focused public CI for task source/reporter ports, runtime registration,
  heartbeat, file queue claim receipts, provider availability filtering, and
  JSONL report records

The controlplane root package owns the public daemon-side ports used by the
future supervisor to register runtimes, send heartbeats, claim tasks, watch
cancellation, and report task start/event/result records. The in-tree
implementations remain black-box local adapters: RAM-only source/reporter for
tests and offline mode, file queue source for JSON task files plus claim
receipts/runtime registry, and file reporter for task-scoped JSONL receipts.

This slice does not move `controlplane/saasplane`, `controlplane/taskdbplane`,
supervisor polling/runtime selection, concrete provider adapters, server HTTP
transport, task DB/project/mwsd local API packages, packaging artifacts, private
infra, secrets, or local machine state.

### RIID-4658 — Claude provider adapter migration

This slice moves the first concrete provider adapter:

- `internal/provider/claude`
- Claude adapter testdata
- docs updates in provider-runtime and daemon migration SSOT files
- focused public CI for Claude command construction, blocked protocol args,
  executable detection, stream-json parser, raw event translator, golden JSONL
  fixtures, and provider input approval frames

The Claude adapter owns only the daemon-side C4 adapter ACL for the external
Claude Code CLI. It does not bundle, install, or distribute the Claude CLI.
Real CLI integration remains opt-in through `AGENTBRIDGE_INTEGRATION=1`; public
CI runs deterministic black-box tests and keeps the integration test skipped
when the external CLI is absent.

This slice does not move Codex/OpenClaw/Cursor adapters, supervisor polling /
runtime selection, SaaS control-plane adapters, task DB/project/mwsd local API
packages, packaging artifacts, private infra, secrets, or local machine state.

## Validation Gates

Required before a daemon migration PR is mergeable:

```bash
go test ./...
go list -m all
```

When the migrated files include the old audit tooling, restore the stronger
private-repo gates in public CI:

```bash
make check
```

Provider real-CLI integration checks stay environment-gated:

```bash
AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/... -run TestIntegration -v
```

## Store Review Invariants

- Provider CLIs are external tools, not bundled app payloads.
- The daemon must expose local-only IPC, not public TCP listeners.
- Unsafe provider modes are opt-in policy decisions, not defaults.
- Host trust tier must reject unsafe bypass.
- App Store and MSIX helper/runtime contracts stay in C11 docs and tests.

## Open Follow-Ups

| Follow-up | Repository |
| --- | --- |
| Move CLI command surface under the separate CLI task. | `riido-daemon` / RIID-4635 |
| Promote shared DTO/schema only after two repositories need the same fact. | `riido-contracts` / RIID-4637 |
| Move SaaS server code separately. | `riido-control-plane` / RIID-4638 |
| Move Terraform/deploy evidence privately. | `riido-infra` / RIID-4639 |
