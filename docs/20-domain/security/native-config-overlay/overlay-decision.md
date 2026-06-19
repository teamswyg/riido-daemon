# Native Config Overlay: T-CFG Decision

[Back to native-config-overlay](../native-config-overlay.md)

T-CFG treats provider-native config overlay as default-deny.

Primary instruction file materialization, such as `CLAUDE.md` and `AGENTS.md`,
is the deterministic default surface owned by C6 Workspace. Provider-native hook
settings, task-scoped config home, wrapper manifest, and MCP config injection are
active only when the policy bundle explicitly allows that surface.

Rules:

1. User-global native config overlay is not allowed by default.
2. C6 never reads or copies global `~/.claude`, `~/.codex`, Cursor, or OpenClaw
   config homes.
3. When a provider-native config home is required, C7 allows a known surface and
   C6 passes only the per-task workdir materialized config home to the adapter.
4. Codex does not create task-scoped `CODEX_HOME`. Codex app-server may use the
   existing user Codex auth store.
5. Codex launch shape is fixed by C4 as
   `codex --sandbox danger-full-access app-server --listen stdio://`.
6. Workdir is provider cwd and evidence root, not filesystem sandbox boundary.
7. Team id, OpenAPI key, and workspace task-location are not Codex auth or
   sandbox bridge inputs.
8. Runtime `.codex` state created under workdir is not C6 materialization and
   must not be used as auth copy, symlink, or SaaS credential snapshot.
9. Dirty workdir policy/config changes use T-POLICY / T-CONFIG runtime upgrade
   flow, not automatic in-place reinjection.
