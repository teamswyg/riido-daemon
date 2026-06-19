# RuntimeActor Boundary

[Back to runtime-actor-boundary.md](../runtime-actor-boundary.md)

`internal/agentbridge/runtimeactor` 는 한 RuntimeID 의 provider execution capacity 를
소유하는 mailbox actor 다. actor goroutine 하나가 in-flight task map 과 slot state 를
단독으로 소유한다.

RuntimeActor 의 책임:

- Start 시 adapter `Detect` 를 실행하고 public C3 `ProviderCapability` 로 reconcile 한다.
- `PolicyBundleVersion` 과 detected executable fingerprint 를 capability fingerprint input 에 포함한다.
- MaxConcurrent slot limit, duplicate task id, unavailable provider, unknown provider 를 fail-closed 로 집행한다.
- `Submit` 을 `session.Start` 로 handoff 하고 optional `ProtocolDriverProvider` 를 session 에 장착한다.
- `Cancel` / `Stop` 은 session cancel 과 process kill cascade 를 일으키고 slot 을 회수한다.
- `Status` / `HeartbeatPayload` 는 local settings UI 와 control-plane heartbeat 가 읽을 수 있는 daemon-side runtime snapshot 을 만든다.

RuntimeActor 의 비책임:

- supervisor polling / task claim / runtime selection
- EventIngestor append / task transition 결정
- workdir preparation / native config injection
- provider-specific parser / adapter implementation
- task DB / project / mwsd / local API persistence
