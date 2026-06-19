# Scope and Core Invariants

[Back to invariants](../invariants.md)

> **이 문서가 C5 Runtime Scheduling 의 SSOT다.**
>
> - 책임: 어떤 runtime 이 어떤 task 를 claim / execute 할 수 있는가, runtime
>   capability 와 task 요구 surface 의 eligibility 판정, runtime lease pinning,
>   heartbeat 의미.
> - 비책임: provider 가 무엇을 할 수 있는가의 정적 모델은 public
>   [`riido-contracts`](https://github.com/teamswyg/riido-contracts) 의 C3
>   계약이 소유한다. provider process 실행과 provider session table schema /
>   retention / adapter 는 C4 daemon migration slice, workdir 생성은 C6 daemon
>   migration slice, lock 획득 primitive 는 [`../../locking.md`](../../locking.md)
>   (C9)가 소유한다.

이 SSOT 는 **C5 Runtime Scheduling** context 를 채운다.

## 0. 핵심 invariant

1. **scheduler 는 capability 로만 분기한다.** provider binary version 문자열로 task dispatch 를 결정하지 않는다. `DetectedVersion` 은 fingerprint raw signal 일 뿐이며, SaaS runtime snapshot 의 `provider_version` 으로 전달되더라도 표시/진단용 projection 이지 eligibility 입력이 아니다.
2. **provider-specific FSM 은 없다.** scheduling 은 task 요구 surface 와 runtime capability 의 boolean / compatibility envelope 를 비교한다. 실행 상태는 C1/C2 IR FSM 이 소유한다.
3. **provider process 는 eligibility 통과 후에만 spawn 된다.** claim 된 task 가 요구 surface 를 만족하지 못하면 pre-submit 단계에서 `blocked` result 로 보고하고 process 를 시작하지 않는다.
4. **experimental runtime 은 명시 opt-in 이 필요하다.** `RequiresExperimentalOptIn=true` runtime 은 task 가 `allow_experimental_runtime=true` 를 명시해야 local daemon scheduler 가 실행할 수 있다.
5. **lease pin 은 `(RuntimeID, CapabilityFingerprint)` 쌍이다.** 같은 runtime id 라도 fingerprint 가 바뀌면 기존 lease 는 stale 이며 무효화된다.
6. **local file queue 는 영속 scheduler 가 아니다.** file queue 는 runtime registry 의 `provider.<name>.available` capability 로 provider mismatch task 를 claim 전에 건너뛴다. claim 된 파일은 top-level task 파일을 원자적으로 `claims/` receipt 로 이동하므로 surface / policy ineligible task 를 “Queued 유지” 로 되돌릴 수 없다. 대신 reporter 에 `blocked` result 를 남긴다. DB/API 기반 production source 는 같은 판정을 task state `Blocked` 또는 `Queued 유지` 로 표현할 수 있다.
7. **first-class local production source 는 `riido-task-db.v1` 이다.** `RIIDO_TASK_DB_SOURCE_PATH` 를 설정한 daemon 은 `Queued` row 만 claim 하며, 모든 claim/progress/result 는 C1 guarded mutation 으로 기록한다. provider 가 `completed` 를 보고해도 task 는 `Completed` 로 직접 전이하지 않고 `Validating` 에 머문다.
8. **C5 does not own provider session table.** C5 는 lease metadata 에 `RuntimeID`, `CapabilityFingerprint`, fencing token 을 보존하고, C4 가 소유한 provider session table 을 필요할 때 참조할 수 있다. 하지만 `riido-provider-session-table.v1` schema / retention / adapter 와 provider-native resume semantics 는 C4 Provider Runtime 이 소유한다.
9. **C5 does not own client task-thread read models.** Scheduling preserves task/run/thread identifiers supplied by the SaaS source so progress can be reported back, but `GET /v1/client/ai-agent/tasks/{task_id}/threads`, `active_stream` HATEOAS selection, and historical thread collection semantics are control-plane/client API facts. The same boundary applies to Figma `node-id=153-8761`: a busy-agent queued row is SaaS assignment state (`queued_by_busy_agent`/`queued`) and client presentation copy, not a daemon-generated comment. It also applies to Figma `node-id=227-19354`: the stopped row after agent deletion is SaaS delete/read-model state (`stopped_by_agent_deleted`/`stopped`) plus client presentation copy. The daemon only observes SaaS cancellation/stop instructions for an assigned runtime, applies them to the provider process, and reports progress/result through existing ports.
