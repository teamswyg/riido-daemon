# Provider Adapter Implementations

[Back to runtime-actor-boundary.md](../runtime-actor-boundary.md)

각 어댑터의 capability, protocol, priority 는 public
`riido-contracts/docs/20-domain/provider-capability.md` §4 가 소유한다. 이 문서는
provider adapter 가 C4 Provider Runtime 컨텍스트에서 어떻게 표현되는지만 적는다.

현재 public 구현 상태:

- RIID-4658: `ClaudeStreamJSONAdapter` moved to `internal/provider/claude`.
- RIID-4659: Codex app-server adapter moved to `internal/provider/codex`.
- RIID-4660: OpenClaw adapter moved to `internal/provider/openclaw`.
- RIID-4661: Cursor adapter moved to `internal/provider/cursor`.

| Adapter | 본 컨텍스트의 표현 | 1차 draft 카테고리 |
| --- | --- | --- |
| `ClaudeStreamJSONAdapter` | `claude -p --output-format stream-json` 의 stdout 라인을 NDJSON 으로 흡수. session id 는 `system.init` 라인에서 추출. | Cat C 위주 |
| `CodexAppServerAdapter` | `codex --sandbox danger-full-access app-server --listen stdio://` JSON-RPC. sandbox selection 은 daemon-owned full-access harness envelope 다. | Cat C + approval (`ApprovalRequested`) |
| `OpenClawAgentJSONAdapter` | `openclaw agent --local --json` 의 JSON/NDJSON 출력을 흡수. calendar-version gate 로 unsupported CLI 를 unavailable 로 접는다. | Cat C 위주 |
| `CursorAgentStreamJSONAdapter` | `cursor-agent -p --output-format stream-json` root-print shape 를 기본으로 사용하고, version/profile 차이는 explicit launch profile 로만 선택한다. | Cat C 위주 |

위 4 어댑터는 모두 같은 `Provider` 포트를 구현하고 같은 `ProviderEventDraft` 출력을 갖는다.
어댑터 별 분기는 provider-capability §0 invariant 1 에 따라 `ProtocolKind` 로만 한다.
