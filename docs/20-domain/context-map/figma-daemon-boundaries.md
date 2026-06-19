# Figma Daemon Boundaries

[Back to context-map.md](../context-map.md)

Figma `node-id=156-19307` menu placement is a client route affordance. The
daemon may power runtime status after a route is opened, but it does not own
menu labels, ordering, selected state, or route availability.

Figma `node-id=162-23090` runtime settings is a client composition over
control-plane device/runtime liveness and the local daemon lifecycle surface.
The daemon owns current-device local facts exposed by `riido daemon status`,
`health`, `ready`, `metrics`, `stop`, `logs`, and `start`. It does not own the
agent hover popover, daemon stop modal copy, restart animation, remote-device
presentation, or SaaS `GET /v1/client/ai-agent/devices` projection.

Figma `node-id=432-37336` agent settings, `node-id=134-6542` agent add,
`node-id=337-24001` / `node-id=337-24013` agent list/add affordance, and
`node-id=432-35713` agent list are client/control-plane composition over agent
bootstrap/create/update/editability APIs and authorized device/runtime read
models.

The daemon owns only runtime execution after an already-authorized runtime
binding/model/instruction is assigned. It does not create agent records, stamp
`created_at`, refresh `updated_at`, enable or disable the save/add button,
decide whether all visible members have selectable runtimes, own row/meatball
edit entry, absolute-time tooltip behavior, no-description row layout,
long-description presentation, status-label copy/color, or the model dropdown
catalog.

Figma `node-id=275-22731` runtime-settings empty states are the same boundary.
The daemon supplies local liveness/detection facts when it is running; it does
not own Windows app waitlist copy, marketing-consent mutation, provider
install-card hover behavior, or external provider installation links. Claude,
Codex, OpenClaw, and Cursor CLIs remain external user-installed tools.

Figma `node-id=153-15935` additional planning content confirms the assignment
target boundary. The daemon does not decide whether a project, milestone,
intake, existing AI property filler, mention surface, task, or subtask can
receive an agent. That target validation is a contracts/control-plane/client
surface rule. The daemon consumes only SaaS assignments that already passed
target-scope policy, then controls the selected runtime.

Direct host-helper actions remain current-device local facts in C11. SaaS
assignment and daemon command authorization are agent-access facts owned
upstream: public agents can delegate indirect runtime execution to workspace
members, and private agents limit that path to admins and owners.
