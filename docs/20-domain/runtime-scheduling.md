# Runtime Scheduling SSOT

> **이 문서가 C5 Runtime Scheduling 의 SSOT다.**
>
> - 책임: 어떤 runtime 이 어떤 task 를 claim / execute 할 수 있는가, runtime capability 와 task 요구 surface 의 eligibility 판정, runtime lease pinning, heartbeat 의미.
> - 비책임: provider 가 무엇을 할 수 있는가의 정적 모델은 public
>   [`riido-contracts`](https://github.com/teamswyg/riido-contracts) 의 C3
>   계약이 소유한다. provider process 실행과 provider session table schema /
>   retention / adapter 는 C4 daemon migration slice, workdir 생성은 C6 daemon
>   migration slice, lock 획득 primitive 는 [`./locking.md`](./locking.md) (C9)가
>   소유한다.

이 SSOT 는 **C5 Runtime Scheduling** context 를 채운다.

## 0. 핵심 invariant

1. **scheduler 는 capability 로만 분기한다.** provider binary version 문자열로 task dispatch 를 결정하지 않는다. `DetectedVersion` 은 fingerprint raw signal 일 뿐이다.
2. **provider-specific FSM 은 없다.** scheduling 은 task 요구 surface 와 runtime capability 의 boolean / compatibility envelope 를 비교한다. 실행 상태는 C1/C2 IR FSM 이 소유한다.
3. **provider process 는 eligibility 통과 후에만 spawn 된다.** claim 된 task 가 요구 surface 를 만족하지 못하면 pre-submit 단계에서 `blocked` result 로 보고하고 process 를 시작하지 않는다.
4. **experimental runtime 은 명시 opt-in 이 필요하다.** `RequiresExperimentalOptIn=true` runtime 은 task 가 `allow_experimental_runtime=true` 를 명시해야 local daemon scheduler 가 실행할 수 있다.
5. **lease pin 은 `(RuntimeID, CapabilityFingerprint)` 쌍이다.** 같은 runtime id 라도 fingerprint 가 바뀌면 기존 lease 는 stale 이며 무효화된다.
6. **local file queue 는 영속 scheduler 가 아니다.** file queue 는 runtime registry 의 `provider.<name>.available` capability 로 provider mismatch task 를 claim 전에 건너뛴다. claim 된 파일은 top-level task 파일을 원자적으로 `claims/` receipt 로 이동하므로 surface / policy ineligible task 를 “Queued 유지” 로 되돌릴 수 없다. 대신 reporter 에 `blocked` result 를 남긴다. DB/API 기반 production source 는 같은 판정을 task state `Blocked` 또는 `Queued 유지` 로 표현할 수 있다.
7. **first-class local production source 는 `riido-task-db.v1` 이다.** `RIIDO_TASK_DB_SOURCE_PATH` 를 설정한 daemon 은 `Queued` row 만 claim 하며, 모든 claim/progress/result 는 C1 guarded mutation 으로 기록한다. provider 가 `completed` 를 보고해도 task 는 `Completed` 로 직접 전이하지 않고 `Validating` 에 머문다.
8. **C5 does not own provider session table.** C5 는 lease metadata 에 `RuntimeID`, `CapabilityFingerprint`, fencing token 을 보존하고, C4 가 소유한 provider session table 을 필요할 때 참조할 수 있다. 하지만 `riido-provider-session-table.v1` schema / retention / adapter 와 provider-native resume semantics 는 C4 Provider Runtime 이 소유한다.
9. **C5 does not own client task-thread read models.** Scheduling preserves task/run/thread identifiers supplied by the SaaS source so progress can be reported back, but `GET /v1/client/ai-agent/tasks/{task_id}/threads`, `active_stream` HATEOAS selection, and historical thread collection semantics are control-plane/client API facts. The same boundary applies to Figma `node-id=153-8761`: a busy-agent queued row is SaaS assignment state (`queued_by_busy_agent`/`queued`) and client presentation copy, not a daemon-generated comment. It also applies to Figma `node-id=227-19354`: the stopped row after agent deletion is SaaS delete/read-model state (`stopped_by_agent_deleted`/`stopped`) plus client presentation copy. The daemon only observes SaaS cancellation/stop instructions for an assigned runtime, applies them to the provider process, and reports progress/result through existing ports.

## 1. Task requirement model

task 는 provider-neutral surface 이름으로 요구 조건을 표현한다.

| Surface | 의미 | capability flag |
| --- | --- | --- |
| `structured-event-stream` | stdout/RPC 에서 구조화된 event stream 을 받을 수 있어야 함 | `SupportsStructuredEventStream` |
| `session-resume` | session/thread resume 이 가능해야 함 | `SupportsResume` |
| `system-prompt` | system/developer instruction 을 native surface 로 전달할 수 있어야 함 | `SupportsSystemPrompt` |
| `max-turns` | turn limit 을 native surface 로 전달할 수 있어야 함 | `SupportsMaxTurns` |
| `mcp` | MCP config / tool bridge 를 지원해야 함 | `SupportsMCP` |
| `tool-hooks` | tool / hook event surface 를 지원해야 함 | `SupportsHookEvents` |
| `usage` | token usage metric 을 제공해야 함 | `SupportsUsageMetrics` |

Go surface 는 `bridge.TaskRequest.RequiredSurfaces []string` 이다. local file queue JSON 에서는 `required_surfaces` 로 쓸 수 있다. 예:

```json
{
  "id": "task-1",
  "provider": "cursor",
  "prompt": "inspect this repo",
  "required_surfaces": ["structured-event-stream"],
  "allow_experimental_runtime": true,
  "metadata": {
    "workspace_id": "ws-1"
  }
}
```

알 수 없는 surface 이름은 “지원한다고 추정” 하지 않는다. eligibility 는 실패한다.

## 2. Runtime eligibility

C5 의 local eligibility 판정 입력:

| 입력 | 출처 |
| --- | --- |
| `Provider` | `TaskRequest.Provider` |
| `RequiredSurfaces` | `TaskRequest.RequiredSurfaces` (+ legacy metadata fallback 가능) |
| `AllowExperimentalRuntime` | `TaskRequest.AllowExperimentalRuntime` |
| `RuntimeID` | runtime status |
| `CapabilityFingerprint` | C3 capability reconciliation |
| `SlotLimit` / `SlotsInUse` | runtime status / heartbeat |
| `Available` / `CompatibilityStatus` / `RequiresExperimentalOptIn` | C3 capability reconciliation |
| provider-neutral support flags | C3 capability reconciliation |

판정 순서:

1. task provider 와 runtime capability provider 가 일치해야 한다.
2. `Available=false` 이면 거절.
3. `CompatibilityStatus=blocked` 이면 거절.
4. `RequiresExperimentalOptIn=true` 이고 task 가 opt-in 하지 않았으면 거절.
5. `SlotLimit > 0` 이고 `SlotsInUse >= SlotLimit` 이면 거절.
6. 모든 `RequiredSurfaces` 의 capability flag 가 true 여야 한다.

## 3. Local daemon implementation contract

이 섹션은 C5 가 다른 daemon runtime package 와 맺는 계약을 설명한다.
RIID-4646 에서는 `internal/scheduling` 순수 domain package 만 public repo 로
이동했다. RIID-4656 은 `runtimeactor` 를, RIID-4662 는 `supervisor` 를 public
repo 로 이동했다. RIID-4683 은 local `riido-task-db.v1` persistence package 와
`controlplane/taskdbplane` adapter 를 public repo 로 이동했다. RIID-4689 는
SaaS assignment source/reporter adapter 인 `controlplane/saasplane` 을 public repo 로
이동했다. RIID-4690 은 이 public runtime/source/reporter 조각들을
`riido daemon ...` process lifecycle CLI 로 연결했다.

현재 local daemon process 는 provider capability boundary 별 RuntimeActor pool 을 가진다. 기본 daemon 은 `claude`, `codex`, `openclaw`, `cursor` adapter 를 각각 별도 RuntimeActor 로 시작하고, SupervisorActor 가 각 runtime 을 control plane 에 등록한 뒤 runtime id 별로 claim / heartbeat / cancellation 을 dispatch 한다. 같은 process 안에서도 “pool 에서 선택된 runtime 이 claim 하고 실행한다”가 기본 의미다. 여러 daemon 이 같은 task DB 를 source 로 공유하는 경우에도 task DB adapter 는 persisted runtime registry 를 pool snapshot 으로 보고, deterministic selector 가 고른 runtime id 에서만 claim 이 성공한다.

구현 위치:

| 코드 | 역할 |
| --- | --- |
| `internal/scheduling` | C5 domain model: `RuntimeLease`, required surface enum, eligibility evaluator |
| `internal/agentbridge/supervisor` | local daemon claim loop. RuntimeActor pool 을 등록/heartbeat 하고, selected runtime status 를 C5 evaluator 입력으로 변환한 뒤 process spawn 전에 gate 적용. public 구현은 RIID-4662 에서 이동됨 |
| `internal/agentbridge/runtimeactor` | C4 실행 actor. production daemon 에서는 provider별 actor 로 구동된다. provider availability 방어를 유지하지만 scheduling 결정을 소유하지 않음. public 구현은 RIID-4656 에서 이동됨 |
| `internal/taskdb` | local `riido-task-db.v1` JSON schema, guarded mutation, command-id idempotent replay, evidence receipt persistence. public 구현은 RIID-4683 에서 이동됨 |
| `internal/agentbridge/controlplane/taskdbplane` | `riido-task-db.v1` first-class source/reporter adapter. `Queued → Claimed → Preparing → Running → Validating/terminal` 을 guarded mutation 으로 기록. runtime registry / lease sidecar / fencing token 검증을 소유한다. public 구현은 RIID-4683 에서 이동됨 |
| `internal/agentbridge/controlplane/saasplane` | SaaS assignment first-class source/reporter adapter. `riido-contracts/assignment` DTO 로 poll/start/cancel/heartbeat/event sync 를 수행하며, claim/runtime scheduling decision 은 supervisor/runtimeactor/controlplane port 경계 뒤에 둔다. public 구현은 RIID-4689 에서 이동됨 |
| `cmd/riido daemon ...` | 12-factor env 와 CLI flags 를 읽어 RuntimeActor pool, SupervisorActor, local Unix socket status surface, and configured TaskSourcePort/TaskReporterPort 를 연결하는 adapter. public 구현은 RIID-4690 에서 이동됨 |

순수 pool selector(`internal/scheduling.SelectRuntime`) 는 여러 runtime capability snapshot 을 같은 evaluator 로 평가하고, eligible 후보 중 deterministic 하게 하나를 고른다. 우선순위는 capability summary(`supported` → `degraded` → `experimental`) 다음 slot headroom, `RuntimeID`, `CapabilityFingerprint` 순이다. 이 함수는 lease 를 획득하지 않고 provider process 를 시작하지 않는다. `riido-task-db.v1` adapter 는 선택 결과의 `(RuntimeID, CapabilityFingerprint)` 로 C9 local file lock + lease sidecar 를 잡은 뒤 claim 한다.

### 3.1 Local file queue shared-runtime behavior

여러 daemon 이 같은 `RIIDO_TASK_QUEUE_DIR` 를 poll 할 때 `FileQueueSource` 는 runtime registry(`runtimes/*.json`) 를 읽어 현재 `runtimeID` 의 `provider.<task.provider>.available` 이 false 이거나 registry 에 없는 provider task 를 claim 하지 않는다. 해당 task 파일은 top-level queue 에 남아 다른 runtime 이 claim 할 수 있다.

이 필터는 provider availability 에만 한정된다. required surface, experimental opt-in, compatibility status 같은 세부 eligibility 는 supervisor 의 C5 evaluator 가 최종 판정한다.

### 3.2 Local task DB production source

`RIIDO_TASK_DB_SOURCE_PATH` 는 `riido-task-db.v1` 파일을 local daemon 의 first-class production source 로 선택한다. source 는 다음 row 만 claim 할 수 있다.

1. `State=Queued`.
2. provider 가 task 또는 DB 추천 provider 에서 결정되고, provider candidate 가 있으면 `available=true`.
3. runtime registry 가 비어 있으면 backward-compatible single-daemon claim 으로 처리한다. registry 가 있으면 `provider.<provider>.available` key 가 있는 runtime 만 후보가 되며, task DB source 는 `internal/scheduling.SelectRuntime` 으로 선택된 runtime id 와 현재 runtime id 가 일치할 때만 claim 한다.
4. 같은 task 에 active foreign lease 가 없어야 한다. expired/released lease 는 C9 fencing token 을 증가시켜 재claim 할 수 있다.
5. prompt 로 쓸 `HarnessNextDirection` 또는 `Title` 이 존재.
6. human approval gate 가 있으면 이전 guarded receipt 의 `approval_id` 가 존재.

claim 은 `TaskClaimed`, start 는 `WorkdirPreparing`, provider running lifecycle 또는 완료 합성은 `RunStarted`, provider completed result 는 `RunReportedDone` 으로 기록한다. 이 마지막 전이의 target 은 `Validating` 이며, `Completed` 는 validation + approval 경로가 따로 충족될 때만 가능하다.

supervisor pre-submit eligibility 가 task 를 `blocked` 로 보고하면 task DB source 는 `BlockerRaised → Blocked` 로 보존한다. 이 경로는 local file queue 처럼 결과 sink 에만 남기지 않고 C1 task state 에 회복 가능한 block 을 남긴다.

task DB source 의 runtime registration / heartbeat 는 task DB schema 를 직접 늘리지 않고 같은 디렉터리의 sidecar registry 에 저장한다. `task-db.json` 의 기본 sidecar 경로는 `task-db.runtimes.json` 이며 schema version 은 `riido-runtime-registry.v1` 이다. registry 는 `schema_version`, `task_db_path`, `updated_at`, `runtimes[]` 를 가진다. `runtimes[]` 의 각 row 는 `controlplane.RegisteredRuntime` shape 이므로 runtime identity, provider capability flags, capability attributes, heartbeat slot state, running task ids, `last_heartbeat` 를 포함한다.

task DB source 의 claim lease 는 `task-db.leases.json` sidecar 에 저장한다. schema version 은 `riido-runtime-lease-registry.v1` 이며, `RuntimeID`, `CapabilityFingerprint`, `LeaseUntil`, `FencingToken` 을 기록한다. claim 에 성공한 `TaskRequest.Metadata` 는 `runtime_lease_id`, `runtime_fencing_token`, `runtime_capability_fingerprint` 를 포함한다. heartbeat 의 `running_task_ids` 는 same runtime + same capability fingerprint 의 active lease deadline 을 refresh 한다. task DB reporter 는 실제 progress/result transition 을 저장하기 전에 active lease 존재/만료와 fencing token value 를 확인한다.

claim 전에 task DB source 는 expired lease 를 조정한다. `Preparing` / `Running` task 는 `Blocked → Queued` 로 handoff 하며, 같은 claim loop 에서 새 selected runtime 이 다시 claim 할 수 있다. `Claimed` task 는 provider execution 전 준비 실패로 보고 `Failed` 로 정리한다. `NeedsInput` task 는 lease 만료로 더 이상 입력을 전달할 runtime 이 없으므로 `TimedOut` 으로 정리한다.

새 daemon 이 같은 task DB source 로 시작하면 sidecar registry 를 로드해 claim gating 에 재사용한다. 즉, daemon status socket 밖에서도 GUI/Zed 는 현재 runtime registry / heartbeat snapshot 을 읽을 수 있다. selector wiring / registry persistence / local claim-time lease primitive / active lease + fencing token enforcement / expired lease handoff / in-process multi-runtime dispatch 는 각각 분리된 책임으로 구현되어 있다.

### 3.3 SaaS assignment production source

`RIIDO_SAAS_URL` 을 선택한 daemon 은 local file queue 나 `riido-task-db.v1` source
대신 SaaS assignment API 를 source/reporter 로 사용한다. public daemon adapter 는
`controlplane/saasplane` 이며, shared contract surface 는
`github.com/teamswyg/riido-contracts/assignment` 가 소유한다.

SaaS mode 에서 runtime registration / heartbeat / claim loop 는 같은
TaskSourcePort 를 사용한다. `saasplane` 은 runtime id 에 포함된 agent/provider scope
로 `/v1/agents/{agent_id}/poll` 을 호출하고, `start` action 만
`bridge.TaskRequest` 로 변환한다. `cancel` action 은 in-flight task watcher 에
취소 cause 를 전달한다. heartbeat 는 local running task id 를 active assignment id 로
변환해 `/heartbeat` 으로 전송한다. progress/result report 는
`/v1/agents/{agent_id}/events` 로 전송한다.

이 adapter 는 remote assignment lease token 과 assignment id 를 request metadata 에
보존하지만, durable assignment store, reassignment blocker policy, SSE fan-out,
request authorization, metrics/health read model 은 소유하지 않는다.

## 4. RuntimeLease

`RuntimeLease` 는 다음을 묶는 도메인 값이다.

```
(LeaseID, TaskID, RuntimeID, CapabilityFingerprint, ClaimedAt, LeaseUntil, FencingToken)
```

규칙:

1. `LeaseUntil` 이 현재 시각보다 과거이면 expired.
2. `(RuntimeID, CapabilityFingerprint)` 중 하나라도 현재 runtime 과 다르면 stale.
3. stale lease holder 는 task 를 진행시킬 수 없다.
4. `FencingToken` 의 원자적 증가 / 비교는 C9 primitive 책임이다. C5 는 token 의미만 소유한다.

## 5. 인접 SSOT 와의 계약

| 인접 context | 본 문서가 받는 / 공급 |
| --- | --- |
| **C3 Provider Capability** | 받는다: `ProviderCapability`, `CompatibilityStatus`, `CapabilityFingerprint`, surface flags. |
| **C4 Provider Runtime** | 공급: eligibility 를 통과한 task 만 process spawn 으로 넘긴다. 받는다: runtime status / capability snapshot. C4 의 provider session table 은 provider-native resume identity 를 소유하며, C5 는 그 schema 를 재정의하지 않는다. |
| **C6 Workspace** | 받는다: workspace feasibility 신호. `WorkspacePrepared` 자체는 claim 사전조건이 아니다. |
| **C7 Security / Policy** | 받는다: experimental runtime opt-in / trust-tier 결정을 capability envelope 로 반영한 결과. |
| **C9 Locking / Lease** | 공급: lease 의미와 fencing token 요구. 받는다: 실제 lock / DB lease primitive. |

## 6. version-affecting changes

- 새 required surface 추가는 `change:additive` 이지만 C3 capability flag 또는 명확한 mapping 이 같은 PR 에 있어야 한다.
- eligibility 판정 순서 변경은 `change:behavioral`.
- lease pinning 규칙 변경은 `change:breaking-policy`.
