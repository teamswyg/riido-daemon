# Store Distribution Architecture: Decisions

[Back to architecture](../architecture.md)

This section owns the high-level product split for store distribution.

Non-goals:

- C11 domain decisions remain in
  [`docs/20-domain/distribution-host-integration.md`](../../../20-domain/distribution-host-integration.md).
- Security policy remains in
  [`docs/20-domain/security.md`](../../../20-domain/security.md).
- Provider capability shared contracts remain in public `riido-contracts`.
- SaaS control-plane and review account details remain in public
  `riido-control-plane`.

Decisions:

1. Provider CLIs are not included in package artifacts. Claude, Codex, OpenClaw,
   and Cursor Agent are user-installed external tools.
2. Mac uses Developer ID notarized distribution as the first target. Mac App
   Store remains a separate constrained target until sandbox/helper/workspace
   grant mode is ready.
3. Windows uses MSIX sideload as the first target. Microsoft Store remains a
   separate target until packaged desktop app, full-trust, and background policy
   evidence are ready.
4. Store App and Local Helper are separated by role. UI owns consent, provider
   status, and workspace grant surfaces. The helper owns local-only IPC and task
   execution orchestration.
5. Store review mode is required. Reviewers must be able to inspect onboarding,
   consent, status, and privacy flows without provider CLIs.
