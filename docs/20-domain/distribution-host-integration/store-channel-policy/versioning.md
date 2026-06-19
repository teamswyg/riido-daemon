# Version-Affecting Changes

[Back to Store Channel Policy](../store-channel-policy.md)

- 새 `DistributionChannel` 추가는 `change:additive` 이지만 StoreChannelPolicy 표 갱신이 필수다.
- provider CLI bundling 금지 invariant 완화는 `change:breaking-policy` 이며 법무/스토어 심사 ADR 없이는 불가하다.
- Local IPC endpoint kind 추가는 C11 additive change 다. C1~C10 이 OS별 IPC 를 import 하게 만들면 context boundary 위반이다.
