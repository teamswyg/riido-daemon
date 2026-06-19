# Store Distribution Architecture: macOS Helper / Login Item

[Back to architecture](../architecture.md)

`developer-id` and `mac-app-store` target the same user experience, but helper
registration and file boundaries differ.

| Channel | Helper shape | Startup / background rule | IPC/data root |
| --- | --- | --- | --- |
| `developer-id` | signed + notarized local helper/broker. `cmd/riido` is the current domain core. | LaunchAgent or Login Item registration is allowed after explicit consent. Revocation disables auto-start. | Unix socket / app data root under `~/Library/Application Support/riido`. External TCP listener is forbidden. |
| `mac-app-store` | sandboxed Store App bundle helper/login item. Helper permission is bound to App Sandbox and entitlement review notes. | Only `SMAppService` / Login Item plus explicit consent is allowed. Direct `~/Library/LaunchAgents` install is forbidden. | Local IPC and app data root live inside app group or sandbox container. Workspace access persists only through security-scoped grant. |

Common rules:

1. Provider CLIs are not included in the helper bundle. C11
   `ExternalToolRegistry` records only user-selected, env-override, and
   auto-detected provenance.
2. Background helper consent is sourced from the C11 `ConsentLedger`
   `background-helper` grant.
3. Channel allowance is checked pre-runtime by C7 `EvaluateStoreChannelPolicy`
   over `StoreSurfaceBackgroundHelper` and
   `StoreSurfaceDirectLaunchAgentInstall`.
4. Mac App Store review notes describe helper purpose, login item consent UX,
   sandbox entitlement use, provider CLI non-bundling, and review/demo mode.

Executable C11 role plan:

- `internal/hostintegration.ResolveHelperRuntimePlan` returns the current plan.
- `mac-app-store` returns sandboxed login item helper, `SMAppService` / Login
  Item registration, helper-owned Unix socket under app group/container data
  root, App Store-managed updates, no provider CLI bundling, no direct
  LaunchAgent install, no shared-location code install, no standalone code
  download, security-scoped workspace grant requirement, and helper purpose /
  entitlement / consent review note surfaces.
- A concrete macOS Store App packaging adapter maps this plan to app bundle
  target, entitlements, login item registration, and security-scoped bookmark
  handling.
