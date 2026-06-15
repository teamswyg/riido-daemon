# Distribution / Host Integration SSOT: Invariants

[Back to distribution-host-integration.md](../distribution-host-integration.md)


> **이 문서가 store channel policy / host integration / external CLI provenance / local IPC / app data root / consent ledger 의 SSOT다.**
>
> - 책임: Riido daemon 이 App Store / Microsoft Store / Developer ID / MSIX sideload 같은 distribution channel 에서 어떤 host surface 를 쓸 수 있는가, provider CLI 를 어떻게 외부 도구로 등록하는가, background/helper 실행과 workspace 접근 동의를 어떻게 기록하는가.
> - 비책임: provider capability 모델은 public `riido-contracts/provider/capability` (C3), workspace materialization 은 [`./workspace.md`](./workspace.md) (C6) 이 소유한다. provider process 실행 의미(C4), security decision matrix(C7), SaaS assignment / polling(C10)은 후속 migration slice 가 각 SSOT 를 이동한다.

이 SSOT 는 **C11 Distribution / Host Integration** context 를 채운다. Context map SSOT 는 [`./context-map.md`](./context-map.md) 가 소유한다.

## 0. 핵심 invariant

1. **Provider CLI 는 번들하지 않는다.** Claude / Codex / OpenClaw / Cursor Agent executable 은 Riido package artifact 안에 들어갈 수 없다. Riido 는 사용자가 설치한 외부 CLI 를 detect / register / verify 할 뿐이다.
2. **Store app 은 control surface 다.** Riido Store App 은 local helper 상태, provider 연결 상태, workspace grant, privacy/telemetry 설정, review/demo mode 를 보여주는 사용자-facing control surface 다. provider runtime 자체를 앱 안에 숨기지 않는다.
3. **Store App GUI adapter 는 C11 계약의 consumer 다.** `riido-daemon` 은 C11 순수 모델, local helper/runtime contract, local IPC API 를 소유한다. 실제 GUI shell, OS entitlement calls, App Store/MSIX project files, file/folder picker, login-item/full-trust registration adapter 는 future desktop/app repository 가 소유할 수 있지만 C11/local API 계약을 우회할 수 없다. 이 결정은 `Q-CTX-005` 를 닫는다.
4. **Background 실행은 사용자 동의가 truth source 다.** helper/login item/startup task/background sync 는 `ConsentLedger` 의 explicit grant 가 없으면 켜지지 않는다.
5. **Local IPC 는 OS별 adapter 뒤에 둔다.** 도메인은 "local-only IPC" 만 안다. macOS Unix domain socket / app group container path, Windows named pipe / package local data path 는 C11 adapter 가 결정한다.
6. **Store channel 은 runtime capability 의 사용 가부를 제한한다.** provider 가 어떤 surface 를 지원해도 `mac-app-store` / `msix-store` policy 가 금지하면 C3 compatibility 또는 C4 pre-execute 단계에서 blocked 로 본다.
7. **Review 환경은 provider CLI 없이도 평가 가능해야 한다.** Store 심사용 demo/offline mode 는 provider 실행 없이 app shell, consent, provider connection status, workspace grant UX 를 검증할 수 있어야 한다.
8. **Public daemon binary 는 GitHub Release asset 으로 배포한다.** Desktop/MSIX launcher 는 release asset 을 사용자 app data 영역에 다운로드하고 checksum 을 검증한 뒤 실행한다. 이 경로도 provider CLI bundling 이 아니며, provider executable 은 계속 외부 사용자 설치 도구다.

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

### 2.1 Store App repository / adapter ownership

`riido-daemon` 은 Store App 이 호출해야 하는 public-safe daemon contract 를 소유한다.
이 범위는 C11 pure domain model, `ExternalToolRegistry`, `ConsentLedger`,
`WorkspaceGrantStore`, `AppDataRoot`, `LocalIPCEndpoint`, helper runtime plan,
store distribution executable contract, and local API request/response envelope 이다.

Store App GUI adapter 는 별도 desktop/app shell 로 분리할 수 있다. 그 adapter 는
화면, native entitlement API 호출, security-scoped bookmark / Windows folder picker
token 보관, `SMAppService` / packaged full-trust startup registration, App Store/MSIX
project files, review note composition 을 소유한다. 단, 다음을 지켜야 한다.

1. provider executable path / login status / workspace grant / consent 상태는 C11
   records 로 변환한 뒤 local helper API 로 전달한다.
2. GUI adapter 는 provider process 를 직접 spawn 하지 않는다. provider 실행은
   local helper (`cmd/riido`) 와 C4 runtime boundary 뒤에 둔다.
3. GUI adapter 는 provider CLI 를 bundle / download / silently install 하지 않는다.
4. GUI adapter 가 별도 repository 로 이동해도 C11 domain facts 를 복사하지 않는다.
   필요한 공유 wire type 이 두 repo 이상에서 필요해지면 `riido-contracts` promotion
   rule 을 따른다.
5. signing profiles, real provisioning secrets, App Store Connect / Partner Center
   credentials, and live submission evidence 는 public daemon repo 에 들어오지 않는다.

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
