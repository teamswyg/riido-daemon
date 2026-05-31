# Riido Daemon Migration Plan

> Riido task: RIID-4636 `[Daemon] 기존 riido_daemon daemon 마이그레이션 계획/문서화`

This document defines how the daemon/runtime part of the former private
`riido_daemon` repository moves into the public `riido-daemon` repository.

## Goal

`riido-daemon` owns the customer-PC daemon runtime, local host integration, and
provider execution boundary. It must stay public, store-reviewable, and
free of non-Riido dependencies unless a later ADR explicitly changes that rule.

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
| Local workspace projection / mwsd ACL | `internal/mwsdbridge`, `internal/project` |
| C11 Distribution / Host Integration | `internal/hostintegration`, local transport pieces in `internal/riidoapi`, `packaging/store`, `tools/storecontract`, `NOTICE.md` |

The documentation source is:

- `docs/20-domain/provider-runtime.md`
- `docs/20-domain/provider-capability.md`
- `docs/20-domain/security.md`
- `docs/20-domain/security-redaction.md`
- `docs/20-domain/distribution-host-integration.md`
- `docs/30-architecture/module-decomposition.md`
- `docs/30-architecture/integration-matrix.md`
- `docs/30-architecture/config-reference.md`
- `docs/30-architecture/store-distribution.md`
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

`riido-daemon` started by consuming
`github.com/teamswyg/riido-contracts v0.1.0` and keeping CI limited to
Riido-owned Go module dependencies. This was a compatibility gate only; it did
not move runtime packages. Later slices may bump the module version when a
newly migrated daemon adapter needs a newer shared contract.

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
`github.com/teamswyg/riido-contracts/provider/capability`; the module version
is the current `go.mod` contract version.

This slice does not move supervisor/runtimeactor/session/provider adapters,
task DB/project/mwsd/local API packages, provider process execution, private
infra, secrets, or local machine state.

### RIID-4647 — workspace native config domain

This slice moves the pure C6 workspace / native config domain:

- `internal/workdir`
- `docs/20-domain/workspace.md`
- native config plan generator support needed by `go generate ./internal/workdir`

The package imports IR types from `github.com/teamswyg/riido-contracts/ir`; the
module version is the current `go.mod` contract version.

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
`github.com/teamswyg/riido-contracts/provider/capability`; the module version
is the current `go.mod` contract version.

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
`github.com/teamswyg/riido-contracts/ir`, and imports the local public C7
policy redaction catalog from `internal/policy`. The module version is the
current `go.mod` contract version.

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

### RIID-4572 — runtime/session backpressure and context boundary closure

This slice closes the discussion-complete C4 runtime/session boundary work:

- process stdout/stderr stream buffers are SSOT constants in `internal/process`
  and stay fixed at 64 chunks each
- session event/result buffers stay fixed at 256 events and 1 terminal result
- runtime actor mailbox defaults to 16 messages
- supervisor actor mailbox defaults to 64 messages
- provider runtime streams remain lossless bounded streams; full buffers block
  and propagate backpressure instead of dropping text/log/warning events
- `internal/agentbridge/session` remains a C4 internal submodel, not a separate
  bounded context

The slice adds focused public CI for these default-size and no-drop
backpressure gates. It does not add provider CLI dependencies, retry queues,
EventIngestor/outbox durability, concrete provider adapter ownership, task
DB/project/mwsd local API packages, packaging artifacts, private infra,
secrets, or local machine state.

### RIID-4570 — Store App repo/adapter ownership closure

This slice closes the Store App ownership discussion by moving `Q-CTX-005` out
of open questions and into C11 / architecture SSOT:

- `riido-daemon` owns C11 pure domain facts, helper runtime planning, local IPC
  server contracts, and store distribution gates
- a future desktop/app repository may own concrete Store App GUI, native
  entitlement calls, picker/bookmark adapters, App Store/MSIX project files,
  and submission UI surfaces
- Store App GUI must remain a client of C11/local API contracts and must not
  spawn provider CLIs directly, bundle provider CLIs, or copy C11 domain facts
- signing/provisioning secrets and live store submission evidence remain
  outside public repositories

The slice adds focused public CI that fails if `Q-CTX-005` returns to daemon
open questions or if the Store App ownership SSOT loses its repository
boundary wording.

### RIID-4571 — macOS external Provider CLI entitlement/review closure

This slice closes `Q-DIST-001` by making the Mac App Store external Provider
CLI strategy executable:

- Claude / Codex / OpenClaw / Cursor CLIs remain external user-installed tools
  and are never bundled, downloaded, or silently installed by the Store App
- `mac-app-store` Provider CLI execution requires both an OS grant
  (`StoreChannelPolicyInput.OSGrantPresent=true`) and App Review approval
  (`StoreChannelPolicyInput.StoreReviewApproved=true`)
- when either proof is missing, the provider may be shown as detected /
  login-required / store-blocked, but C4 must not spawn it
- App Review notes must explain the external-tool execution surface, explicit
  provider-execute consent, security-scoped workspace access, local-only helper,
  provider non-bundling, and provider-free review/demo mode
- executable paths, bookmark bytes, entitlement proof, signing/provisioning
  secrets, and live submission evidence remain local/private and are not sent
  to C10 or checked into public repositories

The slice adds focused public CI for the `Q-DIST-001` closure and C7
store-channel policy test.

### RIID-4573 — Workdir archive/retention/cache/native config closure

This slice closes the public daemon workdir policy discussion by absorbing
`Q-WS-001` through `Q-WS-006` into the C6/C7/runtime-upgrade SSOT:

- local archive default is same-host `keep-in-place`; external archive backends
  require an explicit future adapter/config
- workdir cleanup is disabled by default and only the opt-in TTL env is active;
  there is no implicit size or task-count cleanup
- shared repo cache prune is operator-triggered maintenance only, guarded by
  the short `repo_cache_update.lock`
- native config overlay means per-task materialization; user-global config
  copy/overlay is not a default behavior
- container/VM workdir handoff belongs to the future C4 runtime launcher /
  platform adapter, while C6 only prepares host-side files and manifests
- dirty workdir native-config reinjection threshold is zero; changes after
  `Preparing`/`Running` use the no-silent-upgrade flow and next-run
  recomputation

The slice adds focused public CI for the workdir policy closure and the
existing workdir cleanup/native-config tests.

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

### RIID-4659 — Codex provider adapter migration

This slice moves the Codex concrete provider adapter:

- `internal/provider/codex`
- Codex adapter testdata
- docs updates in provider-runtime and daemon migration SSOT files
- focused public CI for Codex command construction, blocked protocol args,
  unsafe bypass filtering, `CODEX_HOME` isolation, executable detection, JSONL
  parser, raw event translator, golden fixtures, JSON-RPC actor, handshake, and
  protocol-driver approval response path

The Codex adapter owns only the daemon-side C4 adapter ACL for the external
Codex CLI app-server stdio mode. It does not bundle, install, or distribute the
Codex CLI. Real CLI integration remains opt-in through `AGENTBRIDGE_INTEGRATION=1`;
public CI runs deterministic black-box tests and keeps the integration test
skipped when the external CLI is absent.

This slice does not move OpenClaw/Cursor adapters, supervisor polling / runtime
selection, SaaS control-plane adapters, task DB/project/mwsd local API packages,
packaging artifacts, private infra, secrets, or local machine state.

### RIID-4660 — OpenClaw provider adapter migration

This slice moves the OpenClaw concrete provider adapter:

- `internal/provider/openclaw`
- OpenClaw adapter testdata
- docs updates in provider-runtime and daemon migration SSOT files
- focused public CI for OpenClaw command construction, mandatory session id
  resolution, executable detection, calendar-version gate, JSON/NDJSON parser,
  raw event translator, and golden fixtures

The OpenClaw adapter owns only the daemon-side C4 adapter ACL for the external
OpenClaw CLI. It does not bundle, install, or distribute the OpenClaw CLI. Real
CLI integration remains opt-in through `AGENTBRIDGE_INTEGRATION=1`; public CI
runs deterministic black-box tests and keeps the integration test skipped when
the external CLI is absent.

This slice does not move the Cursor adapter, supervisor polling / runtime
selection, SaaS control-plane adapters, task DB/project/mwsd local API packages,
packaging artifacts, private infra, secrets, or local machine state.

### RIID-4661 — Cursor provider adapter migration

This slice moves the Cursor concrete provider adapter:

- `internal/provider/cursor`
- Cursor adapter testdata
- docs updates in provider-runtime and daemon migration SSOT files
- focused public CI for Cursor command construction, launch profiles,
  unsafe `--yolo` policy gate, unsupported feature warnings, executable
  detection, stream-json parser, raw event translator, and golden fixtures

The Cursor adapter owns only the daemon-side C4 adapter ACL for the external
Cursor Agent CLI. It does not bundle, install, or distribute the Cursor Agent
CLI. Real CLI integration remains opt-in through `AGENTBRIDGE_INTEGRATION=1`;
public CI runs deterministic black-box tests and keeps the integration test
skipped when the external CLI is absent.

This slice does not move supervisor polling / runtime selection, SaaS
control-plane adapters, task DB/project/mwsd local API packages, packaging
artifacts, private infra, secrets, or local machine state.

### RIID-4662 — supervisor migration

This slice moves the daemon supervisor actor:

- `internal/agentbridge/supervisor`
- docs updates in provider-runtime, runtime-scheduling, and daemon migration
  SSOT files
- focused public CI for supervisor task claim, RuntimeActor pool dispatch,
  pre-submit C5 eligibility, workdir/native-config injection, EventIngestor
  append delegation, terminal result reporting, stop cancellation, and archive
  gates

The supervisor package owns the in-process Daemon tier control loop. It
registers RuntimeActor instances with the control-plane source, sends
heartbeats, claims tasks by runtime id, evaluates public C3 capability snapshots
through the C5 scheduling evaluator, prepares per-run workdirs, delegates event
append to `internal/ir/ingest`, and reports terminal results through
`TaskReporterPort`.

This slice imports C1/C2/C3 domain types from
`github.com/teamswyg/riido-contracts` and does not reintroduce private
`riido_daemon` internal packages. It does not move `controlplane/saasplane`,
`controlplane/taskdbplane`, task DB/project/mwsd local API packages, server
HTTP transport, packaging artifacts, private infra, secrets, or local machine
state.

### RIID-4683 — taskdbplane local task DB adapter migration

This slice moves the local task DB control-plane adapter:

- `internal/taskdb`
- `internal/agentbridge/controlplane/taskdbplane`
- docs updates in provider-runtime, runtime-scheduling, locking, and daemon
  migration SSOT files
- focused public CI for guarded task DB mutation, command-id idempotent replay,
  runtime registry sidecars, lease sidecars, fencing tokens, expired lease
  handoff, and taskdbplane claim/report black-box scenarios

`internal/taskdb` owns the public daemon copy of the `riido-task-db.v1` JSON
schema and guarded mutation rules. It persists transitions, deterministic
validation evidence, and command receipts without importing private
`riido_daemon` packages or workspace projection code.

`internal/agentbridge/controlplane/taskdbplane` adapts that local JSON DB into
`TaskSourcePort` and `TaskReporterPort`. It claims only eligible `Queued` rows,
records `Queued -> Claimed -> Preparing -> Running -> Validating/terminal`
through C1 guarded transitions, stores runtime registry / lease sidecars next
to the task DB, and rejects progress/result reports without matching active
lease metadata.

This slice imports C1/C2/C3 domain types from
`github.com/teamswyg/riido-contracts` and does not reintroduce private
`riido_daemon` internal packages. It does not move project/mwsd sync, local
API/socket, CLI commands, `controlplane/saasplane`, server HTTP transport,
packaging artifacts, private infra, secrets, or local machine state.

### RIID-4684 — riidoapi local API adapter migration

This slice moves the daemon local-only API adapter:

- `internal/riidoapi`
- docs updates in distribution-host-integration, CLI migration, and daemon
  migration SSOT files
- focused public CI for local API status/tasks/transition/evidence/validate,
  review-demo mode, Unix socket transport, Windows named pipe path behavior, and
  no-public-TCP listener boundary checks

`internal/riidoapi` owns the local JSON envelope and the local transport
adapters used by GUI/Zed/CLI surfaces. The handler exposes `status`, `tasks`,
`transition`, `evidence`, `validate`, and `review-demo` over local IPC only.
Task mutations use public `internal/taskdb` guarded transition/evidence
receipts. Validation uses public `internal/validation` and rejects missing
`approval_id` before running the command.

This slice imports C1/C2 domain types from `github.com/teamswyg/riido-contracts`
and does not reintroduce private `riido_daemon` internal packages. It does not
move `cmd/riido` CLI commands, mwsdbridge/project projection sync,
`controlplane/saasplane`, server HTTP transport, packaging artifacts, private
infra, secrets, or local machine state.

### RIID-4685 — task/api/bridge CLI adapter migration

This slice moves the public-safe part of the local CLI:

- `cmd/riido`
- `riido task list|transition|evidence|validate`
- `riido serve`
- `riido api status|tasks|transition|evidence|validate|review-demo`
- `riido bridge providers|detect`
- focused public CI for CLI build, help output, bridge provider listing,
  guarded task validation scenarios, local API review-demo scenario, public
  boundary import checks, and local-only listener checks

The CLI remains a thin adapter. Task mutations call public `internal/taskdb`,
local IPC calls go through public `internal/riidoapi`, and provider listing uses
the public provider adapter ports. The CLI does not redefine FSM, IR,
validation, provider policy, or local transport decisions.

This slice does not move `riido mwsd ...`, mwsdbridge/project projection sync,
full `riido daemon ...` process lifecycle commands, `controlplane/saasplane`,
server HTTP transport, packaging artifacts, private infra, secrets, or local
machine state.

### RIID-4686 — mwsdbridge/project projection sync migration

This slice moves the public-safe mwsd workspace projection adapter:

- `internal/mwsdbridge`
- `internal/project`
- `riido mwsd snapshot|projection|sync|orchestration|projects|status`
- focused public CI for mwsd Unix-socket handshake, workspace projection,
  project state persistence, project-to-taskdb sync, CLI black-box sync, public
  boundary import checks, no-public-TCP listener checks, and stdlib-only Go
  dependency checks

`internal/mwsdbridge` is the anti-corruption layer for the local
macmini-workspace daemon. It reads only the mwsd JSON socket contracts and must
not parse macmini-workspace files directly.

`internal/project` owns the deterministic `riido-workspace-projection.v1` and
`riido-project-state.v1` files. Its task DB sync adapter is intentionally narrow:
it projects mwsd-discovered document tasks into public `internal/taskdb` records
and initial `TaskCreated` transitions, but guarded task mutation remains owned
by `internal/taskdb`.

The CLI remains a thin adapter for this slice. `riido mwsd sync` writes the
project state file and updates `riido-task-db.v1` through the projection sync
adapter. It does not start provider processes, talk to the SaaS server, open
public TCP listeners, or bundle the mwsd daemon.

This slice does not move full `riido daemon ...` process lifecycle commands,
`controlplane/saasplane`, server HTTP transport, packaging artifacts, private
infra, secrets, or local machine state.

### RIID-4689 — saasplane assignment polling adapter migration

This slice moves the daemon-side SaaS assignment polling/report adapter:

- `internal/agentbridge/controlplane/saasplane`
- docs updates in provider-runtime, runtime-scheduling, and daemon migration
  SSOT files
- focused public CI for assignment poll/start/cancel, heartbeat, progress/event
  sync, terminal result reporting, bearer auth forwarding, and public boundary
  import checks

`controlplane/saasplane` adapts the C10 assignment HTTP API into the existing
`TaskSourcePort` and `TaskReporterPort`. It polls `/v1/agents/{agent_id}/poll`,
turns `start` responses into `bridge.TaskRequest`, watches `cancel` responses
for in-flight task cancellation, forwards heartbeat active assignment ids, and
posts daemon progress/result events to `/v1/agents/{agent_id}/events`.
Task-thread progress intended for the client thread surface is reported as
bounded parsed batches to `/v1/agents/{agent_id}/thread-progress`; the daemon
must preserve SaaS-supplied task/run/thread identity when present and must not
invent the client cold-thread collection. Figma `node-id=153-15931` viewer-away
thread visibility and long-body scroll/focus behavior remain control-plane
cold-collection and client presentation facts, not daemon scheduling or provider
runtime state.
Figma `node-id=236-21379` normal task-thread rendering is the same boundary at
task-screen scale: generic comment input, AI Agent reply input, send-button
state, right details panel, and the visible `중지` button are client/task
presentation. The daemon responds only after SaaS polling returns cancellation
or interrupt state, then applies that to the provider runtime and reports
progress/result through existing ports.
Figma `node-id=153-8761` busy-agent queued rendering is also outside daemon
comment ownership. When SaaS reports that an already-working agent has accepted
a queued assignment/comment, the daemon does not synthesize the Korean
"지금은 다른 작업을 처리 중이에요..." copy or a task-thread row. It only waits for
the SaaS poll result that either grants work to this runtime or reports a
cancel/stop transition, then applies the provider action and reports existing
progress/result events.
Figma `node-id=227-19354` stopped-by-deleted-agent rendering follows the same
boundary. Agent deletion is a client/control-plane command; the daemon does not
decide that deletion, render the "에이전트가 삭제되어..." task-thread copy, or
create a Riido-authored thread row. If the control plane force-stops an assigned
runtime because the agent was deleted, the daemon receives that as the same
SaaS cancellation/stop path, interrupts the provider process when applicable,
and reports terminal progress/result through the existing adapter ports.

Figma `node-id=153-15935` additional planning content is also no daemon runtime
diff. The task/subtask-only Agent assignment target scope belongs upstream to
contracts/control-plane/client composition. The daemon must not infer Agent
targets for projects, milestones, intakes, AI property filling, or mention
surfaces, and it must not implement agent recommendation for the existing AI
property filler. It only executes SaaS assignments after target validation and
continues to treat device/runtime owner-only actions as local current-device
host integration behavior.

The adapter imports shared DTO/state/event/poll constants from
`github.com/teamswyg/riido-contracts/assignment v0.3.0` and does not import
private `internal/riidoaiserver` packages. The control-plane repository still
owns HTTP handlers, the task-thread cold collection read model, store actors,
SSE fan-out, authZ, metrics/health read models, persistence, and production
infra.

This slice does not move full `riido daemon ...` process lifecycle CLI wiring,
server HTTP implementation, SSE transport, Terraform/AWS/deploy evidence,
packaging artifacts, private infra, secrets, provider CLI bundling, or local
machine state.

### RIID-4690 — full daemon lifecycle CLI wiring migration

This slice restores the public-safe `riido daemon ...` process lifecycle
adapter:

- `cmd/riido/daemon.go`
- `cmd/riido/daemon_config.go`
- `riido daemon start|status|health|ready|metrics|stop|logs`
- docs updates in provider-runtime, runtime-scheduling, CLI migration, and
  daemon migration SSOT files
- focused public CI for foreground/background daemon start, local Unix socket
  status/health/ready/metrics, cooperative stop, PID fallback, log tailing,
  12-factor env config loading, control-plane source selection, public
  boundary import checks, and local-only listener checks

The CLI adapter wires the already public runtimeactor/supervisor/provider
adapters to the already public control-plane sources: in-memory offline mode,
file queue, `riido-task-db.v1` via `taskdbplane`, and SaaS assignment HTTP via
`saasplane`. Source selection is 12-factor env based:
`RIIDO_TASK_QUEUE_DIR`, `RIIDO_TASK_DB_SOURCE_PATH`, or `RIIDO_SAAS_URL` with
`RIIDO_SAAS_AGENTS`.

The daemon command imports public daemon packages only and must not import
private `riido_daemon` paths or `internal/riidoaiserver`. It does not bundle,
install, or auto-download Claude/Codex/OpenClaw/Cursor CLIs.

Figma runtime-settings empty states (`node-id=275-22731`) do not change that
boundary. Provider install cards and hover states are client/product
presentation over external provider links, and Windows app waitlist /
marketing-consent mutations are not daemon commands.

Figma web onboarding (`node-id=236-29749`) does not change that boundary either.
The macOS app download CTA is distribution/product routing to a Riido desktop
artifact, not a daemon command to install provider CLIs. Sign-up, terms consent,
member invite, Windows waitlist/marketing consent, chat animation, and
progress-bar references remain client/auth/team/product surfaces unless a future
daemon SSOT explicitly promotes a local helper behavior.

This slice does not move server HTTP implementation, SSE transport,
Terraform/AWS/deploy evidence, packaging artifacts, private infra, secrets,
provider CLI bundling, App Store/MSIX helper packaging, or local machine state.

### RIID-4703 — store distribution contract migration

This slice moves the executable store distribution contract into the public
daemon repository:

- `packaging/store/riido_daemon_store_distribution.riido.json`
- `tools/storecontract`
- `docs/30-architecture/store-distribution.md`
- `NOTICE.md`
- `.github/workflows/store-distribution-contract.yml`

The gate fixes the public daemon boundary for Developer ID, Mac App Store,
MSIX sideload, and Microsoft Store review surfaces. It makes provider CLI
non-bundling, store-managed update rules, local-only IPC, App Sandbox/login
item expectations, Windows named pipe/package local data expectations, demo
review account surface, and privacy metadata allowlist requirements executable.

This slice does not build/sign/notarize app bundles, produce MSIX packages,
submit to App Store Connect or Partner Center, bundle provider CLIs, move
private infra/account artifacts, or change the control-plane review account
runtime seed. The SaaS review account artifact remains owned by
`riido-control-plane`.

### RIID-4711 — architecture SSOT docs migration

This slice moves the public daemon architecture SSOT into `riido-daemon` after
the split-repo package migration.

This slice does:

- add `docs/20-domain/context-map.md` for public daemon bounded-context
  ownership and split-repo dependency direction
- add `docs/30-architecture/module-decomposition.md` for hexagonal package and
  import rules
- add `docs/30-architecture/config-reference.md` for daemon-only Factor 12
  env/flag ownership
- add `docs/30-architecture/integration-matrix.md` for optional real provider
  CLI integration gates
- add `docs/30-architecture/compatibility-gate.md` and
  `docs/30-architecture/runtime-upgrade-flow.md` for pre-execute and
  no-silent-upgrade boundaries
- add `docs/50-roadmap/open-questions.md` for public daemon unresolved
  questions referenced by domain SSOT docs
- add focused public CI for architecture doc presence, stale split-repo
  wording, config coverage, dependency boundary, and Go tests

This slice does not move `cmd/riido_ai_server`, `internal/riidoaiserver`,
Terraform/AWS/deploy evidence, private state, `.riido-local`, provider CLI
binaries, or provider installation automation.

### RIID-4630 — ApprovalRequested timeout owner SSOT cleanup

This slice closes the public daemon `Q-RT-003` open question by moving the
approval wait timeout decision into the provider-runtime SSOT.

This slice does:

- state that C4 session actor run clocks own approval wait timeout policy
- remove `Q-RT-003` from daemon open questions
- keep `EventIngestor` as an append authority only, not a timeout owner
- keep UI/review surfaces as display/response senders, not terminal timeout
  sources
- add a focused public workflow that fails if `Q-RT-003` drifts back into open
  questions
- add a reducer test that `EventToolApprovalNeeded` resets the semantic idle
  watchdog

This slice does not change provider-native approval RPC frames, add UI, change
CLI flags, introduce dependencies, or alter hard/semantic timeout defaults.

## Validation Gates

Required before a daemon migration PR is mergeable:

```bash
go test ./...
go list -m all
go test ./tools/storecontract
go run ./tools/storecontract -contract packaging/store/riido_daemon_store_distribution.riido.json -repo .
test -f docs/20-domain/context-map.md
test -f docs/30-architecture/config-reference.md
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
| Promote shared DTO/schema only after two repositories need the same fact. | `riido-contracts` / RIID-4637 |
| Move SaaS server code separately. | `riido-control-plane` / RIID-4638 |
| Move Terraform/deploy evidence privately. | `riido-infra` / RIID-4639 |
