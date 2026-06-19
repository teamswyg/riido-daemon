# Task Requirement Model

[Back to invariants](../invariants.md)

task 는 provider-neutral surface 이름으로 요구 조건을 표현한다.

| Surface | 의미 | capability flag |
| --- | --- | --- |
| `structured-event-stream` | stdout/RPC 에서 구조화된 event stream 을 받을 수 있어야 함 | `SupportsStructuredEventStream` |
| `session-resume` | session/thread resume 이 가능해야 함 | `SupportsResume` |
| `system-prompt` | system/developer instruction 을 native surface 로 전달할 수 있어야 함 | `SupportsSystemPrompt` |
| `max-turns` | turn limit 을 native surface 로 전달할 수 있어야 함 | `SupportsMaxTurns` |
| `mcp` | MCP config / tool bridge 를 지원해야 함 | `SupportsMCP` |
| `tool-hooks` | tool / hook event surface 를 지원해야 함 | `SupportsHookEvents` |
| `usage` | token usage metric 을 제공해야 함 | `SupportsUsageMetrics` |
| `worktree` | daemon 이 선택한 task-scoped workdir/worktree 를 provider 실행 표면으로 전달할 수 있어야 함 | `SupportsWorktree` |

Go surface 는 `bridge.TaskRequest.RequiredSurfaces []string` 이다. local file queue
JSON 에서는 `required_surfaces` 로 쓸 수 있다. 예:

```json
{
  "id": "task-1",
  "provider": "cursor",
  "prompt": "inspect this repo",
  "required_surfaces": ["structured-event-stream", "worktree"],
  "allow_experimental_runtime": true,
  "metadata": {
    "workspace_id": "ws-1"
  }
}
```

알 수 없는 surface 이름은 “지원한다고 추정” 하지 않는다. eligibility 는 실패한다.
