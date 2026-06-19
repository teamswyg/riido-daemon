# RIID-4656 — Runtimeactor Migration

[Back to macOS Provider CLI Review](../macos-provider-cli-review.md)

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
