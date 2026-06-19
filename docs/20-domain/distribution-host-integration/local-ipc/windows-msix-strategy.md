# Windows MSIX Strategy

[Back to local-ipc.md](../local-ipc.md)

`msix-sideload` 와 `msix-store` 는 Windows 에서 같은 local runtime 책임을 갖되 심사/업데이트 경계가 다르다.

| Channel | Runtime shape | Packaging/data rule | Startup/background rule |
| --- | --- | --- | --- |
| `msix-sideload` | signed MSIX 안의 Store App + local helper/broker. helper 는 package identity 아래에서 provider runtime orchestration 을 수행한다. | package identity, Windows Desktop target device family, package local data root, Windows named pipe local IPC 가 필수다. | explicit `background-helper` consent 가 있어야 하며 Windows service install 은 기본 금지다. |
| `msix-store` | Microsoft Store packaged desktop app + packaged full-trust helper/tray process. helper 는 Store submission note 에 설명된 local-only runtime broker 다. | package identity, Windows Desktop target device family, package local data root, named pipe local IPC, Partner Center / `runFullTrust` review note 가 필수다. | Store-managed update 를 사용하고 self-updater / Windows service default install 은 금지다. background helper 는 consent 와 Store review approval 이 모두 필요하다. |

Common rules:

1. Claude / Codex / OpenClaw / Cursor Agent CLI 는 MSIX package 에 포함하지 않는다.
2. C11 `ExternalToolRegistry` 는 user-selected / env-override / auto-detected provenance 만 보존한다.
3. Store App 과 helper 사이의 control surface 는 Windows named pipe 로만 열린다. external TCP listener 는 금지다.
4. App data 와 daemon state 는 package local data root 아래에 둔다. `%USERPROFILE%` 또는 임의 home scan fallback 은 금지다.
5. 사용자 workspace 접근은 `WorkspaceGrantStore` 의 `windows-folder-picker-grant` 와 `ConsentLedger` 의 `workspace-access:<workspace-id>` 가 모두 active 일 때만 C6 materialization 으로 전달한다.
6. `msix-store` 는 provider 없이도 onboarding, provider status, workspace grant, background consent, privacy setting 을 볼 수 있는 review/demo mode 를 제공해야 한다.

현재 순수 runtime role 모델은 `internal/hostintegration.ResolveHelperRuntimePlan` 이다.
Windows adapter 는 이 plan 을 실제 MSIX manifest / packaged full-trust process / tray
startup task 구현의 입력으로만 써야 하며, 이 함수 자체는 provider process 를 spawn 하거나
named pipe 를 열거나 Windows service 를 설치하지 않는다.

MSIX helper plan invariant:

1. `msix-store` role 은 `msix-packaged-full-trust-helper-tray` 이고 local IPC 는 helper-owned Windows named pipe 다.
2. `msix-store` background 실행은 `background-helper` consent 와 Store review approval 이 모두 있어야 allowed 다.
3. `msix-store` 는 Store-managed updates 를 사용하며 self-updater / Windows service default install / provider CLI bundling 을 허용하지 않는다.
4. `msix-sideload` role 은 `msix-packaged-helper-broker` 이고 background 실행은 explicit consent 만 요구한다.
5. 두 MSIX channel 모두 app data root scope 는 `windows-package-local-data` 여야 한다.
