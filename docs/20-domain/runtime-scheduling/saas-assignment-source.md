# Runtime Scheduling SSOT: SaaS Assignment Source

[Back to runtime-scheduling.md](../runtime-scheduling.md)

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
변환해 `/heartbeat` 으로 전송하며, 기본 cadence 는 5 seconds 다. Control plane 은
active assignment heartbeat 가 20 seconds 동안 refresh 되지 않으면 stale 로
간주한다. heartbeat response 에서 requested active assignment 가 refresh 되지 않으면
daemon 은 그 server-side stale/cancel 판정을 인정하고 local provider run 에 취소
cause 를 전달한다. progress/result report 는 `/v1/agents/{agent_id}/events` 로
전송한다.

SaaS assignment FSM 은 `ready -> running -> terminal` 순서를 요구한다. Provider
adapter 가 별도 running lifecycle event 를 내지 않더라도 supervisor 는 provider
process submit 이 성공한 직후 `assignment_running` report 를 보장해야 한다. Terminal
result 를 `ready` 상태에서 바로 보고하면 control-plane 이 `ready -> completed`
전이를 거부할 수 있고, 그 assignment 가 다시 lease 되어 같은 provider run 이 반복될
수 있다.

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
