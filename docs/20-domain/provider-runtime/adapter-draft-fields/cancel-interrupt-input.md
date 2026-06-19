# Cancel, Interrupt, And Input

[Back to adapter-draft-fields.md](../adapter-draft-fields.md)

| External signal | Adapter action |
| --- | --- |
| ingest calls `Cancel(handle)` | Provider process SIGTERM -> drain -> SIGKILL fallback. Remaining drafts drain before close. |
| local daemon stop / supervisor context cancel | RunController treats in-flight provider run as cancelled, requests process cancel, records `TaskCancelled` transition event and best-effort `WorkdirArchived`, then deregisters runtime. |
| ingest calls `Interrupt(handle)` | Provider-side interrupt message, such as Claude interrupt or Codex `turn/interrupt`; process remains alive. |
| ingest calls `ProvideInput(handle, response)` | Send response through provider stdin / RPC; resulting provider output flows as drafts. |
| ingest calls `ResolveApproval(approvalID, decision)` | Send provider approval protocol response, such as Codex app-server `approval/resolved`. |

Shutdown authority (`none` -> `graceful` -> `forced`) and default shutdown
timeouts are owned by `pkg/lifecycle`. RuntimeActor, SupervisorActor, and
`cmd/riido daemon` must consume that model instead of redefining local stop
level parsing or timeout policy.
