# Provider Runtime / Adapter SSOT: Public Migration Status

[Back to provider-runtime.md](../provider-runtime.md)

## 0. Public Migration Status

RIID-4651 에서 public `riido-daemon` 으로 이동한 구현 범위는 `internal/agentbridge` 루트 package 다. 이 package 는 stdlib-only provider-neutral 도메인으로, `Adapter` port, `RawEvent` / `Parser`, `RunState`, reducer, telemetry parser, tool start gate 를 포함한다.

아직 이 slice 에 포함하지 않은 구현:

- task DB/project/mwsd/local API, server/control-plane/infra/secret/state files

The Figma onboarding planning screen (`v.1.22 AI Agent`, `node-id=42-3014`) is
outside C4 ownership. C4 may report runtime detection and liveness used by the
client's onboarding runtime-choice screen (`node-id=137-6746`), and it may
execute a task using the instruction that SaaS assigns later. C4 does not own
the `감지됨` / `감지 안 됨` labels, radio enabled/disabled state, row dimming,
onboarding fixture catalog, the `리도` / `영실` / `홍도` / `지원` onboarding
fixture rows, the `직접 설정` row, disabled-next presentation, preview skeleton
or popover, direct-setting form composition, workspace selector, no-runtime
skip branch, scroll affordances, two-line ellipsis behavior, or onboarding
fixture copy. Figma planning node `432:46849` may ask clients to collect an
agent draft/configuration before runtime and workspace selection, but C4 still
does not start a provider from that draft. C4 receives only the final
SaaS-authorized runtime/model/instruction snapshot.

The Figma web onboarding section (`node-id=236-29749`) is also outside C4
ownership. Its macOS app download CTA can lead a user to a desktop artifact, but
it is not a provider runtime command and does not authorize C4 to bundle,
download, install, or update Claude/Codex/OpenClaw/Cursor CLIs. Its sign-up,
terms consent, member invite, Windows waitlist, marketing-consent, chat
animation, and progress-bar reference facts belong to client/auth/team/product
surfaces until a separate SSOT promotes executable daemon behavior.

The daemon projection of Figma AI Agent boundaries is
[`../30-architecture/figma-ai-agent-daemon-boundary.md`](../30-architecture/figma-ai-agent-daemon-boundary.md).
C4 follows that projection only for provider-runtime execution inputs; it does
not turn client/control-plane fixture or presentation facts into adapter
behavior.

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

A-51 부터 Claude real CLI integration gate 는 `ResultCompleted` 와 함께 daemon 이
선택한 workdir 안의 expected file artifact 를 확인한다. 이 gate 는
`AGENTBRIDGE_INTEGRATION=1` 과 local Claude Code auth/runtime 이 준비된 operator
environment 에서만 실행된다. Claude adapter 는 process `Dir` 를 task workdir 로
전달하고, integration gate 는 `PermissionModeAcceptEdits` 를 사용해 edit/write tool
승인을 명시적으로 열어둔다. Gate 가 skip 된 경우에는 filesystem side-effect 가
검증된 것이 아니다.

RIID-4659 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/provider/codex` 다. 이 package 는 Codex CLI 를 번들하지 않고,
`codex --sandbox danger-full-access app-server --listen stdio://` command
construction, daemon-owned full-access runtime selection, JSONL parser, raw event
translator, JSON-RPC protocol driver, pending request actor, approval response path
를 소유한다. Codex app-server 자체는 사용자의 기존 Codex auth store 를 쓸 수 있다.
Workdir 은 daemon-selected 작업/evidence root 이지만 filesystem sandbox boundary 가
아니며, provider 는 local full-access automation 으로 실행된다. 이 full-access 는
Codex 의 provider default 또는 caller 입력에 맡기는 값이 아니라 C4 adapter 가
고정하는 harness-managed launch envelope 다. real Codex CLI execution 은
`AGENTBRIDGE_INTEGRATION=1` 로 opt-in 된 경우에만 검증한다. OpenClaw/Cursor
adapter, supervisor polling loop, server/task DB/project/mwsd adapter 는 RIID-4659
당시 후속 migration slice 가 맡는 것으로 남겼다.

A-57 부터 Codex real CLI integration gate 는 `ResultCompleted` 와 함께 daemon 이
선택한 workdir 안의 expected file artifact 를 확인한다. 이 gate 는
`AGENTBRIDGE_INTEGRATION=1` 과 local Codex auth/runtime 이 준비된 operator
environment 에서만 실행된다. Codex adapter 는
`codex --sandbox danger-full-access app-server --listen stdio://` 를 실행한다.
이 gate 는 full-access
provider runtime 이 daemon-selected workdir 안에 expected artifact 를 실제로 만들 수
있는지를 확인한다. Gate 가 skip 된 경우에는 filesystem side-effect 가 검증된 것이
아니다.

Codex runtime model catalog 는 host Codex config 의 `model` 값을
runtime-scoped opaque `model_id` 로 보고할 수 있다. 이 값은 provider credential 이
아니며, control-plane 이 agent 설정/assignment snapshot 에 저장하는 model 선택의
입력이다. Daemon 은 OpenAI/ChatGPT auth token, account identity, API key,
team id, Open API key 로 model catalog 를 추론하지 않는다.

Control-plane fallback catalog id 인 `codex-default`, `claude-default`,
`openclaw-default`, `cursor-auto`, `runtime-default` 는 client read-model 의
선택/표시용 sentinel 이다. 이 값은 assignment metadata 에는 원문 그대로 보존되지만,
provider native process 에 넘기는 model override 가 아니다. `saasplane` 은 이런
synthetic default 를 `StartRequest.Model=""` 로 정규화해야 한다. Provider adapter
가 `--model` 같은 native flag 로 변환할 수 있는 값은 runtime 이 실제 host/provider
catalog 에서 보고했거나 control-plane 이 provider-native 값으로 승인한
non-synthetic `model_id` 뿐이다.

RIID-4660 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/provider/openclaw` 다. 이 package 는 OpenClaw CLI 를 번들하지 않고,
external executable detection, calendar-version gate, `openclaw agent --local --json`
command construction, mandatory session id resolution, JSON/NDJSON parser, raw event
translator 를 소유한다. real OpenClaw CLI execution 은 `AGENTBRIDGE_INTEGRATION=1` 로
opt-in 된 경우에만 검증한다. Cursor adapter, supervisor polling loop, server/task
DB/project/mwsd adapter 는 RIID-4660 당시 후속 migration slice 가 맡는 것으로
남겼다.

OpenClaw `--session-id` 는 provider-native identifier 다. 이미 provider 가 발급한
`ResumeSessionID` 는 그대로 보존하지만, first-run 의 `TaskID` fallback 은 raw Riido
component id 를 그대로 넘기지 않는다. Riido component id 는 `-4ck...` 처럼 하이픈으로
시작할 수 있고 OpenClaw 가 이를 거부할 수 있으므로, adapter 는 task id 로부터
`riido-<sanitized-task-slug>-<short-hash>` 형태의 deterministic provider-safe
session id 를 파생한다. Raw task id 는 run metadata/workdir path 의 SSOT 로 남고,
OpenClaw session id 만 native process boundary 에서 별도로 정규화된다.

A-42 live E2E 에서 OpenClaw 는 `assignment -> daemon -> provider process ->
provider text result -> SaaS completed thread` 경로를 통과했다. 하지만 provider 가
파일 생성 완료라고 응답해도 daemon workdir 에 expected artifact 가 없을 수 있음이
확인되었다. 따라서 OpenClaw text completion 은 filesystem side-effect evidence 가
아니다. OpenClaw 파일 작성 capability 는 별도 provider capability / permission gate
로 검증되기 전까지 PASS 조건으로 선언하지 않는다.

A-48 부터 OpenClaw real CLI integration gate 는 `ResultCompleted` 와 함께 daemon 이
선택한 workdir 안의 expected file artifact 를 확인한다. 이 gate 는
`AGENTBRIDGE_INTEGRATION=1` 과 지원 OpenClaw version 이 모두 충족된 operator
environment 에서만 실행된다. Gate 가 skip 된 경우에는 filesystem side-effect 가
검증된 것이 아니며, SaaS thread completion 만으로 파일 산출을 증명하지 않는다.
현재 OpenClaw CLI 의 `agent` surface 는 per-run `--workspace` / `--cwd` 를 제공하지
않고 `agents add --workspace` 또는 `setup --workspace` 로 사전 구성된 workspace 를
사용하므로, daemon-selected task workdir 을 요구하는 task 는 C5
`required_surfaces=["worktree"]` 로 pre-submit 차단되어야 한다. Runtime capability
reconciliation 은 OpenClaw `SupportsWorktree=false`, Claude/Codex/Cursor
`SupportsWorktree=true` 로 노출한다.

RIID-4661 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/provider/cursor` 다. 이 package 는 Cursor Agent CLI 를 번들하지 않고,
root-print / agent-subcommand / legacy-chat launch profile selection, `--yolo` unsafe
bypass policy gate, daemon task workdir 에 대한 headless workspace trust acknowledgement
`--trust`, external executable detection, stream-json parser, raw event translator 를
소유한다. Cursor `--trust` 는 Cursor Agent 가 daemon 이 선택한 작업 디렉터리에서
interactive trust prompt 로 멈추지 않게 하는 확인값이다. 이는 tool auto-approval 을
켜는 `--yolo` unsafe bypass surface 가 아니며, `--yolo` 는 계속 C7 policy gate 를
통과한 경우에만 사용할 수 있다. real Cursor Agent CLI execution 은
`AGENTBRIDGE_INTEGRATION=1` 로 opt-in 된 경우에만 검증한다. supervisor polling loop,
server/task DB/project/mwsd adapter 는 RIID-4661 당시 후속 migration slice 가 맡는
것으로 남겼다.
