# Workspace Invariants: Native Config Injection

[Back to invariants](../invariants.md)

C6 owns the injection mechanism. Policy bundles decide what must exist; Workspace
creates the files deterministically under `workdir/`, `workdir/.claude/`,
`workdir/.riido/`, or provider-standard locations.

Injected files:

- `CLAUDE.md`, loaded by Claude Code.
- `AGENTS.md`, loaded by Codex.
- Claude `.claude/settings.json`.
- Hook settings such as `.claude/hooks/...` or equivalents.
- Wrapper manifest when present.
- `.riido/native-config-manifest.json`, using
  `riido-native-config-manifest.v1`.
- `.riido/` metadata: `task.json`, `policy-bundle.lock`,
  `native-config.lock`, and related run metadata.

Telemetry contract rule:

- If SaaS task source provides `riido_telemetry_contract`, supervisor injects the
  hard rule `<riido_log>{"code":...,"args":{...}}<end>` independently of
  provider-specific prompt placement.
- Progress code catalog and append-only policy are owned by
  `riido-contracts/progressmessage/catalog.dsl.riido.json`.
- The rule is part of `NativeConfigVersion` input hashes so replay can identify
  telemetry contract changes.

Agent instruction rule:

- SaaS agent instruction values are consumed only as run-scope prompt/native
  config input.
- Storage location, length limits, editability, RBAC, `profile_thumbnail_url`,
  `description`, and other presentation fields belong to public `riido-contracts`
  and `riido-control-plane`.
- Daemon does not inject thumbnail or description values into native config.
