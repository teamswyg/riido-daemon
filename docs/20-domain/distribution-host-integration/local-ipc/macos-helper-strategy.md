# macOS App Store Helper Strategy

[Back to local-ipc.md](../local-ipc.md)

`mac-app-store` 는 sandboxed Store App 과 bundle 내부 helper/login item 의 역할을 분리한다.
Store App 은 onboarding, provider status, workspace grant, background/privacy setting,
review/demo mode 를 보여 주고, helper 는 local-only IPC 와 task execution orchestration 을
담당한다.

Rules:

1. helper role 은 `sandboxed-login-item-helper` 이며, background 등록 방식은 `SMAppService` / Login Item 계열만 허용한다.
2. helper background 실행은 `background-helper` consent 와 App Store review approval 이 모두 있어야 allowed 다.
3. direct `~/Library/LaunchAgents` 설치, self-updater, shared-location code install, standalone code download, provider CLI bundling 은 금지다.
4. app data root 는 app group 또는 sandbox container 여야 하며, user home `Application Support` fallback 은 금지다.
5. local IPC 는 helper-owned Unix domain socket 이고 app group/container root 내부에 있어야 한다.
6. user workspace 접근은 `WorkspaceGrantStore` 의 `security-scoped-bookmark` 와 `ConsentLedger` 의 `workspace-access:<workspace-id>` 가 모두 active 일 때만 C6 materialization 으로 전달한다.
7. App Store review note 는 helper purpose, Login Item consent UX, App Sandbox entitlement 사용 이유, security-scoped workspace grant, provider CLI non-bundling, review/demo mode, privacy metadata allowlist 를 설명해야 한다.

현재 순수 runtime role 모델은 `internal/hostintegration.ResolveHelperRuntimePlan` 이다. 이
함수는 channel-approved `AppDataRoot` 와 `LocalIPCEndpoint` 를 받아 macOS Store
App/helper adapter 가 구현해야 할 role, startup registration, background rule, workspace
grant requirement, update rule, review note surfaces 를 계산한다.

실제 Store App bundle, entitlements, `SMAppService` 호출, security-scoped bookmark bytes 는
C11 adapter / packaging target 소유이며 이 함수는 OS API 를 호출하지 않는다.

macOS Store helper plan invariant:

1. `mac-app-store` role 은 `sandboxed-login-item-helper` 이고 startup registration 은 `service-management-login-item` 이다.
2. `mac-app-store` local IPC 는 helper-owned Unix socket 이며 app group 또는 sandbox container root 아래에 있다.
3. `mac-app-store` background 실행은 `background-helper` consent 와 Store review approval 이 모두 있어야 allowed 다.
4. `mac-app-store` 는 App Store-managed updates 를 사용하며 self-updater / direct LaunchAgent / shared-location install / standalone code download / provider CLI bundling 을 허용하지 않는다.
5. `mac-app-store` workspace grant requirement 는 `security-scoped-bookmark` 다.
