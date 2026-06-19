# Idle Watchdog Semantic Activity

[Back to adapter-draft-fields.md](../adapter-draft-fields.md)

Idle watchdog resets only on events where the provider advances task meaning,
not on raw stdout byte activity. Public daemon implementation:
`internal/agentbridge.EventKind.IsSemanticActivity()`.

Semantic activity:

- `lifecycle`
- `text_delta`
- `thinking_delta`
- `tool_call_started`
- `tool_call_delta`
- `tool_call_completed`
- `tool_call_failed`
- `tool_approval_needed`
- `usage_delta`
- `progress`

`progress` is produced by the common Riido telemetry parser from
`<riido_log>{"code":...,"args":{...}}<end>`; legacy raw phrases are only
compatibility fallback mappings into code/args.

Non-semantic activity:

- `log`, `warning`, `error`
- `result`, `process_exit`
- `cancellation_requested`, `timeout`

stderr heartbeat, log spam, or process signals alone must not reset the idle
watchdog. Otherwise a provider can keep a run alive indefinitely without real
progress.
