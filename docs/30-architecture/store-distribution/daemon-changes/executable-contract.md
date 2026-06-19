# Executable Contract

[Back to Daemon Changes](../daemon-changes.md)

The local preflight contract lives at [`../../../packaging/store/riido_daemon_store_distribution.riido.json`](../../../packaging/store/riido_daemon_store_distribution.riido.json).

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
