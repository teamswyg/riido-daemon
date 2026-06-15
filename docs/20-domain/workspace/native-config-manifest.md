# Workspace / Native Config SSOT: Native Config Manifest

[Back to workspace.md](../workspace.md)

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
