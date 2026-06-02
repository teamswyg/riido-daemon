# Security / Policy SSOT

> **이 문서가 trust tier / policy bundle / unsafe bypass 정책 / protected path / network·secrets·destructive action 정책 / 6 security gate / store channel policy decision 의 hub SSOT다.**
>
> - 책임: 무엇이 “허용” 인가, 어느 trust tier 에서 어떤 surface 가 활성 가능한가, 정책 번들의 구조와 진화, 6 security gate 의 위치·검사·실패 동작.
> - 비책임: 정책 **실행** 은 각 인접 context 가 한다. 본 문서는 **결정** 만 한다. workdir 파일 주입 — [`./workspace.md`](./workspace.md) (C6). validation rule 결과 해석 — [`./validation.md`](./validation.md) (C8). security-compatible runtime 배정 — [`./runtime-scheduling.md`](./runtime-scheduling.md) (C5). provider process 에 flag/env 전달은 C4 Provider Runtime 후속 migration slice 가 소유한다.

이 SSOT 는 **C7 Security / Policy** context 를 채운다. C7 은 cross-cutting 으로 C3·C4·C5·C6·C8 모두에 결정을 공급한다. Context map SSOT 는 [`./context-map.md`](./context-map.md) 가 소유한다.

## 0. 핵심 invariant (단단히 박는다)

다음 여섯 invariant 는 본 도메인 전체의 1차 약속이다. 본문 다른 절은 이들의 정교화다.

1. **`ExposesUnsafePermissionBypass=true` 는 risk signal 이며 사용 허가가 아니다.** capability 가 우회 surface 를 노출했다는 사실만으로 그것을 켤 수 없다 — 사용 가부는 본 문서의 policy bundle 게이트가 단독 결정한다. C3 Provider Capability SSOT 는 public `riido-contracts/provider/capability` 가 소유한다.

2. **`Host` trust tier 에서는 unsafe bypass 모드 활성화 절대 금지.** 다음 surface 가 모두 본 invariant 의 대상이다:
   - Claude `--permission-mode bypassPermissions`
   - Codex `--yolo`
   - Codex `--dangerously-bypass-approvals-and-sandbox`
   - Codex sandbox `danger-full-access`
   - wrapper 매니페스트가 자기신고한 동등 surface

   이 모드들은 **isolated trust tier 에서만**, **policy bundle 이 명시적으로 허용** 할 때만 사용 가능하다.

3. **default-deny.** policy bundle 이 명시적으로 허용하지 않은 모든 surface 는 자동으로 거절된다. 옛 정책 → 새 정책 으로 **downgrade(약화) 금지**. 정책 번들 버전은 항상 증가하는 방향으로만 적용된다.

4. **정책 결정과 실행의 분리.** 본 컨텍스트는 “이 surface 를 켜도 되는가?” / “이 path 가 보호 대상인가?” / “이 commit 이 destructive 한가?” 같은 **결정** 만 한다. 실제 flag 전달 / 파일 주입 / process 격리 / lint 실행 / lease 발급은 인접 context 가 수행한다.

5. **policy bundle 도 capability fingerprint 의 입력이다.** 정책 번들이 바뀌면 `CapabilityFingerprint` 가 바뀌고 lease 가 무효화된다. C3 Provider Capability SSOT 는 public `riido-contracts/provider/capability` 가 소유한다. 즉, **정책 변경은 silent 한 “더 허용해주기” 가 될 수 없다** — 진행 중 task 는 명시적으로 재평가된다.

6. **Store channel policy 는 보안 경계다.** `mac-app-store` / `msix-store` channel 에서 금지한 surface 는 provider capability 가 지원해도 사용할 수 없다. Provider CLI bundling, silent provider auto-install, 사용자 동의 없는 background helper, external TCP listener, arbitrary home scan 은 store channel 에서 항상 거절된다. Channel enum 과 role model 의 SSOT 는 [`./distribution-host-integration.md`](./distribution-host-integration.md) (C11) 이다.

## 1. Trust Tier

trust tier 는 “이 runtime 이 실행되는 환경의 격리 수준” 이다. 5 종으로 고정한다.

| Tier | 의미 | unsafe bypass 가능? |
| --- | --- | --- |
| `Host` | 데몬이 사용자/운영자 호스트 OS 위에 직접 실행 | **❌ 절대 금지** (§0 invariant 2) |
| `IsolatedContainer` | OS-level 컨테이너 격리 (Docker / containerd / Podman). 호스트 FS / network 가 격리됨 | policy bundle 명시 허용 시 |
| `EphemeralVM` | task 마다 1회용 VM (firecracker / micro-VM / 격리된 VM 인스턴스) | policy bundle 명시 허용 시 |
| `CIControlledRunner` | CI 시스템이 관리하는 격리 runner (예: GitHub Actions runner, GitLab runner) | policy bundle 명시 허용 시 — CI 의 추가 격리 보장이 있어야 함 |
| `Unknown` | trust tier 결정 불가 (미설정 / 매니페스트 부재 / 검증 실패) | **❌ 절대 금지** |

규칙:

1. trust tier 는 runtime 등록 시점에 **외부 신호** 로 결정된다. 자기 신고(wrapper 매니페스트, runtime 시작 환경 변수, host 정책 파일)와 검증 가능한 사실(예: cgroup / namespace 격리 detection)을 결합한다.
2. `Unknown` 으로 결정되면 capability 가 G-S1 (§3) 에서 거절된다.
3. trust tier 는 runtime 의 `ProviderCapability` 와 함께 `CapabilityFingerprint` 의 입력은 아니다(검증의 외부 사실). 단, 변경되면 capability 재평가가 트리거된다.

## 2. Policy Bundle

policy bundle 은 본 도메인의 **단일 결정 소스** 다. 모든 게이트는 활성 policy bundle 을 입력으로 받아 “허용/거절/보류” 를 답한다.

### 2.1 구조 (도메인 표현)

```
PolicyBundle {
    SchemaVersion    string             // "riido-policy-bundle.v1"
    Version           string             // opaque policy bundle version; 절대 downgrade 불가 (§0 invariant 3)
    EffectiveSince    time.Time
    SupersededAt      time.Time | null

    TrustTierPolicies map[TrustTier] TrustTierPolicy
    DefaultDeny       []PolicyTarget     // 명시적 black-list (예: 항상 거절할 path)
}

TrustTierPolicy {
    AllowedSurfaces   AllowedSurfaceSet   // 9 target 별 허용 표 (§3)
    Sandbox           SandboxPolicy
    NetworkEgress     EgressPolicy
    Secrets           SecretsPolicy
    ProtectedPaths    []PathPattern
    DestructiveOps    DestructiveOpPolicy
    MCPAllowlist      []MCPServerID
    NativeConfigRules NativeConfigPolicy
}
```

### 2.2 진화

- 새 번들은 **항상 새 Version**. 같은 Version 을 덮어쓰지 않는다.
- 두 번들의 차이가 “더 허용” 이라도 **각 task 는 시작 시점의 번들 버전** 으로 평가된다. silent expand 금지.
- 진행 중 task 가 새 번들로 “옮겨가려면” [`../30-architecture/runtime-upgrade-flow.md`](../30-architecture/runtime-upgrade-flow.md) 의 T-POLICY 흐름을 따른다.

### 2.3 actor

정책 번들은 **운영자(human)** 가 PR 로 배포한다. agent / daemon 이 정책을 변경하는 경로는 없다.

### 2.4 실행 artifact — `riido-policy-bundle.v1`

C7 policy bundle 의 현재 물리 형태는 **단일 JSON 파일** 이다. 파일 경로는 daemon 시작 환경의 `RIIDO_POLICY_BUNDLE_PATH` 로 주입하며, 파일 안의 `version` 이 활성 `PolicyBundleVersion` 이 된다. `RIIDO_POLICY_BUNDLE_VERSION` 을 함께 지정한 경우에는 파일 `version` 과 정확히 일치해야 하며, 다르면 daemon 설정 로드를 실패시킨다. 경로가 없으면 기존 local 개발 기본값 `policy-bundle.local.v0` 을 사용한다.

현재 executable loader 는 C7 의 최소 실행 부분인 `unsafe_bypass`, `native_config_hooks`, `native_config_files`, `tool_use` allowed surface 를 받는다. 알 수 없는 필드는 fail-closed 로 거절한다. `RIIDO_POLICY_BUNDLE_PATH` 가 없으면 daemon 은 built-in `policy-bundle.local.v0` 을 사용하며, 이 번들은 Host trust tier 에서 Claude audit-only command hook (`claude:command-hooks:audit`) 만 허용하고 unsafe bypass / native config file / tool-use risk surface 는 허용하지 않는다. Codex app-server 는 native config file surface 가 아니라 C4 command builder 의 mandatory permission profile 로 sandbox 를 고정한다. `RIIDO_POLICY_BUNDLE_VERSION` 만 지정한 dev-local 실행은 같은 built-in allowed surface 에 version tag 만 바꾼다.

```json
{
  "schema_version": "riido-policy-bundle.v1",
  "version": "policy-bundle.example.v1",
  "effective_since": "2026-05-27T00:00:00Z",
  "superseded_at": null,
  "trust_tier_policies": {
    "IsolatedContainer": {
      "allowed_surfaces": {
        "unsafe_bypass": [
          "codex:--yolo"
        ],
        "native_config_hooks": [
          "claude:command-hooks:audit"
        ],
        "native_config_files": [],
        "tool_use": [
          "tool:network-egress"
        ]
      }
    }
  }
}
```

검증 규칙:

1. `schema_version` 은 `riido-policy-bundle.v1` 이어야 한다.
2. `version`, `effective_since`, `trust_tier_policies` 는 필수다.
3. `superseded_at` 이 있으면 `effective_since` 보다 이후여야 한다.
4. `trust_tier_policies` 의 key 는 §1 의 trust tier enum 만 허용한다.
5. `allowed_surfaces.unsafe_bypass` 값은 §5 의 알려진 provider unsafe bypass surface 만 허용하며 중복될 수 없다.
6. `Host` / `Unknown` trust tier 는 bundle 이 어떤 값을 담아도 unsafe bypass surface 를 허용할 수 없다. loader 는 이 조합을 invalid bundle 로 거절한다.
7. `allowed_surfaces.native_config_hooks` 값은 본 문서의 알려진 provider hook surface 만 허용하며 중복될 수 없다. 현재 surface 는 `claude:command-hooks:audit` 뿐이다.
8. `Unknown` trust tier 는 native config hook surface 를 허용할 수 없다. runtime trust tier 가 확정되지 않으면 hook materialization 은 fail-closed 된다.
9. `allowed_surfaces.native_config_files` 값은 본 문서의 알려진 provider config file/home surface 만 허용하며 중복될 수 없다. `codex:config-home:task-scoped` 는 과거 호환을 위해 parser 가 아는 surface 로 남지만 현재 native config plan 은 이를 materialize 하지 않는다.
10. `Unknown` trust tier 는 native config file surface 를 허용할 수 없다. runtime trust tier 가 확정되지 않으면 provider config-home materialization 은 fail-closed 된다.
11. `allowed_surfaces.tool_use` 값은 본 문서 §6 의 알려진 tool-use risk surface 만 허용하며 중복될 수 없다. 현재 surface 는 `tool:network-egress`, `tool:protected-path-write`, `tool:secret-exposure`, `tool:destructive-command` 이다.
12. `Unknown` trust tier 는 tool-use risk surface 를 허용할 수 없다. runtime trust tier 가 확정되지 않으면 ToolUseSecurityGate 는 `interrupt-and-block` 으로 수렴한다.

## 3. Policy 대상 9 종

각 대상은 `AllowedSurfaceSet` 의 한 멤버이며 trust tier 별로 독립 결정된다.

| ID | Target | 정책 표현 |
| --- | --- | --- |
| T-PERM | permission mode | enum: 어떤 permission mode 가 허용되는가 (예: Claude `default`/`acceptEdits`/`plan`만 / `bypassPermissions` 거부) |
| T-SBX | sandbox mode | enum + 기본값 강제 (`read-only` / `workspace-write` / `danger-full-access` 활성 가부) |
| T-NET | network egress | 모드 + allowlist (default-off / explicit allowlist / unrestricted) |
| T-PATH | protected paths | path glob list (`.git/**`, `.env*`, secrets dirs, prod config 등) |
| T-SEC | secret exposure | scoped-token 정책 (TTL, scope 한정, env 전달 거부, 로그 redaction) |
| T-DESTR | destructive command | shell command 패턴 차단 (`rm -rf`, `dd of=/dev/`, `git push --force`, DB drop 등) |
| T-PUSH | git push / deploy / migration | repo, branch, env 별 허용 매트릭스 (예: `main` push 는 human approval 필요) |
| T-MCP | MCP server allowlist | 등록 가능한 MCP server id 목록 + transport 제한 |
| T-CFG | native config injection | task 별로 어떤 정책 파일이 workdir 에 주입되어야 하는가 (CLAUDE.md / AGENTS.md / hooks settings / wrapper manifest) |

각 target 의 enum 값과 의미는 본 문서가 소유한다. 인접 context 는 **결정 결과** 만 받는다(예: C4 는 “T-SBX → workspace-write” 라는 결과를 받아 provider 에 전달).

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
프로세스는 사용자의 기존 Codex 인증 store 를 쓸 수 있지만, C4 adapter 는 매 실행마다
daemon-owned `default_permissions` profile 을 `-c` 로 주입해 provider tool command 가
workdir 과 minimal platform path 만 읽고 쓰게 하며, 사용자 Codex auth/config home 은
filesystem access `none` 으로 deny 한다. 따라서 team id / OpenAPI key / workspace
task-location 값은 Codex 인증 또는 tool sandbox 의 bridge 로 쓰이지 않는다.
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

### 4.2 StoreChannelPolicyGate

`StoreChannelPolicyGate` 는 위 6 runtime security gate 에 번호를 추가하지 않는다. C11 Distribution / Host Integration 이 distribution channel 과 consent state 를 판정한 뒤 C7 에 다음 질문을 던지는 pre-runtime decision 이다.

| 질문 | 호출 context | 실패 시 |
| --- | --- | --- |
| 이 channel 에서 provider CLI 실행을 시작해도 되는가? | C4 pre-execute | provider 미기동 + blocked reason `STORE_CHANNEL_PROVIDER_EXECUTION_BLOCKED` |
| 이 channel 에서 background helper 를 켜도 되는가? | C11 Store App / helper | startup 비활성 + consent UI 표시 |
| 이 channel 에서 workspace root 를 사용할 수 있는가? | C6 workspace prepare | task claim/prepare blocked |
| 이 channel 에서 server 로 보낼 metadata 인가? | C10 SaaS adapter | path/token/absolute workspace root 제거 |

Store channel policy 의 표는 [`./distribution-host-integration.md`](./distribution-host-integration.md) §6 이 소유한다. 본 문서는 그 표의 결정을 security boundary 로 취급한다는 원칙만 소유한다.

`internal/policy.EvaluateStoreChannelPolicy` 는 이 pre-runtime decision 의 현재 순수 구현이다. C11 이 판정한 `DistributionChannel` 과 consent / OS grant / store-review 사실을 입력으로 받아 allow/block `Decision` 만 반환하며, provider process 실행이나 OS adapter 호출은 하지 않는다.

### 4.1 게이트 결과의 IR 표기

| 결과 | EventType | Producer |
| --- | --- | --- |
| 게이트 통과 | `PolicyBundleLoaded` / `PolicyViolationDetected(...)` 없음 / 다음 단계 진행 | (없음 — 통과는 따로 발행 안 함, 진행 자체가 통과 표시) |
| 게이트 실패 | `PolicyViolationDetected(category, subject, severity)` 후속으로 `BlockerRaised(category=POLICY_*)` | 호출 context (G-S1 → scheduler, G-S2 → adapter/orchestrator 등) |
| 정책 번들 교체 | `PolicyBundleSwitched(from, to)` | C7 자체 |
| scoped token 발급 | `SecretsScopeIssued(scopeID, ttl, purpose)` | C7 |
| scoped token 회수 | `SecretsScopeRevoked(scopeID, reason)` | C7 |

값 / payload 카탈로그의 정식 정의는 public `riido-contracts/ir` C2 event catalog 가 소유한다.

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

### 5.1 코드 집행 위치

`internal/policy.EvaluateUnsafeBypass` 는 위 매트릭스의 순수 결정 함수다. C4 provider adapter 는 unsafe bypass 를 실제 provider flag 로 변환하기 직전에 이 결정을 호출한다. C4 provider-runtime 후속 migration slice 가 집행할 표면:

| Surface | 집행 위치 | Host / Unknown 기본 |
| --- | --- | --- |
| Claude `--permission-mode bypassPermissions` | C4 Claude provider `BuildStart` | 거절 |
| Cursor `--yolo` | C4 Cursor provider `BuildStart` | 거절 |
| Codex `--yolo` | C4 Codex provider custom arg filter (`--yolo`, `--yolo=*`) | 거절 |
| Codex `--dangerously-bypass-approvals-and-sandbox` | C4 Codex provider custom arg filter (`--dangerously-bypass-approvals-and-sandbox`, `--dangerously-bypass-approvals-and-sandbox=*`) | 거절 |
| Codex sandbox `danger-full-access` | C4 Codex provider custom arg filter | 거절 |

Codex 의 unsafe bypass 표면은 현재 local daemon command builder 가 생성하지 않으며, free-form `CustomArgs` 에서도 차단된다. 추후 이 표면을 의도적으로 허용하는 PR 은 `StartOptions` 같은 명시적 입력과 `internal/policy.EvaluateUnsafeBypass` 호출을 먼저 추가해야 하며, Host / Unknown trust tier 는 여전히 항상 거절한다.

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

Codex provider 는 항상 `AGENTS.md` 만 native config 로 주입한다. app-server credential 사용과 tool sandbox 는 C4 Codex command builder 의 mandatory permission profile injection 이 소유한다. `Unknown` trust tier 는 모든 provider config-home materialization 을 차단한다.
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

현재 executable ToolUseSecurityGate subset 은 `internal/policy.EvaluateToolUse` 가 소유한다. 이 함수는 provider tool 을 실행하지 않고 C7 decision 만 반환한다. C4 approval flow 에서는 `internal/agentbridge/toolpolicy` 가 provider-neutral `ToolRef` 의 `Kind` / `Name` / redacted `Args` 를 surface 로 분류한 뒤, policy bundle 이 해당 surface 를 명시 허용한 경우에만 `AutoApprover` 를 통해 provider approval command 를 전송한다. `ToolRef.Args` 는 `internal/agentbridge/toolargs` 가 provider raw input 을 bounded string map 으로 flatten 한 값이며, key 가 secret / token / credential 계열이거나 value 가 [`./security-redaction.md`](./security-redaction.md) §1 secret 패턴과 매치되면 raw value 를 저장하지 않고 redaction marker 만 보존한다. redaction marker 를 가진 `ToolRef.Args` 는 `tool:secret-exposure` surface 로 분류된다. 분류되지 않은 tool, policy bundle 에 없는 surface, `Unknown` tier 는 자동 승인하지 않고 기존 human approval path 에 남긴다. provider-native approval RPC/hook 실행 wiring 은 후속 session/runtimeactor/provider-adapter migration slice 가 맡는다.

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
