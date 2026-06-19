# Review Notes Contract

[Back to Daemon Changes](../daemon-changes.md)

Store submission notes must state:

1. Riido does not include Claude / Codex / OpenClaw / Cursor Agent binaries.
2. Users connect provider CLIs they installed separately.
3. Provider execution requires explicit user consent and visible connected status.
4. Workspace access is limited to user-selected folders.
5. Background helper / startup behavior is opt-in and revocable.
6. SaaS sync only uploads the C10/C11 allowlist: `distribution_channel`, `app_version`, `provider_kind`, `provider_available`, `provider_login_status`, `routing_status`, daemon/runtime identity required for polling/sync, assignment ids, task event state, and provider-neutral progress/result metadata.
7. SaaS sync does not upload provider executable paths, workspace absolute paths, provider tokens, API keys, or raw environment values.
8. Demo/review mode is available when provider CLIs are absent.
9. Reviewers receive a store-review-only account that can inspect onboarding, provider status, workspace grant, privacy/telemetry settings, public/private agent visibility, and non-provider task flow without requiring Claude / Codex / OpenClaw / Cursor Agent to be installed.
10. Review notes identify provider CLIs as external user-installed dependencies, state that Riido does not silently install them, and state that review/demo provider status is synthetic until a real user connects an external CLI.

Required store-review surfaces:

| Surface | Meaning |
| --- | --- |
| `demo-review-account` | App Review / Partner Center receives a SaaS account or review mode path that works without provider CLI installation. The control-plane repository owns the review account seed artifact and its CI validation. |
| `privacy-metadata-allowlist` | Store metadata and public privacy policy enumerate the allowed SaaS fields and the forbidden path/token/raw environment fields. Daemon executable artifact: `internal/hostintegration/privacy_metadata_allowlist.riido.json`; daemon CI and store distribution contract gate own local validation. |
| `helper-purpose-review-note` | Mac App Store review notes explain why the helper/login item exists, what user-visible consent enables it, and that provider runtime orchestration stays local-only. |
| `provider-non-bundling-review-note` | Review notes explicitly say Claude / Codex / OpenClaw / Cursor Agent CLIs are not bundled, redistributed, or silently installed. |
