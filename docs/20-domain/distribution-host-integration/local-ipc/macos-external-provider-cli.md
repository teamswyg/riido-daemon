# macOS External Provider CLI Strategy

[Back to local-ipc.md](../local-ipc.md)

`Q-DIST-001` is resolved here. The Mac App Store target treats Claude / Codex /
OpenClaw / Cursor CLIs as external user-installed executables, never bundled
payloads. The Store App may help the user choose and verify a provider path, but
provider execution remains behind the local helper and C4 runtime boundary.

Policy snapshot: checked Apple App Review Guidelines and App Sandbox entitlement
documentation on 2026-05-28. The source links stay in
[`store-distribution.md`](../../../30-architecture/store-distribution.md) §7.
If Apple changes these rules, this section and the executable policy gate must
change in the same work unit.

Rules:

1. `mac-app-store` must not use a temporary exception entitlement as the default strategy for provider CLI execution.
2. Provider CLI path registration must start from a user action such as file picker / open panel, then reduce adapter-specific proof into C11 facts: `ExternalToolRecord{provenance=user-selected}` and `StoreChannelPolicyInput.OSGrantPresent=true`.
3. A sandbox/security-scoped/user-selected executable grant alone is not enough to execute a provider CLI in the Store channel. `StoreReviewApproved=true` is also required.
4. If either OS grant or Store review approval is missing, the provider may be shown as detected / login-required / store-blocked, but C4 must not spawn it.
5. Review/demo mode must still work without provider CLI installation.
6. The review note must explain: provider CLIs are external user-installed tools; Riido does not bundle, download, or silently install them; execution requires explicit `provider-execute:<provider>` consent; the local helper is local-only; workspace access is security-scoped; no root escalation, LaunchAgent install, standalone code download, or shared-location code install is used.
7. Provider executable paths, security-scoped bookmark bytes, and entitlement proof stay local to the Store App/helper adapter. C10 metadata may receive provider kind and routing status only.
