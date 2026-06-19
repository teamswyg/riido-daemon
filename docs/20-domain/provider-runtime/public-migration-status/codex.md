# Codex Adapter

[Back to Public Migration Status](../public-migration-status.md)

RIID-4659 moved `internal/provider/codex` into public `riido-daemon`.

The package does not bundle Codex CLI. It owns
`codex --sandbox danger-full-access app-server --listen stdio://` construction,
daemon-owned full-access runtime selection, JSONL parser, raw event translator,
JSON-RPC protocol driver, pending request actor, and approval response path.

Codex app-server can use the user's existing Codex auth store. Workdir is a
daemon-selected task/evidence root, not a filesystem sandbox boundary. The
full-access launch shape is C4 adapter-owned harness policy, not provider default
or caller input.

A-57 added a real CLI gate that expects `ResultCompleted` and an expected file
artifact inside daemon-selected workdir. The gate requires
`AGENTBRIDGE_INTEGRATION=1` and local Codex auth/runtime.

Codex runtime model catalog may report host Codex config `model` as
runtime-scoped opaque `model_id`. Daemon never infers model catalog from
OpenAI/ChatGPT tokens, account identity, API key, team id, or Open API key.

Control-plane fallback catalog ids such as `codex-default`, `claude-default`,
`openclaw-default`, `cursor-auto`, and `runtime-default` are client read-model
sentinels. `saasplane` preserves them in assignment metadata but normalizes them
to `StartRequest.Model=""`; provider native flags need real provider model ids.
