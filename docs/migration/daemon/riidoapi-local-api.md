# Riido Daemon Migration Plan: riidoapi Local API

[Back to daemon.md](../daemon.md)

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
The SaaS heartbeat cadence is 5 seconds and the control-plane stale lease
deadline is 20 seconds. If heartbeat response omits a requested active
assignment, the daemon treats the assignment as server-side stale/cancelled and
signals the local cancellation watcher instead of continuing to report that
lease.
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
keeps direct host-helper actions current-device scoped. SaaS authorization for
assigned work and daemon command requests remains an upstream agent-access
policy: public agents can delegate indirect runtime execution to workspace
members, while private agents limit that path to admins and owners.

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
