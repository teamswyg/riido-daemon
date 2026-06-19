# Assignment Lifecycle FSM

[Back to agent-execution-unresolved-design.md](../agent-execution-unresolved-design.md)

Assignment lifecycle is generated from the `riido-contracts` Common Lisp FSM
source. Daemon consumes the generated Go enum/SPI and this page keeps the
reader-facing state diagram close to the executable evidence manifest.

```mermaid
stateDiagram-v2
    [*] --> queued
    queued --> leased: poll_start
    leased --> preparing_workspace: daemon_ready
    preparing_workspace --> running: workspace_prepared
    running --> stop_requested: user_stop
    stop_requested --> provider_cancel_requested: daemon_cancel_sent
    provider_cancel_requested --> cancelled: provider_cancelled
    running --> waiting_approval: approval_required
    waiting_approval --> running: approval_granted
    waiting_approval --> blocked: approval_denied_or_timeout
    running --> completed: final_answer
    running --> failed: provider_failed
    preparing_workspace --> failed: workspace_failed
    leased --> failed: lease_expired
    completed --> [*]
    cancelled --> [*]
    failed --> [*]
    blocked --> [*]
```

FSM metadata:

- start states: `queued`, `leased`
- terminal states: `completed`, `cancelled`, `failed`, `blocked`
- retryable states: `leased`, `preparing_workspace`, transient transport failure
- non-retryable states: policy blocked, approval denied, private repo unsupported
- user-visible active states: `queued`, `leased`, `preparing_workspace`, `running`,
  `waiting_approval`, `stop_requested`, `provider_cancel_requested`

Executable evidence manifest:
[`assignment-lifecycle-evidence.riido.json`](assignment-lifecycle-evidence.riido.json).

Related sections:

- [Stream envelope](stream-envelope.md)
- [Retry and recovery policy](retry-recovery-policy.md)
- [Repo ownership](repo-ownership.md)
- [Implementation slices](implementation-slices.md)
- [Verification evidence](verification-evidence.md)
- [Current daemon slice status](current-daemon-slice-status.md)
