# 5.1 Unsafe Bypass Enforcement

[Back to enforcement locations](../enforcement-locations.md)

`internal/policy.EvaluateUnsafeBypass` 는 위 매트릭스의 순수 결정 함수다. C4
provider adapter 는 unsafe bypass 를 실제 provider flag 로 변환하기 직전에 이
결정을 호출한다. C4 provider-runtime 후속 migration slice 가 집행할 표면:

| Surface | 집행 위치 | Host / Unknown 기본 |
| --- | --- | --- |
| Claude `--permission-mode bypassPermissions` | C4 Claude provider `BuildStart` | 거절 |
| Cursor `--yolo` | C4 Cursor provider `BuildStart` | 거절 |
| Codex `--yolo` | C4 Codex provider custom arg filter (`--yolo`, `--yolo=*`) | 거절 |
| Codex `--dangerously-bypass-approvals-and-sandbox` | C4 Codex provider custom arg filter (`--dangerously-bypass-approvals-and-sandbox`, `--dangerously-bypass-approvals-and-sandbox=*`) | 거절 |

Codex 의 approval-bypass 표면은 현재 local daemon command builder 가 생성하지
않으며, free-form `CustomArgs` 에서도 차단된다. Codex `--sandbox
danger-full-access` 는 §4.3 의 provider full-access runtime envelope 로서 C4
command builder 가 직접 생성하며, §5 unsafe bypass policy bundle surface 로 다루지
않는다. 추후 provider-native approval bypass 표면을 의도적으로 허용하는 PR 은
`StartOptions` 같은 명시적 입력과 `internal/policy.EvaluateUnsafeBypass` 호출을
먼저 추가해야 하며, Host / Unknown trust tier 는 여전히 항상 거절한다.

Cursor `--trust` 는 이 표의 unsafe bypass surface 가 아니다. 이는 Cursor Agent 가
daemon 이 선택한 task-scoped workspace 에서 headless 로 실행될 때 interactive
workspace trust prompt 로 멈추지 않게 하는 workspace acknowledgement 다. `--trust`
를 붙여도 Cursor `--yolo`, `-f`, tool auto-approval, sandbox 우회가 암묵적으로
허용되지 않는다. 따라서 Cursor adapter 는 daemon task workdir 을 지정할 때 `--trust`
를 붙일 수 있지만, `--yolo` 는 위 매트릭스와
`internal/policy.EvaluateUnsafeBypass` 집행을 계속 통과해야 한다.
