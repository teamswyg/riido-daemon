# Security / Policy SSOT: Native Config Overlay

[Back to security.md](../security.md)

### 3.1 T-CFG native config overlay decision

T-CFG 는 provider-native config overlay 를 default-deny 로 다룬다.
`CLAUDE.md` / `AGENTS.md` 같은 primary instruction file materialization 은 C6
Workspace 의 deterministic 기본 surface 이지만, provider-native hook settings,
task-scoped config home, wrapper manifest, MCP config injection 은 policy bundle 이
명시적으로 허용한 surface 일 때만 활성화된다.

No user-global native config overlay is allowed by default. C6 는 사용자 전역
`~/.claude`, `~/.codex`, Cursor/OpenClaw config home 을 읽거나 복사하지 않는다.
Provider-native config home 이 필요한 경우 C7 은 known surface 를 허용하고, C6 는
per-task workdir 안에 materialize 된 config home 만 provider adapter 에 전달한다.
Codex 는 예외적으로 task-scoped `CODEX_HOME` 을 만들지 않는다. Codex app-server
프로세스는 사용자의 기존 Codex 인증 store 를 쓸 수 있고, C4 adapter 는 매 실행마다
`codex --sandbox danger-full-access app-server --listen stdio://` 를 daemon-owned
launch shape 로 고정한다. 따라서 workdir 은 provider 의 기본 작업 위치이자 evidence
root 이지만 filesystem sandbox boundary 가 아니다. team id / OpenAPI key / workspace
task-location 값은 Codex 인증 또는 sandbox bridge 로 쓰이지 않는다.
Codex app-server 가 실행 중 workdir 아래 `.codex` runtime state 를 자체 생성할 수
있지만, 이는 C6 native config materialization 이 아니며 raw auth credential copy,
symlink, SaaS 전달용 credential snapshot 으로 사용하면 안 된다.
정책 또는 native config 변경이 dirty workdir 에 적용되어야 하는 경우 자동
in-place reinjection 을 하지 않고 runtime upgrade flow 의 T-POLICY / T-CONFIG
절차로 넘긴다.

### 3.2 T-SEC runtime secret release evidence

`riido_ai_server` production release 는 runtime secret 값을 evidence 로 저장하지 않는다. T-SEC 의 CaaS release evidence 는 다음 두 축으로 분리한다.

| Evidence kind | Evidence id | 증명 내용 | 금지 내용 |
| --- | --- | --- | --- |
| `runtime-secret-readiness` | `actual-runtime-secret-readiness` | `RIIDO_AI_SERVER_BEARER_TOKEN`, `RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON`, `RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256` 의 reference 존재와 payload shape | raw bearer/authZ/review token 값 |
| `runtime-secret-rotation` | `actual-runtime-secret-rotation` | 위 3개 secret reference 가 `rotatable=true`, `last_rotated_at`, `next_rotation_due_at`, `max_age_seconds` 기준을 만족하고 아직 due 상태가 아님 | raw secret 값, token value, secret payload 본문 |

`runtime-secret-rotation` evidence 는 `riido-runtime-secret-rotation-metadata.v1` input 에서 생성된다. input/evidence 모두 unknown field 를 fail-closed 로 거절해 `value`, `token`, payload 본문 같은 raw secret field 가 섞이면 생성/검증이 실패해야 한다. SSM Parameter Store 를 runtime secret store 로 쓰는 production slice 는 `aws ssm describe-parameters` 의 metadata-only JSON 을 `tools/caasrotationmetadata` 로 변환해 이 input 을 만든다. 이 collector 는 `GetParameter` / `GetParameters` / decrypt 경로를 갖지 않으며 `SecureString` / `Standard` tier / expected parameter name 을 요구하고, SSM `LastModifiedDate` 를 manual overwrite rotation 의 `last_rotated_at` 으로 기록한다. `next_rotation_due_at` 은 `observed_at` 이후여야 하고, `last_rotated_at` 부터 `next_rotation_due_at` 까지의 간격은 각 secret 의 `max_age_seconds` 를 넘을 수 없다. Release packet `apply-ready` mode 는 readiness 와 rotation evidence 를 모두 요구한다.

## 4. 6 Security Gates

본 도메인의 게이트는 **결정 시점** 으로 정렬되어 있다. 각 게이트는 단 하나의 책임을 진다.

| # | Gate | 위치 | 누가 호출 | 입력 | 실패 시 |
| --- | --- | --- | --- | --- | --- |
| G-S1 | **PreClaimSecurityGate** | task claim 직전 | C5 scheduler (G4 의 일부) | (task.requiredSurfaces, runtime.capability, runtime.trustTier, activePolicyBundle) | task `Blocked(category=POLICY_*)` |
| G-S2 | **PreExecuteSecurityGate** | provider process 기동 직전 | C4 adapter / orchestrator (G5 의 일부) | (capability, workdir state, policy bundle) | task `Blocked` + provider 미기동 |
| G-S3 | **ToolUseSecurityGate** | 모든 `ToolCallStarted` 직전 | server transition layer | (toolName, args, runtime.tier, policy bundle) | provider 측 `Interrupt` 또는 `ApprovalRequested` (자세한 결정은 §6) |
| G-S4 | **FileEffectSecurityGate** | 모든 `FileChanged` / `CommandStarted` 직후 | server transition layer | (path, kind, diff, protectedPaths, sandbox) | rollback 요청 + `BlockerRaised(SECURITY_VIOLATION)` |
| G-S5 | **NetworkEgressGate** | 모든 network 시도 (provider sandbox 가 보고하는 시점) | adapter ACL 통해 server | (host, port, protocol, allowlist) | provider 측 deny 결과 + 이벤트 |
| G-S6 | **PreCompleteAuditGate** | `PatchReady → Completed` 전 | C8 validation runner result handler | (전체 IR 로그 / diff / 영향 path 들) | task `Blocked(SECURITY_AUDIT_FAILED)` 또는 `HumanReview` 강제 |

규칙:

1. 게이트는 **결정만** 한다 — 실행은 호출 컨텍스트가 수행한다. 예: G-S2 가 “이 surface 는 거절” 이라고 답하면 C4 는 provider 를 기동하지 않고 C5 가 lease 를 반환한다.
2. 모든 게이트의 결정 결과는 **IR event 로 영속** 된다 (감사 가능성 — §7).
3. 새 게이트 추가는 `change:additive`, 기존 게이트 책임 변경은 `change:breaking-policy`.

### 4.1 StoreChannelPolicyGate

`StoreChannelPolicyGate` 는 위 6 runtime security gate 에 번호를 추가하지 않는다. C11 Distribution / Host Integration 이 distribution channel 과 consent state 를 판정한 뒤 C7 에 다음 질문을 던지는 pre-runtime decision 이다.

| 질문 | 호출 context | 실패 시 |
| --- | --- | --- |
| 이 channel 에서 provider CLI 실행을 시작해도 되는가? | C4 pre-execute | provider 미기동 + blocked reason `STORE_CHANNEL_PROVIDER_EXECUTION_BLOCKED` |
| 이 channel 에서 background helper 를 켜도 되는가? | C11 Store App / helper | startup 비활성 + consent UI 표시 |
| 이 channel 에서 workspace root 를 사용할 수 있는가? | C6 workspace prepare | task claim/prepare blocked |
| 이 channel 에서 server 로 보낼 metadata 인가? | C10 SaaS adapter | path/token/absolute workspace root 제거 |

Store channel policy 의 표는 [`./distribution-host-integration.md`](./distribution-host-integration.md) §6 이 소유한다. 본 문서는 그 표의 결정을 security boundary 로 취급한다는 원칙만 소유한다.

`internal/policy.EvaluateStoreChannelPolicy` 는 이 pre-runtime decision 의 현재 순수 구현이다. C11 이 판정한 `DistributionChannel` 과 consent / OS grant / store-review 사실을 입력으로 받아 allow/block `Decision` 만 반환하며, provider process 실행이나 OS adapter 호출은 하지 않는다.

### 4.2 게이트 결과의 IR 표기

| 결과 | EventType | Producer |
| --- | --- | --- |
| 게이트 통과 | `PolicyBundleLoaded` / `PolicyViolationDetected(...)` 없음 / 다음 단계 진행 | (없음 — 통과는 따로 발행 안 함, 진행 자체가 통과 표시) |
| 게이트 실패 | `PolicyViolationDetected(category, subject, severity)` 후속으로 `BlockerRaised(category=POLICY_*)` | 호출 context (G-S1 → scheduler, G-S2 → adapter/orchestrator 등) |
| 정책 번들 교체 | `PolicyBundleSwitched(from, to)` | C7 자체 |
| scoped token 발급 | `SecretsScopeIssued(scopeID, ttl, purpose)` | C7 |
| scoped token 회수 | `SecretsScopeRevoked(scopeID, reason)` | C7 |

값 / payload 카탈로그의 정식 정의는 public `riido-contracts/ir` C2 event catalog 가 소유한다.

### 4.3 Provider full-access runtime harness

Riido AI Agent 는 provider CLI 를 사용자 PC 위에서 실제로 실행하는 automation 이다.
따라서 provider 에게 충분한 실행 권한을 주지 않으면 “작업을 대신 수행한다”는 제품
목표와 충돌한다. C4 의 현재 canonical 방향은 provider-native full-access/trusted
runtime envelope 를 adapter 가 명시적으로 선택하고, C7/C4/C5/C6 harness 가 실행
전체를 관리하는 것이다. 이는 “provider default 가 full-access” 라는 뜻도 아니고,
“caller 가 full-access 를 고를 수 있다” 는 뜻도 아니다. 반대로 daemon 은 provider
default 나 caller 입력에 기대지 않고, 선택한 runtime envelope 와 그 envelope 를
감싸는 harness 책임을 함께 고정한다.

Codex 의 현재 canonical launch shape 는 다음과 같다.

```text
codex --sandbox danger-full-access app-server --listen stdio://
```

이 값은 default sandbox 가 아니라 **유일하게 daemon 이 생성하는 Codex sandbox
selection** 이다. Caller `CustomArgs`, client codegen, SaaS assignment payload,
policy bundle 은 이를 임의로 바꾸지 않는다. 따라서 Codex 실행 권한의 의미는
“기본값으로 우연히 전권이 되었다”가 아니라 “daemon 이 Codex 를 전권 host automation
으로 실행하고, 그 위험을 하네스가 관리한다”다. C4 adapter 는 caller-provided
`--sandbox`, `--sandbox=*`, `-s`, `-s=*`, `-c`, `--config`, `--enable`,
`--disable`, `--yolo`, `--dangerously-bypass-approvals-and-sandbox` 를 drop 하고
`DroppedArgs` 로 남긴다.

Full-access runtime 을 선택하는 대신 harness 는 다음을 반드시 소유한다.

- immutable assignment snapshot 과 prompt/native-config placement
- daemon-selected workdir 과 결과 evidence root
- provider process start/stop/cancel, terminal result, provider log/progress redaction
- 5초 heartbeat 와 20초 stale 판단을 통한 orphaned assignment 해제
- runtime slot/lease/fencing token, task-level active assignment 정책
- provider 별 real integration gate 와 workdir side-effect 검증

이 결정은 RIID-4881 의 permission-profile 실험을 폐기한다. 그 실험은 Codex auth 401
문제를 해결하려고 user-global auth store 와 provider tool command 권한을 동시에
다루려 했지만, daemon 이 provider 내부 permission semantics 를 과하게 소유하게 만들고
Go/Rust toolchain 같은 정상 작업을 불필요하게 약화시켰다. 새 방향은 권한을 숨기지
않고, 전권 실행을 명시적으로 인정한 뒤 harness / lease / heartbeat / evidence 로
운영한다.

Claude / Cursor / OpenClaw 도 같은 메타 모델을 따른다. 다만 각 provider 의
native full-access/trust flag 는 이름과 효과가 다르므로, Codex 외 provider 를 전권
mode 로 승격하는 PR 은 provider 별 SSOT, command builder, integration evidence 를
같은 PR 에서 갱신해야 한다. Claude `bypassPermissions`, Cursor `--yolo` 처럼
provider 가 “approval bypass” 로 정의한 flag 는 별도 승격 전까지 §5 unsafe bypass
surface 로 남는다.

구조적 판단: 권한을 약하게 만들어 provider 를 반쯤 동작시키는 방식은 Riido 의
agent-ops 목표와 맞지 않는다. 동시에 `danger-full-access` 를 안전한 default 로
부르는 것도 틀렸다. Codex 의 현재 의미는 **default sandbox 가
danger-full-access** 가 아니라 **Codex adapter 가 danger-full-access launch
envelope 만 생성한다** 는 것이다. 그래서 권한은 provider 가 실제 작업을 수행할 수
있을 만큼 명시적으로 열고, 위험은 harness boundary 로 관리한다. 따라서 C7 의 질문은
“full-access 를 기본값으로 둘 것인가?” 가 아니라 “이 provider 에 대해 어떤
trusted-runtime envelope 를 채택했고, 그 envelope 를 어떤 harness evidence 로
운영할 것인가?” 이다. Provider 별 채택 상태와 실행 표면은
[`./provider-runtime.md`](./provider-runtime.md) §2.1 이 소유한다.

## 5. ExposesUnsafePermissionBypass 사용 정책

§0 invariant 1·2 의 정교화. 매트릭스로 못박는다.

> **연구 단계 채택 근거**: 본 §5 의 매트릭스가 "왜 Host × bypass 가 거절인가" 의 결정 자체다. 이 결정은 private source research 의 Claude `--permission-mode bypassPermissions` 기본값 거부, Cursor `--yolo` 거부, Codex unsafe bypass 거부 비교를 통해 도출됐고 본 §5 매트릭스로 흡수됐다. 즉 research 문서는 결정의 **출처(history)** 이고, 본 문서가 결정의 **집행(enforcement)** 이다.

| trust tier × bundle | 동작 |
| --- | --- |
| `Host` × * | **항상 거절**. `BlockedReasons += {Code: "UNSAFE_BYPASS_ON_HOST"}` |
| `IsolatedContainer` × bundle 이 허용 | 허용 — single-task 격리 + protected path 게이트 동반 활성 |
| `IsolatedContainer` × bundle 이 미허용 | 거절 |
| `EphemeralVM` × bundle 이 허용 | 허용 |
| `EphemeralVM` × bundle 이 미허용 | 거절 |
| `CIControlledRunner` × bundle 이 허용 | 허용 (CI 의 격리 보장 검증 후) |
| `Unknown` × * | **항상 거절** |

“bundle 이 허용” 의 현재 실행 의미: `bundle.TrustTierPolicies[<tier>].AllowedSurfaces.UnsafeBypass` 가 해당 provider unsafe bypass surface 를 명시적으로 포함. **추론 / 기본값 허용 없음**.
