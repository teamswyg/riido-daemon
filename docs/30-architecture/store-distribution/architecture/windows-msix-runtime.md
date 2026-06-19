# Store Distribution Architecture: Windows MSIX Runtime

[Back to architecture](../architecture.md)

`msix-sideload` and `msix-store` target the same Windows runtime UX, but differ
in packaging, update, and review evidence.

| Channel | Runtime shape | Packaging rule | Review / update rule |
| --- | --- | --- | --- |
| `msix-sideload` | signed MSIX local helper/broker. Helper owns provider runtime orchestration; Store App owns consent/status control surface. | signed package, package identity, Windows Desktop target device family, package local data, and named pipe local IPC are required. | Store review note is not required. Windows service install is forbidden by default; background helper requires explicit consent. |
| `msix-store` | Microsoft Store packaged desktop app plus packaged full-trust helper/tray process. Helper is described as local-only runtime broker. | package identity, Windows Desktop target device family, package local data, and named pipe local IPC are required. | `runFullTrust` / Partner Center notes, review/demo mode, privacy policy, and Store-managed updates are required. Self-updater is forbidden. |

Common rules:

1. Provider CLIs are not included in the MSIX package. Store App shows only
   user-selected, env-override, auto-detected provenance, and login-required
   status.
2. Local IPC uses Windows named pipe only. External TCP listener is forbidden.
3. App data and daemon state live under package local data root. `%USERPROFILE%`
   home fallback and arbitrary home scanning are forbidden.
4. Workspace access reaches runtime only through Windows folder picker grant and
   C11 consent.
5. `msix-store` review notes describe packaged desktop app / full-trust helper
   purpose, background consent UX, provider CLI non-bundling, privacy scope, and
   review/demo mode.

Executable C11 role plan:

- `internal/hostintegration.ResolveHelperRuntimePlan` returns the current plan.
- `msix-store` returns helper-owned named pipe, package local data root,
  Store-managed updates, no provider CLI bundling, no default Windows service
  install, no self-updater, and `runFullTrust` / Partner Center review note
  surface.
- A concrete Windows packaging adapter maps this plan to manifest, packaged
  full-trust process, and tray startup task implementation.
