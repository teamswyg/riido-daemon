# RIID-4571 — macOS External Provider CLI Entitlement/Review Closure

[Back to macOS Provider CLI Review](../macos-provider-cli-review.md)

This slice closes `Q-DIST-001` by making the Mac App Store external Provider
CLI strategy executable:

- Claude / Codex / OpenClaw / Cursor CLIs remain external user-installed tools
  and are never bundled, downloaded, or silently installed by the Store App
- `mac-app-store` Provider CLI execution requires both an OS grant
  (`StoreChannelPolicyInput.OSGrantPresent=true`) and App Review approval
  (`StoreChannelPolicyInput.StoreReviewApproved=true`)
- when either proof is missing, the provider may be shown as detected /
  login-required / store-blocked, but C4 must not spawn it
- App Review notes must explain the external-tool execution surface, explicit
  provider-execute consent, security-scoped workspace access, local-only helper,
  provider non-bundling, and provider-free review/demo mode
- executable paths, bookmark bytes, entitlement proof, signing/provisioning
  secrets, and live submission evidence remain local/private and are not sent
  to C10 or checked into public repositories

The slice adds focused public CI for the `Q-DIST-001` closure and C7
store-channel policy test.
