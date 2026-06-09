# AI Agent Runtime Lifecycle — Consolidated Review (통합본)

> Review date: 2026-06-08
>
> Scope: `riido-daemon`, `riido-desktop`, `riido-client`, `riido-control-plane`,
> `riido-contracts` (and the broader local Riido workspace).
>
> 이 문서는 두 개의 독립 리뷰를 **중복 제거 + 교차검증**하여 하나로 합친 통합
> 기록이다. 기존 리뷰 파일은 수정하지 않았다.

## 0. 이 문서에 대해 (출처·방법론·표기 규칙)

두 출처를 합쳤다. 진단이 겹치는 항목은 하나로 병합했고, 각 항목에 출처와 검증
판정을 표기한다.

| 출처 | 파일 | 방법론 | 강점 |
| --- | --- | --- | --- |
| **[A]** | `docs/review/ai-agent-runtime-lifecycle-review-2026-06-08.md` | **런타임 실측 + 코드 읽기** (`riido daemon status` 실행, 라이브 프로세스 관찰) | lifecycle 정합성/소유권, repo 바인딩, stop ACID, 런타임 실증 |
| **[B]** | `RIIDO_DAEMON_ISSUES.md` (워크스페이스 루트) | **정적 코드 읽기 + 적대적 검증** (36 에이전트, 29 root cause 반증 검사) | 실패·지연 내부 메커니즘 (timeout/polling/PATH/policy/retry/세션 연속성) |

**표기 규칙**

- `[A]` / `[B]` / `[A+B]` — 어느 리뷰가 제기했는지.
- **verdict** — `confirmed` / `partial` / `refuted` / `latent`. [B]의 항목은 적대적
  검증자의 판정이고, 아래 ✅로 표시된 3건은 본 통합 과정에서 코드로 **추가 독립
  검증**된 항목이다.
- 모든 근거는 `repo/path:line` 형태로 인용한다. 실제 동작 변경은 하지 않았다.

**본 통합에서 새로 코드 검증된 3건 (모두 CONFIRMED)**

1. `[A§6]` 소켓 하이재킹 / socket-lock-pid 비원자성 → **§4.1 D4**
2. `[A§13]` task에 실제 repo/worktree 바인딩 부재 → **§4.4 F3**
3. `[A§15/16]` 늦은 progress가 stopped thread를 부활 (read model terminal fence 부재) → **§4.5 C2**

## 1. Executive Summary

현재 문제는 개별 버그 하나가 아니라 **lifecycle ownership이 분리되어 있고, 어떤
계층도 다른 계층의 상태를 fencing하지 못하는 구조적 문제**다.
`desktop daemon launcher`, `daemon singleton`, `runtime actor`,
`provider CLI process`, `control-plane assignment`, `client SSE active_stream`이
하나의 원자적 lifecycle로 묶여 있지 않다. 한 계층의 성공 응답이 다른 계층이 같은
상태에 도달했음을 보장하지 않는다.

그 결과 같은 계열에서 다음이 발생한다.

- 데스크탑 재시작 시 정리되지 않는 데몬이 남는다 (고아). 서로 다른 lock 경로의
  데몬이 같은 소켓을 탈취할 수 있다.
- 댓글마다 provider CLI가 새 프로세스로 cold start되고, run timeout이 없어 hung
  CLI가 무한 실행된다.
- **댓글로 트리거된 코딩 작업이 실제 소스 repo가 아니라 빈 isolated 디렉터리에서
  실행된다.**
- provider raw text delta가 batching 없이 progress로 전송되어 SSE에 한두 글자씩
  줄바꿈이 쌓인다.
- Stop을 눌러도 늦은 progress가 stopped thread를 다시 running으로 되살린다.
- 새 task는 최대 5초(idle poll) 지연 후에야 claim되고, 다양한 실패가 불투명한
  `EventAssignmentFailed`로 뭉쳐진다.

## 2. 런타임 실측 증거 (출처 [A])

실제 머신에서 확인된 provider CLI:

| Provider | Local executable | Detection |
| --- | --- | --- |
| Codex | `/Users/work/.local/bin/codex` | detected, `codex-cli 0.137.0` |
| Claude | `/Users/work/.local/bin/claude` | detected, `2.1.168 (Claude Code)` |
| Cursor | `cursor-agent` not found | missing |
| OpenClaw | `openclaw` not found | missing |

- 설치된 Desktop daemon binary:
  `/Users/work/riido-ai-agent-development/electron-user-data/ai-agent-daemon/bin/riido`,
  status로 본 버전 `v0.0.14`.
- `riido daemon status`는 Codex/Claude를 detected로 보고.
- **동시에 같은 desktop dev parent 아래 daemon start 프로세스가 2개 확인됨.**
- 두 프로세스는 같은 `daemon.pid` / `daemon.lock` / `daemon.log`를 사용.
- pid file은 한 프로세스만 가리키며, 다른 하나는 singleton lock 대기 상태일
  가능성이 높음.

> **해석 (통합):** 같은 lock 경로를 공유하는 이 "2개"는 *serving 1개 + flock 대기
> 좀비 1개*(§4.1 **D3**)다. 동시에 listen하는 진짜 중복이 아니다. 진짜 동시 실행은
> *서로 다른 lock 경로*(데스크탑 userData lock vs 수동 `~/.riido/.lock`)일 때만
> 발생하며, 그 경우 소켓 하이재킹(§4.1 **D4**)이 일어난다.

## 3. 시스템 경계 (왜 한 트랜잭션이 아닌가)

사용자에게 보이는 한 번의 AI Agent run은 독립적인 여러 시스템을 가로지른다.

- Desktop daemon launcher (`riido-desktop`)
- Daemon singleton / socket / status (`riido-daemon`)
- Runtime actor / session / process exec / SaaS reporter (`riido-daemon`)
- Assignment store / AI Agent read model / runtime snapshot / SSE server
  (`riido-control-plane`)
- Client stop mutation / thread rendering / SSE hook (`riido-client`)

`socket`, `pid file`, `lock file`, `install root`, `device/runtime identity`,
`assignment id`, `client thread projection`은 **하나의 lifecycle identity로
취급되어야 하는데** 그렇지 않다. 그래서 아래 문제들이 단순 UI 이슈가 아니다.

---

## 4. 관심사별 통합 근본 원인

### 4.1 데몬 중복 / 고아 프로세스

#### D1 — 앱 종료 시 데몬을 전혀 stop하지 않음  `[A§4 + B:1-A]` · confirmed · **high**

- **메커니즘:** 데몬 관련 종료 배선은 `before-quit → daemonAutoLaunchController.stop`
  하나뿐. 이 `stop()`은 `stopped=true; clearInterval(interval)`로 30초 폴링 루프만
  멈추고 `daemon stop`/SIGTERM/kill을 전혀 보내지 않는다. `riido daemon stop`을
  발행하는 유일한 함수 `stopDaemonIfRunning`은 startup의 버전드리프트/forceRestart
  경로에서만 호출되고 종료 경로에서는 호출되지 않는다.
- **근거:** `riido-desktop/src/main.ts:606`; `riido-desktop/src/modules/daemonLauncher.ts:964-968`;
  `main.ts:719-726,729-733`; `daemonLauncher.ts:563-591`(stop helper), `:811`(유일 호출처).
- **수정:** `will-quit`/`before-quit` 전용 teardown에서 동기 `daemon stop` 발행
  (`execFileSync`, 짧은 timeout). **난이도: medium.**

#### D2 — detached + unref로 spawn → 의도적 고아화  `[A§4 + B:1-B]` · confirmed · **high**

- **메커니즘:** `buildDaemonLaunchOptions` 기본 `detached:true`, spawn 후
  `child.unref()`. spawn args가 `daemon start --foreground`라 자식이 곧 장수 데몬
  본체. POSIX setsid + unref로 부모(Electron) 종료 시 신호를 받지 않아 생존.
  foreground 데몬은 SIGTERM/SIGINT/소켓 shutdown에서만 종료.
- **근거:** `daemonLauncher.ts:598-624`(detached 기본), `:632-634`(spawn+unref),
  `:711-723`(--foreground args); `riido-daemon/cmd/riido/daemon.go:401`(종료 조건).
- **수정:** spawn된 child(또는 pid)를 추적하고 종료 시 cooperative `daemon stop` →
  SIGTERM → (유예) SIGKILL. 또는 detached/unref 설계 재검토. **난이도: medium.**

#### D3 — 단일 lock이 fast-fail 대신 무한 대기  `[A§5 + B:1-D]` · confirmed · **high**

- **메커니즘:** `runDaemonStartForeground`이 `c9lock.AcquireFile(ctx, lockFile)`로
  flock(`LOCK_EX|LOCK_NB`)을 시도하고, busy면 10ms 폴링으로 해제 또는 ctx.Done까지
  대기. 프로덕션은 `context.Background()`라 ctx가 취소되지 않아, 점유된 lock에 대해
  두 번째 데몬이 **감지-후-종료하지 않고 무한 블록**(lock-file fd를 점유한 좀비).
  'already running' 프로브도 fast-fail도 없음.
- **근거:** `daemon.go:204`(blocking acquire), `:88-90`(context.Background()),
  `internal/lock/filelock.go:24`(폴링), `internal/lock/filelock_unix.go:11`(LOCK_EX|LOCK_NB).
- **수정:** non-blocking `TryAcquireFile`(busy → typed `ErrLockHeld`) 또는 bounded
  ctx(1~2s). lock busy 시 '이미 실행 중' 로깅 후 식별 가능한 코드로 즉시 종료.
  **난이도: small.**

#### D4 — socket/lock/pid identity가 원자적이지 않음 → 소켓 하이재킹  `[A§6]` · ✅ **CONFIRMED (본 통합 검증)** · **critical**

- **메커니즘:** 데스크탑은 `--socket` 없이 데몬을 띄움 → 데몬은 **고정 default
  소켓**(`~/Library/Application Support/riido/agentd.sock`) 사용. serve 시작 시
  소켓 파일을 **무조건 `os.Remove`**. 단일 lock은 lock-file **경로** 기준 flock이라
  서로 다른 lock 경로의 두 데몬은 배제되지 않음. 따라서 데몬 A(lock 경로 A, 소켓 S)
  실행 중 데몬 B(다른 lock 경로 B, 같은 소켓 S)가 뜨면, B가 lock에 막히지 않고
  `os.Remove(S)` + re-listen으로 A의 소켓을 unlink → **A는 살아있지만 도달 불가한
  고아**가 되고 이후 status/stop은 B로 가거나 실패.
- **정밀한 트리거 (검증으로 확정):** 데스크탑 2회 실행은 **안전**(같은 userData
  lock 경로 → 두 번째는 flock 대기 = D3). 진짜 하이재킹은 **데스크탑(userData lock)
  + 수동 `riido daemon start`(default `~/.riido/.lock`)** 조합 — 같은 소켓, 다른
  lock. 데스크탑의 `daemon status`/`daemon stop`도 `--socket`을 생략하므로 하이재킹
  후엔 B에 도달(A는 격리됨).
- **근거:** `daemonLauncher.ts:711-721`(no --socket), `:351-364`(userData lock
  `.../ai-agent-daemon/daemon.lock`); `daemon.go:175-181`(default socket fallback),
  `:843-867`(`defaultAgentDaemonSocket`), `:326-328`(`os.Remove(flags.socket)`),
  `:869-874`(default lock `~/.riido/.lock`); `internal/lock/filelock_unix.go:10-12`(flock).
- **수정:** socket·pid·lock·install root를 하나의 lifecycle identity로. per-instance
  소켓을 쓰거나 데스크탑이 항상 `--socket`을 명시. serve 전 `os.Remove`를 소켓
  liveness/ownership 검사로 대체(살아있는 소켓 unlink 금지). **난이도: medium.**

#### D5 — daemon stop이 best-effort (소켓 교체/stale pid 시 고아 누락)  `[A§7 + B:1-C]` · partial · **medium**

- **메커니즘:** `stopDaemonIfRunning`은 reachable socket으로 status를 읽어 거기서
  socket/pid를 뽑아 stop. 소켓이 다른 데몬에 의해 교체되었거나 pid가 stale하면
  stop 트랜잭션 밖의 고아를 포괄하지 못함.
- **보정(B):** "다음 실행마다 중복 누적"은 기각 — stop 실패 시 새 데몬을 스폰하지
  않고(`daemonLauncher.ts:810-824`) 소켓 응답 데몬이 있으면 단락(`:826-835`). 잔존
  이슈는 *단일 고아/스테일* 데몬.
- **근거:** `daemonLauncher.ts:560,563-591`; (보정) `:810-835`.
- **수정:** spawn 전 probe-and-adopt, stop 실패를 '스폰 금지' 하드 조건화 + D1/D4
  수정. **난이도: medium.**

#### D6 — 비정상 종료 시 CLI 자식 고아 (자체 pgid), startup reaper 없음  `[B:1-E, A§14 evidence]` · confirmed · **medium**

- **메커니즘:** provider CLI 자식은 `Setpgid:true`로 데몬과 **별개 프로세스 그룹**.
  자식 kill 경로는 모두 in-process Go 코드(session ctx 취소 / 명시적 Cancel→Kill).
  데몬이 SIGKILL/크래시되면 실행되지 않고, 자식은 자체 pgid라 신호 미수신 →
  init/launchd로 reparent되어 잔존. `Pdeathsig` 미설정. graceful 경로는 올바름.
- **근거:** `internal/process/processexec/processexec_unix.go:11`(Setpgid),
  `:14-25`; `processexec.go:154-172`; `session.go:135`;
  `cmd/riido/daemon.go:1000-1013`(SIGTERM→SIGKILL escalate).
- **수정:** 데몬 startup 시 워크디렉터리 run 메타데이터의 고아 provider 프로세스
  스캔·종료; 데스크탑은 cooperative stop 선호(timeout 확대). **난이도: medium.**

#### D7 — Windows `.claim` lock이 비정상 종료 후 영구 스테일  `[B:1-F]` · latent · **low**

- **메커니즘:** Windows 단일 lock은 `path+'.claim'`을 `O_CREATE|O_EXCL`로 생성,
  `cleanupLockFile`이 `Release()`에서만 호출됨. SIGKILL/크래시 시 Release 미실행 →
  스테일 `.claim` 잔존 → 이후 모든 start가 무한 폴링. PID/mtime 검사 없음. 현재는
  데몬 아티팩트가 non-darwin에서 null이라 **latent**.
- **근거:** `internal/lock/filelock_windows.go:10-16,26-32`; `filelock.go:56-75`.
- **수정:** `.claim`에 owner PID/mtime staleness 자가치유. **난이도: small.**

### 4.2 LLM CLI 런타임 / 세션 관리

#### R1 — 매 assignment마다 새 단발(single-shot) CLI 프로세스, 재사용 없음  `[A§8/9 + B:2-B]` · confirmed

- **메커니즘:** runtime actor(`MaxConcurrent:1`, provider별 동시성 제한 — 재사용
  아님)가 submit마다 `BuildStart` → `session.Start` → `Process.Start`로 새 자식을
  fork. claude는 stream-json user 프레임 1개 후 CloseStdin, codex는 정확히 1턴 후
  종료. warm 프로세스/지속 REPL 없음. 따라서 댓글 1개 = task 1개 = 새 CLI 1개.
- **근거:** `cmd/riido/daemon.go:454,463`(MaxConcurrent:1);
  `runtimeactor.go:375,389,418`; `session.go:135`; `processexec.go:41`;
  `claude/protocol_driver.go:45`; `codex/protocol_driver.go:306`.
- **수정:** (제품 의도가 단발이면 불필요) 연속성이 필요하면 R2로.

#### R2 — 댓글 간 provider 세션/스레드 연속성 부재 (ResumeSessionID 미설정)  `[B:2-A]` · partial

- **메커니즘:** `taskRequestFromAssignment`가 `ResumeSessionID`를 절대 설정 안 함 →
  claude/cursor `--resume` 생략, codex는 항상 `thread/start`(새 스레드). provider
  세션 id는 `EventSessionIdentified`로 IR 로그에만 기록되고 후속 run에 재주입 안 됨.
  assignment 계약에 resume/thread/session 필드 자체가 없음.
- **보정(B):** control-plane이 후속 메시지 경로에서 **직전 스레드 메시지 + 상태**를
  새 프롬프트에 append(`### Previous Thread Message`)하므로 *제한적 프롬프트
  연속성*은 존재. 빠진 것은 provider 세션 resume과 전체 멀티턴 트랜스크립트.
- **근거:** `saasplane.go:762`; `runtimeactor.go:383`; `codex/protocol_driver.go:274`;
  `claude/command.go:139`; `cursor/command.go:136`; `openclaw/command.go:96`;
  `riido-contracts/assignment/types.go:132`; (보정)
  `riido-control-plane/internal/riidoaiserver/server.go:906-929`.
- **수정:** 계약에 prior-session 필드 추가 → terminal 완료 시 provider 세션 id 보고/
  영속화 → 다음 assignment에서 echo → `ResumeSessionID` 채움(배선은 이미 소비함).
  단, resume eligibility 는 같은 Riido task/thread, 같은 agent, 같은 provider/model,
  같은 runtime identity, 같은 workdir/repo context 에서만 true 다. 다른 task 는
  `/clear` 를 보내는 것이 아니라 `ResumeSessionID` 를 비워 fresh provider session 으로
  시작한다. Codex `app-server` 처럼 process 안에서 `thread/start` / `thread/resume` /
  `turn/start` 를 구조적으로 구분할 수 있는 provider 는 task/run 별 process spawn 을
  제거하고 runtime-scoped long-lived process 로 전환해야 한다. Claude 같은 one-shot
  adapter 는 structured multi-session primitive 확인 전까지 process reuse 대상이 아니다.
  **난이도: large.**

#### R3 — 프로덕션에 hard/idle run timeout 모두 없음 → hung CLI 무한 실행  `[B:2-C/4-B]` · confirmed · **high**

- **메커니즘:** `newDaemonRuntimeActor`가 HardTimeout 미설정,
  `taskRequestFromAssignment`가 Timeout/SemanticIdle 미설정 → 둘 다 0 → session이
  hard/idle 타이머를 arm하지 않음(>0일 때만). hung CLI(승인 대기, 멈춘 툴)는 절대
  timeout되지 않고 단일 슬롯을 영구 점유. 유일한 backstop은 control-plane active
  lease(기본 20초)이며, 만료 시 `failStaleAssignment`가 provider 진단 없는 불투명한
  **"active assignment lease expired"** 실패를 방출.
- **근거:** `daemon.go:454`; `runtimeactor.go:394`; `session.go:174,182`;
  `saasplane.go:778`; `riido-control-plane/.../assignment_operation_port.go:14`(20초),
  `store.go:945,972`.
- **수정:** 데몬측 HardTimeout(10~15분) + SemanticIdle(2~3분) 설정 → 불투명 lease
  만료를 명시적·분류된 timeout으로 전환, hung CLI 결정적 kill. **난이도: small.**

#### R4 — 비정상 종료 후 active assignment가 PollActive로 재폴링 → 처음부터 재스폰  `[B:2-E, A§10]` · confirmed

- **메커니즘:** control-plane은 active-lease assignment 보유 시 PollActive(동일
  assignment)를 반환. `saasplane.ClaimTask`는 PollActive를 PollStart와 동일 처리.
  동일 프로세스 내에선 inFlight dedup으로 무해하나, **재시작/크래시 후** inFlight가
  비면 새 CLI 프로세스로 프롬프트를 **처음부터 재실행**(새 스레드, resume 없음) →
  부작용(커밋/툴) 중복 가능. graceful stop은 terminal 처리되어 자가 치유.
- **근거:** `riido-control-plane/.../store.go:842,861,941,972`; `transition.go:3`;
  `saasplane.go:374`; `supervisor.go:410`.
- **수정:** startup에서 PollActive를 reconcile(resumable id 있을 때만 resume, 아니면
  조용한 fresh start 금지). R2와 결합 시 재시작에는 provider-native thread/session
  resume 을 선호하지만, tool call / 파일 수정 / commit 같은 이미 발생한 부작용의
  idempotency 는 보장하지 않는다. session id 가 없으면 `recovery_fresh_start` 같은 명시
  이벤트를 남기거나 recovery failure 로 처리해야 한다. **난이도: medium.**

#### R5 — 취소 watcher 고루틴/채널이 완료 후 데몬 종료까지 잔존  `[B:2-F]` · confirmed · **low**

- **메커니즘:** task마다 `forwardCancellation` 고루틴이 `cancelWatchers[taskID]`
  채널에서 블록. 정상 완료 시 `CompleteTask`가 map 엔트리는 삭제하나 **채널을 닫지
  않아** 고루틴이 파킹된 채 잔존(누적은 한정적).
- **근거:** `supervisor.go:464,998`; `saasplane.go:437,494`.
- **수정:** `CompleteTask`에서 map 삭제 전 채널 close. **난이도: small.**

### 4.3 SSE 버퍼링 / 줄바꿈

> **핵심:** "줄바꿈"은 삽입된 `\n`이 아니다. (1) 상류가 토큰 delta마다 이벤트 1개를
> 무버퍼 포워딩 + (2) 클라이언트가 각 라인을 별도 block `<p>`로 렌더하기 때문이다.
> "한두 글자/줄"은 codex `item/agentMessage/delta`·cursor `text` 같은 fine-grained
> 토큰 provider에 특정되고, claude·codex `agent_message`(큰 블록)에선 미발생.

#### S1 — 데몬이 모든 텍스트 delta를 coalescing 없이 EventRiidoLog로 `/events` 포워딩  `[A§11 + B:3-A]` · confirmed · **high**

- **메커니즘:** provider 번역기가 텍스트 fragment마다 `EventTextDelta` 1개 방출 →
  session → supervisor가 1:1로 `ReportEvent` → `eventRequestFromAgentEvent`가
  `EventRiidoLog`로 매핑 → delta당 HTTP POST 1회. 누산 버퍼/시간·크기 flush/배치
  전무. 커밋 #97/#98(8dd510e)이 이 포워딩을 'content block당 1 delta'로 도입했으나
  codex/cursor에선 delta=토큰.
- **근거:** `codex/translate.go:74,77`; `cursor/translate.go:33`; `session.go:215`;
  `supervisor.go:342,980`; `saasplane.go:461,828,836`.
- **수정:** task별 텍스트-delta 누산기(50~150ms debounce 또는 크기 임계) + 비텍스트
  이벤트/완료 시 flush → `/thread-progress` 배치로 라우팅(S4). **난이도: medium.**

#### S2 — control-plane이 delta마다 progress line + per-event SSE frame을 즉시 flush  `[A§11 + B:3-B]` · confirmed · **high**

- **메커니즘:** `handleAgentEvent`가 POST당 `RecordAIAgentAssignmentEvent` 1회 →
  EventRiidoLog당 단일 `AgentThreadProgressLine` + 1줄짜리 event → SSE writer가
  이벤트마다 `data:` 프레임 1개 즉시 flush. **서버측 coalescing 없음.** 단, 메시지
  텍스트에 `\n`을 삽입하진 않음(SSE 프레임 종결자만).
- **근거:** `server.go:1251,1271,1274,1610,1623,1628,2120`;
  `ai_agent_client_development.go:1466,1485`.
- **수정:** 연속 EventRiidoLog delta를 active run의 마지막 라인에 머지하거나 fan-out
  throttle. 논리적 업데이트당 1 프레임. **난이도: medium.**

#### S3 — 클라이언트가 각 progress 라인을 block `<p>`로 렌더 → 보이는 줄바꿈  `[A§11 + B:3-C]` · confirmed · **high**

- **메커니즘:** `thread.progress_messages` + seq 정렬 `thread.lines[]` + 스트림
  이벤트를 평탄 배열로 만든 뒤 `messages.slice(-6).map(m => <p>{m}</p>)`로 렌더. 각
  원소가 단일 delta fragment이고 `<p>`가 block-level이라 fragment마다 자기 줄.
  delta가 클라이언트에서 concat되지 않음.
- **근거:** `riido-client/.../AgentThreadCard.tsx:79,80,84,101,117,227`.
- **수정:** 스트리밍 어시스턴트 텍스트 delta를 seq 순으로 단일 문자열 concat →
  하나의 `<p className="whitespace-pre-wrap break-words">`에 렌더(모델 자체 개행만
  보존). discrete 상태/progress는 라인당 `<p>` 유지. **난이도: small/medium.**
- **권고:** 상류 granularity와 무관하게 견고해지므로 **즉효성 1차 수정**.

#### S4 — SSOT 스펙 위반: parsed/bounded `/thread-progress` 배치 설계를 우회  `[A§11]` · confirmed (스펙 근거) · **high**

- **메커니즘:** SSOT 문서는 client thread progress를 parsed `<riido_log>...<end>`
  **bounded batch**로 `POST /v1/agents/{id}/thread-progress`에 올리는 모델을 규정.
  실제 구현은 raw `EventTextDelta`를 `riido_log`로 바꿔 `/events`에 보냄 → 스펙이
  의도한 final-answer(run output) vs progress telemetry 분리를 붕괴시킴.
- **근거(스펙):** `riido-daemon/docs/migration/daemon.md:706`;
  `riido-control-plane/docs/20-domain/saas-control-plane.md:92`;
  `riido-control-plane/docs/20-domain/ai-agent-client-api.md:460`.
- **수정:** S1의 누산 결과를 `/thread-progress` 배치 엔드포인트로 라우팅하여 스펙과
  정합. final assistant text는 progress line이 아니라 output/comment/result로 저장.

#### S5 — request storm / client 503 (이미 코드에 인지됨)  `[A§12 + B:3-A/3-D]` · confirmed

- **메커니즘:** 데몬이 delta마다 동기 HTTP POST → control-plane fan-out → 클라이언트
  refetch storm. 클라이언트가 이미 throttle(`STREAM_INVALIDATE_MIN_INTERVAL_MS=800`)
  주석으로 "refetch storm caused 503s"를 인정. throttle은 증상 완화일 뿐.
- **근거:** `riido-client/.../useAiAgentTask.ts:231-240`.
- **수정:** 근본은 S1/S4(데몬·control-plane 배치). 클라이언트 throttle 유지.

### 4.4 느림 / 잦은 실패

> **요약:** 느림의 지배 원인은 **5초 idle 폴링(F1)**. 실패는 여러 독립 원인이 모두
> 불투명한 `EventAssignmentFailed`로 collapse된다 — 그 중 **task의 repo 바인딩
> 부재(F3)** 가 코딩 작업 실패의 가장 유력한 단일 원인이다.

#### F1 — idle 폴 간격 5초 → 새 댓글 task가 claim 전 최대 1 간격 대기  `[B:4-A]` · confirmed · **high**

- **메커니즘:** supervisor run loop는 폴링 후 idle(claim 0, in-flight 0)이면
  `IdlePollEvery`(주입값 5초)로 타이머 리셋. 데스크탑이 override 미설정 → 5초 기본.
  댓글 task는 다음 idle 폴이 ClaimTask를 발화할 때까지 0~5초 대기. dynamic-bindings
  모드는 폴마다 `GET /v1/daemon/agent-bindings`도 직렬 추가. push wake 없음.
- **근거:** `supervisor.go:314,381`; `cmd/riido/daemon_config.go:180`;
  `saasplane.go:358`(per-poll bindings GET);
  `riido-desktop/.../daemonDeviceCredential.ts:196`(override 미설정).
- **수정:** IdlePollEvery 1초로 하향 / heartbeat 'work-waiting' 힌트로 즉시 wake /
  per-poll bindings GET을 짧은 TTL 캐시로. **난이도: medium.**

#### F2 — 데몬 run timeout 부재 → stalled CLI가 20초 lease 만료의 불투명 실패로만 종료

R3와 동일 근본 원인. 실패가 "active assignment lease expired"로 표면화되어 진짜
원인(어느 provider가 어디서 멈췄는지)을 가린다. **수정: R3.** · confirmed · **high**

#### F3 — task에 실제 repo/worktree 바인딩 없음 → 빈 isolated 디렉터리에서 CLI 실행  `[A§13]` · ✅ **CONFIRMED (본 통합 검증)** · **critical**

- **메커니즘:** `taskRequestFromAssignment`가 `workspace_id`를
  `firstNonEmpty(ComponentID, TaskID)` — **경로가 아니라 ID**로 설정. workdir
  `Prepare`는 `<root>/<workspace>/tasks/<task>/runs/<run>/`에 **빈 하위 디렉터리만**
  생성(소스 코드 미포함; 주입되는 건 `AGENTS.md/CLAUDE.md/.riido` hooks뿐). cwd
  체인이 이 빈 workdir로 귀결 → **CLI가 대상 코드베이스 없이 실행**. clone/symlink/
  worktree/checkout 로직 없음. repo-aware한 `project.LocalPath`는 `mwsd` 진단 전용
  으로 실행 경로와 단절.
- **근거:** `saasplane.go:762,770`; `internal/workdir/workdir.go:234-276,249,755`;
  cwd 체인 `supervisor.go:593,613`, `runtimeactor.go:378,414-415,688`,
  `claude/command.go:155`, `codex/command.go:102`;
  `riido-contracts/assignment/types.go:121-148`(repo 필드 부재);
  `internal/project/projection.go:18-69`(LocalPath, mwsd 전용).
- **수정:** assignment/task 메타데이터에 실제 repository/worktree identity 포함. repo
  바인딩이 없으면 빈 디렉터리에서 CLI를 띄우지 말고 fast-fail 또는 setup 요청.
  **난이도: large (계약 변경 동반).**

#### F4 — spawn된 CLI가 데몬의 최소 PATH 상속 → git/node/rg 'command not found'  `[B:4-D]` · confirmed

- **메커니즘:** `taskRequestFromAssignment`가 `req.Env`를 채우지 않아 os/exec가 데몬
  `os.Environ()`(macOS launchd 최소 PATH `/usr/bin:/bin:...`)를 상속. provider
  바이너리는 절대경로 resolve로 실행되나, codex/claude가 git/node/ripgrep으로 shell
  out하면 자식이 최소 PATH라 실패 → flaky한 툴 실패로 보임. detectutil의 증강 PATH는
  바이너리 위치 찾기에만 쓰이고 자식 env에 전파 안 됨.
- **근거:** `saasplane.go:778`; `processexec.go:76`; `detectutil.go:133`.
- **수정:** detectutil 증강 PATH를 `StartRequest.Env`로 자식 CLI에 전파(F5와 통합).

#### F5 — provider 감지가 최소 launchd PATH에서 1회 실행 후 영구 캐시  `[B:4-C, A§17 부분]` · confirmed

- **메커니즘:** `runtimeactor.Start`가 `Detect`를 동기 1회 실행 후 caps를 액터 생애
  동안 캐시. 데스크탑이 `RIIDO_*_PATH`/증강 PATH 없이 띄우고 DetectEnv가 비어 env-pin
  override가 참조되지 않음. RIID-4921의 login-shell PATH 프로브(3초)·well-known dir로
  완화했으나, 그 일회성 프로브가 실패/timeout하면 `Available=false`로 캐시되어
  **재감지 없이** 재시작까지 해당 provider의 모든 task 실패.
- **보정(B):** live 디스패치 1차 거절은 handleSubmit이 아니라 supervisor eligibility
  검사 → ResultBlocked. 증상(미실행, 재시작까지 지속)은 정확.
- **근거:** `runtimeactor.go:287,370`; `daemon.go:454`(no DetectEnv);
  `claude/detect.go:22`; `detectutil.go:25,29,174`; (보정) `supervisor.go:417-422`.
- **수정:** DetectEnv 배선 + 데스크탑 PATH 핀/증강 + on-claim 재감지(TTL 캐시).
  **난이도: large.**

#### F6 — host-tier 정책이 위험 분류 툴 승인을 차단  `[B:4-E]` · partial

- **메커니즘:** runtime actor가 `TrustTierHost`로 AutoApprover/ToolStartGate 배선.
  기본 host-tier bundle은 tool-use surface를 0개 허용 → 위험 분류 command/patch
  (destructive/network/protected-path)는 Block → ResultBlocked → EventAssignmentFailed.
- **보정(B, scope 대폭 축소):** "일반 shell/edit가 승인 데드락으로 fail"은 기각 —
  비위험 shell(ls/cat/go test)·비보호 경로 edit는 차단 안 됨. codex는
  `--sandbox danger-full-access`로 spawn되어 일반 툴은 승인 hang 경로가 아님. 지배
  경로는 *차단(즉시 종료)*이지 hang이 아님.
- **근거:** `daemon.go:464,528`; `toolpolicy.go:17,89`; `session.go:221`;
  `policy/bundle.go:38-53`; `codex/command.go:88`(danger-full-access).
- **수정:** headless run용 sandbox 적합 auto-approve 라우팅 또는 승인 요청을
  control-plane으로 surface(사람 승인) + SemanticIdle watchdog(R3) 페어링. **medium.**

#### F7 — 일시적 claim/heartbeat HTTP 실패에 retry/backoff 없음  `[B:4-F]` · confirmed

- **메커니즘:** `postJSON`/`getJSON`이 요청을 5초 timeout으로 감싸고 retry 없음.
  `claimOne`은 ClaimTask 에러를 'no work'로 보아 조용히 false. 일시적 블립이나
  >5초 응답은 그 런타임의 이번 사이클 claim을 드롭. heartbeat 에러도 폐기되어
  missed heartbeat가 lease 스테일에 기여.
- **근거:** `saasplane.go:124,524`; `daemon.go:553`; `supervisor.go:324,404`.
- **수정:** bounded retry/backoff(5xx/timeout vs 4xx 구분) + RequestTimeout 구성 +
  실패 로깅/메트릭. **난이도: medium.**

#### F8 — 동기 per-task workspace prepare가 claim 직렬 경로를 블록  `[B:4-G]` · partial · **low**

- **메커니즘:** `claimOne`이 `Submit` 전에 `prepareWorkspace`(FS Prepare,
  InjectRuntimeConfig, ComputeNativeConfigVersion=파일 walk+SHA256, IR 2회 쓰기)를
  단일 supervisor 고루틴에서 동기 실행 → 그동안 다른 런타임 claim/heartbeat도 블록.
- **보정(B):** '20초 lease 예산' 수치는 데몬 코드 근거 없음. 보통 빠르며 tail/느린
  디스크 우려이지 1차 원인 아님.
- **근거:** `supervisor.go:388,425,552,579`.
- **수정:** prepareWorkspace를 claim 직렬 경로에서 분리/비동기. **난이도: medium (낮은 우선순위).**

### 4.5 Stop / Cancel 정합성 (인접 — 출처 [A] 중심)

> 사용자 4개 관심사엔 없었지만 "중지해도 SSE가 계속 동작"과 직결되며 [A]가 깊게
> 다룬 영역이다.

#### C1 — Stop이 cross-layer ACID 아님  `[A§14]` · confirmed

- **메커니즘:** stop API 성공, assignment cancel 상태, 로컬 runtime cancel, provider
  프로세스 종료, 최종 이벤트 보고, client SSE active_stream 해제가 **하나의
  트랜잭션처럼 움직이지 않는다.** control-plane은 client read model을 먼저 stopped로
  쓰고 별도 assignment store를 cancelling/cancelled로 변경. 데몬은 즉시 kill RPC가
  아니라 poll/`WatchCancellation`로 취소를 관찰. runtime cancel은 session cancel을
  enqueue만 하고 프로세스 종료를 기다리지 않음. 결과: assignment=cancelling, client
  thread=stopped, 데몬 runtime=in-flight, provider 자식=버퍼된 stdout 보유, SSE=연결
  유지가 동시에 가능.
- **근거:** client `useAiAgentTask.ts:375,376,383`,
  `AgentTaskThreadPanel.tsx:39,65,66,71`; control-plane
  `server.go:99,1071,1083,1088`, `ai_agent_client_development.go:997,1017,1018,1029`,
  `store.go:717,721,727,811,821`; daemon `saasplane.go:431`, `supervisor.go:998`,
  `runtimeactor.go:459,468,566`, `session.go:345,354,396`,
  `processexec_unix.go:22,23,24`.
- **수정:** assignment 연산을 stop SSOT로, client thread를 projection으로. "cancel
  요청"과 "프로세스 종료 확인"을 분리된 이벤트로. **난이도: large.**

#### C2 — 늦은 progress가 stopped thread를 부활 (read model에 terminal fence 없음)  `[A§15/16]` · ✅ **CONFIRMED (본 통합 검증)** · **critical**

- **메커니즘:** stop 후 늦은 `riido_log`가 assignment id로 thread를 찾는데, 조회에
  **terminal guard가 없다**(stop은 `AssignmentState=stopped`로만 바꾸고 AssignmentID는
  유지하므로 여전히 매칭됨). `EventRiidoLog` 분기가 **state-mapping switch 도달 전에**
  `WorkStatus=running`/`AssignmentState=running`/`CommentKind=runtime_progress`로 단락
  → `appendThreadProgressLocked`가 ThreadID만 보고 **무조건 덮어씀**. 그 결과
  `taskThreadHasActiveStream`이 running을 active로 보아 SSE 재개. 별도로
  `AssignmentCancelling→AgentAssignmentStateStopping`도 "active"라 stop 직후에도 stream
  유지. 데몬 reducer의 terminal 우선(`reducer.go:123-124`)은 **control-plane으로 전파
  안 됨**. exploitable window = stop 시점 in-flight 상태인 log.
- **근거:** `ai_agent_client_development.go:1438,1727,1735`(guard 없는 조회),
  `:1466,1485,1493,1494,1495`(riido_log→running, switch 전 단락),
  `:1500,2116,2119-2122,2127,2128,2129`(무조건 덮어씀), `:2978,2980`(stopping/running=active),
  `:997,1029,2213,2221`(stop이 AssignmentID 유지), `:1781,1783,1894,1896`(cancelling→stopping);
  daemon `reducer.go:56-62,123-125`, `reducer_test.go:108-122`; persistence
  `ai_agent_client_persistence.go:388-396`(guard 없이 delegate).
- **수정:** control-plane 이벤트 ingestion에 terminal fence — assignment/thread가
  stopped/cancelled/terminal이면 late `riido_log`가 active_stream을 재개하지 못하도록
  assignment id/generation/lease token 기반 fencing 또는 terminal-state guard 추가.
  **난이도: medium.**

---

## 5. 검증에서 기각 / 보정된 주장 (출처 [B] 적대적 검증)

- **(기각) "중복 데몬이 다음 실행마다 무한 누적된다"** — 고정 lock의 배타적 flock +
  소켓 응답 단락 + stop 실패 시 비스폰으로 동시 listen 데몬은 누적되지 않음. 실제는
  *단일 고아/스테일 데몬*(D1/D2) + *lock 대기 좀비*(D3). 진짜 동시 실행은
  cross-lock-path(D4)에서만.
- **(기각) "codex가 experimental opt-in 플래그 미스탬프로 자주 실패(critical)"** —
  control-plane이 실험 런타임에 `allow_experimental_runtime`을 자동 스탬프함
  (`riido-control-plane/.../server.go:853`, 데몬 보고 capability
  `ai_agent_daemon_runtime.go:464`). 정상 codex assignment는 게이트 통과. → 4의
  critical 원인이 제거되고 R3/F2(timeout)·F1(폴링)·F3(repo)이 실제 1차로 재정렬됨.
- **(보정) 최종 텍스트 재조립 위치** — control-plane 라인 concat이 아니라 **데몬
  reducer 누산**(`reducer.go:117` → `saasplane.go:482` `res.Output`)에서 일관 텍스트가
  나옴.
- **(보정) "한두 글자/줄"의 보편성** — codex 토큰 delta·cursor text에 특정. claude·
  codex `agent_message`는 큰 블록이라 미발생.
- **(보정) "댓글마다 기억 0"** — provider 세션 resume은 없지만 control-plane이 직전
  메시지를 프롬프트에 append하므로 제한적 연속성 존재(R2).
- **(보정) F5/F6 scope** — F5 1차 거절은 ResultBlocked; F6은 위험 surface만 차단(즉시
  종료, hang 아님), codex는 danger-full-access.

---

## 6. 통합 우선순위 수정 목록

severity × confidence × 사용자 체감 영향 순.

| 순위 | ID | 관심사 | 수정 | 파일 | 난이도 |
| --- | --- | --- | --- | --- | --- |
| 1 | **F3** | 느림/실패 | task를 실제 repo/worktree에 바인딩 (없으면 fast-fail) | `saasplane.go`, `workdir.go`, `assignment/types.go` | large |
| 2 | **R3/F2** | 런타임/실패 | 데몬 HardTimeout + idle watchdog (불투명 lease 만료 → 명시 timeout) | `daemon.go`, `saasplane.go`, `assignment_operation_port.go` | small |
| 3 | **C2** | stop | control-plane ingestion에 terminal fence (late log가 stopped thread 부활 금지) | `ai_agent_client_development.go` | medium |
| 4 | **F1** | 느림 | idle 폴 5→1초 / heartbeat work-waiting wake + bindings GET 캐시 | `supervisor.go`, `daemon_config.go`, `saasplane.go`, `daemonDeviceCredential.ts` | medium |
| 5 | **S3** | SSE | 클라이언트 텍스트 delta를 단일 `<p whitespace-pre-wrap>`로 concat | `AgentThreadCard.tsx` | small |
| 6 | **D4** | 중복/고아 | socket/lock/pid를 하나의 identity로, serve 전 socket liveness 검사 | `daemon.go`, `daemonLauncher.ts` | medium |
| 7 | **D1** | 중복/고아 | 종료 시 `daemon stop` 동기 발행 (will/before-quit teardown) | `main.ts`, `daemonLauncher.ts` | medium |
| 8 | **D3** | 중복/고아 | 단일 lock fast-fail ('already running' 즉시 종료) | `internal/lock/filelock*.go`, `daemon.go` | small |
| 9 | **S1/S4** | SSE | 데몬 텍스트 delta debounce/coalesce → `/thread-progress` 배치 | `supervisor.go`, `saasplane.go`, `session.go` | medium |
| 10 | **C1** | stop | stop을 cross-layer terminal로 (assignment=SSOT, client=projection) | `server.go`, `store.go`, daemon cancel 경로 | large |
| 11 | **F5/F4** | 실패 | provider on-claim 재감지(TTL) + 증강 PATH를 자식 env 전파 | `daemon.go`, `runtimeactor.go`, `detectutil.go`, `processexec.go` | large |
| 12 | **D2/D6** | 중복/고아 | spawn 자식 추적 + 종료 kill, 비정상 종료 startup orphan reaper | `daemonLauncher.ts`, `processexec_unix.go`, `daemon.go` | medium |
| 13 | **R4** | 런타임 | 재시작 PollActive 재스폰 멱등화 | `saasplane.go`, `store.go` | medium |
| 14 | **F7** | 실패 | 일시적 poll/heartbeat HTTP retry/backoff + 메트릭 | `saasplane.go`, `supervisor.go`, `daemon.go` | medium |
| 15 | **F6** | 실패 | headless 승인 모델 (sandbox auto-approve 또는 control-plane 승인) | `daemon.go`, `toolpolicy.go`, `session.go` | medium |
| 16 | **R2** | 런타임 | provider 세션/스레드 id를 후속 assignment로 전달 (진짜 연속성) | `assignment/types.go`, `saasplane.go`, `store.go` | large |
| 17 | **R5/D7/F8** | 잡정리 | 취소 watcher 채널 close / Windows `.claim` 자가치유 / prepare 비동기 | 각 파일 | small~medium |

---

## 7. 테스트 갭

- 데스크탑 launcher가 singleton 데몬 존재 시 추가 foreground start 프로세스를 남기지
  않는지 (D1/D2/D3/D4)
- daemon start가 lock busy일 때 무기한 대기 않고 status 또는 명확한 failure 반환 (D3)
- default socket과 desktop userData lock/pid/log가 같은 lifecycle identity로 묶이고,
  살아있는 소켓을 unlink하지 않는지 (D4)
- 비정상 종료 후 startup이 고아 provider 자식을 reap하는지 (D6)
- SaaS run timeout(hard/idle)이 발화하여 hung CLI를 결정적으로 종료하고 분류된
  timeout 결과를 내는지 (R3)
- 실제 repo 바인딩 없이 assignment가 들어올 때 CLI를 빈 디렉터리에서 띄우지 않고
  blocked/error 처리하는지 (F3)
- spawn된 CLI 자식이 증강 PATH를 상속하는지 (F4)
- stop 이후 late `riido_log`가 stopped/cancelled thread를 running으로 되살리지
  않는지; `AssignmentCancelling → Stopping → active_stream` 경로가 명시적으로
  fencing되는지 (C2)
- `EventTextDelta`가 raw `riido_log`가 아니라 parsed/batched `/thread-progress`로만
  사용자 progress를 갱신하는지 (S1/S4)
- local CLI detected 상태와 SaaS runtime snapshot stale 상태가 UI에서 구분되는지

## 8. 구현 전 결정 필요 (Open Questions)

**설계 결정 ([A] 중심)**

1. 데스크탑이 앱 종료 시 데몬을 shutdown해야 하는가, 아니면 데몬이 background
   agent로 앱 재시작을 가로질러 지속해야 하는가?
2. 지속한다면 lifecycle identity의 단일 진실원천은? (global socket / userData install
   root / device id / daemon id) — D4 해결의 전제.
3. `daemon start --foreground`를 lock busy 시 fast-fail시킬 것인가, 별도 "ensure
   running"(현재 status 반환) 커맨드를 둘 것인가?
4. control-plane이 assignment terminal 이후 `riido_log`를 assignment store / read
   model / 양쪽 중 어디서 거부할 것인가? (C2)
5. provider final assistant text를 progress line이 아닌 final comment/output으로
   저장할 것인가? (S4)
6. 제품 task/comment 컨텍스트의 **로컬 repository/worktree 경로**를 daemon assignment로
   운반하는 계약은? (F3)
7. 헤드리스 댓글 run의 tool-approval 모델은? (F6)

**런타임 로그/실측 필요 ([B] 중심)**

8. macOS에서 Electron 종료 시 detached/setsid 데몬과 CLI 손자 프로세스가 실제로
   살아남는가 (launch → quit → `ps`). — [A]가 2개 프로세스를 관찰해 부분 확인됨.
9. 프로덕션 reporter가 `saasplane`인가 `taskdbplane`인가; SSE store가
   `DevelopmentAIAgentClientStore`인가 `PersistentAIAgentClientStore`인가. (S 분석의
   적용 범위)
10. `DefaultAssignmentActiveLeaseSeconds`(=20)·HeartbeatEvery(=5초)·
    `RIIDO_DAEMON_IDLE_POLL_INTERVAL_SECONDS`의 실제 프로덕션/launchd plist 값. (F1/R3/R4)
11. login-shell PATH 프로브(3초)가 필드에서 얼마나 자주 실패/timeout하는가. (F5)
12. codex/claude 콜드 스타트 + initialize/thread.start/turn.start 핸드셰이크 실측
    지연. (느림)
13. 데스크탑 lockFilePath(userData)와 수동 `riido daemon start`(`~/.riido/.lock`)가
    동시 존재 가능한 실제 시나리오 빈도. (D4)

## 9. 구현 게이트

`riido-daemon/AGENTS.md`는 SSOT-document-first 변경을 요구한다. 본 통합 문서는
documentation-only 기록이며 동작을 변경하지 않는다. 이 리뷰에서 파생되는 모든 동작
변경은 같은 PR에서 소유 도메인/아키텍처 문서를 갱신해야 하고, Riido task 생성 응답의
`branchName`을 Git 브랜치명으로 사용해야 한다.

---

> 통합 출처: `[A]` `docs/review/ai-agent-runtime-lifecycle-review-2026-06-08.md`,
> `[B]` `RIIDO_DAEMON_ISSUES.md`. ✅ 표시 3건(D4·F3·C2)은 본 통합 과정에서 코드로
> 독립 재검증되어 CONFIRMED.

---

# 부록: 구현 기록 (Fix Log)

> 위 진단을 바탕으로 한 실제 수정 내역. 우선순위 표 순서대로 하나씩, 근본적으로
> 해결한다. 각 항목은 **근본 원인 → 근본 접근(왜 이 레이어인가) → 변경 파일 →
> 기본값/근거 → SSOT 문서 → 테스트 → 검증** 으로 기록한다. 모든 변경은 현재 작업
> 브랜치(`JYM-ai-get-done`)에서 진행한다.

## R3/F2 — 데몬 run timeout (hard/idle) + 분류된 실패  ✅ 완료 (2026-06-08)

**상태:** 구현 완료, `go test ./...` 36 ok / 0 fail. 미커밋.

### 근본 원인 (재확인)
세션 액터(C4)는 `HardTimeout`/`SemanticIdle` run-clock 머신러리와
`EventTimeout → ResultTimeout + CommandCancelProvider(프로세스 kill)`를 **이미
완비**하고 있었고(`internal/agentbridge/session/session.go:172-197,356-360`), SSOT
(`provider-runtime.md §5.5/§5.6`)도 "C4 세션이 run clock 소유"를 명시한다. 진짜
문제는 **프로덕션 데몬 배선에서 두 값이 0(비활성)으로 들어가던 것**이다:

- `runtimeactor.Config.HardTimeout`을 `newDaemonRuntimeActor`가 설정하지 않음 → 0.
- `runtimeactor.Config`에 **`SemanticIdle` 필드 자체가 없어** task request에서만 옴
  → `taskRequestFromAssignment`도 미설정 → 0.

그 결과 hung CLI가 무한 실행되고, 유일한 backstop인 control-plane 20초 active lease
가 불투명한 `active assignment lease expired`(liveness 실패)로만 종료시켰다.

### 근본 접근 (왜 이 레이어인가)
값만 박지 않고 레이어·소유권을 정리했다.

- **프로덕션 경로를 먼저 확정:** supervisor → `runtimeactor.Actor` → `handleSubmit`
  → `session.Start`. `bridge.Coordinator`(이미 `DefaultTimeout`/`DefaultSemanticIdle`
  보유)는 프로덕션 어디에서도 생성되지 않는 미사용 경로 → 그걸 고치는 함정을 피하고
  `runtimeactor` 레이어를 고쳤다.
- **대칭 설계:** 기존 `HardTimeout` fallback 패턴에 맞춰 `SemanticIdle`도 actor-config
  fallback으로 추가. per-task(`TaskRequest`) 값이 항상 우선, actor-config는 기본값.
- **소유권 분리:** C4 run clock = 권위적 provider-run 타임아웃, control-plane lease =
  liveness. 데몬 clock이 먼저 발화해 **분류된** 종료를 보고하도록 했다.

### 변경 파일
| 파일 | 변경 |
| --- | --- |
| `internal/agentbridge/runtimeactor/runtimeactor.go` | `Config.SemanticIdle` 필드 추가(`HardTimeout`과 대칭); `handleSubmit`에 idle fallback `if idle <= 0 { idle = a.cfg.SemanticIdle }` |
| `cmd/riido/daemon.go` | `newDaemonRuntimeActor`가 `HardTimeout: settings.RunHardTimeout`, `SemanticIdle: settings.RunSemanticIdle` 배선 |
| `cmd/riido/daemon_config.go` | env 상수 2개 + `defaultRunHardTimeout`/`defaultRunSemanticIdle` + `daemonSettings.RunHardTimeout`/`RunSemanticIdle` + 파싱 + 헬퍼 `parseOptionalDurationSecondsWithDefault`(empty→기본값, `0`→비활성, 음수/오류→reject) |
| `internal/agentbridge/session/session.go` | timeout 메시지에 clock 종류+기간 명시 (`run hard timeout exceeded (%s)` / `no provider progress for %s (semantic idle timeout)`) |

### 기본값 / 근거
- `RIIDO_DAEMON_RUN_HARD_TIMEOUT_SECONDS` 기본 **1800(30m)**, `0`=비활성.
- `RIIDO_DAEMON_RUN_SEMANTIC_IDLE_SECONDS` 기본 **600(10m)**, `0`=비활성.
- **보수적으로 크게** 잡은 이유: 사용자 불만이 "실패가 너무 많음"이므로 과도한
  timeout이 *오히려 실패를 유발*하지 않게 했다. idle은 `IsSemanticActivity`
  (text/tool/usage/progress가 리셋)라 긴 *silent* 빌드는 살리고 *진짜 무진행*만 잡는다.

### 분류된 실패
`CompleteTask`(`saasplane.go:474-490`)는 `ResultTimeout`을 `default` 분기로
`EventAssignmentFailed` + `message=res.Error`로 매핑한다. 세션 메시지를 명시화했으므로,
데몬 timeout이 켜지면 불투명한 lease 만료 대신 **"run hard timeout exceeded (30m0s)"
/ "no provider progress for 10m0s (semantic idle timeout)"** 가 사용자에게 전달된다.

### SSOT 문서
- `docs/20-domain/provider-runtime.md` **§5.7 신설** — 데몬 run-clock 기본 정책이
  필수임 + env knob + `0`=비활성 + C4 clock(권위) vs control-plane lease(liveness) 관계.
- `docs/30-architecture/config-reference.md` — env 표에 2행 추가.

### 테스트
- `runtimeactor_test.go`: `TestRuntimeActorAppliesConfigSemanticIdleTimeout`(actor-config
  idle이 세션까지 도달해 `ResultTimeout` 발화 + 메시지 검증),
  `TestRuntimeActorTaskRequestSemanticIdleOverridesConfig`(per-task 우선순위).
- `daemon_test.go`: `TestLoadDaemonSettingsRunClockDefaultsAndOverrides`(기본값/override/
  `0`-disable/invalid reject).

### 검증
- `gofmt -l` / `go vet` 클린, `go build ./...` OK, **`go test ./...` 36 ok / 0 fail**.
- CI 게이트 로컬 재현 통과: `approval-timeout-ssot`(grep + 타깃 테스트),
  `architecture-docs`(split-repo wording + config-key 커버리지).
- 변경 footprint: 8파일 +208/−3.

## F1 — 댓글→실행 지연: long-poll claim (접근 C)  ✅ 완료 (2026-06-09)

**상태:** 3개 레포 구현 완료. contracts/control-plane/daemon 각각 `go test ./...`
all green. 미커밋. 로컬 빌드는 `replace` 디렉티브로 연결(아래 롤아웃 주의).

### 근본 원인 (재확인)
데몬이 idle 일 때 새 assignment 는 다음 idle 폴(`IdlePollEvery` 기본 5초)까지
0~5초 대기. 데몬은 "일감 생김" 통보 경로가 없고 폴링으로만 발견(push/long-poll
부재). dynamic 모드는 폴마다 bindings GET+poll POST 2왕복.

### 근본 접근 (왜 long-poll 인가)
"폴링으로 발견"을 "서버가 hold 후 즉시 응답"으로 바꿈 — 지연은 RTT 수준으로
떨어지고 idle 부하는 오히려 감소. cross-repo(contracts+control-plane+daemon)이며
`DisallowUnknownFields` 때문에 **server-first 롤아웃**(contracts 태그 → CP → daemon)
이 강제됨. 정찰 워크플로우로 양쪽 표면(store actor, queue 신호 지점, SSE/배포
timeout)을 매핑한 뒤 설계.

### 변경 — contracts (v0.3.3 기반 작업 브랜치 `f1-longpoll-wait-ms`)
| 파일 | 변경 |
| --- | --- |
| `assignment/types.go` | `PollRequest.WaitMs int json:"wait_ms,omitempty"` 추가(additive; omit 시 legacy 와 바이트 동일) |
| `docs/20-domain/assignment-polling.md` | `wait_ms` 서버-클램프 hint·degrade·롤아웃 순서 문서화 |
| `assignment/types_test.go` | omit=legacy 바이트 동일 + 라운드트립 compat 테스트 |
검증: 전체 test, `apicontract verify`, `ssotdeps verify`, stdlib-only(1) 통과.

### 변경 — control-plane (`main`)
| 파일 | 변경 |
| --- | --- |
| `store_state.go` | `agentWaiters map[agentID]map[int64]chan struct{}` + seq |
| `store.go` | per-agent waiter(register/unregister/signal, subscribe 패턴 복제); `signalAgentWaiters` 를 `handleAssign`·`cancelQueuedBlockerForAssignment`·`handleCancelAssignment` 에서 발사; `WaitForAssignment`(actor **밖** wait 루프: 즉시평가 → register → 재평가 → signal/tick/deadline/ctx); poll 메트릭을 held poll 당 1회로 분리(`countRequest`) |
| `store_port.go` | optional `AssignmentLongPollStore` 인터페이스(auto-degrade) |
| `server.go` | `handleAgentPoll` 가 `wait_ms>0` 시 `WaitForAssignment` 로 분기 + budget clamp; `ServerConfig.LongPollMaxHold/Tick`(기본 25s/2s) |
| `cmd/riido_ai_server/main.go` | `RIIDO_AI_SERVER_LONGPOLL_{MAX_HOLD,TICK}_SECONDS` env 배선 |
| docs `saas-control-plane.md`, `config-reference.md` | long-poll 섹션 + env 행 |
| `assignment_longpoll_test.go` | 즉시반환/assign-wake/timeout-none/ctx-cancel/actor-non-block/HTTP clamp 테스트 |
멀티 인스턴스(DynamoDB): in-process signal + 2초 fallback 틱(`agent_queue` GSI
재평가)로 cross-task 발견 ≤2초. http.Server 는 `ReadHeaderTimeout` 만 설정(SSE 도
이미 무한 hold) → 변경 불필요. ALB idle 60s ≫ 25s. 검증: `go test ./...` all green.

### 변경 — daemon (`JYM-ai-get-done`)
| 파일 | 변경 |
| --- | --- |
| `internal/agentbridge/supervisor/supervisor.go` | **claim 을 run goroutine 밖 per-runtime claim goroutine 으로 이동**(블로킹 long-poll 이 heartbeat 를 굶기거나 runtime 을 직렬화하지 않음); free-gate capacity token(`MaxConcurrent:1`); `claimOne`→`startClaimedTask`(run loop, mailbox `claimed` case); poll 타이머 제거; **elapsed≥1s 면 즉시 repoll, 아니면 IdlePollEvery backoff**(fast-none degrade) |
| `internal/agentbridge/controlplane/saasplane/saasplane.go` | poll 에 `wait_ms` 전송 + 전용 `LongPollTimeout`(기본 30s); `http.Client.Timeout` 제거(held poll 안 잘리게); `postJSONTimeout` 도입(heartbeat/events 는 `RequestTimeout` 유지) |
| `cmd/riido/daemon_config.go` | `RIIDO_DAEMON_CLAIM_WAIT_MS`(20000, `0`=비활성)·`RIIDO_DAEMON_LONGPOLL_TIMEOUT_SECONDS`(30) |
| `cmd/riido/daemon.go` | saasplane 에 `ClaimWaitMs`/`LongPollTimeout` 전달 |
| docs `runtime-scheduling.md`, `config-reference.md` | long-poll 데몬측 섹션 + env 행 |
| 테스트 | saasplane: wait_ms 전송 + poll 가 RequestTimeout 아닌 LongPollTimeout 사용(250ms>50ms 성공)·미설정 시 omit; supervisor: held claim 이 heartbeat 안 굶김 |
검증: `gofmt`/`vet` 클린, `go build ./...`, **`go test ./...` 36 ok / 0 fail**,
arch-docs gate(stale wording + no-non-riido-deps) 통과.

### 호환성·롤아웃 (중요)
- additive `wait_ms` + optional-interface auto-degrade + 데몬 fast-none degrade 로
  (구daemon↔신CP), (신daemon↔신CP), (신CP feature-off) 모두 안전.
- **단, 신daemon→구CP 는 `DisallowUnknownFields` 로 400** → 롤아웃 순서 강제:
  contracts 태그 → CP import+배포 → daemon import+배포.
- **로컬 빌드용 `replace github.com/teamswyg/riido-contracts => ../riido-contracts`
  가 daemon·control-plane go.mod 에 추가됨(DEV 전용).** 머지/배포 전에 제거하고
  실제 `go get riido-contracts@<태그>` 로 교체해야 함.

### 후속(이번 범위 밖)
- DynamoDB Streams→EventBridge 로 cross-instance signal fan-in → 2초 틱 제거.
- bindings GET 캐시(현재 long-poll 로 ~25초당 1회라 hot path 아님).

## C2 — 중지 후 늦은 progress의 thread 부활 차단 (terminal/stop fence)  ✅ 완료 (2026-06-09)

**상태:** control-plane 단독 구현 완료. `go test ./...` all green. 미커밋.
(F1 과 독립 — contract 변경 불요. F1 의 dev `replace` 와는 무관하게 단독 성립.)

### 근본 원인 (재확인, 현재 코드)
stop(`StopAIAgentTask`→`markTaskAgentThreadsStoppedLocked`)은 thread 를 `Stopped`로
만들되 `AssignmentID`는 유지. 늦은 `riido_log`가 같은 assignment id 로 도착하면
`taskThreadForAssignmentLocked`가 **terminal guard 없이** 그 stopped thread 를 찾고,
`RecordAIAgentAssignmentEvent`의 riido_log 분기가 thread 상태를 보지 않고 무조건
`WorkStatus/AssignmentState=Running` 으로 flip → `appendThreadProgressLocked` 무조건
덮어쓰기 → `taskThreadHasActiveStream`이 running 을 active 로 보아 SSE 재개 + 에이전트
재잠금. `/thread-progress` 배치 핸들러(`RecordAIAgentThreadProgress`)도 동일 버그.
데몬 reducer 의 terminal 가드는 control-plane read model(별개 projection)로 전파 안 됨.

### 근본 접근
읽기모델 불변식을 박음: **"runtime progress 는 active(queued/running) thread 만
전진시킨다."** 단일 판정자 + 두 ingestion 경로 fence + 최저수준 방어 가드.

### 변경 — control-plane (`main`)
| 파일 | 변경 |
| --- | --- |
| `internal/riidoaiserver/ai_agent_client_development.go` | `agentAssignmentStateAcceptsRuntimeProgress(state)`(= queued/running) 추가; **Path1** `RecordAIAgentAssignmentEvent` riido_log 분기 상단에서 `hadThread && !accepts(previousThread)` 면 `return nil`(드롭); **Path2** `RecordAIAgentThreadProgress`가 stopped thread 면 `AcceptedLines:0` no-op 반환; **방어 가드** `appendThreadProgressLocked`가 non-accepting thread 면 `return`(다른 호출자도 불변식 보장) |
| `docs/20-domain/ai-agent-client-api.md` | stop/terminal fence 규칙 명시(SSOT) |
| `ai_agent_client_stop_fence_test.go` | stop→늦은 riido_log/배치 둘 다 드롭(stopped 유지·active_stream 안 열림·에이전트 안 잠김·line 미추가) + running thread 정상 progress 회귀 |

### 처리 정책
**A(드롭)** — stop 직후 자투리 텍스트는 버림(이미 스트리밍된 본문은 보존). stop UX
("지금 멈춤")에 충실, 최소·저위험. `stopping`도 fence 대상(stop 진행 중 되돌림 방지).

### 검증
`gofmt` 클린, `go build ./...`, **`go test ./...` all green**, ai-agent-client-api
게이트(생성 클라이언트 drift)는 read-model API shape 무변경이라 영향 없음.

### 후속(이번 범위 밖)
- 늦은 **비-riido_log** 이벤트(예: state=running)의 `assignmentEventActionResponse`
  경로 재활성 가능성 — 확정된 riido_log/배치 경로만 막음. action 경로는 별도 검토.

## S1+S3 — SSE 글자단위/줄바꿈: 데몬 텍스트 coalescing + 클라이언트 단일 블록 렌더  ✅ 완료 (2026-06-09)

**상태:** 데몬(`riido-daemon`) + 클라이언트(`riido-client`) 구현 완료. 각 `go test`/
`jest` all green, tsc·eslint(변경 파일) 클린. 미커밋. 계약/control-plane 변경 불요.
(/thread-progress 라우팅(S4)·서버측 coalescing(S2)은 후속.)

### 근본 원인 (재확인)
상류가 provider token 단위 `EventTextDelta`마다 progress line 1개를 만들고(데몬→CP
토큰당 POST), 클라이언트가 각 line을 block `<p>`로 렌더 → "한두 글자씩 줄바꿈" +
요청 storm(클라 throttle은 증상 완화일 뿐).

### 근본 접근 — 소스 buffer(S1) + 시각 마무리(S3), 상보적
1. **데몬 coalescing(S1)** 으로 토큰을 청크로 합쳐 보고(폭주↓, provider cadence↔보고
   분리=async). 2. **클라이언트 concat(S3)** 으로 청크 경계도 한 블록으로 흘림.

### 변경 — daemon (`JYM-ai-get-done`)
| 파일 | 변경 |
| --- | --- |
| `internal/agentbridge/supervisor/supervisor.go` | per-task `forwardSession`에 텍스트 delta coalescer: 버퍼 + flush 조건 = size(`TextFlushBytes`) / **max-interval 타이머**(`TextFlushInterval`, debounce 아님) / 비텍스트 이벤트 직전(순서 보존) / 종료. `Config.TextFlushBytes`·`TextFlushInterval`(둘 0이면 off=passthrough) |
| `cmd/riido/daemon.go`, `daemon_config.go` | `RIIDO_DAEMON_TEXT_FLUSH_BYTES`(256)·`RIIDO_DAEMON_TEXT_FLUSH_MS`(200) env, `0`=비활성; ms 파서 추가 |
| docs `runtime-scheduling.md`, `config-reference.md` | coalescing 보고 규칙 + env 행 |
| 테스트 | coalesce(타이머)·size-flush·비텍스트 boundary flush·terminal flush·passthrough(off); config defaults/override/disable/invalid(claim-wait/long-poll/text-flush 통합) |
`EventProgress`(structured `<riido_log>` 파생)는 비텍스트로 취급 → 버퍼 flush 후 별도
보고 → 답변 텍스트와 상태가 안 섞임.

### 변경 — client (`JYM-ai-get-done`)
| 파일 | 변경 |
| --- | --- |
| `src/components/domain/aiAgentTask/agentThinkingLog.ts` (신규) | 순수 selector: `selectStreamedText`(thread.lines+라이브 progress lines seq순 concat, 구분자 없음)·`selectStatusMessages`(progress_messages+work_status 라벨 분리)·`tailLines`(컴팩트 미리보기) |
| `AgentThreadCard.tsx` | `AgentThinkingLog`가 스트리밍 텍스트를 단일 `<p className="whitespace-pre-wrap break-words">`로(모델 실제 줄바꿈 보존, 토큰 경계 줄바꿈 제거), 상태는 별도 라인. 메인 메시지 `<p>`에도 `whitespace-pre-wrap` |
| `agentThinkingLog.test.ts` (신규) | 순수 selector 8 케이스 |
클라이언트 `AgentThreadProgressLine`엔 `message_code`가 없어(상류에서 렌더됨) line
레벨 텍스트/상태 구분 불가 → **소스 단위**(lines=텍스트 스트림, progress_messages/
work_status=상태)로 분리.

**추가 — raw `<riido_log>` 마커 strip (v0.0.18 데몬 대응):** 디스크 IR 로그로 확인 —
provider 가 AGENTS.md 지시대로 `<riido_log>{...}<end>` 를 **토큰 단위 TextDelta**로
내보내고(예: `<ri`/`ido`/`_log`/`>{"`…), 데몬(v0.0.18)이 이를 파싱(LogLine)은 하되
**화면 텍스트에서 제거하지 않아** 그대로 새어 화면에 마커가 노출됨. concat 하면 원문이
정확히 복원되므로(`<riido_log>{"code":1001,…}<end>…`), `selectStreamedText` 가 concat
**후** `stripRiidoTelemetry` 로 완전 블록 + 미완(열림/부분 opener) 마커를 제거 →
깨끗한 답변만 표시(실제 데이터 694자 raw → 350자 클린 한국어 답변으로 검증). 배포된
v0.0.18 바이너리는 못 바꾸므로 클라이언트가 방어적으로 strip. (정식 수정은 데몬
telemetry 파서가 추출과 동시에 텍스트에서 마커 제거 — 후속.)

**추가 — persisted/live 중복 dedup:** UI에 텍스트·상태가 2~3중으로 중복되는 현상
발견. 원인: 클라이언트가 `thread.lines`(GET persisted)와 `streamEvents`(SSE live)를
**둘 다 합치는데 둘이 같은 control-plane seq 공간을 공유(겹침)**. 기존 코드는
`slice(-6)`으로 마지막 6줄만 보여 가려졌고, S3 concat이 전체를 이어붙여 중복이
드러남. 수정: `selectStreamedText`는 **seq Map으로 병합·dedup**(겹친 line 1회만),
`selectStatusMessages`는 **연속 중복 라벨 collapse**(상태 3중복 → 1). jest 15 통과
(overlap-dedup·연속중복 케이스 포함), tsc·eslint 클린(Map iterator는 `Array.from` 사용).

### 검증
daemon `go test ./...` 36 ok/0 fail, gofmt·arch-docs(stale wording) 통과;
client `jest`(8) 통과, tsc·eslint(변경 파일) 클린(레포 baseline tsc 에러는 무관한 기존분).

### 후속(이번 범위 밖)
- S2/S4: control-plane 측 coalescing + 데몬→`/thread-progress` 배치 라우팅(SSOT 정합).
  현재는 데몬 coalescing으로 `/events` 청크 보고(계약 변경 회피).

## F3 (1차) — repo 없는 빈 workdir: 에이전트가 판단·요청하게 (fail-fast guidance)  ✅ 완료 (2026-06-09)

**상태:** daemon 단독 구현 완료. `go test ./...` 36 ok/0 fail. 미커밋. 계약/CP 변경 불요.
F3 **풀 바인딩(clone/worktree/mwsd resolve)은 여전히 미구현** — 이건 그 전 단계.

### 디스크 증거 (이 머신)
`~/Library/Application Support/riido/workspaces/<id>/tasks/<id>/runs/<asn>/workdir/`
가 provider CLI cwd인데 **소스 코드 0** (주입된 `AGENTS.md`+`.riido`뿐). `workspace_id ==
task_id`(ComponentID 비어 TaskID 폴백). IR 로그에서 LLM이 직접 "수정할 코드
리포지터리는 없습니다"라고 인지. 즉 코딩 task가 빈 곳에서 헛돎.

### 근본 접근 (말만 — A안)
데몬은 "이 task가 repo 필요한지" 모름(신호 없음) → **하드 차단 금지**(repo 불필요
task까지 막힘). 대신 데몬이 이미 쓰는 AGENTS.md/CLAUDE.md(native config)에 가이던스를
주입해 **LLM이 판단**하게: 코딩 task면 빈 디렉터리에서 추측 말고 "이 경로(`<workdir>`)에
프로젝트를 두거나 새로 만들지" 사용자에게 묻고 멈춤; 비코딩 task면 그대로 진행.

### 변경 — daemon (`JYM-ai-get-done`)
| 파일 | 변경 |
| --- | --- |
| `internal/workdir/workdir.go` | `RuntimeConfig.WorkdirGuidance` 필드 + `renderRuntimeConfig`에 "## Working directory" 섹션 렌더 |
| `internal/agentbridge/supervisor/supervisor.go` | `prepareWorkspace`가 `workdirHasWorkContent`(주입 config 외 내용물 감지)로 repo 부재 시 `noRepoWorkdirGuidance(<workdir 절대경로>)`를 주입. 가이던스는 "먼저 코드베이스 필요 여부 판단 → 필요하면 사용자에게 경로 안내+생성 제안, 불필요하면 진행" |
| docs `workspace.md §5.1` | no-repo 가이던스 주입 규칙 (repo mount 시 자동 비활성) |
| 테스트 | workdir: 가이던스 set 시 AGENTS.md에 섹션 렌더/미set 시 부재; supervisor: `workdirHasWorkContent`(빈/config만/실파일)·`noRepoWorkdirGuidance`(경로+판단 문구) |

**자동 비활성:** `workdirHasWorkContent`가 repo mount 후 내용물을 감지하면 가이던스가
사라짐 → F3 풀 바인딩이 들어오면 자연 무효화.

### 검증
gofmt·`go build`·**`go test ./...` 36 ok/0 fail**, `workdir-policy-ssot`(Q-WS 문구
보존 + go test 서브셋)·arch-docs(stale wording) 통과.

### 후속 = F3 풀 바인딩 (의도적으로 다음으로 미룸 — 2026-06-09)

**결정 A 확정 (recon 발견):** repo 정체성은 **이미 control-plane에 존재**.
`ai_agent_assignment_prompt.go`의 `AIAgentTaskContext`가
`Component.BranchName` + `Repositories[]{ FullName, IsPrivate, RepositoryURL,
Source(connected_pull_request|workspace_connected_repository) }` 를 갖고, 지금은
**프롬프트 텍스트에만** 주입(`repository_url:`/`branch_name:`). mwsd 불필요(이 머신엔
mwsd 부재). 즉 **이미 있는 URL/branch를 구조화해 데몬까지 흘리면 됨.**

**구현 체인(3-레포):**
1. contracts: `Assignment`/`AssignRequest`에 repo 필드 추가(additive) —
   `RepositoryURL`/`RepositoryFullName`/`RepositoryIsPrivate`/`BranchName`(+source).
2. control-plane: assignment 생성 시 task context의 repo 값을 구조화 필드로도 채움
   (지금 프롬프트에 쓰는 그 값).
3. daemon: `taskRequestFromAssignment` → `MountRepo`(clone→workdir에 worktree/shallow
   clone, branch=BranchName) → cwd=코드 있는 workdir. clone 실패 시 fail-fast(F3 1차와 연결).

**진짜 crux = git 인증.** 데몬엔 git clone/worktree/인증 코드 전무.
`IsPrivate:true` clone하려면 (i) control-plane이 단기 토큰 발급(GitHub App 등)해
assignment에 실어줌 vs (ii) 사용자 로컬 자격증명. **recon 필요 항목**: 제품의 git
인증 메커니즘, `Repositories[]` 출처(api-server), source별 ref 규칙.

**점진 v1 권장:** public repo는 shallow clone로 체인 end-to-end 검증, private는
"인증 미지원" fail-fast → v2에서 private 인증. mount 모드는 v1 shallow clone(격리·
단순), worktree+cache(workspace.md §4)는 후속. 결과 write-back(commit/push/PR)은 별개.

## D3+D4 — 데몬 single-instance + 단일 lifecycle identity  ✅ 완료 (2026-06-09)

관심사 #1(데몬 중복/고아)의 **토대(foundation)** 항목. 사용자 결정: 데몬은 background
영속 금지(앱 종료 시 동반 종료), **모든 근본 옵션 채택** = 방향 3(lock fast-fail +
socket connect-probe + PID liveness) + (b) app-data 단일 identity + PID-liveness stale 회수.

### 근본 원인 (재확인, 현재 코드)
- **D3:** foreground 진입점이 `c9lock.AcquireFile(ctx, …)`(blocking)로 lock 을 기다렸다.
  prod ctx 는 취소되지 않으므로(`runDaemon → context.Background()`), 두 번째 daemon 은
  영원히 lock-waiter 좀비로 블록될 수 있었다.
- **D4:** lock 기본경로(`$HOME/.riido/.lock`)와 socket 기본경로
  (`~/Library/Application Support/riido/agentd.sock`)가 **서로 다른 디렉터리**였고,
  `serveAgentDaemon` 은 무조건 `os.Remove(flags.socket)` 후 listen → 경로가 다른 두
  daemon 이 같은 socket 을 hijack(살아 있는 socket unlink)해 서로를 고아로 만들 수 있었다.

### 근본 접근
1. **단일 identity (b):** `defaultDaemonAppDataRoot()` 헬퍼로 socket·lock·pid 를 모두
   같은 app-data root 아래로 통일(`daemon.lock`/`daemon.pid` = socket 과 같은 디렉터리).
   manual start 와 desktop-launched daemon 이 **같은** 단일 lock 으로 수렴.
2. **D3 fast-fail + liveness:** `internal/lock` 에 `TryAcquireFile`(non-blocking) +
   `ErrLocked` + `RemoveStaleLock` 추가. `acquireDaemonSingleton` 이 `TryAcquireFile`
   →`ErrLocked` 시 pid liveness 확인 → 살아 있음/불명 = "already running" 깨끗이 종료
   (return nil, 좀비 아님), **확실히 죽은 경우에만** stale 회수+1회 재시도.
   Unix flock 은 죽으면 자동 해제되므로 `ErrLocked ⟺ 살아 있는 owner`(회수 경로는 사실상
   Windows `.claim` 전용). 항상 pid 파일을 기록하고 release 시 정리.
3. **D4 anti-hijack:** `serveAgentDaemon` 이 `os.Remove(socket)` **전에** connect-probe
   (`daemonSocketServing` = `net.DialTimeout("unix", …)`) → 살아 있는 socket 이면 unlink
   하지 않고 step-aside(return nil).
4. **liveness 보수성:** `daemonPIDProbablyAlive` 는 ESRCH/`ErrProcessDone` 만 "죽음"으로
   판정, EPERM·기타 모호한 결과는 "살아 있음"으로 본다(살아 있는 lock 을 빼앗지 않음).
   Windows 는 신뢰 가능한 liveness 부재 → 항상 alive 반환(이중 daemon 방지; D7 별도 추적).

### 변경 — daemon (`JYM-ai-get-done`)
- `internal/lock/filelock.go`: `ErrLocked`, `TryAcquireFile`, `RemoveStaleLock`.
- `cmd/riido/daemon.go`: `defaultDaemonAppDataRoot`(공유), `defaultDaemonLockPath`/
  `defaultDaemonPidPath`(app-data), foreground 의 pid 기본값+always-write,
  `acquireDaemonSingleton`/`readDaemonPID`/`daemonSocketServing`, serve connect-probe.
- `cmd/riido/daemon_process_unix.go`/`_windows.go`: `daemonPIDProbablyAlive`.

### SSOT 문서
- `docs/20-domain/locking.md` §1: `TryAcquireFile`/`RemoveStaleLock` primitive + "daemon
  single-instance (D3)" 절(identity 통일·clean-exit·보수적 회수·앱 종료 동반).
- `docs/30-architecture/config-reference.md`: `--lock-file`/`--pid-file` 기본값을 app-data
  root 로 갱신 + identity 통일/clean-exit 설명, locking SSOT 링크.

### 테스트
- `internal/lock/filelock_test.go`: `TryAcquireFile` ErrLocked/재획득, empty-path,
  `RemoveStaleLock` no-op.
- `cmd/riido/daemon_singleton_test.go`: `acquireDaemonSingleton`(획득→already-running→
  release 후 재획득), `daemonSocketServing`(listener 유무).
- `cmd/riido/daemon_singleton_unix_test.go`(`!windows`): `daemonPIDProbablyAlive`(live/
  dead/0), stale-pid 라도 **살아 있는 flock 은 빼앗지 않음**.
- 기존 `TestDaemonStartHoldsSingletonLock` 을 새 계약(두 번째 start = 에러 아닌 clean-exit
  + 두 번째 listener 미생성)으로 갱신. foreground 테스트들에 `--pid-file`(tempdir) 추가 —
  실제 app-data `daemon.pid` 를 건드리지 않도록 위생 보강.

### 검증
- `go build ./...` + `GOOS=windows go build` OK, `go vet` OK, `go test ./...` 36 pkg 전부 통과.

### 후속(이번 범위 밖, 관심사 #1 잔여)
- **D1/D2** (desktop: 앱 종료 시 daemon kill + probe-and-adopt) — riido-desktop 미착수.
- **D6** (daemon SIGKILL 시 고아 CLI child reaper; `Setpgid:true` 이미 설정됨).
- **D7** (Windows `.claim` 자동 회수 — 신뢰 가능한 Windows liveness 필요; 현재 보수적 latent).

## D1 + D2/D6 — 앱 종료 시 데몬 동반 종료 + 자식 추적/orphan reaper  ✅ 완료 (2026-06-09)

관심사 #1의 desktop↔daemon 수명주기 마감. 사용자 결정: 데몬은 백그라운드 영속 금지(앱
종료 시 동반 종료) / **(1c) 하이브리드** / will-quit+preventDefault **5초** 후 강제
`app.exit(0)` / **(3c) 지속 pgid 레지스트리** reaper.

### 근본 원인 (재확인, file:line)
- **D1:** `riido-desktop/src/main.ts:606` `app.once('before-quit', controller.stop)` 의
  `stop()`(`daemonLauncher.ts`)은 **폴링만 정지**, 데몬은 안 죽임. 데몬은
  `spawn(...{detached:true}) + child.unref()` 로 부모와 수명 분리 → 앱 꺼져도 잔존.
  종료 이벤트도 `before-quit`만, `will-quit` 없음(async teardown 불가).
- **D2:** spawn한 데몬 ChildProcess 참조/PID를 버림(`unref`) → 죽일 핸들 없음. 단
  협조적 `stopDaemonIfRunning`(`riido daemon stop --socket --pid-file`)은 이미 있으나
  종료에 안 엮임.
- **D6:** provider CLI는 자기 pgid(`Setpgid`)로 떠 session 종료/ graceful daemon
  shutdown 시 그룹 kill(`terminateCommand`)로 정리되지만, 데몬 **SIGKILL/크래시** 시
  자식이 reparent되어 고아로 남고, **startup orphan reaper 부재**.

### 근본 접근
- **daemon (D6):** `internal/process/childreg` — spawn한 provider pgid를 pid 파일과
  같은 디렉터리(`daemon-children.pids`)에 지속(spawn 기록/exit 제거 → 살아있는 그룹만
  잔존). `processexec`에 `ChildObserver`(OnSpawn/OnExit) 추가, `NewWithObserver` 로
  레지스트리 주입(`daemon.go` `newDaemonRuntimeActor`). serve 진입 시(singleton lock
  확보 후) `childreg.ReapOrphans` 가 이전 인스턴스의 살아있는 그룹을 `SIGKILL` 후 파일
  reset. graceful 자식 kill은 기존 경로 유지. Windows는 process group 부재로 reaper
  no-op(D7).
- **desktop (D1+D2):** `daemonLauncher.ts` `stopDaemonNow` — 협조적
  `stopDaemonIfRunning` 먼저 → 실패/잔존 시 **pid 파일의 PID로 SIGTERM→(대기)→SIGKILL**
  강제 폴백(spawn/adopt 양쪽 커버, 같은 pid 파일 사용). `main.ts` `will-quit` 에서
  `event.preventDefault()` 후 `Promise.race([stopDaemonNow(), 5s])` → 성공/실패 무관
  `app.exit(0)`. 폴링 정지(before-quit)가 teardown 중 재기동을 막음.

### 변경 파일
- daemon (`JYM-ai-get-done`): `internal/process/childreg/{childreg.go,_unix.go,_windows.go,_test.go,_unix_test.go}`(신규), `internal/process/processexec/processexec.go`(observer), `cmd/riido/daemon.go`(reaper+레지스트리 wiring+`childRegistryPath`), `daemon_test.go`(시그니처), `docs/20-domain/provider-runtime.md` §5.1.1(SSOT).
- desktop (`JYM-ai-get-done`): `src/modules/daemonLauncher.ts`(`stopDaemonNow`+force-kill 헬퍼), `src/main.ts`(`will-quit` teardown).

### 검증
- daemon: `go build ./...` + `GOOS=windows` OK, `go vet` OK, `go test ./...` 전 패키지 통과(childreg가 실제 process-group SIGKILL 검증).
- desktop: `tsc -b` 통과, `eslint --quiet` 0 errors. (riido-desktop은 test runner 부재 → tsc+eslint+로직 검토로 검증.)

### 후속(이번 범위 밖)
- **D2 probe-and-adopt 고도화** / 비정상(SIGKILL) 종료 시 desktop은 OS가 정리(자식은 다음 기동 D6 reaper가 처리).
- **D7** (Windows pgid/`.claim` reaper) — process group 부재로 현재 no-op.

## C4/Codex — rate-limit notification 노출 + 장시간 대기/불투명 실패  ✅ 1차 수정 (2026-06-09)

사용자 관찰:

```text
중지
codex rate limits updated
```

### 근본 원인
- `origin/main`의 Codex translator는 app-server 내부 notification
  `account/rateLimits/updated` / `account_rate_limits_updated`를 user-visible
  `EventLog("codex rate limits updated")`로 변환한다.
- daemon SaaS reporter는 `EventLog`를 `EventProviderLog`로 올리고, control-plane read
  model은 terminal/stopped thread fencing이 없거나 약하면 이 provider log를 thread
  message로 반영할 수 있다.
- 따라서 이 문구는 "Codex rate limit 때문에 실패"라기보다 **Codex 내부 account window
  알림이 stopped/failed thread UI로 새어 나온 것**이다.

### 관련 지연/실패 분석
- Codex run은 간단한 명령도 매번 새 provider process + JSON-RPC handshake
  (`initialize` → `thread/start|resume` → `turn/start`)를 거친다. long-lived CLI reuse가
  아니다.
- `JYM-ai-get-done` 기본 timeout은 hard 30분 / semantic idle 10분이라, 완료 신호가
  누락되거나 non-semantic log만 오면 간단한 작업도 오래 도는 것처럼 보인다.
- `origin/main`은 runtime max concurrency 기본값을 4로 바꿨다. 각 Codex run은 여전히
  별도 app-server turn이라 동시 실행이 Codex account pressure와 실패 빈도를 키울 수 있다.
- SaaS assignment는 `ResumeSessionID`를 채우지 않는다. "다른 작업인데 문맥이 안 바뀜"은
  Codex thread resume 누수보다는, daemon이 실제 repo를 mount/clone하지 않고 빈 isolated
  workdir에서 실행하는 **filesystem context 부재**가 더 유력하다.

### 이번 1차 변경
- SSOT: `docs/20-domain/provider-runtime.md` §5.3에 provider-internal notification은
  user-visible `LogLine`/`EventLog`로 승격하지 않는다는 규칙 추가.
- Review: `docs/review/ai-agent-runtime-lifecycle-review-2026-06-08.md` §22.1에
  rate-limit noise, long-run timeout, failure opacity, context mismatch 원인 추가.
- Code: `internal/provider/codex/translate.go`에서 Codex account rate-limit update와
  structural item/hook/status notification을 drop한다.
- Code: Codex `turn_error` / `turn/failed`의 nested `error.message` / `detail` /
  `error` payload를 terminal failure reason으로 보존한다. 빈 실패 메시지가
  control-plane fallback `agent work failed`로 뭉개지는 경우를 줄인다.
- Test: `internal/provider/codex/translate_test.go`에 internal notification drop 회귀
  테스트와 nested failure reason 회귀 테스트 추가.

### 2차 확인 및 변경 (2026-06-09)
- 참여자에 agent를 등록한 뒤 몇 초 후 체크가 풀리고 `중지 / agent work failed`가
  보이는 현상은 participant mutation rollback이 아니라, assignment가 terminal `failed`
  상태가 되면서 client selector가 terminal thread를 active participant에서 제외하는
  projection 결과다.
- `agent work failed`만 보이는 추가 원인은 control-plane assignment projection이
  `Assignment.State`와 `LastEventSeq`만 보존하고, 마지막 terminal event message를 보존하지
  않았기 때문이다. daemon이 실제 실패 원인을 보냈더라도 thread list/bootstrap reconcile이
  projection state만 보고 `assignmentEventActionResponse(..., message="")`를 호출하면
  fallback copy로 다시 덮일 수 있다.
- Code(control-plane): `AssignmentProjection`에 `LastEvent`를 추가하고, in-memory/DynamoDB
  projection에 `last_event_json`을 저장/로드한다. stale read-model repair는 projection의
  마지막 event message를 사용해 failed/completed/cancelled thread message를 복구한다.
- Code(daemon): supervisor가 runtime eligibility, workspace prepare, runtime submit,
  terminal provider result failure를 daemon.log에 구조화된 한 줄로 남긴다. 필드는
  `phase`, `task_id`, `assignment_id`, `agent_id`, `run_id`, `runtime_id`, `provider`,
  `model`, `workdir`, `status`, `err`, `output`이다.
- Test(control-plane): projection repair가 provider failure message를 보존하는지 검증한다.
  DynamoDB projection load/save도 마지막 event JSON을 검증한다.
- Test(daemon): supervisor/codex 패키지 테스트를 통과시켰다.

### 검증
- `go test ./internal/provider/codex -count=1`
- `go test ./internal/agentbridge/session ./internal/agentbridge/runtimeactor -count=1`
- `go test ./internal/agentbridge/supervisor ./internal/provider/codex -count=1`
- `go test ./...`
- control-plane: `go test ./internal/riidoaiserver -count=1`
- control-plane: `go test ./...`

### 남은 후속
- control-plane terminal read-model fence는 별도 PR/변경으로 유지되어야 한다. daemon에서
  provider noise를 drop해도 late progress/provider log를 terminal thread가 받아들이면
  다른 noise가 같은 방식으로 샐 수 있다.
- 실패 메시지 분류를 제품에 노출해야 한다. `agent work failed`는 마지막 fallback이어야
  하고, 실제 terminal reason(`semantic idle timeout`, `process exited without provider
  result`, provider RPC error, runtime ineligible, no repo/workdir 등)을 보존해야 한다.
  이번 변경은 projection/reconcile 단계의 보존을 고정하지만, 제품 copy와 분류 UI는 아직
  별도 후속이다.
- repo/worktree binding을 추가해야 한다. prompt context가 바뀌어도 process cwd가 빈
  generated workdir이면 coding task 문맥은 실제로 바뀐 것이 아니다.

## 23. 참여자 등록 지연 및 새로고침 후 공백 (2026-06-09)

### 확인된 원인
- 참여자에 agent를 추가하는 최초 표시 자체는 daemon poll, provider CLI spawn, SSE를
  기다릴 필요가 없다. control-plane이 assignment를 만들고 AI Agent task thread read-model을
  갱신하면 client는 바로 표시할 수 있어야 한다.
- control-plane assign handler가 클릭한 task만 reconcile하지 않고
  `reconcileAIAgentTaskThreadProjections(..., "")`를 호출해 workspace의 visible active
  thread 후보 전체를 훑는다. active/queued/stale thread가 많으면 현재 task와 무관한 durable
  assignment projection 조회 때문에 assign 응답이 늦어진다.
- assign handler는 assignment 저장 전에 task context private API를 호출해 prompt를 만든다.
  이 prompt는 runtime에는 필요하지만, 참여자 UI 표시까지 같은 요청에서 막고 있어 체감 지연을
  키운다.
- client `useAiAgentTask.assignAgent`는 mutation 응답의 `AIAgentTaskActionResponse`를
  threads cache에 즉시 반영하지 않고 query invalidate만 한다. 서버가 202 Accepted를 반환해도
  threads refetch가 끝나기 전까지 `currentThread`/`assignedAgent`가 바뀌지 않는다.
- agent 교체 UI는 현재 `stopTask()` -> `unassignAgent()` -> `assignAgent()`를 순차 await한다.
  control-plane `AssignTask`가 기존 assignment replace/cancel을 처리할 수 있으므로, 다른 agent
  선택에는 이 순차 stop/unassign이 불필요한 지연이다.
- 새로고침 후 참여자가 공백이 되는 경우는 두 종류다.
  - 정상 queued/running thread가 있는데 client가 agent profile을 못 찾으면
    `selectedAgentIds=[]`가 되어 공백으로 보인다. 이 경우 client 표시/cache 경로 문제다.
  - thread가 failed/stopped/unassigned terminal 상태라면 selector가 의도적으로 active
    participant에서 제외한다. 이 경우 공백은 UI rollback이 아니라 assignment terminal
    projection 결과이며, root cause는 daemon/runtime/provider failure다.

### 1차 수정 방향
- control-plane assign/create-assignment handler의 reconcile scope를 `""`에서 `taskID`로
  좁힌다. workspace 전체 active thread scan을 클릭 path에서 제거한다.
- client는 assignment mutation 응답을 `tasks/{taskId}/threads` cache에 upsert한다. refetch는
  background correction으로 남기고, 참여자 표시는 REST 재조회 왕복을 기다리지 않는다.
- client agent 교체는 다른 agent를 선택할 때 바로 `assignAgent(newAgentId)`를 호출한다. 선택
  해제일 때만 stop/unassign 경로를 사용한다.
- terminal failed/stopped thread를 active participant로 되살리지는 않는다. 정상 active
  상태의 표시 공백만 막고, terminal failure의 원인 노출은 별도 failure-classification 후속으로
  둔다.
