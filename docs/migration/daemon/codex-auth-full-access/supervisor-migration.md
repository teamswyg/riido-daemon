# RIID-4662: Supervisor Migration

[Back to codex-auth-full-access](../codex-auth-full-access.md)

This slice moves the daemon supervisor actor.

Moved surfaces:

- `internal/agentbridge/supervisor`
- docs updates in provider-runtime, runtime-scheduling, and daemon migration SSOT
  files
- focused public CI for supervisor task claim, RuntimeActor pool dispatch,
  pre-submit C5 eligibility, workdir/native-config injection, EventIngestor append
  delegation, terminal result reporting, stop cancellation, and archive gates

The supervisor package owns the in-process Daemon tier control loop. It registers
RuntimeActor instances with the control-plane source, sends heartbeats, claims
tasks by runtime id, evaluates public C3 capability snapshots through C5
scheduling, prepares per-run workdirs, delegates event append to
`internal/ir/ingest`, and reports terminal results through `TaskReporterPort`.

This slice imports C1/C2/C3 domain types from `github.com/teamswyg/riido-contracts`
and does not reintroduce private `riido_daemon` internal packages. It does not
move `controlplane/saasplane`, `controlplane/taskdbplane`, task DB/project/mwsd
local API packages, server HTTP transport, packaging artifacts, private infra,
secrets, or local machine state.
