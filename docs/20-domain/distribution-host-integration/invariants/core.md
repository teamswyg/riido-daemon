# Scope and Core Invariants

[Back to invariants](../invariants.md)

> **이 문서가 store channel policy / host integration / external CLI provenance /
> local IPC / app data root / consent ledger 의 SSOT다.**
>
> - 책임: Riido daemon 이 App Store / Microsoft Store / Developer ID / MSIX
>   sideload 같은 distribution channel 에서 어떤 host surface 를 쓸 수 있는가,
>   provider CLI 를 어떻게 외부 도구로 등록하는가, background/helper 실행과
>   workspace 접근 동의를 어떻게 기록하는가.
> - 비책임: provider capability 모델은 public
>   `riido-contracts/provider/capability` (C3), workspace materialization 은
>   [`../../workspace.md`](../../workspace.md) (C6) 이 소유한다. provider process
>   실행 의미(C4), security decision matrix(C7), SaaS assignment / polling(C10)은
>   후속 migration slice 가 각 SSOT 를 이동한다.

이 SSOT 는 **C11 Distribution / Host Integration** context 를 채운다. Context map
SSOT 는 [`../../context-map.md`](../../context-map.md) 가 소유한다.

## 0. 핵심 invariant

1. **Provider CLI 는 번들하지 않는다.** Claude / Codex / OpenClaw / Cursor Agent executable 은 Riido package artifact 안에 들어갈 수 없다. Riido 는 사용자가 설치한 외부 CLI 를 detect / register / verify 할 뿐이다.
2. **Store app 은 control surface 다.** Riido Store App 은 local helper 상태, provider 연결 상태, workspace grant, privacy/telemetry 설정, review/demo mode 를 보여주는 사용자-facing control surface 다. provider runtime 자체를 앱 안에 숨기지 않는다.
3. **Store App GUI adapter 는 C11 계약의 consumer 다.** `riido-daemon` 은 C11 순수 모델, local helper/runtime contract, local IPC API 를 소유한다. 실제 GUI shell, OS entitlement calls, App Store/MSIX project files, file/folder picker, login-item/full-trust registration adapter 는 future desktop/app repository 가 소유할 수 있지만 C11/local API 계약을 우회할 수 없다. 이 결정은 `Q-CTX-005` 를 닫는다.
4. **Background 실행은 사용자 동의가 truth source 다.** helper/login item/startup task/background sync 는 `ConsentLedger` 의 explicit grant 가 없으면 켜지지 않는다.
5. **Local IPC 는 OS별 adapter 뒤에 둔다.** 도메인은 "local-only IPC" 만 안다. macOS Unix domain socket / app group container path, Windows named pipe / package local data path 는 C11 adapter 가 결정한다.
6. **Store channel 은 runtime capability 의 사용 가부를 제한한다.** provider 가 어떤 surface 를 지원해도 `mac-app-store` / `msix-store` policy 가 금지하면 C3 compatibility 또는 C4 pre-execute 단계에서 blocked 로 본다.
7. **Review 환경은 provider CLI 없이도 평가 가능해야 한다.** Store 심사용 demo/offline mode 는 provider 실행 없이 app shell, consent, provider connection status, workspace grant UX 를 검증할 수 있어야 한다.
8. **Public daemon binary 는 GitHub Release asset 으로 배포한다.** Desktop/MSIX launcher 는 release asset 을 사용자 app data 영역에 다운로드하고 checksum 을 검증한 뒤 실행한다. 이 경로도 provider CLI bundling 이 아니며, provider executable 은 계속 외부 사용자 설치 도구다.
