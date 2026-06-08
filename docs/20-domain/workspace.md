# Workspace / Native Config SSOT

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

task workdir 에 소스 저장소가 mount 되지 않은 경우(현재는 `MountRepo` 미구현이라
사실상 항상) workspace 는 native config 에 "Working directory" 가이던스를 주입한다.
이 가이던스는 workdir 절대경로를 알리고, 에이전트가 **먼저 이 task 가 코드베이스를
필요로 하는지 판단**하도록 시킨다 — 코딩 작업(코드 읽기/수정/실행/생성)은 필요하고,
비코딩 작업(답변/계획/문서)은 불필요. 필요하면 빈 디렉터리에서 추측·scaffold 하지
말고 멈춰서 사용자에게 이 경로에 프로젝트를 두라고 하거나 새로 만들지 물어보고,
불필요하면 그대로 진행한다. 이 주입은 `workdirHasWorkContent` 가 daemon-injected
config 외 내용물을 감지하면(=repo mount 후) 자동으로 사라진다. 실제 repo
mount(worktree / shallow clone, §4)는 후속 작업이며, 그때 이 가이던스는 비활성된다.

SaaS task source 가 agent instruction 값을 제공하는 경우 C6/C4 는 그 값을
run-scope prompt/native config 입력으로만 소비한다. instruction 의 저장 위치,
길이 제한, 수정 가능성, RBAC, 그리고 `profile_thumbnail_url` / `description`
같은 presentation field 는 public `riido-contracts` 와
`riido-control-plane` 의 SSOT 가 소유한다. Daemon 은 thumbnail 이나
description 값을 native config 에 주입하지 않는다.

### 5.1.1 native config manifest

Provider 별 native config file plan 의 실행 가능한 SSOT 는 `internal/workdir/native_config_plan.riido.json` (`riido-native-config-plan.v1`) 이다. 이 IR 은 `tools/riidogen/templates/native_config_plan.go.gotmpl` 로 `internal/workdir/native_config_plan_gen.go` 를 생성하며, `go generate ./internal/workdir` 로 갱신한다. 코드가 직접 provider→filename/hook/config-home mapping 을 재정의하면 안 된다.

`riido-native-config-manifest.v1` 은 provider-native config 작성의 현재 증적이다. 이 manifest 는 `workdir/.riido/native-config-manifest.json` 과 `native-config/.riido/native-config-manifest.json` 에 같은 내용으로 쓰인다.

필드:

| Field | 의미 |
| --- | --- |
| `schema_version` | 항상 `riido-native-config-manifest.v1` |
| `provider_kind` | provider family (`claude`, `codex`, `openclaw`, `cursor`, unknown fallback 등) |
| `protocol_kind` | C3 가 선택한 protocol kind. 비어 있으면 field 자체를 생략한다 |
| `primary_instruction_file` | provider 가 자동 로드하는 1차 instruction 파일 (`CLAUDE.md`, `AGENTS.md`, `GEMINI.md`) |
| `manifest_file` | manifest 자체의 상대 경로. 현재 `.riido/native-config-manifest.json` |
| `hook_mode` | 현재 hook materialization 방식. `instruction-only` 는 provider-native hook script/settings 를 아직 쓰지 않고 1차 instruction file hard rule 로만 집행한다는 뜻이고, `claude-command-hooks` 는 Claude `.claude/settings.json` 과 command hook script 를 주입했다는 뜻이다 |
| `config_home_dir` | provider 전용 config home 이 task-scoped 로 주입될 때의 상대 경로. 현재 기본 plan 에서는 비어 있다 |
| `provider_settings_files` | provider 가 직접 읽는 settings/config 파일 목록. 현재 Claude `.claude/settings.json` 을 쓸 수 있다 |
| `hook_files` | provider-native hook settings 가 참조하는 script 파일 목록. 현재 Claude command hook script 는 `.riido/hooks/claude-audit-hook.sh` |
| `telemetry_contract_placement` | SaaS source 가 prompt/system prompt 에 telemetry contract 를 둔 위치. 비어 있으면 field 자체를 생략한다 |
| `workflow` | runtime config workflow branch. 비어 있으면 `default` |
| `generated_files` | C6 가 이 run 에 deterministic 하게 쓴 상대 파일 경로 목록 |

`generated_files` 에는 manifest 자신도 포함된다. 따라서 manifest schema, hook mode, telemetry placement, provider filename catalog 변경은 모두 `NativeConfigVersion` 에 반영된다.

현재 `riido-native-config-plan.v1` materialization:

| Provider | 생성 파일 | 의미 |
| --- | --- | --- |
| Claude | `CLAUDE.md`, `.claude/settings.json`, `.riido/hooks/claude-audit-hook.sh`, `.riido/native-config-manifest.json` | Claude Code 의 project settings hook surface 를 task workdir 안에 고정한다. 기본 hook 은 `PreToolUse` / `PostToolUse` 입력 JSON 을 `.riido/hooks/claude-hook-events.jsonl` 로 append 하는 audit-only command hook 이며, exit 0 으로 provider 행동을 차단하지 않는다. 단 `.claude/settings.json` 과 hook script 는 C7 policy bundle 이 `claude:command-hooks:audit` surface 를 허용한 경우에만 materialize 된다. 거절되면 manifest 의 `hook_mode` 은 `instruction-only` 로 기록되고 `CLAUDE.md` 만 남는다. |
| Codex | `AGENTS.md`, `.riido/native-config-manifest.json` | Codex 는 task-scoped `.codex/config.toml` 또는 `CODEX_HOME` overlay 를 materialize 하지 않는다. app-server credential 사용과 full-access runtime envelope 는 C4 Codex adapter 가 `codex --sandbox danger-full-access app-server --listen stdio://` 로 고정한다. Codex process 가 실행 중 workdir `.codex` state 를 만들 수 있지만, C6 manifest/provider settings output 으로 선언하지 않는다. Workdir 은 기본 cwd/evidence root 이며 filesystem sandbox boundary 가 아니다. |
| OpenClaw / Cursor / unknown | `AGENTS.md`, `.riido/native-config-manifest.json` | 현재는 provider-neutral instruction file 주입만 한다. |

### 5.1.2 native config overlay policy

Native config overlay 의 표준은 **user-global config 를 읽거나 복사하지 않는
per-task materialization** 이다. Claude command hook 과 future provider-native
config home 은 모두 C7 policy bundle 의 explicit allow surface 로만 활성화된다.
Codex 는 이 C6 overlay 를 쓰지 않고 C4 adapter 의 mandatory full-access sandbox
selection 으로 provider runtime 을 실행한다. Workdir 은 daemon 이 선택한 cwd 와
결과/evidence root 이지만, Codex full-access mode 에서 provider 가 읽고 쓸 수 있는
유일한 filesystem boundary 는 아니다.

OpenClaw / Cursor / unknown provider 는 이 문서 기준에서 instruction-only
overlay 가 default 이며, provider-native config home 을 자동으로 추론하지 않는다.
새 provider-native overlay surface 를 추가하려면 C7 policy bundle surface,
`riido-native-config-plan.v1`, manifest field, NCV 입력을 같은 PR 에서 갱신한다.

### 5.2 deterministic materialization

같은 (`policy bundle version`, `task plan`) 입력은 항상 같은 파일 셋과 같은 내용을 만든다. 임의 timestamp / hostname / random salt 가 파일 내부에 새지 않게 한다. 이게 `NativeConfigVersion` (§6) 산출의 기반.

### 5.3 protected path / hooks 의 분리

- “어떤 path 가 protected 인가” = C7 결정 (`T-PATH`).
- “protected 를 어떻게 구현하는가” = C6 (예: chattr `+i`, readonly mount, namespace 격리 등).
- “provider hook 으로 protected path edit 을 차단” = adapter 가 받은 hook script 실행 — C4. hook script 의 내용은 정책 번들에서 옴.

## 6. NativeConfigVersion 생성 규칙

`NativeConfigVersion` 은 execution-bound `CanonicalEvent` 의 의무 필드다. 본 문서가 생성 규칙을 소유하고, event schema 자체는 public `riido-contracts` C2 계약이 소유한다.

```
NativeConfigVersion = sha256-hex(
   canonicalJSON({
       policyBundleVersion: <C7 활성 번들 버전>,
       nativeConfigPlan: {
           providerKind:       ...,
           protocolKind:       ...,
           injectedFiles[]:    [{ path, sha256(content) }, ...],
           hookScriptVersions: [{ id, sha256 }, ...],
           wrapperManifestSha: <opt> ,
       },
       schemaVersion: 1
   })
)
```

규칙:

1. 입력에 **모든 주입된 파일의 내용 해시** 가 포함되어야 한다 → 한 줄만 바뀌어도 새 버전.
2. 입력에 `policyBundleVersion` 이 포함되어야 한다 → 정책 번들 변경은 항상 NativeConfigVersion 변경.
3. 주입된 파일에는 `.riido/native-config-manifest.json` 도 포함된다. provider filename catalog, hook materialization mode, telemetry placement 의 변경은 manifest 내용 변경으로 NCV 에 반영된다.
4. 알고리즘 / 입력 schema 자체가 바뀌면 `schemaVersion` 을 올린다. 옛 schemaVersion 의 산출값은 영구 보존 (replay 호환).
5. `NativeConfigVersion` 은 task 시작 시점에 정해진 뒤 그 run 동안 **불변**. 변경하려면 `ReinjectNativeConfig` 또는 새 run (`ReworkQueued → Queued`).
6. local daemon 의 supervisor 는 native config 주입 직후 이 값을 계산해 run metadata `native_config_version` 에 고정한다. `NativeConfigInjected` / `WorkdirArchived` 같은 Cat E 이벤트 append 는 같은 `NativeConfigVersion` 을 EventIngestor 경로로 stamp 한다.

## 7. PolicyBundleVersion ↔ NativeConfigVersion 관계

- `PolicyBundleVersion` 변경 → `NativeConfigVersion` 변경 (§6 입력의 한 멤버이므로).
- `NativeConfigVersion` 변경이 `PolicyBundleVersion` 변경을 함의하지는 않는다 (예: 같은 정책 번들 + 새 wrapper manifest).
- 둘 다 task 시작 시점에 함께 고정 → execution-bound CanonicalEvent 필드로 영속화.
- 진행 중 task 가 두 값 중 하나라도 silent 하게 따라가는 것을 금지. 변경은 runtime upgrade flow 의 T-POLICY / T-CONFIG 분기를 통해서만 가능하며, 해당 architecture SSOT 는 provider-runtime slice 에서 public repo 로 이동한다.

## 8. Workspace lock 정책 (도메인 표현)

본 문서는 lock 의 **사용 정책** 을 도메인 표현으로 갖는다. 실제 `flock` / DB lease primitive 는 [`./locking.md`](./locking.md) (C9).

### 8.1 lock 의 종류

| Lock | scope | 보유 시간 | 사용 시점 |
| --- | --- | --- | --- |
| `repo_cache_update.lock` | `cache/repos/{repo_hash}` 단위 | **짧음** — fetch/prune 동안만 | cache 갱신 시 |
| `task_workdir.lock` | `workspaces/.../runs/{run_id}/` 단위 | run 동안 보유 (로컬 OS 동시성 보호) | adapter / RunController / validation 이 같은 workdir 을 동시 mutate 못하게 |
| `archive_pipeline.lock` | archive 단계 단위 | archive 동안만 | `ArchiveWorkspace` |

### 8.2 금지

- **`repo_lock` 으로 agent run 전체를 감싸지 않는다.** 같은 repo 의 여러 task 가 직렬화되어 처리량이 깨진다.
- **`task_workdir.lock` 으로 다른 task 의 workdir 를 보호하려 하지 않는다.** 격리는 디렉토리 분리 + protected path / sandbox 로 한다.

### 8.3 분산 환경에서

여러 데몬이 같은 호스트 / 같은 cache 를 공유할 때는 `repo_cache_update.lock` 만 의미가 있다. 다른 host 의 데몬과의 cache 공유는 본 SSOT 비범위 — 보통 host 별 독립 cache 를 둔다.

## 9. 인접 SSOT 와의 계약

| 인접 context | 본 문서가 받는 / 공급 |
| --- | --- |
| **C7 Security / Policy** | 받는다: `T-PATH` protected paths, `T-SBX` sandbox 정책, `T-CFG` native config plan, `T-MCP` MCP allowlist (workdir 안 MCP 설정 주입에 사용). 공급: 주입 완료 IR 이벤트 (`NativeConfigInjected`). |
| **C4 Provider Runtime / Adapter** | 공급: workdir 경로, native config 사본, `NativeConfigVersion`. 받지 않음: adapter 가 직접 workdir 을 만들지 않는다. |
| **C5 Runtime Scheduling** | 받는다: claim 된 task 의 `runID`, `RuntimeID`, `CapabilityFingerprint`. 공급: `WorkspacePrepared` 신호 (claim → lease 활성 사전조건). |
| **C2 IR Event Log** | Cat E (workspace/config) 의 1차 producer. `WorkdirCreated` / `NativeConfigInjected` / `WorkdirArchived` / `ConfigTemplateReinjected` 발행은 EventIngestor API 를 통해서만 수행한다. local workdir adapter 는 run root 의 `ir/events.jsonl` sink 를 제공할 수 있지만, envelope 확정 권한은 C2 EventIngestor 에 있다. C2 event schema 는 public `riido-contracts` 가 소유한다. |
| **C8 Validation** | 공급: workdir 의 base / final 상태, diff, artifacts. validation 은 본 디렉토리들을 **읽기 전용** 으로 본다. |
| **C9 Locking** | 받는다: `flock` primitive (§8 의 도메인 lock 들이 실제로 사용하는 메커니즘). |
| **컨테이너 / VM 매니저 (외부)** | tier=`IsolatedContainer`/`EphemeralVM` 인 경우 C6 는 host-side run root 와 manifest 를 준비하고, container/VM 안으로 mount/전달하는 책임은 C4 runtime launcher / platform adapter 가 갖는다. C6 는 mount primitive 나 VM lifecycle 을 소유하지 않는다. |

## 10. Resolved workdir policy decisions

아래 항목은 RIID-4573 에서 본 C6 SSOT 로 흡수된 결정이다. 다시 open
question 으로 복제하지 않는다.

| ID | Decision | Follow-up owner |
| --- | --- | --- |
| `Q-WS-001` | Local daemon archive backend default 는 same-host run root `keep-in-place` 다. S3 / 압축 bundle / 외부 storage 는 default 가 아니며, 별도 archive adapter 와 config/env 가 생기기 전에는 자동 선택하지 않는다. | future infra/archive adapter |
| `Q-WS-002` | Default workdir retention 은 disabled 다. `RIIDO_WORKDIR_RETENTION_SECONDS` 가 명시된 경우에만 archived run TTL cleanup 이 켜지고, size / task-count cleanup 은 default 로 존재하지 않는다. | daemon config + workdir cleanup |
| `Q-WS-003` | Shared repo cache prune 은 자동 주기가 없다. 필요 시 operator-triggered maintenance 로만 실행하고 `repo_cache_update.lock` 을 짧게 잡는다. | future cache maintenance adapter |
| `Q-WS-004` | Native config overlay 는 per-task materialization 이 표준이다. User-global config overlay/copy 는 default 로 금지하고, provider-native config home 은 C7 explicit allow surface + manifest/NCV 반영이 있을 때만 쓴다. | C7 policy + C6 workdir |
| `Q-WS-005` | Container/VM workdir 전달 owner 는 C4 runtime launcher / platform adapter 다. C6 는 host-side run root, materialized files, manifest 만 공급한다. | future isolated runtime launcher |
| `Q-WS-006` | Dirty workdir 에 대한 automatic in-place `ReinjectNativeConfig` threshold 는 zero 다. `Preparing`/`Running` 이후 policy/native-config 변경은 runtime-upgrade flow 를 통해 cancel/fail and next-run 재평가로 처리한다. | runtime upgrade flow + supervisor |

## 11. version-affecting changes

- 새 operation 추가는 `change:additive` (단 IR Cat E 이벤트 동시 갱신).
- directory layout 변경은 `change:breaking-policy` + migration tool 필수.
- `NativeConfigVersion` 알고리즘 변경은 `change:breaking-ir` (옛 schemaVersion 산출값을 영원히 보존 — replay 호환).
- lock 정책 변경은 `change:breaking-policy` (분산 환경의 안전성에 영향).
- protected path 구현 방식 변경(예: chattr → namespace) 은 `change:behavioral` (정책 결정은 C7 가 owner, 구현은 본 문서 자유).
