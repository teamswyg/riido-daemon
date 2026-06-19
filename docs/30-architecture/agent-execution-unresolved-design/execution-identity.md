# Execution Identity

[Back to Overview](overview.md)

`ExecutionIdentity` should be shared vocabulary in `riido-contracts`. Daemon
uses it as the execution key; task id remains a UI/read-model grouping key.

| Field | Meaning |
| --- | --- |
| `assignment_id` | primary execution id for in-flight, watcher, heartbeat, report |
| `task_id` | client/read-model grouping id |
| `component_id` | workspace/task thread projection scope |
| `agent_id` | assigned agent profile/runtime binding id |
| `runtime_id` | daemon runtime actor id |
| `run_id` | local workdir/IR run id; defaults to `assignment_id` |
| `attempt` | same-assignment retry attempt number |
| `provider_session_id` | provider-native resume id stored as durable event |

Rules:

- daemon run tables, watchers, and stream buffers key by `assignment_id` or `run_id`.
- heartbeat reports `active_assignment_ids` directly.
- task-id cancel is compatibility-only; actor boundaries consume execution id.
- user-facing task thread keeps task id and reads assignment lifecycle projection.
