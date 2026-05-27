# Security Redaction SSOT

> **이 문서가 secret redaction 정책의 SSOT다.**
>
> - 책임: C7 secret redaction 금지 패턴 카탈로그, marker 형식, C4/C2 실행 책임, redaction audit event payload, version-affecting change 규칙.
> - 비책임: provider raw event 를 `ProviderEventDraft` 로 변환하는 concrete adapter ACL 구현은 C4 Provider Runtime 후속 provider-adapter migration slice 가 소유한다. `CanonicalEvent` schema/envelope 는 public `riido-contracts/ir` C2 계약이 소유하고, local daemon append-time completion/redaction implementation 은 public `internal/ir/ingest` 가 소유한다. `ToolRef.Args` flattening 은 public `internal/agentbridge/toolargs` 가 소유한다.

이 SSOT 는 [security.md](./security.md) 의 C7 Security / Policy context 안에서 secret exposure target 의 세부 결정을 분리해 소유한다. `security.md` 는 보안 정책 hub 이며, 이 문서의 내용을 재정의하지 않는다.

## 0. Invariants

1. **redact 결정은 C7 이 소유한다.** C4 Adapter ACL, C4 ToolRef.Args, C2 EventIngestor 는 C7 catalog 를 호출할 뿐 패턴 의미를 재정의하지 않는다.
2. **raw secret 은 IR `Payload` / `Unknown` / `ToolRef.Args` 에 저장하지 않는다.** redaction 이 필요하면 marker 만 남긴다.
3. **adapter 는 관측만 한다.** C4 1차 redaction 이 발생해도 adapter 는 별도 audit event 를 발행하지 않는다.
4. **2차 redaction audit 은 C2 EventIngestor 가 append 한다.** EventIngestor append 직전 redaction 이 발생하면 redacted source event 와 같은 sink batch 에 `PolicyViolationDetected` audit event 를 함께 append 한다.
5. **pattern catalog 진화는 policy bundle version change 로 다룬다.** 패턴 추가는 additive, 패턴 제거는 breaking-policy 다.

## 1. 금지 패턴 카탈로그

현재 C7 실행 카탈로그는 `internal/policy` 의 `secretRedactionPatterns` 가 구현한다. 패턴 ID 는 stable external evidence/audit 값이므로 이름 변경은 breaking-policy 다.

| pattern ID | 의미 |
| --- | --- |
| `aws-access-key` | AWS access key (`AKIA...`, `ASIA...`) |
| `gcp-api-key` | GCP API key (`AIza...`) |
| `github-token` | GitHub classic/fine-grained token 계열 |
| `gitlab-token` | GitLab personal/project token 계열 |
| `anthropic-api-key` | Anthropic API key |
| `openai-api-key` | OpenAI API key |
| `jwt` | JWT 모양의 bearer token |
| `pem-private-key` | PEM private key header |
| `basic-auth-url` | `https://user:pass@host` 형태의 basic-auth URL |
| `env-secret-assignment` | env assignment 형태의 `*_TOKEN=...`, `*_SECRET=...`, `*_KEY=...` 값 |

## 2. Marker

Canonical IR payload redaction marker 는 다음 형식이다.

```text
[REDACTED:<patternID>]
```

`ToolRef.Args` 는 policy classification 입력용 bounded summary 이므로 pattern ID 를 보존하지 않고 단일 marker 를 쓴다.

```text
[redacted]
```

`ToolRef.Args` 에 `[redacted]` marker 가 있으면 `tool:secret-exposure` risk surface 로 분류한다.

## 3. 실행 책임

| 단계 | 주체 | 책임 |
| --- | --- | --- |
| 정책 소유 | C7 Security Redaction (본 문서) | 금지 패턴 catalog, marker 형식, audit category/severity/subject 규칙, version-affecting change 규칙 |
| 1차 redaction | C4 Adapter ACL | provider raw 출력을 `ProviderEventDraft` 로 변환할 때 string 필드를 C7 catalog 로 스캔하고 `[REDACTED:<patternID>]` 로 치환한다. `Payload` / `Raw` / `Unknown` 모두 같은 규칙을 따른다. |
| ToolRef.Args redaction | C4 ToolRef.Args | provider raw tool input 을 bounded string map 으로 flatten 할 때 sensitive key 또는 C7 pattern match 값을 `[redacted]` 로 치환한다. |
| 2차 redaction + audit | C2 EventIngestor | `AppendDraft(...)` 에서 `CanonicalEvent` 적재 직전 `Payload` / `Unknown` 의 string 값을 다시 스캔한다. 매치되면 `[REDACTED:<patternID>]` 를 적용하고, redacted source event 와 같은 sink batch 에 audit event 를 함께 append 한다. |

비-string scalar 는 그대로 둔다. string value 를 포함하는 nested map/slice 는 같은 catalog 로 재귀 스캔한다.

## 4. Audit Event

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

## 5. Redaction Failure

패턴이 매치되었지만 redaction 을 적용할 수 없는 payload 형태가 들어오면 EventIngestor 는 해당 draft 를 거절하고 `BlockerRaised(category=SECURITY_REDACTION_FAILED)` 를 발행해야 한다. 현재 실행 subset 은 string / map / slice redaction 을 지원하고 비-string scalar 는 redaction 대상이 아니므로 실패로 보지 않는다.

## 6. Versioning

- 새 secret pattern 추가: `change:additive`.
- 기존 pattern 제거: `change:breaking-policy`.
- pattern ID rename: `change:breaking-policy`.
- Canonical marker 형식 변경: `change:breaking-policy`.
- `ToolRef.Args` marker 형식 변경: `change:breaking-policy`.
- audit payload field 추가: `change:additive`.
- audit category/severity/subject 의미 변경: `change:breaking-policy`.

## 7. 검증 게이트

- `internal/policy` tests 는 C7 pattern catalog 와 marker 형식을 검증한다.
- `internal/agentbridge/toolargs` tests 는 ToolRef.Args redaction marker 와 safe value non-redaction 을 검증한다.
- `internal/ir/ingest` tests 는 C2 EventIngestor 2차 redaction 과 `PolicyViolationDetected` audit append 를 검증한다.
- 후속 SSOT drift workflow 는 `security.md` 가 redaction 세부 규칙을 재정의하지 않고 본 문서를 링크하는지 확인한다.
