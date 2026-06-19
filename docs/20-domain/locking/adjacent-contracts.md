# Adjacent SSOT Contracts

[Back to Locking / Lease SSOT](../locking.md)

| 인접 context | 본 문서가 받는 / 공급 |
| --- | --- |
| **C5 Runtime Scheduling** | 받는다: selected `(RuntimeID, CapabilityFingerprint)`, lease expiry 의미. 공급: local claim primitive 와 fencing token. |
| **C1 Task Lifecycle** | 받는다: `Queued → Claimed` 등 상태 전이. 공급: 전이를 저장하기 전의 single-writer critical section. |
| **C4 Provider Runtime** | 공급: provider process 가 시작되기 전 claim 이 단일 runtime 으로 fenced 됐다는 사실. |
| **C8 Validation** | provider lease 가 release 된 뒤 검증 command 가 별도 guarded mutation 으로 진행된다는 경계를 공유한다. |
