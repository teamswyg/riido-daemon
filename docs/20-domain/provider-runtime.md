# Provider Runtime / Adapter SSOT

> **이 문서가 Provider 어댑터 포트 / process · session · run lifecycle / adapter ACL 출력 타입(`ProviderEventDraft`) / cancel · resume · needs-input 처리의 SSOT다.**
>
> - 책임: provider 를 “어떻게 실행하는가” 의 도메인 모델. Provider port interface, process / session / run 의 lifecycle, adapter ACL 변환 규칙, draft 생성 책임.
> - 비책임: provider 가 “무엇을 할 수 있는가” 의 정적 모델 — public `riido-contracts` 의 `docs/20-domain/provider-capability.md` (C3). 어느 task 를 어느 runtime 이 claim 하는가 — [`./runtime-scheduling.md`](./runtime-scheduling.md) (C5). workspace 생성 — [`./workspace.md`](./workspace.md) (C6). 정책 결정 — [`./security.md`](./security.md) (C7). validation 결과 판단 — [`./validation.md`](./validation.md) (C8). **event append authority** — public `riido-contracts` 의 `docs/20-domain/ir-event-log.md` §5.0 와 daemon-side [`internal/ir/ingest`](../../internal/ir/ingest) 구현이 함께 소유한다.

이 SSOT 는 split-repo context map 의 **C4 Provider Runtime / Adapter** context 를 채운다. C1/C2/C3 contract SSOT 는 public `riido-contracts` repository 가 소유하고, 이 repository 는 customer-PC daemon 의 실행 boundary 를 소유한다. C3 ↔ C4 경계는 §2 가 못박는다.

## 0. Public Migration Status

RIID-4651 에서 public `riido-daemon` 으로 이동한 구현 범위는 `internal/agentbridge` 루트 package 다. 이 package 는 stdlib-only provider-neutral 도메인으로, `Adapter` port, `RawEvent` / `Parser`, `RunState`, reducer, telemetry parser, tool start gate 를 포함한다.

아직 이 slice 에 포함하지 않은 구현:

- task DB/project/mwsd/local API, server/control-plane/infra/secret/state files

The Figma onboarding planning screen (`v.1.22 AI Agent`, `node-id=42-3014`) is
outside C4 ownership. C4 may report runtime detection and liveness used by the
client's onboarding runtime-choice screen, and it may execute a task using the
instruction that SaaS assigns later. C4 does not own the onboarding template
catalog, direct-setting form composition, workspace selector, no-runtime skip
branch, scroll affordances, two-line ellipsis behavior, or starter-agent copy.

The Figma web onboarding section (`node-id=236-29749`) is also outside C4
ownership. Its macOS app download CTA can lead a user to a desktop artifact, but
it is not a provider runtime command and does not authorize C4 to bundle,
download, install, or update Claude/Codex/OpenClaw/Cursor CLIs. Its sign-up,
terms consent, member invite, Windows waitlist, marketing-consent, chat
animation, and progress-bar reference facts belong to client/auth/team/product
surfaces until a separate SSOT promotes executable daemon behavior.

RIID-4652 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/agentbridge/toolargs` 와 `internal/agentbridge/toolpolicy` 다. 이 package
들은 provider raw tool input 을 bounded/redacted `ToolRef.Args` 로 요약하고,
provider-neutral `ToolRef` 를 C7 ToolUse risk surface 로 분류해 `AutoApprover` /
`ToolStartGate` 를 구성한다. provider-native approval RPC/hook 실행 wiring 은 여전히
후속 runtimeactor/provider-adapter migration slice 가 맡는다.

RIID-4653 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/agentbridge/session` 이다. 이 package 는 one-run session actor 로서
Process → Parser/ProtocolDriver → reducer → bounded Events/Result stream 을
연결하고, hard/semantic-idle timeout, cancellation, process-exit ordering, telemetry
extraction, `AutoApprover`, `ToolStartGate` fail-closed block, adapter temp-file
cleanup 을 소유한다. runtime pool / task claim loop / concrete provider adapter 는
여전히 후속 runtimeactor/supervisor/provider-adapter migration slice 가 맡는다.

RIID-4654 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/agentbridge/bridge` 와 `internal/agentbridge/detectutil` 이다. `bridge` 는
provider registry / detect / run entrypoint 로서 `StartCommand` 를 public
`internal/process` port 로 변환하고 `session` actor 를 시작한다. `detectutil` 은
provider adapter 들이 공유할 PATH lookup / env override pin / version probe helper
이며 concrete provider adapter 자체는 아니다. runtime pool / supervisor / task claim
loop / concrete provider adapter 는 여전히 후속 migration slice 가 맡는다.

RIID-4656 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/agentbridge/runtimeactor` 이다. 이 package 는 one runtime capability
boundary 의 mailbox-owned actor 로서 adapter detect, C3 capability reconciliation,
bounded slot pool, Submit/Cancel/Status/Heartbeat, session handoff, Stop idempotency,
and cancellation cascade 를 소유한다. supervisor task claim loop / control-plane
transport / concrete provider adapter 는 여전히 후속 migration slice 가 맡는다.

RIID-4657 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/agentbridge/controlplane` 의 root port package 다. 이 package 는 daemon 이
task source 에 runtime registration / heartbeat / claim / cancel-watch 를 요청하고
task reporter 에 start/event/complete 를 보고하는 provider-neutral port 계약을
소유한다. `controlplane/saasplane`, `controlplane/taskdbplane`, supervisor polling
loop, server HTTP/SSE transport, task DB/project/mwsd adapter 는 RIID-4657 당시
후속 migration slice 가 맡는 것으로 남겼다.

RIID-4658 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/provider/claude` 다. 이 package 는 Claude Code CLI 를 번들하지 않고,
external executable detection, command construction, stream-json parser, raw event
translator, stdin protocol driver, provider input approval frame builder 를 소유한다.
real Claude CLI execution 은 `AGENTBRIDGE_INTEGRATION=1` 로 opt-in 된 경우에만 검증한다.
Codex/OpenClaw/Cursor adapter, supervisor polling loop, server/task DB/project/mwsd
adapter 는 RIID-4658 당시 후속 migration slice 가 맡는 것으로 남겼다.

RIID-4659 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/provider/codex` 다. 이 package 는 Codex CLI 를 번들하지 않고,
`codex app-server --listen stdio://` command construction, per-task `CODEX_HOME`
isolation, JSONL parser, raw event translator, JSON-RPC protocol driver, pending
request actor, approval response path 를 소유한다. real Codex CLI execution 은
`AGENTBRIDGE_INTEGRATION=1` 로 opt-in 된 경우에만 검증한다. OpenClaw/Cursor
adapter, supervisor polling loop, server/task DB/project/mwsd adapter 는 RIID-4659
당시 후속 migration slice 가 맡는 것으로 남겼다.

RIID-4660 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/provider/openclaw` 다. 이 package 는 OpenClaw CLI 를 번들하지 않고,
external executable detection, calendar-version gate, `openclaw agent --local --json`
command construction, mandatory session id resolution, JSON/NDJSON parser, raw event
translator 를 소유한다. real OpenClaw CLI execution 은 `AGENTBRIDGE_INTEGRATION=1` 로
opt-in 된 경우에만 검증한다. Cursor adapter, supervisor polling loop, server/task
DB/project/mwsd adapter 는 RIID-4660 당시 후속 migration slice 가 맡는 것으로
남겼다.

RIID-4661 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/provider/cursor` 다. 이 package 는 Cursor Agent CLI 를 번들하지 않고,
root-print / agent-subcommand / legacy-chat launch profile selection, `--yolo` unsafe
bypass policy gate, external executable detection, stream-json parser, raw event
translator 를 소유한다. real Cursor Agent CLI execution 은 `AGENTBRIDGE_INTEGRATION=1`
로 opt-in 된 경우에만 검증한다. supervisor polling loop, server/task DB/project/mwsd
adapter 는 RIID-4661 당시 후속 migration slice 가 맡는 것으로 남겼다.

RIID-4662 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/agentbridge/supervisor` 다. 이 package 는 Daemon tier control loop 로서
RuntimeActor pool registration / heartbeat, task claim, pre-submit C5 eligibility,
workdir preparation, EventIngestor append delegation, terminal result reporting, and
shutdown cancellation/archive 를 연결한다. RIID-4662 당시에는
`controlplane/saasplane`, `controlplane/taskdbplane`, task DB/project/mwsd/local API,
server HTTP transport, infra/secret/state files 를 후속 migration slice 또는 private
repo 가 맡기로 남겼다.

RIID-4683 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/taskdb` 와 `internal/agentbridge/controlplane/taskdbplane` 이다.
`internal/taskdb` 는 `riido-task-db.v1` schema, guarded transition/evidence
mutation, command-id idempotent replay, and deterministic validation evidence
receipt 를 소유한다. `taskdbplane` 은 해당 JSON DB 를 first-class local
control-plane source/reporter 로 사용하며, runtime registry sidecar, lease sidecar,
fencing token 검증, expired lease handoff 를 같은 C9 file lock 아래에서 수행한다.
이 slice 는 project/mwsd sync, local API/socket, CLI commands, `saasplane`, server
HTTP transport, infra/secret/state files 를 이동하지 않는다.

RIID-4684 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/riidoapi` local API adapter 다. 이 adapter 는 local IPC envelope 와
Unix-socket / Windows named-pipe transport 를 소유하고, public `internal/taskdb`
guarded mutation 과 `internal/validation` 을 호출한다. provider runtime 은 이 local
API transport 를 소유하지 않는다.

RIID-4686 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/mwsdbridge`, `internal/project`, and `riido mwsd ...` 이다.
`mwsdbridge` 는 macmini-workspace daemon 의 local JSON socket contract 만 읽는
anti-corruption layer 이고, `project` 는 `riido-workspace-projection.v1` /
`riido-project-state.v1` 과 project-to-taskdb projection sync 를 소유한다. 이 sync 는
문서 기반 task source 를 public `internal/taskdb` row 로 투영할 뿐, provider process
execution / runtime session / SaaS transport 를 소유하지 않는다.

RIID-4689 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/agentbridge/controlplane/saasplane` 이다. 이 adapter 는
`github.com/teamswyg/riido-contracts/assignment v0.3.0` 의 shared DTO/state/event
contract 를 사용해 SaaS assignment poll/heartbeat/event HTTP API 를
TaskSourcePort/TaskReporterPort 로 번역한다. HTTP handler, store actor, SSE,
authZ, metrics/health, persistence, Terraform/AWS/deploy evidence 는 여전히
`riido-control-plane` 또는 `riido-infra` 가 소유한다.

RIID-4690 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`cmd/riido daemon ...` lifecycle adapter 다. 이 adapter 는 public provider
adapters, `runtimeactor`, `supervisor`, `taskdbplane`, and `saasplane` 을 하나의
customer-PC process 로 조립하고 local-only Unix socket 에 status/health/ready/metrics
JSON 을 노출한다. Provider CLI binary bundling, server HTTP/SSE implementation,
Terraform/AWS/deploy evidence, and private machine-local state 는 이 context 밖에
남는다.

## 1. 책임 한 줄

> Provider Runtime 은 **provider 의 실행 표면** (process, session, run, raw event stream)을 도메인 안으로 끌어들여 **정규화된 draft** 로 변환한다. 그것이 본 컨텍스트의 시작이고 끝이다.

본 컨텍스트는 다음을 **하지 않는다**:

- `CanonicalEvent` 를 IR 로그에 append (이 권한은 `riido-contracts` IR append authority 계약과 public daemon [`internal/ir/ingest`](../../internal/ir/ingest) 구현).
- `ProviderCapability` 를 만들거나 update (C3 의 권한).
- agent 설정을 생성, 저장, 수정. Agent profile / description / instruction 의미와
  API shape 는 upstream contracts/control-plane SSOT 가 소유하고, C4 는 이미
  배정된 run 입력만 provider process 로 전달한다. Agent list `updated_at`,
  add-screen save enablement, row/meatball edit entry, long-description
  presentation, and absolute-time tooltips are client/control-plane facts, not
  provider-runtime inputs.
- Figma menu placement (`node-id=156-19307`) 또는 client route selected state 를
  해석한다. 메뉴는 runtime 실행 입력이 아니며, C4 는 route 진입 이후 배정된 run 만
  소비한다.
- Figma task-thread annotations (`node-id=153-15931`) 의 scroll, hover, modal,
  animation reference 또는 viewer-away 상태의 thread 표시 방식을 해석한다.
  `riido.aiAgent.events.stream` / `riido.aiAgent.tasks.stop` /
  `riido.aiAgent.tasks.threads` 는 control-plane/client generated path evidence
  이고, C4 는 upstream ingest/orchestrator 가 내린 cancel/interrupt 와
  `<riido_log>...<end>` telemetry marker 만 실행 표면으로 소비한다.
  Client-facing cold thread collection, active stream link selection,
  persisted viewer-away thread visibility, and rendered thread composition are
  not provider-runtime facts.
- Figma normal task-thread screen (`node-id=236-21379`) 의 generic comment
  input, AI Agent reply input, send-button state, right-side task details panel,
  또는 `중지` button rendering 을 해석한다. C4 는 browser/client click 을 직접
  관찰하지 않는다. SaaS polling/assignment response 가 cancel/interrupt 를
  내려준 뒤에만 provider runtime process 에 중단을 반영하고, progress/result
  는 SaaS thread-progress/reporting port 로 올린다.
- Figma participant dropdown annotations (`node-id=153-12742`) 의 member/agent
  정렬, 긴 이름 표시, max height, scrollbar width, checkbox layout 을 해석한다.
  assignable-agent response 와 client composition 은 control-plane/client
  boundary 이며, C4 는 dropdown 표시 순서나 멤버 목록을 만들지 않는다.
- Figma runtime settings (`node-id=162-23090`) 의 agent hover popover, daemon
  stop modal copy, restart animation, or remote-device table presentation 을
  해석한다. C4 는 provider process/run lifecycle 과 runtime status 를 공급하고,
  `cmd/riido daemon ...` local lifecycle adapter 가 current-device status/stop
  surface 를 조립한다. SaaS device/runtime read model 은 control-plane contract
  projection 이며 C4 가 만들지 않는다.
- Figma runtime settings empty-state (`node-id=275-22731`) 의 provider
  install-card hover, Windows app waitlist copy, or marketing-consent state 를
  해석한다. C4 는 provider CLI 를 bundle/download/install 하지 않으며, waitlist
  mutation 이 필요하면 client/product/control-plane SSOT 가 먼저 소유해야 한다.
- Figma web onboarding (`node-id=236-29749`) 의 macOS app download CTA,
  sign-up/terms/member-invite flows, Windows waitlist/marketing consent,
  chat animation, or progress-bar reference 를 해석한다. C4 는 external provider
  executable detection/execution boundary 만 소유하고, auth/team/distribution
  presentation 을 runtime command 로 바꾸지 않는다.
- Figma agent settings (`node-id=164-50215`) 와 agent add
  (`node-id=134-6542`) 의 create/update form, save-button enablement, row edit
  entry, absolute-time tooltip, long-description UI, runtime dropdown rendering,
  or model dropdown catalog 를 해석한다. Runtime binding 과 runtime-scoped
  `model_id` 는 upstream assignment/configuration input 으로 소비할 수 있다.
  하지만 provider-specific model catalog 와 label 은 public contracts 의
  `runtime_model_catalog.v1` / control-plane read model 이 소유하며, C4 는
  이미 승인된 실행 요청의 model 값만 provider adapter argument 로 변환한다.
- task 를 lease / claim / heartbeat (C5).
- workdir / native config 작성 (C6).
- 정책 / sandbox / 보호 경로 결정 (C7).
- validation 결과 판단 (C8).

## 2. C3 ↔ C4 경계 (단단히 박는다)

| 질문 | 답 (소유 context) |
| --- | --- |
| “이 provider 는 무엇을 할 수 있는가?” (surface flag, EventStreamFormat, fingerprint) | **C3 Provider Capability** (public `riido-contracts/docs/20-domain/provider-capability.md`) |
| “이 task 를 지금 어떻게 실행하는가?” (process 기동, session resume, stdout 파싱, raw → draft) | **C4 Provider Runtime / Adapter** (본 문서) |
| “raw event 가 어떤 도메인 의미를 가지는가?” (어댑터 ACL 매핑) | **C4** (본 문서 §6) |
| “이 task 의 lease 는 어느 runtime 이 가지는가?” | **C5 Runtime Scheduling** ([`./runtime-scheduling.md`](./runtime-scheduling.md)) |

C4 는 C3 의 `ProviderCapability` 를 **읽기 전용으로 import** 한다. 반대 방향은 금지(public `riido-contracts/provider/capability` 가 daemon runtime package 를 import 해서는 안 된다).

## 3. Provider 어댑터 포트 (도메인 표현)

본 문서는 시그니처의 **도메인 표현** 을 박는다. public Go boundary 의 현재 구현은 [`internal/agentbridge`](../../internal/agentbridge) 루트 package, [`internal/agentbridge/session`](../../internal/agentbridge/session), [`internal/agentbridge/bridge`](../../internal/agentbridge/bridge), [`internal/agentbridge/detectutil`](../../internal/agentbridge/detectutil), [`internal/agentbridge/runtimeactor`](../../internal/agentbridge/runtimeactor), [`internal/agentbridge/controlplane`](../../internal/agentbridge/controlplane), [`internal/agentbridge/supervisor`](../../internal/agentbridge/supervisor), [`internal/provider/claude`](../../internal/provider/claude), [`internal/provider/codex`](../../internal/provider/codex), [`internal/provider/openclaw`](../../internal/provider/openclaw), [`internal/provider/cursor`](../../internal/provider/cursor), and `cmd/riido daemon ...` adapter, 그리고 [`docs/migration/daemon.md`](../migration/daemon.md) 의 RIID-4651 / RIID-4653 / RIID-4654 / RIID-4656 / RIID-4657 / RIID-4658 / RIID-4659 / RIID-4660 / RIID-4661 / RIID-4662 / RIID-4689 / RIID-4690 slice 가 확정한다.

```
Provider {
    // 식별
    Capability() ProviderCapability        // 현재 pinned capability snapshot (불변)

    // run lifecycle (한 task 의 한 run 동안 한 번)
    StartRun(ctx, RunRequest) -> RunHandle
    Cancel(ctx, RunHandle) -> error
    Interrupt(ctx, RunHandle) -> error

    // input / approval (NeedsInput / ApprovalRequested 흐름)
    ProvideInput(ctx, RunHandle, response) -> error
    ResolveApproval(ctx, approvalID, decision) -> error

    // observation stream (output channel)
    Drafts() <-chan ProviderEventDraft     // raw → draft. 이 채널이 본 컨텍스트의 출력.

    // session lifecycle
    PinSession(ctx, RunHandle, providerSessionID) -> error
    ResumeSession(ctx, providerSessionID) -> RunHandle   // 새 RunID 는 ingest 계층이 부여
}
```

규칙:

1. `Provider` 인스턴스 한 개는 **하나의 RuntimeID + CapabilityFingerprint 페어에 묶인다**. 그 페어가 변하면 새 `Provider` 인스턴스가 만들어진다(같은 인스턴스 재사용 금지 — runtime pinning invariant).
2. `Drafts()` 채널은 **adapter ACL 의 유일한 출력 경로**. 다른 경로로 raw 를 외부에 노출해서는 안 된다.
3. `Provider` 는 IR 로그 writer 를 직접 import 하지 않는다(append authority 분리, §7).

## 4. ProviderEventDraft — adapter ACL 출력 타입

`ProviderEventDraft` 는 어댑터가 만들 수 있는 **유일한 도메인 출력** 이다. `EventIngestor` (단일 Append API — `riido-contracts/docs/20-domain/ir-event-log.md` §5.0 와 public daemon [`internal/ir/ingest`](../../internal/ir/ingest)) 가 이 draft 를 받아 append-only record 에 필요한 identity / ordering / runtime identity / attribution / schema / timestamp 정책을 **최종 확정** 한 뒤 `CanonicalEvent` 로 적재한다. authorized caller(FSM Orchestrator / server transition layer 등)는 EventIngestor API 를 호출하는 방식으로만 append 에 관여하고, 직접 writer 를 갖지 않는다.

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

### 5.4 cancel / interrupt / needs-input

| 외부 신호 | adapter 액션 |
| --- | --- |
| ingest 가 `Cancel(handle)` 호출 | provider process SIGTERM → drain → SIGKILL fallback. 잔여 draft drain 후 close. |
| local daemon stop / supervisor context cancel | RunController 는 in-flight provider run 을 cancel 로 취급한다. process cancel 을 요청한 뒤 `TaskCancelled` transition event 와 `WorkdirArchived` manifest 를 best-effort 로 기록하고 runtime deregistration 을 진행한다. |
| ingest 가 `Interrupt(handle)` 호출 | provider 측 interrupt 메시지(`claude` interrupt, `codex` `turn/interrupt`). process 는 유지. |
| ingest 가 `ProvideInput(handle, response)` 호출 | provider stdin / RPC 로 응답 전송. provider 가 응답을 받아 계속 진행하면 그 결과는 평소처럼 draft 로 흐른다. |
| ingest 가 `ResolveApproval(approvalID, decision)` 호출 | provider approval 프로토콜 응답. Codex app-server `approval/resolved` 같은 메시지. |

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
- `progress` (provider 공통 Riido telemetry parser 가 `<riido_log>...<end>` 에서 생성)

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
- `StartCommand` 를 `process.Command` 로 변환하고 `internal/agentbridge/session` 을 시작한다.
- adapter 가 `ProtocolDriverProvider` 이면 one-run protocol driver 를 생성해 session 에 장착한다.
- session handle facade 를 반환하고 adapter `DroppedArgs` / `TempFiles` 를 session 경계까지 보존한다.

`bridge` 는 scheduler, task claim, EventIngestor append, workdir preparation, policy
decision, provider-specific parsing 을 소유하지 않는다. 이 책임들은 각각 C5, C2/C4
RunController, C6, C7, concrete adapter slice 가 소유한다.

`internal/agentbridge/detectutil` 은 concrete provider adapters 가 공유하는 탐지 helper
다. env override 는 hint 가 아니라 pin 이므로 override path 가 없거나 directory 이면
PATH fallback 을 하지 않고 fail-closed 한다. version probe helper 는 missing binary /
timeout / unclassifiable signal 을 unavailable 로 접고, strict probe 는 command completion
여부와 exit code 를 노출해 adapter 가 non-zero output 을 version 으로 오인하지 않게 한다.

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
| `CodexAppServerAdapter` | `codex app-server --listen stdio://` JSON-RPC. `initialize` 핸드셰이크와 pending request actor 로 approval response 를 처리한다. | Cat C + approval (Cat C `ApprovalRequested`) |
| `OpenClawAgentJSONAdapter` | `openclaw agent --local --json` 의 JSON/NDJSON 출력을 흡수. calendar-version gate 로 unsupported CLI 를 unavailable 로 접는다. | Cat C 위주 |
| `CursorAgentStreamJSONAdapter` | `cursor-agent -p --output-format stream-json` root-print shape 를 기본으로 사용하고, version/profile 차이는 explicit launch profile 로만 선택한다. | Cat C 위주 |

위 4 어댑터는 모두 같은 `Provider` 포트를 구현하고 같은 `ProviderEventDraft` 출력을 갖는다. **어댑터 별 분기는 ProtocolKind 로만** (provider-capability §0 invariant 1).

## 11. 미결정 / 오픈 이슈

Open questions roadmap 문서는 [`../50-roadmap/open-questions.md`](../50-roadmap/open-questions.md) 가 소유한다. `Q-RT-001` 과 legacy `Q-MULTICA-005` 는 §7.5 로 닫혔고, `Q-CTX-001` 은 §7.5/§7.7 로 닫혔으며, `Q-RT-003` 은 §5.6 으로 닫혔고, `Q-RT-005` 는 §8 로 닫혔다.

- `Q-RT-002`: provider process crash 와 lease handoff 사이의 정확한 ordering (`ConnectionLost` draft → ingest → handoff orchestration).
- `Q-RT-004`: wrapper 매니페스트의 표준 위치 / 형식(공개 spec vs 사내 전용).
- `Q-RT-006`: Codex app-server `thread/fork` 같은 experimental surface 의 사용 가부 — `task.allowExperimentalRuntime` 외에 어떤 추가 게이트가 필요한가.

## 12. version-affecting changes

- `ProviderEventDraft` 의 허용 필드 추가는 `change:additive`.
- 허용 필드 제거 또는 의미 변경은 `change:breaking-protocol` (어댑터 ↔ ingest 와이어 깨짐).
- 금지 필드 표(§4.2)에 새 필드 추가는 `change:breaking-policy` (ingest 계층의 채움 책임을 늘리기 때문).
- `Provider` 포트 시그니처 변경(`StartRun` / `Cancel` / `Drafts()` 등)은 `change:breaking-protocol`.
- 어댑터 4 종 (§10) 중 한 종의 `ProtocolKind` 분류가 바뀌면 public `riido-contracts/docs/20-domain/provider-capability.md` §4 와 동시 갱신.
