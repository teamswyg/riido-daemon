# Distribution / Host Integration SSOT: Part 02

[Back to distribution-host-integration.md](../distribution-host-integration.md)

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

### 4.1.1 macOS external Provider CLI entitlement / review strategy

`Q-DIST-001` is resolved here. The Mac App Store target treats Claude / Codex /
OpenClaw / Cursor CLIs as external user-installed executables, never bundled
payloads. The Store App may help the user choose and verify a provider path, but
provider execution remains behind the local helper and C4 runtime boundary.

Policy snapshot: checked Apple App Review Guidelines and App Sandbox entitlement
documentation on 2026-05-28. The source links stay in
[`../30-architecture/store-distribution.md`](../30-architecture/store-distribution.md)
§7. If Apple changes these rules, this section and the executable policy gate
must change in the same work unit.

Rules:

1. `mac-app-store` must not use a temporary exception entitlement as the default
   strategy for provider CLI execution. Temporary exceptions require a new SSOT
   work unit and App Review note update.
2. Provider CLI path registration must start from a user action such as file
   picker / open panel, then reduce adapter-specific proof into C11 facts:
   `ExternalToolRecord{provenance=user-selected}` and
   `StoreChannelPolicyInput.OSGrantPresent=true`.
3. A sandbox/security-scoped/user-selected executable grant alone is not enough
   to execute a provider CLI in the Store channel. `StoreReviewApproved=true`
   is also required, representing App Review acceptance of the review note for
   this external-tool execution surface.
4. If either OS grant or Store review approval is missing, the provider may be
   shown as detected / login-required / store-blocked, but C4 must not spawn it.
   Review/demo mode must still work without provider CLI installation.
5. The review note must explain: provider CLIs are external user-installed
   tools; Riido does not bundle, download, or silently install them; execution
   requires explicit `provider-execute:<provider>` consent; the local helper is
   local-only; workspace access is security-scoped; no root escalation,
   LaunchAgent install, standalone code download, or shared-location code
   install is used.
6. Provider executable paths, security-scoped bookmark bytes, and entitlement
   proof stay local to the Store App/helper adapter. C10 metadata may receive
   provider kind and routing status only.

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

