# Runtime Eligibility

[Back to invariants](../invariants.md)

C5 의 local eligibility 판정 입력:

| 입력 | 출처 |
| --- | --- |
| `Provider` | `TaskRequest.Provider` |
| `RequiredSurfaces` | `TaskRequest.RequiredSurfaces` (+ legacy metadata fallback 가능) |
| `AllowExperimentalRuntime` | `TaskRequest.AllowExperimentalRuntime` |
| `RuntimeID` | runtime status |
| `CapabilityFingerprint` | C3 capability reconciliation |
| `SlotLimit` / `SlotsInUse` | runtime status / heartbeat |
| `Available` / `CompatibilityStatus` / `RequiresExperimentalOptIn` | C3 capability reconciliation |
| provider-neutral support flags | C3 capability reconciliation |

판정 순서:

1. task provider 와 runtime capability provider 가 일치해야 한다.
2. `Available=false` 이면 거절.
3. `CompatibilityStatus=blocked` 이면 거절.
4. `RequiresExperimentalOptIn=true` 이고 task 가 opt-in 하지 않았으면 거절.

SaaS assignment polling uses the same gate. `riido-control-plane` derives
`Assignment.allow_experimental_runtime` from the daemon-reported runtime fact at
assignment creation time, and `saasplane` copies that snapshot into
`TaskRequest.AllowExperimentalRuntime`. The daemon must not infer opt-in from a
provider name, local environment variable, team id, or Open API key.

5. `SlotLimit > 0` 이고 `SlotsInUse >= SlotLimit` 이면 거절.
6. 모든 `RequiredSurfaces` 의 capability flag 가 true 여야 한다.
