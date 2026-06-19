# Security Invariants: Policy Bundle Loader

[Back to invariants](../invariants.md)

The current physical C7 policy bundle form is one JSON file.

Load rules:

- The file path is injected through `RIIDO_POLICY_BUNDLE_PATH`.
- File `version` becomes the active `PolicyBundleVersion`.
- If `RIIDO_POLICY_BUNDLE_VERSION` is also set, it must exactly match file
  `version` or daemon config loading fails.
- When no path is set, daemon uses built-in `policy-bundle.local.v0`.
- Unknown fields are rejected fail-closed.

The executable loader currently accepts C7's minimal execution subset:

- `unsafe_bypass`
- `native_config_hooks`
- `native_config_files`
- `tool_use`

Canonical JSON shape example: [policy-bundle-example.md](policy-bundle-example.md).

Built-in local development default:

- Host trust tier allows only Claude audit-only command hook:
  `claude:command-hooks:audit`.
- Unsafe bypass, native config file, and tool-use risk surfaces are not allowed.
- Codex app-server full-access sandbox envelope is not a native config file
  surface; it is selected by the C4 command builder.
- `RIIDO_POLICY_BUNDLE_VERSION` without a bundle path changes only the built-in
  version tag.

Validation rules:

1. `schema_version` is `riido-policy-bundle.v1`.
2. `version`, `effective_since`, and `trust_tier_policies` are required.
3. `superseded_at`, when present, is after `effective_since`.
4. `trust_tier_policies` keys are trust tier enum values only.
5. `allowed_surfaces.unsafe_bypass` values are known provider unsafe bypass
   surfaces and cannot repeat.
6. `Host` and `Unknown` cannot allow unsafe bypass surfaces.
7. `allowed_surfaces.native_config_hooks` values are known provider hook surfaces
   and cannot repeat. Current surface is `claude:command-hooks:audit`.
8. `Unknown` cannot allow native config hook surfaces.
9. `allowed_surfaces.native_config_files` values are known provider config
   file/home surfaces and cannot repeat.
10. `codex:config-home:task-scoped` stays parser-known for compatibility, but the
    current native config plan does not materialize it.
11. `Unknown` cannot allow native config file surfaces.
12. `allowed_surfaces.tool_use` values are known tool-use risk surfaces and
    cannot repeat.
13. `Unknown` cannot allow tool-use risk surfaces.
