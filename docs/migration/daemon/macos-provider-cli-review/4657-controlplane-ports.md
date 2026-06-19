# RIID-4657 — Controlplane Ports Migration

[Back to macOS Provider CLI Review](../macos-provider-cli-review.md)

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
