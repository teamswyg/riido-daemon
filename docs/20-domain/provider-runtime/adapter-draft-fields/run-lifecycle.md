# Run Lifecycle

[Back to adapter-draft-fields.md](../adapter-draft-fields.md)

```text
RunStarted
   -> TextDelta / ReasoningDelta / ToolCallStarted / ToolCallFinished
   -> FileChanged / CommandStarted / CommandFinished / StatusUpdate
   -> UsageDelta / LogLine
   -> InputRequested -> ProvideInput -> continue
   -> ApprovalRequested -> ResolveApproval -> continue
   -> RunReportedDone
```

Rules:

1. Each stage may produce one or more `ProviderEventDraft` values.
2. `RunReportedDone` means only "the agent reported done"; task completion still depends on C8 validation.
3. Local RunController translates terminal provider `Result(completed)` into `RunReportedDone` transition event.
4. `Result(failed|blocked|aborted|cancelled|timeout)` does not set task state directly.
5. Local RunController translates terminal results into `TaskFailed`, `TaskCancelled`, or `TaskTimedOut` transition events and stamps `FSMVersion`.
6. Provider transport errors outrank provider self-reported completion.

Codex JSON-RPC pending request error responses, Codex `error` notifications
followed by empty `turn_completed` / process exit, and OpenClaw empty
`full_result` payloads fail closed as `Result(failed)`. If meaningful
`TextDelta` or non-empty `Result.Output` appears after an error notification,
the run may be treated as recovered.
