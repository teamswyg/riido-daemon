# 검증 게이트

[Back to Security Redaction SSOT](../security-redaction.md)

- `internal/policy` tests 는 C7 pattern catalog 와 marker 형식을 검증한다.
- `internal/agentbridge/toolargs` tests 는 ToolRef.Args redaction marker 와 safe value non-redaction 을 검증한다.
- `internal/ir/ingest` tests 는 C2 EventIngestor 2차 redaction 과 `PolicyViolationDetected` audit append 를 검증한다.
- `tools/redactiondrift` 는 `security.md` / `security/**` 가 redaction 세부 규칙을 재정의하지 않고 본 문서를 링크하는지 검증한다.
