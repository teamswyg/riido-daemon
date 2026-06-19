# AppDataRoot

[Back to invariants](../invariants.md)

OS별 app data root 는 `internal/hostintegration.AppDataRoot` 가 실행하는 C11
순수 모델이다. 이 모델은 OS API 를 호출하지 않고, channel / host OS / adapter 가
제공한 root 후보가 store-safe 한지만 검증한다.

```text
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
