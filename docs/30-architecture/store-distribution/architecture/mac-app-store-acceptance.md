# Store Distribution Architecture: Mac App Store Acceptance

[Back to architecture](../architecture.md)

Policy snapshot: as of 2026-05-28, this section reflects Apple App Review
Guidelines, App Sandbox entitlement documentation, App Store Connect App
Sandbox information, Service Management / `SMAppService`, and security-scoped
bookmark guidance. If those policies change, update this architecture and the
C11 distribution SSOT in the same work unit.

`mac-app-store` is possible only after constrained-mode redesign:

1. The app target enables App Sandbox and declares only required entitlements.
   Temporary exceptions require App Store Connect review notes.
2. Store App, helper, and broker are self-contained in the app bundle. Third-party
   installers, shared-location code install, and standalone code download are
   forbidden.
3. Background helper registration uses `SMAppService` / Login Item with user
   consent. Direct `~/Library/LaunchAgents` installation is forbidden.
4. Workspace access persists only through security-scoped bookmark or
   user-selected document/folder grant allowed by App Sandbox.
5. External provider CLI execution requires both user-selected/sandbox/
   security-scoped OS grant and App Review approval. Otherwise the Store App may
   show detected/login-required/store-blocked status, but the local helper must
   not spawn the provider process.
6. Local control is exposed only through app group/container local IPC. External
   TCP listeners are forbidden.
7. Updates use Mac App Store update paths. In-app self-updaters are forbidden.
8. Root privilege escalation and setuid-like behavior are forbidden.
9. Review/demo mode verifies onboarding, provider status, workspace grant, and
   privacy/telemetry settings without provider CLIs.
10. Privacy policy and Store metadata state values not sent to SaaS: provider
   executable path, workspace absolute path, token, and API key.

Executable verification:

- `tools/storecontract` checks App Sandbox, app group/container IPC,
  security-scoped workspace grant, Service Management login-item consent, helper
  purpose review note, App Sandbox review notes, App Store-managed updates,
  privacy policy, review/demo mode, demo/review account, privacy metadata
  allowlist, provider non-bundling review note, and forbidden surfaces.
- RIID-4571 external provider CLI execution is enforced by C7
  `EvaluateStoreChannelPolicy`, which requires both OS grant and Store review
  approval.
- `.github/workflows/store-distribution-contract.yml` runs the contract gate.
