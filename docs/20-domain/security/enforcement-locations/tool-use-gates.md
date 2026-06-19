# 6 ToolUse / FileEffect / NetworkEgress Branches

[Back to enforcement locations](../enforcement-locations.md)

세 게이트(G-S3 / G-S4 / G-S5)는 빈도가 높아 “매번 거절 vs 허용” 단순 결정 외에
다음 분기를 둘 수 있다.

| 결과 | 의미 |
| --- | --- |
| `allow` | 진행 |
| `allow-but-audit` | 진행 + `OperatorNote` 자동 추가 |
| `require-approval` | provider 측 approval 프로토콜 (Codex app-server) 또는 task `NeedsInput` 으로 전이 |
| `interrupt-and-block` | provider interrupt + `BlockerRaised(SECURITY_VIOLATION)` |
| `quarantine` | provider process 종료 + workdir 격리 보존(분석용) + `TaskFailed(reason=SECURITY_QUARANTINE)` |

현재 executable ToolUseSecurityGate subset 은 `internal/policy.EvaluateToolUse` 가
소유한다. 이 함수는 provider tool 을 실행하지 않고 C7 decision 만 반환한다. C4
approval flow 에서는 `internal/agentbridge/toolpolicy` 가 provider-neutral `ToolRef`
의 `Kind` / `Name` / redacted `Args` 를 surface 로 분류한 뒤, policy bundle 이 해당
surface 를 명시 허용한 경우에만 `AutoApprover` 를 통해 provider approval command 를
전송한다. `ToolRef.Args` 는 `internal/agentbridge/toolargs` 가 provider raw input 을
bounded string map 으로 flatten 한 값이며, key 가 secret / token / credential 계열이거나
value 가 [`../security-redaction.md`](../security-redaction.md) §1 secret 패턴과
매치되면 raw value 를 저장하지 않고 redaction marker 만 보존한다. redaction marker 를
가진 `ToolRef.Args` 는 `tool:secret-exposure` surface 로 분류된다. 분류되지 않은 tool,
policy bundle 에 없는 surface, `Unknown` tier 는 자동 승인하지 않고 기존 human
approval path 에 남긴다. 현재 daemon-local 실행 wiring 은 provider 가 노출한 approval
request 에 자동 승인/거절 응답을 줄 수 있고, approval round-trip 없이 이미 시작된
classified tool 은 fail-closed 로 provider 를 취소한다. provider-native hook/RPC 로 tool
실행 **직전** 에 차단하는 pre-start interrupt 와 SaaS/web approval request/decision
handoff 는 후속 work unit 이 맡는다.

| Surface | 의미 | `allowed_surfaces.tool_use` 미포함 + approval 가능 | approval 불가 / Unknown tier |
| --- | --- | --- | --- |
| `tool:network-egress` | provider tool 이 외부 network 로 나가려는 risk surface | `require-approval` | `interrupt-and-block` |
| `tool:protected-path-write` | protected path 에 쓰기/삭제/권한 변경을 시도하는 risk surface | `require-approval` | `interrupt-and-block` |
| `tool:secret-exposure` | secret/raw token 이 tool input/output 으로 노출될 risk surface | `require-approval` | `interrupt-and-block` |
| `tool:destructive-command` | destructive shell/db/git/deploy command risk surface | `require-approval` | `interrupt-and-block` |

정책 번들이 해당 surface 를 trust tier 별로 명시하면 `allow` 다. 명시하지 않았지만
provider/runtime 이 human approval 경로를 제공하면 `require-approval` 을 반환한다.
approval 경로가 없거나 trust tier 가 `Unknown` 이면 `interrupt-and-block` 이다. 현재
C4 wiring 은 provider 가 `ApprovalRequested` 를 노출하는 경로의 자동 승인 여부,
provider-neutral `ToolCallStarted` / `ApprovalRequested` IR payload 의 redacted args
보존, approval round-trip 없이 관측된 classified `ToolCallStarted` 의 fail-closed
provider kill + `ResultBlocked` 종료까지 실행한다. provider-native hook/RPC 로 tool 실행
**직전** 에 차단하는 pre-start interrupt 는 후속 work unit 이 맡는다. C4/C5/C8 은 이
결정을 실행으로 옮기고 IR event 영속화를 맡는다.

각 게이트가 어떤 분기들을 지원하는가는 `AllowedSurfaceSet` 의 plan 에 따라 다르다.
모든 분기는 §4.1 의 IR event 로 영속된다.
