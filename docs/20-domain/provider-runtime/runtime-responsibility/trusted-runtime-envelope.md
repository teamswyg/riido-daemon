# Provider Trusted-Runtime Envelope

[Back to runtime-responsibility.md](../runtime-responsibility.md)

C4 provider command builders do not hide provider-native work permissions or
delegate them to provider defaults. If a provider must operate on repo,
toolchain, or workspace surfaces on the user's PC, the adapter must create an
explicit trusted-runtime envelope.

Riido harness owns:

- immutable SaaS assignment snapshot
- daemon-selected workdir and evidence root
- provider process start / stop / cancel
- runtime slot, lease, fencing token, heartbeat, stale judgment
- dropped arg evidence and provider log/progress redaction
- real integration gate and filesystem side-effect verification

This does not mean "full-access is the default." C4 does not infer sandbox or
approval-bypass meaning from provider defaults, caller args, or SaaS payload.
The adapter-owned launch envelope and harness responsibilities are fixed together.

| Provider | Current C4 trusted-runtime envelope status |
| --- | --- |
| Codex | Adopted. Adapter creates only `codex --sandbox danger-full-access app-server --listen stdio://`; caller `--sandbox`, config override, and unsafe bypass args are dropped with evidence. |
| Claude | Not adopted. `PermissionMode` is explicit input; `bypassPermissions` requires C7 unsafe-bypass gate approval in an isolated tier. |
| Cursor | Not adopted. `--trust` acknowledges daemon-selected workdir; `--yolo` remains a C7 unsafe-bypass gate target. |
| OpenClaw | Not adopted. Worktree-required tasks must be blocked by C5 because `supports_worktree=false`. |

Promoting another provider to a Codex-like trusted/full-access runtime requires
one PR to update this table, C7 security decisions, command builder,
deterministic tests, and real integration evidence.
