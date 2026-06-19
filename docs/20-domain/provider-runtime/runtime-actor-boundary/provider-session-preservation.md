# Provider Session Preservation

[Back to runtime-actor-boundary.md](../runtime-actor-boundary.md)

**C4 Provider Runtime owns the provider session table.** 이 table 은 provider
native session identity 를 Riido runtime identity 에 매핑하는 C4 저장소다. C5 Runtime
Scheduling 은 이 table 의 schema / retention / adapter 를 소유하지 않고, task claim
lease 의 `(RuntimeID, CapabilityFingerprint)` 와 heartbeat 의미만 소유한다.

C5 lease 는 "어느 runtime 이 task 를 진행할 수 있는가" 를 답하고, C4 provider session
table 은 "그 runtime 위에서 어떤 provider-native session/thread 를 resume 할 수 있는가"
를 답한다.

session resume 의 안전성은 다음 페어 보존에 달려있다.

| 페어 키 | 보존 위치 |
| --- | --- |
| (`TaskID`, `RunID`, `ProviderSessionID`) | IR 이벤트(`SessionPinned` payload) + C5 lease metadata 의 runtime pin |
| (`ProviderSessionID`, `RuntimeID`) | C4 provider session table (`riido-provider-session-table.v1`) |

adapter 는 session id 를 자체 메모리에 들고 있지 않고, 즉시 draft 로 ingest 에 넘긴다.
crash 후 resume 은 IR 로그 + session table 에서 복구한다.

`riido-provider-session-table.v1` 의 최소 key 는 `(ProviderSessionID, RuntimeID)` 이다.
row 는 provider kind / protocol kind / last seen run identity / resume capability
provenance 를 담을 수 있지만, runtime eligibility 나 fencing token 을 다시 해석하지 않는다.
lease expiry / stale fingerprint / task handoff 는 C5/C9 가 소유한다.
