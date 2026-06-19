# WorkspaceGrantStore

[Back to invariants](../invariants.md)

사용자 workspace root 접근 grant 는 `internal/hostintegration.WorkspaceGrantStore`
가 실행하는 C11 순수 모델이다. C6 는 이 store 의 active grant record 만 받아
workdir prepare 단계에서 snapshot / worktree / shallow clone 으로 materialize 한다.
OS별 bookmark / picker token bytes 자체는 adapter 소유이며, 도메인은 grant method
와 subject 만 검증한다.

```text
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
