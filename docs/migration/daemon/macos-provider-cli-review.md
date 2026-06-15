# Riido Daemon Migration Plan: macOS Provider CLI Review

[Back to daemon.md](../daemon.md)

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
  unsafe bypass filtering, `CODEX_HOME` non-materialization, executable detection,
  JSONL parser, raw event translator, golden fixtures, JSON-RPC actor, handshake,
  and protocol-driver approval response path

The Codex adapter owns only the daemon-side C4 adapter ACL for the external
Codex CLI app-server stdio mode. It does not bundle, install, or distribute the
Codex CLI. Real CLI integration remains opt-in through `AGENTBRIDGE_INTEGRATION=1`;
public CI runs deterministic black-box tests and keeps the integration test
skipped when the external CLI is absent.
