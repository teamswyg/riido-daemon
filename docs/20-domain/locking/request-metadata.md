# Request Metadata

[Back to Locking / Lease SSOT](../locking.md)

task DB source 가 claim 에 성공하면 `bridge.TaskRequest.Metadata` 에 다음 값을 넣는다.

| key | 의미 |
| --- | --- |
| `runtime_lease_id` | local lease sidecar 의 `lease_id` |
| `runtime_fencing_token` | task 별 monotonic fencing token |
| `runtime_capability_fingerprint` | C5 selector 가 사용한 capability fingerprint |

supervisor 는 claim metadata 에서 이 값을 typed report context 로 추출해 `StartTask` / `ReportEvent` / `CompleteTask` 호출 context 에 싣는다. task DB reporter 는 active lease 확인과 함께 `runtime_lease_id`, `runtime_fencing_token`, `runtime_capability_fingerprint` 를 sidecar 의 현재 active lease 와 비교한다. 값이 없거나 맞지 않으면 progress mutation 은 거절된다.
