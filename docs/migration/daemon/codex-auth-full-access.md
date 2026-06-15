# Riido Daemon Migration Plan: Codex Auth / Full Access

[Back to daemon.md](../daemon.md)

### RIID-4881 / RIID-4917 — Codex app-server auth and full-access harness correction

The initial task-scoped `CODEX_HOME` model prevented Codex from reading the
user's global config but also removed Codex-managed ChatGPT auth, causing real
SaaS assignments to fail with provider 401 responses. RIID-4881 first corrected
that by keeping the real Codex auth store available while trying a daemon-owned
permission profile for provider tool commands. RIID-4917 supersedes that
permission-profile experiment: the daemon no longer tries to own Codex internal
filesystem permission semantics. The corrected boundary is:

- C6 no longer materializes Codex `.codex/config.toml` or
  `native_config_home=<workdir>/.codex`.
- C4 `internal/provider/codex` starts
  `codex --sandbox danger-full-access app-server --listen stdio://`.
- `danger-full-access` is not a provider default, caller-provided default, or
  hidden fallback. It is the only Codex sandbox selection the daemon generates
  for the local provider runtime, and that choice is paired with the daemon
  harness responsibilities below.
- Free-form custom args cannot pass `-c`, `--config`, `--enable`, or `--disable`
  because those could rewrite the daemon-owned launch/trust shape.
- Free-form custom args cannot pass `--sandbox`, `--sandbox=*`, `-s`, `-s=*`,
  `--yolo`, or `--dangerously-bypass-approvals-and-sandbox` because sandbox
  selection and approval-bypass surfaces are not caller-owned.
- The harness owns assignment snapshot, daemon-selected workdir/evidence root,
  provider process start/stop/cancel, heartbeat/stale lease handling, terminal
  result reporting, dropped arg evidence, and provider integration gates.
- Codex runtime registration reports the host Codex config `model` value as the
  default runtime-scoped model catalog entry. This prevents control-plane agent
  assignment from snapshotting an invented `codex-default` model that Codex
  app-server rejects for ChatGPT-account runs.

RIID-4917 also records the structural split of this decision:
`docs/20-domain/security.md` owns the C7 judgment that full-access/trusted
runtime envelopes are explicit harness-managed choices rather than provider or
caller defaults, while `docs/20-domain/provider-runtime.md` §2.1 owns the C4
provider-by-provider adoption table. Codex is currently adopted; Claude,
Cursor, and OpenClaw must not be treated as silently adopted without their own
SSOT, command-builder tests, and real integration evidence.

This keeps real Codex auth usable for development E2E and treats Codex as
trusted local automation. The structural decision is not "make full access the
default"; it is "when a provider must work on the user's machine, make the
trusted/full-access launch explicit and make Riido's harness own lifecycle,
workdir, heartbeat, lease, cancellation, and evidence." It does not use team id,
OpenAPI key, or task-location metadata as any part of Codex identity or sandbox
binding. Other providers should follow the same full-access/trusted-runtime
meta model only through provider-specific SSOT, command builder changes, and
integration evidence.

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
