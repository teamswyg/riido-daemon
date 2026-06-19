# Provider Stdin Command ACL

[Back to runtime-actor-boundary.md](../runtime-actor-boundary.md)

Reducer command 는 provider-neutral 이다. C4 adapter 가 provider stdin control protocol
을 가진 경우에만 `agentbridge.ProviderInputBuilder` 를 구현해 `CommandApproveTool` /
`CommandRejectTool` / `CommandWriteProviderInput` 을 concrete byte frame 으로 바꾼다.

현재 집행 표면:

| Provider | 입력 command | Concrete frame |
| --- | --- | --- |
| Claude | `CommandApproveTool` / `CommandRejectTool` | `control_response` stream-json frame. `control_request.request_id` 는 `ToolRef.ProviderRequestID` 로 보존한다. |
| Codex | JSON-RPC protocol driver 내부 | pending request id 를 driver actor 가 소유하고 JSON-RPC response 로 처리한다. |
