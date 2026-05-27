# Distribution / Host Integration SSOT

> **이 문서가 store channel policy / host integration / external CLI provenance / local IPC / app data root / consent ledger 의 SSOT다.**
>
> - 책임: Riido daemon 이 App Store / Microsoft Store / Developer ID / MSIX sideload 같은 distribution channel 에서 어떤 host surface 를 쓸 수 있는가, provider CLI 를 어떻게 외부 도구로 등록하는가, background/helper 실행과 workspace 접근 동의를 어떻게 기록하는가.
> - 비책임: provider capability 모델은 public `riido-contracts/provider/capability` (C3), workspace materialization 은 [`./workspace.md`](./workspace.md) (C6) 이 소유한다. provider process 실행 의미(C4), security decision matrix(C7), SaaS assignment / polling(C10)은 후속 migration slice 가 각 SSOT 를 이동한다.

이 SSOT 는 **C11 Distribution / Host Integration** context 를 채운다. Context map SSOT 는 후속 architecture-doc migration slice 에서 public repo 로 이동한다.

## 0. 핵심 invariant

1. **Provider CLI 는 번들하지 않는다.** Claude / Codex / OpenClaw / Cursor Agent executable 은 Riido package artifact 안에 들어갈 수 없다. Riido 는 사용자가 설치한 외부 CLI 를 detect / register / verify 할 뿐이다.
2. **Store app 은 control surface 다.** Riido Store App 은 local helper 상태, provider 연결 상태, workspace grant, privacy/telemetry 설정, review/demo mode 를 보여주는 사용자-facing control surface 다. provider runtime 자체를 앱 안에 숨기지 않는다.
3. **Background 실행은 사용자 동의가 truth source 다.** helper/login item/startup task/background sync 는 `ConsentLedger` 의 explicit grant 가 없으면 켜지지 않는다.
4. **Local IPC 는 OS별 adapter 뒤에 둔다.** 도메인은 "local-only IPC" 만 안다. macOS Unix domain socket / app group container path, Windows named pipe / package local data path 는 C11 adapter 가 결정한다.
5. **Store channel 은 runtime capability 의 사용 가부를 제한한다.** provider 가 어떤 surface 를 지원해도 `mac-app-store` / `msix-store` policy 가 금지하면 C3 compatibility 또는 C4 pre-execute 단계에서 blocked 로 본다.
6. **Review 환경은 provider CLI 없이도 평가 가능해야 한다.** Store 심사용 demo/offline mode 는 provider 실행 없이 app shell, consent, provider connection status, workspace grant UX 를 검증할 수 있어야 한다.

## 1. Distribution channel enum

| Channel | 의미 | 1차 목적 |
| --- | --- | --- |
| `developer-id` | macOS Developer ID signed + notarized outside Mac App Store | macOS 1차 배포 |
| `mac-app-store` | Mac App Store sandboxed app | 장기 목표 / 제한 모드 |
| `msix-sideload` | signed MSIX outside Microsoft Store | Windows 1차 배포 |
| `msix-store` | Microsoft Store MSIX / packaged desktop app | Windows Store 배포 |
| `dev-local` | repo / launchd / foreground development | 개발자 로컬 |

Channel 은 build artifact identity 다. 같은 daemon binary 라도 channel 이 다르면 허용 surface 가 달라진다.

## 2. Role model

| Role | 책임 | 포함하면 안 되는 것 |
| --- | --- | --- |
| **Store App** | onboarding, provider path 등록, workspace 선택, background/privacy 설정, status/review UI | provider CLI binary, silent installer, unmanaged updater |
| **Local Helper / Broker** | local-only IPC, task claim loop, provider runtime orchestration, status/metrics | 외부 TCP listener, 사용자 동의 없는 autostart |
| **Provider Connector** | user-selected/env/auto-detected executable provenance, version/login/capability probe | vendor code redistribution, provider license/TOS 대체 |
| **SaaS Control Plane** | assignment, polling, event stream, policy-aware routing metadata | 고객 PC provider process 실행 |

## 3. ExternalToolRegistry

Provider CLI 등록은 executable path 자체보다 **provenance** 가 중요하다.

현재 순수 도메인 모델은 `internal/hostintegration.ExternalToolRecord` /
`ExternalToolRegistry` 가 실행한다. 이 패키지는 PATH 탐색, provider process
spawn, OS bookmark / named pipe 같은 adapter 일을 하지 않는다. adapter 는 검증된
record 를 이 모델로 넘긴다.

```
ExternalToolRecord {
    provider              ProviderKind
    executablePath         string
    provenance             "user-selected" | "env-override" | "auto-detected"
    detectedVersion        string
    loginStatus            "unknown" | "logged-in" | "login-required"
    compatibilityStatus    CompatibilityStatus
    lastVerifiedAt         time
}
```

규칙:

1. `user-selected` 가 가장 강한 신호다. Store App 에서 사용자가 file picker 로 지정한 path 다.
2. `env-override` 는 `RIIDO_<PROVIDER>_PATH` 로 들어온 값이다. Store channel 에서는 UI 에 "환경 변수 override" 로 표시해야 한다.
3. `auto-detected` 는 PATH / known install path 탐지 결과다. Store channel 에서는 실행 전 user confirmation 이 필요하다.
4. `login-required` 는 failure 가 아니다. scheduler / UI 가 해당 provider 를 task routing 후보에서 제외할 수 있게 하는 상태다.

### 3.1 AppDataRoot

OS별 app data root 는 `internal/hostintegration.AppDataRoot` 가 실행하는 C11
순수 모델이다. 이 모델은 OS API 를 호출하지 않고, channel / host OS / adapter 가
제공한 root 후보가 store-safe 한지만 검증한다.

```
AppDataRoot {
    channel DistributionChannel
    hostOS  "darwin" | "windows"
    scope   "user-application-support" | "sandbox-container" | "app-group" |
            "windows-local-app-data" | "windows-package-local-data"
    path    string
}
```

기본 규칙:

1. `dev-local` / `developer-id` + macOS 는 `$HOME/Library/Application Support/riido` 를 app data root 로 둔다.
2. `mac-app-store` 는 app group root 또는 sandbox container root 를 adapter 가 넘겨야 한다. 사용자 home fallback 은 금지다.
3. `msix-sideload` / `msix-store` 는 Windows package local data root 를 adapter 가 넘겨야 한다. `%USERPROFILE%` home fallback 은 금지다.
4. C6 workdir root 는 `AppDataRoot.WorkdirRoot()` 의 결과만 materialize 한다. 즉 app data root 아래 `workspaces/` 다.
5. C11 app data root 는 user workspace root 가 아니다. 사용자가 선택한 repository / workspace folder 는 별도 `WorkspaceGrantStore` 가 허용한 root 로만 들어온다.

### 3.2 WorkspaceGrantStore

사용자 workspace root 접근 grant 는 `internal/hostintegration.WorkspaceGrantStore`
가 실행하는 C11 순수 모델이다. C6 는 이 store 의 active grant record 만 받아
workdir prepare 단계에서 snapshot / worktree / shallow clone 으로 materialize 한다.
OS별 bookmark / picker token bytes 자체는 adapter 소유이며, 도메인은 grant method
와 subject 만 검증한다.

```
WorkspaceGrantRecord {
    workspaceID string
    channel     DistributionChannel
    hostOS      "darwin" | "windows"
    method      "dev-local-path" | "user-selected-folder" |
                "security-scoped-bookmark" | "windows-folder-picker-grant"
    rootPath    string
    grantedAt   time
    revokedAt   time?
}
```

규칙:

1. `workdir root` 와 `user workspace root` 는 항상 분리한다.
2. `mac-app-store` 는 `security-scoped-bookmark` method 없이는 grant 를 받지 않는다.
3. `msix-store` 는 `windows-folder-picker-grant` method 없이는 grant 를 받지 않는다.
4. grant record 가 active 여도 `ConsentLedger` 의 `workspace-access:<workspace-id>` 가 없으면 C6 materialization 은 blocked 다.
5. revoke 는 record 를 삭제하지 않고 현재 active grant view 에서만 제외한다.

## 4. Local IPC

도메인 포트는 다음 semantic 만 노출한다.

```
LocalIPCEndpoint {
    channel      DistributionChannel
    hostOS       "darwin" | "windows"
    endpointKind "unix-socket" | "named-pipe"
    path         string
    owner        "store-app" | "helper"
}
```

현재 순수 모델은 `internal/hostintegration.LocalIPCEndpoint` /
`DefaultLocalIPCEndpoint` 가 실행한다. 이 모델은 endpoint descriptor 만 계산한다.
실제 listener adapter 는 public `internal/riidoapi` 의 local API transport 가
소유하며, dev-local / Developer ID 는 Unix socket, Windows Store/MSIX path 는
Windows named pipe 를 사용한다. RIID-4684 에서 이 adapter 는 private
`project/mwsd` 의존성 없이 public `internal/taskdb` guarded mutation 을 사용하도록
이동됐다.

Channel 별 adapter:

| Channel | IPC adapter | 경로 owner |
| --- | --- | --- |
| `developer-id` | Unix domain socket | `~/Library/Application Support/riido` 또는 user config dir |
| `mac-app-store` | Unix domain socket inside app group/container | App Sandbox container / app group |
| `msix-sideload` | Windows named pipe | package local data / app identity |
| `msix-store` | Windows named pipe | package local data / app identity |
| `dev-local` | Unix domain socket | 현재 launchd/dev path |

`cmd/riido` 의 Unix socket API 는 `dev-local` / `developer-id` adapter 로 본다.
Windows named pipe API 는 `--transport windows-named-pipe` 로 선택하는 C11 local
transport adapter 이며, 같은 request envelope 와 `riidoapi` handler 를 재사용한다.
C1~C10 은 OS별 listener 를 import 하지 않는다.

현재 `cmd/riido daemon` 의 dev-local 기본 socket path 는 C11
`AppDataRoot` + `LocalIPCEndpoint` 를 통해 기존
`$HOME/Library/Application Support/riido/agentd.sock` 로 계산한다. HTTP listener
추가는 여전히 금지다.

### 4.1 macOS App Store helper / login item strategy

`mac-app-store` 는 sandboxed Store App 과 bundle 내부 helper/login item 의 역할을 분리한다. Store App 은 onboarding, provider status, workspace grant, background/privacy setting, review/demo mode 를 보여 주고, helper 는 local-only IPC 와 task execution orchestration 을 담당한다.

규칙:

1. helper role 은 `sandboxed-login-item-helper` 이며, background 등록 방식은 `SMAppService` / Login Item 계열만 허용한다.
2. helper background 실행은 `background-helper` consent 와 App Store review approval 이 모두 있어야 allowed 다.
3. direct `~/Library/LaunchAgents` 설치, self-updater, shared-location code install, standalone code download, provider CLI bundling 은 금지다.
4. app data root 는 app group 또는 sandbox container 여야 하며, user home `Application Support` fallback 은 금지다.
5. local IPC 는 helper-owned Unix domain socket 이고 app group/container root 내부에 있어야 한다.
6. user workspace 접근은 `WorkspaceGrantStore` 의 `security-scoped-bookmark` 와 `ConsentLedger` 의 `workspace-access:<workspace-id>` 가 모두 active 일 때만 C6 materialization 으로 전달한다.
7. App Store review note 는 helper purpose, Login Item consent UX, App Sandbox entitlement 사용 이유, security-scoped workspace grant, provider CLI non-bundling, review/demo mode, privacy metadata allowlist 를 설명해야 한다.

현재 순수 runtime role 모델은 `internal/hostintegration.ResolveHelperRuntimePlan`
이다. 이 함수는 channel-approved `AppDataRoot` 와 `LocalIPCEndpoint` 를 받아
macOS Store App/helper adapter 가 구현해야 할 role, startup registration,
background rule, workspace grant requirement, update rule, review note surfaces 를
계산한다. 실제 Store App bundle, entitlements, `SMAppService` 호출, security-scoped
bookmark bytes 는 C11 adapter / packaging target 소유이며 이 함수는 OS API 를
호출하지 않는다.

macOS Store helper plan invariant:

1. `mac-app-store` role 은 `sandboxed-login-item-helper` 이고 startup registration 은 `service-management-login-item` 이다.
2. `mac-app-store` local IPC 는 helper-owned Unix socket 이며 app group 또는 sandbox container root 아래에 있다.
3. `mac-app-store` background 실행은 `background-helper` consent 와 Store review approval 이 모두 있어야 allowed 다.
4. `mac-app-store` 는 App Store-managed updates 를 사용하며 self-updater / direct LaunchAgent / shared-location install / standalone code download / provider CLI bundling 을 허용하지 않는다.
5. `mac-app-store` workspace grant requirement 는 `security-scoped-bookmark` 다.

### 4.2 Windows MSIX runtime / packaging strategy

`msix-sideload` 와 `msix-store` 는 Windows 에서 같은 local runtime 책임을 갖되 심사/업데이트 경계가 다르다.

| Channel | Runtime shape | Packaging/data rule | Startup/background rule |
| --- | --- | --- | --- |
| `msix-sideload` | signed MSIX 안의 Store App + local helper/broker. helper 는 package identity 아래에서 provider runtime orchestration 을 수행한다. | package identity, Windows Desktop target device family, package local data root, Windows named pipe local IPC 가 필수다. | explicit `background-helper` consent 가 있어야 하며 Windows service install 은 기본 금지다. |
| `msix-store` | Microsoft Store packaged desktop app + packaged full-trust helper/tray process. helper 는 Store submission note 에 설명된 local-only runtime broker 다. | package identity, Windows Desktop target device family, package local data root, named pipe local IPC, Partner Center / `runFullTrust` review note 가 필수다. | Store-managed update 를 사용하고 self-updater / Windows service default install 은 금지다. background helper 는 consent 와 Store review approval 이 모두 필요하다. |

공통 규칙:

1. Claude / Codex / OpenClaw / Cursor Agent CLI 는 MSIX package 에 포함하지 않는다. C11 `ExternalToolRegistry` 는 user-selected / env-override / auto-detected provenance 만 보존한다.
2. Store App 과 helper 사이의 control surface 는 Windows named pipe 로만 열린다. external TCP listener 는 금지다.
3. App data 와 daemon state 는 package local data root 아래에 둔다. `%USERPROFILE%` 또는 임의 home scan fallback 은 금지다.
4. 사용자 workspace 접근은 `WorkspaceGrantStore` 의 `windows-folder-picker-grant` 와 `ConsentLedger` 의 `workspace-access:<workspace-id>` 가 모두 active 일 때만 C6 materialization 으로 전달한다.
5. `msix-store` 는 provider 없이도 onboarding, provider status, workspace grant, background consent, privacy setting 을 볼 수 있는 review/demo mode 를 제공해야 한다.

현재 순수 runtime role 모델은 `internal/hostintegration.ResolveHelperRuntimePlan`
이다. 이 함수는 channel-approved `AppDataRoot` 와 `LocalIPCEndpoint` 를 받아
Store App / helper adapter 가 구현해야 할 role, background rule, update rule 을
계산한다. Windows adapter 는 이 plan 을 실제 MSIX manifest / packaged full-trust
process / tray startup task 구현의 입력으로만 써야 하며, 이 함수 자체는 provider
process 를 spawn 하거나 named pipe 를 열거나 Windows service 를 설치하지 않는다.

MSIX helper plan invariant:

1. `msix-store` role 은 `msix-packaged-full-trust-helper-tray` 이고 local IPC 는 helper-owned Windows named pipe 다.
2. `msix-store` background 실행은 `background-helper` consent 와 Store review approval 이 모두 있어야 allowed 다.
3. `msix-store` 는 Store-managed updates 를 사용하며 self-updater / Windows service default install / provider CLI bundling 을 허용하지 않는다.
4. `msix-sideload` role 은 `msix-packaged-helper-broker` 이고 background 실행은 explicit consent 만 요구한다.
5. 두 MSIX channel 모두 app data root scope 는 `windows-package-local-data` 여야 한다.

## 5. ConsentLedger

다음 동의는 독립 record 로 저장한다.

현재 순수 모델은 `internal/hostintegration.ConsentLedger` /
`ConsentRecord` 가 실행한다. ledger 는 append-only facts 를 보존하고, 현재 허용
상태는 reducer view(`ConsentState`) 로 계산한다.

| Consent | 의미 | 없을 때 |
| --- | --- | --- |
| `background-helper` | app 종료 후 helper/login item/startup task 실행 | helper 자동 시작 금지 |
| `provider-execute:<provider>` | 해당 provider CLI 실행 허용 | detect 는 가능, execute 는 blocked |
| `workspace-access:<workspace-id>` | 사용자가 선택한 workspace root 접근 | task claim blocked |
| `telemetry-sync` | SaaS progress/result metadata sync | local-only mode |
| `review-demo-mode` | provider 없이 심사용 demo 동작 허용 | 실제 provider status 만 표시 |

Consent 는 mutable setting 이지만 audit 관점에서는 append-only ledger 로 남긴다. 현재 상태 view 는 마지막 accepted/revoked record 로 계산한다.

subject 규칙:

1. `provider-execute` 는 `ProviderKind` subject 가 필수이며 workspace id 를 함께 담지 않는다.
2. `workspace-access` 는 `workspace_id` subject 가 필수이며 provider 를 함께 담지 않는다.
3. `background-helper` / `telemetry-sync` / `review-demo-mode` 는 global consent 이므로 provider / workspace subject 를 담지 않는다.
4. `revoked` record 는 삭제가 아니라 새 fact 다. 현재 view 만 false 로 계산된다.

## 6. Store channel policy

| Surface | `developer-id` | `mac-app-store` | `msix-sideload` | `msix-store` |
| --- | --- | --- | --- | --- |
| Provider CLI bundling | 금지 | 금지 | 금지 | 금지 |
| Provider CLI user-selected path | 허용 | 제한 허용: sandbox/security-scoped grant 필요 | 허용 | 허용 |
| Silent provider auto-install | 금지 | 금지 | 금지 | 금지 |
| Background helper | 허용: 명시 동의 | 제한 허용: login item/helper + sandbox 검토 | 허용: 명시 동의 | 제한 허용: full trust / Store policy 검토 |
| Direct LaunchAgent install | 허용 가능 | 금지 | 해당 없음 | 해당 없음 |
| Windows service install | 해당 없음 | 해당 없음 | 기본 금지 | Store 리스크: 기본 금지 |
| External TCP listener | 금지 | 금지 | 금지 | 금지 |
| Local IPC | 허용 | container/app group 안에서 허용 | 허용 | 허용 |
| Self-updater | 허용 가능 | 금지 | 허용 가능 | 금지: Store 업데이트 우선 |
| Arbitrary home scan | 금지 | 금지 | 금지 | 금지 |

Store policy gate 는 새 runtime security gate 번호를 추가하지 않는다. C11 이 channel 을 판정하고 C7 에 "이 surface 가 이 channel 에서 허용되는가" 를 묻는다. 결과는 C3 compatibility / C4 pre-execute / C6 workspace grant 단계에서 사용된다.

실행 함수는 후속 C7 policy migration slice 의 `internal/policy.EvaluateStoreChannelPolicy` 가 소유한다. 이 함수는 위 표를 그대로 실행하며 provider CLI 실행, helper 설치, IPC listener open, OS entitlement inspection 을 하지 않는다. `dev-local` 은 개발자 배포 실험을 위해 `developer-id` 와 같은 완화 축으로 취급하지만, Provider CLI bundling / silent provider auto-install / external TCP listener / arbitrary home scan 같은 항상 금지 surface 는 동일하게 거절한다.

## 7. Server-facing metadata

Daemon 이 C10 으로 보낼 수 있는 distribution metadata 의 executable model 은
`internal/hostintegration.BuildServerFacingClientMetadata` 다. 이 함수는 local
`ExternalToolRegistry` 를 읽되 C10 에는 routing 에 필요한 최소 status 만 넘긴다.
전송 가능/금지 field 의 실행 가능한 policy artifact 는
`internal/hostintegration/privacy_metadata_allowlist.riido.json`
(`riido-privacy-metadata-allowlist.v1`) 이며, `LoadPrivacyMetadataAllowlist`
와 privacy metadata tests 가 아래 shape 와 artifact 의 일치를 검증한다.

```
ServerFacingClientMetadata {
    distribution_channel
    app_version
    providers[] {
        provider_kind
        provider_available
        provider_login_status
        routing_status
    }
}
```

Field boundary:

| Field | 전송 가능? | 이유 |
| --- | --- | --- |
| `distribution_channel` | 가능 | server routing / store-safe policy |
| `app_version` | 가능 | compatibility / rollout |
| `provider_kind` | 가능 | assignment routing |
| `provider_available` | 가능 | assignment routing |
| `provider_login_status` | 가능 | routing 후보 제외 / UI 안내 |
| `routing_status` | 가능 | C10 sync API / UI / scheduler 공통 vocabulary |
| `provider_executable_path` | 금지 | user filesystem privacy |
| `workspace_root_path` | 금지 | user filesystem privacy |
| provider token / API key | 금지 | secret |

규칙:

1. `distribution_channel` / `app_version` 은 envelope 에 한 번만 담는다.
2. provider list 는 `ProviderKind` 기준 deterministic order 로 만든다.
3. `provider_available=false` 는 실패가 아니라 C10 routing 에서 후보 제외 신호다.
4. `routing_status` vocabulary 는 `available`, `login-required`, `unsupported`, `store-blocked` 네 값만 허용한다. `store-blocked` 는 provider 가 설치되어 있어도 store channel policy 때문에 C10 routing 후보에서 제외되는 상태다.
5. C10 은 이 metadata 를 assignment/capability routing 입력으로만 쓰며, executable path / workspace absolute path / token / API key 를 받거나 저장하지 않는다.
6. full capability fingerprint 나 binary version 은 C3/C4 capability sync 의 별도 계약이 생기기 전까지 이 metadata 에 섞지 않는다.
7. C10 `provider-status` request 에서 받을 수 있는 subset 도 같은 artifact 의 `c10-provider-status-sync-request` surface 가 소유한다. 이 request 는 daemon/runtime identity 와 `distribution_channel`, optional `app_version`, `providers[].provider_kind`, `providers[].routing_status` 만 받으며 `provider_available` / `provider_login_status` 는 C11 projection 내부 field 로만 남는다.

## 8. Review / demo mode

Store 심사용 demo mode 는 다음을 제공한다.

1. provider CLI 미설치 상태 표시.
2. provider 연결 flow 의 read-only preview.
3. workspace grant flow 의 fake/demo root.
4. background helper consent UI.
5. SaaS 없이 local-only status 화면.
6. privacy / telemetry 설정 화면.

Demo mode 는 실제 provider process 를 spawn 하지 않는다. 따라서 C4 Provider Runtime 이 아니라 C11 Store App / Local Helper control surface 의 기능이다.

현재 순수 결정 함수는 `internal/hostintegration.EvaluateReviewDemoMode` 다. 이 함수는 store-managed channel(`mac-app-store`, `msix-store`) 에서 `ConsentLedger` 의 `review-demo-mode` grant 가 있을 때만 review/demo surface 를 활성화한다. 활성화된 demo mode 도 provider execution 과 telemetry sync 를 허용하지 않는다. 즉 reviewer 가 보는 것은 onboarding / provider status preview / workspace grant flow / background consent / privacy setting / local status 화면이며, 실제 C4 provider spawn 과 C10 telemetry sync 는 열리지 않는다.

Store App / helper adapter 가 쓰는 local control surface 는 후속 local API migration slice 의 `internal/riidoapi` `review-demo` method 다. Request 는 `distribution_channel` 과 `review_demo_consent_granted` 만 받으며, response schema 는 `riido-api-review-demo.v1` 이다. Response 는 `enabled`, `surfaces`, `provider_status_mode`, `provider_execution_allowed=false`, `telemetry_sync_allowed=false`, `local_only=true` 로 reviewer-facing UI 가 어느 화면을 열 수 있는지 알려준다. `riido api review-demo --channel mac-app-store --review-demo-consent-granted true` 는 같은 계약을 CLI 로 검증하는 adapter 이며, provider CLI 를 탐지하거나 실행하지 않는다.

## 9. Open questions

`open-questions.md` 위임.

- `Q-DIST-001`: `mac-app-store` 에서 외부 provider CLI 실행을 어떤 entitlement / security-scoped bookmark 조합으로 허용 가능한지의 최종 심사 전략.
- `Q-DIST-002`: RESOLVED by §4.2. `msix-store` 는 packaged full-trust helper/tray process 를 local runtime broker 로 쓰고 Windows service default install 은 금지한다.
- `Q-DIST-003`: ConsentLedger 의 저장 substrate 를 local JSON append log 로 둘지, OS secure storage 를 섞을지.
- `Q-DIST-004`: RESOLVED in the private source open-questions SSOT as `Q-DIST-004` → `Q-CTX-005`. Store App repo ownership 은 C11 context ownership 질문 하나로만 추적한다.

## 10. version-affecting changes

- 새 `DistributionChannel` 추가는 `change:additive` 이지만 StoreChannelPolicy 표 갱신이 필수다.
- provider CLI bundling 금지 invariant 완화는 `change:breaking-policy` 이며 법무/스토어 심사 ADR 없이는 불가하다.
- Local IPC endpoint kind 추가는 C11 additive change 다. C1~C10 이 OS별 IPC 를 import 하게 만들면 context boundary 위반이다.
