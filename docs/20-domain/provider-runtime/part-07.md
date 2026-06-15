# Provider Runtime / Adapter SSOT: Part 07

[Back to provider-runtime.md](../provider-runtime.md)

### 7.7 RuntimeActor boundary

`internal/agentbridge/runtimeactor` 는 한 RuntimeID 의 provider execution capacity 를
소유하는 mailbox actor 다. actor goroutine 하나가 in-flight task map 과 slot state 를
단독으로 소유한다.

RuntimeActor 의 책임:

- Start 시 adapter `Detect` 를 실행하고 public C3 `ProviderCapability` 로 reconcile 한다.
- `PolicyBundleVersion` 과 detected executable fingerprint 를 capability fingerprint input
  에 포함한다.
- MaxConcurrent slot limit, duplicate task id, unavailable provider, unknown provider 를
  fail-closed 로 집행한다.
- `Submit` 을 `session.Start` 로 handoff 하고 optional `ProtocolDriverProvider` 를
  session 에 장착한다.
- `Cancel` / `Stop` 은 session cancel 과 process kill cascade 를 일으키고 slot 을 회수한다.
- `Status` / `HeartbeatPayload` 는 local settings UI 와 control-plane heartbeat 가 읽을
  수 있는 daemon-side runtime snapshot 을 만든다.

RuntimeActor 의 비책임:

- supervisor polling / task claim / runtime selection
- EventIngestor append / task transition 결정
- workdir preparation / native config injection
- provider-specific parser / adapter implementation
- task DB / project / mwsd / local API persistence

### 7.8 ControlPlanePort boundary

`internal/agentbridge/controlplane` 은 daemon supervisor 와 실제 task source/reporter
adapter 사이의 provider-neutral port 계약이다. 이 package 는 "어디에서 task 를
가져오는가" 와 "어디로 결과를 보고하는가" 를 interface 와 local black-box adapter 로
표현하지만, 어떤 runtime 이 선택되는지나 어떤 원격 프로토콜을 쓰는지는 결정하지 않는다.

ControlPlane root package 의 책임:

- `TaskSourcePort`: runtime registration, deregistration, heartbeat, task claim,
  cancellation watch port.
- `TaskReporterPort`: task start, normalized event, terminal result reporting port.
- claim-time lease metadata 를 reporter 호출 context 로 전달하는
  `TaskReportContext` helper.
- `MemorySource` / `MemoryReporter`: tests 와 offline mode 용 RAM-only port
  implementation.
- `FileQueueSource`: top-level JSON task file 을 atomically claim 하고 claim receipt /
  runtime registry record 를 남기는 local queue implementation.
- `FileReporter`: task-scoped JSONL report record writer.

ControlPlane root package 의 비책임:

- supervisor polling loop, runtime selection, slot scheduling
- `runtimeactor` session handoff / process execution
- `controlplane/saasplane` HTTP polling / event sync adapter. Public
  `controlplane/saasplane` owns that adapter outside the root package.
- task DB source/reporter adapters outside the root package. Public
  `controlplane/taskdbplane` owns `riido-task-db.v1`; public `internal/project`
  owns project/mwsd projection sync outside this context.
- `riidoaiserver`, local API, project persistence, packaging, infra, secrets

### 7.9 Supervisor boundary

`internal/agentbridge/supervisor` 는 Daemon tier RunController 다. Provider adapter 도
영속 scheduler 도 아니며, public daemon 안에서 이미 분리된 C4 RuntimeActor / C5
Scheduling / C6 Workdir / C2 EventIngestor / control-plane port 를 한 task run 단위로
조립한다.

Supervisor 의 책임:

- Start 시 RuntimeActor pool 을 control-plane source 에 등록하고 heartbeat 를 보낸다.
- runtime id 별로 task 를 claim 하고, duplicate in-flight task 를 방지한다.
- `internal/scheduling` eligibility evaluator 로 provider / surface /
  experimental-runtime opt-in 을 process spawn 전에 검증한다.
- workdir adapter 가 설정된 경우 task/run workspace 를 준비하고 native config 를
  주입한다.
- provider event 와 terminal result 를 daemon-side `internal/ir/ingest` 에 draft 로
  위임해 `CanonicalEvent` 로 append 하게 한다.
- terminal result 를 `TaskReporterPort` 로 보고하고, stop/cancel 시 in-flight run 을
  cancelled 로 정리하며 archive 를 best-effort 로 남긴다.

Supervisor 의 비책임:

- `riido-task-db.v1` guarded mutation, local task DB lease sidecars,
  project/mwsd sync, local API, SaaS HTTP/SSE transport, or infra/state/secret
  ownership. Public `internal/taskdb` and `controlplane/taskdbplane` own the
  task DB pieces, public `internal/project` owns project/mwsd projection sync,
  public `internal/riidoapi` owns local API, and SaaS/infra remain separate
  repos or adapters.
- concrete provider parser/command/protocol implementation.
- C1/C2/C3 schema ownership. 이 타입들은 public `riido-contracts` 에서 import 한다.
- persistent lease registry / fencing-token primitive. Public
  `controlplane/taskdbplane` owns this adapter boundary.

## 8. provider session 보존 (영속화의 1 차 키)

**C4 Provider Runtime owns the provider session table.** 이 table 은 provider
native session identity 를 Riido runtime identity 에 매핑하는 C4 저장소다. C5
Runtime Scheduling 은 이 table 의 schema / retention / adapter 를 소유하지 않고,
task claim lease 의 `(RuntimeID, CapabilityFingerprint)` 와 heartbeat 의미만 소유한다.
즉, C5 lease 가 "어느 runtime 이 task 를 진행할 수 있는가" 를 답하고, C4 provider
session table 은 "그 runtime 위에서 어떤 provider-native session/thread 를 resume 할
수 있는가" 를 답한다.

session resume 의 안전성은 다음 페어 보존에 달려있다.

| 페어 키 | 보존 위치 |
| --- | --- |
| (`TaskID`, `RunID`, `ProviderSessionID`) | IR 이벤트(`SessionPinned` payload) + C5 lease metadata 의 runtime pin |
| (`ProviderSessionID`, `RuntimeID`) | C4 provider session table (`riido-provider-session-table.v1`) |

adapter 는 session id 를 **자체 메모리에 들고 있지 않고**, 즉시 draft 로 ingest 에 넘긴다. crash 후 resume 은 IR 로그 + session table 에서 복구.

`riido-provider-session-table.v1` 의 최소 key 는 `(ProviderSessionID, RuntimeID)` 이다.
row 는 provider kind / protocol kind / last seen run identity / resume capability
provenance 를 담을 수 있지만, runtime eligibility 나 fencing token 을 다시 해석하지
않는다. lease expiry / stale fingerprint / task handoff 는 C5/C9 가 소유한다.

## 9. 인접 SSOT 와의 계약

| 인접 SSOT | 본 문서가 호출 / 위임 / 요구 |
| --- | --- |
| public `riido-contracts/docs/20-domain/provider-capability.md` (C3) | `Provider.Capability()` 는 C3 `ProviderCapability` 를 그대로 반환. capability 의 변경은 C3 의 책임. |
| public `riido-contracts/docs/20-domain/ir-event-log.md` (C2) | adapter 가 만드는 draft 의 `Type` 은 §3 카탈로그에 등록된 것만. append 권한은 §5.0 분리. |
| public `riido-contracts/docs/20-domain/ir-schema-versioning.md` | 9+2+FSMVersion 의무 필드 중 adapter 가 채울 수 있는 것은 §4.1 의 허용 필드만. 나머지는 ingest 가 확정. |
| public `riido-contracts/docs/20-domain/task-lifecycle.md` (C1) | adapter 는 transition 판정을 하지 않는다. `RunReportedDone` 같은 신호는 ingest/orchestrator 가 transition 으로 해석. |
| [`./runtime-scheduling.md`](./runtime-scheduling.md) (C5) | `Provider` 인스턴스 한 개 ↔ lease 한 개 ↔ (`RuntimeID`, `CapabilityFingerprint`) 한 페어. |
| [`./workspace.md`](./workspace.md) (C6) | adapter 는 workdir 경로 / native config 파일을 **읽기만** 한다. workdir 생성은 C6 가 사전에 수행. |
| [`./security.md`](./security.md) (C7) | `ExposesUnsafePermissionBypass` 가 true 라도 sandbox / permission / hook 활성 여부는 C7 의 정책 게이트가 결정. adapter 는 그 결정의 결과로만 동작. |
| [`../30-architecture/compatibility-gate.md`](../30-architecture/compatibility-gate.md) | G5 (Pre-Execute) 핸드셰이크는 adapter 측 `initialize` / `firstline` probe 를 호출. 그 결과로 lease 가 활성화. |
| [`../30-architecture/runtime-upgrade-flow.md`](../30-architecture/runtime-upgrade-flow.md) | adapter 가 `Running` 도중 `RuntimeID`/`CapabilityFingerprint` 변경을 감지하면 즉시 draft `Type=RuntimePinViolated` 를 발행하고 process 를 stop. |

### 9.1 Provider stdin command ACL

Reducer command 는 provider-neutral 이다. C4 adapter 가 provider stdin control protocol 을 가진 경우에만 `agentbridge.ProviderInputBuilder` 를 구현해 `CommandApproveTool` / `CommandRejectTool` / `CommandWriteProviderInput` 을 concrete byte frame 으로 바꾼다.

현재 집행 표면:

| Provider | 입력 command | Concrete frame |
| --- | --- | --- |
| Claude | `CommandApproveTool` / `CommandRejectTool` | `control_response` stream-json frame. `control_request.request_id` 는 `ToolRef.ProviderRequestID` 로 보존한다. |
| Codex | JSON-RPC protocol driver 내부 | pending request id 를 driver actor 가 소유하고 JSON-RPC response 로 처리한다. |

## 10. Provider adapter implementations — 본 컨텍스트 안에서의 위치

각 어댑터의 capability·protocol·우선순위는 public `riido-contracts/docs/20-domain/provider-capability.md` §4 가 소유한다. 본 문서는 **그 provider adapter 가 본 컨텍스트에서 어떻게 표현되는지** 만 적는다.

현재 public 구현 상태: RIID-4658 은 `ClaudeStreamJSONAdapter` 를
[`internal/provider/claude`](../../internal/provider/claude) 로 이동했고, RIID-4659 는
Codex app-server adapter 를 [`internal/provider/codex`](../../internal/provider/codex) 로
이동했으며, RIID-4660 은 OpenClaw adapter 를
[`internal/provider/openclaw`](../../internal/provider/openclaw) 로 이동했고, RIID-4661 은
Cursor adapter 를 [`internal/provider/cursor`](../../internal/provider/cursor) 로
이동했다.

| Adapter | 본 컨텍스트의 표현 | 1차 draft 카테고리 |
| --- | --- | --- |
| `ClaudeStreamJSONAdapter` | `claude -p --output-format stream-json` 의 stdout 라인을 NDJSON 으로 흡수. session id 는 `system.init` 라인에서 추출. | Cat C 위주 |
| `CodexAppServerAdapter` | `codex --sandbox danger-full-access app-server --listen stdio://` JSON-RPC. sandbox selection 은 default/caller 값이 아니라 daemon-owned full-access harness envelope 다. `initialize` 핸드셰이크와 pending request actor 로 approval response 를 처리한다. | Cat C + approval (Cat C `ApprovalRequested`) |
| `OpenClawAgentJSONAdapter` | `openclaw agent --local --json` 의 JSON/NDJSON 출력을 흡수. calendar-version gate 로 unsupported CLI 를 unavailable 로 접는다. | Cat C 위주 |
| `CursorAgentStreamJSONAdapter` | `cursor-agent -p --output-format stream-json` root-print shape 를 기본으로 사용하고, version/profile 차이는 explicit launch profile 로만 선택한다. | Cat C 위주 |

위 4 어댑터는 모두 같은 `Provider` 포트를 구현하고 같은 `ProviderEventDraft` 출력을 갖는다. **어댑터 별 분기는 ProtocolKind 로만** (provider-capability §0 invariant 1).

## 11. 미결정 / 오픈 이슈

Open questions roadmap 문서는 [`../50-roadmap/open-questions.md`](../50-roadmap/open-questions.md) 가 소유한다. `Q-RT-001` 과 legacy `Q-MULTICA-005` 는 §7.5 로 닫혔고, `Q-CTX-001` 은 §7.5/§7.7 로 닫혔으며, `Q-RT-003` 은 §5.6 으로 닫혔고, `Q-RT-005` 는 §8 로 닫혔다.

- `Q-RT-002`: provider process crash 와 lease handoff 사이의 정확한 ordering (`ConnectionLost` draft → ingest → handoff orchestration).
- `Q-RT-004`: wrapper 매니페스트의 표준 위치 / 형식(공개 spec vs 사내 전용).
- `Q-RT-006`: Codex app-server `thread/fork` 같은 experimental surface 의 사용 가부 — `task.allowExperimentalRuntime` 외에 어떤 추가 게이트가 필요한가.

