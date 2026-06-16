# Security / Policy SSOT: Enforcement Locations

[Back to security.md](../security.md)

### 5.1 코드 집행 위치

`internal/policy.EvaluateUnsafeBypass` 는 위 매트릭스의 순수 결정 함수다. C4 provider adapter 는 unsafe bypass 를 실제 provider flag 로 변환하기 직전에 이 결정을 호출한다. C4 provider-runtime 후속 migration slice 가 집행할 표면:

| Surface | 집행 위치 | Host / Unknown 기본 |
| --- | --- | --- |
| Claude `--permission-mode bypassPermissions` | C4 Claude provider `BuildStart` | 거절 |
| Cursor `--yolo` | C4 Cursor provider `BuildStart` | 거절 |
| Codex `--yolo` | C4 Codex provider custom arg filter (`--yolo`, `--yolo=*`) | 거절 |
| Codex `--dangerously-bypass-approvals-and-sandbox` | C4 Codex provider custom arg filter (`--dangerously-bypass-approvals-and-sandbox`, `--dangerously-bypass-approvals-and-sandbox=*`) | 거절 |

Codex 의 approval-bypass 표면은 현재 local daemon command builder 가 생성하지 않으며, free-form `CustomArgs` 에서도 차단된다. Codex `--sandbox danger-full-access` 는 §4.3 의 provider full-access runtime envelope 로서 C4 command builder 가 직접 생성하며, §5 unsafe bypass policy bundle surface 로 다루지 않는다. 추후 provider-native approval bypass 표면을 의도적으로 허용하는 PR 은 `StartOptions` 같은 명시적 입력과 `internal/policy.EvaluateUnsafeBypass` 호출을 먼저 추가해야 하며, Host / Unknown trust tier 는 여전히 항상 거절한다.

Cursor `--trust` 는 이 표의 unsafe bypass surface 가 아니다. 이는 Cursor Agent 가
daemon 이 선택한 task-scoped workspace 에서 headless 로 실행될 때 interactive
workspace trust prompt 로 멈추지 않게 하는 workspace acknowledgement 다. `--trust`
를 붙여도 Cursor `--yolo`, `-f`, tool auto-approval, sandbox 우회가 암묵적으로
허용되지 않는다. 따라서 Cursor adapter 는 daemon task workdir 을 지정할 때 `--trust`
를 붙일 수 있지만, `--yolo` 는 위 매트릭스와
`internal/policy.EvaluateUnsafeBypass` 집행을 계속 통과해야 한다.

### 5.2 Native config hook materialization 정책

`internal/policy.EvaluateNativeConfigHook` 은 T-CFG provider-native hook surface 의 순수 결정 함수다. 현재 실행 surface 는 Claude Code command hook 설정/스크립트 주입을 audit-only 로 허용하는 `claude:command-hooks:audit` 하나다. C6 `internal/workdir` 는 hook 정책을 결정하지 않고, C4/C6 경계의 supervisor 가 active policy bundle 을 평가해 `claude-command-hooks` 또는 `instruction-only` hook mode 를 넘긴다.

| Surface | 집행 위치 | 기본 local policy |
| --- | --- | --- |
| Claude audit-only command hooks (`.claude/settings.json`, `.riido/hooks/claude-audit-hook.sh`) | C4 supervisor → `internal/workdir.InjectRuntimeConfig` | Host 에서 허용 |

policy bundle 이 해당 surface 를 허용하지 않으면 Claude provider 여도 `CLAUDE.md` 만 주입하고 `.claude/settings.json` / hook script 는 만들지 않는다. `Unknown` trust tier 는 항상 `instruction-only` 로 수렴한다.

### 5.3 Native config file materialization 정책

`internal/policy.EvaluateNativeConfigFile` 은 T-CFG provider-native config file/home surface 의 순수 결정 함수다. 현재 기본 local bundle 은 native config file surface 를 허용하지 않는다. `codex:config-home:task-scoped` 는 과거 compatibility surface 로 남아 있지만 현재 `riido-native-config-plan.v1` 은 Codex `.codex/config.toml` 또는 adapter `CODEX_HOME=<workdir>/.codex` metadata 를 materialize 하지 않는다.

| Surface | 집행 위치 | 기본 local policy |
| --- | --- | --- |
| Codex task-scoped config home (`.codex/config.toml`, adapter `CODEX_HOME`) | compatibility parser surface only | Host 에서 비허용 |

Codex provider 는 항상 `AGENTS.md` 만 native config 로 주입한다. app-server credential 사용과 full-access sandbox envelope 는 C4 Codex command builder 가 소유한다. `Unknown` trust tier 는 모든 provider config-home materialization 을 차단한다.
Codex process 가 실행 중 생성하는 workdir `.codex` state 는 이 표의 허용 대상이 아니며,
daemon 은 그 경로를 provider-native config overlay 로 선언하지 않는다.

### 5.4 MCP temp file lifecycle 정책

MCP raw JSON config 는 provider command line 에 직접 inline 하지 않고 adapter 가 소유한 temp file path 로만 전달한다. C4 adapter 는 `StartCommand.TempFiles` 에 이 path 를 보고하고, C4 session actor 는 provider process exit / cancellation / timeout 으로 run 이 종료될 때 해당 temp file 을 삭제한다.

이 규칙은 Claude `--mcp-config` 처럼 provider protocol 이 file path 를 요구하는 경우의 lifecycle gate 다. Temp file 삭제 실패는 run-scope warning 으로만 남기고 terminal result 를 바꾸지 않는다. 같은 path 가 중복 보고되거나 이미 삭제된 경우는 idempotent cleanup 으로 처리한다. Provider adapter 는 temp file path 를 `DroppedArgs` / telemetry payload / command warning 에 raw config 내용으로 풀어 쓰면 안 된다.

## 6. ToolUse / FileEffect / NetworkEgress 의 분기

세 게이트(G-S3 / G-S4 / G-S5)는 빈도가 높아 “매번 거절 vs 허용” 단순 결정 외에 다음 분기를 둘 수 있다.

| 결과 | 의미 |
| --- | --- |
| `allow` | 진행 |
| `allow-but-audit` | 진행 + `OperatorNote` 자동 추가 |
| `require-approval` | provider 측 approval 프로토콜 (Codex app-server) 또는 task `NeedsInput` 으로 전이 |
| `interrupt-and-block` | provider interrupt + `BlockerRaised(SECURITY_VIOLATION)` |
| `quarantine` | provider process 종료 + workdir 격리 보존(분석용) + `TaskFailed(reason=SECURITY_QUARANTINE)` |

현재 executable ToolUseSecurityGate subset 은 `internal/policy.EvaluateToolUse` 가 소유한다. 이 함수는 provider tool 을 실행하지 않고 C7 decision 만 반환한다. C4 approval flow 에서는 `internal/agentbridge/toolpolicy` 가 provider-neutral `ToolRef` 의 `Kind` / `Name` / redacted `Args` 를 surface 로 분류한 뒤, policy bundle 이 해당 surface 를 명시 허용한 경우에만 `AutoApprover` 를 통해 provider approval command 를 전송한다. `ToolRef.Args` 는 `internal/agentbridge/toolargs` 가 provider raw input 을 bounded string map 으로 flatten 한 값이며, key 가 secret / token / credential 계열이거나 value 가 [`./security-redaction.md`](./security-redaction.md) §1 secret 패턴과 매치되면 raw value 를 저장하지 않고 redaction marker 만 보존한다. redaction marker 를 가진 `ToolRef.Args` 는 `tool:secret-exposure` surface 로 분류된다. 분류되지 않은 tool, policy bundle 에 없는 surface, `Unknown` tier 는 자동 승인하지 않고 기존 human approval path 에 남긴다. 현재 daemon-local 실행 wiring 은 provider 가 노출한 approval request 에 자동 승인/거절 응답을 줄 수 있고, approval round-trip 없이 이미 시작된 classified tool 은 fail-closed 로 provider 를 취소한다. provider-native hook/RPC 로 tool 실행 **직전** 에 차단하는 pre-start interrupt 와 SaaS/web approval request/decision handoff 는 후속 work unit 이 맡는다.

| Surface | 의미 | `allowed_surfaces.tool_use` 미포함 + approval 가능 | approval 불가 / Unknown tier |
| --- | --- | --- | --- |
| `tool:network-egress` | provider tool 이 외부 network 로 나가려는 risk surface | `require-approval` | `interrupt-and-block` |
| `tool:protected-path-write` | protected path 에 쓰기/삭제/권한 변경을 시도하는 risk surface | `require-approval` | `interrupt-and-block` |
| `tool:secret-exposure` | secret/raw token 이 tool input/output 으로 노출될 risk surface | `require-approval` | `interrupt-and-block` |
| `tool:destructive-command` | destructive shell/db/git/deploy command risk surface | `require-approval` | `interrupt-and-block` |

정책 번들이 해당 surface 를 trust tier 별로 명시하면 `allow` 다. 명시하지 않았지만 provider/runtime 이 human approval 경로를 제공하면 `require-approval` 을 반환한다. approval 경로가 없거나 trust tier 가 `Unknown` 이면 `interrupt-and-block` 이다. 현재 C4 wiring 은 provider 가 `ApprovalRequested` 를 노출하는 경로의 자동 승인 여부, provider-neutral `ToolCallStarted` / `ApprovalRequested` IR payload 의 redacted args 보존, approval round-trip 없이 관측된 classified `ToolCallStarted` 의 fail-closed provider kill + `ResultBlocked` 종료까지 실행한다. provider-native hook/RPC 로 tool 실행 **직전** 에 차단하는 pre-start interrupt 는 후속 work unit 이 맡는다. C4/C5/C8 은 이 결정을 실행으로 옮기고 IR event 영속화를 맡는다.

각 게이트가 어떤 분기들을 지원하는가는 `AllowedSurfaceSet` 의 plan 에 따라 다르다. 모든 분기는 §4.1 의 IR event 로 영속된다.

## 7. Audit invariant

1. **모든 게이트 실패는 IR 이벤트로 남는다.** 운영자가 “왜 거절되었는지” 를 항상 replay 할 수 있다.
2. **scoped token 의 값** 은 IR 에 적지 않는다. `scopeID`, `ttlSeconds`, `purpose` 만.
3. **secret 패턴** 이 IR payload 에 들어갈 위험이 있는 경우, adapter ACL 단계에서 redact 한다(`Unknown` 에도 적지 않는다). 정확한 redaction 규칙은 [`./security-redaction.md`](./security-redaction.md).
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

Secret redaction 세부 결정은 [`./security-redaction.md`](./security-redaction.md) 가 소유한다.

본 문서는 보안 정책 hub 로서 secret exposure target 을 추적하지만, 금지 패턴 카탈로그 / marker 형식 / C4 1차 redaction / C4 `ToolRef.Args` redaction / C2 EventIngestor 2차 redaction + audit 규칙을 재정의하지 않는다. 다른 문서와 코드는 `security-redaction.md` 를 링크하거나 C7 `internal/policy` helper 를 호출해야 한다.

## 10. 미결정 / 오픈 이슈

[`../50-roadmap/open-questions.md`](../50-roadmap/open-questions.md) 위임.

- `Q-SEC-001`: trust tier 결정의 **외부 신호 우선순위** (wrapper 매니페스트 vs daemon 시작 env vs 호스트 정책 파일).
- `Q-SEC-003`: scoped token 의 발급 / 회수 메커니즘 (자체 발급 vs Vault / cloud KMS 위임).
- `Q-SEC-004`: G-S6 (PreCompleteAuditGate)가 자동 `HumanReview` 로 강제하는 조건의 임계값.
- `Q-SEC-006`: `quarantine` 분기의 workdir 격리 보존 기간 / 자동 삭제 정책.
- `Q-SEC-007`: `CIControlledRunner` tier 의 “격리 보장 검증” 알고리즘 (어떤 신호를 신뢰할지).
- `Q-SEC-008`: StoreChannelPolicyGate 의 blocked reason 을 Cat F IR 이벤트로 별도 추가할지, 기존 `PolicyViolationDetected` payload 로 흡수할지.

## 11. version-affecting changes

- 새 trust tier 추가는 `change:breaking-policy` (모든 정책 번들이 새 tier 의 정책을 추가해야 함).
- 새 policy target 추가는 `change:breaking-policy` (정책 번들 schema 가 확장됨).
- 새 게이트 추가는 `change:additive`. 게이트의 책임 변경 / 제거는 `change:breaking-policy`.
- 우회 surface 의 “허용 매트릭스” (§5) 변경은 항상 `change:breaking-policy` + 정책 번들 버전 강제 증가.
- secret redaction 변경의 version-affecting 규칙은 [`./security-redaction.md`](./security-redaction.md) §6 이 소유한다.
- store channel 에서 금지된 surface 를 완화하는 변경은 `change:breaking-policy` 이며 C11 distribution SSOT 와 같은 PR 에서 갱신해야 한다.
