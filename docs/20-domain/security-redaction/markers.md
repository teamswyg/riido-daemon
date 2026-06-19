# Marker

[Back to Security Redaction SSOT](../security-redaction.md)

Canonical IR payload redaction marker 는 다음 형식이다.

```text
[REDACTED:<patternID>]
```

`ToolRef.Args` 는 policy classification 입력용 bounded summary 이므로 pattern ID 를 보존하지 않고 단일 marker 를 쓴다.

```text
[redacted]
```

`ToolRef.Args` 에 `[redacted]` marker 가 있으면 `tool:secret-exposure` risk surface 로 분류한다.
