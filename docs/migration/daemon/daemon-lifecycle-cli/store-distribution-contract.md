# RIID-4703: Store Distribution Contract

[Back to daemon-lifecycle-cli](../daemon-lifecycle-cli.md)

This slice moves the executable store distribution contract into the public
daemon repository.

Moved surfaces:

- `packaging/store/riido_daemon_store_distribution.riido.json`
- `tools/storecontract`
- `docs/30-architecture/store-distribution.md`
- `NOTICE.md`
- `.github/workflows/store-distribution-contract.yml`

The gate fixes the public daemon boundary for Developer ID, Mac App Store, MSIX
sideload, and Microsoft Store review surfaces. It makes provider CLI
non-bundling, store-managed update rules, local-only IPC, App Sandbox/login item
expectations, Windows named pipe/package local data expectations, demo review
account surface, and privacy metadata allowlist requirements executable.

This slice does not build/sign/notarize app bundles, produce MSIX packages,
submit to App Store Connect or Partner Center, bundle provider CLIs, move private
infra/account artifacts, or change the control-plane review account runtime seed.
The SaaS review account artifact remains owned by `riido-control-plane`.
