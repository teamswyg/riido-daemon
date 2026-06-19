# RIID-4661: Cursor Provider Adapter

[Back to codex-auth-full-access](../codex-auth-full-access.md)

This slice moves the Cursor concrete provider adapter.

Moved surfaces:

- `internal/provider/cursor`
- Cursor adapter testdata
- docs updates in provider-runtime and daemon migration SSOT files
- focused public CI for Cursor command construction, launch profiles, unsafe
  `--yolo` policy gate, unsupported feature warnings, executable detection,
  stream-json parser, raw event translator, and golden fixtures

The Cursor adapter owns only the daemon-side C4 adapter ACL for the external
Cursor Agent CLI. It does not bundle, install, or distribute the Cursor Agent
CLI.

Real CLI integration remains opt-in through `AGENTBRIDGE_INTEGRATION=1`. Public
CI runs deterministic black-box tests and keeps integration skipped when the
external CLI is absent.

This slice does not move supervisor polling/runtime selection, SaaS control-plane
adapters, task DB/project/mwsd local API packages, packaging artifacts, private
infra, secrets, or local machine state.
