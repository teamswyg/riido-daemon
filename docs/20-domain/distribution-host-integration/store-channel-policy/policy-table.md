# Store Channel Policy

[Back to Store Channel Policy](../store-channel-policy.md)

| Surface | `developer-id` | `mac-app-store` | `msix-sideload` | `msix-store` |
| --- | --- | --- | --- | --- |
| Provider CLI bundling | 금지 | 금지 | 금지 | 금지 |
| Provider CLI user-selected path | 허용 | 제한 허용: sandbox/security-scoped/user-selected executable grant + Store review approval 필요 | 허용 | 허용 |
| Silent provider auto-install | 금지 | 금지 | 금지 | 금지 |
| Background helper | 허용: 명시 동의 | 제한 허용: login item/helper + sandbox 검토 | 허용: 명시 동의 | 제한 허용: full trust / Store policy 검토 |
| Direct LaunchAgent install | 허용 가능 | 금지 | 해당 없음 | 해당 없음 |
| Windows service install | 해당 없음 | 해당 없음 | 기본 금지 | Store 리스크: 기본 금지 |
| External TCP listener | 금지 | 금지 | 금지 | 금지 |
| Local IPC | 허용 | container/app group 안에서 허용 | 허용 | 허용 |
| Self-updater | 허용 가능 | 금지 | 허용 가능 | 금지: Store 업데이트 우선 |
| Arbitrary home scan | 금지 | 금지 | 금지 | 금지 |

Store policy gate 는 새 runtime security gate 번호를 추가하지 않는다. C11 이 channel 을 판정하고 C7 에 "이 surface 가 이 channel 에서 허용되는가" 를 묻는다. 결과는 C3 compatibility / C4 pre-execute / C6 workspace grant 단계에서 사용된다.

실행 함수는 후속 C7 policy migration slice 의 `internal/policy.EvaluateStoreChannelPolicy` 가 소유한다. 이 함수는 위 표를 그대로 실행하며 provider CLI 실행, helper 설치, IPC listener open, OS entitlement inspection 을 하지 않는다. `dev-local` 은 개발자 배포 실험을 위해 `developer-id` 와 같은 완화 축으로 취급하지만, Provider CLI bundling / silent provider auto-install / external TCP listener / arbitrary home scan 같은 항상 금지 surface 는 동일하게 거절한다.
