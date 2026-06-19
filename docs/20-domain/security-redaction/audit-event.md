# Audit Event

[Back to Security Redaction SSOT](../security-redaction.md)

EventIngestor 2차 redaction 이 발생하면 `PolicyViolationDetected` 를 append 한다.

| field | 값 |
| --- | --- |
| `Type` | `PolicyViolationDetected` |
| `Scope` | source draft 와 같은 `TaskScope` 또는 `RunScope` |
| `ActorKind` / `ActorID` | source event 와 같은 attribution |
| `Payload.category` | `SECRET_LEAK_ATTEMPTED` |
| `Payload.subject` | redacted pattern IDs, comma-separated stable order |
| `Payload.severity` | `high` |
| `Payload.sourceEventID` | redacted source event ID |
| `Payload.sourceEventType` | redacted source event type |
| `Payload.redactedFields` | redaction 이 발생한 payload/unknown field path 목록 |

audit event 는 redacted source event 를 mutate 하지 않는다. append-only correction 이 필요한 경우에는 public `riido-contracts/ir` C2 correction 규칙을 따른다.
