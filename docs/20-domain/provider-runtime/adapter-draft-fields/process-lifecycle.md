# Process Lifecycle

[Back to adapter-draft-fields.md](../adapter-draft-fields.md)

| Step | Adapter responsibility |
| --- | --- |
| spawn | On `RunHandle.start`, start the provider CLI or app-server and hold exit code, stderr, and stdout handles. |
| observe | Normalize stdout / JSON-RPC notifications into `ProviderEventDraft`. |
| interrupt | `Interrupt(ctx, handle)` stops the current stream, for example Claude SIGINT or Codex `turn/interrupt`; draft emission may continue with interrupted signal. |
| stop | `Cancel(ctx, handle)` terminates the process, drains remaining stdout, then closes. |

If the process dies, the adapter emits an observation such as
`ProviderEventDraft(Type=LogLine, level="fatal", text=...)`. It does not emit a
task transition draft such as `TaskFailed`; transition decisions belong to the
server orchestrator.
