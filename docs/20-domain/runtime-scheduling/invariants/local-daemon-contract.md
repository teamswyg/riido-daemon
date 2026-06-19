# Local Daemon Implementation Contract

[Back to invariants](../invariants.md)

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

## 3.1 Local file queue shared-runtime behavior

여러 daemon 이 같은 `RIIDO_TASK_QUEUE_DIR` 를 poll 할 때 `FileQueueSource` 는 runtime registry(`runtimes/*.json`) 를 읽어 현재 `runtimeID` 의 `provider.<task.provider>.available` 이 false 이거나 registry 에 없는 provider task 를 claim 하지 않는다. 해당 task 파일은 top-level queue 에 남아 다른 runtime 이 claim 할 수 있다.

이 필터는 provider availability 에만 한정된다. required surface, experimental opt-in, compatibility status 같은 세부 eligibility 는 supervisor 의 C5 evaluator 가 최종 판정한다.

## 3.2 Local task DB production source

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
