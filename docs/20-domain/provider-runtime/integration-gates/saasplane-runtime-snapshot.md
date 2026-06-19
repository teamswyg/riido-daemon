# SaaS Runtime Snapshot Semantics

[Back to Integration Gates](../integration-gates.md)

`saasplane` 은 runtime-snapshot 을 device 단위 full set 으로 보고한다.

`RegisterRuntime` 과 heartbeat refresh 모두 단일 runtime 이 아니라 현재까지 등록된 모든
provider runtime 을 RuntimeID 정렬 순서로 post 한다.

따라서 미탐지 provider 도 `detection_state=missing` 으로 항상 set 안에 남고, snapshot
replace 의미를 쓰는 서버 projection 이 device runtime 을 빈 `[]` 로 덮어쓰지 않는다.

detected/missing 판정 자체는 runtime capability(`provider.<name>.available`)에서 파생된다.

control-plane device projection(`GET .../ai-agent/devices`)의 최종 표현/필드는 여전히
`riido-control-plane` 이 소유한다.
