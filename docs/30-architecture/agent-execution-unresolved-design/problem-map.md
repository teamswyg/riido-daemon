# Problem Map

[Back to Overview](overview.md)

| ID | Symptom | Structural cause | Direction |
| --- | --- | --- | --- |
| F3 | AI works in empty temp folder | repo/branch was prompt text | public clone, private fail-closed, later token-ref broker |
| C1 | stop feels inconsistent | stop intent, process kill, projection state split | generated lifecycle states |
| S2/S4/S5 | stream cleanup is client-hidden | delta/progress/final share a path | `StreamEnvelope` split |
| F4/F5 | child tools missing | detection and launch env diverged | frozen launch PATH and TTL re-detect |
| F6 | headless approvals blocked | web approval absent | DTO plus fail-closed fallback |
| F7 | transient network failures skip retry | transport lacks retry taxonomy | safe/idempotent retry wrapper |
| R4 | restart can rerun from scratch | session id not durable enough | resume or explicit fresh-start refusal |
| R5 | watcher can leak | terminal release was not invariant | execution-id watcher cleanup |
| D5 | stale PID could kill wrong process | PID alone was trusted | identity probe before kill |
| D7 | Windows stale `.claim` can block | lock claim had no freshness metadata | owner/refreshed_at reclaim |
| F8 | workspace prep can block heartbeat | preparation ran on actor path | async activation goroutine |
| R1 | provider starts per task | one-shot process model | long-lived session policy later |
