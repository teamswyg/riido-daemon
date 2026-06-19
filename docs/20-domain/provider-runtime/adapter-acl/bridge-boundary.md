# Bridge Boundary

[Back to adapter-acl.md](../adapter-acl.md)

`internal/agentbridge/bridge` 는 C4 provider runtime 의 provider-neutral library entrypoint 다. caller 는 adapter 목록과 process port 를 주입하고, bridge 는 다음만 수행한다.

- adapter registry 를 만들고 provider name 중복 / empty name 을 거부한다.
- `Detect(ctx)` 호출을 provider name 기준 stable order 로 반환한다.
- `TaskRequest` 를 `agentbridge.StartRequest` 로 변환해 adapter `BuildStart` 를 호출한다.
- SaaS assignment source 에서 온 `Assignment.agent_instruction` 을 provider 별 runtime instruction 으로 materialize 한다.
- `StartCommand` 를 `process.Command` 로 변환하고 `internal/agentbridge/session` 을 시작한다.
- adapter 가 `ProtocolDriverProvider` 이면 one-run protocol driver 를 생성해 session 에 장착한다.
- session handle facade 를 반환하고 adapter `DroppedArgs` / `TempFiles` 를 session 경계까지 보존한다.

Claude / OpenClaw 는 system prompt surface 를 쓰고, Codex / Cursor 는 prompt prefix 를 쓴다. instruction 값의 의미와 1000자 제한은 `riido-contracts` 가 소유한다.

`bridge` 는 scheduler, task claim, EventIngestor append, workdir preparation, policy decision, provider-specific parsing 을 소유하지 않는다. 이 책임들은 각각 C5, C2/C4 RunController, C6, C7, concrete adapter slice 가 소유한다.
