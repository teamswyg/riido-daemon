# Adjacent SSOT Contracts

[Back to runtime-actor-boundary.md](../runtime-actor-boundary.md)

| 인접 SSOT | 본 문서가 호출 / 위임 / 요구 |
| --- | --- |
| public `riido-contracts/docs/20-domain/provider-capability.md` (C3) | `Provider.Capability()` 는 C3 `ProviderCapability` 를 그대로 반환. capability 의 변경은 C3 의 책임. |
| public `riido-contracts/docs/20-domain/ir-event-log.md` (C2) | adapter 가 만드는 draft 의 `Type` 은 §3 카탈로그에 등록된 것만. append 권한은 §5.0 분리. |
| public `riido-contracts/docs/20-domain/ir-schema-versioning.md` | 9+2+FSMVersion 의무 필드 중 adapter 가 채울 수 있는 것은 §4.1 의 허용 필드만. 나머지는 ingest 가 확정. |
| public `riido-contracts/docs/20-domain/task-lifecycle.md` (C1) | adapter 는 transition 판정을 하지 않는다. `RunReportedDone` 같은 신호는 ingest/orchestrator 가 transition 으로 해석. |
| [`./runtime-scheduling.md`](../../runtime-scheduling.md) (C5) | `Provider` 인스턴스 한 개, lease 한 개, (`RuntimeID`, `CapabilityFingerprint`) 한 페어. |
| [`./workspace.md`](../../workspace.md) (C6) | adapter 는 workdir 경로 / native config 파일을 읽기만 한다. workdir 생성은 C6 가 사전에 수행. |
| [`./security.md`](../../security.md) (C7) | `ExposesUnsafePermissionBypass` 가 true 라도 sandbox / permission / hook 활성 여부는 C7 의 정책 게이트가 결정. |
| [`../30-architecture/compatibility-gate.md`](../../../30-architecture/compatibility-gate.md) | G5 (Pre-Execute) 핸드셰이크는 adapter 측 `initialize` / `firstline` probe 를 호출. 그 결과로 lease 가 활성화. |
| [`../30-architecture/runtime-upgrade-flow.md`](../../../30-architecture/runtime-upgrade-flow.md) | adapter 가 `Running` 도중 `RuntimeID`/`CapabilityFingerprint` 변경을 감지하면 draft `Type=RuntimePinViolated` 를 발행하고 process 를 stop. |
