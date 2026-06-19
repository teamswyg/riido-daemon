# Boundary Criteria

[Back to Figma AI Agent Daemon Boundary](../figma-ai-agent-daemon-boundary.md)

- Figma is product/design evidence, not daemon durable SSOT.
- contracts/control-plane owns agent, workspace, thread, and generated API meaning first.
- daemon consumes accepted assignments, runtime/model/instruction snapshots,
  provider detection/liveness, and stop/cancel/lifecycle commands.
- daemon does not own client copy, sorting, dropdown, modal, scroll, animation,
  timestamp, fixture row, workspace selection, waitlist, or marketing consent.
