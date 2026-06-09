# RIID-4925 Branch Review - 2026-06-09

## Scope

기준:

- `riido-client`: `origin/main...HEAD`
- `riido-control-plane`: `origin/main...HEAD`
- `riido-daemon`: `origin/main...HEAD`
- `riido-desktop`: `origin/main...HEAD`
- `riido-contracts`: `HEAD` (`v0.3.5`)

작업트리는 리뷰 시점에 clean 상태였다. 최신 커밋 단일 diff도 확인했지만, daemon/client/desktop의 실제 변경은 브랜치 전체 diff에 더 많이 포함되어 있으므로 findings는 브랜치 전체 기준으로 작성한다.

## Findings

### 1. High - desktop PID fallback kill이 stale PID를 검증하지 않음

파일:

- `/Users/work/work/riido-official/riido-desktop/src/modules/daemonLauncher.ts:621`

`forceKillDaemonByPidFile`은 pid file의 숫자만 읽고 `SIGTERM` 뒤 `SIGKILL`을 보낸다. pid file이 stale이고 OS가 같은 PID를 다른 프로세스에 재사용하면 Riido와 무관한 프로세스를 종료할 수 있다.

필요한 보강:

- pid file에 daemon identity를 같이 저장하거나,
- kill 전에 command line / executable path / process start time / status socket 응답을 검증하거나,
- cooperative stop이 실패한 경우에도 pid-file-only kill은 같은 daemon identity가 확인될 때만 수행해야 한다.

### 2. High - Codex persistent runner가 실제로는 assignment마다 재시작될 가능성이 큼

파일:

- `/Users/work/work/riido-official/riido-daemon/cmd/riido/codex_persistent_runner.go:96`
- `/Users/work/work/riido-official/riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:831`
- `/Users/work/work/riido-official/riido-daemon/internal/agentbridge/supervisor/supervisor.go:773`

`codexPersistentRunner`는 `Executable + Dir + Args + Env`가 완전히 같을 때만 기존 process를 재사용한다. 그런데 SaaS request는 `run_id=assignment.ID`를 metadata에 넣고, supervisor는 `Workdir.Prepare(... Run: runID)`로 assignment별 workdir을 만든다.

결과적으로 새 assignment/comment마다 cwd가 달라져 persistent runner가 기존 Codex app-server를 kill하고 새로 띄울 수 있다. 사용자가 원한 "Codex CLI/app-server 1개 유지" 목표를 달성하지 못할 가능성이 높다.

필요한 보강:

- Codex app-server process cwd를 runtime-level stable directory로 고정하고, turn별 작업 경로는 protocol/system prompt/tool context로 전달하거나,
- 같은 task/thread resume의 workdir identity를 안정화해야 한다.

### 3. High - supervisor in-flight key가 task id라 multi-agent/additive assignment와 충돌함

파일:

- `/Users/work/work/riido-official/riido-daemon/internal/agentbridge/supervisor/supervisor.go:526`
- `/Users/work/work/riido-official/riido-daemon/internal/agentbridge/supervisor/supervisor.go:589`
- `/Users/work/work/riido-official/riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:839`
- `/Users/work/work/riido-official/riido-control-plane/internal/riidoaiserver/server.go:787`

daemon supervisor의 `inFlight` key는 `req.ID`이고, SaaS `TaskRequest.ID`는 `assignment.ID`가 아니라 `assignment.TaskID`다. control-plane에는 `AssignTaskAdditive` 경로가 있으므로 같은 task에 여러 agent assignment가 동시에 존재할 수 있다.

이 경우 두 번째 assignment는 같은 task id duplicate로 판단되어 session이 시작되지 않고, terminal `CompleteTask`도 보고되지 않을 수 있다. lease/queue가 꼬이거나 assignment가 active 상태로 남을 수 있다.

필요한 보강:

- runtime lifecycle key는 최소 `assignment_id` 기준이어야 한다.
- cancellation watcher / reporter / workspace run id / in-flight map도 같은 lifecycle id로 맞춰야 한다.

### 4. Medium-High - assign 전 global reconcile 제거가 stale busy/queued를 남길 수 있음

파일:

- `/Users/work/work/riido-official/riido-control-plane/internal/riidoaiserver/server.go:736`
- `/Users/work/work/riido-official/riido-control-plane/internal/riidoaiserver/ai_agent_client_development.go:1644`

assign handler의 reconcile scope를 `""`에서 `taskID`로 줄인 것은 클릭 path 지연 개선에는 맞다. 하지만 assignability는 `visibleAgents()`가 모든 task thread를 훑어 agent status를 계산한다.

다른 task에 stale active thread가 남아 있으면 현재 task만 reconcile해도 agent가 계속 busy/queued로 보일 수 있고, 새 assignment가 불필요하게 queued될 수 있다.

필요한 보강:

- assign 전에는 workspace 전체가 아니라 selected agent 관련 active thread만 reconcile하거나,
- agent status projection을 durable assignment state 기준으로 계산해야 한다.

### 5. Medium - cancellation watcher goroutine/channel 누수가 남아 있음

파일:

- `/Users/work/work/riido-official/riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:483`
- `/Users/work/work/riido-official/riido-daemon/internal/agentbridge/controlplane/saasplane/saasplane.go:547`
- `/Users/work/work/riido-official/riido-daemon/internal/agentbridge/supervisor/supervisor.go:1339`

`WatchCancellation`은 channel을 `cancelWatchers`에 저장한다. terminal cleanup은 `delete`만 하고 channel을 close하지 않는다. `forwardCancellation`은 actor ctx 또는 stoppedCh가 닫히기 전까지 해당 channel receive에서 기다린다.

완료된 task마다 goroutine/channel이 actor 종료까지 남을 수 있다.

필요한 보강:

- `closeCancelWatcherLocked(taskID)` helper를 만들고,
- terminal cleanup, watcher replacement, cancel delivery, heartbeat stale cleanup에서 모두 close+delete를 수행해야 한다.

### 6. Medium - client cache patch가 신규 active_stream을 만들지 않음

파일:

- `/Users/work/work/riido-official/riido-client/src/lib/hooks/queries/aiAgent/useAiAgentTask.ts:350`

assignment mutation response를 threads cache에 upsert해 참여자 표시는 빨라졌다. 하지만 `active_stream`은 기존 `active_stream.thread_id`가 같은 thread일 때만 보존한다. 신규 assign 직후에는 기존 active_stream이 없으므로 SSE 연결은 여전히 refetch 성공에 의존한다.

증상:

- 참여자는 즉시 보일 수 있다.
- 그러나 refetch가 늦거나 실패하면 "등록됐는데 진행이 안 보임"이 남을 수 있다.

필요한 보강:

- action response가 active 상태라면 server가 stream link를 action response에 포함하거나,
- client가 thread-stream-subscription endpoint를 호출해 active_stream을 보강해야 한다.

### 7. Medium - desktop auto-update quit flow와 daemon teardown이 충돌할 수 있음

파일:

- `/Users/work/work/riido-official/riido-desktop/src/main.ts:734`
- `/Users/work/work/riido-official/riido-desktop/src/modules/updater.ts:110`

`will-quit`에서 항상 `preventDefault()` 후 async daemon teardown을 수행하고 마지막에 `app.exit(0)`을 호출한다. updater는 `autoUpdater.quitAndInstall(false, true)`로 Electron quit lifecycle을 사용한다.

업데이트 설치/재시작 흐름에서 daemon teardown handler가 updater의 quit/install sequencing을 가로챌 가능성이 있다.

필요한 보강:

- update install path에서는 daemon teardown 후 `quitAndInstall`을 호출하도록 sequencing을 명확히 하거나,
- `will-quit` handler가 updater-initiated quit을 구분해 `app.exit(0)`으로 덮지 않도록 해야 한다.

### 8. Low-Medium - pnpm repo에 npm package-lock.json이 커밋됨

파일:

- `/Users/work/work/riido-official/riido-client/package.json:7`
- `/Users/work/work/riido-official/riido-client/package-lock.json:1`
- `/Users/work/work/riido-official/riido-desktop/package.json:211`
- `/Users/work/work/riido-official/riido-desktop/package-lock.json:1`

client와 desktop 모두 `packageManager`가 pnpm인데 npm `package-lock.json`이 새로 들어왔다. CI나 개발자가 npm을 실행하면 pnpm lock과 다른 dependency graph가 생길 수 있다.

필요한 보강:

- npm lockfile이 의도된 것이 아니라면 제거해야 한다.
- 의도된 것이라면 package manager policy와 CI install command를 문서화해야 한다.

## Notes

- `riido-contracts` v0.3.5의 session continuity 필드는 additive로 보이며, JSON shape test도 갱신되어 있다.
- control-plane의 provider session persistence/resume stamp 방향은 맞지만, resume은 provider/model/agent context만으로 충분하지 않을 수 있다. workdir/repo/runtime identity까지 eligibility에 포함하는 것이 더 안전하다.
- terminal failed/stopped thread를 active participant로 되살리지 않는 client selector 정책은 맞다. 다만 failure reason 노출은 별도 후속으로 남아 있다.

## Recommended Order

1. desktop PID kill identity check
2. Codex persistent process cwd/reuse 설계 재검토
3. daemon supervisor lifecycle key를 assignment id 기준으로 전환
4. cancellation watcher close lifecycle 정리
5. assign path에서 selected-agent scoped reconcile 추가
6. client active_stream 보강
7. updater quit sequencing 검증
8. package-lock 정리
