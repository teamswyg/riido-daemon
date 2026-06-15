# Workspace / Native Config SSOT: Invariants

[Back to workspace.md](../workspace.md)


> **이 문서가 per-task workdir 트리 / native config 주입 / `NativeConfigVersion` 생성 / repo cache 분리 / workspace lifecycle 의 SSOT다.**
>
> - 책임: workdir 디렉토리 구조와 lifecycle, native config 파일(CLAUDE.md / AGENTS.md / hooks settings / wrapper manifest 등)을 task workdir 에 **deterministic 하게 materialize**, `NativeConfigVersion` 부여 규칙, repo cache 와 task workdir 의 격리, lock 사용 정책의 도메인 표현.
> - 비책임: **정책 결정** 은 본 문서가 하지 않는다. C7 Security /
>   Policy 후속 migration slice 가 소유한다. 실제 lock acquisition primitive 는
>   [`./locking.md`](./locking.md) (C9). provider 에 어떤 flag/env 를 넘기는가
>   는 C4 Provider Runtime 후속 migration slice 가 소유한다. lease /
>   scheduling 은 [`./runtime-scheduling.md`](./runtime-scheduling.md) (C5).

이 SSOT 는 **C6 Workspace / Native Config** context 를 채운다.

## 0. 핵심 invariant

> **Workspace 는 policy 를 결정하지 않는다. Workspace 는 C7 Security 가 허용한 policy / native config bundle 을 task workdir 에 deterministic 하게 materialize 한다.**

추가 invariant:

1. **per-task isolated workdir.** 모든 task 의 run 은 독립 workdir 트리(§3)를 갖는다. 동일 task 의 두 run 도 별도 `run_id` 디렉토리로 분리된다.
2. **`Running` 진입 전에 `WorkspacePrepared` 가 invariant — 단 `Claim` 의 사전조건은 아니다.** C5 Runtime Scheduling 은 task 를 claim 할 때 workspace feasibility(예: cache 가용성, 디스크 용량)는 참조할 수 있지만 **`WorkspacePrepared` state 자체를 claim 전에 요구하지 않는다**. `Preparing → Running` 전이의 사전 조건일 뿐이다. workdir 이 prepare 되지 않은 채로 provider process 가 기동되면 안 된다. Task lifecycle 자체는 public `riido-contracts` C1 계약이 소유한다.
3. **`NativeConfigVersion` 없이 `Running` 진입 금지.** native config 가 주입되지 않은 task 는 어떤 정책으로 시작했는지 replay 불가능하므로 시작 자체를 거절한다.
4. **`PolicyBundleVersion` 과 `NativeConfigVersion` 은 `Running` 진입 전 확정된다.** `NativeConfigVersion` 은 execution-bound RunScope 이벤트의 의무 필드이며, TaskScope / SystemScope / RuntimeScope 및 pre-execute RunScope 이벤트(`TaskClaimed`, `WorkdirPreparing`, `WorkdirCreated`, `RuntimePinned`, `RuntimeHandshakeOK` 등)에는 요구되지 않는다. IR event schema 자체는 public `riido-contracts` C2 계약이 소유한다. `PolicyBundleVersion` 은 모든 scope 의 공통 envelope 필드(이벤트 발행 시점의 활성 정책 번들). 같은 run 의 execution-bound 구간 안에서 NCV 가 silent 하게 바뀔 수 없다 — 변경은 `ConfigTemplateReinjected` 또는 새 run.
5. **`protected path` 는 본 문서가 결정하지 않는다.** Security(C7) 가 결정한 path glob 목록을 받아 workspace 가 실제 파일 시스템 권한 / 가시성으로 구현할 뿐이다.
6. **shared repo cache lock 은 짧게.** repo cache 업데이트 / shallow fetch 같은 짧은 구간에만 lock. **agent run 전체** 를 repo lock 으로 감싸는 것은 invariant 위반(같은 repo 의 여러 task 가 직렬화되어 처리량이 떨어진다).

## 1. Workspace lifecycle (state)

도메인 레벨의 lifecycle. 코드 구현은 단순 operation 으로 가도 됨(§2).

| State | 의미 | 진입 조건 |
| --- | --- | --- |
| `WorkspaceUnprepared` | task 가 생성됐지만 workdir 미생성 | `TaskCreated` ~ `TaskQueued` |
| `WorkspacePreparing` | workdir 생성 / repo mount / native config 주입 진행 중 | `TaskClaimed` → `WorkdirPreparing` |
| `WorkspacePrepared` | 모든 준비 완료. `NativeConfigVersion` 확정. provider process 기동 가능 | `RunStarted` 의 사전 조건 |
| `WorkspaceDirty` | run 진행 중 또는 종료 후 산출물 / artifact 가 남아있는 상태 | `RunStarted` 이후, `ArchiveWorkspace` 이전 |
| `WorkspaceArchived` | 산출물 보존 위치가 기록되고 retention 대상이 된 상태. 로컬 기본값은 workdir 을 삭제하지 않고 run root 에 archive manifest 를 남기는 `keep-in-place` 보존이다 | task terminal 후 |
| `WorkspaceFailed` | prepare / archive 실패. 정리 / 분석 대상 | 어느 단계든 회복 불가 실패 시 |

본 lifecycle 은 `TaskState` (C1) 와 1:1 매핑되지 않는다. task 가 `Failed` 로 가더라도 workspace 는 `WorkspaceDirty` → 운영자 분석 → `WorkspaceArchived` 로 따로 진행될 수 있다.

## 2. Workspace operations

각 operation 은 단일 책임 + 결정적 출력. 모두 IR 이벤트로 영속된다.

| Operation | 입력 | 출력 / IR event |
| --- | --- | --- |
| `PrepareWorkspace` | `taskID`, `runID`, repo ref(commit/branch), policy bundle 결정, native config plan | workdir 트리 생성. `WorkdirCreated` |
| `MountRepo` | repo cache 위치, ref, isolation mode (worktree / shallow clone) | workdir 안에 코드 trees 가시화. (repo lock 은 §7 규칙 적용) |
| `InjectNativeConfig` | policy bundle 결정 + native config plan | workdir 안에 CLAUDE.md / AGENTS.md / settings / hooks / wrapper manifest 작성. `NativeConfigInjected(files[], nativeConfigVersion)` |
| `RecordBaseline` | workdir 상태 hash | run 시작 직전 base 상태 영속화 (이후 diff 의 기준) |
| `CollectArtifacts` | run 진행 결과 | output / logs / artifacts 디렉토리에 저장 |
| `ArchiveWorkspace` | run 종료 후 | workdir 트리 archive. 로컬 기본 adapter 는 run root 의 `archive.json` 에 `riido-workdir-archive.v1` manifest 를 쓰고, `archiveURI` 를 같은 host 의 run root 로 기록한다. `WorkdirArchived(workdirPath, archiveURI)` |
| `CleanupWorkspace` | terminal + 보관 정책 만료 | 실제 파일 시스템 제거 |
| `ReinjectNativeConfig` | T-CONFIG trigger | workdir 의 native config 만 갱신. 다른 산출물은 유지. `ConfigTemplateReinjected(from, to)` |

## 3. Directory layout

per-host (또는 per-server) 루트 아래 다음 트리를 만든다. 정확한 workdir root path 는 Factor 12 env `RIIDO_WORKDIR_ROOT` 가 소유한다. `dev-local` / `developer-id` 기본값은 macOS 지원 배포면 `$HOME/Library/Application Support/riido/workspaces` 이다. `mac-app-store` / `msix-store` 같은 store channel 의 app data root 는 C11 Distribution / Host Integration 후속 migration slice 가 소유하고, C6 는 C11 이 허용한 root 만 materialize 한다.

현재 C11 순수 모델은 후속 migration slice 대상이다. C6 는 C11 이 계산한
channel-approved app data root 결과를 받아 materialize 할 뿐, store channel 에서
user home fallback 을 만들지 않는다.

```
$RIIDO_WORKDIR_ROOT/
  {workspace_id}/                 # workspace = task 들의 조직 단위 (org / project / channel 등)
    tasks/
      {task_id}/
        runs/
          {run_id}/
            workdir/              # ★ agent 가 보는 root. agent 의 cwd.
            output/               # 산출물 (PR diff, generated files 등)
            logs/                 # adapter / RunController / validation 측 로그
            artifacts/            # validation / build 결과물
            native-config/        # 주입된 CLAUDE.md / AGENTS.md / hooks / wrapper manifest 원본 사본
            ir/                   # 이 run 의 IR snapshot 캐시 (선택)
            archive.json          # riido-workdir-archive.v1, ArchiveWorkspace 결과 manifest
```

규칙:

1. **agent root = `workdir/` 뿐**. provider process 의 cwd 는 항상 `workdir/`. agent 가 `../output/` 같은 상대 경로로 형제 디렉토리에 접근하지 못하도록 sandbox 가 막는다. 실제 sandbox 정책은 C7 의 T-SBX / T-PATH 가 소유한다.
2. `output/`, `logs/`, `artifacts/`, `native-config/`, `ir/` 의 **protected 여부** 는 본 문서가 결정하지 않는다 — C7 의 `T-PATH` 정책을 받아 workspace 가 mount 옵션 / 권한 / 별도 OS 사용자 / namespace 로 구현한다.
3. terminal 후 로컬 기본 archive 는 `keep-in-place` 다. 즉 `archive.json` 에 `archiveURI=file://<run-root>` 를 기록하고 실제 디렉토리는 삭제하지 않는다. S3 / 다른 storage / 압축 bundle 같은 외부 archive backend 는 운영 정책이며, 위치 표현은 `WorkdirArchived.archiveURI`.
4. local daemon stop 으로 in-flight run 이 cancel 되는 경우도 terminal workspace lifecycle 로 취급한다. RunController 는 `TaskCancelled` 와 같은 terminal transition 을 기록한 뒤 같은 `ArchiveWorkspace` 경로로 `archive.json` 을 남긴다. 실제 삭제는 `CleanupWorkspace` / retention 정책 만료 후에만 수행한다.
5. local daemon 의 filesystem cleanup 은 기본 비활성이다. 운영자가 Factor 12 env `RIIDO_WORKDIR_RETENTION_SECONDS` 를 명시하면 daemon 은 `archive.json.archived_at` 이 cutoff 보다 오래된 `keep-in-place` run root 만 삭제한다. `archive.json` 이 없는 run 은 active 또는 dirty 로 간주해 삭제하지 않는다.
6. local daemon 은 기본 size / task-count based cleanup 을 갖지 않는다. `RIIDO_WORKDIR_RETENTION_SECONDS=0` 이 default 이며, disk quota 나 N-task pruning 은 별도 operator policy / adapter 가 생기기 전까지 자동으로 추론하지 않는다.

### 3.1 Store channel workspace grants

Store channel 에서는 workdir root 와 user workspace root 를 분리한다.

| Root | Owner | Store rule |
| --- | --- | --- |
| `workdir root` | C11 app data root + C6 materialization | app container / package local data / app group 안에서 생성 |
| `user workspace root` | C11 WorkspaceGrantStore | 사용자가 선택한 folder grant 없이는 접근 금지 |
| `repo cache` | C6, path selection constrained by C11 | store channel 에서는 arbitrary home scan 금지 |

규칙:

1. macOS App Store target 은 security-scoped bookmark 또는 app group/container grant 로 표현되는 workspace grant 를 요구한다. C6 는 bookmark 자체를 해석하지 않고 C11 이 제공한 allowed root 만 받는다.
2. Windows MSIX target 은 package identity 와 사용자가 선택한 folder grant 를 요구한다. C6 는 Windows picker / capability 구현을 알지 않는다.
3. `dev-local` channel 의 현재 `~/Library/Application Support/riido` 경로는 store artifact 에 들어갈 수 없다. Store packaging 은 channel-specific app data root 를 명시해야 한다.
4. Provider process cwd 는 여전히 `workdir/` 뿐이다. 사용자가 선택한 workspace 는 prepare 단계에서 snapshot / worktree / shallow clone 으로 materialize 된다.
5. C11 WorkspaceGrantStore 는 후속 migration slice 대상이다. C6 는 active grant 와 `ConsentLedger` 의 `workspace-access:<workspace-id>` view 가 모두 true 일 때만 user workspace root 를 materialize 한다.

### 3.2 Local archive manifest

로컬 filesystem adapter 가 쓰는 archive manifest 의 schema version 은 `riido-workdir-archive.v1` 이다. 이 manifest 는 `ArchiveWorkspace` 가 terminal result 를 관찰한 뒤 run root 에 원자적으로 기록한다.

필드:

| Field | 의미 |
| --- | --- |
| `schema_version` | 항상 `riido-workdir-archive.v1` |
| `workdir_path` | provider process 의 cwd 였던 `workdir/` 절대 경로 |
| `archive_uri` | 로컬 기본값은 `file://<run-root>` |
| `retention_mode` | 로컬 기본값은 `keep-in-place` |
| `result_status` | terminal result status (`completed` / `failed` / `cancelled` / `timeout` / `aborted` / `blocked`) |
| `archived_at` | archive manifest 기록 시각. 이 timestamp 는 archive operation 의 산출물이며 `NativeConfigVersion` 입력에는 포함되지 않는다 |

## 4. Repo cache 와 task workdir 분리

대전제: **task 가 코드를 바꾸는 곳(workdir)과 repo 원본(cache)은 다른 디렉토리.**

### 4.1 두 가지 isolation mode

| Mode | 동작 | 비용 | 권장 |
| --- | --- | --- | --- |
| **git worktree** | `cache/repos/{repo_hash}` 의 shared bare repo 에 `git worktree add` 로 task workdir 생성 | 디스크 절약 (object 공유), prepare 빠름 | 같은 repo 에 빈번한 task 가 들어오는 경우 |
| **shallow clone** | `git clone --depth 1 <cache> <workdir>` 식으로 task 마다 독립 복제 | 디스크 더 사용, prepare 조금 느림. 완전 격리 | 격리가 더 중요한 경우(예: 한 task 가 다른 task 의 untracked 파일을 보면 안 됨) |

선택은 task 정의(`task.isolationMode`)와 정책 번들 (`T-SBX` 가 강한 격리 요구) 의 결합으로 결정. 본 문서는 두 mode 의 **의미** 만 소유, 선택 결정은 task 정의 + C7.

### 4.2 cache 갱신 정책

- `cache/repos/{repo_hash}` 는 multi-task 공유. fetch / prune 갱신 시에만 짧은 lock 필요(§7).
- task workdir 은 cache 의 시점 스냅샷 위에서 동작. cache 가 갱신되더라도 진행 중 task 의 worktree 는 영향을 받지 않는다(`git worktree` 의 ref 또는 `clone` 의 분리).
- shared repo cache 의 자동 prune 은 local daemon default 가 아니다. prune 이 필요하면 operator-triggered maintenance 로만 수행하며, 반드시 `repo_cache_update.lock` 안에서 진행하고 active task workdir / run root 를 삭제 대상으로 삼지 않는다.

## 5. Native config injection (C6 owns the injection mechanism)

### 5.1 무엇이 주입되는가

`InjectNativeConfig` 가 `workdir/` 또는 `workdir/.claude/` / `workdir/.riido/` 같은 표준 위치에 다음 파일을 생성한다(provider 별 차이는 어댑터별 가이드가 있겠지만, **결정** 은 정책 번들이, **생성 동작** 은 workspace 가).

- `CLAUDE.md` — Claude Code 가 자동 로드. provider research SSOT 는 provider-runtime slice 에서 public repo 로 이동한다.
- `AGENTS.md` — Codex 가 자동 로드. provider research SSOT 는 provider-runtime slice 에서 public repo 로 이동한다.
- Claude `.claude/settings.json` (managed/local/project precedence — 우리는 보통 `project` 위치)
- hooks settings (`.claude/hooks/...` 또는 동등)
- wrapper manifest (있을 때)
- `.riido/native-config-manifest.json` — C6 가 실제 생성한 provider-native config 파일 목록과 hook materialization mode 를 기록하는 `riido-native-config-manifest.v1`
- `.riido/` 메타: `task.json`, `policy-bundle.lock`, `native-config.lock` 등

SaaS task source 가 `riido_telemetry_contract` metadata 를 제공하면 supervisor 는 provider 별 prompt placement 와 별개로 native config hard rule 에 `<riido_log>{"code":...,"args":{...}}<end>` progress telemetry contract 를 주입한다. Progress code catalog 와 append-only 정책은 `riido-contracts/progressmessage/catalog.dsl.riido.json` 이 소유하고, workspace/native config 는 해당 rule 을 provider 가 읽는 위치에 투영만 한다. 이 rule 도 `NativeConfigVersion` 입력 파일 해시에 포함되므로, telemetry contract 변경은 run replay 에서 식별 가능해야 한다.

SaaS task source 가 agent instruction 값을 제공하는 경우 C6/C4 는 그 값을
run-scope prompt/native config 입력으로만 소비한다. instruction 의 저장 위치,
길이 제한, 수정 가능성, RBAC, 그리고 `profile_thumbnail_url` / `description`
같은 presentation field 는 public `riido-contracts` 와
`riido-control-plane` 의 SSOT 가 소유한다. Daemon 은 thumbnail 이나
description 값을 native config 에 주입하지 않는다.
