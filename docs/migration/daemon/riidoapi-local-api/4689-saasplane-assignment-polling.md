# RIID-4689 — saasplane Assignment Polling Adapter Migration

[Back to riidoapi local API](../riidoapi-local-api.md)

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
