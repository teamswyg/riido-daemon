# Store Distribution Architecture SSOT: Part 01

[Back to store-distribution.md](../store-distribution.md)


> **이 문서가 Riido daemon 을 App Store / Microsoft Store / Developer ID / MSIX sideload 로 배포하기 위한 architecture, role split, packaging target, review gate 의 SSOT다.**
>
> - 책임: store 심사를 통과하기 위해 어떤 제품 역할과 packaging target 으로 나누는가, 어떤 artifact 를 만들고 어떤 자동 검증을 통과해야 하는가.
> - 비책임: C11 도메인 결정은 [`../20-domain/distribution-host-integration.md`](../20-domain/distribution-host-integration.md), security policy 는 [`../20-domain/security.md`](../20-domain/security.md) 가 소유한다. Provider capability shared contract 는 public `riido-contracts`, SaaS control-plane / review account 는 public `riido-control-plane` 이 소유한다.

## 0. 결정

1. **Provider CLI 는 package artifact 에 포함하지 않는다.** Claude / Codex / OpenClaw / Cursor Agent 는 사용자 설치 외부 도구다.
2. **Mac 은 Developer ID notarized 배포를 1차 target 으로 둔다.** Mac App Store 는 sandbox/helper/workspace grant 제한 모드가 준비된 뒤 별도 target 으로 다룬다.
3. **Windows 는 MSIX sideload 를 1차 target 으로 둔다.** Microsoft Store 는 packaged desktop app / full trust / background policy 검증이 붙은 뒤 별도 target 으로 다룬다.
4. **Store App 과 Local Helper 를 역할상 분리한다.** UI 가 consent / provider status / workspace grant 를 소유하고, helper 는 local-only IPC 와 task execution orchestration 을 소유한다.
5. **스토어 심사용 demo/review mode 를 제공한다.** 심사자는 provider CLI 없이도 onboarding, consent, status, privacy flow 를 확인할 수 있어야 한다.

## 1. Target matrix

| Target | Artifact | Status | 핵심 blocker |
| --- | --- | --- | --- |
| `developer-id` | signed/notarized macOS app + helper | preferred first | signing/notarization pipeline, helper consent UI |
| `mac-app-store` | sandboxed Mac App Store app | requires redesign | App Sandbox, app group/helper, security-scoped workspace grant, no direct LaunchAgent install |
| `msix-sideload` | signed MSIX | preferred first | Windows named pipe, package local data, manifest/signing |
| `msix-store` | Microsoft Store MSIX packaged desktop app | requires policy gate | runFullTrust explanation, no service install by default, privacy/review notes |
| `dev-local` | `go run` / launchd plist | existing | not a store artifact |

The helper binary that Desktop/MSIX launchers download is published as a
GitHub Release asset by
[`release-artifacts.md`](release-artifacts.md). Store package updates remain a
Desktop packaging concern; the daemon release asset is the user-data helper
binary source and must not contain provider CLIs or secrets.

### 1.1 MSIX acceptance criteria

정책 snapshot: 2026-05-26 기준 Microsoft Learn 의 Microsoft Store Policies v7.19(2025-09-10 published, 2025-10-14 effective), packaged desktop app distribution, MSIX package upload, MSIX signing 문서를 확인했다. 이 외부 정책이 바뀌면 본 문서와 C11 distribution SSOT 를 같은 work unit 에서 갱신한다.

`msix-sideload` 는 현재 제품 구조에서 **가능한 1차 Windows 배포 target** 이다. 완료 기준은 다음이다.

1. `.msix` / `.msixbundle` 산출물이 신뢰 가능한 certificate 로 signed 되어야 한다. Store 밖 배포에서는 사용자/조직 device 가 그 certificate 를 신뢰해야 한다.
2. manifest 는 packaged desktop app 으로 package identity 를 제공하고 target device family 를 Windows Desktop 으로 둔다.
3. daemon state 는 package local data root 아래에만 저장한다. 임의 home scan / 하드코딩된 user path fallback 은 금지다.
4. local control surface 는 Windows named pipe 로만 열고 external TCP listener 는 금지다.
5. background helper/startup 동작은 explicit consent 가 있어야 하며 Windows service install 은 기본 금지다.
6. Claude / Codex / OpenClaw / Cursor Agent CLI 는 package 에 포함하지 않고 사용자가 설치한 외부 tool 로만 등록한다.

`msix-store` 는 **가능하되 Store 심사 evidence 가 필요한 target** 이다. sideload 기준 전체에 더해 다음이 필요하다.

1. packaged desktop app / full-trust 사용 이유를 Partner Center submission note 에 설명한다. provider runtime orchestration 이 필요한 local helper 책임과 사용자가 볼 수 있는 consent UI 를 같이 적는다.
2. update 는 Microsoft Store package update 경로를 우선한다. 앱 내부 self-updater 는 금지 surface 로 둔다.
3. Windows service install 은 기본 금지다. background 실행은 packaged app/full-trust 정책 검토와 explicit consent 를 동시에 요구한다.
4. review/demo mode 로 provider CLI 없이 onboarding, provider connection status, workspace grant, privacy/telemetry setting 을 검증할 수 있어야 한다.
5. privacy policy / Store metadata 는 SaaS 로 전송하지 않는 값(provider executable path, workspace absolute path, token/API key)을 명시해야 한다.

현재 executable 기준은 `packaging/store/riido_daemon_store_distribution.riido.json` 과 `tools/storecontract` 가 집행한다. `go run ./tools/storecontract -contract packaging/store/riido_daemon_store_distribution.riido.json -repo .` 는 MSIX channel 에서 signed package, Windows Desktop target, named pipe IPC, package local data, runFullTrust review note, Store-managed update, review notes, provider CLI non-bundling 금지, demo/review account, privacy metadata allowlist 를 검사한다. 또한 `msix-store` 의 runtime role 이 `msix-packaged-full-trust-helper-tray`, background rule 이 `explicit-consent-and-store-review`, IPC transport 가 `windows-named-pipe`, data root 가 `windows-package-local-data`, update mechanism 이 `store-managed` 인지도 검사한다. `NOTICE.md` 에 Multica provenance / Modified Apache License 2.0 fact / provider CLI non-bundling 문구가 남아 있는지 확인한다. GitHub Actions 의 `.github/workflows/store-distribution-contract.yml` 이 PR / `main` push 에서 같은 gate 를 실행한다. Provider-free review/demo local control surface 는 `internal/hostintegration` 테스트와 public daemon CI 가 검증한다.

### 1.2 Mac App Store acceptance criteria

정책 snapshot: 2026-05-28 기준 Apple App Review Guidelines, App Sandbox entitlement documentation, App Store Connect App Sandbox information, Service Management / `SMAppService`, security-scoped bookmark 문서를 확인했다. 이 외부 정책이 바뀌면 본 문서와 C11 distribution SSOT 를 같은 work unit 에서 갱신한다.

`mac-app-store` 는 현재 제품 구조에서 **가능하지만 제한 모드 redesign 이 필요한 target** 이다. 완료 기준은 다음이다.

1. app target 은 App Sandbox 를 켜고 필요한 entitlement 만 선언한다. temporary exception entitlement 가 필요하면 App Store Connect review note 에 assess 방법과 필요 이유를 적는다.
2. Store App / helper / broker 는 self-contained app bundle 안에 있어야 한다. third-party installer, shared location code install, standalone code download 는 금지다.
3. background helper 는 `SMAppService` / Login Item 계열로 사용자 동의 후 등록하고, 직접 `~/Library/LaunchAgents` 를 설치하는 방식은 금지다.
4. workspace 접근은 security-scoped bookmark 또는 App Sandbox 가 허용하는 user-selected document/folder grant 로만 지속된다.
5. 외부 Provider CLI 실행은 user-selected/sandbox/security-scoped OS grant 와 App Review approval 이 모두 있어야 한다. 둘 중 하나라도 없으면 Store App 은 provider 를 detected / login-required / store-blocked 로 보여줄 수 있지만 local helper 는 provider process 를 spawn 하지 않는다.
6. local control surface 는 app group/container 내부 local IPC 로만 열고 external TCP listener 는 금지다.
7. update 는 Mac App Store update 경로를 사용한다. 앱 내부 self-updater 는 금지 surface 다.
8. root privilege escalation / setuid 성격의 동작은 금지다.
9. review/demo mode 로 provider CLI 없이 onboarding, provider connection status, workspace grant, privacy/telemetry setting 을 검증할 수 있어야 한다.
10. privacy policy / Store metadata 는 SaaS 로 전송하지 않는 값(provider executable path, workspace absolute path, token/API key)을 명시해야 한다.

현재 executable 기준은 `tools/storecontract` 가 `mac-app-store` channel 에서 App Sandbox, app group/container IPC, security-scoped workspace grant, Service Management login item consent, helper purpose review note, App Sandbox review notes, App Store-managed updates, privacy policy, review/demo mode, demo/review account, privacy metadata allowlist, provider non-bundling review note 와 금지 surface(직접 LaunchAgent, self-updater, third-party installer, shared location code install, standalone code download, root privilege escalation)를 검사한다. RIID-4571 의 외부 Provider CLI 실행 decision 은 C7 `EvaluateStoreChannelPolicy` 가 user-selected/sandbox/security-scoped OS grant 와 Store review approval 을 동시에 요구함으로써 집행한다. GitHub Actions 의 `.github/workflows/store-distribution-contract.yml` 이 이 contract 를 별도 gate 로 검증한다.

## 2. Package boundaries

RIID-4570 decision: `riido-daemon` owns the C11 Store App contracts and local
helper runtime shape. A future desktop/app repository may own the concrete
Store App GUI adapter and OS entitlement calls, but that adapter must consume
the C11/local API contracts rather than redefining domain facts.

```
Store App
  -> concrete GUI / OS entitlement adapter (outside daemon domain)
  -> C11 Host Integration contracts
  -> Local IPC client
  -> ConsentLedger view
  -> ExternalToolRegistry view

Local Helper / Broker
  -> C11 local IPC adapter
  -> C3 ProviderCapability
  -> C4 ProviderRuntime
  -> C5 RuntimeScheduling
  -> C6 Workspace
  -> C7 SecurityPolicy
  -> C10 SaaS polling/sync adapter

SaaS Control Plane
  -> C10 assignment / SSE / routing
  -> receives distribution metadata, never provider executable paths
```

`cmd/riido` remains the local helper binary in this repository. A future GUI wrapper may live in another repo, but it must call the C11 contracts rather than bypass them.

### 2.0 Repository ownership

| Surface | Owner | Non-owner |
| --- | --- | --- |
| C11 domain facts and pure models | `riido-daemon` | Store App GUI repo must not copy/redefine them |
| Local helper / broker executable | `riido-daemon` (`cmd/riido`) | Store App GUI must not run provider CLIs directly |
| Local IPC handler and request envelope | `riido-daemon` | Store App GUI may only be a client |
| Store distribution executable contract | `riido-daemon` | Private infra must not weaken public review invariants |
| Store App native UI, entitlement calls, picker/bookmark adapter | future desktop/app repository | `riido-daemon` domain packages do not import GUI frameworks |
| Signing, provisioning, submission credentials, live evidence | private operator/infra environment | public repositories never store secrets |
| Shared DTO/schema needed by multiple repos | `riido-contracts` after promotion | no repo may fork the same fact |

### 2.1 macOS helper / login item strategy

`developer-id` 와 `mac-app-store` 는 같은 사용자 경험을 목표로 하지만 helper 등록 방식과 파일 경계가 다르다.

| Channel | Helper shape | Startup / background rule | IPC/data root |
| --- | --- | --- | --- |
| `developer-id` | signed + notarized local helper/broker. `cmd/riido` 가 현재 이 역할의 domain core 이다. | explicit consent 후 LaunchAgent 또는 Login Item 계열 등록 가능. revoke 시 자동 시작을 끈다. | `~/Library/Application Support/riido` 아래 Unix socket / app data root. 외부 TCP listener 금지. |
| `mac-app-store` | sandboxed Store App bundle 안의 helper/login item. helper 권한은 App Sandbox / entitlement review note 에 묶는다. | `SMAppService` / Login Item 계열 + explicit consent 만 허용. 직접 `~/Library/LaunchAgents` 설치는 금지. | app group 또는 sandbox container 내부 local IPC / app data root. workspace 는 security-scoped grant 로만 지속. |

공통 규칙:

1. Provider CLI 는 helper bundle 에 포함하지 않는다. C11 `ExternalToolRegistry` 가 user-selected / env-override / auto-detected provenance 만 기록한다.
2. Background helper consent 는 C11 `ConsentLedger` 의 `background-helper` grant 가 truth source 다.
3. Channel 별 허용 여부는 C7 `EvaluateStoreChannelPolicy` 의 `StoreSurfaceBackgroundHelper` / `StoreSurfaceDirectLaunchAgentInstall` 결정으로 pre-runtime 에서 확인한다.
4. Mac App Store review note 는 helper 목적, login item consent UX, sandbox entitlement 사용 이유, provider CLI non-bundling, review/demo mode 를 함께 설명해야 한다.

현재 executable C11 role plan 은 `internal/hostintegration.ResolveHelperRuntimePlan` 이다.
`mac-app-store` 는 sandboxed login item helper, `SMAppService` / Login Item
registration, helper-owned Unix socket under app group/container data root, App
Store-managed updates, no provider CLI bundling, no direct LaunchAgent install,
no shared-location code install, no standalone code download, security-scoped
workspace grant requirement, helper purpose / entitlement / consent review note
surfaces 를 한 plan 으로 반환한다. 실제 macOS Store App packaging adapter 는 이
plan 을 app bundle target, entitlements, login item registration, security-scoped
bookmark handling 에 매핑해야 한다.

### 2.2 Windows MSIX runtime / packaging strategy

`msix-sideload` 와 `msix-store` 는 같은 Windows runtime UX 를 목표로 하지만 packaging, update, review evidence 가 다르다.

| Channel | Runtime shape | Packaging rule | Review / update rule |
| --- | --- | --- | --- |
| `msix-sideload` | signed MSIX 안의 local helper/broker. provider runtime orchestration 은 helper 가 맡고 Store App 은 consent/status control surface 다. | signed package, package identity, Windows Desktop target device family, package local data, named pipe local IPC 가 필수다. | Store review note 는 필요 없지만 Windows service install 은 기본 금지이고 background helper 는 explicit consent 가 필요하다. |
| `msix-store` | Microsoft Store packaged desktop app + packaged full-trust helper/tray process. helper 는 local-only runtime broker 로 설명한다. | package identity, Windows Desktop target device family, package local data, named pipe local IPC 가 필수다. | `runFullTrust` / Partner Center notes, review/demo mode, privacy policy, Store-managed updates 가 필수이고 self-updater 는 금지다. |

공통 규칙:

1. Provider CLI 는 MSIX package 안에 포함하지 않는다. Store App 은 user-selected / env-override / auto-detected provenance 와 login-required 상태만 보여 준다.
2. Local IPC 는 Windows named pipe 만 사용한다. external TCP listener 는 금지다.
3. App data / daemon state 는 package local data root 아래에 둔다. `%USERPROFILE%` home fallback 과 arbitrary home scan 은 금지다.
4. Workspace 접근은 Windows folder picker grant + C11 consent 로만 runtime 에 전달한다.
5. `msix-store` 심사 note 는 packaged desktop app / full-trust helper 목적, background consent UX, provider CLI non-bundling, privacy scope, review/demo mode 를 함께 설명해야 한다.

현재 executable C11 role plan 은 `internal/hostintegration.ResolveHelperRuntimePlan` 이다.
`msix-store` 는 helper-owned named pipe, package local data root, Store-managed
updates, no provider CLI bundling, no Windows service default install, no
self-updater, `runFullTrust` / Partner Center review note surface 를 한 plan 으로
반환한다. 실제 Windows packaging adapter 는 이 plan 을 manifest / packaged
full-trust process / tray startup task 구현에 매핑해야 한다.

