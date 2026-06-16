# Agent Execution Unresolved Design: Overview

[Back to agent-execution-unresolved-design.md](../agent-execution-unresolved-design.md)


> Riido task: RIID-4964 `에이전트 프로필을 업로드 하기 위한 endpoint 필요`
>
> This document analyzes the unresolved AI Agent execution issues from the
> daemon-side architecture point of view. It is intentionally structural: the
> goal is to fix the shared model first, then split implementation PRs across
> `riido-contracts`, `riido-control-plane`, `riido-daemon`, `riido-infra`, and
> client/desktop surfaces.

## 0. 결론

아래 미해결 항목들은 서로 다른 버그처럼 보이지만, 구조적으로는 같은 문제에서
나온다.

1. daemon 실행 단위가 `task_id` 중심이다. SaaS assignment 는 이미
   `assignment_id` 를 갖지만, runtime/supervisor/heartbeat/cancel watcher 는 여전히
   `task_id` 를 in-flight key 로 쓴다.
2. control-plane 의 assignment snapshot 이 "무엇을 어디에서 실행해야 하는가"를
   구조화해서 넘기지 못하고, repo/branch/task context 가 prompt 문자열 안에 묻혀
   있다.
3. stop, stream, retry, approval, resume 이 하나의 lifecycle/FSM 위에 있지 않고
   각 레이어의 부수 효과로 흩어져 있다.
4. provider 실행 환경이 "감지된 executable"까지만 고정되고, PATH/toolchain/env/cwd
   전체가 launch envelope 로 고정되지 않는다.

따라서 1차 설계 방향은 다음과 같다.

- `assignment_id` 기반 `ExecutionIdentity` 를 shared contract 로 승격한다.
- assignment snapshot 에 `WorkspacePlan` 과 최소 `RuntimeLaunchPolicy` 를 추가한다.
- daemon 은 `ExecutionIdentity` 로 in-flight/run/watcher/heartbeat 를 관리하고,
  `task_id` 는 UI grouping 과 read-model 조회용으로만 쓴다.
- stop/progress/final/retry/resume 을 generated FSM 과 event envelope 위로 올린다.
- private repo/auth/token 은 public contract 에 값으로 들어가지 않고, token reference
  또는 unsupported/fail-closed 상태만 공개 가능한 vocabulary 로 표현한다.

## 1. 현재 구조 증거

| 관찰 | 현재 코드 / SSOT | 구조적 의미 |
| --- | --- | --- |
| SaaS assignment id 는 execution id 로 우선 사용되고 logical `task_id` 는 metadata 로 보존된다. | `internal/agentbridge/controlplane/saasplane/runtime_id.go` `assignmentExecutionID`, `taskRequestFromAssignment` | 같은 task 에 여러 assignment 가 붙어도 daemon 실행 key 가 충돌하지 않는다. |
| `saasplane` state 는 `assignmentsByExecution`, `runtimeIDsByExecution`, `cancelWatchers`, `partialBodies` 를 execution id 로 저장한다. | `state_loop.go` `planeState`, `saveAssignmentRuntime`, `WatchCancellation`, `CompleteTask` | watcher/runtime/partial-body cleanup 은 terminal path 에서 assignment 단위로 수렴한다. |
| supervisor `inFlight` map 과 duplicate guard 는 `TaskRequest.ID` 기준이며, SaaS adapter 는 그 값을 execution id 로 채운다. | `internal/agentbridge/supervisor/supervisor.go` `claimOne`, `saasplane/state_loop.go` `taskRequestFromAssignment` | agent 교체/동시 assignment/retry attempt 는 logical task metadata 와 분리된 실행 단위로 모델링된다. |
| runtimeactor heartbeat 는 `RunningTaskIDs` 를 유지하지만 daemon adapter 가 execution id 를 assignment id 로 매핑한다. | `saasplane/http_client.go` `activeAssignmentIDsForHeartbeat` | heartbeat 는 task id 역변환 없이 active assignment refresh 요청을 만든다. |
| workspace SSOT 는 per-run workdir, repo cache 분리, `run_id` 를 이미 정의한다. | `docs/20-domain/workspace.md` | 도메인 방향은 맞지만 SaaS assignment DTO 가 repo/workspace plan 을 구조화하지 않아 구현이 prompt 에 기대게 된다. |
| provider spawn 은 `detectutil` 의 augmented/frozen launch PATH 를 `StartRequest.Env` 와 process env 에 동일하게 주입한다. | `runtimeactor/actor_submit.go`, `detectutil.EnvMapWithLaunchPATH`, `detectutil.EnvListWithLaunchPATHFromMap` | GUI/launchd 최소 PATH 에서도 provider child tool 탐색 surface 가 detect/launch 사이에 갈라지지 않는다. |
| C7 security 는 approval decision vocabulary 를 갖고 있으나 headless web approval handoff 는 아직 runtime lifecycle 과 연결되지 않았다. | `docs/20-domain/security.md` §6 | "승인 필요"가 provider wait 상태인지, task needs-input 인지, fail-closed 인지 레이어마다 해석될 수 있다. |

## 2. 미해결 항목별 구조 원인

| ID | 증상 | 구조 원인 | 설계 방향 |
| --- | --- | --- | --- |
| F3 | AI 가 실제 코드가 아닌 빈 temp folder 에서 작업 | repo/branch 가 prompt text 이고 daemon 의 `WorkspacePlan` 이 아님 | control-plane 이 assignment snapshot 에 repo ref 를 구조화하고 daemon 이 public repo clone/worktree 를 materialize |
| C1 | stop 이 일관된 상태가 아님 | stop intent, process kill, projection terminal state 가 분리됨 | assignment lifecycle FSM 에 `stop_requested`, `provider_cancel_requested`, terminal states 추가 |
| S2/S4/S5 | server-side streaming cleanup 미완 | text delta/progress/final answer 가 같은 progress path 를 공유 | `StreamEnvelope` 를 `progress_event`, `answer_delta`, `final_answer` 로 분리 |
| F4/F5 | provider child tools 가 `git`/`node` 를 못 찾음 | **Resolved for launch PATH and submit-time re-detect.** Detection uses augmented search dirs, launch PATH is frozen into adapter/spawn env, and unavailable capability is refreshed after TTL. | detectutil/runtimeactor PATH and TTL refresh 회귀 테스트 유지 |
| F6 | headless tool approvals blocked | approval policy 가 C7 decision 으로는 있으나 SaaS approval round-trip 이 없음 | safe auto-approve allowlist + web approval request/decision event 추가 |
| F7 | transient network error 재시도 없음 | **Resolved for SaaS safe/idempotent JSON transport.** Event POST 는 idempotency key 가 없어 의도적으로 재시도하지 않는다. | `poll`, `heartbeat`, `agent-bindings`, `runtime-snapshot` transient retry + permanent 4xx no-retry 회귀 테스트 유지 |
| R4 | daemon restart 후 처음부터 반복 | provider session id / run attempt / recovery policy 가 durable assignment 와 분리 | assignment run table 에 `provider_session_id`, `attempt`, `recovery_mode` 기록 |
| R5 | cancellation watcher leak | watcher 가 task id map 에 저장되고 terminal cleanup 에서 close 되지 않음 | `ExecutionIdentity` keyed watcher registry + terminal close/release invariant |
| D5 | desktop stop 이 stale PID 를 kill 할 수 있음 | **Resolved.** PID fallback 은 sidecar identity, socket, foreground daemon command line 검증이 없으면 fail-closed 한다. | PID identity probe 회귀 테스트 유지 |
| D7 | Windows `.claim` stale lock | **Resolved.** Windows claim file 이 owner/refreshed_at metadata 를 쓰고 active holder 가 주기적으로 갱신한다. 오래된 legacy/metadata claim 만 회수한다. | claim metadata stale recovery 테스트 + Windows compile check 유지 |
| F8 | sync preparation 이 poll/heartbeat path 를 막을 수 있음 | **Resolved for supervisor claim/heartbeat loop.** Workspace materialization/provider submit 은 in-flight 등록 뒤 activation goroutine 에서 수행되고 결과만 actor mailbox 로 돌아온다. | slow worktree materialization 중 heartbeat 지속 회귀 테스트 유지 |
| R1 | provider process per task cold start | one-shot run 만 모델링되고 conversational session 과 resume 정책이 없음 | provider session table 과 long-lived process policy 를 분리 |
| Review | PID kill, cwd mismatch, multi-agent conflict, stale busy, SSE refetch dependency | runtime key, lifecycle, read-model update authority 가 레이어마다 다름 | assignment identity + FSM + client optimistic upsert + stale assignment reconciliation |

## 3. Target Domain Model

### 3.1 ExecutionIdentity

`ExecutionIdentity` 는 `riido-contracts` 로 승격해야 하는 shared vocabulary 다.
daemon 내부에서는 task id 대신 이 값을 실행 key 로 사용한다.

| Field | 의미 |
| --- | --- |
| `assignment_id` | SaaS assignment 의 primary execution id. in-flight, watcher, heartbeat, event report key |
| `task_id` | client/read-model grouping id. 여러 assignment 가 같은 task 를 가질 수 있음 |
| `component_id` | workspace/task thread projection scope |
| `agent_id` | assigned agent profile/runtime binding id |
| `runtime_id` | daemon runtime actor id |
| `run_id` | local workdir/IR run id. 기본값은 `assignment_id`; retry attempt 는 suffix 또는 attempt field 로 구분 |
| `attempt` | same assignment retry attempt number. process retry 와 user reassign 을 구분 |
| `provider_session_id` | provider native resume id. provider 가 보고한 뒤 durable event 로 저장 |

규칙:

1. daemon `inFlight`, runtimeactor session table, cancellation watcher,
   partial streaming buffer 는 `assignment_id` 또는 `run_id` 로 keying 한다.
2. heartbeat 는 `active_assignment_ids` 를 직접 보고한다. `running_task_ids` 는
   backward compatibility field 로만 남긴다.
3. `task_id` 로 runtime cancel 을 호출하는 path 는 compatibility adapter 에서만
   허용하고, 내부 actor 경계는 `execution_id` 를 사용한다.
4. user-facing task thread 는 task id 를 유지한다. terminal/active projection 은
   assignment lifecycle state 를 보고 결정한다.

### 3.2 WorkspacePlan

F3 를 해결하려면 "repo/branch 를 prompt 로 알려준다"가 아니라 "daemon 이 어디를
clone/materialize 해야 하는지"가 assignment snapshot 에 있어야 한다.

| Field | Phase | 설명 |
| --- | --- | --- |
| `source_kind` | P0 | `empty`, `git_public`, `git_private_unsupported`, `git_private_token_ref` |
| `repo_url` | P0 public only | public clone URL. private repo URL 은 public log/RAG 에 raw 로 남기지 않는다. |
| `repo_full_name` | P0 | 표시/진단용 stable id. private 이면 visibility 와 함께 최소화 |
| `branch_name` | P0 | Riido task branchName 또는 selected branch |
| `commit_sha` | P1 | 있으면 exact checkout. 없으면 branch head fetch |
| `visibility` | P0 | `public` / `private` / `unknown` |
| `auth_mode` | P0 | `none` / `unsupported` / `token_ref` |
| `auth_ref` | P2 private | token 값이 아니라 control-plane/daemon secret broker reference |
| `isolation_mode` | P0 | `git_worktree` / `shallow_clone` / `empty_explicit` |
| `required_surfaces` | P0 | scheduler 에 전달될 worktree/session/tool surface |

Phase P0 에서는 public repo 만 clone 한다. private repo 는 `git_private_unsupported`
로 fail-closed 하며, provider prompt 나 public event 에 token, signed URL, raw
credential 을 넣지 않는다.

### 3.3 RuntimeLaunchEnvelope

provider executable detection 은 runtime capability 의 일부이고, provider process
spawn 은 별도 launch envelope 이다. 둘을 분리하면 "감지는 됐지만 child tool 이
없다" 문제를 줄일 수 있다.

| Field | Owner | 설명 |
| --- | --- | --- |
| `selected_executable` | daemon C4 | adapter detect 가 고른 executable absolute path |
| `path` | daemon C4 | login-shell PATH + well-known install dirs + policy env 를 합친 frozen PATH |
| `env` | daemon C4/C7 | child process 에 명시적으로 넘기는 sanitized env |
| `cwd` | daemon C6 | prepared workdir |
| `toolchain_probe` | daemon C4 | `git`, `node`, package manager 등 provider child tool 가용성 |
| `detected_at` / `ttl` | daemon C4 | cached detection freshness |
| `approval_policy` | daemon C7 + contracts | auto-approve surface, approval-required surface, fail-closed surface |

`RuntimeLaunchEnvelope` 전체를 곧바로 contracts 로 올릴 필요는 없다. 다만
`selected_executable`, `provider_version`, `toolchain_probe`, `approval_policy_id`
처럼 control-plane/client 가 읽어야 하는 field 는 shared DTO 로 승격한다.
