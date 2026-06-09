# AI Agent Runtime Lifecycle Review

> Review date: 2026-06-08
>
> Scope: `riido-daemon`, `riido-desktop`, `riido-client`,
> `riido-control-plane`, and related local Riido repositories.

이 문서는 댓글 기반 AI Agent 실행에서 관찰된 daemon 중복 실행, provider CLI
runtime 관리, SSE progress streaming, stop/cancel lifecycle, provider CLI 감지
불일치를 한 번에 묶은 통합 리뷰 기록이다. 구현 변경 전 현재 코드와 로컬 실행
상태에서 확인한 fact만 기록한다.

## 1. Local Repository Scope

확인된 로컬 Riido 관련 repository:

| Path | Status |
| --- | --- |
| `/Users/work/work/riido-official/riido-daemon` | current repo, branch `JYM-ai-get-done` |
| `/Users/work/work/riido-official/riido-desktop` | related desktop launcher |
| `/Users/work/work/riido-official/riido-client` | related web/client UI |
| `/Users/work/work/riido-official/riido-control-plane` | related AI agent SaaS/control-plane |
| `/Users/work/work/riido-official/riido-api-server` | sibling Riido repo |
| `/Users/work/work/riido-official/riido-contracts` | sibling contracts repo |
| `/Users/work/work/riido-official/riido-infra` | sibling infra repo |
| `/Users/work/work/riido-official/riido-mcp-server` | sibling MCP repo |
| `/Users/work/work/riido-official/riido-plugin` | sibling plugin repo |
| `/Users/work/work/projects/riido-iphone-app` | extra local Riido app repo |

Non-primary / not used as source of truth in this review:

- `/Users/work/work/oss/VibeVoice`
- `/Users/work/work/projects/riido-meeting-stt` if present locally; it was not
  part of the Git-backed daemon/client/control-plane lifecycle evidence used
  here.
- retired/private source names such as `riido_daemon_private` or
  `riido-daemon-private`

`riido-contracts` appears both as a local sibling repository in this workspace
and as the Go module dependency consumed by `riido-daemon`. This review uses the
checked local source only for repository inventory; daemon behavior evidence
comes from the currently checked out daemon/control-plane/client/desktop code.

## 1.1 Review Method

This was a code-reading and local-process-state review. No runtime behavior was
changed while collecting the evidence. The investigation walked the following
paths:

- Desktop daemon install/launch/auto-launch code in `riido-desktop`
- Local daemon process/singleton/socket/status code in `riido-daemon`
- Provider runtime actor, session, process execution, and SaaS reporter paths in
  `riido-daemon`
- Assignment store, AI Agent client read model, runtime snapshot, and SSE server
  paths in `riido-control-plane`
- Client stop mutation, task thread rendering, onboarding runtime selection, and
  SSE hook paths in `riido-client`

The important distinction throughout the review is that a user-visible AI Agent
run spans several independent systems. A successful HTTP response from one layer
does not imply the other layers have reached the same state.

## 2. Executive Summary

현재 문제는 개별 버그 하나가 아니라 lifecycle ownership 분리가 맞지 않는 구조적
문제다. `desktop daemon launcher`, `daemon singleton`, `runtime actor`,
`provider CLI process`, `control-plane assignment`, `client SSE active_stream`이
하나의 원자적 lifecycle로 묶여 있지 않다.

그 결과 다음 현상이 같은 계열에서 발생한다.

- Desktop app을 껐다 켤 때 daemon start process가 중복으로 누적된다.
- 기존 daemon이 live socket을 잃거나 lock 대기 process가 남을 수 있다.
- 댓글마다 provider CLI가 새 process로 cold start된다.
- provider raw text delta가 `riido_log` progress로 직접 전송되어 SSE에 한 글자,
  두 글자 단위 line이 쌓인다.
- Stop을 눌러도 late `riido_log`가 stopped thread를 다시 running으로 되살릴 수
  있다.
- local CLI 감지는 성공하지만 UI/SaaS runtime 상태는 stale하거나 다른 daemon
  instance를 보고 있을 수 있다.

## 3. Current Local Runtime Evidence

현재 머신에서 확인된 provider CLI:

| Provider | Local executable | Detection result |
| --- | --- | --- |
| Codex | `/Users/work/.local/bin/codex` | detected, `codex-cli 0.137.0` |
| Claude | `/Users/work/.local/bin/claude` | detected, `2.1.168 (Claude Code)` |
| Cursor | `cursor-agent` not found | missing |
| OpenClaw | `openclaw` not found | missing |

설치된 Desktop daemon binary:

- `/Users/work/riido-ai-agent-development/electron-user-data/ai-agent-daemon/bin/riido`
- daemon version observed through status: `v0.0.14`

중요한 로컬 상태:

- `riido daemon status`는 Codex/Claude를 detected로 보고했다.
- 동시에 같은 desktop dev parent 아래 daemon start process가 두 개 확인됐다.
- 두 process는 같은 `daemon.pid`, `daemon.lock`, `daemon.log`를 사용했다.
- pid file은 한 process만 가리키며, 다른 process는 singleton lock 대기 상태일
  가능성이 높다.

## 4. Desktop Launch Creates Duplicate Daemon Processes

Desktop launcher는 daemon을 detached child process로 실행하고 즉시 unref한다.

Evidence:

- `riido-desktop/src/modules/daemonLauncher.ts:593`
  - `buildDaemonLaunchOptions` default `detached = true`
- `riido-desktop/src/modules/daemonLauncher.ts:631`
  - `spawn(...)`
- `riido-desktop/src/modules/daemonLauncher.ts:633`
  - `child.unref()`
- `riido-desktop/src/main.ts:604`
  - `setupDaemonAutoLaunch()`
- `riido-desktop/src/main.ts:606`
  - `before-quit`에서 auto-launch controller stop만 수행

현재 desktop은 앱 종료 시 daemon process 자체를 명확히 stop하지 않는다. 앱 실행 중
30초 auto launch loop 또는 credential refresh path가 반복 실행되면 daemon start
attempt가 누적될 수 있다.

Detailed failure sequence:

1. Desktop starts and schedules the daemon ensure loop.
2. The ensure loop checks install/status and decides daemon should be running.
3. It spawns `riido daemon start --foreground ...` as a detached process.
4. Desktop does not keep the child as a managed process handle because it calls
   `unref()`.
5. On app quit, Desktop stops the auto-launch loop, not necessarily the daemon.
6. On the next app start, the same ensure path can spawn another foreground
   daemon start attempt.

The second process may not become the serving daemon. It can simply wait on the
singleton lock. From a process list, however, it still looks like another daemon
process. That is why "duplicate process" and "orphan process" are both valid
symptoms of the same launch model.

## 5. Daemon Singleton Lock Waits Instead Of Failing Fast

Daemon foreground start는 singleton lock을 잡는다.

Evidence:

- `riido-daemon/cmd/riido/daemon.go:192`
  - `runDaemonStartForeground`
- `riido-daemon/cmd/riido/daemon.go:204`
  - `c9lock.AcquireFile(ctx, flags.lockFile)`
- `riido-daemon/internal/lock/filelock.go:23`
  - `AcquireFile` waits until lock can be acquired
- `riido-daemon/internal/lock/filelock.go:35`
  - 10ms ticker retry loop
- `riido-daemon/internal/lock/filelock_unix.go:10`
  - `LOCK_EX | LOCK_NB`

즉 두 번째 daemon start는 "이미 daemon이 떠 있음"으로 종료되지 않는다. lock을 얻을
때까지 대기 process로 살아남는다. desktop이 이 process를 detached/unref로 띄우면
사용자에게는 보이지 않는 대기 process가 남는다.

This is not a lock implementation failure. The lock is doing what it was asked
to do: wait until the exclusive lock becomes available. The product problem is
that Desktop uses this wait-oriented foreground command as an "ensure running"
operation. An ensure operation should either return the current daemon status,
fail fast with "already running", or own the waiting child process so it can be
cancelled when Desktop exits.

## 6. Socket, Lock, PID Boundaries Are Not Atomic

Desktop은 userData 아래에 install root, log, pid, lock path를 만든다.

Evidence:

- `riido-desktop/src/modules/daemonLauncher.ts:351`
  - `defaultDaemonInstallRoot() = app.getPath('userData')/ai-agent-daemon`
- `riido-desktop/src/modules/daemonLauncher.ts:357`
  - daemon log path
- `riido-desktop/src/modules/daemonLauncher.ts:360`
  - daemon pid path
- `riido-desktop/src/modules/daemonLauncher.ts:363`
  - daemon lock path

하지만 launch args에는 `--socket`이 없다.

Evidence:

- `riido-desktop/src/modules/daemonLauncher.ts:711`
  - args: `daemon start --foreground --log-file --pid-file --lock-file`

Daemon은 `--socket`이 없으면 default socket을 사용한다.

Evidence:

- `riido-daemon/cmd/riido/daemon.go:175`
  - missing socket uses `defaultAgentDaemonSocket`
- `riido-daemon/cmd/riido/daemon.go:843`
  - default socket path construction

그리고 daemon serve 시작 시 socket file을 무조건 제거한다.

Evidence:

- `riido-daemon/cmd/riido/daemon.go:326`
  - `serveAgentDaemon`
- `riido-daemon/cmd/riido/daemon.go:328`
  - `_ = os.Remove(flags.socket)`

이 조합은 위험하다. 서로 다른 launcher/userData/lock path를 가진 daemon들이 같은
default socket을 공유할 수 있고, 새 daemon이 기존 live socket을 unlink해서 기존
daemon을 살아 있지만 접근 불가능한 orphan으로 만들 수 있다.

Concrete orphan sequence:

1. Daemon A starts with lock path A and default socket S.
2. Daemon B starts with lock path B but also default socket S.
3. Because the lock paths differ, B is not protected by A's lock.
4. B enters `serveAgentDaemon` and removes S as a "stale" socket.
5. A is still alive, but its socket path has been unlinked.
6. Status/stop calls that dial S now reach B or fail, while A can keep running
   without a reachable control socket.

This is why socket, pid file, lock file, install root, and device/runtime
identity must be treated as one lifecycle identity. Managing only one of those
paths does not guarantee daemon singleton behavior.

## 7. Daemon Stop Is Best-Effort

Desktop `stopDaemonIfRunning`은 먼저 running daemon status를 읽고 그 status에서
socket/pid를 뽑아 stop한다.

Evidence:

- `riido-desktop/src/modules/daemonLauncher.ts:560`
  - comment: `daemon stop` requires `--socket` and/or `--pid-file`
- `riido-desktop/src/modules/daemonLauncher.ts:563`
  - `stopDaemonIfRunning`
- `riido-desktop/src/modules/daemonLauncher.ts:567`
  - `readRunningDaemonStatus`
- `riido-desktop/src/modules/daemonLauncher.ts:572`
  - args `daemon stop`

문제는 status가 현재 reachable socket에만 의존한다는 점이다. socket이 다른 daemon에
의해 교체되었거나 pid file이 stale하면 stop은 orphan daemon까지 포괄하지 못한다.

This also explains why "stop daemon before replacement" can still miss a
process. The stop path first has to discover the correct daemon. If discovery is
already pointed at a different socket owner, or if the pid file was overwritten
by a later start attempt, the old daemon is outside the stop transaction.

## 8. Provider Runtime Model

Daemon은 built-in provider adapter마다 runtime actor를 만든다.

Evidence:

- `riido-daemon/cmd/riido/daemon.go:640`
  - `builtinDaemonAdapters`
- `riido-daemon/cmd/riido/daemon.go:454`
  - `newDaemonRuntimeActor`
- `riido-daemon/cmd/riido/daemon.go:463`
  - `MaxConcurrent: 1`

`MaxConcurrent`는 provider별 동시 실행 제한이다. provider CLI process reuse를
의미하지 않는다.

Branch/version note:

- The reviewed `JYM-ai-get-done` daemon branch keeps runtime actor concurrency
  at `MaxConcurrent: 1`.
- `origin/main` later changed the default runtime max concurrency to 4 through
  `RIIDO_RUNTIME_MAX_CONCURRENT` in `cmd/riido/daemon_config.go`.
- That mainline change increases throughput but can also amplify Codex account
  rate-limit pressure because each accepted run still starts its own Codex
  app-server process and turn.

The runtime actor is the in-memory owner for scheduling and capability status.
It is not a persistent provider CLI daemon. It may serialize or bound work for a
provider, but each accepted task becomes a run-scoped session with its own
process and state reducer. Therefore runtime actor health does not imply there
is a long-lived Codex/Claude process that can continue across Desktop or daemon
restarts.

## 9. Every Task Starts A New Provider CLI Process

Runtime actor는 submit마다 adapter `BuildStart`로 command를 만들고 session을 새로
start한다.

Evidence:

- `riido-daemon/internal/agentbridge/runtimeactor/runtimeactor.go:375`
  - `agentbridge.StartRequest`
- `riido-daemon/internal/agentbridge/runtimeactor/runtimeactor.go:389`
  - `adapter.BuildStart(startReq)`
- `riido-daemon/internal/agentbridge/runtimeactor/runtimeactor.go:418`
  - `session.Start(...)`
- `riido-daemon/internal/agentbridge/session/session.go:135`
  - `cfg.Process.Start(ctx, cfg.Spawn)`

따라서 댓글 하나가 assignment/task 하나로 들어오면 provider CLI process가 새로
뜬다. long-lived Codex/Claude CLI 하나를 계속 재사용하는 구조가 아니다.

This matters for performance and cancellation:

- Each comment-triggered run pays provider CLI startup and handshake cost.
- Any in-flight provider child process is owned by the daemon process that
  spawned it.
- If another daemon instance starts, it does not inherit the original daemon's
  in-flight session state.
- If duplicate daemons poll the same control-plane, SaaS assignment fencing must
  be strong enough to prevent double claim/report. The local process model alone
  does not provide that guarantee.

## 10. Codex And Claude Spawn Shapes

Codex:

- `riido-daemon/internal/provider/codex/command.go:87`
  - `--sandbox danger-full-access`
- `riido-daemon/internal/provider/codex/command.go:89`
  - `app-server`
- `riido-daemon/internal/provider/codex/command.go:91`
  - `--listen stdio://`

Claude:

- `riido-daemon/internal/provider/claude/command.go:117`
  - `-p`
- `riido-daemon/internal/provider/claude/command.go:119`
  - `--output-format stream-json`
- `riido-daemon/internal/provider/claude/command.go:120`
  - `--input-format stream-json`
- `riido-daemon/internal/provider/claude/command.go:121`
  - `--verbose`
- `riido-daemon/internal/provider/claude/command.go:122`
  - `--permission-mode ...`

Daemon이 새로 시작하면 in-memory runtime actor/session state는 사라진다. 중복 daemon
상태에서는 여러 daemon이 같은 SaaS assignment source를 poll/report할 위험이 생긴다.

The current spawn shapes also mean provider-specific runtime failures surface as
daemon session failures, not as persistent CLI health failures. A provider can
be detected and still fail on a run because the per-run command shape, cwd, env,
MCP config, prompt placement, permission mode, or repo context is wrong.

## 11. SSE Progress Has No Effective Batch Boundary

SSOT 문서는 client thread progress가 parsed/bounded batch로
`POST /v1/agents/{agent_id}/thread-progress`에 올라가는 모델을 말한다.

Evidence:

- `riido-daemon/docs/migration/daemon.md:706`
  - task-thread progress should be reported as bounded parsed batches
- `riido-control-plane/docs/20-domain/saas-control-plane.md:92`
  - runtime progress is ingested as bounded daemon batches on `/thread-progress`
- `riido-control-plane/docs/20-domain/ai-agent-client-api.md:460`
  - daemon progress ingest accepts parsed `<riido_log>...<end>` batches

하지만 실제 daemon은 provider `EventTextDelta`를 standard assignment event
`riido_log`로 바꿔 `/events`에 보낸다.

Evidence:

- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:461`
  - `ReportEvent`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:828`
  - `EventTextDelta`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:836`
  - `req.EventType = EventRiidoLog`

Control-plane은 각 `riido_log`를 `agent_thread_progress`로 fanout한다.

Evidence:

- `riido-control-plane/internal/riidoaiserver/server.go:1610`
  - `handleAgentEvent`
- `riido-control-plane/internal/riidoaiserver/server.go:1623`
  - assignment event record
- `riido-control-plane/internal/riidoaiserver/server.go:1628`
  - AI Agent read model recorder
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1466`
  - `EventRiidoLog` creates progress line
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1485`
  - `AgentThreadProgressEvent`

SSE writer는 event마다 즉시 write/flush한다.

Evidence:

- `riido-control-plane/internal/riidoaiserver/server.go:1251`
  - SSE response setup
- `riido-control-plane/internal/riidoaiserver/server.go:1271`
  - write each live event
- `riido-control-plane/internal/riidoaiserver/server.go:1274`
  - flush each event
- `riido-control-plane/internal/riidoaiserver/server.go:2120`
  - `writeAIAgentClientSSE`

Client는 persisted thread lines와 live stream event lines를 그대로 합쳐 표시한다.

Evidence:

- `riido-client/src/components/domain/aiAgentTask/AgentThreadCard.tsx:79`
  - `progressMessages`
- `riido-client/src/components/domain/aiAgentTask/AgentThreadCard.tsx:80`
  - `thread.lines`
- `riido-client/src/components/domain/aiAgentTask/AgentThreadCard.tsx:84`
  - live `streamEvents`

결론: 한글 한 글자/두 글자씩 줄바꿈되는 현상은 provider raw text delta가
batch/semantic progress boundary 없이 progress line으로 처리되기 때문이다.

The intended semantic split appears to be:

- Provider final assistant text or answer text is run output.
- `<riido_log>...<end>` is progress telemetry.
- Parsed progress telemetry is batched and posted to `/thread-progress`.
- Client SSE receives `agent_thread_progress` as meaningful progress lines.

The implemented path collapses the first two surfaces. Raw assistant text delta
is treated as progress. If Codex/Claude emits small chunks, every chunk becomes a
separate progress line, a separate control-plane event, and potentially a
separate client-rendered line. This is why the bug is visible as Korean text
splitting rather than just higher network volume.

## 12. Request Storm And Slowness

Daemon은 raw text delta마다 HTTP POST를 보낸다.

Evidence:

- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:461`
  - `ReportEvent`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:470`
  - `postAgentEvent`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:524`
  - `postJSON`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:525`
  - per-request timeout

Client에서도 이 문제가 이미 주석으로 인정되어 있다.

Evidence:

- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:231`
  - SSE invalidation scoped/throttled
- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:232`
  - daemon heartbeat events and one event per assistant text delta
- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:233`
  - refetch storm caused 503s
- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:240`
  - `STREAM_INVALIDATE_MIN_INTERVAL_MS = 800`

Client throttle는 증상 완화일 뿐이다. 근본 원인은 daemon/control-plane 쪽 progress
batching 부재와 raw text delta forwarding이다.

The slow/failure symptom likely has three reinforcing causes:

1. Provider CLI cold start for every comment assignment.
2. Synchronous HTTP reporting for high-frequency text deltas.
3. Missing real repository/worktree binding, causing the provider CLI to operate
   without the target codebase context.

These causes compound. A run can start slowly because of CLI startup, then spend
extra time discovering that the expected files are absent, while also flooding
control-plane/client with text-delta progress events.

Current branch follow-up:

- `JYM-ai-get-done` coalesces provider `EventTextDelta` in supervisor before it
  reaches the reporter (`TextFlushBytes=256`, `TextFlushInterval=200ms` by
  default).
- That reduces request storm volume, but `saasplane` still maps the coalesced
  assistant text to `EventRiidoLog`, so assistant answer text and progress log
  semantics remain collapsed.
- `origin/main` changes this again: raw text deltas are accumulated and posted as
  a special evolving partial body progress line, while final completion falls
  back to the accumulated body when Codex reports an empty result.
- Neither branch has a true repo-binding step yet; the LLM still runs inside the
  generated isolated workdir, not the actual project checkout.

## 13. Workdir And Repo Binding Are Insufficient

SaaS assignment를 daemon task request로 변환할 때 metadata의 `workspace_id`가 실제
local repo path가 아니다.

Evidence:

- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:762`
  - `taskRequestFromAssignment`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:770`
  - `"workspace_id": firstNonEmpty(assignment.ComponentID, assignment.TaskID)`

Workdir adapter는 isolated task tree만 만든다.

Evidence:

- `riido-daemon/internal/workdir/workdir.go:232`
  - `Prepare`
- `riido-daemon/internal/workdir/workdir.go:249`
  - task root path from workspace/task/run ids

따라서 댓글에서 LLM CLI를 실행할 때 실제 `riido-daemon`, `riido-client`,
`riido-desktop`, `riido-api-server` repository 안에서 작업한다는 보장이 없다.
LLM CLI가 빈 task workdir 또는 repo context 없는 위치에서 실행되면 느리거나 실패할
가능성이 높다.

This is a separate problem from CLI detection. A provider can be correctly
detected and still perform badly if the run's `cwd` and prompt metadata do not
point at the intended repository. "Codex is installed" only proves the executable
can be found and version-probed; it does not prove the assignment has a usable
worktree.

The current branch added explicit no-repo guidance in the generated provider
config, but that is only a fail-fast instruction to the LLM. It does not mount,
clone, or checkout the selected repository. If the assignment prompt says a
repository exists but the daemon workdir contains only generated runtime config,
the agent has task metadata but no files to inspect or edit.

This explains a common "context did not switch" symptom: the control-plane
prompt can change to a different task snapshot, but the process `cwd` remains an
empty generated run directory. From the provider CLI's perspective, every coding
assignment has the same filesystem shape unless a repo/worktree binding step
materializes the actual project into that run workdir.

## 14. Stop/Cancel Is Not ACID

Stop button flow, as observed from code:

1. Client calls `tasks/{taskId}/stop` and then invalidates queries.
2. Control-plane first writes the AI Agent client read model as stopped.
3. Control-plane then mutates the separate assignment store to cancelling or
   cancelled.
4. Daemon does not receive an immediate process-kill RPC. It observes
   cancellation through polling / `WatchCancellation`.
5. Runtime cancel does not wait for provider process termination. It signals the
   session cancel path, and actual kill/report happens later through the session
   and supervisor.

Client stop mutation:

- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:375`
  - `stopTask`
- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:376`
  - `stopMutation.mutateAsync`
- `riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:383`
  - query invalidation

Client SSE stream:

- `riido-client/src/components/domain/aiAgentTask/AgentTaskThreadPanel.tsx:39`
  - `activeStreamHref`
- `riido-client/src/components/domain/aiAgentTask/AgentTaskThreadPanel.tsx:65`
  - effect controls stream
- `riido-client/src/components/domain/aiAgentTask/AgentTaskThreadPanel.tsx:66`
  - stop only if no `activeStreamHref`
- `riido-client/src/components/domain/aiAgentTask/AgentTaskThreadPanel.tsx:71`
  - start stream if active

Control-plane stop handler first updates AI Agent client read model and then
cancels durable assignment.

Evidence:

- `riido-control-plane/internal/riidoaiserver/server.go:1071`
  - `handleAIAgentClientStopTask`
- `riido-control-plane/internal/riidoaiserver/server.go:1083`
  - `s.aiAgent.StopAIAgentTask`
- `riido-control-plane/internal/riidoaiserver/server.go:1088`
  - `cancelAIAgentAssignmentFromAction`
- `riido-control-plane/internal/riidoaiserver/server.go:99`
  - assignment cancellation helper

AI Agent client read model marks thread stopped.

Evidence:

- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:997`
  - `StopAIAgentTask`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1017`
  - `WorkStatus: idle`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1018`
  - `AssignmentState: stopped`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1029`
  - mark threads stopped

Assignment store moves assignment to cancelling or cancelled.

Evidence:

- `riido-control-plane/internal/riidoaiserver/store.go:717`
  - terminal/cancelling guard
- `riido-control-plane/internal/riidoaiserver/store.go:721`
  - `AssignmentCancelling`
- `riido-control-plane/internal/riidoaiserver/store.go:727`
  - queued assignment becomes cancelled

Daemon receives cancellation through poll/cancel watcher, not immediate push.

Evidence:

- `riido-control-plane/internal/riidoaiserver/store.go:811`
  - cancelling assignment
- `riido-control-plane/internal/riidoaiserver/store.go:821`
  - `PollCancel`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:431`
  - `WatchCancellation`
- `riido-daemon/internal/agentbridge/supervisor/supervisor.go:998`
  - `forwardCancellation`

Runtime actor cancel only enqueues session cancel and returns.

Evidence:

- `riido-daemon/internal/agentbridge/runtimeactor/runtimeactor.go:459`
  - `handleCancel`
- `riido-daemon/internal/agentbridge/runtimeactor/runtimeactor.go:468`
  - `session.Cancel`
- `riido-daemon/internal/agentbridge/runtimeactor/runtimeactor.go:566`
  - public `Cancel`

Session cancellation emits cancellation and later kills process.

Evidence:

- `riido-daemon/internal/agentbridge/session/session.go:345`
  - cancellation select
- `riido-daemon/internal/agentbridge/session/session.go:354`
  - `emitAndTerminate(EventCancellation)`
- `riido-daemon/internal/agentbridge/session/session.go:396`
  - `proc.Kill`

Unix process kill sends SIGTERM and immediately SIGKILL.

Evidence:

- `riido-daemon/internal/process/processexec/processexec_unix.go:22`
  - `SIGTERM`
- `riido-daemon/internal/process/processexec/processexec_unix.go:23`
  - `SIGKILL`
- `riido-daemon/internal/process/processexec/processexec_unix.go:24`
  - single PID fallback kill

결론: stop API success, assignment cancel state, local runtime cancellation, provider
process termination, final event reporting, client SSE active stream clearing이 하나의
transaction처럼 움직이지 않는다.

The precise inconsistency is:

- Assignment state can be `cancelling`.
- Client thread read model can already be `stopped`.
- Daemon runtime can still have an in-flight session.
- Provider child process can still have buffered stdout/stderr.
- SSE can still be connected because `active_stream` is derived from the current
  thread projection.

No single layer fences all later events by the accepted stop operation. That is
why this is more than a UI issue.

## 15. Late Progress Can Re-Activate A Stopped Thread

Stop 이후 provider/daemon에서 늦은 `riido_log`가 들어오면, control-plane read model은
assignment id로 기존 thread를 찾는다.

Evidence:

- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1438`
  - `taskThreadForAssignmentLocked`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1727`
  - lookup by assignment id
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1735`
  - returns matching thread without terminal guard

그 다음 `riido_log`는 무조건 running progress event를 만든다.

Evidence:

- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1466`
  - `EventRiidoLog`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1485`
  - `AgentThreadProgressEvent`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1493`
  - `WorkStatus: running`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1494`
  - `AssignmentState: running`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1495`
  - `CommentKind: runtime_progress`

Append path는 thread state를 event state로 덮는다.

Evidence:

- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:2116`
  - `appendThreadProgressLocked`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:2127`
  - overwrite `WorkStatus`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:2128`
  - overwrite `AssignmentState`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:2129`
  - overwrite `CommentKind`

Active stream selection treats running/queued/stopping as active.

Evidence:

- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:2978`
  - `taskThreadHasActiveStream`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:2980`
  - queued/running/stopping are active

결론: stop 이후 late progress가 stopped thread를 다시 active stream으로 살릴 수 있다.
사용자가 관찰한 "중지해도 SSE가 계속 동작"과 직접 연결된다.

There is a second active-stream continuation path immediately after stop. Thread
collection reads reconcile active thread projections from durable assignment
projection:

- `riido-control-plane/internal/riidoaiserver/server.go:86`
  - `reconcileAIAgentTaskThreadProjections`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:752`
  - `ReconcileAIAgentActiveThreadProjections`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1781`
  - `assignmentStateCanRepairTaskThread`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1783`
  - `AssignmentCancelling` can repair a task thread
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1894`
  - `AssignmentCancelling`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1896`
  - maps to `AgentAssignmentStateStopping`

`AgentAssignmentStateStopping` is itself considered active:

- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:2978`
  - `taskThreadHasActiveStream`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:2980`
  - `AgentAssignmentStateStopping` returns active

So stop can keep a stream open even without the late-progress resurrection bug:
the durable assignment is `cancelling`, the client projection becomes
`stopping`, and `stopping` still emits `active_stream`. Late `riido_log` then
makes the situation worse by changing the same thread back to running progress.

## 16. Provider Session Reducer Has Local Terminal Guard But Read Model Does Not

Daemon session reducer 자체는 cancellation이 later result보다 우선한다.

Evidence:

- `riido-daemon/internal/agentbridge/reducer.go:123`
  - `EventCancellation`
- `riido-daemon/internal/agentbridge/reducer.go:124`
  - terminate cancelled
- `riido-daemon/internal/agentbridge/reducer_test.go:108`
  - cancellation preempts later events test

하지만 이 terminal invariant가 control-plane read model event ingestion까지 전파되지
않는다. control-plane은 late `riido_log`를 assignment/thread terminal state와 fencing
해서 거부하지 않는다.

The daemon-side reducer protects a single local session state. The control-plane
read model is a separate derived model and needs its own fence. Without that
fence, a locally cancelled session can still be followed by network-delivered
events that mutate the SaaS/client-visible state.

## 17. CLI Detection Logic And UI State Are Different Surfaces

Provider detect code:

- `riido-daemon/internal/provider/codex/detect.go:10`
  - `RIIDO_CODEX_PATH`
- `riido-daemon/internal/provider/codex/detect.go:13`
  - `ResolveExecutable`
- `riido-daemon/internal/provider/claude/detect.go:12`
  - `RIIDO_CLAUDE_PATH`
- `riido-daemon/internal/provider/claude/detect.go:22`
  - `ResolveExecutable`
- `riido-daemon/internal/provider/cursor/detect.go:11`
  - `RIIDO_CURSOR_PATH`
- `riido-daemon/internal/provider/openclaw/detect.go:13`
  - `RIIDO_OPENCLAW_PATH`

Resolve semantics:

- `riido-daemon/internal/agentbridge/detectutil/detectutil.go:40`
  - env override is a pin, not a hint
- `riido-daemon/internal/agentbridge/detectutil/detectutil.go:67`
  - non-empty override is checked first
- `riido-daemon/internal/agentbridge/detectutil/detectutil.go:68`
  - invalid override returns no candidates
- `riido-daemon/internal/agentbridge/detectutil/detectutil.go:93`
  - process PATH lookup
- `riido-daemon/internal/agentbridge/detectutil/detectutil.go:102`
  - augmented search dirs
- `riido-daemon/internal/agentbridge/detectutil/detectutil.go:121`
  - production search dirs include login shell PATH and well-known install dirs

현재 머신에서는 Codex/Claude detect가 성공했다. 따라서 "daemon이 내 로컬에 깔린
CLI를 감지하지 못한다"는 현상이 UI에서 보인다면 다음 원인을 먼저 봐야 한다.

- UI가 local daemon status가 아니라 stale SaaS runtime snapshot을 보고 있다.
- 중복 daemon 중 다른 process/socket/pid state를 보고 있다.
- Desktop dev/prod userData가 다르고, 다른 daemon install root를 보고 있다.
- `RIIDO_*_PATH`가 Desktop/Finder/launchd env에 잘못 pin되어 있다.
- selected agent runtime binding이 detected runtime과 다르다.
- Cursor/OpenClaw는 실제로 local PATH에 없다.

The important diagnostic split is:

- `riido bridge detect` answers "can this binary resolve provider CLIs from this
  process environment?"
- `riido daemon status` answers "what does the daemon reachable through this
  socket think its runtime capabilities are?"
- SaaS device runtime APIs answer "what runtime snapshot has the control-plane
  last accepted for this device/workspace?"
- Agent settings answer "which runtime id is this agent bound to?"

Those four answers can diverge. A user-facing "not detected" label must identify
which surface is missing; otherwise a local detection success can be misreported
as an agent/runtime availability failure.

## 18. Local Daemon Runtime Status Projection

Desktop local runtime lookup은 installed daemon binary로 `riido daemon status`를
실행하고 그 `runtimes`를 그대로 projection한다.

Evidence:

- `riido-desktop/src/modules/daemonLocalRuntimes.ts:38`
  - `runDaemonStatusJSON`
- `riido-desktop/src/modules/daemonLocalRuntimes.ts:40`
  - `spawn(binaryPath, ['daemon', 'status'])`
- `riido-desktop/src/modules/daemonLocalRuntimes.ts:73`
  - `getLocalDaemonRuntimes`
- `riido-desktop/src/modules/daemonLocalRuntimes.ts:79`
  - maps `parsed.runtimes`
- `riido-desktop/src/modules/ipc.ts:92`
  - IPC `ai-agent:local-runtimes`

Client onboarding uses local runtime result when available.

Evidence:

- `riido-client/src/components/domain/aiAgentOnboarding/useLocalDesktopRuntimes.ts:50`
  - calls Electron local runtimes IPC
- `riido-client/src/components/domain/aiAgentOnboarding/useLocalDesktopRuntimes.ts:66`
  - `connected: runtime.available`
- `riido-client/src/components/domain/aiAgentOnboarding/AiAgentSelectStep.tsx:190`
  - local daemon runtimes preferred for desktop
- `riido-client/src/components/domain/aiAgentOnboarding/AiAgentSelectStep.tsx:198`
  - `hasLocalRuntimes`
- `riido-client/src/components/domain/aiAgentOnboarding/AiAgentSelectStep.tsx:199`
  - `effectiveRuntimes = hasLocalRuntimes ? localRuntimes : runtimeOptions`

이 surface에서는 `riido daemon status`가 어떤 socket/daemon에 붙는지가 곧 UI 결과다.

## 19. SaaS Runtime Snapshot Projection

Daemon reports runtime snapshots to control-plane.

Evidence:

- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:60`
  - `RuntimeSnapshotRecord`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:150`
  - runtime registration snapshot path
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:169`
  - posts full accumulated provider set
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:327`
  - `postRuntimeSnapshot`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:331`
  - `/v1/daemon/runtime-snapshot`

Control-plane merges incoming runtime records into device read model.

Evidence:

- `riido-control-plane/internal/riidoaiserver/server.go:134`
  - route `/v1/daemon/runtime-snapshot`
- `riido-control-plane/internal/riidoaiserver/server.go:264`
  - `handleDaemonRuntimeSnapshot`
- `riido-control-plane/internal/riidoaiserver/ai_agent_daemon_runtime.go:63`
  - `SyncAIAgentDaemonRuntimeSnapshot`
- `riido-control-plane/internal/riidoaiserver/ai_agent_daemon_runtime.go:139`
  - `upsertDeviceRuntimeSnapshotLocked`
- `riido-control-plane/internal/riidoaiserver/ai_agent_daemon_runtime.go:169`
  - `DeviceRuntimeSnapshotEvent`

따라서 local status, SaaS device runtime list, selected agent runtime binding은 서로
다른 read surface다. "CLI detected"와 "agent가 해당 runtime을 쓸 수 있음"은 같은
상태가 아니다.

A stale runtime snapshot is especially plausible when daemon singleton is
unstable. One daemon may detect and report correctly, another may be waiting on
lock, and the UI may be reading either local daemon status or the last SaaS
snapshot depending on screen and environment. This is why CLI detection should
be debugged only after process identity and socket ownership are stable.

## 20. Infinite Queued State

추가 확인일: 2026-06-09.

`queued`는 하나의 원인으로만 생기는 상태가 아니다. 현재 코드에서 queued가
무한히 지속될 수 있는 주요 경로는 다음 세 가지다.

1. Daemon이 해당 agent/runtime binding을 poll하지 않는 경우.
2. 새 assignment가 이전 assignment에 blocked 되어 있고, blocker가 terminal로
   바뀌지 않는 경우.
3. Client read model이 queued/active thread를 계속 active stream으로 선택하는 경우.

Assignment queue 생성 경로:

- `riido-control-plane/internal/riidoaiserver/store.go:798`
  - 새 assignment는 `AssignmentQueued`로 생성된다.
- `riido-control-plane/internal/riidoaiserver/store.go:809`
  - `EventAssignmentQueued`를 append한다.
- `riido-control-plane/internal/riidoaiserver/store.go:1011`
  - daemon poll이 queued assignment를 lease하려면
    `assignmentBlockerCleared(...)`가 true여야 한다.
- `riido-control-plane/internal/riidoaiserver/store.go:1547`
  - blocker가 있으면 blocker assignment가 terminal일 때만 clear로 간주한다.

즉 queue 자체는 정상 대기 상태다. 문제는 queue 탈출 조건이 polling, runtime
binding, previous assignment terminal transition에 강하게 의존한다는 점이다.

### 20.1 Daemon Poll Requires Exact Runtime Binding Match

Daemon SaaS plane은 dynamic binding 모드에서 `/v1/daemon/agent-bindings`를 읽고,
현재 runtime actor의 `RuntimeID`와 정확히 일치하는 binding만 poll한다.

Evidence:

- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:369`
  - `ClaimTask(ctx, runtimeID)`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:371`
  - dynamic binding mode
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:376`
  - iterate returned bindings
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:377`
  - `binding.RuntimeProvider != provider` or `binding.RuntimeID != runtimeID`
    skips the binding
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:380`
  - only then `pollAgent(...)`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:406`
  - no matching binding returns `nil`

Control-plane also validates daemon poll requests against the registered binding.

Evidence:

- `riido-control-plane/internal/riidoaiserver/agent_registry.go:110`
  - `validateDaemonBinding`
- `riido-control-plane/internal/riidoaiserver/agent_registry.go:127`
  - daemon id must match
- `riido-control-plane/internal/riidoaiserver/agent_registry.go:130`
  - device id must match if binding has one
- `riido-control-plane/internal/riidoaiserver/agent_registry.go:133`
  - runtime id must match

Therefore a queued assignment can sit forever if it is assigned to an agent whose
runtime binding no longer matches the live daemon runtime id.

Local evidence from the current machine:

- Live daemon id observed on 2026-06-09:
  `dev_39b7268cd02e004b16333b16e047fe6a`
- Live runtime ids:
  - `dev_39b7268cd02e004b16333b16e047fe6a:codex`
  - `dev_39b7268cd02e004b16333b16e047fe6a:claude`
  - `dev_39b7268cd02e004b16333b16e047fe6a:openclaw`
  - `dev_39b7268cd02e004b16333b16e047fe6a:cursor`
- `GET /v1/daemon/agent-bindings` with the live device credential returned only
  one Codex binding:
  `agent-elqlhnquvfopbpupxwqgw-dev-39b7268cd02e004b16333b16e047fe6a-codex`
  -> `dev_39b7268cd02e004b16333b16e047fe6a:codex`
- Local daemon status showed Codex and Claude executable detection succeeded,
  but only Codex had an agent binding returned by control-plane.
- Current process list still showed multiple desktop-launched
  `riido daemon start --foreground` processes using the same desktop
  `daemon.pid`, `daemon.lock`, and `daemon.log`.

Implication:

- A queued assignment for the current Codex agent should be discoverable by this
  daemon if it is not blocked.
- A queued assignment for Claude/OpenClaw/Cursor, an old `agentd-local:*`
  runtime id, or a previous daemon id will not be polled by the current live
  daemon.
- Local CLI detection does not prove the SaaS agent has a live binding. Codex
  and Claude can both be detected locally while only one SaaS agent is actually
  routable.

The control-plane source already acknowledges this class of failure.

Evidence:

- `riido-control-plane/internal/riidoaiserver/ai_agent_daemon_runtime.go:219`
  - device may be connected to many workspaces
- `riido-control-plane/internal/riidoaiserver/ai_agent_daemon_runtime.go:222`
  - otherwise an agent assigned from another connected workspace is never polled
- `riido-control-plane/internal/riidoaiserver/ai_agent_daemon_runtime.go:223`
  - explicit comment: its assignment stays queued forever

### 20.2 Blocked Queued Assignments Can Wait Forever

When a task is reassigned while an existing assignment is still active,
control-plane queues the new assignment and points it at the old assignment as a
blocker.

Evidence:

- `riido-control-plane/internal/riidoaiserver/store.go:762`
  - replacement/blocker bookkeeping begins
- `riido-control-plane/internal/riidoaiserver/store.go:778`
  - active previous assignment is moved to `AssignmentCancelling`
- `riido-control-plane/internal/riidoaiserver/store.go:782`
  - new assignment records `BlockedByAssignmentID = current.ID`
- `riido-control-plane/internal/riidoaiserver/store.go:1011`
  - blocked queued assignments are skipped by poll
- `riido-control-plane/internal/riidoaiserver/store.go:1547`
  - blocker clears only when blocker is terminal

This is correct sequencing if the old assignment eventually receives cancel,
fails stale, or completes. It becomes an infinite queue if the old assignment's
agent no longer polls.

The stale/cancel cleanup path is poll-driven by agent id:

- `riido-control-plane/internal/riidoaiserver/store.go:935`
  - poll reads `state.agentAssignments[agentID]`
- `riido-control-plane/internal/riidoaiserver/store.go:938`
  - `AssignmentCancelling` is handled only during that agent's poll
- `riido-control-plane/internal/riidoaiserver/store.go:943`
  - expired cancelling lease is failed stale only there
- `riido-control-plane/internal/riidoaiserver/store.go:976`
  - active assignments are also checked only in that agent's poll path
- `riido-control-plane/internal/riidoaiserver/store.go:983`
  - expired active lease is failed stale only there

Therefore:

1. Assignment A is leased/running for old agent/runtime.
2. User reassigns or comments in a way that creates assignment B.
3. A becomes `cancelling`, B becomes `queued` with `BlockedByAssignmentID=A`.
4. If old agent/runtime A is no longer returned by `agent-bindings`, or its
   daemon/runtime id changed, no daemon polls A.
5. A never reaches terminal.
6. B is skipped forever by `assignmentBlockerCleared`.
7. Client sees B as queued forever.

This directly connects the daemon singleton/runtime binding problem with
infinite queue. Duplicate daemon starts, daemon id changes, stale runtime
snapshots, deleted agents, and unavailable providers can all strand the blocker
that the next queued assignment depends on.

### 20.3 Claim Errors Are Silent In The Daemon

Daemon supervisor claim loop does not log claim errors.

Evidence:

- `riido-daemon/internal/agentbridge/supervisor/supervisor.go:444`
  - calls `Source.ClaimTask(ctx, status.RuntimeID)`
- `riido-daemon/internal/agentbridge/supervisor/supervisor.go:449`
  - `err != nil || req == nil || req.ID == ""`
- `riido-daemon/internal/agentbridge/supervisor/supervisor.go:450`
  - sleep/pacing branch
- no log line records the claim error, poll rejection, empty binding list, or
  binding mismatch reason

So if `/v1/daemon/agent-bindings` fails, returns no matching binding, or
`/v1/agents/{agent_id}/poll` rejects the daemon/runtime binding, the local user
usually sees only "queued". The daemon log mostly shows status requests and
startup lines, not the reason claim made no progress.

### 20.4 Client Treats Queued As Active

Client task thread selectors intentionally treat queued as working/active.

Evidence:

- `riido-client/src/lib/hooks/queries/aiAgent/taskThreadSelectors.ts:10`
  - working statuses include `queued`
- `riido-client/src/lib/hooks/queries/aiAgent/taskThreadSelectors.ts:18`
  - active assignment states include `queued`
- `riido-client/src/lib/hooks/queries/aiAgent/taskThreadSelectors.ts:37`
  - `isThreadWorking`
- `riido-client/src/lib/hooks/queries/aiAgent/taskThreadSelectors.ts:65`
  - `active_stream.thread_id` wins current-thread selection

Control-plane also marks queued as an active stream state.

Evidence:

- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:2997`
  - `taskThreadHasActiveStream`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:2999`
  - queued/running/stopping return true

Thus the client is not the root cause. It faithfully renders the server
projection. If the assignment store never moves queued -> leased/running or
queued -> terminal, UI remains queued indefinitely.

### 20.5 Queue Fix Direction

Recommended fixes for the queue class:

1. Add explicit queue diagnostics:
   - assignment id
   - agent id
   - runtime provider
   - expected daemon id
   - expected runtime id
   - current daemon id
   - current runtime id
   - `BlockedByAssignmentID`
   - blocker state
   - last poll time per agent
2. Make daemon claim errors visible:
   - log `/agent-bindings` failures
   - log no-matching-binding decisions at a throttled interval
   - log `validateDaemonBinding` rejection details without secrets
3. Add a global stale-blocker sweeper or make `ClaimNextAssignment` capable of
   terminalizing stale blockers even when the old agent no longer polls.
4. When a daemon/runtime snapshot moves an agent binding to a new runtime id,
   re-evaluate queued assignments for that agent and clear or fail assignments
   bound to stale runtime ids.
5. Ensure stop/cancel lifecycle always produces a durable terminal event for the
   assignment store, not just a client read-model update.
6. In client, keep queued rendering, but expose a stale queue reason when the
   server reports no poller, binding mismatch, or blocked-by stale assignment.

Priority order for this specific symptom:

1. Confirm the affected queued assignment id and agent id.
2. Check whether the affected agent appears in live `agent-bindings`.
3. If it appears, check whether `BlockedByAssignmentID` is set and whether the
   blocker is terminal.
4. If blocker is not terminal, inspect whether the blocker agent/runtime is still
   pollable by the current daemon.
5. Add server-side repair/fail-stale behavior so queue does not depend on the old
   agent polling forever.

### 20.6 Stepwise Implementation Plan

The implementation should be split so the first change fixes the infinite queue
class without also changing daemon singleton ownership, SSE batching, or
worktree binding. Those are connected issues, but mixing them into one patch
will make regressions hard to isolate.

Step 1: Define queue repair behavior in the SSOT docs.

- Update the owning control-plane assignment lifecycle document before code.
- State that a queued assignment blocked by a stale non-terminal assignment must
  not depend forever on the old agent polling.
- Define whether stale blockers become `failed` or `cancelled`.
- Define the event message and metadata that explain the repair.
- Define diagnostic fields exposed for queued assignments:
  `assignment_id`, `agent_id`, `runtime_provider`, `blocked_by_assignment_id`,
  blocker state, expected daemon id, expected runtime id, and last poll time if
  available.

Step 2: Add a focused control-plane regression test.

- Construct assignment A as leased/running or cancelling.
- Create assignment B that replaces A and is queued with
  `BlockedByAssignmentID=A`.
- Simulate the old agent/runtime no longer polling.
- Poll or claim from the new assignment path.
- Assert A is terminalized by repair and B becomes leaseable.
- Assert a durable event records why A was terminalized.

This test should fail before the fix. It is the guard for the infinite queue
bug.

Step 3: Implement stale blocker repair in the assignment store.

- Put the repair close to queue claim, not only in old-agent poll handling.
- When scanning queued candidates, inspect `BlockedByAssignmentID`.
- If the blocker is already terminal, proceed normally.
- If the blocker is non-terminal but stale or unclaimable, terminalize it and
  append an assignment event.
- Then clear or bypass the blocker so the queued assignment can lease.

The repair must be idempotent. Re-running the same poll/claim should not append
duplicate terminal events or produce conflicting assignment states.

Step 4: Add minimal diagnostics for queue no-progress.

- Add server-side diagnostic information sufficient to answer:
  "what is this queued assignment waiting for?"
- At minimum, expose blocked-by assignment id and blocker state through the
  read model or a diagnostic endpoint/event.
- Do not expose device secrets, bearer tokens, provider env, or prompt bodies.

Step 5: Make daemon claim failures observable.

- Add throttled logging for:
  - `/v1/daemon/agent-bindings` failure
  - no binding matching the current runtime id
  - poll request rejection
  - repeated no-progress while a matching binding exists
- Keep the logs concise and secret-free.
- Prefer counters/metrics if the existing metrics surface can carry them.

Step 6: Verify the live stuck queue case.

- Check the affected queued assignment id and agent id.
- Check live `agent-bindings` for that device.
- Check whether the queued assignment has `BlockedByAssignmentID`.
- If it has a blocker, verify the blocker is terminalized and the queued
  assignment leases after the fix.
- Confirm the client leaves permanent queued state without needing a page reload
  beyond the existing query/SSE flow.

Step 7: Continue with the connected lifecycle fixes in separate patches.

- Stop/cancel terminal fence across assignment store, daemon runtime, provider
  process, and read model.
- Daemon singleton process/socket/pid/lock ownership.
- SSE progress batching and raw text delta separation.
- Real repository/worktree binding for provider CLI runs.

### 20.7 Implementation Progress

2026-06-09 first queue repair patch:

- Updated the control-plane SSOT document
  `riido-control-plane/docs/20-domain/saas-control-plane.md` to define queue
  claim repair semantics.
- Added an in-memory assignment store regression test for a queued assignment
  blocked by a stale active/cancelling assignment.
- Added DynamoDB claim-path coverage for the same stale-blocker case.
- Implemented queue-claim repair in the control-plane assignment store:
  - missing blocker: clear `blocked_by_assignment_id` and continue claim;
  - queued blocker: cancel the historical queued blocker before claim;
  - stale active/cancelling blocker: fail the blocker and clear the queued
    candidate before claim;
  - terminal blocker: allow claim and clear the blocker on the claimed
    projection.
- Implemented DynamoDB repair as part of claim transaction payload when the
  durable claimer sees stale blockers. The transaction persists both the blocker
  repair operation and the current assignment poll-start operation.
- Extended claimed assignment application so durable claim repair operations are
  also reflected in the actor's in-memory state/event history.
- Added optional client thread `queue_diagnostics` so a queued thread can expose
  `blocked_by_assignment_id`, blocker state, blocker agent id, and blocker
  runtime provider when the assignment projection has that information.
- Added daemon-side throttled claim observability for source claim errors,
  dynamic binding mismatches, poll errors, and no-assignment results with a
  matching binding.

Verified:

- `go test ./internal/riidoaiserver -run 'TestStoreActorPollRepairsStaleBlockedQueuedAssignment|TestDynamoDBAssignmentOperationStoreClaimRepairsStaleBlockedAssignment|TestDynamoDBAssignmentOperationStoreSkipsBlockedAssignmentUntilBlockerTerminal' -count=1`
- `go test ./internal/riidoaiserver -count=1 -timeout=120s`
- `go test ./internal/agentbridge/supervisor ./internal/agentbridge/controlplane/saasplane -count=1`

Remaining from this queue section:

- Verify the live stuck queue case with a real affected assignment id.
- Continue the related stop/SSE/runtime/provider process lifecycle fixes as
  separate patches.

## 20.1. Terminal Thread Still Appears As Selected Participant

New symptom:

- After an agent run fails (`agent work failed`) or is stopped, removing the
  agent from the task participant dropdown appears to work optimistically.
- After refresh, the same agent is selected again.
- Selecting a different agent can also appear to revert after refresh.

Root cause candidate:

- The client participant dropdown derives selected AI agents from the current
  task thread. The current thread selector intentionally returns the latest
  terminal thread when there is no active stream, so the task panel can show the
  latest failure/completion history.
- The participant dropdown then treats any current thread that is not
  `unassigned` as selected. That is too broad: `failed`, `stopped`, and
  `completed` are history states, not active participant states.
- The control-plane unassign path has a matching persistence gap. It looks up
  the target thread via `activeTaskThreadForAgentLocked`, so a `failed` terminal
  thread is not found. The response can lose the original assignment/thread id
  and write a synthetic stopped thread instead of updating the failed thread the
  user is removing.

Required fix:

- Client: expose an explicit "active participant" predicate and use it for the
  participant dropdown. Terminal task threads may remain visible in the task
  history/panel, but they must not render as checked participants.
- Control-plane: explicit user unassign/stop should locate the latest thread for
  the agent, not only active-stream threads, so terminal failed/stopped runs can
  be associated with the original assignment/thread id and fenced from being
  re-selected.
- Tests: cover terminal `failed` thread unassign and client participant
  selection so a failed historical run does not become selected again after
  refresh.

## 21. Existing Test Gaps

필요한데 빠진 test:

- desktop launcher가 이미 singleton daemon이 있을 때 추가 foreground start process를
  남기지 않는지
- daemon start가 lock busy일 때 무기한 대기하지 않고 기존 daemon status 또는 명확한
  failure를 반환하는지
- default socket과 desktop userData lock/pid/log path가 같은 lifecycle identity로
  묶이는지
- stop 이후 late `riido_log`가 stopped/cancelled thread를 running으로 되살리지
  않는지
- assignment cancelling/cancelled 이후 daemon/provider event가 fencing token 없이
  read model을 mutate하지 않는지
- `EventTextDelta`가 raw `riido_log` line으로 전송되지 않고 parsed/batched
  `/thread-progress`로만 사용자 progress를 갱신하는지
- local CLI detected 상태와 SaaS runtime snapshot stale 상태가 UI에서 구분되는지
- 실제 repo path 없이 assignment가 들어올 때 provider CLI run을 시작하지 않거나
  명시적으로 blocked/error 처리하는지
- blocked queued assignment의 blocker agent가 더 이상 poll되지 않는 경우 queue가
  repair/fail되지 않는지
- daemon claim error, empty binding, runtime binding mismatch가 로그/metrics로
  노출되는지
- terminal `failed`/`stopped` task thread가 참가자 dropdown에서 selected agent로
  다시 해석되는지
- terminal thread unassign이 원래 assignment/thread id를 보존하며 read model을
  stopped/unassigned-equivalent 상태로 정리하는지

## 22. Primary Risks

| Severity | Risk | Why it matters |
| --- | --- | --- |
| Critical | duplicated/orphan daemon processes | multiple daemon instances can poll/report, socket can be hijacked, local lifecycle becomes unobservable |
| Critical | stop is not terminal for SSE/read model | user-requested stop can be reversed by late progress |
| Critical | blocked queued assignments can become permanent | stale blockers require an old agent poll before the next assignment can lease |
| High | raw text delta becomes progress line | huge event volume, bad Korean line rendering, request storms |
| High | task workdir lacks real repo binding | LLM CLI runs without the needed codebase context |
| High | CLI detection UI conflates local/SaaS/binding states | user sees "not detected" even when local daemon detects CLI |
| Medium | provider process kill has no graceful wait | output/cleanup truncation and race with final reporting |
| Medium | desktop stop is best-effort | stale pid/socket/orphan processes can survive |

## 22.1 Codex Rate-Limit Noise, Long Runs, And Context Switching

추가 확인일: 2026-06-09.

User-visible symptom:

```text
중지
codex rate limits updated
```

This text is unlikely to mean "the user request failed because of a Codex rate
limit". In `origin/main`, the Codex app-server notification
`account/rateLimits/updated` / `account_rate_limits_updated` is translated into
a user-visible `EventLog` with text `codex rate limits updated`.

Mainline evidence:

- `riido-daemon/internal/provider/codex/translate.go`
  - maps `account/rateLimits/updated` to `EventLog`
- `riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go`
  - maps `EventLog` to `EventProviderLog`
- `riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go`
  - provider log messages can become the visible thread message unless fenced by
    terminal assignment/thread state

Interpretation:

- `중지` means the assignment/thread has reached stopped/cancelled state.
- `codex rate limits updated` is an informational Codex app-server account
  window notification.
- They appear together when a provider log arrives near or after stop and leaks
  into the stopped thread's visible message.

Root fix:

1. Daemon should not report `account/rateLimits/updated` as user-visible
   provider log. It should be internal diagnostics only.
2. Control-plane should reject or ignore provider log/progress updates for
   terminal assignment/thread rows.
3. Client filtering can hide this exact phrase, but that is only a presentation
   fallback.

### 22.1.1 Why Simple Codex Work Can Run Too Long

Codex work is not a cheap command execution path. Each assignment starts a fresh
provider process and app-server handshake:

1. Start Codex process.
2. Send JSON-RPC `initialize`.
3. Send `initialized`.
4. Send `thread/start` or `thread/resume`.
5. Send `turn/start`.
6. Wait for `turn/completed` or terminal `thread/status/changed`.

Current branch defaults also make "hung" look slow:

- `RIIDO_DAEMON_RUN_HARD_TIMEOUT_SECONDS` defaults to 30 minutes.
- `RIIDO_DAEMON_RUN_SEMANTIC_IDLE_SECONDS` defaults to 10 minutes.

So if Codex does not emit a recognized terminal result, or emits only
non-semantic log/noise, the run can sit until semantic idle or hard timeout.
`EventLog` does not reset semantic idle, but the idle default is still long
enough that a simple request can appear stuck for minutes.

`origin/main` adds another pressure point:

- default `RIIDO_RUNTIME_MAX_CONCURRENT` is 4.
- each concurrent Codex run still starts its own app-server process/turn.
- multiple simultaneous app-server turns can increase account rate-limit pressure
  and make "rate limit updated" notifications appear more often.

### 22.1.2 Why Failures Are Opaque

The daemon session can fail for several distinct reasons:

- `semantic idle timeout`
- `hard timeout`
- `process exited without provider result`
- Codex JSON-RPC request error
- Codex runtime `error` notification
- provider executable detected but run command fails
- runtime ineligible because selected agent binding does not match live runtime
- empty/no-repo workdir for a coding task

But the product often collapses the visible message to `agent work failed` when
the terminal assignment event has no message. That fallback lives in the
control-plane AI Agent client read model.

One concrete daemon-side source of empty failure messages is provider-specific
error shape mismatch. Codex `turn_error` / `turn/failed` notifications may carry
the reason as `message`, `detail`, plain `error`, or nested `error.message`.
The translator must normalize those shapes into `Result.Error`; otherwise the
control-plane only receives a failed terminal state with no useful message.

There is also a control-plane projection source of message loss. Assignment
projection previously kept `Assignment.State` and `LastEventSeq`, but not the
last assignment event body. When a client thread read model was stale, thread
list/bootstrap reconcile could repair the thread from the projection by calling
`assignmentEventActionResponse(..., message="")`. That made a real provider
failure reason disappear behind the fallback `agent work failed`.

Therefore "agent work failed" is not enough diagnostic data. The terminal event
must preserve a classified failure reason and the UI should render that reason
instead of a generic fallback whenever one exists.

2026-06-09 follow-up:

- Control-plane `AssignmentProjection` now carries the last assignment event.
- DynamoDB assignment projection stores/loads `last_event_json`.
- Stale read-model repair uses the projection's last event message, so a daemon
  terminal reason such as `codex turn/start rpc error: ...`, `semantic idle
  timeout`, or `runtime ineligible` is not replaced by `agent work failed`.
- Daemon supervisor logs runtime failures with phase/runtime/provider/model/
  assignment/workdir/error fields so the next repro can be diagnosed from
  `daemon.log` even before the UI copy is refined.

### 22.1.3 Why Context May Not Switch

SaaS assignment conversion does not currently pass a provider
`ResumeSessionID`, so the observed context issue is probably not Codex
`thread/resume` accidentally reusing an old thread in the daemon path.

The more likely context mismatch is filesystem and assignment scope:

- control-plane composes a new prompt snapshot from task context;
- daemon converts the assignment to a task request;
- supervisor creates an isolated workdir under
  `<workdir_root>/<workspace>/tasks/<task>/runs/<assignment>/workdir`;
- no code in the reviewed daemon path clones, mounts, or checks out the selected
  repository into that workdir.

So the prompt context can change while the process filesystem context remains a
fresh generated folder with only runtime config. For coding work, that looks like
"context did not switch" or "the agent cannot find the right project".

Follow-up thread messages intentionally include the previous thread id, previous
run id, previous status, previous visible message, and the new user instruction.
That is correct for a follow-up inside the same task thread, but it must not be
confused with cross-task repo/workdir context switching.

## 23. Recommended Fix Order

1. Fix daemon singleton and desktop launch lifecycle first.
   - A second start must not create a detached lock-waiting process.
   - `socket`, `pid-file`, `lock-file`, and install root must represent the same
     lifecycle identity.
   - `daemon start --foreground` should not be used as a fire-and-forget ensure
     operation unless Desktop owns and cancels the child process.

2. Make stop/cancel terminal across control-plane and daemon.
   - Once assignment/thread is stopped/cancelled/terminal, late progress must not
     re-open `active_stream`.
   - Add fencing by assignment id/generation/lease token or terminal-state guard.
   - Treat assignment operation as the stop SSOT; client thread state should be a
     projection, not an independent pre-write that can conflict with assignment
     state.
   - Separate "cancel requested" from "provider process termination confirmed" in
     daemon/control-plane events.

3. Replace raw `EventTextDelta -> riido_log -> /events` progress path.
   - User-visible progress should come from parsed `<riido_log>...<end>` telemetry
     and `/thread-progress` batching.
   - Final assistant text should not be split into progress lines.
   - Terminal/final answer content should be stored as output/comment/result, not
     as a stream of progress rows.

4. Add queue repair and diagnostics.
   - Expose blocked-by state, poller state, runtime binding state, and last poll
     time.
   - Fail or repair stale blockers without requiring the old agent to poll.
   - Add throttled daemon claim error logging and metrics.

5. Bind tasks to real worktrees.
   - Assignment/task metadata must include actual repository/worktree identity.
   - If no repo binding exists, runtime should fail fast or ask for setup instead
     of launching provider CLI in an empty isolated directory.

6. Hide provider internal notifications from user-visible thread messages.
   - Codex `account/rateLimits/updated` is an internal app-server account window
     notification, not task progress or a terminal failure reason.
   - Provider lifecycle/log noise should remain diagnostic metadata unless it is
     explicitly promoted to user-facing progress.

7. Preserve classified failure reasons.
   - Terminal messages should distinguish timeout, process exit without result,
     provider RPC error, runtime ineligible, and missing workdir/repo context.
   - `agent work failed` should be a final fallback, not the normal visible
     result for daemon/provider failure.

8. Separate CLI detection states in UI and API.
   - Local daemon status
   - SaaS runtime snapshot freshness
   - Agent runtime binding
   - Provider executable/version probe result

9. Add lifecycle regression tests.
   - Especially late progress after stop, duplicate daemon start, and raw delta
     batching.
   - Include the `AssignmentCancelling -> AgentAssignmentStateStopping ->
     active_stream` path so stop behavior is explicit rather than incidental.

## 24. Open Questions Before Implementation

- Should Desktop own daemon shutdown on app quit, or should daemon persist across
  app restarts as a background agent?
- If daemon persists, what is the single source of truth for lifecycle identity:
  global socket path, userData install root, device id, or daemon id?
- Should `daemon start --foreground` fail fast on lock busy, or should desktop use
  a separate "ensure running" command that returns current status?
- Should control-plane reject `riido_log` after assignment terminal state at the
  assignment store layer, AI Agent read-model layer, or both?
- Should provider final assistant text be stored as final comment/output instead
  of progress lines?
- What contract carries local repository/worktree path from product task/comment
  context into daemon assignment?
- How should stale SaaS runtime snapshot be presented when local daemon status is
  currently detected?
- What should the product do when a queued assignment is blocked by an old
  assignment whose agent/runtime binding is no longer live?
- Should queued assignment timeout be explicit, or should stale blockers be
  repaired as soon as no poller is available?

## 25. Implementation Gate

`riido-daemon/AGENTS.md` requires SSOT-document-first changes. This review is a
documentation-only record and does not change behavior. Any behavior change that
follows from this review must update the owning domain/architecture document in
the same PR and must use the Riido task creation response `branchName` as the Git
branch name.
