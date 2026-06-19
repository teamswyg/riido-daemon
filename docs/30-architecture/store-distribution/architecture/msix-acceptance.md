# Store Distribution Architecture: MSIX Acceptance

[Back to architecture](../architecture.md)

Policy snapshot: as of 2026-05-26, this section reflects Microsoft Store
Policies v7.19, packaged desktop app distribution, MSIX package upload, and
MSIX signing guidance. If those policies change, update this architecture and
the C11 distribution SSOT in the same work unit.

`msix-sideload` is the first viable Windows distribution target:

1. `.msix` / `.msixbundle` artifacts must be signed by a trusted certificate.
2. The manifest provides package identity and targets Windows Desktop.
3. Daemon state is stored only under package local data. Arbitrary home scanning
   and hardcoded user-path fallback are forbidden.
4. Local control is exposed only through Windows named pipe IPC. External TCP
   listeners are forbidden.
5. Background helper/startup behavior requires explicit consent. Windows service
   installation is forbidden by default.
6. Provider CLIs are never included in the package and are registered only as
   user-installed external tools.

`msix-store` is possible only with Store-review evidence:

1. Partner Center notes explain packaged desktop app / full-trust use, local
   helper responsibility, and visible consent UI.
2. Updates use Microsoft Store package updates first. In-app self-updaters are a
   forbidden surface.
3. Windows service installation is forbidden by default. Background execution
   requires packaged-app/full-trust policy review and explicit consent.
4. Review/demo mode verifies onboarding, provider status, workspace grant, and
   privacy/telemetry settings without provider CLIs.
5. Privacy policy and Store metadata state values not sent to SaaS: provider
   executable path, workspace absolute path, token, and API key.

Executable verification:

- `go run ./tools/storecontract -contract packaging/store/riido_daemon_store_distribution.riido.json -repo .`
  verifies signed package, Windows Desktop target, named pipe IPC, package local
  data, runFullTrust notes, Store-managed updates, review notes, provider CLI
  non-bundling, demo/review account, and privacy metadata allowlist.
- The same gate verifies the `msix-store` role, background rule, IPC transport,
  data root, and update mechanism.
- `NOTICE.md` must keep Multica provenance, Modified Apache License 2.0 fact,
  and provider CLI non-bundling language.
- `.github/workflows/store-distribution-contract.yml` runs the same gate on PRs
  and `main` pushes.
- Provider-free review/demo local control is verified by `internal/hostintegration`
  tests and public daemon CI.
