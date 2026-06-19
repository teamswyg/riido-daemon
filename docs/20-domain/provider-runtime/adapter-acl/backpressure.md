# Draft And Session Backpressure

[Back to adapter-acl.md](../adapter-acl.md)

C4 Provider Runtime owns provider process stream, provider draft / session event channel, actor mailbox sizes, and drop policy. C6 workspace, C7 policy, and C10 server do not redefine these values. This closes `Q-RT-001`, legacy `Q-MULTICA-005`, and the runtime/session boundary portion of `Q-CTX-001`.

`internal/agentbridge/session` is a C4 internal submodel, not a separate bounded context. Claude/Codex/OpenClaw/Cursor session id differences remain concrete adapter ACL/protocol differences.

| Surface | Implementation constant | Value | Policy |
| --- | --- | --- | --- |
| process stdout chunk stream | `internal/process.DefaultStdoutBuffer` | `64` | no-drop, blocking process backpressure |
| process stderr chunk stream | `internal/process.DefaultStderrBuffer` | `64` | no-drop, blocking process backpressure |
| provider/session semantic event stream | `internal/agentbridge/session.DefaultEventBuffer` | `256` | no-drop, blocking backpressure |
| terminal result stream | `internal/agentbridge/session.DefaultResultBuffer` | `1` | exactly one terminal result |
| runtime actor mailbox | `internal/agentbridge/runtimeactor.DefaultMailboxSize` | `16` | caller-context bounded send |
| supervisor actor mailbox | `internal/agentbridge/supervisor.DefaultMailboxSize` | `64` | caller-context bounded send |

Rules:

1. stdout/stderr channels block when full; text/log/warning chunks are not dropped, overwritten, or reordered.
2. session actor blocks when the event buffer is full until the consumer drains `Events()`.
3. runtime actor and supervisor mailbox sends are bounded sends; caller context owns deadline / cancellation.
4. callers must drain `Events()` until close. Result-only callers must discard-drain.
5. C4 currently has no in-memory retry queue. Durable retry / outbox is a future C2/C10 decision.
6. changing buffer/mailbox values requires one work unit to update this doc, implementation constants, default-size tests, and the `provider-runtime-backpressure` workflow.
