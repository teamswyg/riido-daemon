# RIID-4660: OpenClaw Provider Adapter

[Back to codex-auth-full-access](../codex-auth-full-access.md)

This slice moves the OpenClaw concrete provider adapter.

Moved surfaces:

- `internal/provider/openclaw`
- OpenClaw adapter testdata
- docs updates in provider-runtime and daemon migration SSOT files
- focused public CI for OpenClaw command construction, mandatory session id
  resolution, executable detection, calendar-version gate, JSON/NDJSON parser,
  raw event translator, and golden fixtures

The OpenClaw adapter owns only the daemon-side C4 adapter ACL for the external
OpenClaw CLI. It does not bundle, install, or distribute the OpenClaw CLI.

Real CLI integration remains opt-in through `AGENTBRIDGE_INTEGRATION=1`. Public
CI runs deterministic black-box tests and keeps integration skipped when the
external CLI is absent.

This slice does not move Cursor adapter, supervisor polling/runtime selection,
SaaS control-plane adapters, task DB/project/mwsd local API packages, packaging
artifacts, private infra, secrets, or local machine state.
