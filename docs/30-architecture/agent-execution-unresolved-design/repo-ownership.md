# Repo Ownership

[Back to Assignment Lifecycle FSM](assignment-lifecycle-fsm.md)

| Repo | SSOT responsibility |
| --- | --- |
| `riido-contracts` | `ExecutionIdentity`, `WorkspacePlan`, FSM, stream envelope, approval DTO, generated enum/FSM SPI |
| `riido-control-plane` | task context to `WorkspacePlan`, assignment snapshot, scoped reconcile, stream coalescer, approval endpoint |
| `riido-daemon` | assignment-id in-flight model, workspace materializer, launch envelope/PATH, retry wrapper, watcher release |
| `riido-infra` | private auth broker/storage only when needed; no raw secret/evidence in public docs |
| client/desktop | optimistic cache upsert, SSE subscription, daemon PID probe, stale lock recovery, update/quit handoff |

If vocabulary is needed in two or more repos, prefer promotion to
`riido-contracts` before copying it into daemon-local code.
