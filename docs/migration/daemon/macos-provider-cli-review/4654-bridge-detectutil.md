# RIID-4654 — Bridge/Detectutil Migration

[Back to macOS Provider CLI Review](../macos-provider-cli-review.md)

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
