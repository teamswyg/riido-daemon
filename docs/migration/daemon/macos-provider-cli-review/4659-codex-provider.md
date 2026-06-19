# RIID-4659 — Codex Provider Adapter Migration

[Back to macOS Provider CLI Review](../macos-provider-cli-review.md)

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
