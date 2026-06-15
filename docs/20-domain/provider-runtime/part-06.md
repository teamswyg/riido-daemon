# Provider Runtime / Adapter SSOT: Part 06

[Back to provider-runtime.md](../provider-runtime.md)

## 6. raw → draft 변환 규칙 (어댑터 ACL)

본 문서가 강제하는 변환 규칙:

1. **알려진 raw type → 도메인 `EventType` 매핑.** 매핑 표는 어댑터마다 자기 코드 안에 두지만(예: `claude-stream-json` 의 `assistant.delta` → `TextDelta`), 정규화된 `Type` 은 public `riido-contracts/docs/20-domain/ir-event-log.md` §3 카탈로그에 등록된 것만 사용한다.
2. **알려지지 않은 raw type** → `Type=ProviderUnknownEvent`, `RawType=<원본>`, `Raw=<페이로드>`. FSM transition 절대 발생시키지 않는다(`riido-contracts` IR event log §6).
3. **알려진 raw type 이지만 모르는 raw 필드** → 알려진 필드는 정규화된 `Payload` 에, 모르는 필드는 `Unknown` 으로 보존. drop 금지.
4. **해석으로 의미가 추가된 경우** → `Payload.derived=true` 를 표기 (예: provider 가 “파일 수정” 을 자연어로만 말한 것을 `FileChanged` 로 추론한 경우).
5. **provider 가 transition-after-side 사실을 보고** (예: `RunReportedDone`) → adapter 는 draft 를 발행하지만, transition 자체는 ingest 가 결정.

## 7. EventIngestor 와의 계약 (단일 Append API + RunController 가 drain)

EventIngestor 의 append authority contract 는 public `riido-contracts/docs/20-domain/ir-event-log.md` §5.0 가 소유하고, daemon-side 구현은 public [`internal/ir/ingest`](../../internal/ir/ingest) 가 소유한다. 단, C4 ↔ ingest 사이의 계약은 본 문서가 박는다.

### 7.1 RunController — C4 의 orchestration layer

`Provider.Drafts()` 채널을 **누가 읽어서** `EventIngestor` 의 single append API 를 호출하는가? — adapter 구현체가 아니라 **RunController** 가 한다. RunController 는 C4 의 orchestration layer 다(adapter 자체가 아니다). 책임:

- 한 run 동안 `Provider.Drafts()` 채널을 drain.
- 받은 `ProviderEventDraft` 를 EventIngestor 의 단일 API 로 넘김 (authorized caller). 현재 public Go API 는 `ingest.Ingestor.Append(ctx, ingest.Draft)` 다.
- adapter 의 lifecycle 호출(`Cancel` / `Interrupt` / `ProvideInput` / `ResolveApproval`) 을 외부 orchestrator 신호에 따라 수행.
- adapter 가 `Drafts()` 채널을 닫으면 run lifecycle 을 종료시키고 cleanup.

> **단단히 박는 한 줄**: Adapter 구현체는 EventIngestor 를 모른다. RunController 가 `Provider.Drafts()` 를 drain 하고 EventIngestor single append API 를 authorized caller 로서 호출한다. RunController 는 adapter 가 아니라 orchestration layer 다.

### 7.2 흐름

```
provider raw stdout/RPC
   ↓ (adapter ACL 변환 — adapter 코드, [`security-redaction.md`](./security-redaction.md) 기준 1차 secret redaction 포함)
ProviderEventDraft
   ↓ Provider.Drafts() 채널
RunController (C4 orchestration, authorized caller)
   ↓ EventIngestor single append API
EventIngestor (single Append API, 유일한 writer 보유)
   ↓ identity / ordering / runtime identity / attribution / schema / timestamp 정책 확정
   ↓ [`security-redaction.md`](./security-redaction.md) 기준 2차 secret redaction / audit check
CanonicalEvent (append-only)
```

### 7.3 책임 표

| 방향 | 입력 | 출력 | 주체 |
| --- | --- | --- | --- |
| Adapter → Drafts() | provider raw stdout/RPC | `ProviderEventDraft` | adapter 구현체 |
| Drafts() → ingest | `Provider.Drafts()` 채널 drain | EventIngestor single append API 호출 | **RunController (C4 orchestration)** |
| 다른 authorized caller → ingest | 외부 신호(API / validation / scheduler / 운영자) | EventIngestor single append API 호출 | FSM Orchestrator, server transition layer, validation runner result handler, runtime scheduler result handler |
| ingest → 적재 | draft + lease 조회 + actor 정책 + 활성 schema | identity / ordering / runtime identity / attribution / schema / timestamp 확정 + `CanonicalEvent` append | EventIngestor (유일한 writer 보유) |
| RunController → adapter | 외부 사용자 신호 (cancel / interrupt / provideInput / resolveApproval) | adapter lifecycle 호출 | RunController |

### 7.4 규칙

1. **단일 API**: `CanonicalEvent` 를 append 할 수 있는 코드 경로는 EventIngestor single append API 하나뿐이다.
2. **Adapter 구현체는 EventIngestor 를 import 하지 않는다**. adapter 는 `Drafts()` 채널을 채울 뿐.
3. **Reducer 는 EventIngestor 를 호출할 수 없다** — 순수 함수.
4. RunController 는 adapter 구현체와 분리된 패키지에 산다 — adapter 코드는 RunController 를 모른다(반대 방향 import 만 허용).
5. RunController 는 `Drafts()` drain 외의 신호(예: provider process 죽음, stderr fatal)도 받아 EventIngestor 로 적재.
6. adapter 가 `Drafts()` 를 닫으면 RunController 가 run lifecycle 을 종료시키고 cleanup. ingest 다운 시 RunController 는 §7.5 의 no-drop backpressure 계약을 유지한다.

### 7.5 Draft/session event backpressure

C4 Provider Runtime 이 provider process stream, provider draft / session event channel,
actor mailbox 의 숫자와 drop 정책을 소유한다. C6 workspace, C7 policy, C10 server 는
이 값을 재정의하지 않는다. 이 절은 `Q-RT-001`, legacy `Q-MULTICA-005`, 그리고
`Q-CTX-001` 의 runtime/session boundary 답이다.

`internal/agentbridge/session` 은 C4 내부 submodel 이며 별도 bounded context 가 아니다.
Claude/Codex/OpenClaw/Cursor 의 session id 차이는 concrete adapter ACL/protocol 차이로
처리하고, runtime/session lifecycle split decision 으로 승격하지 않는다.

| Surface | 구현 상수 | 값 | 정책 |
| --- | --- | --- | --- |
| process stdout chunk stream | `internal/process.DefaultStdoutBuffer` | `64` | no-drop, blocking process backpressure |
| process stderr chunk stream | `internal/process.DefaultStderrBuffer` | `64` | no-drop, blocking process backpressure |
| provider/session semantic event stream | `internal/agentbridge/session.DefaultEventBuffer` | `256` | no-drop, blocking backpressure |
| terminal result stream | `internal/agentbridge/session.DefaultResultBuffer` | `1` | exactly one terminal result |
| runtime actor mailbox | `internal/agentbridge/runtimeactor.DefaultMailboxSize` | `16` | caller-context bounded send |
| supervisor actor mailbox | `internal/agentbridge/supervisor.DefaultMailboxSize` | `64` | caller-context bounded send |

규칙:

1. process stdout/stderr channel 이 가득 차면 stream writer 는 block 한다. text/log/warning chunk 를 drop / overwrite / reorder 하지 않는다.
2. session actor 는 event buffer 가 가득 차면 event 를 drop / overwrite / reorder 하지 않고 consumer 가 `Events()` 를 drain 할 때까지 block 한다.
3. runtime actor 와 supervisor mailbox send 는 bounded send 이며 caller context 가 deadline / cancellation 을 소유한다. mailbox-full 을 숨겨진 retry queue 로 우회하지 않는다.
4. caller 는 `Events()` 를 close 될 때까지 drain 해야 한다. result-only caller 는 discard-drain 해야 하며, drain 하지 않는 caller 는 provider runtime 을 backpressure 로 세운 것이지 adapter bug 가 아니다.
5. C4 는 현재 in-memory channel 에서 retry queue 를 두지 않는다. EventIngestor / sink append failure 는 warning event 로 표면화하고, durable retry / outbox 는 C2/C10 future decision 이 소유한다.
6. buffer / mailbox 값을 바꾸는 slice 는 본 문서, 구현 상수, default-size tests, 그리고 `provider-runtime-backpressure` workflow 를 같은 work unit 에서 갱신해야 한다.

### 7.6 Bridge/detect helper boundary

`internal/agentbridge/bridge` 는 C4 provider runtime 의 provider-neutral library entrypoint
다. caller 는 adapter 목록과 process port 를 주입하고, bridge 는 다음만 수행한다.

- adapter registry 를 만들고 provider name 중복 / empty name 을 거부한다.
- `Detect(ctx)` 호출을 provider name 기준 stable order 로 반환한다.
- `TaskRequest` 를 `agentbridge.StartRequest` 로 변환해 adapter `BuildStart` 를 호출한다.
- SaaS assignment source 에서 온 `Assignment.agent_instruction` 은 provider 별
  runtime instruction 으로 materialize 된 뒤 `TaskRequest` 에 들어온다.
  Claude / OpenClaw 는 system prompt surface 를 쓰고, Codex / Cursor 는 prompt
  prefix 를 쓴다. instruction 값의 의미와 1000자 제한은 `riido-contracts` 가
  소유한다.
- `StartCommand` 를 `process.Command` 로 변환하고 `internal/agentbridge/session` 을 시작한다.
- adapter 가 `ProtocolDriverProvider` 이면 one-run protocol driver 를 생성해 session 에 장착한다.
- session handle facade 를 반환하고 adapter `DroppedArgs` / `TempFiles` 를 session 경계까지 보존한다.

`bridge` 는 scheduler, task claim, EventIngestor append, workdir preparation, policy
decision, provider-specific parsing 을 소유하지 않는다. 이 책임들은 각각 C5, C2/C4
RunController, C6, C7, concrete adapter slice 가 소유한다.

### 7.6.1 Agent instruction placement and effectiveness probe

C4 Provider Runtime owns only how an assignment-created
`Assignment.agent_instruction` snapshot is delivered to each provider surface.
It does not own the client-authored instruction text, the 1000-character limit,
or the assignment-time snapshot decision; those remain in `riido-contracts`
AI Agent policy and assignment polling contracts.

Current placement matrix:

| Provider | Instruction placement | Telemetry placement | Deterministic gate |
| --- | --- | --- | --- |
| Claude Code | `system-prompt` | `system-prompt` | `go test ./internal/agentbridge` |
| OpenClaw | `system-prompt-inline` | `system-prompt-inline` | `go test ./internal/agentbridge` |
| Codex | `prompt` prefix | `prompt` prefix | `go test ./internal/agentbridge` |
| Cursor Agent | `prompt` prefix | `prompt` prefix | `go test ./internal/agentbridge` |

The matrix is implemented by `RuntimeInstructionStrategies()` and consumed by
`ApplyRuntimeInstructionContract`. Public CI verifies deterministic placement,
metadata, idempotent section composition, and the provider-neutral effectiveness
probe shape without executing external provider CLIs.

Provider-specific "effectiveness" means the real provider obeys the delivered
instruction after it is placed on that provider's chosen surface. It is verified
by `BuildInstructionEffectivenessProbe` and
`ValidateInstructionEffectivenessOutput`: the probe asks the provider to echo a
provider-specific marker such as `RIIDO_INSTRUCTION_ACK:codex`, and the validator
accepts only outputs containing that marker. Real provider execution remains an
opt-in integration/evidence gate because provider CLIs, credentials, model
selection, latency, and vendor behavior are external attached resources.

`internal/agentbridge/detectutil` 은 concrete provider adapters 가 공유하는 탐지 helper
다. env override 는 hint 가 아니라 pin 이므로 override path 가 없거나 directory 이면
PATH fallback 을 하지 않고 fail-closed 한다. version probe helper 는 missing binary /
timeout / unclassifiable signal 을 unavailable 로 접고, strict probe 는 command completion
여부와 exit code 를 노출해 adapter 가 non-zero output 을 version 으로 오인하지 않게 한다.
`ResolveExecutableCandidates` 는 no-override PATH 후보 목록을 PATH 순서대로 제공하지만,
그 후보를 하나만 쓸지 여러 개 probe 할지는 concrete adapter 의 호환성 정책이다. 현재
OpenClaw 만 calendar-version gate 특성상 지원 버전 후보를 찾을 때까지 later PATH
candidate 를 probe 할 수 있다. `RIIDO_OPENCLAW_PATH` 가 설정된 경우에는 여전히 pin 이며
구버전/오류여도 PATH fallback 을 하지 않는다.

override 가 없을 때 후보 탐색은 process `PATH` 만이 아니라 augmented search path 를
쓴다. Desktop app / launchd / service 로 기동된 daemon 은 최소 `PATH`(macOS launchd 는
보통 `/usr/bin:/bin:/usr/sbin:/sbin`)만 상속해 Homebrew·per-user 디렉터리에 설치된
`claude`/`codex`/`cursor-agent`/`openclaw` 를 못 찾고 `detection_state=missing` 로
보고하던 문제를 막기 위함이다. 탐색 순서는 process `PATH` → login-shell `PATH`(프로세스당
1회 `$SHELL -lc` 로 읽어 캐시, Windows·`$SHELL` 미설정·timeout 시 skip) → well-known
install 디렉터리다. 이는 unset-override lookup 의 탐색 범위만 넓히며 `RIIDO_<PROVIDER>_PATH`
pin 의 fail-closed 의미는 그대로다. 카탈로그는
[`../30-architecture/config-reference.md`](../30-architecture/config-reference.md) 가 소유한다.

Detect 가 선택한 executable path 는 capability snapshot 의 실행 사실이다. `bridge.Run`
과 `runtimeactor.Submit` 은 이 값을 `StartRequest.Executable` 로 `BuildStart` 까지
전달하고, concrete provider adapter 는 이를 다시 `PATH` 에서 재해석하지 않는다. Adapter
specific `StartOptions.Executable` 만 이 값을 override 할 수 있으며, 둘 다 비어 있을 때만
provider default executable name 을 사용한다.

