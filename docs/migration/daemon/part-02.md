# Riido Daemon Migration Plan: Part 02

[Back to daemon.md](../daemon.md)

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

