# Retry And Recovery Policy

[Back to Assignment Lifecycle FSM](assignment-lifecycle-fsm.md)

Retry is not blind rerun. It asks which stage can recover under the same
assignment/run identity.

| Error class | Retry | Rule |
| --- | --- | --- |
| `transport_transient` | yes | timeout, reset, 502/503/504; idempotent request only |
| `transport_permanent` | no | 400/401/403/404 or contract violation |
| `workspace_prepare_transient` | maybe | network clone timeout or lock contention |
| `workspace_prepare_permanent` | no | private auth unsupported or branch not found |
| `provider_spawn_transient` | maybe | executable temporarily unavailable after TTL re-detect |
| `provider_policy_blocked` | no | explicit policy deny |
| `provider_session_lost` | no blind rerun | require `recovery_fresh_start_required` without resume id |
