# Supervisor Boundary

[Back to runtime-actor-boundary.md](../runtime-actor-boundary.md)

`internal/agentbridge/supervisor` 는 Daemon tier RunController 다. Provider adapter 도
영속 scheduler 도 아니며, public daemon 안에서 이미 분리된 C4 RuntimeActor / C5
Scheduling / C6 Workdir / C2 EventIngestor / control-plane port 를 한 task run 단위로
조립한다.

Supervisor 의 책임:

- Start 시 RuntimeActor pool 을 control-plane source 에 등록하고 heartbeat 를 보낸다.
- runtime id 별로 task 를 claim 하고, duplicate in-flight task 를 방지한다.
- `internal/scheduling` eligibility evaluator 로 provider / surface / experimental-runtime opt-in 을 process spawn 전에 검증한다.
- workdir adapter 가 설정된 경우 task/run workspace 를 준비하고 native config 를 주입한다.
- provider event 와 terminal result 를 daemon-side `internal/ir/ingest` 에 draft 로 위임해 `CanonicalEvent` 로 append 하게 한다.
- terminal result 를 `TaskReporterPort` 로 보고하고, stop/cancel 시 in-flight run 을 cancelled 로 정리하며 archive 를 best-effort 로 남긴다.

Supervisor 의 비책임:

- `riido-task-db.v1` guarded mutation, local task DB lease sidecars, project/mwsd sync, local API, SaaS HTTP/SSE transport, or infra/state/secret ownership
- concrete provider parser/command/protocol implementation
- C1/C2/C3 schema ownership
- persistent lease registry / fencing-token primitive

Public `internal/taskdb` and `controlplane/taskdbplane` own task DB pieces.
Public `internal/project` owns project/mwsd projection sync, public
`internal/riidoapi` owns local API, and SaaS/infra remain separate repos or
adapters. C1/C2/C3 types are imported from public `riido-contracts`.
