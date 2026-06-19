# ConsentLedger

[Back to local-ipc.md](../local-ipc.md)

다음 동의는 독립 record 로 저장한다.

현재 순수 모델은 `internal/hostintegration.ConsentLedger` / `ConsentRecord` 가 실행한다.
ledger 는 append-only facts 를 보존하고, 현재 허용 상태는 reducer view(`ConsentState`) 로
계산한다.

| Consent | 의미 | 없을 때 |
| --- | --- | --- |
| `background-helper` | app 종료 후 helper/login item/startup task 실행 | helper 자동 시작 금지 |
| `provider-execute:<provider>` | 해당 provider CLI 실행 허용 | detect 는 가능, execute 는 blocked |
| `workspace-access:<workspace-id>` | 사용자가 선택한 workspace root 접근 | task claim blocked |
| `telemetry-sync` | SaaS progress/result metadata sync | local-only mode |
| `review-demo-mode` | provider 없이 심사용 demo 동작 허용 | 실제 provider status 만 표시 |

Consent 는 mutable setting 이지만 audit 관점에서는 append-only ledger 로 남긴다. 현재 상태
view 는 마지막 accepted/revoked record 로 계산한다.

subject rules:

1. `provider-execute` 는 `ProviderKind` subject 가 필수이며 workspace id 를 함께 담지 않는다.
2. `workspace-access` 는 `workspace_id` subject 가 필수이며 provider 를 함께 담지 않는다.
3. `background-helper` / `telemetry-sync` / `review-demo-mode` 는 global consent 이므로 provider / workspace subject 를 담지 않는다.
4. `revoked` record 는 삭제가 아니라 새 fact 다. 현재 view 만 false 로 계산된다.
