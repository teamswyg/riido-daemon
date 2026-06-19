# RIID-4658 — Claude Provider Adapter Migration

[Back to macOS Provider CLI Review](../macos-provider-cli-review.md)

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
