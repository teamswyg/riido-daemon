# Distribution Channels and Role Model

[Back to invariants](../invariants.md)

## 1. Distribution channel enum

| Channel | 의미 | 1차 목적 |
| --- | --- | --- |
| `developer-id` | macOS Developer ID signed + notarized outside Mac App Store | macOS 1차 배포 |
| `mac-app-store` | Mac App Store sandboxed app | 장기 목표 / 제한 모드 |
| `msix-sideload` | signed MSIX outside Microsoft Store | Windows 1차 배포 |
| `msix-store` | Microsoft Store MSIX / packaged desktop app | Windows Store 배포 |
| `dev-local` | repo / launchd / foreground development | 개발자 로컬 |

Channel 은 build artifact identity 다. 같은 daemon binary 라도 channel 이 다르면
허용 surface 가 달라진다.

## 2. Role model

| Role | 책임 | 포함하면 안 되는 것 |
| --- | --- | --- |
| **Store App** | onboarding, provider path 등록, workspace 선택, background/privacy 설정, status/review UI | provider CLI binary, silent installer, unmanaged updater |
| **Local Helper / Broker** | local-only IPC, task claim loop, provider runtime orchestration, status/metrics | 외부 TCP listener, 사용자 동의 없는 autostart |
| **Provider Connector** | user-selected/env/auto-detected executable provenance, version/login/capability probe | vendor code redistribution, provider license/TOS 대체 |
| **SaaS Control Plane** | assignment, polling, event stream, policy-aware routing metadata | 고객 PC provider process 실행 |

## 2.1 Store App repository / adapter ownership

`riido-daemon` 은 Store App 이 호출해야 하는 public-safe daemon contract 를 소유한다.
이 범위는 C11 pure domain model, `ExternalToolRegistry`, `ConsentLedger`,
`WorkspaceGrantStore`, `AppDataRoot`, `LocalIPCEndpoint`, helper runtime plan,
store distribution executable contract, and local API request/response envelope 이다.

Store App GUI adapter 는 별도 desktop/app shell 로 분리할 수 있다. 그 adapter 는
화면, native entitlement API 호출, security-scoped bookmark / Windows folder picker
token 보관, `SMAppService` / packaged full-trust startup registration, App Store/MSIX
project files, review note composition 을 소유한다. 단, 다음을 지켜야 한다.

1. provider executable path / login status / workspace grant / consent 상태는 C11 records 로 변환한 뒤 local helper API 로 전달한다.
2. GUI adapter 는 provider process 를 직접 spawn 하지 않는다. provider 실행은 local helper (`cmd/riido`) 와 C4 runtime boundary 뒤에 둔다.
3. GUI adapter 는 provider CLI 를 bundle / download / silently install 하지 않는다.
4. GUI adapter 가 별도 repository 로 이동해도 C11 domain facts 를 복사하지 않는다. 필요한 공유 wire type 이 두 repo 이상에서 필요해지면 `riido-contracts` promotion rule 을 따른다.
5. signing profiles, real provisioning secrets, App Store Connect / Partner Center credentials, and live submission evidence 는 public daemon repo 에 들어오지 않는다.
