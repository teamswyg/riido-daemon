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

## Detail Surfaces

1. [금지 패턴 카탈로그](security-redaction/pattern-catalog.md)
2. [Marker](security-redaction/markers.md)
3. [실행 책임](security-redaction/execution-responsibility.md)
4. [Audit Event](security-redaction/audit-event.md)
5. [Redaction Failure](security-redaction/failure.md)
6. [Versioning](security-redaction/versioning.md)
7. [검증 게이트](security-redaction/verification-gates.md)
