# Provider Runtime / Adapter SSOT: Part 04

[Back to provider-runtime.md](../provider-runtime.md)

## 1. 책임 한 줄

> Provider Runtime 은 **provider 의 실행 표면** (process, session, run, raw event stream)을 도메인 안으로 끌어들여 **정규화된 draft** 로 변환한다. 그것이 본 컨텍스트의 시작이고 끝이다.

본 컨텍스트는 다음을 **하지 않는다**:

- `CanonicalEvent` 를 IR 로그에 append (이 권한은 `riido-contracts` IR append authority 계약과 public daemon [`internal/ir/ingest`](../../internal/ir/ingest) 구현).
- `ProviderCapability` 를 만들거나 update (C3 의 권한).
- agent 설정을 생성, 저장, 수정. Agent profile / description / instruction 의미와
  API shape 는 upstream contracts/control-plane SSOT 가 소유하고, C4 는 이미
  배정된 run 입력만 provider process 로 전달한다. Agent list `created_at` /
  `updated_at`, add-screen save enablement, row/meatball edit entry,
  no-description row layout, status-label copy/color, long-description
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
  `<riido_log>{"code":...,"args":{...}}<end>` telemetry marker 만 실행
  표면으로 소비한다. Progress code catalog 와 append-only 정책은
  `riido-contracts/progressmessage/catalog.dsl.riido.json` 이 upstream SSOT 로
  소유하며, C4 는 provider output 을 code/args 로 정규화하는 projection 만
  실행한다.
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
- Figma additional planning section (`node-id=153-15935`) 의 task/subtask-only
  assignment target scope, AI property filler recommendation exclusion, or agent
  mention exclusion 을 해석한다. C4 는 project/milestone/intake/property-filler/
  mention surface 에서 agent 후보를 계산하지 않으며, task/subtask target 검증도
  하지 않는다. C4 는 SaaS 가 이미 승인한 assignment 만 실행 입력으로 소비한다.
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
- Figma agent settings (`node-id=432-37336`), agent add
  (`node-id=134-6542`), agent list/add affordance (`node-id=337-24001` /
  `node-id=337-24013`), and agent list (`node-id=432-35713`) 의 create/update
  form, save/add-button enablement, "모든 멤버가 런타임이 없으면" 표시 조건,
  row edit/delete entry, created/update date stamping, absolute-time tooltip,
  no-description row layout, status-label copy/color, long-description UI,
  runtime dropdown rendering, required-control state, model-default request
  semantics, or model dropdown catalog 를 해석한다. Runtime binding 과
  runtime-scoped `model_id` 는 upstream assignment/configuration input 으로
  소비할 수 있다. 하지만 provider-specific model catalog 와 label, 그리고
  omitted `model_id` 를 default model 로 해석하는 규칙은 public contracts 의
  `runtime_model_catalog.v1` / control-plane read model 이 소유하며, C4 는
  이미 승인된 실행 요청의 model 값만 provider adapter argument 로 변환한다.
- Figma onboarding direct-setting expansion (`node-id=164-26969`) 의 `이름`,
  `설명`, `지침` form composition, placeholder copy, dimmed fixture rows, or
  scroll behavior 를 해석한다. C4 는 control-plane 이 생성/할당한 agent
  configuration 의 instruction/runtime/model 값만 실행 입력으로 소비한다.
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

### 2.1 Provider trusted-runtime envelope rule

C4 의 provider command builder 는 provider-native "작업 가능 권한" 을 숨기거나
provider default 에 맡기지 않는다. Provider 가 실제 사용자 PC 에서 repo / toolchain /
workspace 를 다뤄야 한다면, adapter 는 그 provider 의 trusted-runtime envelope 를
명시적으로 만든다. 그 대신 Riido harness 가 다음을 소유한다.

- SaaS 가 승인한 immutable assignment snapshot
- daemon-selected workdir 과 evidence root
- provider process start / stop / cancel
- runtime slot, lease, fencing token, heartbeat, stale 판단
- dropped arg evidence 와 provider log/progress redaction
- real integration gate 와 filesystem side-effect 검증

이 규칙은 **"full-access 가 기본값"** 이라는 뜻이 아니다. 오히려 반대다. C4 는
provider default/caller args/SaaS payload 에서 sandbox 또는 approval-bypass 의미를
추론하지 않고, adapter 가 생성하는 단 하나의 launch envelope 와 harness 관리 책임을
함께 고정한다. Codex 의 경우 "default sandbox 가 danger-full-access" 가 아니라
"Codex adapter 가 danger-full-access envelope 만 생성하고 그 위험을 Riido harness 가
관리한다" 가 정확한 표현이다.

| Provider | 현재 C4 trusted-runtime envelope 상태 |
| --- | --- |
| Codex | 채택됨. `codex --sandbox danger-full-access app-server --listen stdio://` 만 adapter 가 생성한다. 이 값은 provider default/caller 선택이 아니라 daemon-owned trusted-runtime envelope 다. Caller `--sandbox`, config override, unsafe bypass arg 는 drop evidence 로 남긴다. |
| Claude | 전권 승격 미채택. `PermissionMode` 는 explicit input 이며, `bypassPermissions` 는 C7 unsafe-bypass gate 를 통과한 isolated tier 에서만 가능하다. |
| Cursor | 전권 승격 미채택. `--trust` 는 daemon-selected workdir acknowledgement 일 뿐이고, `--yolo` 는 계속 C7 unsafe-bypass gate 대상이다. |
| OpenClaw | 전권/worktree envelope 미채택. 현재 worktree-required task 는 `supports_worktree=false` 로 C5 에서 차단되어야 한다. |

다른 provider 를 Codex 와 같은 trusted/full-access runtime 으로 승격하는 PR 은 이 표,
[`./security.md`](./security.md) 의 C7 결정, command builder, deterministic tests, real
integration evidence 를 같은 변경으로 갱신해야 한다. SSOT 없이 provider flag 만
강하게 만드는 변경은 C4/C7 경계 위반이다.

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

