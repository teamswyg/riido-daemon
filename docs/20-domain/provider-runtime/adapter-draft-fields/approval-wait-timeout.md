# Approval Wait Timeout Ownership

[Back to adapter-draft-fields.md](../adapter-draft-fields.md)

`Q-RT-003` is closed here: C4 Provider Runtime / Adapter owns approval wait
timeout policy through the session actor's run clocks.

Rules:

1. provider adapters surface provider-native approval requests as `tool_approval_needed` / `ApprovalRequested`.
2. `tool_approval_needed` is semantic activity, so the first approval request resets `SemanticIdle`.
3. After approval request, the same C4 `SemanticIdle` clock expires the run if there is no provider progress, auto-approval response, human approval response, cancellation, or terminal provider result.
4. `HardTimeout` remains the whole-run upper bound and applies while waiting for approval.
5. `EventIngestor` appends observed draft events but does not own approval timers, expiry policy, or terminal timeout decisions.
6. UI / review surfaces may display pending approval and send a response, but they are not the source of truth for timing out the provider run.

When the C4 clock expires, the session actor emits `EventTimeout`; the reducer
turns it into `ResultTimeout` plus `CommandCancelProvider`, and the session
actor kills the provider process. Provider-native timeout/error observations
remain provider events, but Riido's provider-run timeout decision remains the
C4 session actor decision.
