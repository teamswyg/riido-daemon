# 7-9 Audit Invariants and Adjacent SSOT Contracts

[Back to enforcement locations](../enforcement-locations.md)

## 7. Audit invariant

1. **모든 게이트 실패는 IR 이벤트로 남는다.** 운영자가 “왜 거절되었는지” 를 항상 replay 할 수 있다.
2. **scoped token 의 값** 은 IR 에 적지 않는다. `scopeID`, `ttlSeconds`, `purpose` 만.
3. **secret 패턴** 이 IR payload 에 들어갈 위험이 있는 경우, adapter ACL 단계에서 redact 한다(`Unknown` 에도 적지 않는다). 정확한 redaction 규칙은 [`../security-redaction.md`](../security-redaction.md).
4. **trust tier 결정 결과** 는 `RuntimeRegistered` payload 에 포함되어 영속된다.

## 8. 인접 SSOT 와의 계약 (경계 단언)

본 컨텍스트가 “결정” 만 함을 다시 못박는다.

| 인접 context | 본 문서가 공급 | 본 문서가 받지 않음 (그 context 가 owns) |
| --- | --- | --- |
| **C3 Provider Capability** | `ExposesUnsafePermissionBypass` 의 사용 가부, trust tier 보강 입력 | capability detection / fingerprint 계산 / surface flag 집합 |
| **C4 Provider Runtime / Adapter** | provider 에 전달할 flag / env / sandbox 모드 / approval policy 등의 **결정 값** | 실제 process 기동, flag argv 조립, raw → draft 변환 |
| **C5 Runtime Scheduling** | runtime.trustTier × policy bundle 의 호환성, “이 runtime 은 이 task 를 claim 할 수 있다/없다” 의 **결정** | lease DB 행, claim SQL, heartbeat |
| **C6 Workspace** | 어떤 native config 템플릿이 task 의 workdir 에 들어가야 하는가의 **결정** | workdir 디렉토리 생성, 파일 쓰기, 권한 chmod |
| **C8 Validation** | validation rule 목록 / 정책 규칙 셋의 **활성 버전** | test/lint/diff/secret-scan 실제 실행, 결과 해석 |
| **C2 IR Event Log** | Cat F 이벤트의 발행 사유 / payload 스키마 (해당 cat 의 1차 producer) | EventType 카탈로그 자체 |
| **C1 Task Lifecycle** | `Blocked(category=POLICY_*)` 의 사유 카테고리 | TaskState 집합, transition matrix |
| **C11 Distribution / Host Integration** | store channel 에서 금지/허용되는 surface 의 security decision | OS helper 설치 방식, local IPC 구현, app data root 선택, consent ledger 저장 방식 |

## 9. Secret redaction

Secret redaction 세부 결정은 [`../security-redaction.md`](../security-redaction.md)
가 소유한다.

본 문서는 보안 정책 hub 로서 secret exposure target 을 추적하지만, 금지 패턴 카탈로그
/ marker 형식 / C4 1차 redaction / C4 `ToolRef.Args` redaction / C2 EventIngestor
2차 redaction + audit 규칙을 재정의하지 않는다. 다른 문서와 코드는
`security-redaction.md` 를 링크하거나 C7 `internal/policy` helper 를 호출해야 한다.
