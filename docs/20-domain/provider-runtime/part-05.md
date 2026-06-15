# Provider Runtime / Adapter SSOT: Part 05

[Back to provider-runtime.md](../provider-runtime.md)

### 4.1 허용 필드 (어댑터가 채울 수 있는 것)

```
ProviderEventDraft {
    Type              ir.EventType          // 정규화된 EventType (Cat C 위주, transition 후보면 ingest 가 검증)
    Payload           map[string]any        // 정규화된 payload (schema 는 riido-contracts ir-event-log.md §3)
    Unknown           map[string]any        // 알려지지 않은 raw 필드 보존 (ACL "unknown" 잔여)

    // provider 측 식별자 (raw 그대로)
    ProviderSessionID string                // session 식별자 (Claude session id / Codex thread id)
    ProviderTurnID    string                // turn 식별자 (있을 때)

    // 원본 보존 (replay / 재해석 자산)
    RawType           string                // provider 가 보고한 raw event type
    Raw               map[string]any        // raw payload 사본

    // 시점
    ObservedAt        time.Time             // adapter 가 line 을 읽은 시각
}
```

### 4.2 금지 필드 (어댑터가 채워서는 안 되는 것)

다음 필드는 **ingest 계층이 결정** 한다. adapter 가 직접 채우면 `riido-contracts` IR append authority 권한 분리 위반이다.

| 필드 | 누가 채우는가 | 왜 adapter 가 아닌가 |
| --- | --- | --- |
| `EventID` | EventIngestor | 중앙에서 ULID/UUID7 발급해야 단조 정렬 가능 |
| sequence / ordering metadata | EventIngestor | 같은 task 의 이벤트가 단조여야 함 |
| `RuntimeID` | EventIngestor (lease 조회) | adapter 인스턴스가 자신의 RuntimeID 를 알더라도 lease 가 실제 owning runtime |
| `CapabilityFingerprint` | EventIngestor (lease 조회) | 짝지어진 lease 의 fingerprint 와 일치해야 함 |
| `ActorKind` | authorized caller / EventIngestor config | adapter 가 결정하면 attribution invariant 위반 |
| `ActorID` | server transition layer | 동상 |
| `EventSchemaVersion` | EventIngestor | 현재 활성 reducer 가 인지 가능한 버전을 부여. adapter 마다 다른 버전 결정 금지 |
| `FSMVersion` | server transition layer | transition event 일 때만, 활성 FSM schema 기준으로 결정 |
| `OccurredAt` vs `IngestedAt` 정책 | EventIngestor | OccurredAt = `draft.ObservedAt` 으로 둘지, IngestedAt 따로 둘지를 ingest 계층이 결정 |

### 4.3 한 줄 invariant

> **Adapter 는 관측한다. Adapter ACL 은 정규화 초안을 만든다. EventIngestor 만 append 한다.**
>
> `Only EventIngestor / FSM Orchestrator / server transition layer may append CanonicalEvent. Provider Adapter may only produce ProviderEventDraft.`

## 5. process / session / run lifecycle

### 5.1 process lifecycle

| 단계 | adapter 책임 |
| --- | --- |
| **spawn** | `RunHandle.start` 시 provider CLI 또는 app-server 를 기동. exit code, stderr, stdout 핸들을 보유. |
| **observe** | stdout/JSON-RPC notification 을 `ProviderEventDraft` 로 normalize. |
| **interrupt** | `Interrupt(ctx, handle)` — 진행 중 stream 을 멈춤(예: `claude` SIGINT, `codex app-server` `turn/interrupt`). draft 발행은 멈추지 않고 “interrupted” 신호를 보낸다. |
| **stop** | `Cancel(ctx, handle)` — process 종료. 잔여 stdout 을 drain 후 close. |

규칙:

- process 가 죽으면 adapter 가 `ProviderEventDraft(Type=LogLine, level="fatal", text=...)` 를 발행한다. **`TaskFailed` 같은 transition draft 를 직접 발행하지 않는다** — transition 결정은 server orchestrator 가 한다(adapter 는 “죽었다” 라는 관측만 제공).

### 5.2 session lifecycle

| 단계 | adapter 책임 |
| --- | --- |
| **pin** | provider 가 session id 를 보고하면 즉시 `ProviderEventDraft(Type=SessionPinned, ProviderSessionID=...)` 발행. “early pin” invariant 를 따른다. |
| **resume** | `ResumeSession(ctx, providerSessionID)` — 기존 provider session 으로 새 run 시작. 새 `RunID` 는 ingest 계층이 부여(adapter 는 알 필요 없음). |
| **fork** | (선택 — Codex `thread/fork` 처럼 experimental surface). resume 과 같은 흐름이지만 `Payload.fork=true`. |
| **close** | process stop 직후 session 은 자동 close 로 간주. 별도 draft 발행 없음. |

`SessionPinned` draft 발행 시점은 가능한 한 **provider 가 session id 를 처음 노출하는 순간** 이다. 늦게 발행하면 crash 후 resume 이 깨진다.

### 5.3 run lifecycle (단일 run 동안)

```
RunStarted (ingest 결정 — adapter 는 draft 발행: Type=RunStarted 후보 + raw)
   ↓
TextDelta / ReasoningDelta / ToolCallStarted / ToolCallFinished /
FileChanged / CommandStarted / CommandFinished / StatusUpdate / UsageDelta / LogLine
   ↓ (반복)
InputRequested            (사용자/외부 응답 필요 시)
   ↓ ProvideInput
계속
   ↓
ApprovalRequested         (Codex app-server approval 흐름)
   ↓ ResolveApproval
계속
   ↓
RunReportedDone (provider 자기 보고)
```

- 위 모든 화살표 단계마다 adapter 는 한 개 또는 여러 개의 `ProviderEventDraft` 를 발행한다.
- `RunReportedDone` 은 “agent 가 끝났다고 신고” 일 뿐 task 완료가 아니다. local RunController(supervisor) 는 terminal provider `Result(completed)` 를 `RunReportedDone` transition event 로 append 하고, completion 판정은 validation gate (C8) 가 한다.
- `Result(failed|blocked|aborted|cancelled|timeout)` 은 adapter 가 직접 task 상태를 set 하지 않는다. local RunController 가 각각 `TaskFailed` / `TaskCancelled` / `TaskTimedOut` transition event 로 번역하고 `FSMVersion` 을 stamp 한다.
- provider transport 오류는 provider 자기보고 완료보다 우선한다. 특히 Codex JSON-RPC pending request 가 error response 로 닫히거나, Codex `error` notification 이후 의미 있는 assistant output 없이 빈 `turn_completed` / process exit 로 끝난 경우 adapter 는 `Result(failed)` 를 발행해야 한다. OpenClaw `full_result` 도 명시적 `error` 가 없더라도 `text` / payload text 가 비어 있으면 산출 없는 terminal result 이므로 `Result(failed)` 로 fail-closed 한다. 단, 오류 notification 이후 `TextDelta` 또는 non-empty `Result.Output` 이 관측되면 회복된 실행으로 보고 일반 완료를 허용한다.

### 5.4 cancel / interrupt / needs-input

| 외부 신호 | adapter 액션 |
| --- | --- |
| ingest 가 `Cancel(handle)` 호출 | provider process SIGTERM → drain → SIGKILL fallback. 잔여 draft drain 후 close. |
| local daemon stop / supervisor context cancel | RunController 는 in-flight provider run 을 cancel 로 취급한다. process cancel 을 요청한 뒤 `TaskCancelled` transition event 와 `WorkdirArchived` manifest 를 best-effort 로 기록하고 runtime deregistration 을 진행한다. |
| ingest 가 `Interrupt(handle)` 호출 | provider 측 interrupt 메시지(`claude` interrupt, `codex` `turn/interrupt`). process 는 유지. |
| ingest 가 `ProvideInput(handle, response)` 호출 | provider stdin / RPC 로 응답 전송. provider 가 응답을 받아 계속 진행하면 그 결과는 평소처럼 draft 로 흐른다. |
| ingest 가 `ResolveApproval(approvalID, decision)` 호출 | provider approval 프로토콜 응답. Codex app-server `approval/resolved` 같은 메시지. |

Shutdown authority (`none` → `graceful` → `forced`) and default shutdown
timeouts are owned by `pkg/lifecycle`. RuntimeActor, SupervisorActor, and
`cmd/riido daemon` must consume that model instead of redefining local stop
level parsing or timeout policy.

### 5.5 idle watchdog semantic activity

Idle watchdog 은 stdout byte activity 가 아니라 **provider 가 task 의미를 전진시킨 이벤트** 로만 갱신된다. 이 정의의 public daemon 구현은 `internal/agentbridge.EventKind.IsSemanticActivity()` 이다.

Semantic activity:

- `lifecycle`
- `text_delta`
- `thinking_delta`
- `tool_call_started`
- `tool_call_delta`
- `tool_call_completed`
- `tool_call_failed`
- `tool_approval_needed`
- `usage_delta`
- `progress` (provider 공통 Riido telemetry parser 가
  `<riido_log>{"code":...,"args":{...}}<end>` 에서 생성하고, legacy raw 문구는
  호환 fallback 으로만 code/args 에 매핑)

Non-semantic activity:

- `log`, `warning`, `error`
- `result`, `process_exit`
- `cancellation_requested`, `timeout`

즉 stderr heartbeat / log spam / process 신호만으로 idle watchdog 을 reset 하지 않는다. 이 규칙이 깨지면 provider 가 실제 진행 없이 로그만 뿜어도 run 이 무기한 살아남는다.

### 5.6 Approval wait timeout ownership

`Q-RT-003` is closed here: C4 Provider Runtime / Adapter owns approval wait
timeout policy through the session actor's run clocks.

The rule:

- provider adapters surface provider-native approval requests as
  `tool_approval_needed` / `ApprovalRequested`
- `tool_approval_needed` is semantic activity, so the first approval request
  resets `SemanticIdle`
- after the approval request, the same C4 `SemanticIdle` clock expires the run
  if there is no provider progress, auto-approval response, human approval
  response, cancellation, or terminal provider result
- `HardTimeout` remains the whole-run upper bound and also applies while waiting
  for approval
- `EventIngestor` appends observed draft events but does not own approval
  timers, expiry policy, or terminal timeout decisions
- UI / review surfaces may display the pending approval and send a response,
  but they are not the source of truth for timing out the provider run

When the C4 clock expires, the session actor emits `EventTimeout`; the reducer
turns it into `ResultTimeout` plus `CommandCancelProvider`, and the session
actor kills the provider process. If a provider reports its own timeout/error,
that raw observation is still translated as a provider event, but Riido's
provider-run timeout decision remains the C4 session actor decision.

