# 9 Adjacent SSOT Contracts

[Back to native config manifest](../native-config-manifest.md)

| 인접 context | 본 문서가 받는 / 공급 |
| --- | --- |
| **C7 Security / Policy** | 받는다: `T-PATH` protected paths, `T-SBX` sandbox 정책, `T-CFG` native config plan, `T-MCP` MCP allowlist (workdir 안 MCP 설정 주입에 사용). 공급: 주입 완료 IR 이벤트 (`NativeConfigInjected`). |
| **C4 Provider Runtime / Adapter** | 공급: workdir 경로, native config 사본, `NativeConfigVersion`. 받지 않음: adapter 가 직접 workdir 을 만들지 않는다. |
| **C5 Runtime Scheduling** | 받는다: claim 된 task 의 `runID`, `RuntimeID`, `CapabilityFingerprint`. 공급: `WorkspacePrepared` 신호 (claim → lease 활성 사전조건). |
| **C2 IR Event Log** | Cat E (workspace/config) 의 1차 producer. `WorkdirCreated` / `NativeConfigInjected` / `WorkdirArchived` / `ConfigTemplateReinjected` 발행은 EventIngestor API 를 통해서만 수행한다. local workdir adapter 는 run root 의 `ir/events.jsonl` sink 를 제공할 수 있지만, envelope 확정 권한은 C2 EventIngestor 에 있다. C2 event schema 는 public `riido-contracts` 가 소유한다. |
| **C8 Validation** | 공급: workdir 의 base / final 상태, diff, artifacts. validation 은 본 디렉토리들을 **읽기 전용** 으로 본다. |
| **C9 Locking** | 받는다: `flock` primitive (§8 의 도메인 lock 들이 실제로 사용하는 메커니즘). |
| **컨테이너 / VM 매니저 (외부)** | tier=`IsolatedContainer`/`EphemeralVM` 인 경우 C6 는 host-side run root 와 manifest 를 준비하고, container/VM 안으로 mount/전달하는 책임은 C4 runtime launcher / platform adapter 가 갖는다. C6 는 mount primitive 나 VM lifecycle 을 소유하지 않는다. |
