# Store Distribution Architecture SSOT: Part 02

[Back to store-distribution.md](../store-distribution.md)

## 3. Required daemon changes

| Work | Owner context | Output |
| --- | --- | --- |
| Local IPC abstraction | C11 | Unix socket / Windows named pipe adapters behind one port |
| App data root abstraction | C11/C6 | dev-local / Developer ID / App Store / MSIX path selection |
| ExternalToolRegistry | C11/C3 | provider path provenance + version/login status |
| ConsentLedger | C11/C7 | background/provider/workspace/telemetry grants |
| WorkspaceGrantStore | C11/C6 | macOS security-scoped bookmark / Windows folder grant representation |
| StoreChannelPolicyGate | C11/C7 | channel-specific blocked reasons |
| Store-safe demo mode | C11 | `EvaluateReviewDemoMode` + local API `review-demo` reviewable/offline UX without provider spawn or SaaS telemetry sync |
| Distribution contract gate | C11 | executable check that provider binaries are not bundled |

## 4. Required server changes

| Work | Owner context | Output |
| --- | --- | --- |
| Distribution metadata | C10 | daemon poll/register payload includes channel/app version/status only |
| Provider status sync | C10 | available/login-required/unsupported without path/token |
| Capability routing gate | C10 + C3/C7 | task assignment excludes store-blocked or login-required runtimes |
| Review/demo account | C10 | `review_account_seed.riido.json` store-review-only SaaS seed/provisioning without real provider CLI |
| Privacy policy alignment | C10 | API collection scope matches `privacy_metadata_allowlist.riido.json` public policy metadata allowlist |

## 5. Review notes contract

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

## 6. Executable contract

The local preflight contract lives at [`../../packaging/store/riido_daemon_store_distribution.riido.json`](../../packaging/store/riido_daemon_store_distribution.riido.json).

Run the executable contract locally with:

```bash
go run ./tools/storecontract -contract packaging/store/riido_daemon_store_distribution.riido.json -repo .
```

The contract validates:

1. provider CLI bundling is explicitly forbidden;
2. required SSOT documents exist;
3. required store channels are declared;
4. `developer-id` declares signing, notarization, user-consented background helper, and local-only IPC surfaces;
5. `mac-app-store` declares sandbox, app group/container IPC, security-scoped workspace grant, Service Management login item consent, helper purpose review note, App Sandbox review notes, Store-managed updates, privacy policy, review/demo mode, demo/review account, privacy metadata allowlist, and provider non-bundling review note surfaces;
6. `msix-sideload` / `msix-store` declare signed/package identity, local IPC/data, review notes, Store update surfaces, demo/review account, privacy metadata allowlist, and provider non-bundling review note surfaces appropriate to each channel;
7. store artifact roots do not contain files named like provider executables;
8. store artifact roots do not contain hardcoded developer user paths.

For store-managed helper/runtime channels the contract also validates role-level fields:
`runtime_role`, `background_rule`, `local_ipc_transport`, `data_root`, and
`update_mechanism`. This keeps the `mac-app-store` sandboxed login item helper
decision and the `msix-store` full-trust helper/tray decision executable instead
of leaving them only in prose.

This gate does not prove App Store / Microsoft Store acceptance. It prevents the repository from drifting away from the product shape required for submission.

## 7. External sources

Policy can change, so these are fact sources, not copied rules:

- Apple App Review Guidelines: <https://developer.apple.com/app-store/review/guidelines/>
- Apple Developer ID: <https://developer.apple.com/support/developer-id/>
- Microsoft Store Policies: <https://learn.microsoft.com/en-us/windows/apps/publish/store-policies>
- MSIX packaged desktop apps: <https://learn.microsoft.com/en-us/windows/msix/desktop/desktop-to-uwp-behind-the-scenes>

If these external policies change, this document and the C11 domain SSOT must be updated in the same work unit.
